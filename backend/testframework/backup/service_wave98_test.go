package backup_test
import (
	"github.com/valerii/hub/backend/internal/backup"
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractZIP_OnlyMongoFile(t *testing.T) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	fw, err := w.Create("mongo.archive")
	require.NoError(t, err)
	_, err = fw.Write([]byte("mongo-data"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	svc := &backup.BackupService{}
	pg, mongo, err := backup.ExportBackupExtractZIP(svc, buf.Bytes())
	require.NoError(t, err)
	assert.Nil(t, pg)
	assert.Equal(t, []byte("mongo-data"), mongo)
}

func TestExtractZIP_OnlyPostgresFile(t *testing.T) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	fw, err := w.Create("postgres.dump")
	require.NoError(t, err)
	_, err = fw.Write([]byte("pg-data"))
	require.NoError(t, err)
	require.NoError(t, w.Close())

	svc := &backup.BackupService{}
	pg, mongo, err := backup.ExportBackupExtractZIP(svc, buf.Bytes())
	require.NoError(t, err)
	assert.Equal(t, []byte("pg-data"), pg)
	assert.Nil(t, mongo)
}

func TestExtractZIP_SkipsUnknownEntries(t *testing.T) {
	buf := new(bytes.Buffer)
	w := zip.NewWriter(buf)
	for _, name := range []string{"readme.txt", "postgres.dump", "mongo.archive"} {
		fw, err := w.Create(name)
		require.NoError(t, err)
		_, err = fw.Write([]byte("x"))
		require.NoError(t, err)
	}
	require.NoError(t, w.Close())

	svc := &backup.BackupService{}
	pg, mongo, err := backup.ExportBackupExtractZIP(svc, buf.Bytes())
	require.NoError(t, err)
	assert.NotNil(t, pg)
	assert.NotNil(t, mongo)
}
