package backup_test
import (
	"github.com/valerii/hub/backend/internal/backup"
	"archive/zip"
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateZIP проверяет архивирование pg_dump и mongodump данных в ZIP.
func TestCreateZIP(t *testing.T) {
	svc := &backup.BackupService{}

	pgData := []byte("pg_dump_content_here")
	mongoData := []byte("mongodump_content_here")

	zipData, err := backup.ExportBackupCreateZIP(svc, pgData, mongoData)
	require.NoError(t, err)
	require.NotEmpty(t, zipData)

	// Проверяем содержимое ZIP
	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	require.NoError(t, err)
	assert.Len(t, reader.File, 2)

	names := make(map[string]bool)
	for _, f := range reader.File {
		names[f.Name] = true
		rc, err := f.Open()
		require.NoError(t, err)
		var buf bytes.Buffer
		_, err = buf.ReadFrom(rc)
		require.NoError(t, err)
		rc.Close()

		switch f.Name {
		case "postgres.dump":
			assert.Equal(t, pgData, buf.Bytes())
		case "mongo.archive":
			assert.Equal(t, mongoData, buf.Bytes())
		}
	}

	assert.True(t, names["postgres.dump"], "ZIP должен содержать postgres.dump")
	assert.True(t, names["mongo.archive"], "ZIP должен содержать mongo.archive")
}

// TestCreateZIPEmptyData проверяет архивирование пустых данных.
func TestCreateZIPEmptyData(t *testing.T) {
	svc := &backup.BackupService{}

	zipData, err := backup.ExportBackupCreateZIP(svc, []byte{}, []byte{})
	require.NoError(t, err)
	require.NotEmpty(t, zipData)

	reader, err := zip.NewReader(bytes.NewReader(zipData), int64(len(zipData)))
	require.NoError(t, err)
	assert.Len(t, reader.File, 2)
}
