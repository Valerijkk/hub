package storage_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/internal/storage"
)

func TestEntRepository_CreatePendingAndConfirm(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	repo := storage.NewEntRepository(client)

	fileID := uuid.New()
	meta, err := repo.CreatePending(ctx, storage.CreateFileInput{
		ID:          fileID,
		Filename:    "a.txt",
		ContentType: "text/plain",
		S3Key:       "k/a.txt",
		S3Bucket:    "hub-files",
		UploaderID:  userID,
		ProjectID:   projectID,
	})
	require.NoError(t, err)
	assert.Equal(t, filemetadata.StatusPending, meta.Status)

	require.NoError(t, repo.ConfirmUpload(ctx, fileID, "issue", "HUB-1"))
	got, err := repo.GetActiveByID(ctx, fileID)
	require.NoError(t, err)
	assert.Equal(t, filemetadata.StatusConfirmed, got.Status)
}

func TestEntRepository_ListOrphanPending(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	repo := storage.NewEntRepository(client)

	oldID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(oldID).
		SetFilename("orphan.txt").
		SetContentType("text/plain").
		SetSize(0).
		SetS3Key("orphan").
		SetS3Bucket("hub-files").
		SetStatus(filemetadata.StatusPending).
		SetUploaderID(userID).
		SetProjectID(projectID).
		SetCreatedAt(time.Now().Add(-48 * time.Hour)).
		Save(ctx)
	require.NoError(t, err)

	list, err := repo.ListOrphanPending(ctx, time.Now().Add(-24*time.Hour), 10)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, oldID, list[0].ID)
}

func TestEntRepository_SoftDelete(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	repo := storage.NewEntRepository(client)

	fileID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("del.txt").
		SetContentType("text/plain").
		SetSize(1).
		SetS3Key("k").
		SetS3Bucket("b").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	at := time.Now().UTC()
	require.NoError(t, repo.SoftDelete(ctx, fileID, at))
	meta, err := client.FileMetadata.Get(ctx, fileID)
	require.NoError(t, err)
	assert.Equal(t, filemetadata.StatusDeleted, meta.Status)
}

func TestEntRepository_ProjectIDForActiveFile(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	repo := storage.NewEntRepository(client)
	fileID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("x").
		SetContentType("text/plain").
		SetSize(1).
		SetS3Key("k").
		SetS3Bucket("b").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	got, err := repo.ProjectIDForActiveFile(ctx, fileID)
	require.NoError(t, err)
	assert.Equal(t, projectID, got)
}
