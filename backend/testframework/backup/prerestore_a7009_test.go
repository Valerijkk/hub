package backup_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/backuprecord"
)

// TestRestoreBackup_AbortsWhenPreBackupFails_A7_009 — no destructive restore
// without a successful safety-net backup.
func TestRestoreBackup_AbortsWhenPreBackupFails_A7_009(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	// No S3 → PerformBackup (pre-restore) fails → restore must abort.
	record, err := client.BackupRecord.Create().
		SetS3Key("backups/completed.zip").
		SetStatus(backuprecord.StatusCompleted).
		SetFileSize(100).
		Save(ctx)
	require.NoError(t, err)

	err = svc.RestoreBackup(ctx, record.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "pre-restore backup failed")
}
