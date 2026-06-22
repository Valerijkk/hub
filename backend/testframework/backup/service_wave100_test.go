package backup_test
import (
	"github.com/valerii/hub/backend/internal/backup"
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/backuprecord"
)

type fakeS3Store struct {
	mu      sync.Mutex
	objects map[string][]byte
}

func newFakeS3Client(t *testing.T, bucket string) (*s3.Client, *fakeS3Store) {
	t.Helper()
	store := &fakeS3Store{objects: make(map[string][]byte)}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key := strings.TrimPrefix(r.URL.Path, "/"+bucket+"/")
		if key == r.URL.Path {
			key = strings.TrimPrefix(r.URL.Path, "/")
		}
		switch r.Method {
		case http.MethodPut:
			body, _ := io.ReadAll(r.Body)
			store.mu.Lock()
			store.objects[key] = body
			store.mu.Unlock()
			w.WriteHeader(http.StatusOK)
		case http.MethodGet:
			store.mu.Lock()
			data, ok := store.objects[key]
			store.mu.Unlock()
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write(data)
		case http.MethodDelete:
			store.mu.Lock()
			delete(store.objects, key)
			store.mu.Unlock()
			w.WriteHeader(http.StatusNoContent)
		default:
			w.WriteHeader(http.StatusMethodNotAllowed)
		}
	}))
	t.Cleanup(srv.Close)

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("test", "test", "")),
	)
	require.NoError(t, err)

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(srv.URL)
		o.UsePathStyle = true
	})
	return client, store
}

func TestPerformBackup_UploadsToS3(t *testing.T) {
	installFakeBackupBinaries(t, false)

	svc, client, ctx := newBackupTestService(t)
	backup.ExportBackupSetURLs(svc, "postgres://localhost/hub_test", "mongodb://localhost:27017/hub_test")
	s3Client, _ := newFakeS3Client(t, "hub-backups")
	backup.ExportBackupSetS3(svc, "hub-backups", s3Client)

	resp, err := svc.PerformBackup(ctx)
	require.NoError(t, err)
	require.Equal(t, "completed", resp.Status)

	records, err := client.BackupRecord.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, records, 1)
	assert.Equal(t, backuprecord.StatusCompleted, records[0].Status)
}

func TestRestoreBackup_FromS3Zip(t *testing.T) {
	installFakeBackupBinaries(t, false)

	svc, client, ctx := newBackupTestService(t)
	backup.ExportBackupSetURLs(svc, "postgres://localhost/hub_test", "mongodb://localhost:27017/hub_test")
	s3Client, store := newFakeS3Client(t, "hub-backups")
	backup.ExportBackupSetS3(svc, "hub-backups", s3Client)

	zipData, err := backup.ExportBackupCreateZIP(svc, []byte("pg-dump"), []byte("mongo-archive"))
	require.NoError(t, err)
	s3Key := "backups/restore-me.zip"
	store.mu.Lock()
	store.objects[s3Key] = zipData
	store.mu.Unlock()

	record, err := client.BackupRecord.Create().
		SetS3Key(s3Key).
		SetStatus(backuprecord.StatusCompleted).
		SetFileSize(int64(len(zipData))).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.RestoreBackup(ctx, record.ID))

	updated, err := client.BackupRecord.Get(ctx, record.ID)
	require.NoError(t, err)
	require.NotNil(t, updated.ErrorMessage)
	assert.Contains(t, *updated.ErrorMessage, "restored at")
}

func TestRestoreBackup_S3GetObjectFails(t *testing.T) {
	installFakeBackupBinaries(t, false)

	svc, client, ctx := newBackupTestService(t)
	s3Client, _ := newFakeS3Client(t, "hub-backups")
	backup.ExportBackupSetS3(svc, "hub-backups", s3Client)

	record, err := client.BackupRecord.Create().
		SetS3Key("backups/missing.zip").
		SetStatus(backuprecord.StatusCompleted).
		Save(ctx)
	require.NoError(t, err)

	err = svc.RestoreBackup(ctx, record.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "S3")
}

func TestExtractZIP_ReadErrorOnCorruptEntry(t *testing.T) {
	svc := &backup.BackupService{}
	// Корректный zip-заголовок, но усечённое содержимое вызывает ошибку чтения при open.
	buf := new(bytes.Buffer)
	buf.Write([]byte("PK\x03\x04"))
	_, _, err := backup.ExportBackupExtractZIP(svc, buf.Bytes())
	require.Error(t, err)
}
