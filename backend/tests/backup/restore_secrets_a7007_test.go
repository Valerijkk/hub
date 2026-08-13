package backup_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/backuprecord"
	"github.com/valerii/hub/backend/internal/backup"
)

// TestRestoreHandler_DoesNotLeakSecrets_A7_007 — HTTP 500 body must not contain
// connection strings from CombinedOutput / service errors.
func TestRestoreHandler_DoesNotLeakSecrets_A7_007(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	record, err := client.BackupRecord.Create().
		SetS3Key("backups/done.zip").
		SetStatus(backuprecord.StatusCompleted).
		Save(ctx)
	require.NoError(t, err)

	h := backup.NewHandler(svc)
	app := fiber.New()
	h.RegisterRoutes(app, func(c fiber.Ctx) error { return c.Next() }, func(c fiber.Ctx) error { return c.Next() })

	body, _ := json.Marshal(backup.RestoreRequest{Confirm: "RESTORE"})
	req := httptest.NewRequest(http.MethodPost, "/admin/backups/"+record.ID.String()+"/restore", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)

	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	msg := string(raw)
	assert.NotContains(t, msg, "postgres://")
	assert.NotContains(t, msg, "mongodb://")
	assert.NotContains(t, msg, "--uri=")
	assert.Contains(t, msg, "restore failed")
}
