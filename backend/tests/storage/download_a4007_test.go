package storage_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/internal/storage"
)

// TestGetActiveByID_RejectsPending_A4_007 — pending files must not be downloadable
// via GetActiveByID (mutation: drop StatusConfirmed filter → FAIL).
func TestGetActiveByID_RejectsPending_A4_007(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	repo := storage.NewEntRepository(client)

	pendingID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(pendingID).
		SetFilename("pending.pdf").
		SetContentType("application/pdf").
		SetSize(0).
		SetS3Key(projectID.String() + "/files/" + pendingID.String() + "/pending.pdf").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	_, err = repo.GetActiveByID(ctx, pendingID)
	require.True(t, ent.IsNotFound(err), "pending must not resolve for download")

	confirmedID := uuid.New()
	_, err = client.FileMetadata.Create().
		SetID(confirmedID).
		SetFilename("ok.pdf").
		SetContentType("application/pdf").
		SetSize(10).
		SetS3Key(projectID.String() + "/files/" + confirmedID.String() + "/ok.pdf").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	got, err := repo.GetActiveByID(ctx, confirmedID)
	require.NoError(t, err)
	require.Equal(t, confirmedID, got.ID)
}
