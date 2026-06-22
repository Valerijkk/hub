//go:build hubtest

package gitlab

import (
	"regexp"

	"github.com/valerii/hub/backend/ent"
)

// ExportToIntegrationResponse конвертирует ent.GitLabIntegration для hubtest.
func ExportToIntegrationResponse(ig *ent.GitLabIntegration) *IntegrationResponse {
	return toIntegrationResponse(ig)
}

// ExportIssueKeyRegex возвращает regex ключей задач для hubtest.
func ExportIssueKeyRegex() *regexp.Regexp {
	return issueKeyRegex
}

// ExportSlugify нормализует имя ветки для hubtest.
func ExportSlugify(s string) string {
	return slugify(s)
}

// ExportTruncate обрезает строку для hubtest.
func ExportTruncate(s string, maxLen int) string {
	return truncate(s, maxLen)
}

// ExportParseGitLabMRURL парсит URL merge request для hubtest.
func ExportParseGitLabMRURL(raw string) (host, projectPath string, iid int, err error) {
	return parseGitLabMRURL(raw)
}

// ExportMaxWebhookBodyBytes возвращает лимит тела GitLab webhook для hubtest.
func ExportMaxWebhookBodyBytes() int {
	return maxWebhookBodyBytes
}
