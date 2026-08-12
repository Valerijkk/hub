//go:build hubtest

package storage_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/internal/storage"
)

// TestOrphanCleanup_DrainsBeyondBatch_Task067 — W7-A4-015 / P3-BE-15:
// orphan > batch limit drained in one Cleanup.
// Мутация: убрать for-цикл → останутся хвосты после Cleanup.
func TestOrphanCleanup_DrainsBeyondBatch_Task067(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	repo := storageRepo(client)

	restore := storage.ExportSetOrphanBatchLimit(2)
	t.Cleanup(restore)

	const total = 5
	ids := make([]uuid.UUID, 0, total)
	for i := 0; i < total; i++ {
		id := uuid.New()
		ids = append(ids, id)
		_, err := client.FileMetadata.Create().
			SetID(id).
			SetFilename(fmt.Sprintf("orphan-%d.txt", i)).
			SetContentType("text/plain").
			SetSize(0).
			SetS3Key(fmt.Sprintf("%s/orphan-%d.txt", projectID, i)).
			SetS3Bucket("hub-files").
			SetStatus(filemetadata.StatusPending).
			SetUploaderID(userID).
			SetProjectID(projectID).
			SetCreatedAt(time.Now().Add(-48 * time.Hour)).
			Save(ctx)
		require.NoError(t, err)
	}

	svc := storage.NewOrphanCleanupService(repo, nil, "hub-files")
	require.NoError(t, svc.Cleanup(ctx))

	for _, id := range ids {
		_, err := client.FileMetadata.Get(ctx, id)
		assert.Error(t, err, "orphan %s must be drained", id)
	}
}
