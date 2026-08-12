package tms

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"regexp"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/schema"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Типы импорта TestRail
// ---------------------------------------------------------------------------

// TestRailCase представляет тест-кейс TestRail из CSV/XML экспорта.
type TestRailCase struct {
	ID             string `json:"id" xml:"id"`
	Title          string `json:"title" xml:"title"`
	SectionName    string `json:"section" xml:"section"`
	Type           string `json:"type" xml:"type"`         // Automated, Manual и т.п.
	Priority       string `json:"priority" xml:"priority"` // Critical, High, Medium, Low
	Estimate       string `json:"estimate" xml:"estimate"` // например "30s", "1m", "1h"
	Preconditions  string `json:"preconditions" xml:"preconditions"`
	Steps          string `json:"steps" xml:"steps"`           // шаги в HTML или простом тексте
	ExpectedResult string `json:"expected" xml:"expected"`
	References     string `json:"refs" xml:"refs"` // ID задач через запятую
}

// TestRailXMLRoot — корневой элемент XML-экспорта TestRail.
type TestRailXMLRoot struct {
	XMLName xml.Name       `xml:"cases"`
	Cases   []TestRailCase `xml:"case"`
}

// ---------------------------------------------------------------------------
// Маппинг приоритетов TestRail → Hub TMS
// ---------------------------------------------------------------------------

// mapTestRailPriority преобразует приоритет TestRail в приоритет Hub TMS.
func mapTestRailPriority(trPriority string) string {
	switch strings.ToLower(strings.TrimSpace(trPriority)) {
	case "critical", "1 - critical", "p1":
		return "critical"
	case "high", "2 - high", "p2":
		return "high"
	case "medium", "3 - medium", "p3":
		return "medium"
	case "low", "4 - low", "p4":
		return "low"
	case "trivial", "5 - trivial":
		return "trivial"
	default:
		return "medium"
	}
}

// mapTestRailType преобразует тип TestRail в case_type Hub TMS (classic/bdd).
// TestRail «Type» — категория кейса (Manual, Automated, Functional, …), не формат шагов.
// BDD-шаблоны и Gherkin-сценарии маппятся в bdd, остальное — classic.
func mapTestRailType(trType string) string {
	switch strings.ToLower(strings.TrimSpace(trType)) {
	case "bdd", "behavior", "behaviour", "scenario", "gherkin", "cucumber", "feature":
		return "bdd"
	case "classic":
		return "classic"
	default:
		return "classic"
	}
}

// ---------------------------------------------------------------------------
// Парсинг шагов TestRail
// ---------------------------------------------------------------------------

// htmlTagRegexp используется для удаления HTML-тегов из шагов TestRail.
var htmlTagRegexp = regexp.MustCompile(`<[^>]*>`)

// W2-A4-011 / P3-BE-07: compile once — trimStepNumbering вызывается на каждый шаг CSV.
var (
	stepNumberPrefixRegexp = regexp.MustCompile(`^\d+[\.\)]\s*`)
	stepLabelPrefixRegexp  = regexp.MustCompile(`(?i)^step\s*\d+[:\s]*`)
)

// stripHTML удаляет HTML-теги и обрезает пробелы.
func stripHTML(s string) string {
	text := htmlTagRegexp.ReplaceAllString(s, "")
	return strings.TrimSpace(text)
}

// parseTestRailSteps конвертирует текст шагов TestRail (возможно HTML) в структурированные шаги.
// Шаги TestRail обычно пронумерованы или разделены переносами строк.
func parseTestRailSteps(stepsText, expectedText string) []schema.TestStep {
	if stepsText == "" {
		return nil
	}

	plain := stripHTML(stepsText)
	lines := strings.Split(plain, "\n")

	expectedPlain := stripHTML(expectedText)
	expectedLines := strings.Split(expectedPlain, "\n")

	var steps []schema.TestStep
	order := 1
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		// Убираем ведущую нумерацию вида "1. ", "1) ", "Step 1: "
		line = trimStepNumbering(line)
		if line == "" {
			continue
		}

		expected := ""
		if order-1 < len(expectedLines) {
			exp := strings.TrimSpace(expectedLines[order-1])
			exp = trimStepNumbering(exp)
			expected = exp
		}

		steps = append(steps, schema.TestStep{
			Order:          order,
			Action:         line,
			ExpectedResult: expected,
		})
		order++
	}

	// Если структурированных шагов нет, но есть ожидаемый текст — создаём один шаг.
	if len(steps) == 0 && expectedText != "" {
		steps = append(steps, schema.TestStep{
			Order:          1,
			Action:         stripHTML(stepsText),
			ExpectedResult: stripHTML(expectedText),
		})
	}

	return steps
}

