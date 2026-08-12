package backup_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/backup"
)

// TestRestorePostgres_DoesNotEmbedDumpOutput_A7_007 — CombinedOutput / DSN must not
// appear in the returned error (handler also maps to generic "restore failed").
// Mutation: `return fmt.Errorf("pg_restore failed: %s", output)` or wrap pgURL → red.
func TestRestorePostgres_DoesNotEmbedDumpOutput_A7_007(t *testing.T) {
	const secret = "supersecret-pg-pass"
	pgURL := "postgres://hub:" + secret + "@db.internal:5432/hub?sslmode=disable"
	svc := backup.ExportNewBackupServiceForURLTest(pgURL, "")
	err := backup.ExportBackupRestorePostgres(svc, context.Background(), []byte("fake-dump"))
	require.Error(t, err)
	assert.Equal(t, "pg_restore failed", err.Error())
	assert.NotContains(t, err.Error(), secret)
	assert.NotContains(t, err.Error(), "db.internal")
	assert.NotContains(t, err.Error(), pgURL)
}

// TestRestoreMongo_DoesNotEmbedDumpOutput_A7_007 — same sanitize for mongorestore.
// Mutation: embed CombinedOutput or --uri= into error → red.
func TestRestoreMongo_DoesNotEmbedDumpOutput_A7_007(t *testing.T) {
	const secret = "supersecret-mongo-pass"
	mongoURL := "mongodb://hub:" + secret + "@mongo.internal:27017/?authSource=admin"
	svc := backup.ExportNewBackupServiceForURLTest("", mongoURL)
	err := backup.ExportBackupRestoreMongo(svc, context.Background(), []byte("fake-archive"))
	require.Error(t, err)
	assert.Equal(t, "mongorestore failed", err.Error())
	assert.NotContains(t, err.Error(), secret)
	assert.NotContains(t, err.Error(), "mongo.internal")
	assert.NotContains(t, err.Error(), mongoURL)
}
