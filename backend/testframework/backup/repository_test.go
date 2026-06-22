package backup_test
import (
	"github.com/valerii/hub/backend/internal/backup"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/backuprecord"
	"github.com/valerii/hub/backend/ent/enttest"
)

func newBackupRepo(t *testing.T) (*backup.EntRepository, context.Context) {
	t.Helper()
	dsn := "file:backuprepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return backup.NewEntRepository(client), context.Background()
}

func TestEntRepository_CreateAndMarkCompleted(t *testing.T) {
	repo, ctx := newBackupRepo(t)
	rec, err := repo.CreateInProgress(ctx, "backups/test.zip")
	require.NoError(t, err)
	assert.Equal(t, backuprecord.StatusInProgress, rec.Status)

	done, err := repo.MarkCompleted(ctx, rec.ID, 4096)
	require.NoError(t, err)
	assert.Equal(t, backuprecord.StatusCompleted, done.Status)
	assert.Equal(t, int64(4096), done.FileSize)
}

func TestEntRepository_MarkFailed(t *testing.T) {
	repo, ctx := newBackupRepo(t)
	rec, err := repo.CreateInProgress(ctx, "backups/fail.zip")
	require.NoError(t, err)

	require.NoError(t, repo.MarkFailed(ctx, rec.ID, "dump error"))
	got, err := repo.Get(ctx, rec.ID)
	require.NoError(t, err)
	assert.Equal(t, backuprecord.StatusFailed, got.Status)
	require.NotNil(t, got.ErrorMessage)
	assert.Equal(t, "dump error", *got.ErrorMessage)
}

func TestEntRepository_ListCompletedNewestFirst(t *testing.T) {
	repo, ctx := newBackupRepo(t)
	_, err := repo.CreateInProgress(ctx, "backups/a.zip")
	require.NoError(t, err)
	b, err := repo.CreateInProgress(ctx, "backups/b.zip")
	require.NoError(t, err)
	_, err = repo.MarkCompleted(ctx, b.ID, 100)
	require.NoError(t, err)

	list, err := repo.ListCompletedNewestFirst(ctx)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "backups/b.zip", list[0].S3Key)
}

func TestEntRepository_SetRestoredNoteAndDelete(t *testing.T) {
	repo, ctx := newBackupRepo(t)
	rec, err := repo.CreateInProgress(ctx, "backups/restore.zip")
	require.NoError(t, err)
	_, err = repo.MarkCompleted(ctx, rec.ID, 2048)
	require.NoError(t, err)

	require.NoError(t, repo.SetRestoredNote(ctx, rec.ID, "restored 2026-06-21"))
	got, err := repo.Get(ctx, rec.ID)
	require.NoError(t, err)
	require.NotNil(t, got.ErrorMessage)
	assert.Equal(t, "restored 2026-06-21", *got.ErrorMessage)

	all, err := repo.ListAll(ctx)
	require.NoError(t, err)
	require.Len(t, all, 1)

	require.NoError(t, repo.Delete(ctx, rec.ID))
	_, err = repo.Get(ctx, rec.ID)
	require.Error(t, err)
	assert.True(t, backup.ExportBackupIsNotFound(err))
}
