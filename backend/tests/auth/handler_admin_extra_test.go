package auth_test

// Тесты HTTP-обработчика аутентификации: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/auth"
	"bytes"
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

// TestHandler_SetRegistrationMode_InvalidJSON проверяет отклонение при HTTP-обработчик set регистрация mode некорректный json.
func TestHandler_SetRegistrationMode_InvalidJSON(t *testing.T) {
	h, svc, client, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	admin, err := client.User.Create().
		SetEmail("admin-json-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, admin.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, auth.RequireAuth(svc))

	req := httptest.NewRequest(http.MethodPut, "/api/admin/registration-mode", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
