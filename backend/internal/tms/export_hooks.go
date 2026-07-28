//go:build hubtest

package tms

import "github.com/google/uuid"

// ExportJUnitTestCaseToAllure конвертирует JUnit testcase в AllureResult для hubtest.
func ExportJUnitTestCaseToAllure(suiteName string, tc JUnitTestCase) AllureResult {
	return junitTestCaseToAllure(suiteName, tc)
}
func ExportParseAllureResults(data []byte) ([]AllureResult, error) {
	return ParseAllureResults(data)
}

// ExportParseAllureResult парсит один JSON-объект Allure для hubtest.
func ExportParseAllureResult(data []byte) (*AllureResult, error) {
	return ParseAllureResult(data)
}

// ExportExtractLabelValue извлекает значение метки Allure по имени для hubtest.
func ExportExtractLabelValue(labels []AllureLabel, name string) string {
	return extractLabelValue(labels, name)
}

// ExportAllureStatusToTMS маппит статус Allure в статус TMS для hubtest.
func ExportAllureStatusToTMS(status string) string {
	return allureStatusToTMS(status)
}

// ExportParseUUIDs парсит строковые UUID для hubtest.
func ExportParseUUIDs(strs []string) ([]uuid.UUID, error) {
	return parseUUIDs(strs)
}

// ExportDetectImportFormat — case-insensitive format detect (P3-BE-09).
func ExportDetectImportFormat(filename string) string {
	return detectImportFormat(filename)
}
