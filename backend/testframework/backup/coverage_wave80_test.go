package backup_test
import (
	"github.com/valerii/hub/backend/internal/backup"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/backuprecord"
)

func TestNewBackupService_NilRootContext(t *testing.T) {
	svc := backup.NewBackupService(nil, nil, nil, "bucket", "", "")
	require.NotNil(t, svc)
	assert.NotNil(t, backup.ExportBackupLifecycleCtx(svc))
}

func TestDumpMongo_ExecUnavailable(t *testing.T) {
	svc := backup.ExportNewBackupServiceForURLTest("", "mongodb://localhost:27017/test")
	_, err := backup.ExportBackupDumpMongo(svc, context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mongodump")
}

func TestRestorePostgres_ExecUnavailable(t *testing.T) {
	svc := backup.ExportNewBackupServiceForURLTest("postgres://localhost/test", "")
	err := backup.ExportBackupRestorePostgres(svc, context.Background(), []byte("fake-dump"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pg_restore")
}

func TestRestoreMongo_ExecUnavailable(t *testing.T) {
	svc := backup.ExportNewBackupServiceForURLTest("", "mongodb://localhost:27017/test")
	err := backup.ExportBackupRestoreMongo(svc, context.Background(), []byte("fake-archive"))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mongorestore")
}

func TestRestoreBackup_NoS3Client(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	record, err := client.BackupRecord.Create().
		SetS3Key("backups/completed.zip").
		SetStatus(backuprecord.StatusCompleted).
		SetFileSize(100).
		Save(ctx)
	require.NoError(t, err)

	err = svc.RestoreBackup(ctx, record.ID)
	require.Error(t, err)
	// A7-009: aborts at pre-restore when dump/S3 unavailable (message may cite pg_dump or S3).
	assert.Contains(t, err.Error(), "pre-restore backup failed")
}

func TestRestoreBackup_GetNotFound(t *testing.T) {
	svc, _, ctx := newBackupTestService(t)
	err := svc.RestoreBackup(ctx, uuid.New())
	require.Error(t, err)
}

func TestDumpPostgres_ExecUnavailable(t *testing.T) {
	svc := backup.ExportNewBackupServiceForURLTest("postgres://localhost/test", "")
	_, err := backup.ExportBackupDumpPostgres(svc, context.Background())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pg_dump")
}

func TestMarkFailed_UpdatesRecord(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	record, err := client.BackupRecord.Create().
		SetS3Key("backups/fail.zip").
		SetStatus(backuprecord.StatusInProgress).
		Save(ctx)
	require.NoError(t, err)

	backup.ExportBackupMarkFailed(svc, ctx, record.ID, assert.AnError)
	updated, err := client.BackupRecord.Get(ctx, record.ID)
	require.NoError(t, err)
	assert.Equal(t, backuprecord.StatusFailed, updated.Status)
	require.NotNil(t, updated.ErrorMessage)
}
