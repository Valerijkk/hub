package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"regexp"
	"strings"
	"time"
	"unicode"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// issueKeyRegex находит ключи задач вида ABC-123 с границами слов.
var issueKeyRegex = regexp.MustCompile(`\b([A-Z]+-\d+)\b`)

// GitLabService реализует бизнес-логику интеграции с GitLab.
// Обрабатывает webhook-события, управляет интеграциями,
// запрашивает пайплайны и создаёт ветки через GitLab API.
type GitLabService struct {
	repo  Repository
	nats  *natsclient.Client
	entDB *ent.Client
}

// NewGitLabService создаёт новый сервис GitLab-интеграции.
func NewGitLabService(repo Repository, nats *natsclient.Client, entDB ...*ent.Client) *GitLabService {
	svc := &GitLabService{repo: repo, nats: nats}
	if len(entDB) > 0 {
		svc.entDB = entDB[0]
	}
	return svc
}

// ---------------------------------------------------------------------------
// Интеграции
// ---------------------------------------------------------------------------

// CreateIntegration создаёт интеграцию проекта с GitLab-репозиторием.
func (s *GitLabService) CreateIntegration(ctx context.Context, projectID uuid.UUID, req CreateIntegrationRequest) (*IntegrationResponse, error) {
	gitlabURL := req.GitLabURL
	if gitlabURL == "" {
		gitlabURL = "https://gitlab.com"
	}

	input := CreateIntegrationInput{
		GitLabURL:     gitlabURL,
		ProjectPath:   req.ProjectPath,
		AccessToken:   req.AccessToken,
		WebhookSecret: req.WebhookSecret,
	}

	integration, err := s.repo.CreateIntegration(ctx, projectID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания интеграции с GitLab")
	}

	return toIntegrationResponse(integration), nil
}

// ListIntegrations возвращает все GitLab-интеграции проекта.
func (s *GitLabService) ListIntegrations(ctx context.Context, projectID uuid.UUID) ([]IntegrationResponse, error) {
	integrations, err := s.repo.ListIntegrations(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения GitLab-интеграций")
	}

	result := make([]IntegrationResponse, 0, len(integrations))
	for _, ig := range integrations {
		result = append(result, *toIntegrationResponse(ig))
	}
	return result, nil
}

// DeleteIntegration удаляет интеграцию по ID.
func (s *GitLabService) DeleteIntegration(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteIntegration(ctx, id)
}

// ---------------------------------------------------------------------------
// Webhook обработчики
// ---------------------------------------------------------------------------

// HandleWebhook маршрутизирует webhook-событие к соответствующему обработчику.
func (s *GitLabService) HandleWebhook(ctx context.Context, eventType string, payload []byte) error {
	switch eventType {
	case "Push Hook":
		return s.handlePushWebhook(ctx, payload)
	case "Merge Request Hook":
		return s.handleMergeRequestWebhook(ctx, payload)
	case "Pipeline Hook":
		return s.handlePipelineWebhook(ctx, payload)
	default:
		slog.Debug("игнорируем GitLab webhook событие", "event", eventType)
		return nil
	}
}

// handlePushWebhook обрабатывает push-событие, привязывая коммиты к задачам.
func (s *GitLabService) handlePushWebhook(ctx context.Context, payload []byte) error {
	var event WebhookPushEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return apperrors.NewBadRequest("некорректный payload push-события GitLab")
	}

	projectPath := event.Project.PathWithNamespace

	_, err := s.repo.FindIntegrationByProjectPath(ctx, projectPath)
	if err != nil {
		slog.Warn("GitLab-интеграция не найдена для проекта", "project_path", projectPath)
		return nil
	}

	for _, commit := range event.Commits {
		message := commit.Message
		keys := issueKeyRegex.FindAllString(message, -1)

		for _, key := range keys {
			_, findErr := s.repo.FindIssueByKey(ctx, key)
			if findErr != nil {
				continue
			}

			// Публикуем NATS-событие о привязке коммита к задаче
			s.publishEvent("gitlab.push.commit_linked", map[string]interface{}{
				"issue_key":    key,
				"commit_sha":   commit.ID,
				"commit_url":   commit.URL,
				"commit_msg":   truncate(commit.Message, 500),
				"author":       commit.Author.Name,
				"project_path": projectPath,
			})
		}
	}

	return nil
}

