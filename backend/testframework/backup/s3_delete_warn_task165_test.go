//go:build hubtest

package backup_test

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/backuprecord"
	"github.com/valerii/hub/backend/internal/backup"
)

type failingDeleteS3Store struct {
	mu      sync.Mutex
	objects map[string][]byte
	deletes int
}

func newFailingDeleteS3Client(t *testing.T, bucket string) (*s3.Client, *failingDeleteS3Store) {
	t.Helper()
	store := &failingDeleteS3Store{objects: make(map[string][]byte)}
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
			store.deletes++
			store.mu.Unlock()
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("delete failed"))
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

type slogCaptureHandler struct {
	mu   sync.Mutex
	msgs []string
}

func (h *slogCaptureHandler) Enabled(context.Context, slog.Level) bool { return true }
func (h *slogCaptureHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.msgs = append(h.msgs, r.Message)
	return nil
}
func (h *slogCaptureHandler) WithAttrs([]slog.Attr) slog.Handler { return h }
func (h *slogCaptureHandler) WithGroup(string) slog.Handler      { return h }
func (h *slogCaptureHandler) has(substr string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, m := range h.msgs {
		if strings.Contains(m, substr) {
			return true
		}
	}
	return false
}

// TestDeleteBackup_S3FailClosedWarn_A3W74_002 — S3 DeleteObject err → Warn + DB row остаётся.
// Mutation: _, _ = DeleteObject без log/fail → err==nil и запись удалена.
func TestDeleteBackup_S3FailClosedWarn_A3W74_002(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	s3Client, store := newFailingDeleteS3Client(t, "hub-backups")
	backup.ExportBackupSetS3(svc, "hub-backups", s3Client)

	key := "backups/" + uuid.New().String() + ".zip"
	store.mu.Lock()
	store.objects[key] = []byte("zip")
	store.mu.Unlock()

	record, err := client.BackupRecord.Create().
		SetS3Key(key).
		SetFileSize(3).
		SetStatus(backuprecord.StatusCompleted).
		Save(ctx)
	require.NoError(t, err)

	cap := &slogCaptureHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(cap))
	t.Cleanup(func() { slog.SetDefault(prev) })

	err = svc.DeleteBackup(ctx, record.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "delete backup from s3")
	assert.True(t, cap.has("s3 delete backup failed"), "expected WarnContext message")

	still, err := client.BackupRecord.Get(ctx, record.ID)
	require.NoError(t, err)
	assert.Equal(t, record.ID, still.ID)
	assert.GreaterOrEqual(t, store.deletes, 1)
}

// TestCleanupOldBackups_S3FailWarn_A3W74_002 — cleanup при S3 err логирует Warn, DB row не удаляет.
func TestCleanupOldBackups_S3FailWarn_A3W74_002(t *testing.T) {
	svc, client, ctx := newBackupTestService(t)
	backup.ExportBackupSetMaxBackups(svc, 1)
	s3Client, store := newFailingDeleteS3Client(t, "hub-backups")
	backup.ExportBackupSetS3(svc, "hub-backups", s3Client)

	var ids []uuid.UUID
	for i := 0; i < 2; i++ {
		key := "old/" + uuid.New().String() + ".zip"
		store.mu.Lock()
		store.objects[key] = []byte("x")
		store.mu.Unlock()
		rec, err := client.BackupRecord.Create().
			SetS3Key(key).
			SetFileSize(1).
			SetStatus(backuprecord.StatusCompleted).
			Save(ctx)
		require.NoError(t, err)
		ids = append(ids, rec.ID)
	}

	cap := &slogCaptureHandler{}
	prev := slog.Default()
	slog.SetDefault(slog.New(cap))
	t.Cleanup(func() { slog.SetDefault(prev) })

	backup.ExportBackupCleanupOldBackups(svc, ctx)

	assert.True(t, cap.has("s3 delete old backup failed"), "expected Warn on cleanup S3 fail")
	for _, id := range ids {
		_, err := client.BackupRecord.Get(ctx, id)
		require.NoError(t, err, "DB row must remain when S3 delete fails")
	}
}
