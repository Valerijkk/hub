package backup

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testBackupApp(t *testing.T) *fiber.App {
	t.Helper()
	svc, _ := newBackupTestService(t)
	h := NewHandler(svc)

	app := fiber.New()
	h.RegisterRoutes(app, func(c fiber.Ctx) error { return c.Next() }, func(c fiber.Ctx) error { return c.Next() })
	return app
}

func TestHandler_RestoreBackup_Validation(t *testing.T) {
	app := testBackupApp(t)

	t.Run("invalid backup id", func(t *testing.T) {
		body, _ := json.Marshal(RestoreRequest{Confirm: "RESTORE"})
		req := httptest.NewRequest(http.MethodPost, "/admin/backups/not-uuid/restore", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})

	t.Run("missing confirm", func(t *testing.T) {
		id := uuid.New()
		body, _ := json.Marshal(RestoreRequest{Confirm: "NOPE"})
		req := httptest.NewRequest(http.MethodPost, "/admin/backups/"+id.String()+"/restore", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func TestHandler_DeleteBackup_InvalidID(t *testing.T) {
	app := testBackupApp(t)
	req := httptest.NewRequest(http.MethodDelete, "/admin/backups/bad-id", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ListBackups_Empty(t *testing.T) {
	app := testBackupApp(t)
	req := httptest.NewRequest(http.MethodGet, "/admin/backups", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var list []BackupResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&list))
	assert.Empty(t, list)
}

func TestHandler_TriggerBackup_CreatesRecord(t *testing.T) {
	app := testBackupApp(t)
	req := httptest.NewRequest(http.MethodPost, "/admin/backups", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	// pg_dump unavailable in unit test env — service marks failed but returns 201
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestHandler_RestoreBackup_InvalidJSON(t *testing.T) {
	app := testBackupApp(t)
	id := uuid.New()
	req := httptest.NewRequest(http.MethodPost, "/admin/backups/"+id.String()+"/restore", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestCreateZIP_EmptyPayloads(t *testing.T) {
	svc := &BackupService{}
	zipData, err := svc.createZIP(nil, nil)
	require.NoError(t, err)
	gotPG, gotMongo, err := svc.extractZIP(zipData)
	require.NoError(t, err)
	assert.Empty(t, gotPG)
	assert.Empty(t, gotMongo)
}
