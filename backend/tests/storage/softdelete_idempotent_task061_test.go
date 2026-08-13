//go:build hubtest

package storage_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/internal/storage"
)

// TestEntRepository_SoftDelete_IdempotentNotFound_P3BE05 — повторный SoftDelete → NotFound.
// Mutation: UpdateOneID без DeletedAtIsNil → второй delete «успешен» → FAIL.
func TestEntRepository_SoftDelete_IdempotentNotFound_P3BE05(t *testing.T) {
	client, ctx, userID, projectID := newStorageTestClient(t)
	repo := storage.NewEntRepository(client)
	fileID := uuid.New()
	_, err := client.FileMetadata.Create().
		SetID(fileID).
		SetFilename("once.txt").
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
	err = repo.SoftDelete(ctx, fileID, at.Add(time.Second))
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}
