// Package github реализует интеграцию с GitHub: репозитории, PR, пайплайны, вебхуки.
package github

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"unicode"

	gh "github.com/google/go-github/v68/github"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testcase"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// issueKeyRegex находит ключи задач вида ABC-123 с границами слов.
var issueKeyRegex = regexp.MustCompile(`\b([A-Z]+-\d+)\b`)

// GitHubService реализует бизнес-логику интеграции с GitHub.
// Обрабатывает webhook-события, автоматически связывает PR/коммиты с задачами,
// управляет интеграциями и создаёт ветки через GitHub API.
type GitHubService struct {
	repo   Repository
	nats   *natsclient.Client
	entDB  *ent.Client
}

// NewGitHubService создаёт новый сервис GitHub-интеграции.
// entDB используется для TMS sync_status обновлений при push-событиях.
func NewGitHubService(repo Repository, nats *natsclient.Client, entDB ...*ent.Client) *GitHubService {
	svc := &GitHubService{repo: repo, nats: nats}
	if len(entDB) > 0 {
		svc.entDB = entDB[0]
	}
	return svc
}

// ---------------------------------------------------------------------------
// Интеграции
// ---------------------------------------------------------------------------

// CreateIntegration создаёт интеграцию проекта с GitHub-репозиторием.
func (s *GitHubService) CreateIntegration(ctx context.Context, projectID uuid.UUID, req CreateIntegrationRequest) (*IntegrationResponse, error) {
	input := CreateIntegrationInput{
		RepoOwner:     req.RepoOwner,
		RepoName:      req.RepoName,
		AccessToken:   req.AccessToken,
		WebhookSecret: req.WebhookSecret,
	}

	if req.MergeStatusID != "" {
		id, err := uuid.Parse(req.MergeStatusID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный merge_status_id")
		}
		input.MergeStatusID = &id
	}

	integration, err := s.repo.CreateIntegration(ctx, projectID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания интеграции")
	}

	return toIntegrationResponse(integration), nil
}

// ListIntegrations возвращает все интеграции проекта.
func (s *GitHubService) ListIntegrations(ctx context.Context, projectID uuid.UUID) ([]IntegrationResponse, error) {
	integrations, err := s.repo.ListIntegrations(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения интеграций")
	}

	result := make([]IntegrationResponse, 0, len(integrations))
	for _, ig := range integrations {
		result = append(result, *toIntegrationResponse(ig))
	}
	return result, nil
}

// DeleteIntegration удаляет интеграцию по ID.
func (s *GitHubService) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteIntegration(ctx, id)
}

// ---------------------------------------------------------------------------
// Webhook обработчики
// ---------------------------------------------------------------------------

// HandleWebhook маршрутизирует webhook-событие к соответствующему обработчику.
func (s *GitHubService) HandleWebhook(ctx context.Context, eventType string, payload []byte) error {
	switch eventType {
	case "push":
		return s.handlePushWebhook(ctx, payload)
	case "pull_request":
		return s.handlePRWebhook(ctx, payload)
	default:
		slog.Debug("игнорируем webhook событие", "event", eventType)
		return nil
	}
}

// handlePushWebhook обрабатывает push-событие, привязывая коммиты к задачам (GHUB-02).
func (s *GitHubService) handlePushWebhook(ctx context.Context, payload []byte) error {
	var event gh.PushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return apperrors.NewBadRequest("некорректный payload push-события")
	}

	repoOwner := event.GetRepo().GetOwner().GetLogin()
	repoName := event.GetRepo().GetName()

	integration, err := s.repo.FindIntegrationByRepo(ctx, repoOwner, repoName)
	if err != nil {
		slog.Warn("интеграция не найдена для репозитория", "owner", repoOwner, "repo", repoName)
		return nil
	}

	// Сбор изменённых файлов для TMS sync_status
	var changedFiles []string

	for _, commit := range event.Commits {
		message := commit.GetMessage()
		keys := issueKeyRegex.FindAllString(message, -1)

		for _, key := range keys {
			issue, err := s.repo.FindIssueByKey(ctx, key)
			if err != nil {
				continue
			}

			_, err = s.repo.CreateGitHubLink(ctx, issue.ID, integration.ID, CreateGitHubLinkInput{
				LinkType:   "commit",
				ExternalID: commit.GetSHA(),
				Title:      truncate(message, 500),
				URL:        commit.GetURL(),
				Status:     "none",
				Author:     commit.GetAuthor().GetName(),
			})
			if err != nil {
				slog.Error("не удалось создать GitHub link для коммита",
					"commit", commit.GetSHA(), "issue", key, "error", err)
			}
		}

		// Собираем изменённые файлы для проверки TMS sync_status
		changedFiles = append(changedFiles, commit.Added...)
		changedFiles = append(changedFiles, commit.Modified...)
	}

	// Обновляем sync_status тест-кейсов, чьи файлы изменились
	s.markTestCasesOutdated(ctx, integration, changedFiles)

	return nil
}

