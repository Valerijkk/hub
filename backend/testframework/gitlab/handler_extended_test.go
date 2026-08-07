package gitlab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/valerii/hub/backend/internal/gitlab"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/rbac"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}

type gitlabUpstreamRequest struct {
	method      string
	scheme      string
	host        string
	escapedPath string
	rawQuery    string
	token       string
	url         string
}

type gitlabRoutesEnv struct {
	app       *fiber.App
	ownerID   uuid.UUID
	projectID uuid.UUID
	requests  <-chan gitlabUpstreamRequest
}

func newGitLabRoutesEnv(t *testing.T) gitlabRoutesEnv {
	t.Helper()
	dsn := "file:glroutes_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
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
		SetName("GL Routes").
		SetKey(testGitLabProjectKey()).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	requests := make(chan gitlabUpstreamRequest, 1)
	httpClient := &http.Client{Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
		requests <- gitlabUpstreamRequest{
			method:      req.Method,
			scheme:      req.URL.Scheme,
			host:        req.URL.Host,
			escapedPath: req.URL.EscapedPath(),
			rawQuery:    req.URL.RawQuery,
			token:       req.Header.Get("PRIVATE-TOKEN"),
			url:         req.URL.String(),
		}

		return &http.Response{
			StatusCode: http.StatusOK,
			Status:     "200 OK",
			Header:     http.Header{"Content-Type": []string{"application/json"}},
			Body:       io.NopCloser(strings.NewReader("[]")),
			Request:    req,
		}, nil
	})}

	repo := gitlab.NewEntRepository(client, nil)
	svc := gitlab.NewGitLabServiceWithHTTPClient(repo, nil, httpClient)
	h := gitlab.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	return gitlabRoutesEnv{app: app, ownerID: owner.ID, projectID: proj.ID, requests: requests}
}

func TestHandler_RegisterRoutes_WebhookAndIntegrations(t *testing.T) {
	env := newGitLabRoutesEnv(t)
	base := "/api/projects/" + env.projectID.String() + "/gitlab"

	// gitlab_url не задаём: loopback из httptest блокируется ValidateOutboundURL (SEC-04);
	// дефолт https://gitlab.com проходит SSRF-guard. webhook_secret обязателен (A2W8-2-007).
	payload := `{"project_path":"group/routes","access_token":"glpat-routes","webhook_secret":"route-secret"}`
	req := httptest.NewRequest(http.MethodPost, base+"/integrations", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created gitlab.IntegrationResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEmpty(t, created.ID)
	req = httptest.NewRequest(http.MethodGet, base+"/integrations", nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var listed []gitlab.IntegrationResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&listed))
	require.Len(t, listed, 1)

	req = httptest.NewRequest(http.MethodGet, base+"/pipelines", nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var upstream gitlabUpstreamRequest
	select {
	case upstream = <-env.requests:
	default:
		require.FailNow(t, "GitLab pipelines request was not captured")
	}
	require.Equal(t, http.MethodGet, upstream.method)
	require.Equal(t, "https", upstream.scheme)
	require.Equal(t, "gitlab.com", upstream.host)
	require.Equal(t, "/api/v4/projects/group%2Froutes/pipelines", upstream.escapedPath)
	require.Equal(t, "per_page=20&order_by=id&sort=desc", upstream.rawQuery)
	require.Equal(t, "glpat-routes", upstream.token)
	require.Equal(t,
		"https://gitlab.com/api/v4/projects/group%2Froutes/pipelines?per_page=20&order_by=id&sort=desc",
		upstream.url,
	)

	req = httptest.NewRequest(http.MethodDelete, base+"/integrations/"+created.ID, nil)
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNoContent, resp.StatusCode)

	body, _ := json.Marshal(gitlab.WebhookPushEvent{
		ObjectKind: "push",
		Project:    gitlab.WebhookProject{PathWithNamespace: "unknown/project"},
	})
	req = httptest.NewRequest(http.MethodPost, "/api/gitlab/webhook", bytes.NewReader(body))
	req.Header.Set("X-Gitlab-Event", "Push Hook")
	req.Header.Set("X-Gitlab-Token", "route-secret")
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestHandler_RegisterRoutes_InvalidProjectID(t *testing.T) {
	env := newGitLabRoutesEnv(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/not-uuid/gitlab/integrations", nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
