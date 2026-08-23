package mail_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/mail"
)

type mockBroadcastSender struct {
	sent   int
	failOn int // 1-based call index to fail; 0 = never
}

func (m *mockBroadcastSender) SendAlert(_ context.Context, _ mail.SendRequest) error {
	m.sent++
	if m.failOn > 0 && m.sent == m.failOn {
		return mail.ErrInvalidRequest
	}
	return nil
}

func testAdminMailApp(t *testing.T, isAdmin bool) (*fiber.App, *ent.Client, *mockBroadcastSender) {
	t.Helper()
	dsn := "file:mailadmin_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	adminUser, err := client.User.Create().
		SetEmail("admin-" + uuid.New().String()[:6] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	store := mail.NewEntSettingsStore(client)
	adminSvc := mail.NewAdminService(store, "http://127.0.0.1:3001", "alerts@hub.local", false)
	sender := &mockBroadcastSender{}
	isAdminFn := func(_ fiber.Ctx, id uuid.UUID) (bool, error) {
		return isAdmin && id == adminUser.ID, nil
	}
	h := mail.NewAdminHandler(adminSvc, store, sender, isAdminFn)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", adminUser.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, client, sender
}

func TestAdminHandler_SettingsCRUD(t *testing.T) {
	app, _, _ := testAdminMailApp(t, true)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/mail/settings", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	enabled := true
	limit := 30
	updateBody, _ := json.Marshal(mail.UpdateSettingsRequest{
		HubMailURL:    "http://mail:3001",
		FromAlerts:    "Hub <noreply@test.local>",
		AlertsEnabled: &enabled,
		DailyLimit:    &limit,
	})
	req = httptest.NewRequest(http.MethodPut, "/api/admin/mail/settings", bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var updated mail.SettingsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	assert.Equal(t, "http://mail:3001", updated.HubMailURL)
	assert.True(t, updated.AlertsEnabled)
}

func TestAdminHandler_BroadcastDryRun(t *testing.T) {
	app, client, _ := testAdminMailApp(t, true)
	ctx := context.Background()

	_, err := client.User.Create().
		SetEmail("member-" + uuid.New().String()[:6] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	body, _ := json.Marshal(mail.BroadcastRequest{
		Subject:  "Team update",
		HTMLBody: "<p>Hello team</p>",
		DryRun:   true,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/mail/broadcast", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result mail.BroadcastResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.True(t, result.DryRun)
	assert.GreaterOrEqual(t, result.Recipients, 2)
	assert.Equal(t, 0, result.Failed)
}

func TestAdminHandler_BroadcastFailedCount(t *testing.T) {
	// W7-A1-004: partial SendAlert failure → Failed > 0, Sent for successes.
	app, client, sender := testAdminMailApp(t, true)
	ctx := context.Background()
	_, err := client.User.Create().
		SetEmail("member-" + uuid.New().String()[:6] + "@test.com").
		Save(ctx)
	require.NoError(t, err)
	sender.failOn = 1

	body, _ := json.Marshal(mail.BroadcastRequest{
		Subject:  "Team update",
		HTMLBody: "<p>Hello team</p>",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/admin/mail/broadcast", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var result mail.BroadcastResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&result))
	assert.Equal(t, 1, result.Failed)
	assert.Equal(t, result.Recipients-1, result.Sent)
}

func TestAdminHandler_ForbiddenNonAdmin(t *testing.T) {
	app, _, _ := testAdminMailApp(t, false)

	req := httptest.NewRequest(http.MethodGet, "/api/admin/mail/settings", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusForbidden, resp.StatusCode)
}

func TestAdminHandler_Unauthorized(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:mailnoauth_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })

	store := mail.NewEntSettingsStore(client)
	h := mail.NewAdminHandler(mail.NewAdminService(store, "", "", false), store, nil, func(_ fiber.Ctx, _ uuid.UUID) (bool, error) {
		return true, nil
	})

	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodGet, "/api/admin/mail/settings", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusUnauthorized, resp.StatusCode)
}