// handlePRWebhook обрабатывает pull_request событие (GHUB-01, GHUB-03, GHUB-04).
func (s *GitHubService) handlePRWebhook(ctx context.Context, payload []byte) error {
	var event gh.PullRequestEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return apperrors.NewBadRequest("некорректный payload PR-события")
	}

	action := event.GetAction()
	pr := event.GetPullRequest()
	repoOwner := event.GetRepo().GetOwner().GetLogin()
	repoName := event.GetRepo().GetName()

	integration, err := s.repo.FindIntegrationByRepo(ctx, repoOwner, repoName)
	if err != nil {
		slog.Warn("интеграция не найдена", "owner", repoOwner, "repo", repoName)
		return nil
	}

	// Ищем ключи задач в title и body PR
	searchText := pr.GetTitle() + " " + pr.GetBody()
	keys := issueKeyRegex.FindAllString(searchText, -1)

	// Дедупликация ключей
	seen := make(map[string]bool)
	uniqueKeys := make([]string, 0, len(keys))
	for _, k := range keys {
		if !seen[k] {
			seen[k] = true
			uniqueKeys = append(uniqueKeys, k)
		}
	}

	prNumber := fmt.Sprintf("%d", pr.GetNumber())

	for _, key := range uniqueKeys {
		issue, err := s.repo.FindIssueByKey(ctx, key)
		if err != nil {
			continue
		}

		switch action {
		case "opened", "reopened":
			existing, findErr := s.repo.FindGitHubLink(ctx, prNumber, "pull_request", integration.ID)
			if findErr != nil {
				// Создаём новую связь
				_, err = s.repo.CreateGitHubLink(ctx, issue.ID, integration.ID, CreateGitHubLinkInput{
					LinkType:   "pull_request",
					ExternalID: prNumber,
					Title:      pr.GetTitle(),
					URL:        pr.GetHTMLURL(),
					Status:     "open",
					Author:     pr.GetUser().GetLogin(),
					Branch:     pr.GetHead().GetRef(),
				})
			} else {
				// Обновляем существующую
				err = s.repo.UpdateGitHubLinkStatus(ctx, existing.ID, "open")
			}

		case "closed":
			existing, findErr := s.repo.FindGitHubLink(ctx, prNumber, "pull_request", integration.ID)
			if findErr != nil {
				continue
			}

			if pr.GetMerged() {
				err = s.repo.UpdateGitHubLinkStatus(ctx, existing.ID, "merged")

				// GHUB-04: автосмена статуса при мерже PR.
				// Используем merge_status_id из интеграции, либо ищем статус категории "done".
				var targetStatusID *uuid.UUID
				if integration.MergeStatusID != nil {
					targetStatusID = integration.MergeStatusID
				} else if issue.Edges.Project != nil {
					doneStatus, findErr := s.repo.FindDoneStatus(ctx, issue.Edges.Project.ID)
					if findErr == nil && doneStatus != nil {
						targetStatusID = &doneStatus.ID
					}
				}

				if targetStatusID != nil {
					if updateErr := s.repo.UpdateIssueStatus(ctx, issue.ID, *targetStatusID); updateErr != nil {
						slog.Error("не удалось обновить статус задачи при мерже PR",
							"issue", key, "error", updateErr)
					} else {
						// Записываем активность автосмены статуса.
						_ = s.repo.LogActivity(ctx, issue.ID,
							fmt.Sprintf("Статус изменён автоматически (PR #%s merged)", prNumber))
					}
				}

				// Публикуем NATS-событие
				s.publishEvent("github.pr.merged", map[string]interface{}{
					"issue_key": key,
					"pr_number": prNumber,
					"pr_title":  pr.GetTitle(),
				})
			} else {
				err = s.repo.UpdateGitHubLinkStatus(ctx, existing.ID, "closed")
			}
		}

		if err != nil {
			slog.Error("ошибка обработки PR webhook", "action", action, "issue", key, "error", err)
		}
	}

	return nil
}

// ---------------------------------------------------------------------------
// Создание ветки (GHUB-05)
// ---------------------------------------------------------------------------

