package gitlab

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

const maxMRPatchBytes = 12_000

var gitlabMRURLRe = regexp.MustCompile(`(?i)(https?://[^/\s]+)/(.+?)/-/merge_requests/(\d+)`)

// MRReviewContext — GitLab MR для AI review.
type MRReviewContext struct {
	URL          string
	Title        string
	Body         string
	Author       string
	SourceBranch string
	TargetBranch string
	ChangedFiles []string
	PatchSummary string
}

// FetchMRForReview — MR metadata и diff через GitLab API v4.
func (s *GitLabService) FetchMRForReview(ctx context.Context, mrURL string) (*MRReviewContext, error) {
	baseHost, projectPath, iid, err := parseGitLabMRURL(mrURL)
	if err != nil {
		return nil, err
	}
	integration, err := s.repo.FindIntegrationByProjectPath(ctx, projectPath)
	if err != nil {
		return nil, apperrors.NewNotFound("GitLab интеграция для проекта не найдена")
	}
	if !strings.Contains(strings.TrimRight(integration.GitlabURL, "/"), strings.TrimRight(baseHost, "/")) &&
		!strings.Contains(strings.TrimRight(baseHost, "/"), strings.TrimRight(integration.GitlabURL, "/")) {
		return nil, apperrors.NewNotFound("GitLab интеграция для этого хоста не найдена")
	}
	glBase := strings.TrimRight(integration.GitlabURL, "/")
	encodedPath := strings.ReplaceAll(integration.ProjectPath, "/", "%2F")

	mr, err := s.fetchGitLabMR(ctx, glBase, encodedPath, integration.AccessToken, iid)
	if err != nil {
		return nil, err
	}
	out := &MRReviewContext{
		URL:          getString(mr, "web_url"),
		Title:        getString(mr, "title"),
		Body:         getString(mr, "description"),
		SourceBranch: getString(mr, "source_branch"),
		TargetBranch: getString(mr, "target_branch"),
	}
	if author, ok := mr["author"].(map[string]interface{}); ok {
		out.Author = getString(author, "username")
	}
	changes, err := s.fetchGitLabMRChanges(ctx, glBase, encodedPath, integration.AccessToken, iid)
	if err != nil {
		return out, nil
	}
	var patch strings.Builder
	for _, ch := range changes {
		path := getString(ch, "new_path")
		if path == "" {
			path = getString(ch, "old_path")
		}
		if path != "" {
			out.ChangedFiles = append(out.ChangedFiles, path)
		}
		if patch.Len() >= maxMRPatchBytes {
			continue
		}
		if diff := getString(ch, "diff"); diff != "" {
			fmt.Fprintf(&patch, "\n--- %s ---\n%s\n", path, diff)
		}
	}
	out.PatchSummary = strings.TrimSpace(patch.String())
	if len(out.PatchSummary) > maxMRPatchBytes {
		out.PatchSummary = out.PatchSummary[:maxMRPatchBytes] + "\n...(truncated)"
	}
	return out, nil
}

func (s *GitLabService) fetchGitLabMR(ctx context.Context, base, encodedPath, token string, iid int) (map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d", base, encodedPath, iid)
	return s.gitlabJSON(ctx, url, token)
}

func (s *GitLabService) fetchGitLabMRChanges(ctx context.Context, base, encodedPath, token string, iid int) ([]map[string]interface{}, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%s/merge_requests/%d/changes", base, encodedPath, iid)
	body, err := s.gitlabJSON(ctx, url, token)
	if err != nil {
		return nil, err
	}
	raw, ok := body["changes"].([]interface{})
	if !ok {
		return nil, nil
	}
	out := make([]map[string]interface{}, 0, len(raw))
	for _, item := range raw {
		if m, ok := item.(map[string]interface{}); ok {
			out = append(out, m)
		}
	}
	return out, nil
}

func (s *GitLabService) gitlabJSON(ctx context.Context, url, token string) (map[string]interface{}, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("PRIVATE-TOKEN", token)
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, apperrors.NewInternal("GitLab API недоступен")
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxGitLabResponseBodyBytes))
	if err != nil {
		return nil, apperrors.NewInternal("ошибка чтения ответа GitLab")
	}
	if resp.StatusCode >= 400 {
		return nil, apperrors.NewInternal(fmt.Sprintf("GitLab API %d", resp.StatusCode))
	}
	var out map[string]interface{}
	if err := json.Unmarshal(raw, &out); err != nil {
		return nil, apperrors.NewInternal("некорректный JSON GitLab")
	}
	return out, nil
}

func parseGitLabMRURL(raw string) (host, projectPath string, iid int, err error) {
	m := gitlabMRURLRe.FindStringSubmatch(strings.TrimSpace(raw))
	if len(m) < 4 {
		return "", "", 0, apperrors.NewBadRequest("ожидается URL вида https://gitlab.com/group/project/-/merge_requests/123")
	}
	iid, err = strconv.Atoi(m[3])
	if err != nil || iid < 1 {
		return "", "", 0, apperrors.NewBadRequest("некорректный номер MR")
	}
	return m[1], m[2], iid, nil
}

func getString(m map[string]interface{}, key string) string {
	if v, ok := m[key].(string); ok {
		return v
	}
	return ""
}
