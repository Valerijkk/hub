//go:build hubtest

package github

import (
	"context"
	"net/http"
	"regexp"

	gh "github.com/google/go-github/v68/github"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
)

// ExportToIntegrationResponse конвертирует ent.GitHubIntegration для hubtest.
func ExportToIntegrationResponse(ig *ent.GitHubIntegration) *IntegrationResponse {
	return toIntegrationResponse(ig)
}

// ExportToGitHubLinkResponse конвертирует ent.GitHubLink для hubtest.
func ExportToGitHubLinkResponse(l *ent.GitHubLink) *GitHubLinkResponse {
	return toGitHubLinkResponse(l)
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

// ExportMatchTestCaseFileByPath сопоставляет путь файла с тест-кейсом для hubtest.
func ExportMatchTestCaseFileByPath(className, fullName, filePath string) bool {
	return matchTestCaseFileByPath(className, fullName, filePath)
}

// ExportParseGitHubPRURL парсит URL pull request для hubtest.
func ExportParseGitHubPRURL(raw string) (owner, repo string, number int, err error) {
	return parseGitHubPRURL(raw)
}

// ExportToWorkflowRunResponse конвертирует GitHub API WorkflowRun для hubtest.
func ExportToWorkflowRunResponse(run *gh.WorkflowRun) WorkflowRunResponse {
	return toWorkflowRunResponse(run)
}

// ExportParseProjectID извлекает projectId из Fiber context для hubtest.
func ExportParseProjectID(c fiber.Ctx) (uuid.UUID, error) {
	return parseProjectID(c)
}

// ExportMaxWebhookBodyBytes возвращает лимит тела GitHub webhook для hubtest.
func ExportMaxWebhookBodyBytes() int {
	return maxWebhookBodyBytes
}

// ExportEntRepositoryEncryptField шифрует поле для hubtest.
func ExportEntRepositoryEncryptField(r *EntRepository, v string) (string, error) {
	return r.encryptField(v)
}

// ExportEntRepositoryDecryptField расшифровывает поле для hubtest.
func ExportEntRepositoryDecryptField(r *EntRepository, v string) string {
	return r.decryptField(v)
}

// ExportNewEntRepositoryForEncryptTest создаёт EntRepository только с cipher для hubtest.
func ExportNewEntRepositoryForEncryptTest(cipher *cryptopkg.Cipher) *EntRepository {
	return &EntRepository{cipher: cipher}
}

// ExportLinkPreviewResolveExternalPreview вызывает resolveExternalPreview для hubtest.
func ExportLinkPreviewResolveExternalPreview(s *LinkPreviewService, ctx context.Context, rawURL string) *LinkPreviewResponse {
	return s.resolveExternalPreview(ctx, rawURL)
}

// ExportGitHubServiceMarkTestCasesOutdated вызывает markTestCasesOutdated для hubtest.
func ExportGitHubServiceMarkTestCasesOutdated(s *GitHubService, ctx context.Context, changedFiles []string) {
	s.markTestCasesOutdated(ctx, changedFiles)
}

// ExportGitHubServiceHTTPClient возвращает HTTP-клиент сервиса для hubtest.
func ExportGitHubServiceHTTPClient(s *GitHubService) *http.Client {
	return s.httpClient
}
