package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

func testGitHubHandlerApp(t *testing.T, repo *githubMockRepo) *fiber.App {
	t.Helper()
	svc := github.NewGitHubService(repo, nil)
	h := github.NewHandler(svc, nil)

	app := fiber.New()
	app.Post("/github/webhook", h.HandleWebhook)
	app.Get("/issues/:issueKey/github-links", func(c fiber.Ctx) error {
		if c.Locals("userID") == nil {
			c.Locals("userID", uuid.New().String())
		}
		return c.Next()
	}, h.ListGitHubLinks)
	app.Post("/projects/:projectId/issues/:issueKey/github/create-branch", func(c fiber.Ctx) error { return c.Next() }, h.CreateBranch)
	app.Get("/projects/:projectId/github/workflows", func(c fiber.Ctx) error { return c.Next() }, h.ListWorkflowRuns)
	app.Get("/projects/:projectId/github/workflows/:runId", func(c fiber.Ctx) error { return c.Next() }, h.GetWorkflowRunDetail)
	return app
}

func TestHandler_ListGitHubLinks_IssueNotFound(t *testing.T) {
	app := testGitHubHandlerApp(t, &githubMockRepo{})
	req := httptest.NewRequest(http.MethodGet, "/issues/MISSING-1/github-links", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.GreaterOrEqual(t, resp.StatusCode, 400)
}

func TestHandler_ListWorkflowRuns_InvalidProjectID(t *testing.T) {
	app := testGitHubHandlerApp(t, &githubMockRepo{})
	req := httptest.NewRequest(http.MethodGet, "/projects/not-a-uuid/github/workflows", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ListWorkflowRuns_Success(t *testing.T) {
	app := testGitHubHandlerApp(t, &githubMockRepo{})
	projectID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String()+"/github/workflows", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var runs []github.WorkflowRunResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&runs))
	assert.Empty(t, runs)
}

func TestHandler_GetWorkflowRunDetail_InvalidRunID(t *testing.T) {
	app := testGitHubHandlerApp(t, &githubMockRepo{})
	projectID := uuid.New()
	req := httptest.NewRequest(http.MethodGet, "/projects/"+projectID.String()+"/github/workflows/not-int", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestParseProjectID_Validation(t *testing.T) {
	t.Run("nil uuid rejected", func(t *testing.T) {
		app := fiber.New()
		app.Get("/p/:projectId", func(c fiber.Ctx) error {
			_, err := github.ExportParseProjectID(c)
			return err
		})
		req := httptest.NewRequest(http.MethodGet, "/p/"+uuid.Nil.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestHandler_HandleWebhook_ValidSignature(t *testing.T) {
	secret := "webhook-secret"
	repo := &githubMockRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"acme/hub": {
				ID:            uuid.New(),
				RepoOwner:     "acme",
				RepoName:      "hub",
				WebhookSecret: secret,
			},
		},
	}
	app := testGitHubHandlerApp(t, repo)

	body := []byte(`{"zen":"test","repository":{"name":"hub","owner":{"login":"acme"}}}`)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/github/webhook", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", sig)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_HandleWebhook_MissingSignature(t *testing.T) {
	app := testGitHubHandlerApp(t, &githubMockRepo{})
	req := httptest.NewRequest(http.MethodPost, "/github/webhook", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-GitHub-Event", "push")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestHandler_HandleWebhook_OversizedPayload(t *testing.T) {
	secret := "webhook-secret"
	repo := &githubMockRepo{webhookSecrets: []string{secret}}
	svc := github.NewGitHubService(repo, nil)
	h := github.NewHandler(svc, nil)
	app := fiber.New(fiber.Config{BodyLimit: github.ExportMaxWebhookBodyBytes() + 1024})
	app.Post("/github/webhook", h.HandleWebhook)

	body := make([]byte, github.ExportMaxWebhookBodyBytes()+1)
	body[0] = '{'
	body[len(body)-1] = '}'

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(body)
	sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))

	req := httptest.NewRequest(http.MethodPost, "/github/webhook", bytes.NewReader(body))
	req.Header.Set("X-GitHub-Event", "ping")
	req.Header.Set("X-Hub-Signature-256", sig)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusRequestEntityTooLarge, resp.StatusCode)
}
