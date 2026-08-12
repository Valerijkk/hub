//go:build hubtest && !release

package dev_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/dev"
)

// TestLogin_RejectsNonActive_Task067 — W7-A5-012 / P3-BE-16:
// deactivated user /api/dev/login → 403.
// Мутация: убрать StatusActive check → 200 + JWT.
func TestLogin_RejectsNonActive_Task067(t *testing.T) {
	app, client := newDevHandlerApp(t, true)
	ctx := context.Background()

	email := "banned-" + uuid.New().String()[:8] + "@test.local"
	_, err := client.User.Create().
		SetEmail(email).
		SetName("Banned").
		SetStatus(user.StatusDeactivated).
		Save(ctx)
	require.NoError(t, err)

	body, _ := json.Marshal(dev.DevLoginRequest{Email: email})
	req := httptest.NewRequest(http.MethodPost, "/api/dev/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}

// TestSeedData_NoTelegramLoginCopy_Task067 — W7-A5-014 / P3-BE-16:
// демо без «Telegram вход/авторизация».
func TestSeedData_NoTelegramLoginCopy_Task067(t *testing.T) {
	var titles []string
	for _, iss := range dev.ExportDemoIssues() {
		titles = append(titles, iss.Title)
	}
	for _, tc := range dev.ExportDemoTestCases() {
		titles = append(titles, tc.Title)
	}
	for _, title := range titles {
		lower := strings.ToLower(title)
		assert.NotContains(t, lower, "telegram авторизация")
		assert.NotContains(t, lower, "telegram вход")
		assert.NotContains(t, lower, "вход через telegram")
	}
	joined := strings.ToLower(strings.Join(titles, " "))
	assert.Contains(t, joined, "telegram")
	assert.Contains(t, joined, "уведомлен")
}
