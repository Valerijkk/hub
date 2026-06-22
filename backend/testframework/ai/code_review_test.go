package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	githubmod "github.com/valerii/hub/backend/internal/github"
	gitlabmod "github.com/valerii/hub/backend/internal/gitlab"
)

func TestGithubPRToReview_Nil(t *testing.T) {
	assert.Nil(t, ai.ExportGithubPRToReview(nil))
}

func TestGitlabMRToReview_Nil(t *testing.T) {
	assert.Nil(t, ai.ExportGitlabMRToReview(nil))
}

func TestUnifiedCodeReviewProvider_NoVCS(t *testing.T) {
	p := ai.NewUnifiedCodeReviewProvider(nil, nil)
	_, err := p.FetchReview(t.Context(), "https://gitlab.com/g/p/-/merge_requests/5")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no VCS")
}

func TestGithubPRToReview(t *testing.T) {
	rc := ai.ExportGithubPRToReview(&githubmod.PRReviewContext{
		URL:          "https://github.com/o/r/pull/1",
		Title:        "Fix bug",
		Author:       "dev",
		BaseBranch:   "main",
		HeadBranch:   "fix",
		ChangedFiles: []string{"a.go"},
		PatchSummary: "diff",
	})
	require.NotNil(t, rc)
	assert.Equal(t, "Fix bug", rc.Title)
	assert.Equal(t, "main", rc.BaseBranch)
}

func TestGitlabMRToReview(t *testing.T) {
	rc := ai.ExportGitlabMRToReview(&gitlabmod.MRReviewContext{
		URL:          "https://gitlab.com/g/p/-/merge_requests/2",
		Title:        "Feature",
		SourceBranch: "feat",
		TargetBranch: "main",
		ChangedFiles: []string{"b.ts"},
	})
	require.NotNil(t, rc)
	assert.Equal(t, "feat", rc.HeadBranch)
	assert.Equal(t, "main", rc.BaseBranch)
}

func TestUnifiedCodeReviewProviderNil(t *testing.T) {
	p := &ai.UnifiedCodeReviewProvider{}
	_, err := p.FetchReview(t.Context(), "https://github.com/o/r/pull/1")
	assert.Error(t, err)
}

func TestBuildReviewPromptUnified(t *testing.T) {
	sys, user := ai.ExportBuildReviewPrompt(&ai.ReviewContext{
		URL:   "https://gitlab.com/g/p/-/merge_requests/3",
		Title: "MR title",
	})
	assert.Contains(t, sys, "reviewer")
	assert.Contains(t, user, "MR title")
}

func TestExtractPRURLGitLab(t *testing.T) {
	url := ai.ExportExtractPRURL("review https://gitlab.com/group/proj/-/merge_requests/99 please")
	assert.Equal(t, "https://gitlab.com/group/proj/-/merge_requests/99", url)
}
