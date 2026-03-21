package tms

import (
	"encoding/csv"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
)

// ---------------------------------------------------------------------------
// Экспорт тест-кейсов
// ---------------------------------------------------------------------------

// ExportTestCasesToCSV записывает тест-кейсы в CSV-формат.
// Колонки: Key, Title, Type, Priority, Status, Steps, Precondition, Suite.
func ExportTestCasesToCSV(writer io.Writer, cases []TestCaseExportData) error {
	w := csv.NewWriter(writer)
	defer w.Flush()

	// Заголовки
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
	XMLName   xml.Name             `xml:"testcases"`
	TestCases []xmlExportTestCase  `xml:"testcase"`
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

// ---------------------------------------------------------------------------
// Экспорт результатов прогонов
// ---------------------------------------------------------------------------

// ExportRunResultsToCSV записывает результаты прогона в CSV-формат.
func ExportRunResultsToCSV(writer io.Writer, data TestRunExportData) error {
	w := csv.NewWriter(writer)
	defer w.Flush()

	// Мета-информация
	_ = w.Write([]string{"Run Title", data.Title})
	_ = w.Write([]string{"Created At", data.CreatedAt})
	_ = w.Write([]string{"Total", fmt.Sprintf("%d", data.Total)})
	_ = w.Write([]string{"Passed", fmt.Sprintf("%d", data.Passed)})
	_ = w.Write([]string{"Failed", fmt.Sprintf("%d", data.Failed)})
	_ = w.Write([]string{"Pass Rate", fmt.Sprintf("%.1f%%", data.PassRate)})
	_ = w.Write([]string{""})

	// Заголовки результатов
	_ = w.Write([]string{"Case Key", "Title", "Status", "Duration", "Comment"})

	for _, r := range data.Results {
		if err := w.Write([]string{r.CaseKey, r.Title, r.Status, r.Duration, r.Comment}); err != nil {
			return fmt.Errorf("ошибка записи строки CSV: %w", err)
		}
	}

	return nil
}
