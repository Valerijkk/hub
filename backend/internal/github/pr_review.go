package github

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	gh "github.com/google/go-github/v68/github"
	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

const maxPRPatchBytes = 12_000

var prURLRe = regexp.MustCompile(`(?i)github\.com/([^/\s]+)/([^/\s]+)/pull/(\d+)`)

// PRReviewContext — данные PR для AI code review (M3-017).
type PRReviewContext struct {
	URL          string
	Title        string
	Body         string
	Author       string
	BaseBranch   string
	HeadBranch   string
	ChangedFiles []string
	PatchSummary string
}

// FetchPRForReview — PR metadata и усечённые diffs через integration token.
func (s *GitHubService) FetchPRForReview(ctx context.Context, prURL string) (*PRReviewContext, error) {
	owner, repo, number, err := parseGitHubPRURL(prURL)
	if err != nil {
		return nil, err
	}
	integration, err := s.repo.FindIntegrationByRepo(ctx, owner, repo)
	if err != nil {
		return nil, apperrors.NewNotFound("GitHub интеграция для репозитория не найдена — подключите repo в настройках проекта")
	}
	client := s.apiClient(integration.AccessToken)
	pr, _, err := client.PullRequests.Get(ctx, owner, repo, number)
	if err != nil {
		return nil, apperrors.NewInternal("не удалось загрузить PR из GitHub")
	}
	out := &PRReviewContext{
		URL:        pr.GetHTMLURL(),
		Title:      pr.GetTitle(),
		Body:       pr.GetBody(),
		Author:     pr.GetUser().GetLogin(),
		BaseBranch: pr.GetBase().GetRef(),
		HeadBranch: pr.GetHead().GetRef(),
	}
	files, _, err := client.PullRequests.ListFiles(ctx, owner, repo, number, &gh.ListOptions{PerPage: 30})
	if err != nil {
		return out, nil
	}
	var patch strings.Builder
	for _, f := range files {
		out.ChangedFiles = append(out.ChangedFiles, f.GetFilename())
		if patch.Len() >= maxPRPatchBytes {
			continue
		}
		if f.Patch != nil && *f.Patch != "" {
			fmt.Fprintf(&patch, "\n--- %s (%s) ---\n%s\n", f.GetFilename(), f.GetStatus(), *f.Patch)
		}
	}
	out.PatchSummary = strings.TrimSpace(patch.String())
	if len(out.PatchSummary) > maxPRPatchBytes {
		out.PatchSummary = out.PatchSummary[:maxPRPatchBytes] + "\n...(truncated)"
	}
	return out, nil
}

// ProjectIDForPRURL — Hub projectID интеграции по URL PR (для AI bot authz, P1-BE-31).
func (s *GitHubService) ProjectIDForPRURL(ctx context.Context, prURL string) (uuid.UUID, error) {
	owner, repo, _, err := parseGitHubPRURL(prURL)
	if err != nil {
		return uuid.Nil, err
	}
	integration, err := s.repo.FindIntegrationByRepo(ctx, owner, repo)
	if err != nil {
		return uuid.Nil, apperrors.NewNotFound("GitHub интеграция для репозитория не найдена — подключите repo в настройках проекта")
	}
	if integration.Edges.Project != nil {
		return integration.Edges.Project.ID, nil
	}
	proj, err := integration.QueryProject().Only(ctx)
	if err != nil {
		return uuid.Nil, apperrors.NewNotFound("GitHub интеграция для репозитория не найдена — подключите repo в настройках проекта")
	}
	return proj.ID, nil
}

func parseGitHubPRURL(raw string) (owner, repo string, number int, err error) {
	m := prURLRe.FindStringSubmatch(strings.TrimSpace(raw))
	if len(m) < 4 {
		return "", "", 0, apperrors.NewBadRequest("ожидается URL вида https://github.com/owner/repo/pull/123")
	}
	number, err = strconv.Atoi(m[3])
	if err != nil || number < 1 {
		return "", "", 0, apperrors.NewBadRequest("некорректный номер PR")
	}
	return m[1], m[2], number, nil
}
