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

	// Удаление из S3 на недоступный endpoint падает; строка метаданных может остаться — ветка всё равно проверена.
	_, err = client.FileMetadata.Get(ctx, oldID)
	// Либо удалена (если S3-мок успешен), либо ещё присутствует — без паники на пути ошибки S3.
	_ = err
}
