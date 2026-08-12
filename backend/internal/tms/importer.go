package tms

import (
	"bytes"
	"context"
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/xuri/excelize/v2"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/schema"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Форматы импорта
// ---------------------------------------------------------------------------

// ImportFormat определяет формат файла для импорта.
type ImportFormat string

// Настройки импорта тест-кейсов.
const (
	ImportFormatCSV   ImportFormat = "csv"
	ImportFormatJSON  ImportFormat = "json"
	ImportFormatXML   ImportFormat = "xml"
	ImportFormatExcel ImportFormat = "excel"

	// maxImportFileBytes — лимит тела файла импорта (как TestRail / Fiber BodyLimit).
	maxImportFileBytes = 25 * 1024 * 1024
	// maxImportRecords — hard-cap числа кейсов за один импорт (DoS/memory).
	maxImportRecords = 5000
	// maxImportNestDepth — лимит вложенности JSON/XML (bomb amplification).
	maxImportNestDepth = 32
)

// FieldMapping описывает соответствие колонок источника полям TMS.
type FieldMapping map[string]string

// AutoDetectMapping автоматически определяет маппинг по заголовкам колонок.
// Поддерживает русские и английские названия полей.
func AutoDetectMapping(headers []string) FieldMapping {
	mapping := make(FieldMapping)
	knownFields := map[string]string{
		"title": "title", "название": "title", "test name": "title", "name": "title",
		"description": "description", "описание": "description",
		"precondition": "precondition", "предусловие": "precondition",
		"steps": "steps", "шаги": "steps",
		"priority": "priority", "приоритет": "priority",
		"status": "status", "статус": "status",
		"type": "case_type", "тип": "case_type",
	}

	for _, header := range headers {
		lower := strings.ToLower(strings.TrimSpace(header))
		if field, ok := knownFields[lower]; ok {
			mapping[header] = field
		}
	}

	return mapping
}

// ---------------------------------------------------------------------------
// Парсеры
// ---------------------------------------------------------------------------

// ImportFromCSV парсит CSV-файл и возвращает список входных данных для создания тест-кейсов.
func ImportFromCSV(reader io.Reader, mapping FieldMapping) ([]CreateTestCaseInput, error) {
	csvReader := csv.NewReader(reader)
	records, err := csvReader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения CSV: %w", err)
	}

	if len(records) < 2 {
		return nil, fmt.Errorf("CSV файл пуст или содержит только заголовки")
	}

	if len(records)-1 > maxImportRecords {
		return nil, fmt.Errorf("слишком много записей для импорта (макс. %d)", maxImportRecords)
	}

	headers := records[0]
	if len(mapping) == 0 {
		mapping = AutoDetectMapping(headers)
	}

	// Строим индекс: номер колонки -> поле TMS
	colToField := make(map[int]string)
	for i, header := range headers {
		if field, ok := mapping[header]; ok {
			colToField[i] = field
		}
	}

	var inputs []CreateTestCaseInput
	for _, row := range records[1:] {
		input := CreateTestCaseInput{}
		for i, val := range row {
			field, ok := colToField[i]
			if !ok {
				continue
			}
			applyFieldToInput(&input, field, val)
		}
		inputs = append(inputs, input)
	}

	return inputs, nil
}

// jsonTestCase — структура для JSON-импорта тест-кейсов.
type jsonTestCase struct {
	Title        string `json:"title"`
	Description  string `json:"description"`
	Precondition string `json:"precondition"`
	Priority     string `json:"priority"`
	Status       string `json:"status"`
	CaseType     string `json:"case_type"`
	Steps        string `json:"steps"`
}

// ImportFromJSON парсит JSON-массив и возвращает входные данные для создания тест-кейсов.
func ImportFromJSON(reader io.Reader) ([]CreateTestCaseInput, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxImportFileBytes+1))
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения JSON: %w", err)
	}
	if len(data) > maxImportFileBytes {
		return nil, fmt.Errorf("файл слишком большой")
	}
	if err := assertJSONMaxDepth(data, maxImportNestDepth); err != nil {
		return nil, err
	}

	var items []jsonTestCase
	if err := json.Unmarshal(data, &items); err != nil {
		return nil, fmt.Errorf("ошибка разбора JSON: %w", err)
	}
	if len(items) > maxImportRecords {
		return nil, fmt.Errorf("слишком много записей для импорта (макс. %d)", maxImportRecords)
	}

	inputs := make([]CreateTestCaseInput, 0, len(items))
	for _, item := range items {
		input := CreateTestCaseInput{
			Title:        item.Title,
			Description:  item.Description,
			Precondition: item.Precondition,
			Priority:     item.Priority,
			Status:       item.Status,
			CaseType:     item.CaseType,
			Steps:        parseTestRailSteps(item.Steps, ""),
		}
		inputs = append(inputs, input)
	}

	return inputs, nil
}

