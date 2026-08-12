//go:build hubtest

package tms

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/schema"
)

// ExportMapTestRailPriority маппит приоритет TestRail для hubtest.
func ExportMapTestRailPriority(trPriority string) string {
	return mapTestRailPriority(trPriority)
}
func ExportParseTestRailCSV(data []byte) ([]TestRailCase, error) {
	return parseTestRailCSV(data)
}

// ExportTryReadCSV читает CSV с разделителем для hubtest.
func ExportTryReadCSV(data []byte, sep rune) ([][]string, error) {
	return tryReadCSV(data, sep)
}

// ExportParseTestRailXML парсит XML TestRail для hubtest.
func ExportParseTestRailXML(data []byte) ([]TestRailCase, error) {
	return parseTestRailXML(data)
}

// ExportImportTestRailCases импортирует кейсы TestRail для hubtest.
func ExportImportTestRailCases(s *TMSService, ctx context.Context, projectID, userID uuid.UUID, cases []TestRailCase) (*ImportResult, error) {
	return s.importTestRailCases(ctx, projectID, userID, cases)
}

// ExportMapTestRailType маппит тип TestRail для hubtest.
func ExportMapTestRailType(trType string) string {
	return mapTestRailType(trType)
}

// ExportStripHTML удаляет HTML-теги для hubtest.
func ExportStripHTML(s string) string {
	return stripHTML(s)
}

// ExportParseTestRailSteps парсит шаги TestRail для hubtest.
func ExportParseTestRailSteps(stepsText, expectedText string) []schema.TestStep {
	return parseTestRailSteps(stepsText, expectedText)
}

// ExportCSVField читает поле CSV-строки для hubtest.
func ExportCSVField(row []string, colMap map[string]int, field string) string {
	return csvField(row, colMap, field)
}

// ExportTrimStepNumbering убирает нумерацию шага для hubtest.
func ExportTrimStepNumbering(s string) string {
	return trimStepNumbering(s)
}

// ExportTruncate обрезает строку для hubtest.
func ExportTruncate(s string, maxLen int) string {
	return truncate(s, maxLen)
}

// ExportFormatTestStepsForExport сериализует шаги в plain text (A3W2-4-001).
func ExportFormatTestStepsForExport(steps []schema.TestStep) string {
	return formatTestStepsForExport(steps)
}
