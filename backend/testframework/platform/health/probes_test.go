//go:build hubtest

package health_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/platform/health"
)

func TestLive_AlwaysOK_A14_001(t *testing.T) {
	app := fiber.New()
	app.Get("/api/live", health.Live)

	req := httptest.NewRequest(http.MethodGet, "/api/live", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), `"status":"ok"`)
}

func TestReady_DepsOK_A14_001(t *testing.T) {
	app := fiber.New()
	ok := func(context.Context) error { return nil }
	app.Get("/api/ready", health.Ready(health.Deps{
		PingPostgres: ok,
		PingRedis:    ok,
		PingMongo:    ok,
		PingNATS:     ok,
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/ready", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), `"postgres":"ok"`)
	assert.Contains(t, string(body), `"nats":"ok"`)
}

func TestReady_DepFail503_A14_001(t *testing.T) {
	app := fiber.New()
	ok := func(context.Context) error { return nil }
	app.Get("/api/ready", health.Ready(health.Deps{
		PingPostgres: ok,
		PingRedis: func(context.Context) error {
			return errors.New("redis down")
		},
		PingMongo: ok,
		PingNATS:  ok,
	}))

	req := httptest.NewRequest(http.MethodGet, "/api/ready", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), `"redis":"error"`)
	assert.Contains(t, string(body), `"status":"degraded"`)
}

// TestReady_NoProbesConfigured_503 — W4-A4-007 / P2-BE-10:
// Deps{} (все nil) → 503 fail-closed, не 200 «ok».
func TestReady_NoProbesConfigured_503(t *testing.T) {
	app := fiber.New()
	app.Get("/api/ready", health.Ready(health.Deps{}))

	req := httptest.NewRequest(http.MethodGet, "/api/ready", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	assert.Contains(t, string(body), `"status":"degraded"`)
}