// xmlTestCases — корневой XML-элемент для импорта.
type xmlTestCases struct {
	XMLName   xml.Name      `xml:"testcases"`
	TestCases []xmlTestCase `xml:"testcase"`
}

// xmlTestCase — элемент тест-кейса в XML.
type xmlTestCase struct {
	Title        string `xml:"title"`
	Description  string `xml:"description"`
	Precondition string `xml:"precondition"`
	Priority     string `xml:"priority"`
	Status       string `xml:"status"`
	CaseType     string `xml:"case_type"`
}

// ImportFromXML парсит XML с корневым элементом <testcases> и возвращает входные данные.
func ImportFromXML(reader io.Reader) ([]CreateTestCaseInput, error) {
	data, err := io.ReadAll(io.LimitReader(reader, maxImportFileBytes+1))
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения XML: %w", err)
	}
	if len(data) > maxImportFileBytes {
		return nil, fmt.Errorf("файл слишком большой")
	}
	if err := assertXMLMaxDepth(data, maxImportNestDepth); err != nil {
		return nil, err
	}

	var root xmlTestCases
	if err := xml.Unmarshal(data, &root); err != nil {
		return nil, fmt.Errorf("ошибка разбора XML: %w", err)
	}
	if len(root.TestCases) > maxImportRecords {
		return nil, fmt.Errorf("слишком много записей для импорта (макс. %d)", maxImportRecords)
	}

	inputs := make([]CreateTestCaseInput, 0, len(root.TestCases))
	for _, item := range root.TestCases {
		input := CreateTestCaseInput{
			Title:        item.Title,
			Description:  item.Description,
			Precondition: item.Precondition,
			Priority:     item.Priority,
			Status:       item.Status,
			CaseType:     item.CaseType,
		}
		inputs = append(inputs, input)
	}

	return inputs, nil
}

// ImportFromExcel парсит XLSX-файл и возвращает входные данные для создания тест-кейсов.
// Использует библиотеку excelize для чтения Excel.
// Первая строка — заголовки. По умолчанию читает первый лист.
func ImportFromExcel(reader io.Reader, sheetName string, mapping FieldMapping) ([]CreateTestCaseInput, error) {
	f, err := excelize.OpenReader(reader)
	if err != nil {
		return nil, fmt.Errorf("ошибка открытия Excel: %w", err)
	}
	defer f.Close()

	// Определяем лист
	sheet := sheetName
	if sheet == "" {
		sheet = f.GetSheetName(0)
	}

	rows, err := f.GetRows(sheet)
	if err != nil {
		return nil, fmt.Errorf("ошибка чтения листа %s: %w", sheet, err)
	}

	if len(rows) < 2 {
		return nil, fmt.Errorf("Excel файл пуст или содержит только заголовки")
	}

	if len(rows)-1 > maxImportRecords {
		return nil, fmt.Errorf("слишком много записей для импорта (макс. %d)", maxImportRecords)
	}

	headers := rows[0]
	if len(mapping) == 0 {
		mapping = AutoDetectMapping(headers)
	}

	// Строим индекс: номер колонки -> поле TMS
	colToField := make(map[int]string)
	for i, header := range headers {
		if field, ok := mapping[header]; ok {
			colToField[i] = field
		}
	}

	var inputs []CreateTestCaseInput
	for _, row := range rows[1:] {
		input := CreateTestCaseInput{}
		for i, val := range row {
			field, ok := colToField[i]
			if !ok {
				continue
			}
			applyFieldToInput(&input, field, val)
		}
		inputs = append(inputs, input)
	}

	return inputs, nil
}

