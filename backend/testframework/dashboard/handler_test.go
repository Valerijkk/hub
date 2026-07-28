package dashboard_test
import (
	"github.com/valerii/hub/backend/internal/dashboard"
	"bytes"
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
	"github.com/valerii/hub/backend/ent/user"
)

func testDashboardApp(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	svc, _, _, userID := newDashboardTestEnv(t)
	h := dashboard.NewHandler(svc)

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	})
	h.RegisterRoutes(app, func(c fiber.Ctx) error { return c.Next() }, nil)
	return app, userID
}

func TestHandler_SaveLayout_Validation(t *testing.T) {
	app, _ := testDashboardApp(t)

	t.Run("missing dashboard_type", func(t *testing.T) {
		body, _ := json.Marshal(dashboard.SaveLayoutRequest{Items: []dashboard.LayoutItem{}})
		req := httptest.NewRequest(http.MethodPut, "/dashboard/layout", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid json", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPut, "/dashboard/layout", bytes.NewReader([]byte("{")))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestHandler_GetLayout_InvalidProjectID(t *testing.T) {
	app, _ := testDashboardApp(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/layout?type=project&project_id=bad-id", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ApplyTemplate_Validation(t *testing.T) {
	app, _ := testDashboardApp(t)

	t.Run("missing template", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"project_id": uuid.New().String()})
		req := httptest.NewRequest(http.MethodPost, "/dashboard/templates/apply", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("invalid project_id", func(t *testing.T) {
		body, _ := json.Marshal(map[string]string{"project_id": "x", "template": "dev"})
		req := httptest.NewRequest(http.MethodPost, "/dashboard/templates/apply", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestHandler_ListTemplates(t *testing.T) {
	app, _ := testDashboardApp(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/templates", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var templates map[string][]dashboard.LayoutItem
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&templates))
	assert.Contains(t, templates, "dev")
}

func TestHandler_GetLayout_SuccessEmpty(t *testing.T) {
	app, _ := testDashboardApp(t)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/layout?type=personal", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var layout dashboard.LayoutDTO
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&layout))
	assert.Equal(t, "personal", layout.DashboardType)
	assert.Empty(t, layout.Items)
}

func TestHandler_Personal_ReturnsData(t *testing.T) {
	svc, _, ctx, userID := newDashboardTestEnv(t)
	h := dashboard.NewHandler(svc)

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	})
	h.RegisterRoutes(app, func(c fiber.Ctx) error { return c.Next() }, nil)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/personal", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var data dashboard.PersonalData
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&data))
	assert.NotNil(t, data.MyTasks)
	_ = ctx
}

func TestHandler_Admin_ForbiddenForMember(t *testing.T) {
	svc, _, ctx, userID := newDashboardTestEnv(t)
	h := dashboard.NewHandler(svc)

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	})
	h.RegisterRoutes(app, func(c fiber.Ctx) error { return c.Next() }, nil)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/admin", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	_ = ctx
}

func TestHandler_Admin_AllowedForAdmin(t *testing.T) {
	svc, client, ctx, userID := newDashboardTestEnv(t)
	_, err := client.User.UpdateOneID(userID).SetSystemRole(user.SystemRoleAdmin).Save(ctx)
	require.NoError(t, err)

	h := dashboard.NewHandler(svc)
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	})
	h.RegisterRoutes(app, func(c fiber.Ctx) error { return c.Next() }, nil)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/admin", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Admin_AllowedForFounder_A7_002 — founder must reach admin dashboard (ORG-04).
func TestHandler_Admin_AllowedForFounder_A7_002(t *testing.T) {
	svc, client, ctx, userID := newDashboardTestEnv(t)
	_, err := client.User.UpdateOneID(userID).SetSystemRole(user.SystemRoleFounder).Save(ctx)
	require.NoError(t, err)

	h := dashboard.NewHandler(svc)
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	})
	h.RegisterRoutes(app, func(c fiber.Ctx) error { return c.Next() }, nil)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/admin", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_SaveLayout_Success(t *testing.T) {
	app, _ := testDashboardApp(t)

	body, _ := json.Marshal(dashboard.SaveLayoutRequest{
		DashboardType: "personal",
		Items:         []dashboard.LayoutItem{{I: "w1", WidgetType: "my-tasks", W: 6, H: 4}},
	})
	req := httptest.NewRequest(http.MethodPut, "/dashboard/layout", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandler_Project_ReturnsStats(t *testing.T) {
	repo, _, ctx, userID, projectID := newDashboardRepoEnv(t)
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	svc := dashboard.NewDashboardService(repo, rdb, nil)
	h := dashboard.NewHandler(svc)

	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	})
	h.RegisterRoutes(app, func(c fiber.Ctx) error { return c.Next() }, nil)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/project/"+projectID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var stats dashboard.ProjectStats
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&stats))
	assert.NotNil(t, stats.TasksByStatus)
	_ = ctx
}

func TestHandler_ApplyTemplate_Success(t *testing.T) {
	svc, client, ctx, userID := newDashboardTestEnv(t)
	proj, err := client.Project.Create().SetName("Apply").SetKey("APL").Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(userID).
		SetProjectID(proj.ID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	h := dashboard.NewHandler(svc)
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID)
		return c.Next()
	})
	h.RegisterRoutes(app, func(c fiber.Ctx) error { return c.Next() }, nil)

	body, _ := json.Marshal(map[string]string{
		"project_id": proj.ID.String(),
		"template":   "qa",
	})
	req := httptest.NewRequest(http.MethodPost, "/dashboard/templates/apply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestGetUserID_Variants(t *testing.T) {
	t.Run("uuid type", func(t *testing.T) {
		id := uuid.New()
		var got uuid.UUID
		app := fiber.New()
		app.Use(func(c fiber.Ctx) error {
			c.Locals("userID", id)
			return c.Next()
		})
		app.Get("/u", func(c fiber.Ctx) error {
			var err error
			got, err = dashboard.ExportGetUserID(c)
			return err
		})
		req := httptest.NewRequest(http.MethodGet, "/u", nil)
		_, err := app.Test(req)
		require.NoError(t, err)
		assert.Equal(t, id, got)
	})

	t.Run("string type", func(t *testing.T) {
		id := uuid.New()
		var got uuid.UUID
		app2 := fiber.New()
		app2.Use(func(c fiber.Ctx) error {
			c.Locals("userID", id.String())
			return c.Next()
		})
		app2.Get("/u", func(c fiber.Ctx) error {
			var err error
			got, err = dashboard.ExportGetUserID(c)
			return err
		})
		req := httptest.NewRequest(http.MethodGet, "/u", nil)
		_, err := app2.Test(req)
		require.NoError(t, err)
		assert.Equal(t, id, got)
	})
}

func TestHandler_Project_Success(t *testing.T) {
	svc, client, ctx, userID := newDashboardTestEnv(t)
	proj, err := client.Project.Create().
		SetName("Stats Project").
		SetKey("STP").
		Save(ctx)
	require.NoError(t, err)
	_, err = client.ProjectMember.Create().
		SetUserID(userID).
		SetProjectID(proj.ID).
		SetProjectRole("developer").
		Save(ctx)
	require.NoError(t, err)

	h := dashboard.NewHandler(svc)
	app := fiber.New()
	app.Use(func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	})
	h.RegisterRoutes(app, func(c fiber.Ctx) error { return c.Next() }, nil)

	req := httptest.NewRequest(http.MethodGet, "/dashboard/project/"+proj.ID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
