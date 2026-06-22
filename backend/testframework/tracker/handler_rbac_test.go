package tracker_test

// FA-RBAC-FIBER3: регрессия порядка middleware Fiber v3 — RBAC после handler.
// Посторонний пользователь не может создавать задачи в чужом проекте (IDOR).
import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/rbac"
	"github.com/valerii/hub/backend/internal/tracker"
)

func newTrackerRBACEnv(t *testing.T) (app *fiber.App, projectID, memberID, outsiderID uuid.UUID) {
	t.Helper()
	dsn := "file:trrbac_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
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
		SetEmail("tr-member-" + uuid.New().String()[:6] + "@ex.com").
		Save(ctx)
	require.NoError(t, err)

	outsider, err := client.User.Create().
		SetEmail("tr-outsider-" + uuid.New().String()[:6] + "@ex.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Tracker RBAC").
		SetKey("TB").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetUserID(member.ID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, tracker.NewEntRepository(client, nil).SeedDefaultWorkflow(ctx, proj.ID))

	repo := tracker.NewEntRepository(client, nil)
	svc := tracker.NewTrackerService(repo, nil)
	h := tracker.NewHandler(svc, tracker.NewAutomationService(repo))

	app = fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", c.Get("X-Test-User"))
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	return app, proj.ID, member.ID, outsider.ID
}

func TestCreateIssue_RBAC_MemberAllowed(t *testing.T) {
	app, projectID, memberID, _ := newTrackerRBACEnv(t)
	base := "/api/projects/" + projectID.String() + "/issues/"

	body := `{"title":"RBAC issue","issue_type":"task","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestCreateIssue_RBAC_OutsiderForbidden(t *testing.T) {
	app, projectID, _, outsiderID := newTrackerRBACEnv(t)
	base := "/api/projects/" + projectID.String() + "/issues/"

	body := `{"title":"IDOR attempt","issue_type":"task","priority":"medium"}`
	req := httptest.NewRequest(http.MethodPost, base, bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", outsiderID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestTrackerRoutes_RBAC_OutsiderForbidden(t *testing.T) {
	app, projectID, _, outsiderID := newTrackerRBACEnv(t)
	base := "/api/projects/" + projectID.String()

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"update issue", http.MethodPut, base + "/issues/TB-1", `{"title":"x"}`},
		{"create board", http.MethodPost, base + "/boards/", `{"name":"B","type":"kanban"}`},
		{"create label", http.MethodPost, base + "/labels/", `{"name":"bug","color":"#f00"}`},
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
