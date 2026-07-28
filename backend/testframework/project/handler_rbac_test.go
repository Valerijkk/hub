package project_test

// TASK-009: stranger без membership получает 403 на mutate project.
import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/project"
)

func newProjectRBACDenialEnv(t *testing.T) (app *fiber.App, projectID, ownerID, outsiderID uuid.UUID) {
	t.Helper()
	dsn := "file:prrbac_deny_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@test.com").
		SetName("Owner").
		Save(ctx)
	require.NoError(t, err)

	outsider, err := client.User.Create().
		SetEmail("out-" + uuid.New().String()[:8] + "@test.com").
		SetName("Outsider").
		Save(ctx)
	require.NoError(t, err)

	repo := project.NewEntRepository(client, nil)
	svc := project.NewProjectService(repo, nil, nil)
	created, err := svc.CreateProject(ctx, owner.ID, "RBAC Proj", "RBP", "")
	require.NoError(t, err)
	projectID, err = uuid.Parse(created.ID)
	require.NoError(t, err)

	h := project.NewHandler(svc)
	rbacSvc := newRBACServiceForClient(t, client)

	app = fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", c.Get("X-Test-User"))
		return c.Next()
	}
	h.RegisterRoutes(api, authMw, rbacSvc)

	return app, projectID, owner.ID, outsider.ID
}

func TestProjectRoutes_RBAC_OutsiderForbidden(t *testing.T) {
	app, projectID, _, outsiderID := newProjectRBACDenialEnv(t)
	base := "/api/projects/" + projectID.String()

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{"delete", http.MethodDelete, base, ""},
		{"archive", http.MethodPost, base + "/archive", ""},
		{"update", http.MethodPut, base, `{"name":"hacked"}`},
		{"add member", http.MethodPost, base + "/members", `{"user_id":"` + uuid.New().String() + `","role":"developer"}`},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var req *http.Request
			if tc.body != "" {
				req = httptest.NewRequest(tc.method, tc.path, bytes.NewReader([]byte(tc.body)))
				req.Header.Set("Content-Type", "application/json")
			} else {
				req = httptest.NewRequest(tc.method, tc.path, nil)
			}
			req.Header.Set("X-Test-User", outsiderID.String())
			resp, err := app.Test(req)
			require.NoError(t, err)
			defer resp.Body.Close()
			require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
		})
	}
}

func TestProjectRoutes_RBAC_OwnerAllowedDelete(t *testing.T) {
	app, projectID, ownerID, _ := newProjectRBACDenialEnv(t)
	req := httptest.NewRequest(http.MethodDelete, "/api/projects/"+projectID.String(), nil)
	req.Header.Set("X-Test-User", ownerID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)
}
