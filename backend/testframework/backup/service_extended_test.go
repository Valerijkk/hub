package backup_test
import (
	"github.com/valerii/hub/backend/internal/backup"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/backuprecord"
	"github.com/valerii/hub/backend/ent/enttest"
)

func newBackupTestService(t *testing.T) (*backup.BackupService, *ent.Client, context.Context) {
	t.Helper()
	dsn := "file:backuptest_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	return backup.NewBackupService(context.Background(), backup.NewEntRepository(client), nil, "hub-backups", "", ""), client, context.Background()
}

func TestExtractZIP_RoundTrip(t *testing.T) {
	svc := &backup.BackupService{}
	pgData := []byte("postgres dump content")
	mongoData := []byte("mongo archive content")

	zipData, err := backup.ExportBackupCreateZIP(svc, pgData, mongoData)
	require.NoError(t, err)

	gotPG, gotMongo, err := backup.ExportBackupExtractZIP(svc, zipData)
	require.NoError(t, err)
	assert.Equal(t, pgData, gotPG)
	assert.Equal(t, mongoData, gotMongo)
}

func TestExtractZIP_InvalidArchive(t *testing.T) {
	svc := &backup.BackupService{}
	_, _, err := backup.ExportBackupExtractZIP(svc, []byte("not-a-zip"))
	require.Error(t, err)
}

func TestToBackupResponse(t *testing.T) {
	now := time.Now().UTC()
	errMsg := "dump failed"
	record := &ent.BackupRecord{
		ID:           uuid.New(),
		S3Key:        "backups/test.zip",
		FileSize:     1024,
		Status:       backuprecord.StatusFailed,
		ErrorMessage: &errMsg,
		CreatedAt:    now,
	}

	resp := backup.ExportToBackupResponse(record)
	require.NotNil(t, resp)
	assert.Equal(t, record.ID.String(), resp.ID)
	assert.Equal(t, "failed", resp.Status)
	assert.Equal(t, "dump failed", resp.ErrorMessage)
	assert.Equal(t, int64(1024), resp.FileSize)
}

func TestToBackupResponse_NoErrorMessage(t *testing.T) {
	record := &ent.BackupRecord{
		ID:     uuid.New(),
		S3Key:  "backups/ok.zip",
		Status: backuprecord.StatusCompleted,
	}
	resp := backup.ExportToBackupResponse(record)
	assert.Empty(t, resp.ErrorMessage)
}

func TestListBackups_Ordered(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)

	older, err := client.BackupRecord.Create().
		SetS3Key("backups/old.zip").
		SetStatus(backuprecord.StatusCompleted).
		Save(ctx)
	require.NoError(t, err)

	newer, err := client.BackupRecord.Create().
		SetS3Key("backups/new.zip").
		SetStatus(backuprecord.StatusCompleted).
		SetFileSize(2048).
		Save(ctx)
	require.NoError(t, err)
	_ = older

	list, err := svc.ListBackups(ctx)
	require.NoError(t, err)
	require.Len(t, list, 2)
	assert.Equal(t, newer.ID.String(), list[0].ID)
	assert.Equal(t, "backups/new.zip", list[0].S3Key)
}

func TestDeleteBackup_NotFound(t *testing.T) {
	svc, _, ctx := newBackupTestService(t)
	err := svc.DeleteBackup(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

func TestRestoreBackup_NotFound(t *testing.T) {
	svc, _, ctx := newBackupTestService(t)
	err := svc.RestoreBackup(ctx, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

func TestRestoreBackup_WrongStatus(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	record, err := client.BackupRecord.Create().
		SetS3Key("backups/in-progress.zip").
		SetStatus(backuprecord.StatusInProgress).
		Save(ctx)
	require.NoError(t, err)

	err = svc.RestoreBackup(ctx, record.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "невозможно восстановить")
}

func TestStrPtr(t *testing.T) {
	s := backup.ExportStrPtr("test")
	require.NotNil(t, s)
	assert.Equal(t, "test", *s)
}

// TestPerformBackup_PgDumpFailReturnsError_A7_004 — dump fail → error (не HTTP success).
// Mutation: return toBackupResponse(record), nil → require.Error FAIL.
func TestPerformBackup_PgDumpFailReturnsError_A7_004(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	resp, err := svc.PerformBackup(ctx)
	require.Error(t, err, "A7-004: failed dump must not return nil error")
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "pg_dump")

	records, err := client.BackupRecord.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, backuprecord.StatusFailed, records[0].Status)
	require.NotNil(t, records[0].ErrorMessage)
}

func TestDeleteBackup_Success(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	record, err := client.BackupRecord.Create().
		SetS3Key("backups/delete-me.zip").
		SetStatus(backuprecord.StatusCompleted).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.DeleteBackup(ctx, record.ID))
	_, err = client.BackupRecord.Get(ctx, record.ID)
	require.Error(t, err)
}

func TestCleanupOldBackups_RemovesExcess(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	backup.ExportBackupSetMaxBackups(svc, 2)
	base := time.Now().UTC().Add(-time.Hour)

	for i := 0; i < 4; i++ {
		_, err := client.BackupRecord.Create().
			SetS3Key("backups/old-" + string(rune('a'+i)) + ".zip").
			SetStatus(backuprecord.StatusCompleted).
			SetFileSize(int64(100 + i)).
			SetCreatedAt(base.Add(time.Duration(i) * time.Minute)).
			Save(ctx)
		require.NoError(t, err)
	}

	backup.ExportBackupCleanupOldBackups(svc, ctx)
	remaining, err := client.BackupRecord.Query().All(ctx)
	require.NoError(t, err)
	assert.Len(t, remaining, 2)
}

func TestSetMaintenanceMode_NoPanic(t *testing.T) {
	svc, _, _ := newBackupTestService(t)
	backup.ExportBackupSetMaintenanceMode(svc, true)
	backup.ExportBackupSetMaintenanceMode(svc, false)
}
