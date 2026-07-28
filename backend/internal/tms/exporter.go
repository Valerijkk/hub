package tms

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/col"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/core"
	"github.com/johnfercher/maroto/v2/pkg/props"
	"github.com/xuri/excelize/v2"
)

// ---------------------------------------------------------------------------
// Экспорт тест-кейсов
// ---------------------------------------------------------------------------

// ExportTestCasesToCSV записывает тест-кейсы в CSV-формат.
// Колонки: Key, Title, Type, Priority, Status, Steps, Precondition, Suite.
func ExportTestCasesToCSV(writer io.Writer, cases []TestCaseExportData) error {
	w := csv.NewWriter(writer)
	defer w.Flush()

	if err := w.Write([]string{"Key", "Title", "Type", "Priority", "Status", "Steps", "Precondition", "Suite"}); err != nil {
		return fmt.Errorf("ошибка записи заголовков CSV: %w", err)
	}

	for _, tc := range cases {
		if err := w.Write([]string{
			tc.Key, tc.Title, tc.Type, tc.Priority, tc.Status, tc.Steps, tc.Precondition, tc.Suite,
		}); err != nil {
			return fmt.Errorf("ошибка записи строки CSV: %w", err)
		}
	}

	return nil
}

// ExportTestCasesToJSON записывает тест-кейсы в JSON-формат (массив объектов).
func ExportTestCasesToJSON(writer io.Writer, cases []TestCaseExportData) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")
	return encoder.Encode(cases)
}

// xmlExportTestCases — корневой XML-элемент для экспорта.
type xmlExportTestCases struct {
	XMLName   xml.Name            `xml:"testcases"`
	TestCases []xmlExportTestCase `xml:"testcase"`
}

// xmlExportTestCase — элемент тест-кейса в XML-экспорте.
type xmlExportTestCase struct {
	Key          string `xml:"key"`
	Title        string `xml:"title"`
	Type         string `xml:"type"`
	Priority     string `xml:"priority"`
	Status       string `xml:"status"`
	Steps        string `xml:"steps,omitempty"`
	Precondition string `xml:"precondition,omitempty"`
	Suite        string `xml:"suite,omitempty"`
}

// ExportTestCasesToXML записывает тест-кейсы в XML-формат с корневым элементом <testcases>.
func ExportTestCasesToXML(writer io.Writer, cases []TestCaseExportData) error {
	root := xmlExportTestCases{
		TestCases: make([]xmlExportTestCase, 0, len(cases)),
	}

	for _, tc := range cases {
		root.TestCases = append(root.TestCases, xmlExportTestCase{
			Key:          tc.Key,
			Title:        tc.Title,
			Type:         tc.Type,
			Priority:     tc.Priority,
			Status:       tc.Status,
			Steps:        tc.Steps,
			Precondition: tc.Precondition,
			Suite:        tc.Suite,
		})
	}

	if _, err := writer.Write([]byte(xml.Header)); err != nil {
		return fmt.Errorf("ошибка записи XML-заголовка: %w", err)
	}

	encoder := xml.NewEncoder(writer)
	encoder.Indent("", "  ")
	return encoder.Encode(root)
}

// ExportTestCasesToExcel записывает тест-кейсы в XLSX-формат.
// Жирные заголовки, автоширина колонок, чередующийся цвет строк.
func ExportTestCasesToExcel(writer io.Writer, cases []TestCaseExportData) error {
	f := excelize.NewFile()
	defer f.Close()

	sheet := "Test Cases"
	idx, _ := f.NewSheet(sheet)
	f.SetActiveSheet(idx)
	_ = f.DeleteSheet("Sheet1")

	// Стиль заголовков
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true, Size: 11},
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"E2EFDA"}},
	})

	// Стиль чётных строк
	evenStyle, _ := f.NewStyle(&excelize.Style{
		Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"F5F5F5"}},
	})

	headers := []string{"Key", "Title", "Type", "Priority", "Status", "Steps", "Precondition", "Suite"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		_ = f.SetCellValue(sheet, cell, h)
		_ = f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	for rowIdx, tc := range cases {
		row := rowIdx + 2
		values := []string{tc.Key, tc.Title, tc.Type, tc.Priority, tc.Status, tc.Steps, tc.Precondition, tc.Suite}
		for colIdx, val := range values {
			cell, _ := excelize.CoordinatesToCellName(colIdx+1, row)
			_ = f.SetCellValue(sheet, cell, val)
			if rowIdx%2 == 1 {
				_ = f.SetCellStyle(sheet, cell, cell, evenStyle)
			}
		}
	}

	// Автоширина
	colWidths := []float64{12, 40, 10, 10, 12, 30, 30, 20}
	for i, w := range colWidths {
		colName, _ := excelize.ColumnNumberToName(i + 1)
		_ = f.SetColWidth(sheet, colName, colName, w)
	}

	return f.Write(writer)
}

// ---------------------------------------------------------------------------
// Экспорт результатов прогонов
// ---------------------------------------------------------------------------

