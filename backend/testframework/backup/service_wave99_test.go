package backup_test
import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/backuprecord"
	"github.com/valerii/hub/backend/internal/backup"
)

// installFakeBackupBinaries добавляет временный каталог в начало PATH со стабами pg_dump/mongodump/pg_restore/mongorestore.
func installFakeBackupBinaries(t *testing.T, mongoFail bool) {
	t.Helper()
	dir := t.TempDir()

	pgDump := "#!/bin/sh\necho 'fake pg dump'\n"
	mongoDump := "#!/bin/sh\necho 'fake mongo archive'\n"
	if mongoFail {
		mongoDump = "#!/bin/sh\nexit 1\n"
	}
	pgRestore := "#!/bin/sh\nexit 0\n"
	mongoRestore := "#!/bin/sh\nexit 0\n"

	if runtime.GOOS == "windows" {
		writeBat := func(name, body string) {
			require.NoError(t, os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644))
		}
		writeBat("pg_dump.bat", "@echo off\r\necho fake pg dump\r\n")
		if mongoFail {
			writeBat("mongodump.bat", "@echo off\r\nexit /b 1\r\n")
		} else {
			writeBat("mongodump.bat", "@echo off\r\necho fake mongo archive\r\n")
		}
		writeBat("pg_restore.bat", "@echo off\r\nexit /b 0\r\n")
		writeBat("mongorestore.bat", "@echo off\r\nexit /b 0\r\n")
	} else {
		writeSh := func(name, body string) {
			p := filepath.Join(dir, name)
			require.NoError(t, os.WriteFile(p, []byte(body), 0o755))
		}
		writeSh("pg_dump", pgDump)
		writeSh("mongodump", mongoDump)
		writeSh("pg_restore", pgRestore)
		writeSh("mongorestore", mongoRestore)
	}

	oldPath := os.Getenv("PATH")
	t.Setenv("PATH", dir+string(os.PathListSeparator)+oldPath)
}

func TestPerformBackup_CompletesWithoutS3(t *testing.T) {
	installFakeBackupBinaries(t, false)

	svc, client, ctx := newBackupTestService(t)
	backup.ExportBackupSetURLs(svc, "postgres://localhost/hub_test", "mongodb://localhost:27017/hub_test")

	resp, err := svc.PerformBackup(ctx)
	require.NoError(t, err)
	require.NotNil(t, resp)

	records, err := client.BackupRecord.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, backuprecord.StatusCompleted, records[0].Status)
	assert.Greater(t, records[0].FileSize, int64(0))

	// Даём горутине очистки завершиться.
	time.Sleep(50 * time.Millisecond)
}

func TestPerformBackup_MongoDumpFailsAfterPgSuccess(t *testing.T) {
	installFakeBackupBinaries(t, true)

	svc, client, ctx := newBackupTestService(t)
	backup.ExportBackupSetURLs(svc, "postgres://localhost/hub_test", "mongodb://localhost:27017/hub_test")

	resp, err := svc.PerformBackup(ctx)
	require.Error(t, err, "A7-004: mongodump failure must not return nil error")
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "mongodump")

	records, err := client.BackupRecord.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, backuprecord.StatusFailed, records[0].Status)
	require.NotNil(t, records[0].ErrorMessage)
	assert.Contains(t, *records[0].ErrorMessage, "mongodump")
}

func TestCleanupOldBackups_RepoListError(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	require.NoError(t, client.Close())
	backup.ExportBackupCleanupOldBackups(svc, ctx)
}

func TestDeleteBackup_WithNilS3Client(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	record, err := client.BackupRecord.Create().
		SetS3Key("backups/nil-s3.zip").
		SetStatus(backuprecord.StatusCompleted).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.DeleteBackup(ctx, record.ID))
	_, err = client.BackupRecord.Get(ctx, record.ID)
	require.Error(t, err)
}

func TestMarkFailed_LogsWhenRepoUpdateFails(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	record, err := client.BackupRecord.Create().
		SetS3Key("backups/mark-fail.zip").
		SetStatus(backuprecord.StatusInProgress).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.Close())
	backup.ExportBackupMarkFailed(svc, context.Background(), record.ID, assert.AnError)
}