// ValidateImportData проверяет корректность входных данных импорта.
// Возвращает список ошибок с номерами строк.
func ValidateImportData(inputs []CreateTestCaseInput) []ImportValidationError {
	var errors []ImportValidationError
	for i, input := range inputs {
		row := i + 2 // +1 для 0-index, +1 для заголовка
		if input.Title == "" {
			errors = append(errors, ImportValidationError{
				Row:     row,
				Field:   "title",
				Message: "название обязательно",
			})
		}
		if input.Priority != "" {
			validPriorities := map[string]bool{
				"critical": true, "high": true, "medium": true, "low": true, "trivial": true,
			}
			if !validPriorities[input.Priority] {
				errors = append(errors, ImportValidationError{
					Row:     row,
					Field:   "priority",
					Message: "недопустимый приоритет: " + input.Priority,
				})
			}
		}
	}
	return errors
}

// ---------------------------------------------------------------------------
// Обработка импорта
// ---------------------------------------------------------------------------

// detectImportFormat — case-insensitive определение формата по имени файла (P3-BE-09 / W2-A4-012).
func detectImportFormat(filename string) string {
	name := strings.ToLower(filename)
	switch {
	case strings.HasSuffix(name, ".csv"):
		return "csv"
	case strings.HasSuffix(name, ".json"):
		return "json"
	case strings.HasSuffix(name, ".xml"):
		return "xml"
	case strings.HasSuffix(name, ".xlsx"):
		return "excel"
	default:
		return ""
	}
}

// handleImport обрабатывает HTTP-запрос импорта тест-кейсов.
func handleImport(svc *TMSService, c fiber.Ctx, projectID, userID uuid.UUID) (*ImportResult, error) {
	file, err := c.FormFile("file")
	if err != nil {
		return nil, apperrors.NewBadRequest("файл обязателен")
	}

	format := c.FormValue("format", "")
	if format == "" {
		// Автоопределение формата по расширению (case-insensitive — P3-BE-09 / W2-A4-012).
		format = detectImportFormat(file.Filename)
		if format == "" {
			return nil, apperrors.NewBadRequest("не удалось определить формат файла")
		}
	}

	f, err := file.Open()
	if err != nil {
		return nil, apperrors.NewInternal("ошибка открытия файла")
	}
	defer f.Close()

	// TASK-016 / P1-BE-13: LimitReader + явный 413 при превышении лимита (не полагаться только на Fiber BodyLimit).
	data, err := io.ReadAll(io.LimitReader(f, maxImportFileBytes+1))
	if err != nil {
		return nil, apperrors.NewInternal("ошибка чтения файла")
	}
	if len(data) > maxImportFileBytes {
		return nil, fiber.NewError(fiber.StatusRequestEntityTooLarge, "файл слишком большой")
	}
	limited := bytes.NewReader(data)

	// Маппинг из формы
	mappingStr := c.FormValue("mapping", "")
	var mapping FieldMapping
	if mappingStr != "" {
		if err := json.Unmarshal([]byte(mappingStr), &mapping); err != nil {
			return nil, apperrors.NewBadRequest("некорректный JSON маппинга")
		}
	}

	var inputs []CreateTestCaseInput
	switch format {
	case "csv":
		inputs, err = ImportFromCSV(limited, mapping)
	case "json":
		inputs, err = ImportFromJSON(limited)
	case "xml":
		inputs, err = ImportFromXML(limited)
	case "excel":
		inputs, err = ImportFromExcel(limited, "", mapping)
	default:
		return nil, apperrors.NewBadRequest("неподдерживаемый формат: " + format)
	}
	if err != nil {
		msg := err.Error()
		if strings.Contains(msg, "файл слишком большой") {
			return nil, fiber.NewError(fiber.StatusRequestEntityTooLarge, "файл слишком большой")
		}
		// A3W2-4-003: фиксированное сообщение клиенту; internals только в slog (как A2W2-4-007).
		slog.Error("ошибка парсинга файла импорта TMS", "error", err, "format", format)
		return nil, apperrors.NewBadRequest("некорректный файл импорта")
	}

	// Валидация
	validationErrors := ValidateImportData(inputs)

	result := &ImportResult{
		ValidationErrors: validationErrors,
	}

	// Создаём тест-кейсы для валидных записей
	validationMap := make(map[int]bool)
	for _, ve := range validationErrors {
		validationMap[ve.Row] = true
	}

	ctx := c.Context()
	for i, input := range inputs {
		row := i + 2
		if validationMap[row] {
			result.Skipped++
			result.Errors++
			continue
		}
		input.CreatedByID = userID
		_, err := svc.repo.CreateTestCase(ctx, projectID, input)
		if err != nil {
			result.Errors++
			continue
		}
		result.Created++
	}

	// A2W2-4-004: явный partial — частичный импорт без скрытого «успеха».
	result.Partial = result.Errors > 0 || result.Skipped > 0

	return result, nil
}

