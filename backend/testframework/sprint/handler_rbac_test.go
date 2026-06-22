package sprint_test

// FA-RBAC-FIBER3: регрессия порядка middleware Fiber v3 для спринтов.
import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/internal/rbac"
	"github.com/valerii/hub/backend/internal/sprint"
)

func newSprintRBACEnv(t *testing.T) (app *fiber.App, projectID, memberID, outsiderID uuid.UUID) {
	t.Helper()
	dsn := "file:sprrbac_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
	ctx := context.Background()

	member, err := client.User.Create().
		SetEmail("spr-member-" + uuid.New().String()[:6] + "@ex.com").
		Save(ctx)
	require.NoError(t, err)

	outsider, err := client.User.Create().
		SetEmail("spr-outsider-" + uuid.New().String()[:6] + "@ex.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Sprint RBAC").
		SetKey("SB").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(member.ID).
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

	app = fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", c.Get("X-Test-User"))
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	return app, proj.ID, member.ID, outsider.ID
}

func TestCreateSprint_RBAC_MemberAllowed(t *testing.T) {
	app, projectID, memberID, _ := newSprintRBACEnv(t)
	base := "/api/projects/" + projectID.String() + "/sprints/"

	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().Add(14 * 24 * time.Hour).Format(time.RFC3339)
	body := `{"name":"S1","goal":"g","start_date":"` + start + `","end_date":"` + end + `"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestCreateSprint_RBAC_OutsiderForbidden(t *testing.T) {
	app, projectID, _, outsiderID := newSprintRBACEnv(t)
	base := "/api/projects/" + projectID.String() + "/sprints/"

	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().Add(14 * 24 * time.Hour).Format(time.RFC3339)
	body := `{"name":"IDOR","goal":"g","start_date":"` + start + `","end_date":"` + end + `"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", outsiderID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestSprintRoutes_RBAC_OutsiderForbidden(t *testing.T) {
	app, projectID, _, outsiderID := newSprintRBACEnv(t)
	base := "/api/projects/" + projectID.String() + "/sprints"

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"update sprint", http.MethodPut, base + "/" + uuid.New().String(), `{"name":"x"}`},
		{"start sprint", http.MethodPost, base + "/" + uuid.New().String() + "/start", `{}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, bytes.NewReader([]byte(tc.body)))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("X-Test-User", outsiderID.String())
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		})
	}
}
