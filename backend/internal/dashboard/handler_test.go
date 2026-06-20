package dashboard

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/user"
)

func testDashboardApp(t *testing.T) (*fiber.App, uuid.UUID) {
	t.Helper()
	svc, _, userID := newDashboardTestEnv(t)
	h := NewHandler(svc)

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
		body, _ := json.Marshal(SaveLayoutRequest{Items: []LayoutItem{}})
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

	var templates map[string][]LayoutItem
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

	var layout LayoutDTO
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&layout))
	assert.Equal(t, "personal", layout.DashboardType)
	assert.Empty(t, layout.Items)
}

func TestHandler_Personal_ReturnsData(t *testing.T) {
	svc, ctx, userID := newDashboardTestEnv(t)
	h := NewHandler(svc)

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

	var data PersonalData
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&data))
	assert.NotNil(t, data.MyTasks)
	_ = ctx
}

func TestHandler_Admin_ForbiddenForMember(t *testing.T) {
	svc, ctx, userID := newDashboardTestEnv(t)
	h := NewHandler(svc)

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
	svc, ctx, userID := newDashboardTestEnv(t)
	_, err := svc.db.User.UpdateOneID(userID).SetSystemRole(user.SystemRoleAdmin).Save(ctx)
	require.NoError(t, err)

	h := NewHandler(svc)
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

	body, _ := json.Marshal(SaveLayoutRequest{
		DashboardType: "personal",
		Items:         []LayoutItem{{I: "w1", WidgetType: "my-tasks", W: 6, H: 4}},
	})
	req := httptest.NewRequest(http.MethodPut, "/dashboard/layout", bytes.NewReader(body))
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
			got, err = getUserID(c)
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
			got, err = getUserID(c)
			return err
		})
		req := httptest.NewRequest(http.MethodGet, "/u", nil)
		_, err := app2.Test(req)
		require.NoError(t, err)
		assert.Equal(t, id, got)
	})
}