// ExportRunResultsToCSV записывает результаты прогона в CSV-формат.
func ExportRunResultsToCSV(writer io.Writer, data TestRunExportData) error {
	w := csv.NewWriter(writer)

	headerRows := [][]string{
		{"Run Title", data.Title},
		{"Created At", data.CreatedAt},
		{"Total", fmt.Sprintf("%d", data.Total)},
		{"Passed", fmt.Sprintf("%d", data.Passed)},
		{"Failed", fmt.Sprintf("%d", data.Failed)},
		{"Pass Rate", fmt.Sprintf("%.1f%%", data.PassRate)},
		{""},
		{"Case Key", "Title", "Status", "Duration", "Comment"},
	}
	for _, row := range headerRows {
		if err := w.Write(row); err != nil {
			return fmt.Errorf("ошибка записи строки CSV: %w", err)
		}
	}

	for _, r := range data.Results {
		if err := w.Write([]string{r.CaseKey, r.Title, r.Status, r.Duration, r.Comment}); err != nil {
			return fmt.Errorf("ошибка записи строки CSV: %w", err)
		}
	}

	// encoding/csv буферизует; ошибка io.Writer всплывает на Flush (W2-A4-013 / P3-BE-01).
	w.Flush()
	if err := w.Error(); err != nil {
		return fmt.Errorf("ошибка записи CSV: %w", err)
	}
	return nil
}

// ExportTestRunToPDF генерирует PDF-отчёт прогона тестов.
// Включает заголовок, таблицу статистики, таблицу результатов.
func ExportTestRunToPDF(writer io.Writer, data TestRunExportData) error {
	m := maroto.New()

	// Заголовок
	m.AddRow(15,
		col.New(12).Add(
			text.New(data.Title, props.Text{
				Size:  16,
				Style: fontstyle.Bold,
				Align: align.Center,
			}),
		),
	)

	// Дата и статистика
	m.AddRow(8,
		col.New(12).Add(
			text.New(fmt.Sprintf("Создан: %s", data.CreatedAt), props.Text{
				Size:  9,
				Align: align.Center,
				Color: &props.Color{Red: 100, Green: 100, Blue: 100},
			}),
		),
	)

	m.AddRow(5)

	// Таблица статистики
	statsHeaders := []core.Col{
		text.NewCol(2, "Всего", props.Text{Style: fontstyle.Bold, Size: 9}),
		text.NewCol(2, "Passed", props.Text{Style: fontstyle.Bold, Size: 9, Color: &props.Color{Green: 150}}),
		text.NewCol(2, "Failed", props.Text{Style: fontstyle.Bold, Size: 9, Color: &props.Color{Red: 200}}),
		text.NewCol(2, "Skipped", props.Text{Style: fontstyle.Bold, Size: 9}),
		text.NewCol(2, "Broken", props.Text{Style: fontstyle.Bold, Size: 9, Color: &props.Color{Red: 200, Green: 150}}),
		text.NewCol(2, "Pass Rate", props.Text{Style: fontstyle.Bold, Size: 9}),
	}
	m.AddRow(10, statsHeaders...)

	statsValues := []core.Col{
		text.NewCol(2, fmt.Sprintf("%d", data.Total), props.Text{Size: 10}),
		text.NewCol(2, fmt.Sprintf("%d", data.Passed), props.Text{Size: 10, Color: &props.Color{Green: 150}}),
		text.NewCol(2, fmt.Sprintf("%d", data.Failed), props.Text{Size: 10, Color: &props.Color{Red: 200}}),
		text.NewCol(2, fmt.Sprintf("%d", data.Skipped), props.Text{Size: 10}),
		text.NewCol(2, fmt.Sprintf("%d", data.Broken), props.Text{Size: 10, Color: &props.Color{Red: 200, Green: 150}}),
		text.NewCol(2, fmt.Sprintf("%.1f%%", data.PassRate), props.Text{Size: 10}),
	}
	m.AddRow(10, statsValues...)

	m.AddRow(8)

	// Заголовки таблицы результатов
	m.AddRow(8,
		text.NewCol(2, "Ключ", props.Text{Style: fontstyle.Bold, Size: 8}),
		text.NewCol(4, "Название", props.Text{Style: fontstyle.Bold, Size: 8}),
		text.NewCol(2, "Статус", props.Text{Style: fontstyle.Bold, Size: 8}),
		text.NewCol(2, "Время", props.Text{Style: fontstyle.Bold, Size: 8}),
		text.NewCol(2, "Коммент.", props.Text{Style: fontstyle.Bold, Size: 8}),
	)

	// Строки результатов
	for _, r := range data.Results {
		statusColor := &props.Color{}
		switch r.Status {
		case "passed":
			statusColor = &props.Color{Green: 150}
		case "failed":
			statusColor = &props.Color{Red: 200}
		case "broken":
			statusColor = &props.Color{Red: 200, Green: 150}
		}

		comment := r.Comment
		if len(comment) > 30 {
			comment = comment[:30] + "..."
		}

		m.AddRow(7,
			text.NewCol(2, r.CaseKey, props.Text{Size: 7}),
			text.NewCol(4, truncate(r.Title, 40), props.Text{Size: 7}),
			text.NewCol(2, r.Status, props.Text{Size: 7, Color: statusColor}),
			text.NewCol(2, r.Duration, props.Text{Size: 7}),
			text.NewCol(2, comment, props.Text{Size: 7}),
		)
	}

	doc, err := m.Generate()
	if err != nil {
		return fmt.Errorf("ошибка генерации PDF: %w", err)
	}

	buf := doc.GetBytes()
	_, err = writer.Write(buf)
	return err
}

// truncate обрезает строку до указанной длины.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen]) + "..."
}

// Гарантируем использование импортов maroto
var (
	_ = row.New
)