// handleMergeRequestWebhook обрабатывает merge_request событие.
func (s *GitLabService) handleMergeRequestWebhook(ctx context.Context, payload []byte) error {
	var event WebhookMergeRequestEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return apperrors.NewBadRequest("некорректный payload merge_request-события GitLab")
	}

	projectPath := event.Project.PathWithNamespace
	mr := event.ObjectAttributes

	_, err := s.repo.FindIntegrationByProjectPath(ctx, projectPath)
	if err != nil {
		slog.Warn("GitLab-интеграция не найдена", "project_path", projectPath)
		return nil
	}

	// Ищем ключи задач в title и description MR
	searchText := mr.Title + " " + mr.Description
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

	for _, key := range uniqueKeys {
		_, findErr := s.repo.FindIssueByKey(ctx, key)
		if findErr != nil {
			continue
		}

		// Публикуем NATS-событие о привязке MR к задаче
		s.publishEvent("gitlab.mr.linked", map[string]interface{}{
			"issue_key":     key,
			"mr_iid":        mr.IID,
			"mr_title":      mr.Title,
			"mr_url":        mr.URL,
			"mr_state":      mr.State,
			"mr_action":     mr.Action,
			"source_branch": mr.SourceBranch,
			"project_path":  projectPath,
		})
	}

	return nil
}

// handlePipelineWebhook обрабатывает pipeline событие.
func (s *GitLabService) handlePipelineWebhook(ctx context.Context, payload []byte) error {
	var event WebhookPipelineEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return apperrors.NewBadRequest("некорректный payload pipeline-события GitLab")
	}

	projectPath := event.Project.PathWithNamespace
	pipeline := event.ObjectAttributes

	_, err := s.repo.FindIntegrationByProjectPath(ctx, projectPath)
	if err != nil {
		slog.Warn("GitLab-интеграция не найдена", "project_path", projectPath)
		return nil
	}

	// Публикуем NATS-событие об обновлении пайплайна
	s.publishEvent("gitlab.pipeline.updated", map[string]interface{}{
		"pipeline_id":  pipeline.ID,
		"status":       pipeline.Status,
		"ref":          pipeline.Ref,
		"sha":          pipeline.SHA,
		"project_path": projectPath,
	})

	return nil
}

// ---------------------------------------------------------------------------
// Пайплайны (GitLab API)
// ---------------------------------------------------------------------------

// ListPipelines возвращает последние пайплайны для всех GitLab-интеграций проекта.
func (s *GitLabService) ListPipelines(ctx context.Context, projectID uuid.UUID) ([]PipelineResponse, error) {
	integrations, err := s.repo.ListIntegrations(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения GitLab-интеграций")
	}

	var allPipelines []PipelineResponse

	for _, integration := range integrations {
		pipelines, err := s.fetchPipelines(ctx, integration)
		if err != nil {
			slog.Error("ошибка получения пайплайнов из GitLab",
				"project_path", integration.ProjectPath, "error", err)
			continue
		}
		allPipelines = append(allPipelines, pipelines...)
	}

	return allPipelines, nil
}