// applyFieldToInput устанавливает значение поля во входных данных тест-кейса.
func applyFieldToInput(input *CreateTestCaseInput, field, value string) {
	switch field {
	case "title":
		input.Title = value
	case "description":
		input.Description = value
	case "precondition":
		input.Precondition = value
	case "steps":
		// TASK-016 / P1-BE-14: колонка Steps/Шаги больше не отбрасывается молча.
		input.Steps = parseTestRailSteps(value, "")
	case "priority":
		input.Priority = strings.ToLower(value)
	case "status":
		input.Status = strings.ToLower(value)
	case "case_type":
		input.CaseType = strings.ToLower(value)
	}
}

// assertJSONMaxDepth отклоняет JSON с вложенностью выше maxDepth (XML/JSON bomb).
func assertJSONMaxDepth(data []byte, maxDepth int) error {
	dec := json.NewDecoder(bytes.NewReader(data))
	depth := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("ошибка разбора JSON: %w", err)
		}
		switch tok {
		case json.Delim('{'), json.Delim('['):
			depth++
			if depth > maxDepth {
				return fmt.Errorf("превышена максимальная глубина JSON (%d)", maxDepth)
			}
		case json.Delim('}'), json.Delim(']'):
			depth--
		}
	}
}

// assertXMLMaxDepth отклоняет XML с вложенностью выше maxDepth.
func assertXMLMaxDepth(data []byte, maxDepth int) error {
	dec := xml.NewDecoder(bytes.NewReader(data))
	depth := 0
	for {
		tok, err := dec.Token()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return fmt.Errorf("ошибка разбора XML: %w", err)
		}
		switch tok.(type) {
		case xml.StartElement:
			depth++
			if depth > maxDepth {
				return fmt.Errorf("превышена максимальная глубина XML (%d)", maxDepth)
			}
		case xml.EndElement:
			depth--
		}
	}
}

// ---------------------------------------------------------------------------
// Экспорт данных для тест-кейсов
// ---------------------------------------------------------------------------

// TestCaseExportData содержит данные тест-кейса для экспорта.
type TestCaseExportData struct {
	Key          string
	Title        string
	Type         string
	Priority     string
	Status       string
	Steps        string
	Precondition string
	Suite        string
}

// formatTestStepsForExport сериализует шаги в plain text для колонки Steps.
// Формат совместим с parseTestRailSteps / ImportFromCSV (A3W2-4-001).
func formatTestStepsForExport(steps []schema.TestStep) string {
	if len(steps) == 0 {
		return ""
	}
	var b strings.Builder
	for i, step := range steps {
		if i > 0 {
			b.WriteByte('\n')
		}
		order := step.Order
		if order <= 0 {
			order = i + 1
		}
		action := strings.TrimSpace(step.Action)
		if kw := strings.TrimSpace(step.Keyword); kw != "" {
			action = strings.TrimSpace(kw + " " + action)
		}
		fmt.Fprintf(&b, "%d. %s", order, action)
	}
	return b.String()
}

// TestRunExportData содержит данные прогона для экспорта.
type TestRunExportData struct {
	Title     string
	CreatedAt string
	Total     int
	Passed    int
	Failed    int
	Skipped   int
	Broken    int
	PassRate  float64
	Results   []ResultExportData
}

// ResultExportData содержит данные результата для экспорта.
type ResultExportData struct {
	CaseKey  string
	Title    string
	Status   string
	Duration string
	Comment  string
}

