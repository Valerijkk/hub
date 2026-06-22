//go:build !release

package dev_test
import (
	"github.com/valerii/hub/backend/internal/dev"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/auth"
)

func newDevHandlerApp(t *testing.T, devMode bool) (*fiber.App, *ent.Client) {
	t.Helper()

	dsn := "file:devhandler_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	magic := auth.NewMagicLinkService(rdb, nil, "dev@test.local")
	authSvc := auth.NewAuthService(
		auth.NewEntRepository(client),
		rdb, nil,
		"dev-handler-secret",
		15*time.Minute, 30*24*time.Hour,
		magic, "bot", nil,
	)

	h := dev.NewHandler(client, authSvc)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, devMode)
	return app, client
}

func TestHandler_DevModeGuard(t *testing.T) {
	app, _ := newDevHandlerApp(t, false)

	req := httptest.NewRequest(http.MethodGet, "/api/dev/status", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestHandler_Status_DevMode(t *testing.T) {
	app, _ := newDevHandlerApp(t, true)

	req := httptest.NewRequest(http.MethodGet, "/api/dev/status", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body map[string]bool
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.True(t, body["dev_mode"])
}

func TestHandler_Login_ValidationAndNotFound(t *testing.T) {
	app, _ := newDevHandlerApp(t, true)

	req := httptest.NewRequest(http.MethodPost, "/api/dev/login", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, _ := json.Marshal(dev.DevLoginRequest{Email: ""})
	req = httptest.NewRequest(http.MethodPost, "/api/dev/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)

	body, _ = json.Marshal(dev.DevLoginRequest{Email: "missing@dev.local"})
	req = httptest.NewRequest(http.MethodPost, "/api/dev/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestHandler_Login_Success(t *testing.T) {
	app, client := newDevHandlerApp(t, true)
	ctx := context.Background()

	email := "dev-login-" + uuid.New().String()[:8] + "@test.local"
	_, err := client.User.Create().SetEmail(email).SetName("Dev User").Save(ctx)
	require.NoError(t, err)

	body, _ := json.Marshal(dev.DevLoginRequest{Email: email})
	req := httptest.NewRequest(http.MethodPost, "/api/dev/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var result map[string]any
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.NotEmpty(t, result["access_token"])
	assert.NotEmpty(t, result["refresh_token"])
}

func TestHandler_Seed_Idempotent(t *testing.T) {
	app, _ := newDevHandlerApp(t, true)

	for i := 0; i < 2; i++ {
		req := httptest.NewRequest(http.MethodPost, "/api/dev/seed", nil)
		resp, err := app.Test(req, 30*time.Second)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	}
}
