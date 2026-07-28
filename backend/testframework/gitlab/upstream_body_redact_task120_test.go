package gitlab_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/gitlab"
)

// TestTriggerPipeline_UpstreamBodyNotLeaked_W7A3006 — тело GitLab API не попадает в ошибку клиенту.
// Mutation: вернуть string(respBody) в NewInternal TriggerPipeline → assert.NotContains FAIL.
func TestTriggerPipeline_UpstreamBodyNotLeaked_W7A3006(t *testing.T) {
	const marker = "SECRET_UPSTREAM_TOKEN_LEAK_xyz"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		_, _ = w.Write([]byte(`{"message":"` + marker + `"}`))
	}))
	t.Cleanup(srv.Close)

	integrationID := uuid.New()
	projectID := uuid.New()
	ig := &ent.GitLabIntegration{
		ID:          integrationID,
		GitlabURL:   srv.URL,
		ProjectPath: "team/app",
		AccessToken: "glpat-test",
	}
	repo := &gitlabMockRepo{
		byID:      map[uuid.UUID]*ent.GitLabIntegration{integrationID: ig},
		byProject: map[uuid.UUID][]*ent.GitLabIntegration{projectID: {ig}},
	}
	svc := gitlabServiceWithTestHTTPClient(t, repo)

	_, err := svc.TriggerPipeline(context.Background(), projectID, gitlab.TriggerPipelineRequest{
		IntegrationID: integrationID.String(),
		Ref:           "main",
	})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), marker)
	assert.Contains(t, err.Error(), "HTTP 502")
}

// TestCreateBranch_UpstreamBodyNotLeaked_W7A3006 — то же для create branch.
// Mutation: вернуть string(body) в NewInternal CreateBranch → assert.NotContains FAIL.
func TestCreateBranch_UpstreamBodyNotLeaked_W7A3006(t *testing.T) {
	const marker = "BRANCH_UPSTREAM_SECRET_abc"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte(marker))
	}))
	t.Cleanup(srv.Close)

	integrationID := uuid.New()
	projectID := uuid.New()
	issueID := uuid.New()
	ig := &ent.GitLabIntegration{
		ID:          integrationID,
		GitlabURL:   srv.URL,
		ProjectPath: "team/app",
		AccessToken: "glpat-test",
	}
	repo := &gitlabMockRepo{
		byID:      map[uuid.UUID]*ent.GitLabIntegration{integrationID: ig},
		byProject: map[uuid.UUID][]*ent.GitLabIntegration{projectID: {ig}},
		issues: map[string]*ent.Issue{
			"HUB-7": {
				ID:       issueID,
				IssueKey: "HUB-7",
				Title:    "Branch me",
				Edges: ent.IssueEdges{
					Project: &ent.Project{ID: projectID},
				},
			},
		},
	}
	svc := gitlabServiceWithTestHTTPClient(t, repo)

	_, err := svc.CreateBranch(context.Background(), projectID, "HUB-7", gitlab.CreateBranchRequest{
		IntegrationID: integrationID.String(),
	})
	require.Error(t, err)
	assert.NotContains(t, err.Error(), marker)
	assert.Contains(t, err.Error(), "HTTP 403")
}
