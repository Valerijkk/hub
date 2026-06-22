package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
)

func TestListWorkflowRuns_MockGitHubAPI(t *testing.T) {
	now := time.Now().UTC()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/actions/runs") && r.Method == http.MethodGet {
			_, _ = w.Write([]byte(`{"total_count":1,"workflow_runs":[{
				"id":11,"name":"Build","status":"completed","conclusion":"success",
				"head_branch":"main","event":"push",
				"html_url":"https://github.com/acme/routes/actions/runs/11",
				"created_at":"` + now.Format(time.RFC3339) + `",
				"updated_at":"` + now.Format(time.RFC3339) + `",
				"head_sha":"abc","run_number":1
			}]}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)

	projectID := uuid.New()
	repo := &githubMockRepo{integrations: map[string]*ent.GitHubIntegration{}}
	_, err := repo.CreateIntegration(context.Background(), projectID, github.CreateIntegrationInput{
		RepoOwner: "acme", RepoName: "routes", AccessToken: "ghp_mock",
	})
	require.NoError(t, err)

	svc := github.NewGitHubServiceWithHTTPClient(repo, nil, srv.Client(), srv.URL)
	runs, err := svc.ListWorkflowRuns(context.Background(), projectID)
	require.NoError(t, err)
	require.Len(t, runs, 1)
	assert.Equal(t, int64(11), runs[0].ID)
	assert.Equal(t, "Build", runs[0].Name)
}

func TestListWorkflowRuns_ContinuesOnAPIError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "rate limited", http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	projectID := uuid.New()
	repo := &githubMockRepo{integrations: map[string]*ent.GitHubIntegration{}}
	_, err := repo.CreateIntegration(context.Background(), projectID, github.CreateIntegrationInput{
		RepoOwner: "acme", RepoName: "broken", AccessToken: "ghp_mock",
	})
	require.NoError(t, err)

	svc := github.NewGitHubServiceWithHTTPClient(repo, nil, srv.Client(), srv.URL)
	runs, err := svc.ListWorkflowRuns(context.Background(), projectID)
	require.NoError(t, err)
	assert.Empty(t, runs)
}
