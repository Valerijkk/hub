package storage_test
import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/internal/storage"
)

func TestOrphanCleanup_RemovesStalePending(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	repo := storageRepo(client)

	oldID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(oldID).
		SetFilename("orphan.txt").
		SetContentType("text/plain").
		SetSize(0).
		SetS3Key(projectID.String()+"/orphan.txt").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		SetCreatedAt(time.Now().Add(-48 * time.Hour)).
		Save(ctx)
	require.NoError(t, err)

	freshID := uuid.New()
	_, err = client.FileMetadata.Create().
		SetID(freshID).
		SetFilename("fresh.txt").
		SetContentType("text/plain").
		SetSize(0).
		SetS3Key(projectID.String()+"/fresh.txt").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	svc := storage.NewOrphanCleanupService(repo, nil, "hub-files")
	require.NoError(t, svc.Cleanup(ctx))

	_, err = client.FileMetadata.Get(ctx, oldID)
	assert.Error(t, err)

	_, err = client.FileMetadata.Get(ctx, freshID)
	require.NoError(t, err)
}

func TestOrphanCleanup_NoOrphans(t *testing.T) {
	client, ctx, _, _ := newStorageTestClient(t)
	svc := storage.NewOrphanCleanupService(storageRepo(client), nil, "hub-files")
	require.NoError(t, svc.Cleanup(ctx))
}
