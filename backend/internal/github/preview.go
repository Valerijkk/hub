package github

import (
	"context"
	"log/slog"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"encoding/json"

	safehttp "github.com/valerii/hub/backend/internal/platform/http"
)

// issueURLRegex распознаёт внутренние ссылки на задачи.
var issueURLRegex = regexp.MustCompile(`/issues/([A-Z]+-\d+)`)

// testCaseURLRegex распознаёт внутренние ссылки на тест-кейсы.
var testCaseURLRegex = regexp.MustCompile(`/tms/.*test-cases/([A-Z]+-TC-\d+)`)

// previewCacheTTL -- время жизни кеша превью.
const previewCacheTTL = 1 * time.Hour

// LinkPreviewService предоставляет превью для внутренних и внешних ссылок (MSG-11).
// Кеширует результаты в Redis на 1 час.
type LinkPreviewService struct {
	repo  Repository
	redis *redis.Client
}

// NewLinkPreviewService создаёт новый сервис превью ссылок.
func NewLinkPreviewService(repo Repository, redis *redis.Client) *LinkPreviewService {
	return &LinkPreviewService{repo: repo, redis: redis}
}

// GetLinkPreview определяет тип ссылки и возвращает превью-данные.
func (s *LinkPreviewService) GetLinkPreview(ctx context.Context, rawURL string) (*LinkPreviewResponse, error) {
	// Проверяем кеш
	cacheKey := "link_preview:" + rawURL
	if s.redis != nil {
		cached, err := s.redis.Get(ctx, cacheKey).Result()
		if err == nil && cached != "" {
			var resp LinkPreviewResponse
			if json.Unmarshal([]byte(cached), &resp) == nil {
				return &resp, nil
			}
		}
	}

	var resp *LinkPreviewResponse

	// Проверяем внутреннюю ссылку на задачу
	if matches := issueURLRegex.FindStringSubmatch(rawURL); len(matches) > 1 {
		resp = s.resolveIssuePreview(ctx, matches[1])
	}

	// Проверяем внутреннюю ссылку на тест-кейс
	if resp == nil {
		if matches := testCaseURLRegex.FindStringSubmatch(rawURL); len(matches) > 1 {
			resp = &LinkPreviewResponse{
				URL:   rawURL,
				Type:  "test_case",
				Title: matches[1],
			}
		}
	}

	// Внешний URL -- OpenGraph
	if resp == nil {
		resp = s.resolveExternalPreview(ctx, rawURL)
	}

	if resp == nil {
		resp = &LinkPreviewResponse{
			URL:   rawURL,
			Type:  "external",
			Title: rawURL,
		}
	}

	// Кешируем
	if s.redis != nil {
		data, _ := json.Marshal(resp)
		if err := s.redis.Set(ctx, cacheKey, string(data), previewCacheTTL).Err(); err != nil {
			slog.Error("не удалось кешировать превью", "url", rawURL, "error", err)
		}
	}

	return resp, nil
}

// resolveIssuePreview загружает данные задачи из БД для превью.
func (s *LinkPreviewService) resolveIssuePreview(ctx context.Context, issueKey string) *LinkPreviewResponse {
	issue, err := s.repo.FindIssueByKey(ctx, issueKey)
	if err != nil {
		return nil
	}

	resp := &LinkPreviewResponse{
		URL:       "",
		Type:      "issue",
		Title:     issue.Title,
		IssueKey:  issue.IssueKey,
		IssueType: string(issue.IssueType),
	}

	if issue.Edges.Status != nil {
		resp.Status = issue.Edges.Status.Name
	}
	if issue.Edges.Assignee != nil {
		resp.AssigneeName = issue.Edges.Assignee.Name
		resp.AvatarURL = issue.Edges.Assignee.AvatarURL
	}

	return resp
}

// resolveExternalPreview получает OpenGraph метаданные для внешнего URL через SafeHTTP (NEW-02).
func (s *LinkPreviewService) resolveExternalPreview(ctx context.Context, rawURL string) *LinkPreviewResponse {
	og, err := safehttp.FetchOpenGraphMetadata(ctx, rawURL)
	if err != nil {
		slog.Debug("не удалось получить OG данные", "url", rawURL, "error", err)
		return nil
	}

	parsed, _ := url.Parse(rawURL)
	domain := ""
	if parsed != nil {
		domain = parsed.Host
	}

	resp := &LinkPreviewResponse{
		URL:         rawURL,
		Type:        "external",
		Title:       og.Title,
		Description: truncate(og.Description, 200),
		Domain:      domain,
	}

	if og.ImageURL != "" {
		imgURL := og.ImageURL
		if strings.HasPrefix(imgURL, "/") && parsed != nil {
			imgURL = parsed.Scheme + "://" + parsed.Host + imgURL
		}
		resp.ImageURL = imgURL
	}

	return resp
}
