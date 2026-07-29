//go:build hubtest

package tms

import "github.com/valerii/hub/backend/ent"

// ExportToSuiteResponse конвертирует ent.TestSuite в SuiteResponse для hubtest.
func ExportToSuiteResponse(suite *ent.TestSuite) *SuiteResponse {
	return toSuiteResponse(suite)
}

// ExportToComponentResponse конвертирует ent.Component в ComponentResponse для hubtest.
func ExportToComponentResponse(c *ent.Component) *ComponentResponse {
	return toComponentResponse(c)
}

// ExportToTestRunResponse конвертирует ent.TestRun в TestRunResponse для hubtest.
func ExportToTestRunResponse(run *ent.TestRun) *TestRunResponse {
	return toTestRunResponse(run)
}

// ExportToTestResultResponse конвертирует ent.TestResult в TestResultResponse для hubtest.
func ExportToTestResultResponse(r *ent.TestResult) *TestResultResponse {
	return toTestResultResponse(r)
}

// ExportToTestCaseResponse конвертирует ent.TestCase в TestCaseResponse для hubtest.
func ExportToTestCaseResponse(tc *ent.TestCase) *TestCaseResponse {
	return toTestCaseResponse(tc)
}

// ExportToTestRunDetailResponse конвертирует ent.TestRun в TestRunDetailResponse для hubtest.
func ExportToTestRunDetailResponse(run *ent.TestRun) *TestRunDetailResponse {
	return toTestRunDetailResponse(run)
}

// ExportBuildSnapshot строит snapshot тест-кейса для hubtest.
func ExportBuildSnapshot(tc *ent.TestCase) (map[string]interface{}, error) {
	return buildSnapshot(tc)
}

// ExportNormalizeErrorPattern нормализует паттерн ошибки для hubtest.
func ExportNormalizeErrorPattern(msg string) string {
	return normalizeErrorPattern(msg)
}

// ExportSnapshotToUpdateInput конвертирует snapshot в UpdateTestCaseInput для hubtest.
func ExportSnapshotToUpdateInput(snapshot map[string]interface{}) (UpdateTestCaseInput, error) {
	return snapshotToUpdateInput(snapshot)
}

// ExportBuildSuiteTree строит дерево сьюитов для hubtest.
func ExportBuildSuiteTree(suites []*ent.TestSuite) []SuiteResponse {
	return buildSuiteTree(suites)
}

// ExportToTestCaseVersionResponse конвертирует ent.TestCaseVersion в TestCaseVersionResponse для hubtest.
func ExportToTestCaseVersionResponse(v *ent.TestCaseVersion) *TestCaseVersionResponse {
	return toTestCaseVersionResponse(v)
}
