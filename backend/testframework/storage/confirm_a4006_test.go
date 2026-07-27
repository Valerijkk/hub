package storage_test

import (
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
