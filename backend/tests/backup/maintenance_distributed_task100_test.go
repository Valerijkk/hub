package backup_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/backup"
)

// TestMaintenanceGate_DistributedRedis_Task100 — две «реплики» с общим Redis:
// Set на A → Enabled/503 на B (W7-A4-003 / U04).
//
// Mutation: вернуть Enabled только на local atomic → assert на gateB падает.
func TestMaintenanceGate_DistributedRedis_Task100(t *testing.T) {
	mr := miniredis.RunT(t)
	rdbA := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	rdbB := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdbA.Close(); _ = rdbB.Close() })

	gateA := backup.NewRedisMaintenanceGate(rdbA)
	gateB := backup.NewRedisMaintenanceGate(rdbB)
	ctx := context.Background()

	require.False(t, gateA.Enabled(ctx))
	require.False(t, gateB.Enabled(ctx), "replica B idle")

	gateA.Set(ctx, true)
	require.True(t, gateA.Enabled(ctx))
	require.True(t, gateB.Enabled(ctx), "U04: replica B must see Redis flag set by A")

	// Middleware на «реплике B» отдаёт 503, health — bypass.
	appB := fiber.New()
	apiB := appB.Group("/api")
	apiB.Use(backup.MaintenanceMiddleware(gateB))
	apiB.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	apiB.Get("/projects", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	resp, err := appB.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	assert.Contains(t, string(body), "maintenance")

	req = httptest.NewRequest(http.MethodGet, "/api/health", nil)
	resp, err = appB.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	gateA.Set(ctx, false)
	require.False(t, gateB.Enabled(ctx), "U04: clear on A clears B")

	req = httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	resp, err = appB.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
}

// TestMaintenanceGate_LeaseTTL_Task100 — TTL safety lease истекает без явного clear.
func TestMaintenanceGate_LeaseTTL_Task100(t *testing.T) {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	gate := backup.NewRedisMaintenanceGate(rdb)
	ctx := context.Background()
	gate.Set(ctx, true)
	require.True(t, gate.Enabled(ctx))

	// Ускоряем TTL через miniredis FastForward (prod TTL = 2h).
	ttl := mr.TTL("hub:backup:maintenance")
	require.Greater(t, ttl, time.Duration(0), "SET must install lease TTL")
	mr.FastForward(ttl + time.Second)

	require.False(t, gate.Enabled(ctx), "expired lease → maintenance off")
}