// handleExportTestCases экспортирует тест-кейсы проекта.
// A2W2-4-005: пагинация до конца (ListTestCases cap 100) — без тихой усечённой выгрузки.
func handleExportTestCases(svc *TMSService, ctx context.Context, projectID uuid.UUID, format string) (io.Reader, string, error) {
	const pageSize = 100
	offset := 0
	var all []*ent.TestCase
	for {
		items, total, err := svc.repo.ListTestCases(ctx, projectID, ListTestCasesParams{Offset: offset, Limit: pageSize})
		if err != nil {
			return nil, "", apperrors.NewInternal("ошибка получения тест-кейсов для экспорта")
		}
		all = append(all, items...)
		offset += len(items)
		if offset >= total || len(items) == 0 {
			break
		}
	}

	exportData := make([]TestCaseExportData, 0, len(all))
	for _, tc := range all {
		suite := ""
		if tc.Edges.Suite != nil {
			suite = tc.Edges.Suite.Name
		}
		exportData = append(exportData, TestCaseExportData{
			Key:          tc.CaseKey,
			Title:        tc.Title,
			Type:         string(tc.CaseType),
			Priority:     string(tc.Priority),
			Status:       string(tc.Status),
			// A3W2-4-001: plain text (нумерованные строки) — round-trip через CSV Steps / parseTestRailSteps.
			Steps:        formatTestStepsForExport(tc.Steps),
			Precondition: tc.Precondition,
			Suite:        suite,
		})
	}

	var buf bytes.Buffer
	var filename string
	var err error

	switch format {
	case "csv":
		filename = "test-cases.csv"
		err = ExportTestCasesToCSV(&buf, exportData)
	case "json":
		filename = "test-cases.json"
		err = ExportTestCasesToJSON(&buf, exportData)
	case "xml":
		filename = "test-cases.xml"
		err = ExportTestCasesToXML(&buf, exportData)
	case "excel":
		filename = "test-cases.xlsx"
		err = ExportTestCasesToExcel(&buf, exportData)
	default:
		filename = "test-cases.csv"
		err = ExportTestCasesToCSV(&buf, exportData)
	}
	if err != nil {
		return nil, "", apperrors.NewInternal("ошибка экспорта: " + err.Error())
	}

	return bytes.NewReader(buf.Bytes()), filename, nil
}

// handleExportTestRunReport экспортирует отчёт прогона (W4-run-export / A3-012).
func handleExportTestRunReport(svc *TMSService, ctx context.Context, projectID, runID uuid.UUID, format string) (io.Reader, string, error) {
	run, err := svc.repo.GetTestRunWithResultsInProject(ctx, projectID, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, "", apperrors.NewNotFound("прогон не найден")
		}
		return nil, "", apperrors.NewInternal("ошибка получения прогона для экспорта")
	}

	passRate := float64(0)
	if run.TotalCount > 0 {
		passRate = float64(run.PassedCount) / float64(run.TotalCount) * 100
	}

	exportData := TestRunExportData{
		Title:     run.Title,
		CreatedAt: run.CreatedAt.Format("2006-01-02 15:04:05"),
		Total:     run.TotalCount,
		Passed:    run.PassedCount,
		Failed:    run.FailedCount,
		Skipped:   run.SkippedCount,
		Broken:    run.BrokenCount,
		PassRate:  passRate,
		Results:   make([]ResultExportData, 0, len(run.Edges.Results)),
	}

	for _, r := range run.Edges.Results {
		caseKey := ""
		title := ""
		if r.Edges.TestCase != nil {
			caseKey = r.Edges.TestCase.CaseKey
			title = r.Edges.TestCase.Title
		}
		duration := ""
		if r.DurationMs != nil {
			duration = fmt.Sprintf("%dms", *r.DurationMs)
		}
		exportData.Results = append(exportData.Results, ResultExportData{
			CaseKey:  caseKey,
			Title:    title,
			Status:   string(r.Status),
			Duration: duration,
			Comment:  r.Comment,
		})
	}

	var buf bytes.Buffer
	var filename string

	switch format {
	case "json":
		filename = "run-report.json"
		err = json.NewEncoder(&buf).Encode(exportData)
	case "csv":
		filename = "run-report.csv"
		err = ExportRunResultsToCSV(&buf, exportData)
	case "pdf":
		filename = "run-report.pdf"
		err = ExportTestRunToPDF(&buf, exportData)
	default:
		filename = "run-report.csv"
		err = ExportRunResultsToCSV(&buf, exportData)
	}
	if err != nil {
		return nil, "", apperrors.NewInternal("ошибка экспорта отчёта: " + err.Error())
	}

	return bytes.NewReader(buf.Bytes()), filename, nil
}