// CreateBranch создаёт ветку от default branch через GitHub API.
func (s *GitHubService) CreateBranch(ctx context.Context, issueKey string, req CreateBranchRequest) (*CreateBranchResponse, error) {
	issue, err := s.repo.FindIssueByKey(ctx, issueKey)
	if err != nil {
		return nil, apperrors.NewNotFound("задача не найдена")
	}

	integrationID, err := uuid.Parse(req.IntegrationID)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный integration_id")
	}

	integration, err := s.repo.GetIntegration(ctx, integrationID)
	if err != nil {
		return nil, apperrors.NewNotFound("интеграция не найдена")
	}

	// Генерируем имя ветки
	branchName := req.BranchName
	if branchName == "" {
		branchName = slugify(issueKey + "-" + issue.Title)
	}

	// Создаём GitHub client с токеном интеграции
	client := gh.NewClient(nil).WithAuthToken(integration.AccessToken)

	// Получаем default branch ref
	repo, _, err := client.Repositories.Get(ctx, integration.RepoOwner, integration.RepoName)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения репозитория из GitHub")
	}

	defaultBranch := repo.GetDefaultBranch()
	ref, _, err := client.Git.GetRef(ctx, integration.RepoOwner, integration.RepoName, "refs/heads/"+defaultBranch)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения ref default branch")
	}

	// Создаём новую ветку
	newRef := &gh.Reference{
		Ref:    gh.Ptr("refs/heads/" + branchName),
		Object: &gh.GitObject{SHA: ref.Object.SHA},
	}

	_, _, err = client.Git.CreateRef(ctx, integration.RepoOwner, integration.RepoName, newRef)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания ветки: " + err.Error())
	}

	return &CreateBranchResponse{BranchName: branchName}, nil
}

// ---------------------------------------------------------------------------
// CI/CD Pipeline Status (Workflow Runs)
// ---------------------------------------------------------------------------

// ListWorkflowRuns возвращает последние запуски workflow для всех интеграций проекта.
// Для каждой интеграции запрашивает GitHub API /repos/{owner}/{repo}/actions/runs
// и объединяет результаты, сортируя по дате создания (новые первые).
func (s *GitHubService) ListWorkflowRuns(ctx context.Context, projectID uuid.UUID) ([]WorkflowRunResponse, error) {
	integrations, err := s.repo.ListIntegrations(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения интеграций")
	}

	var allRuns []WorkflowRunResponse

	for _, integration := range integrations {
		client := gh.NewClient(nil).WithAuthToken(integration.AccessToken)

		opts := &gh.ListWorkflowRunsOptions{
			ListOptions: gh.ListOptions{PerPage: 20},
		}
		runs, _, err := client.Actions.ListRepositoryWorkflowRuns(ctx, integration.RepoOwner, integration.RepoName, opts)
		if err != nil {
			slog.Error("ошибка получения workflow runs из GitHub",
				"owner", integration.RepoOwner, "repo", integration.RepoName, "error", err)
			continue
		}

		for _, run := range runs.WorkflowRuns {
			allRuns = append(allRuns, toWorkflowRunResponse(run))
		}
	}

	return allRuns, nil
}

// GetWorkflowRunDetail возвращает детали конкретного запуска workflow, включая джобы.
// Ищет интеграцию проекта, подходящую для данного runID, и запрашивает GitHub API.
func (s *GitHubService) GetWorkflowRunDetail(ctx context.Context, projectID uuid.UUID, runID int64) (*WorkflowRunDetailResponse, error) {
	integrations, err := s.repo.ListIntegrations(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения интеграций")
	}

	// Пробуем каждую интеграцию — runID может принадлежать любому из репозиториев.
	for _, integration := range integrations {
		client := gh.NewClient(nil).WithAuthToken(integration.AccessToken)

		run, _, err := client.Actions.GetWorkflowRunByID(ctx, integration.RepoOwner, integration.RepoName, runID)
		if err != nil {
			continue // Пробуем следующую интеграцию
		}

		resp := &WorkflowRunDetailResponse{
			WorkflowRunResponse: toWorkflowRunResponse(run),
		}

		// Загружаем джобы для данного run
		jobs, _, err := client.Actions.ListWorkflowJobs(ctx, integration.RepoOwner, integration.RepoName, runID, &gh.ListWorkflowJobsOptions{
			ListOptions: gh.ListOptions{PerPage: 50},
		})
		if err == nil && jobs != nil {
			for _, job := range jobs.Jobs {
				j := WorkflowJobResponse{
					ID:         job.GetID(),
					Name:       job.GetName(),
					Status:     job.GetStatus(),
					Conclusion: job.GetConclusion(),
					StartedAt:  job.GetStartedAt().Time,
				}
				if job.CompletedAt != nil {
					t := job.GetCompletedAt().Time
					j.CompletedAt = &t
				}
				resp.Jobs = append(resp.Jobs, j)
			}
		}

		return resp, nil
	}

	return nil, apperrors.NewNotFound("workflow run не найден")
}

