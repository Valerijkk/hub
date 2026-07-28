//go:build hubtest

package imex

import (
	"github.com/johnfercher/maroto/v2/pkg/core"
)

func ExportBuildFieldMap(mappings []FieldMapping, headers []string) map[string]int {
	return buildFieldMap(mappings, headers)
}

func ExportGetField(record []string, fieldMap map[string]int, field string) string {
	return getField(record, fieldMap, field)
}

func ExportMapJiraIssueType(jiraType string) string {
	return mapJiraIssueType(jiraType)
}

func ExportMapJiraPriority(jiraPriority string) string {
	return mapJiraPriority(jiraPriority)
}

func ExportJiraDescriptionToText(desc interface{}) string {
	return jiraDescriptionToText(desc)
}

func ExportParseJiraJSON(data []byte) ([]JiraIssue, error) {
	return parseJiraJSON(data)
}

func ExportBuildHeaderRow(title string) core.Row {
	return buildHeaderRow(title)
}

func ExportBuildInfoRow(label, value string) core.Row {
	return buildInfoRow(label, value)
}

func ExportBuildTableHeader(headers []string) core.Row {
	return buildTableHeader(headers)
}

func ExportBuildTableRow(values []string) core.Row {
	return buildTableRow(values)
}

// ExportTruncateCell — P3-BE-04 / TASK-060: rune-safe PDF cell truncate для hubtest.
func ExportTruncateCell(s string, maxRunes int) string {
	return truncateCell(s, maxRunes)
}
