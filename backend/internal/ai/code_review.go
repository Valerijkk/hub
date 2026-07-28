package ai

import (
	"context"
	"fmt"
	"regexp"

	"github.com/google/uuid"
	githubmod "github.com/valerii/hub/backend/internal/github"
	gitlabmod "github.com/valerii/hub/backend/internal/gitlab"
)

var gitlabMRURLRe = regexp.MustCompile(`(?i)(https?://[^/\s]+)/(.+?)/-/merge_requests/(\d+)`)

// CodeReviewProvider загружает PR/MR context для /ai review.
// P1-BE-31: userID — для membership check до org-wide token fetch.
type CodeReviewProvider interface {
	FetchReview(ctx context.Context, userID uuid.UUID, url string) (*ReviewContext, error)
}

// UnifiedCodeReviewProvider маршрутизирует GitHub PR и GitLab MR.
type UnifiedCodeReviewProvider struct {
	github *githubmod.GitHubService
	gitlab *gitlabmod.GitLabService
	access ProjectAccessChecker
}

// NewUnifiedCodeReviewProvider — мульти-VCS загрузчик ревью.
func NewUnifiedCodeReviewProvider(github *githubmod.GitHubService, gitlab *gitlabmod.GitLabService) *UnifiedCodeReviewProvider {
	return &UnifiedCodeReviewProvider{github: github, gitlab: gitlab}
}

// SetAccessChecker включает проверку membership проекта интеграции (P1-BE-31).
func (p *UnifiedCodeReviewProvider) SetAccessChecker(c ProjectAccessChecker) {
	if p == nil {
		return
	}
	p.access = c
}

// FetchReview загружает review context из GitHub или GitLab после authz.
func (p *UnifiedCodeReviewProvider) FetchReview(ctx context.Context, userID uuid.UUID, rawURL string) (*ReviewContext, error) {
	if p == nil {
		return nil, fmt.Errorf("code review not configured")
	}
	if err := p.authorizeReview(ctx, userID, rawURL); err != nil {
		return nil, err
	}
	// W5-A3-013: GitLab-like URL при nil gitlab — явная ошибка, не fallback на GitHub.
	if gitlabMRURLRe.MatchString(rawURL) {
		if p.gitlab == nil {
			return nil, fmt.Errorf("GitLab not configured")
		}
		mr, err := p.gitlab.FetchMRForReview(ctx, rawURL)
		if err != nil {
			return nil, err
		}
		return gitlabMRToReview(mr), nil
	}
	if p.github != nil {
		pr, err := p.github.FetchPRForReview(ctx, rawURL)
		if err != nil {
			return nil, err
		}
		return githubPRToReview(pr), nil
	}
	return nil, fmt.Errorf("no VCS integration configured")
}

func (p *UnifiedCodeReviewProvider) authorizeReview(ctx context.Context, userID uuid.UUID, rawURL string) error {
	if p.access == nil {
		return nil
	}
	var (
		projectID uuid.UUID
		err       error
	)
	switch {
	case gitlabMRURLRe.MatchString(rawURL):
		if p.gitlab == nil {
			return fmt.Errorf("GitLab not configured")
		}
		projectID, err = p.gitlab.ProjectIDForMRURL(ctx, rawURL)
	case p.github != nil:
		projectID, err = p.github.ProjectIDForPRURL(ctx, rawURL)
	default:
		return fmt.Errorf("no VCS integration configured")
	}
	if err != nil {
		return err
	}
	ok, err := p.access.CanViewProject(ctx, userID, projectID)
	if err != nil {
		return err
	}
	if !ok {
		return fmt.Errorf("code review unavailable for this project")
	}
	return nil
}

func githubPRToReview(pr *githubmod.PRReviewContext) *ReviewContext {
	if pr == nil {
		return nil
	}
	return &ReviewContext{
		URL:          pr.URL,
		Title:        pr.Title,
		Body:         pr.Body,
		Author:       pr.Author,
		BaseBranch:   pr.BaseBranch,
		HeadBranch:   pr.HeadBranch,
		ChangedFiles: append([]string(nil), pr.ChangedFiles...),
		PatchSummary: pr.PatchSummary,
	}
}

func gitlabMRToReview(mr *gitlabmod.MRReviewContext) *ReviewContext {
	if mr == nil {
		return nil
	}
	return &ReviewContext{
		URL:          mr.URL,
		Title:        mr.Title,
		Body:         mr.Body,
		Author:       mr.Author,
		BaseBranch:   mr.TargetBranch,
		HeadBranch:   mr.SourceBranch,
		ChangedFiles: append([]string(nil), mr.ChangedFiles...),
		PatchSummary: mr.PatchSummary,
	}
}
