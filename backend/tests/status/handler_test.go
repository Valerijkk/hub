package status_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/status"
)

func TestHandler_GetStatus_PublicAggregate_W7A5002(t *testing.T) {
	svc := status.NewService(
		func(ctx context.Context) error { return errors.New("down") },
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
	)
	h := status.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		if c.Get("X-Test-User-ID") == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "missing auth")
		}
		c.Locals("userID", c.Get("X-Test-User-ID"))
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	// Публичный: degraded, но без id компонентов (recon).
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	var pub status.Response
	require.NoError(t, json.Unmarshal(body, &pub))
	assert.Equal(t, "degraded", pub.Status)
	require.Len(t, pub.Components, 1)
	assert.Equal(t, "hub", pub.Components[0].ID)
	assert.NotContains(t, string(body), "postgres")
	assert.NotContains(t, string(body), `"redis"`)
	assert.NotContains(t, string(body), `"mongo"`)
	assert.NotContains(t, string(body), `"nats"`)

	// Без auth — details → 401.
	req = httptest.NewRequest(http.MethodGet, "/api/status/details", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// С auth — полный разбор.
	req = httptest.NewRequest(http.MethodGet, "/api/status/details", nil)
	req.Header.Set("X-Test-User-ID", "user-1")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)

	body, err = io.ReadAll(resp.Body)
	require.NoError(t, err)
	var det status.Response
	require.NoError(t, json.Unmarshal(body, &det))
	require.NotEmpty(t, det.Components)
	byID := map[string]string{}
	for _, c := range det.Components {
		byID[c.ID] = c.Status
	}
	assert.Equal(t, "outage", byID["postgres"])
}

func TestHandler_GetStatus_Operational(t *testing.T) {
	svc := status.NewService(
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
		func(ctx context.Context) error { return nil },
	)
	h := status.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	h.RegisterRoutes(api, func(c fiber.Ctx) error { return c.Next() })

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
