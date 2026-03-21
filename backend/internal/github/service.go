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
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// issueKeyRegex находит ключи задач вида ABC-123 с границами слов.
var issueKeyRegex = regexp.MustCompile(`\b([A-Z]+-\d+)\b`)

// GitHubService реализует бизнес-логику интеграции с GitHub.
// Обрабатывает webhook-события, автоматически связывает PR/коммиты с задачами,
// управляет интеграциями и создаёт ветки через GitHub API.
type GitHubService struct {
	repo Repository
	nats *natsclient.Client
}

// NewGitHubService создаёт новый сервис GitHub-интеграции.
func NewGitHubService(repo Repository, nats *natsclient.Client) *GitHubService {
	return &GitHubService{repo: repo, nats: nats}
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
	}

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

				// GHUB-04: автосмена статуса при мерже
				if integration.MergeStatusID != nil {
					if updateErr := s.repo.UpdateIssueStatus(ctx, issue.ID, *integration.MergeStatusID); updateErr != nil {
						slog.Error("не удалось обновить статус задачи при мерже PR",
							"issue", key, "error", updateErr)
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

// truncate обрезает строку до maxLen символов.
func truncate(s string, maxLen int) string {
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	return string(runes[:maxLen])
}