// trimStepNumbering удаляет типичные префиксы нумерации шагов: "1. ", "1) ", "Step 1: ".
func trimStepNumbering(s string) string {
	s = stepNumberPrefixRegexp.ReplaceAllString(s, "")
	s = stepLabelPrefixRegexp.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

// ---------------------------------------------------------------------------
// Импорт CSV
// ---------------------------------------------------------------------------

// testRailCSVHeaders маппит известные названия столбцов CSV TestRail в наши внутренние имена полей.
var testRailCSVHeaders = map[string]string{
	"id":             "id",
	"case id":        "id",
	"title":          "title",
	"name":           "title",
	"section":        "section",
	"suite":          "section",
	"section name":   "section",
	"type":           "type",
	"case type":      "type",
	"priority":       "priority",
	"estimate":       "estimate",
	"preconditions":  "preconditions",
	"precondition":   "preconditions",
	"steps":          "steps",
	"steps (text)":   "steps",
	"expected result": "expected",
	"expected":       "expected",
	"references":     "refs",
	"refs":           "refs",
}

// ImportTestRailCSV импортирует тест-кейсы из CSV-экспорта TestRail.
func (s *TMSService) ImportTestRailCSV(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, data []byte) (*ImportResult, error) {
	cases, err := parseTestRailCSV(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing TestRail CSV: %w", err)
	}
	return s.importTestRailCases(ctx, projectID, userID, cases)
}

// ImportTestRailXML импортирует тест-кейсы из XML-экспорта TestRail.
func (s *TMSService) ImportTestRailXML(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, data []byte) (*ImportResult, error) {
	cases, err := parseTestRailXML(data)
	if err != nil {
		return nil, fmt.Errorf("error parsing TestRail XML: %w", err)
	}
	return s.importTestRailCases(ctx, projectID, userID, cases)
}

// importTestRailCases создаёт папки и тест-кейсы из распарсенных данных TestRail.
func (s *TMSService) importTestRailCases(ctx context.Context, projectID uuid.UUID, userID uuid.UUID, cases []TestRailCase) (*ImportResult, error) {
	result := &ImportResult{}

	if len(cases) == 0 {
		return result, nil
	}

	// Создаём папки из уникальных имён секций.
	suiteName2ID := make(map[string]*uuid.UUID)
	for _, tc := range cases {
		section := strings.TrimSpace(tc.SectionName)
		if section == "" {
			continue
		}
		if _, exists := suiteName2ID[section]; exists {
			continue
		}
		suite, err := s.repo.CreateSuite(ctx, projectID, CreateSuiteInput{
			Name:      section,
			SortOrder: len(suiteName2ID),
		})
		if err != nil {
			// A3W2-4-005: fail CreateSuite не маскируется как полный успех.
			result.Errors++
			result.ValidationErrors = append(result.ValidationErrors, ImportValidationError{
				Row:     0,
				Field:   "section",
				Message: "failed to create suite: " + section,
			})
			slog.Error("failed to create suite for TestRail import", "section", section, "error", err)
			continue
		}
		suiteName2ID[section] = &suite.ID
	}

	// Создаём тест-кейсы.
	for i, tc := range cases {
		title := strings.TrimSpace(tc.Title)
		if title == "" {
			result.Skipped++
			result.Errors++
			result.ValidationErrors = append(result.ValidationErrors, ImportValidationError{
				Row:     i + 2,
				Field:   "title",
				Message: "title is required",
			})
			continue
		}

		steps := parseTestRailSteps(tc.Steps, tc.ExpectedResult)

		input := CreateTestCaseInput{
			Title:        title,
			Precondition: stripHTML(tc.Preconditions),
			Priority:     mapTestRailPriority(tc.Priority),
			CaseType:     mapTestRailType(tc.Type),
			Steps:        steps,
			CreatedByID:  userID,
		}

		section := strings.TrimSpace(tc.SectionName)
		if suiteID, ok := suiteName2ID[section]; ok {
			input.SuiteID = suiteID
		}

		_, err := s.repo.CreateTestCase(ctx, projectID, input)
		if err != nil {
			result.Errors++
			slog.Error("failed to create test case during TestRail import",
				"row", i+2, "title", title, "error", err)
			continue
		}
		result.Created++
	}

	result.Partial = result.Errors > 0 || result.Skipped > 0
	return result, nil
}

// ---------------------------------------------------------------------------
// Парсеры
// ---------------------------------------------------------------------------

// parseTestRailCSV парсит CSV-экспорт TestRail в структуры TestRailCase.
// A3W2-4-002: records cap (паритет с ImportFromCSV / maxImportRecords).
func parseTestRailCSV(data []byte) ([]TestRailCase, error) {
	// W2-A4-010 / P3-BE-07: только tryReadCSV (мёртвый csv.NewReader удалён).
	records, err := tryReadCSV(data, ',')
	if err != nil || len(records) < 2 {
		records, err = tryReadCSV(data, ';')
		if err != nil || len(records) < 2 {
			records, err = tryReadCSV(data, '\t')
			if err != nil {
				return nil, fmt.Errorf("failed to parse CSV with any separator: %w", err)
			}
		}
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV file is empty or contains only headers")
	}

	if len(records)-1 > maxImportRecords {
		return nil, fmt.Errorf("слишком много записей для импорта (макс. %d)", maxImportRecords)
	}

	// Строим индекс столбцов.
	headers := records[0]
	colMap := make(map[string]int) // поле -> индекс столбца
	for i, h := range headers {
		lower := strings.ToLower(strings.TrimSpace(h))
		if field, ok := testRailCSVHeaders[lower]; ok {
			colMap[field] = i
		}
	}

	var cases []TestRailCase
	for _, row := range records[1:] {
		tc := TestRailCase{
			ID:             csvField(row, colMap, "id"),
			Title:          csvField(row, colMap, "title"),
			SectionName:    csvField(row, colMap, "section"),
			Type:           csvField(row, colMap, "type"),
			Priority:       csvField(row, colMap, "priority"),
			Estimate:       csvField(row, colMap, "estimate"),
			Preconditions:  csvField(row, colMap, "preconditions"),
			Steps:          csvField(row, colMap, "steps"),
			ExpectedResult: csvField(row, colMap, "expected"),
			References:     csvField(row, colMap, "refs"),
		}
		cases = append(cases, tc)
	}

	return cases, nil
}

// tryReadCSV пытается прочитать все записи из CSV-данных с указанным разделителем.
func tryReadCSV(data []byte, sep rune) ([][]string, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.Comma = sep
	reader.LazyQuotes = true
	reader.TrimLeadingSpace = true
	return reader.ReadAll()
}

// csvField извлекает значение поля из CSV-строки по маппированному имени столбца.
func csvField(row []string, colMap map[string]int, field string) string {
	if idx, ok := colMap[field]; ok && idx < len(row) {
		return row[idx]
	}
	return ""
}

// parseTestRailXML парсит XML-экспорт TestRail в структуры TestRailCase.
// A3W2-4-002: depth + records cap (паритет с ImportFromXML / allure/junit).
func parseTestRailXML(data []byte) ([]TestRailCase, error) {
	if err := assertXMLMaxDepth(data, maxImportNestDepth); err != nil {
		return nil, err
	}
	var root TestRailXMLRoot
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("error parsing TestRail XML: %w", err)
	}
	if len(root.Cases) > maxImportRecords {
		return nil, fmt.Errorf("слишком много записей для импорта (макс. %d)", maxImportRecords)
	}
	return root.Cases, nil
}

