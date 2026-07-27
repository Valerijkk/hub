package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	gh "github.com/google/go-github/v68/github"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestToWorkflowRunResponse(t *testing.T) {
	now := time.Now()
	run := &gh.WorkflowRun{
		ID:         gh.Ptr(int64(42)),
		Name:       gh.Ptr("CI"),
		Status:     gh.Ptr("completed"),
		Conclusion: gh.Ptr("success"),
		HeadBranch: gh.Ptr("main"),
		Event:      gh.Ptr("push"),
		HTMLURL:    gh.Ptr("https://github.com/acme/repo/actions/runs/42"),
		CreatedAt:  &gh.Timestamp{Time: now},
		UpdatedAt:  &gh.Timestamp{Time: now},
		HeadSHA:    gh.Ptr("abc123"),
		RunNumber:  gh.Ptr(7),
	}

	resp := github.ExportToWorkflowRunResponse(run)
	assert.Equal(t, int64(42), resp.ID)
	assert.Equal(t, "CI", resp.Name)
	assert.Equal(t, "success", resp.Conclusion)
	assert.Equal(t, 7, resp.RunNumber)
}

func TestHandler_GetLinkPreview_MissingURL(t *testing.T) {
	env := newGitHubRoutesEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/link-preview", nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateBranch_InvalidJSON(t *testing.T) {
	env := newGitHubRoutesEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+env.projectID.String()+"/issues/HUB-1/github/create-branch", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
