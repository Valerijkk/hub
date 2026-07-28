package gitlab_test
import (
	"github.com/valerii/hub/backend/internal/gitlab"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

func testGitLabHandlerApp(t *testing.T, repo gitlab.Repository) *fiber.App {
	t.Helper()
	svc := gitlab.NewGitLabService(repo, nil)
	h := gitlab.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", uuid.New().String())
		return c.Next()
	}

	api.Post("/gitlab/webhook", h.HandleWebhook)
	proj := api.Group("/projects/:projectId", authMw)
	proj.Post("/gitlab/integrations", h.CreateIntegration)
	proj.Get("/gitlab/integrations", h.ListIntegrations)
	proj.Delete("/gitlab/integrations/:integrationId", h.DeleteIntegration)
	proj.Get("/gitlab/pipelines", h.ListPipelines)
	proj.Post("/gitlab/pipelines/trigger", h.TriggerPipeline)
	proj.Post("/issues/:issueKey/gitlab/create-branch", h.CreateBranch)
	return app
}

func TestHandler_HandleWebhook_MissingToken(t *testing.T) {
	app := testGitLabHandlerApp(t, &gitlabMockRepo{})
	req := httptest.NewRequest(http.MethodPost, "/api/gitlab/webhook", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-Gitlab-Event", "Push Hook")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_HandleWebhook_ValidToken(t *testing.T) {
	repo := &gitlabMockRepo{
		integrations: map[string]*ent.GitLabIntegration{
			"hub/backend": {
				ID:            uuid.New(),
				ProjectPath:   "hub/backend",
				WebhookSecret: "secret-token",
			},
		},
	}
	app := testGitLabHandlerApp(t, repo)

	body, err := json.Marshal(gitlab.WebhookPushEvent{
		ObjectKind: "push",
		Project:    gitlab.WebhookProject{PathWithNamespace: "hub/backend"},
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/gitlab/webhook", bytes.NewReader(body))
	req.Header.Set("X-Gitlab-Event", "Push Hook")
	req.Header.Set("X-Gitlab-Token", "secret-token")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusOK, resp.StatusCode)
}

func TestHandler_HandleWebhook_OversizedPayload(t *testing.T) {
	repo := &gitlabMockRepo{webhookSecrets: []string{"secret-token"}}
	svc := gitlab.NewGitLabService(repo, nil)
	h := gitlab.NewHandler(svc)
	app := fiber.New(fiber.Config{BodyLimit: gitlab.ExportMaxWebhookBodyBytes() + 1024})
	app.Post("/api/gitlab/webhook", h.HandleWebhook)

	body := make([]byte, gitlab.ExportMaxWebhookBodyBytes()+1)
	body[0] = '{'
	body[len(body)-1] = '}'

	req := httptest.NewRequest(http.MethodPost, "/api/gitlab/webhook", bytes.NewReader(body))
	req.Header.Set("X-Gitlab-Event", "Push Hook")
	req.Header.Set("X-Gitlab-Token", "secret-token")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)
}

func TestHandler_CreateIntegration_Validation(t *testing.T) {
	app := testGitLabHandlerApp(t, &gitlabMockRepo{})
	projectID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/gitlab/integrations",
		bytes.NewReader([]byte(`{"project_path":"","access_token":""}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateIntegration_Success(t *testing.T) {
	app := testGitLabHandlerApp(t, &gitlabMockRepo{})
	projectID := uuid.New()

	payload := `{"project_path":"group/repo","access_token":"glpat-test","webhook_secret":"wh-test"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/gitlab/integrations",
		bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusCreated, resp.StatusCode)

	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), "group/repo")
}

func TestHandler_ListIntegrations_InvalidProjectID(t *testing.T) {
	app := testGitLabHandlerApp(t, &gitlabMockRepo{})
	req := httptest.NewRequest(http.MethodGet, "/api/projects/not-uuid/gitlab/integrations", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_TriggerPipeline_MissingIntegrationID(t *testing.T) {
	app := testGitLabHandlerApp(t, &gitlabMockRepo{})
	projectID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/gitlab/pipelines/trigger",
		bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateBranch_MissingIssueKey(t *testing.T) {
	app := testGitLabHandlerApp(t, &gitlabMockRepo{})
	projectID := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/issues//gitlab/create-branch", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.True(t, resp.StatusCode == fiber.StatusBadRequest || resp.StatusCode == fiber.StatusNotFound)
}