// ---------------------------------------------------------------------------
// HTTP-обработчик
// ---------------------------------------------------------------------------

// maxTestRailImportBytes — лимит CSV/XML файла TestRail (совпадает с Fiber BodyLimit 25 MB).
const maxTestRailImportBytes = 25 * 1024 * 1024

// ImportTestRailCases обрабатывает POST /api/projects/:projectId/tms/import/testrail.
func (h *Handler) ImportTestRailCases(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}

	f, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "error opening file")
	}
	defer f.Close()

	data, err := io.ReadAll(io.LimitReader(f, maxTestRailImportBytes+1))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "error reading file")
	}
	if len(data) > maxTestRailImportBytes {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "file too large")
	}

	// Определяем формат по расширению файла или полю формы.
	format := c.FormValue("format", "")
	if format == "" {
		name := strings.ToLower(file.Filename)
		switch {
		case strings.HasSuffix(name, ".xml"):
			format = "xml"
		default:
			format = "csv"
		}
	}

	var result *ImportResult
	switch format {
	case "xml":
		result, err = h.service.ImportTestRailXML(c.Context(), projectID, userID, data)
	default:
		result, err = h.service.ImportTestRailCSV(c.Context(), projectID, userID, data)
	}
	if err != nil {
		// A2W2-4-007: не отдаём internals парсера клиенту.
		slog.Error("ошибка импорта TestRail", "error", err, "format", format)
		return fiber.NewError(fiber.StatusBadRequest, "некорректный файл TestRail")
	}

	return c.JSON(result)
}
