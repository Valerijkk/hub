//go:build hubtest

package ws_test

import (
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/ws"
)

// TestWSHandler_NilAuthorizer_RefusesUpgrade — W3-A4-003 / P2-BE-10:
// authz==nil → fail-closed (503), не allowAllAuthorizer.
func TestWSHandler_NilAuthorizer_RefusesUpgrade(t *testing.T) {
	hub := ws.NewHub()
	svc, magic, _ := newWSAuthService(t)
	token := wsAccessToken(t, svc, magic)

	app := fiber.New()
	app.Get("/ws", ws.Handler(hub, svc, nil))

	req := httptest.NewRequest(fiber.MethodGet, "/ws?token="+token, nil)
	resp, err := app.Test(req, 5*time.Second)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusServiceUnavailable, resp.StatusCode)
}
