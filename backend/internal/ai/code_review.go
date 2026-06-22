package ai

import (
	"context"
	"fmt"
	"regexp"

	githubmod "github.com/valerii/hub/backend/internal/github"
	gitlabmod "github.com/valerii/hub/backend/internal/gitlab"
)

var gitlabMRURLRe = regexp.MustCompile(`(?i)(https?://[^/\s]+)/(.+?)/-/merge_requests/(\d+)`)

// CodeReviewProvider загружает PR/MR context для /ai review.
type CodeReviewProvider interface {
	FetchReview(ctx context.Context, url string) (*ReviewContext, error)
}

// UnifiedCodeReviewProvider маршрутизирует GitHub PR и GitLab MR.
type UnifiedCodeReviewProvider struct {
	github *githubmod.GitHubService
	gitlab *gitlabmod.GitLabService
}

// NewUnifiedCodeReviewProvider — мульти-VCS загрузчик ревью.
func NewUnifiedCodeReviewProvider(github *githubmod.GitHubService, gitlab *gitlabmod.GitLabService) *UnifiedCodeReviewProvider {
	return &UnifiedCodeReviewProvider{github: github, gitlab: gitlab}
}

// FetchReview загружает review context из GitHub или GitLab.
func (p *UnifiedCodeReviewProvider) FetchReview(ctx context.Context, rawURL string) (*ReviewContext, error) {
	if p == nil {
		return nil, fmt.Errorf("code review not configured")
	}
	if p.gitlab != nil && gitlabMRURLRe.MatchString(rawURL) {
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
