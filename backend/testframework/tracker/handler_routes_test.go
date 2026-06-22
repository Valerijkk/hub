package tracker_test

// Тесты HTTP-обработчика трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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

func newTrackerRoutesEnv(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	dsn := "file:trroutes_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
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
		SetEmail("tr-" + uuid.New().String()[:6] + "@ex.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Tracker Routes").
		SetKey("TR").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, tracker.NewEntRepository(client, nil).SeedDefaultWorkflow(ctx, proj.ID))

	repo := tracker.NewEntRepository(client, nil)
	svc := tracker.NewTrackerService(repo, nil)
	h := tracker.NewHandler(svc, tracker.NewAutomationService(repo))

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	return app, proj.ID
}

// TestHandler_RegisterRoutes_CreateDeleteIssue проверяет HTTP-слой: HTTP-обработчик register маршруты создание удаление issue.
func TestHandler_RegisterRoutes_CreateDeleteIssue(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String()

	body := `{"title":"Route issue","issue_type":"task","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base+"/issues/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created tracker.IssueResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	req = httptest.NewRequest(http.MethodDelete, base+"/issues/"+created.IssueKey, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_RegisterRoutes_CustomFields проверяет HTTP-слой: HTTP-обработчик register маршруты custom fields.
func TestHandler_RegisterRoutes_CustomFields(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	base := "/api/projects/" + projectID.String() + "/custom-fields"

	body := `{"name":"Risk","field_type":"text","required":false}`
	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_RegisterRoutes_ListWorkflow проверяет HTTP-слой: HTTP-обработчик register маршруты список workflow.
func TestHandler_RegisterRoutes_ListWorkflow(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/workflow/statuses", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_SearchIssuesJQL_InvalidBody проверяет отклонение при HTTP-обработчик search issues jql некорректный body.
func TestHandler_SearchIssuesJQL_InvalidBody(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/issues/jql", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_RegisterRoutes_CreateBoard проверяет HTTP-слой: HTTP-обработчик register маршруты создание board.
func TestHandler_RegisterRoutes_CreateBoard(t *testing.T) {
	app, projectID := newTrackerRoutesEnv(t)
	body := `{"name":"Sprint board"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/boards/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

// TestHandler_NewHandler проверяет HTTP-слой: HTTP-обработчик new HTTP-обработчик.
func TestHandler_NewHandler(t *testing.T) {
	svc, repo := newTestService()
	h := tracker.NewHandler(svc, tracker.NewAutomationService(repo))
	require.NotNil(t, h)
}
