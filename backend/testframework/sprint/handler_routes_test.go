package sprint_test
import (
	"github.com/valerii/hub/backend/internal/sprint"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/internal/rbac"
)

type sprintRoutesEnv struct {
	app       *fiber.App
	projectID uuid.UUID
	ownerID   uuid.UUID
}

func newSprintRoutesEnv(t *testing.T) sprintRoutesEnv {
	t.Helper()
	dsn := "file:sprroutes_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
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
		SetEmail("spr-" + uuid.New().String()[:6] + "@ex.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Sprint Routes").
		SetKey("RT").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(owner.ID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	repo := sprint.NewEntRepository(client)
	svc := sprint.NewSprintService(repo, rdb, nil)
	poker := sprint.NewPokerService(rdb, nil)
	h := sprint.NewHandler(svc, poker)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	return sprintRoutesEnv{app: app, projectID: proj.ID, ownerID: owner.ID}
}

func TestHandler_RegisterRoutes_UpdateAndAddIssues(t *testing.T) {
	env := newSprintRoutesEnv(t)
	base := "/api/projects/" + env.projectID.String() + "/sprints"

	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().AddDate(0, 0, 14).Format(time.RFC3339)
	body := fmt.Sprintf(`{"name":"Route Sprint","goal":"coverage","start_date":%q,"end_date":%q}`, start, end)
	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var created sprint.SprintResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	updateBody := `{"name":"Updated Sprint Name"}`
	req = httptest.NewRequest(http.MethodPut, base+"/"+created.ID, bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var updated sprint.SprintResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	assert.Equal(t, "Updated Sprint Name", updated.Name)

	issueID := uuid.New()
	moveBody := fmt.Sprintf(`{"issue_ids":[%q]}`, issueID.String())
	req = httptest.NewRequest(http.MethodPost, base+"/"+created.ID+"/issues", bytes.NewReader([]byte(moveBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandler_AddIssuesToSprint_InvalidIssueID(t *testing.T) {
	env := newSprintRoutesEnv(t)
	base := "/api/projects/" + env.projectID.String() + "/sprints"

	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().AddDate(0, 0, 7).Format(time.RFC3339)
	body := fmt.Sprintf(`{"name":"Bad issues","start_date":%q,"end_date":%q}`, start, end)
	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	var sp sprint.SprintResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sp))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, base+"/"+sp.ID+"/issues", bytes.NewReader([]byte(`{"issue_ids":["not-uuid"]}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_UpdateSprint_InvalidBody(t *testing.T) {
	h := sprint.NewHandler(sprint.NewSprintService(newMockRepository(), nil, nil), nil)
	app := fiber.New()
	app.Put("/api/projects/:projectId/sprints/:sprintId", h.UpdateSprint)

	req := httptest.NewRequest(http.MethodPut, "/api/projects/"+uuid.New().String()+"/sprints/"+uuid.New().String(), bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_RegisterRoutes_ListViaRBAC(t *testing.T) {
	env := newSprintRoutesEnv(t)
	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+env.projectID.String()+"/sprints", nil)
	resp, err := env.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