// toWorkflowRunResponse конвертирует GitHub API WorkflowRun в DTO ответа.
func toWorkflowRunResponse(run *gh.WorkflowRun) WorkflowRunResponse {
	return WorkflowRunResponse{
		ID:         run.GetID(),
		Name:       run.GetName(),
		Status:     run.GetStatus(),
		Conclusion: run.GetConclusion(),
		Branch:     run.GetHeadBranch(),
		Event:      run.GetEvent(),
		HTMLURL:    run.GetHTMLURL(),
		CreatedAt:  run.GetCreatedAt().Time,
		UpdatedAt:  run.GetUpdatedAt().Time,
		HeadSHA:    run.GetHeadSHA(),
		RunNumber:  run.GetRunNumber(),
	}
}

// ---------------------------------------------------------------------------
// GitHub-ссылки
// ---------------------------------------------------------------------------

// ListGitHubLinks возвращает все GitHub-ссылки для задачи по ключу.
func (s *GitHubService) ListGitHubLinks(ctx context.Context, issueKey string) ([]GitHubLinkResponse, error) {
	issue, err := s.repo.FindIssueByKey(ctx, issueKey)
	if err != nil {
		return nil, apperrors.NewNotFound("задача не найдена")
	}

	links, err := s.repo.ListGitHubLinks(ctx, issue.ID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения GitHub ссылок")
	}

	result := make([]GitHubLinkResponse, 0, len(links))
	for _, l := range links {
		result = append(result, *toGitHubLinkResponse(l))
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Валидация подписи
// ---------------------------------------------------------------------------

// ValidateWebhookSignature проверяет X-Hub-Signature-256 с использованием HMAC-SHA256.
func ValidateWebhookSignature(payload []byte, signature, secret string) bool {
	if !strings.HasPrefix(signature, "sha256=") {
		return false
	}

	sig, err := hex.DecodeString(signature[7:])
	if err != nil {
		return false
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	expected := mac.Sum(nil)

	return hmac.Equal(sig, expected)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// publishEvent публикует событие в NATS (nil-safe).
func (s *GitHubService) publishEvent(subject string, payload interface{}) {
	if s.nats == nil {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("не удалось сериализовать GitHub событие", "subject", subject, "error", err)
		return
	}
	if err := s.nats.Publish(subject, data); err != nil {
		slog.Error("не удалось опубликовать GitHub событие", "subject", subject, "error", err)
	}
}

// slugify преобразует строку в URL-безопасный формат для имени ветки.
func slugify(s string) string {
	s = strings.ToLower(s)
	var result strings.Builder
	prevDash := false
	for _, r := range s {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			result.WriteRune(r)
			prevDash = false
		} else if !prevDash && result.Len() > 0 {
			result.WriteRune('-')
			prevDash = true
		}
	}
	out := result.String()
	return strings.TrimRight(out, "-")
}

// markTestCasesOutdated проверяет изменённые файлы и помечает sync_status = "outdated"
// для тест-кейсов, чьи class_name или full_name совпадают с путём файла.
func (s *GitHubService) markTestCasesOutdated(ctx context.Context, integration *ent.GitHubIntegration, changedFiles []string) {
	if s.entDB == nil || len(changedFiles) == 0 {
		return
	}

	// Ищем тест-кейсы с заполненным class_name или full_name и статусом synced
	cases, err := s.entDB.TestCase.Query().
		Where(
			testcase.Or(
				testcase.ClassNameNotNil(),
				testcase.FullNameNotNil(),
			),
			testcase.SyncStatusEQ(testcase.SyncStatusSynced),
		).
		All(ctx)
	if err != nil {
		slog.Error("ошибка получения тест-кейсов для sync_status", "error", err)
		return
	}

	for _, tc := range cases {
		for _, file := range changedFiles {
			if matchTestCaseFile(tc, file) {
				if err := s.entDB.TestCase.UpdateOneID(tc.ID).
					SetSyncStatus(testcase.SyncStatusOutdated).
					Exec(ctx); err != nil {
					slog.Error("ошибка обновления sync_status", "case_key", tc.CaseKey, "error", err)
				}
				// Публикуем уведомление
				s.publishEvent("tms.test_case.outdated", map[string]interface{}{
					"case_key":  tc.CaseKey,
					"file_path": file,
				})
				break
			}
		}
	}
}

// matchTestCaseFile проверяет, соответствует ли изменённый файл тест-кейсу
// по class_name или full_name (с учётом паттернов Java/Python-пакетов).
func matchTestCaseFile(tc *ent.TestCase, filePath string) bool {
	// Преобразуем путь файла: src/test/java/com/example/MyTest.java -> com.example.MyTest
	normalized := strings.ReplaceAll(filePath, "/", ".")
	normalized = strings.ReplaceAll(normalized, "\\", ".")

	if tc.ClassName != "" && strings.Contains(normalized, tc.ClassName) {
		return true
	}
	if tc.FullName != "" && strings.Contains(normalized, tc.FullName) {
		return true
	}
	return false
}

// truncate обрезает строку до maxLen символов.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
