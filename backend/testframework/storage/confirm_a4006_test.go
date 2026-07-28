package storage_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/internal/storage"
)

// TestConfirmUpload_RejectsAlreadyConfirmed_A4_006 — повторный confirm не
// перепривязывает entity (mutation: UpdateOneID без StatusPending → FAIL).
func TestConfirmUpload_RejectsAlreadyConfirmed_A4_006(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	repo := storage.NewEntRepository(client)

	fileID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("doc.pdf").
		SetContentType("application/pdf").
		SetSize(0).
		SetS3Key(projectID.String() + "/files/" + fileID.String() + "/doc.pdf").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, repo.ConfirmUpload(ctx, fileID, "issue", "ISSUE-1"))

	err = repo.ConfirmUpload(ctx, fileID, "issue", "ISSUE-HIJACK")
	require.ErrorIs(t, err, storage.ErrAlreadyConfirmed)

	got, err := client.FileMetadata.Get(ctx, fileID)
	require.NoError(t, err)
	assert.Equal(t, "ISSUE-1", got.EntityID, "entity must not be rebinding on re-confirm")
	assert.Equal(t, filemetadata.StatusConfirmed, got.Status)
}

// TestHandler_ConfirmUpload_AlreadyConfirmed_409_A4_006 — ErrAlreadyConfirmed
// maps to HTTP 409 at the handler. Mutation: drop StatusConflict branch → FAIL.
func TestHandler_ConfirmUpload_AlreadyConfirmed_409_A4_006(t *testing.T) {
	app, projectID, userID, client := newStorageRoutesEnv(t)
	ctx := context.Background()

	fileID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("already.pdf").
		SetContentType("application/pdf").
		SetSize(0).
		SetS3Key(projectID.String() + "/files/" + fileID.String() + "/already.pdf").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	body := `{"entityType":"message","entityId":"MSG-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/files/"+fileID.String()+"/confirm", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/files/"+fileID.String()+"/confirm", bytes.NewReader([]byte(
		`{"entityType":"message","entityId":"MSG-HIJACK"}`,
	)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode, "A4-006: re-confirm must be 409")

	got, err := client.FileMetadata.Get(ctx, fileID)
	require.NoError(t, err)
	assert.Equal(t, "MSG-1", got.EntityID, "entity must not rebind on HTTP re-confirm")
	assert.Equal(t, filemetadata.StatusConfirmed, got.Status)
}
