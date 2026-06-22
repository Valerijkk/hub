package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

func TestParseGitHubPRURL(t *testing.T) {
	owner, repo, num, err := github.ExportParseGitHubPRURL("https://github.com/acme/hub/pull/42")
	require.NoError(t, err)
	assert.Equal(t, "acme", owner)
	assert.Equal(t, "hub", repo)
	assert.Equal(t, 42, num)

	_, _, _, err = github.ExportParseGitHubPRURL("https://gitlab.com/a/b/merge_requests/1")
	require.Error(t, err)

	_, _, _, err = github.ExportParseGitHubPRURL("https://github.com/acme/hub/pull/0")
	require.Error(t, err)
}

func TestFetchPRForReview_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/files"):
			_, _ = w.Write([]byte(`[
				{"filename":"auth.go","status":"modified","patch":"+func fix() {}"}
			]`))
		case strings.Contains(r.URL.Path, "/pulls/"):
			_, _ = w.Write([]byte(`{
				"html_url":"https://github.com/acme/hub/pull/42",
				"title":"Fix auth",
				"body":"Security patch",
				"user":{"login":"dev"},
				"base":{"ref":"main"},
				"head":{"ref":"fix-auth"}
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	repo := &githubMockRepo{
		integrations: map[string]*ent.GitHubIntegration{},
	}
	_, err := repo.CreateIntegration(context.Background(), uuid.New(), github.CreateIntegrationInput{
		RepoOwner: "acme", RepoName: "hub", AccessToken: "ghp_test",
	})
	require.NoError(t, err)

	svc := github.NewGitHubServiceWithHTTPClient(repo, nil, srv.Client(), srv.URL)
	rc, err := svc.FetchPRForReview(context.Background(), "https://github.com/acme/hub/pull/42")
	require.NoError(t, err)
	require.NotNil(t, rc)
	assert.Equal(t, "Fix auth", rc.Title)
	assert.Equal(t, "dev", rc.Author)
	assert.Equal(t, "main", rc.BaseBranch)
	assert.Equal(t, "fix-auth", rc.HeadBranch)
	assert.Contains(t, rc.ChangedFiles, "auth.go")
	assert.Contains(t, rc.PatchSummary, "auth.go")
}

func TestFetchPRForReview_NoIntegration(t *testing.T) {
	svc := github.NewGitHubService(&githubMockRepo{}, nil)
	_, err := svc.FetchPRForReview(context.Background(), "https://github.com/acme/hub/pull/1")
	require.Error(t, err)
}

func TestFetchPRForReview_FilesErrorReturnsMetadata(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/pulls/") && !strings.HasSuffix(r.URL.Path, "/files") {
			_, _ = w.Write([]byte(`{
				"html_url":"https://github.com/acme/hub/pull/7",
				"title":"Partial",
				"user":{"login":"qa"},
				"base":{"ref":"main"},
				"head":{"ref":"feat"}
			}`))
			return
		}
		http.Error(w, "files unavailable", http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	repo := &githubMockRepo{integrations: map[string]*ent.GitHubIntegration{}}
	_, err := repo.CreateIntegration(context.Background(), uuid.New(), github.CreateIntegrationInput{
		RepoOwner: "acme", RepoName: "hub", AccessToken: "tok",
	})
	require.NoError(t, err)

	svc := github.NewGitHubServiceWithHTTPClient(repo, nil, srv.Client(), srv.URL)
	rc, err := svc.FetchPRForReview(context.Background(), "https://github.com/acme/hub/pull/7")
	require.NoError(t, err)
	assert.Equal(t, "Partial", rc.Title)
	assert.Empty(t, rc.ChangedFiles)
}
