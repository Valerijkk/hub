package backup_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/backup"
)

// TestMaintenanceGate_BlocksAPI_AllowsHealth_A7_003 — restore-гейт отклоняет
// бизнес-API 503, /api/health пропускает.
func TestMaintenanceGate_BlocksAPI_AllowsHealth_A7_003(t *testing.T) {
	gate := backup.NewMaintenanceGate()
	svc := backup.ExportNewBackupServiceWithGate(t.Context(), nil, gate)

	app := fiber.New()
	api := app.Group("/api")
	api.Use(backup.MaintenanceMiddleware(gate))
	api.Get("/health", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"status": "ok"})
	})
	api.Get("/projects", func(c fiber.Ctx) error {
		return c.JSON(fiber.Map{"ok": true})
	})

	// До maintenance — projects OK.
	req := httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()

	backup.ExportBackupSetMaintenanceMode(svc, true)
	require.True(t, backup.ExportBackupIsMaintenanceMode(svc))

	req = httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusServiceUnavailable, resp.StatusCode, "A7-003: non-health must be 503")
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	assert.Contains(t, string(body), "maintenance")

	req = httptest.NewRequest(http.MethodGet, "/api/health", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode, "A7-003: health must bypass")
	_ = resp.Body.Close()

	backup.ExportBackupSetMaintenanceMode(svc, false)
	require.False(t, backup.ExportBackupIsMaintenanceMode(svc))

	req = httptest.NewRequest(http.MethodGet, "/api/projects", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = resp.Body.Close()
}

// TestSetMaintenanceMode_FlipsGate_A7_003 — setMaintenanceMode реально ставит флаг.
func TestSetMaintenanceMode_FlipsGate_A7_003(t *testing.T) {
	gate := backup.NewMaintenanceGate()
	svc := backup.ExportNewBackupServiceWithGate(t.Context(), nil, gate)
	require.False(t, gate.Enabled())
	backup.ExportBackupSetMaintenanceMode(svc, true)
	require.True(t, gate.Enabled())
	backup.ExportBackupSetMaintenanceMode(svc, false)
	require.False(t, gate.Enabled())
}
