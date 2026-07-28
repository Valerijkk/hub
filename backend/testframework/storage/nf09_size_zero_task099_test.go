package storage_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/storage"
)

// TestGenerateUploadURL_RejectsZeroSize_W7_A4_001 — size==0 legacy → ErrInvalidFileSize.
// Mutation: вернуть size<=0 к allow → err==nil FAIL (NF-09 / ADR-STORAGE-NF09-SIZE-2026).
func TestGenerateUploadURL_RejectsZeroSize_W7_A4_001(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := newStorageServiceWithPresign(t, client)

	_, err := svc.GenerateUploadURL(ctx, "legacy.bin", "application/octet-stream", userID, projectID, 0)
	require.ErrorIs(t, err, storage.ErrInvalidFileSize)

	_, err = svc.GenerateUploadURL(ctx, "neg.bin", "application/octet-stream", userID, projectID, -1)
	require.ErrorIs(t, err, storage.ErrInvalidFileSize)
}

// TestGenerateUploadURL_BindsContentLengthWhenSizePositive_W7_A4_001 —
// size>0 проходит валидацию и создаёт pending (ContentLength bind в putInput).
// Mutation: снять size<=0 guard и не ставить ContentLength → oversized path снова открыт.
func TestGenerateUploadURL_BindsContentLengthWhenSizePositive_W7_A4_001(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	svc := newStorageServiceWithPresign(t, client)

	result, err := svc.GenerateUploadURL(ctx, "ok.bin", "application/octet-stream", userID, projectID, 1024)
	require.NoError(t, err)
	require.NotNil(t, result)
	assert.NotEmpty(t, result.UploadURL)
	assert.Contains(t, result.UploadURL, "X-Amz-") // signed URL
}

// TestHandler_GenerateUploadURL_RejectsOmittedSize_W7_A4_001 — JSON без size → 400.
func TestHandler_GenerateUploadURL_RejectsOmittedSize_W7_A4_001(t *testing.T) {
	app, projectID, _, _ := newStorageRoutesEnv(t)

	payload := fmt.Sprintf(`{"filename":"a.txt","contentType":"text/plain","projectId":%q}`, projectID.String())
	req := httptest.NewRequest(http.MethodPost, "/api/files/upload-url", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)

	var body map[string]any
	_ = json.NewDecoder(resp.Body).Decode(&body)
}
