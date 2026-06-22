package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	cryptopkg "github.com/valerii/hub/backend/internal/platform/crypto"
	"github.com/valerii/hub/backend/internal/rbac"
)

func TestHandler_GetWorkflowRunDetail_HTTPIntegration(t *testing.T) {
	now := time.Now().UTC()
	ghSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/actions/runs/55") && !strings.Contains(r.URL.Path, "/jobs") {
			_, _ = w.Write([]byte(`{
				"id":55,"name":"Lint","status":"completed","conclusion":"success",
				"head_branch":"main","event":"push",
				"html_url":"https://github.com/acme/hub/actions/runs/55",
				"created_at":"` + now.Format(time.RFC3339) + `",
				"updated_at":"` + now.Format(time.RFC3339) + `",
				"head_sha":"sha55","run_number":2
			}`))
			return
		}
		if strings.Contains(r.URL.Path, "/jobs") {
			_, _ = w.Write([]byte(`{"jobs":[]}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(ghSrv.Close)

	dsn := "file:ghwf_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	cipher, err := cryptopkg.NewCipherFromKey(cryptopkg.DeriveDevKey())
	require.NoError(t, err)
	repo := github.NewEntRepository(client, cipher)
	svc := github.NewGitHubServiceWithHTTPClient(repo, nil, ghSrv.Client(), ghSrv.URL)
	preview := github.NewLinkPreviewService(repo, nil)
	h := github.NewHandler(svc, preview)

	app := fiber.New()
	api := app.Group("/api")

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })
	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))

	ownerID := uuid.New()
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", ownerID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	ctx := context.Background()
	_, err = client.User.Create().SetID(ownerID).SetEmail("wf-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().SetName("WF").SetKey(testGitHubProjectKey()).Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(ownerID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	payload := `{"repo_owner":"acme","repo_name":"hub","access_token":"ghp_wf","webhook_secret":"wh"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+proj.ID.String()+"/github/integrations", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+proj.ID.String()+"/github/workflows/55", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var detail github.WorkflowRunDetailResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&detail))
	assert.Equal(t, int64(55), detail.ID)
	assert.Equal(t, "Lint", detail.Name)
}

func TestHandler_CreateBranch_MissingIssueKey(t *testing.T) {
	env := newGitHubRoutesEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/api/issues//github/create-branch", bytes.NewReader([]byte(`{"base_branch":"main"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestHandler_ListGitHubLinks_NotFound(t *testing.T) {
	env := newGitHubRoutesEnv(t)
	req := httptest.NewRequest(http.MethodGet, "/api/issues/MISSING-99/github/links", nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.True(t, resp.StatusCode >= fiber.StatusBadRequest)
}
