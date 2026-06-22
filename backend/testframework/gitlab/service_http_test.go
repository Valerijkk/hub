package gitlab_test
import (
	"github.com/valerii/hub/backend/internal/gitlab"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

func gitlabServiceWithTestHTTPClient(t *testing.T, repo gitlab.Repository) *gitlab.GitLabService {
	t.Helper()
	return gitlab.NewGitLabServiceWithHTTPClient(repo, nil, &http.Client{Timeout: 5 * time.Second})
}

func TestListPipelines_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, "test-token", r.Header.Get("PRIVATE-TOKEN"))
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"id": 101, "status": "success", "ref": "main", "sha": "abc",
				"web_url": "https://gitlab.example/p/101", "source": "push",
				"created_at": time.Now().UTC(), "updated_at": time.Now().UTC(),
			},
		})
	}))
	t.Cleanup(srv.Close)

	projectID := uuid.New()
	integrationID := uuid.New()
	repo := &gitlabMockRepo{
		byProject: map[uuid.UUID][]*ent.GitLabIntegration{
			projectID: {{
				ID:          integrationID,
				GitlabURL:   srv.URL,
				ProjectPath: "group/repo",
				AccessToken: "test-token",
			}},
		},
	}
	svc := gitlabServiceWithTestHTTPClient(t, repo)

	pipelines, err := svc.ListPipelines(context.Background(), projectID)
	require.NoError(t, err)
	require.Len(t, pipelines, 1)
	assert.Equal(t, int64(101), pipelines[0].ID)
	assert.Equal(t, "success", pipelines[0].Status)
}

func TestListPipelines_SkipsFailedIntegration(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	t.Cleanup(srv.Close)

	projectID := uuid.New()
	repo := &gitlabMockRepo{
		byProject: map[uuid.UUID][]*ent.GitLabIntegration{
			projectID: {{
				ID:          uuid.New(),
				GitlabURL:   srv.URL,
				ProjectPath: "group/repo",
				AccessToken: "bad-token",
			}},
		},
	}
	svc := gitlabServiceWithTestHTTPClient(t, repo)

	pipelines, err := svc.ListPipelines(context.Background(), projectID)
	require.NoError(t, err)
	assert.Empty(t, pipelines)
}

func TestTriggerPipeline_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"id": 55, "web_url": "https://gitlab.example/p/55", "status": "pending", "ref": "main",
		})
	}))
	t.Cleanup(srv.Close)

	integrationID := uuid.New()
	repo := &gitlabMockRepo{
		byID: map[uuid.UUID]*ent.GitLabIntegration{
			integrationID: {
				ID:          integrationID,
				GitlabURL:   srv.URL,
				ProjectPath: "team/app",
				AccessToken: "glpat-test",
			},
		},
	}
	svc := gitlabServiceWithTestHTTPClient(t, repo)

	resp, err := svc.TriggerPipeline(context.Background(), uuid.New(), gitlab.TriggerPipelineRequest{
		IntegrationID: integrationID.String(),
		Ref:           "main",
		Variables:     map[string]string{"CI_ENV": "staging"},
	})
	require.NoError(t, err)
	assert.Equal(t, int64(55), resp.PipelineID)
	assert.Equal(t, "pending", resp.Status)
}

func TestTriggerPipeline_InvalidIntegrationID(t *testing.T) {
	svc := gitlabServiceWithTestHTTPClient(t, &gitlabMockRepo{})
	_, err := svc.TriggerPipeline(context.Background(), uuid.New(), gitlab.TriggerPipelineRequest{
		IntegrationID: "not-uuid",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "integration_id")
}

func TestTriggerPipeline_IntegrationNotFound(t *testing.T) {
	svc := gitlabServiceWithTestHTTPClient(t, &gitlabMockRepo{})
	_, err := svc.TriggerPipeline(context.Background(), uuid.New(), gitlab.TriggerPipelineRequest{
		IntegrationID: uuid.New().String(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}

func TestCreateBranch_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method)
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(srv.Close)

	integrationID := uuid.New()
	repo := &gitlabMockRepo{
		issues: map[string]*ent.Issue{
			"HUB-5": {ID: uuid.New(), IssueKey: "HUB-5", Title: "Add endpoint"},
		},
		byID: map[uuid.UUID]*ent.GitLabIntegration{
			integrationID: {
				ID:          integrationID,
				GitlabURL:   srv.URL,
				ProjectPath: "hub/backend",
				AccessToken: "token",
			},
		},
	}
	svc := gitlabServiceWithTestHTTPClient(t, repo)

	resp, err := svc.CreateBranch(context.Background(), "HUB-5", gitlab.CreateBranchRequest{
		IntegrationID: integrationID.String(),
		BranchName:    "feature/hub-5",
		Ref:           "main",
	})
	require.NoError(t, err)
	assert.Equal(t, "feature/hub-5", resp.BranchName)
}

func TestCreateBranch_AutoSlugifyBranchName(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))
	t.Cleanup(srv.Close)

	integrationID := uuid.New()
	repo := &gitlabMockRepo{
		issues: map[string]*ent.Issue{
			"HUB-9": {ID: uuid.New(), IssueKey: "HUB-9", Title: "Fix login"},
		},
		byID: map[uuid.UUID]*ent.GitLabIntegration{
			integrationID: {
				ID:          integrationID,
				GitlabURL:   srv.URL,
				ProjectPath: "hub/backend",
				AccessToken: "token",
			},
		},
	}
	svc := gitlabServiceWithTestHTTPClient(t, repo)

	resp, err := svc.CreateBranch(context.Background(), "HUB-9", gitlab.CreateBranchRequest{
		IntegrationID: integrationID.String(),
	})
	require.NoError(t, err)
	assert.Equal(t, "hub-9-fix-login", resp.BranchName)
}

func TestCreateBranch_IssueNotFound(t *testing.T) {
	svc := gitlabServiceWithTestHTTPClient(t, &gitlabMockRepo{})
	_, err := svc.CreateBranch(context.Background(), "NONE-1", gitlab.CreateBranchRequest{
		IntegrationID: uuid.New().String(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}

func TestListMergeRequests_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/merge_requests")
		_ = json.NewEncoder(w).Encode([]map[string]interface{}{
			{"iid": 3, "title": "Fix bug", "state": "opened"},
		})
	}))
	t.Cleanup(srv.Close)

	projectID := uuid.New()
	repo := &gitlabMockRepo{
		byProject: map[uuid.UUID][]*ent.GitLabIntegration{
			projectID: {{
				ID:          uuid.New(),
				GitlabURL:   srv.URL,
				ProjectPath: "org/service",
				AccessToken: "token",
			}},
		},
	}
	svc := gitlabServiceWithTestHTTPClient(t, repo)

	mrs, err := svc.ListMergeRequests(context.Background(), projectID)
	require.NoError(t, err)
	require.Len(t, mrs, 1)
	assert.Equal(t, float64(3), mrs[0]["iid"])
}
