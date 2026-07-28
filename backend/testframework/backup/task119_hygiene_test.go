//go:build hubtest

package backup_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/backuprecord"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/backup"
	"github.com/valerii/hub/backend/internal/rbac"
)

// TestMarkFailed_SanitizesErrorMessage_W7A4007 — ErrorMessage без host/DSN.
// Mutation: MarkFailed(ctx, id, err.Error()) с wrap URI → Contains secret.
func TestMarkFailed_SanitizesErrorMessage_W7A4007(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	rec, err := client.BackupRecord.Create().
		SetS3Key("backups/sanitize.zip").
		SetStatus(backuprecord.StatusInProgress).
		Save(ctx)
	require.NoError(t, err)

	const secret = "supersecret-pg-pass"
	raw := fmt.Errorf("ошибка pg_dump: dial tcp db.internal:5432: connection refused (postgres://hub:%s@db.internal/hub)", secret)
	backup.ExportBackupMarkFailed(svc, ctx, rec.ID, raw)

	got, err := client.BackupRecord.Get(ctx, rec.ID)
	require.NoError(t, err)
	require.NotNil(t, got.ErrorMessage)
	assert.Equal(t, "pg_dump failed", *got.ErrorMessage)
	assert.NotContains(t, *got.ErrorMessage, secret)
	assert.NotContains(t, *got.ErrorMessage, "db.internal")
	assert.NotContains(t, *got.ErrorMessage, "postgres://")
}

// TestPgDumpArgsEnv_NoPasswordInArgv_W7A4013 — PGPASSWORD в env, URI без пароля.
// Mutation: args содержат raw URL с :pass@ → red.
func TestPgDumpArgsEnv_NoPasswordInArgv_W7A4013(t *testing.T) {
	const secret = "argv-leak-pg-secret"
	pgURL := "postgres://hub:" + secret + "@db.internal:5432/hub?sslmode=disable"
	args, env, err := backup.ExportPgDumpArgsEnv(pgURL, "--format=custom")
	require.NoError(t, err)

	joined := strings.Join(args, " ")
	assert.NotContains(t, joined, secret)
	assert.Contains(t, joined, "--dbname=")
	assert.Contains(t, joined, "hub@db.internal")

	found := false
	for _, e := range env {
		if strings.HasPrefix(e, "PGPASSWORD=") {
			found = true
			assert.Equal(t, "PGPASSWORD="+secret, e)
		}
	}
	assert.True(t, found, "PGPASSWORD must be set in env")
}

// TestMongoConfig_NoPasswordInArgv_W7A4013 — секрет только в temp --config.
// Mutation: --uri=mongodb://user:pass@... в argv → red.
func TestMongoConfig_NoPasswordInArgv_W7A4013(t *testing.T) {
	const secret = "argv-leak-mongo-secret"
	mongoURL := "mongodb://hub:" + secret + "@mongo.internal:27017/?authSource=admin"
	path, cleanup, err := backup.ExportWriteMongoConfig(t.TempDir(), mongoURL)
	require.NoError(t, err)
	defer cleanup()

	arg := backup.ExportMongoConfigArg(path)
	assert.True(t, strings.HasPrefix(arg, "--config="))
	assert.NotContains(t, arg, secret)
	assert.NotContains(t, arg, "mongo.internal")

	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Contains(t, string(raw), secret)
	assert.Contains(t, string(raw), "password:")
	assert.NotContains(t, arg, filepath.Base(path)+secret)
	// Unix: 0600; Windows ACL игнорирует chmod — не assert Mode().Perm().
	_ = os.FileMode(0o600)
}

// TestHandler_RestoreDelete_NotFound404_W7A4011 — missing id → 404, не 500.
// Mutation: handler всегда 500 → StatusNotFound FAIL.
func TestHandler_RestoreDelete_NotFound404_W7A4011(t *testing.T) {
	svc, _, _ := newBackupTestService(t)
	h := backup.NewHandler(svc)
	app := fiber.New()
	h.RegisterRoutes(app, func(c fiber.Ctx) error { return c.Next() }, func(c fiber.Ctx) error { return c.Next() })

	missing := uuid.New()

	t.Run("restore", func(t *testing.T) {
		body, _ := json.Marshal(backup.RestoreRequest{Confirm: "RESTORE"})
		req := httptest.NewRequest(http.MethodPost, "/admin/backups/"+missing.String()+"/restore", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
		raw, _ := io.ReadAll(resp.Body)
		assert.Contains(t, string(raw), "backup not found")
	})

	t.Run("delete", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodDelete, "/admin/backups/"+missing.String(), nil)
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestHandler_RequireAdmin_Denial_W7A4012 — member → 403 с реальным RequireAdmin.
// Mutation: adminMw = Next() → member 200/500 вместо 403.
func TestHandler_RequireAdmin_Denial_W7A4012(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)

	member, err := client.User.Create().
		SetEmail("member-" + uuid.New().String()[:8] + "@example.com").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	admin, err := client.User.Create().
		SetEmail("admin-" + uuid.New().String()[:8] + "@example.com").
		SetSystemRole(user.SystemRoleAdmin).
		Save(ctx)
	require.NoError(t, err)

	h := backup.NewHandler(svc)
	adminMw := rbac.RequireAdmin(rbac.NewEntRepository(client))

	newApp := func(userID uuid.UUID) *fiber.App {
		app := fiber.New()
		authMw := func(c fiber.Ctx) error {
			c.Locals("userID", userID.String())
			return c.Next()
		}
		h.RegisterRoutes(app, authMw, adminMw)
		return app
	}

	t.Run("member forbidden", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/backups", nil)
		resp, err := newApp(member.ID).Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("admin allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/admin/backups", nil)
		resp, err := newApp(admin.ID).Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})
}