// fetchPipelines запрашивает последние пайплайны из GitLab API.
func (s *GitLabService) fetchPipelines(ctx context.Context, integration *ent.GitLabIntegration) ([]PipelineResponse, error) {
	baseURL := strings.TrimRight(integration.GitlabURL, "/")
	encodedPath := strings.ReplaceAll(integration.ProjectPath, "/", "%2F")
	url := fmt.Sprintf("%s/api/v4/projects/%s/pipelines?per_page=20&order_by=id&sort=desc", baseURL, encodedPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", integration.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab API вернул %d: %s", resp.StatusCode, string(body))
	}

	var pipelines []struct {
		ID        int64     `json:"id"`
		Status    string    `json:"status"`
		Ref       string    `json:"ref"`
		SHA       string    `json:"sha"`
		WebURL    string    `json:"web_url"`
		Source    string    `json:"source"`
		CreatedAt time.Time `json:"created_at"`
		UpdatedAt time.Time `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&pipelines); err != nil {
		return nil, err
	}

	result := make([]PipelineResponse, 0, len(pipelines))
	for _, p := range pipelines {
		result = append(result, PipelineResponse{
			ID:        p.ID,
			Status:    p.Status,
			Ref:       p.Ref,
			SHA:       p.SHA,
			WebURL:    p.WebURL,
			Source:    p.Source,
			CreatedAt: p.CreatedAt,
			UpdatedAt: p.UpdatedAt,
		})
	}

	return result, nil
}

// ---------------------------------------------------------------------------
// Создание ветки (GitLab API)
// ---------------------------------------------------------------------------

// CreateBranch создаёт ветку в GitLab-репозитории через API.
func (s *GitLabService) CreateBranch(ctx context.Context, issueKey string, req CreateBranchRequest) (*CreateBranchResponse, error) {
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

	ref := req.Ref
	if ref == "" {
		ref = "main"
	}

	// Создаём ветку через GitLab API
	baseURL := strings.TrimRight(integration.GitlabURL, "/")
	encodedPath := strings.ReplaceAll(integration.ProjectPath, "/", "%2F")
	url := fmt.Sprintf("%s/api/v4/projects/%s/repository/branches?branch=%s&ref=%s",
		baseURL, encodedPath, branchName, ref)

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, url, nil)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания запроса к GitLab API")
	}
	httpReq.Header.Set("PRIVATE-TOKEN", integration.AccessToken)

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка вызова GitLab API: " + err.Error())
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return nil, apperrors.NewInternal(fmt.Sprintf("ошибка создания ветки в GitLab: %d %s", resp.StatusCode, string(body)))
	}

	return &CreateBranchResponse{BranchName: branchName}, nil
}

// ---------------------------------------------------------------------------
// Merge Requests (GitLab API)
// ---------------------------------------------------------------------------

// ListMergeRequests возвращает открытые MR для всех GitLab-интеграций проекта.
func (s *GitLabService) ListMergeRequests(ctx context.Context, projectID uuid.UUID) ([]map[string]interface{}, error) {
	integrations, err := s.repo.ListIntegrations(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения GitLab-интеграций")
	}

	var allMRs []map[string]interface{}

	for _, integration := range integrations {
		mrs, err := s.fetchMergeRequests(ctx, integration)
		if err != nil {
			slog.Error("ошибка получения MR из GitLab",
				"project_path", integration.ProjectPath, "error", err)
			continue
		}
		allMRs = append(allMRs, mrs...)
	}

	return allMRs, nil
}

// fetchMergeRequests запрашивает MR из GitLab API.
func (s *GitLabService) fetchMergeRequests(ctx context.Context, integration *ent.GitLabIntegration) ([]map[string]interface{}, error) {
	baseURL := strings.TrimRight(integration.GitlabURL, "/")
	encodedPath := strings.ReplaceAll(integration.ProjectPath, "/", "%2F")
	url := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests?state=opened&per_page=20&order_by=updated_at&sort=desc", baseURL, encodedPath)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", integration.AccessToken)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("GitLab API вернул %d: %s", resp.StatusCode, string(body))
	}

	var mrs []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&mrs); err != nil {
		return nil, err
	}

	return mrs, nil
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// toIntegrationResponse конвертирует ent-модель в DTO.
func toIntegrationResponse(ig *ent.GitLabIntegration) *IntegrationResponse {
	return &IntegrationResponse{
		ID:          ig.ID.String(),
		GitLabURL:   ig.GitlabURL,
		ProjectPath: ig.ProjectPath,
		CreatedAt:   ig.CreatedAt,
	}
}

// publishEvent публикует событие в NATS (nil-safe).
func (s *GitLabService) publishEvent(subject string, payload interface{}) {
	if s.nats == nil {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("не удалось сериализовать GitLab событие", "subject", subject, "error", err)
		return
	}
	if err := s.nats.Publish(subject, data); err != nil {
		slog.Error("не удалось опубликовать GitLab событие", "subject", subject, "error", err)
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
