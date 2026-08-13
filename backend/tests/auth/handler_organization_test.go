package auth_test

// Тесты HTTP-обработчика аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/user"
)

func founderSession(t *testing.T, h *auth.Handler, svc *auth.AuthService, client *ent.Client, ctx context.Context) string {
	t.Helper()
	seedOrganization(t, svc, ctx)
	founder, err := client.User.Query().
		Where(user.SystemRoleEQ(user.SystemRoleFounder)).
		Only(ctx)
	require.NoError(t, err)
	pair, err := svc.CreateSession(ctx, founder.ID, "browser", "127.0.0.1")
	require.NoError(t, err)
	return pair.AccessToken
}

// TestHandler_ListPendingUsers_FounderOK проверяет HTTP-слой: HTTP-обработчик список pending users founder ok.
func TestHandler_ListPendingUsers_FounderOK(t *testing.T) {
	h, svc, client, _ := newHandlerTestEnv(t)
	ctx := t.Context()
	token := founderSession(t, h, svc, client, ctx)

	org, err := svc.Org.GetOrganization(ctx)
	require.NoError(t, err)
	_, err = client.User.Create().
		SetEmail("pending-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusPending).
		SetOrganizationID(org.ID).
		SetPendingSince(time.Now().UTC()).
		Save(ctx)
	require.NoError(t, err)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/users/pending", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ApprovePendingUser_Success проверяет HTTP-слой: HTTP-обработчик approve pending user успех.
func TestHandler_ApprovePendingUser_Success(t *testing.T) {
	h, svc, client, _ := newHandlerTestEnv(t)
	ctx := t.Context()
	token := founderSession(t, h, svc, client, ctx)

	org, err := svc.Org.GetOrganization(ctx)
	require.NoError(t, err)
	pending, err := client.User.Create().
		SetEmail("approve-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusPending).
		SetOrganizationID(org.ID).
		SetPendingSince(time.Now().UTC()).
		Save(ctx)
	require.NoError(t, err)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/"+pending.ID.String()+"/approve", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_RejectPendingUser_NotFound проверяет HTTP-слой: HTTP-обработчик reject pending user not found.
func TestHandler_RejectPendingUser_NotFound(t *testing.T) {
	h, svc, client, _ := newHandlerTestEnv(t)
	ctx := t.Context()
	token := founderSession(t, h, svc, client, ctx)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodPost, "/api/admin/users/"+uuid.New().String()+"/reject", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// TestHandler_GetOrgInvite_FounderOK проверяет HTTP-слой: HTTP-обработчик получение org приглашение founder ok.
func TestHandler_GetOrgInvite_FounderOK(t *testing.T) {
	h, svc, client, _ := newHandlerTestEnv(t)
	ctx := t.Context()
	token := founderSession(t, h, svc, client, ctx)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodGet, "/api/admin/organization/invite", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
