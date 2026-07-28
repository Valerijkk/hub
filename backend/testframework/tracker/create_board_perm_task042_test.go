package tracker_test

// TASK-042 / P2-BE-12 / W1-A1-006: CreateBoard требует PermTaskEdit (viewer → 403).
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

// Mutation: вернуть boards.Post PermProjectView → viewer создаст доску (201).
func TestCreateBoard_ViewerForbidden_P2_BE_12(t *testing.T) {
	dsn := "file:trboardperm_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rbacSvc := rbac.NewRBACService(rbac.NewEntRepository(client), rbac.NewPermissionCache(rdb))
	ctx := context.Background()

	owner, err := client.User.Create().SetEmail("own-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
	require.NoError(t, err)
	viewer, err := client.User.Create().SetEmail("view-" + uuid.New().String()[:6] + "@ex.com").Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().SetName("Board Perm").SetKey("BP").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().SetUserID(owner.ID).SetProjectID(proj.ID).SetProjectRole("owner").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().SetUserID(viewer.ID).SetProjectID(proj.ID).SetProjectRole("viewer").Save(ctx)
	require.NoError(t, err)
	require.NoError(t, tracker.NewEntRepository(client, nil).SeedDefaultWorkflow(ctx, proj.ID))

	repo := tracker.NewEntRepository(client, nil)
	h := tracker.NewHandler(tracker.NewTrackerService(repo, nil), tracker.NewAutomationService(repo))
	app := fiber.New()
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", c.Get("X-Test-User"))
		return c.Next()
	}
	h.RegisterRoutes(app.Group("/api"), authMw, rbacSvc)

	body := `{"name":"Secret board"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+proj.ID.String()+"/boards/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User", viewer.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
