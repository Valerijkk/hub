package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
)

func TestEntRepository_UpdateIssueStatus(t *testing.T) {
	env := newGitHubRepoEnv(t)
	iss := env.createIssue(t)

	done, err := env.client.WorkflowStatus.Create().
		SetName("Done").
		SetCategory("done").
		SetProjectID(env.projectID).
		Save(env.ctx)
	require.NoError(t, err)

	require.NoError(t, env.repo.UpdateIssueStatus(env.ctx, iss.ID, done.ID))

	reloaded, err := env.client.Issue.Query().
		Where(issue.IDEQ(iss.ID)).
		WithStatus().
		Only(env.ctx)
	require.NoError(t, err)
	require.NotNil(t, reloaded.Edges.Status)
	assert.Equal(t, done.ID, reloaded.Edges.Status.ID)
}

func TestHandler_GetLinkPreview_WithPreviewService(t *testing.T) {
	env := newGitHubRepoEnv(t)
	svc := github.NewGitHubService(env.repo, nil)
	preview := github.NewLinkPreviewService(&previewMockRepo{
		issues: map[string]*ent.Issue{},
	}, nil)
	h := github.NewHandler(svc, preview)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", env.userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, nil)

	req := httptest.NewRequest(http.MethodGet, "/api/link-preview?url=https://hub.local/projects/x/tms/cases/test-cases/HUB-TC-1", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
