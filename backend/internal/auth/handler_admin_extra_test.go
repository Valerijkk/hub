package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/user"
)

func TestHandler_SetRegistrationMode_InvalidJSON(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()

	admin, err := svc.db.User.Create().
		SetEmail("admin-json-" + uuid.New().String()[:6] + "@example.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	pair, err := svc.CreateSession(ctx, admin.ID, "browser", "127.0.0.1")
	require.NoError(t, err)

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, RequireAuth(svc))

	req := httptest.NewRequest(http.MethodPut, "/api/admin/registration-mode", bytes.NewReader([]byte(`{`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+pair.AccessToken)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_LoginTelegram_InviteOnly(t *testing.T) {
	h, svc, _ := newHandlerTestEnv(t)
	ctx := t.Context()
	require.NoError(t, svc.SetRegistrationMode(ctx, "invite_only"))

	app := testApp(http.MethodPost, "/auth/telegram", h.LoginTelegram)

	botToken := "bot"
	authDate := strconv.FormatInt(time.Now().Unix(), 10)
	userJSON := `{"id":999888,"first_name":"New","username":"newtg"}`
	params := url.Values{}
	params.Set("auth_date", authDate)
	params.Set("user", userJSON)
	params.Set("query_id", "q1")
	params.Set("hash", computeTelegramHash(t, botToken, params))

	body, _ := json.Marshal(TelegramAuthRequest{InitData: params.Encode()})
	req := httptest.NewRequest(http.MethodPost, "/auth/telegram", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
