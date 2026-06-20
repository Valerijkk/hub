package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/user"
)

func newAdminHandlerApp(t *testing.T, h *Handler, svc *AuthService) *fiber.App {
	t.Helper()
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, RequireAuth(svc))
	return app
}

func TestHandler_GetRegistrationMode_AdminOK(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	admin, err := svc.db.User.Create().
		SetEmail("admin-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, admin.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/registration-mode", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_GetRegistrationMode_Forbidden(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	member, err := svc.db.User.Create().
		SetEmail("member-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, member.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/registration-mode", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

func TestHandler_SetRegistrationMode_AdminOK(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	admin, err := svc.db.User.Create().
		SetEmail("admin-set-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, admin.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, RequireAuth(svc))

	body, _ := json.Marshal(SetRegistrationModeRequest{Mode: "invite_only"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/registration-mode", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	mode, err := svc.GetRegistrationMode(ctx)
	require.NoError(t, err)
	assert.Equal(t, "invite_only", mode)
}

func TestHandler_SetRegistrationMode_InvalidMode(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	admin, err := svc.db.User.Create().
		SetEmail("admin-bad-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, admin.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, RequireAuth(svc))

	body, _ := json.Marshal(SetRegistrationModeRequest{Mode: "closed"})
	req := httptest.NewRequest(http.MethodPut, "/api/admin/registration-mode", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestNewHandler(t *testing.T) {
	_, svc, _ := newHandlerTestEnv(t)
	got := NewHandler(svc, nil)
	require.NotNil(t, got)
	assert.Equal(t, svc, got.service)
}

func TestHandler_RegisterRoutes_Smoke(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodPost, "/api/auth/refresh", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
}
