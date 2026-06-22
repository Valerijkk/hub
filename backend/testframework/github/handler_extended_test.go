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

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/rbac"
)

type githubRoutesEnv struct {
	app       *fiber.App
	projectID uuid.UUID
}

func newGitHubRoutesEnv(t *testing.T) githubRoutesEnv {
	t.Helper()
	dsn := "file:ghroutes_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("own-" + uuid.New().String()[:6] + "@ex.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("GH Routes").
		SetKey(testGitHubProjectKey()).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	ghSrv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.Contains(r.URL.Path, "/actions/runs") {
			_, _ = w.Write([]byte(`{"total_count":0,"workflow_runs":[]}`))
			return
		}
		http.NotFound(w, r)
	}))
	t.Cleanup(ghSrv.Close)

	repo := github.NewEntRepository(client, nil)
	svc := github.NewGitHubServiceWithHTTPClient(repo, nil, ghSrv.Client(), ghSrv.URL)
	h := github.NewHandler(svc, nil)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	return githubRoutesEnv{app: app, projectID: proj.ID}
}

func TestHandler_RegisterRoutes_IntegrationsCRUD(t *testing.T) {
	env := newGitHubRoutesEnv(t)
	base := "/api/projects/" + env.projectID.String() + "/github"

	payload := `{"repo_owner":"acme","repo_name":"routes","access_token":"ghp-routes","webhook_secret":"gh-wh"}`
	req := httptest.NewRequest(http.MethodPost, base+"/integrations", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created github.IntegrationResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEmpty(t, created.ID)

	req = httptest.NewRequest(http.MethodGet, base+"/integrations", nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/workflows", nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, base+"/integrations/"+created.ID, nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)
}

func TestHandler_RegisterRoutes_WebhookMissingSignature(t *testing.T) {
	env := newGitHubRoutesEnv(t)

	req := httptest.NewRequest(http.MethodPost, "/api/github/webhook", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("X-GitHub-Event", "push")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_RegisterRoutes_InvalidProjectID(t *testing.T) {
	env := newGitHubRoutesEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/bad-id/github/integrations", nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
