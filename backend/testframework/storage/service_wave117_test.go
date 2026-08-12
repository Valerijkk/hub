package storage_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/internal/storage"
)

func TestOrphanCleanup_AttemptsS3DeleteWhenClientConfigured(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	repo := storageRepo(client)
	svc := newStorageServiceWithPresign(t, client)

	oldID := uuid.New()
	s3Key := projectID.String() + "/files/" + oldID.String() + "/orphan.txt"
	_, err := client.FileMetadata.Create().
		SetID(oldID).
		SetFilename("orphan.txt").
		SetContentType("text/plain").
		SetSize(0).
		SetS3Key(s3Key).
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		SetCreatedAt(time.Now().Add(-48 * time.Hour)).
		Save(ctx)
	require.NoError(t, err)

	cleanup := storage.NewOrphanCleanupService(repo, svc.S3Client(), svc.Bucket())
	require.NoError(t, cleanup.Cleanup(ctx))

	// W7-A4-020 / P3-BE-02: S3 delete на недоступный endpoint падает → meta остаётся.
	// require.NoError(Get) краснеет, если ветка ошибочно hard-delete'ит meta при S3 err.
	_, err = client.FileMetadata.Get(ctx, oldID)
	require.NoError(t, err, "orphan meta must remain when S3 delete fails")
}
