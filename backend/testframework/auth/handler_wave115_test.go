package auth_test

// Тесты HTTP-обработчика аутентификации (волна 115): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
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

// TestHandler_ApprovePendingUser_InvalidUserID проверяет отклонение при HTTP-обработчик approve pending user некорректный user id.
func TestHandler_ApprovePendingUser_InvalidUserID(t *testing.T) {
	h, svc, client, _ := newHandlerTestEnv(t)
	token := founderSession(t, h, svc, client, t.Context())

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/not-a-uuid/approve", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_ApprovePendingUser_NotFound проверяет HTTP-слой: HTTP-обработчик approve pending user not found.
func TestHandler_ApprovePendingUser_NotFound(t *testing.T) {
	h, svc, client, _ := newHandlerTestEnv(t)
	token := founderSession(t, h, svc, client, t.Context())

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/"+uuid.New().String()+"/approve", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

// TestHandler_ListPendingUsers_ForbiddenForMember проверяет отклонение при HTTP-обработчик список pending users запрещённый for участник.
func TestHandler_ListPendingUsers_ForbiddenForMember(t *testing.T) {
	h, svc, client, _ := newHandlerTestEnv(t)
	ctx := t.Context()
	seedOrganization(t, svc, ctx)

	member, err := client.User.Create().
		SetEmail("member-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, member.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users/pending", nil)
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}
