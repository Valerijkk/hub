//go:build hubtest

package audit_test

import (
	"context"
	"database/sql"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/bson"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/audit"
	"github.com/valerii/hub/backend/testframework/mongotest"
)

const task225S3Bucket = "hub-recordings"

func openAuditPostgresTask225(t *testing.T) *ent.Client {
	t.Helper()

	adminURL := os.Getenv("POSTGRES_TEST_URL")
	if adminURL == "" {
		adminURL = "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
	}

	admin, err := sql.Open("postgres", adminURL)
	require.NoError(t, err)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	require.NoError(t, admin.PingContext(ctx), "PostgreSQL is required for TASK-225")

	dbName := "hub_audit_task225_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	_, err = admin.ExecContext(ctx, "CREATE DATABASE "+dbName)
	require.NoError(t, err)
	t.Cleanup(func() {
		cleanupCtx, cleanupCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cleanupCancel()
		_, _ = admin.ExecContext(cleanupCtx,
			"SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()",
			dbName,
		)
		_, _ = admin.ExecContext(cleanupCtx, "DROP DATABASE IF EXISTS "+dbName)
		_ = admin.Close()
	})

	dsn, err := url.Parse(adminURL)
	require.NoError(t, err)
	dsn.Path = "/" + dbName
	client := enttest.Open(t, "postgres", dsn.String())
	t.Cleanup(func() { _ = client.Close() })
	return client
}

type retentionS3Task225 struct {
	mu       sync.Mutex
	statuses map[string]int
	deletes  map[string]int
}

func (s *retentionS3Task225) setStatus(key string, status int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.statuses[key] = status
}

func (s *retentionS3Task225) deleteCount(key string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deletes[key]
}

func newRetentionS3Task225(t *testing.T) (*s3.Client, *retentionS3Task225) {
	t.Helper()

	store := &retentionS3Task225{
		statuses: make(map[string]int),
		deletes:  make(map[string]int),
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		key := strings.TrimPrefix(r.URL.Path, "/"+task225S3Bucket+"/")
		store.mu.Lock()
		store.deletes[key]++
		status := store.statuses[key]
		store.mu.Unlock()
		if status == 0 {
			status = http.StatusNoContent
		}
		if status >= http.StatusBadRequest {
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(status)
			_, _ = w.Write([]byte(`<Error><Code>InternalError</Code><Message>delete failed</Message></Error>`))
			return
		}
		w.WriteHeader(status)
	}))
	t.Cleanup(server.Close)

	cfg, err := awsconfig.LoadDefaultConfig(context.Background(),
		awsconfig.WithRegion("us-east-1"),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider("task225", "task225", "")),
	)
	require.NoError(t, err)
	client := s3.NewFromConfig(cfg, func(options *s3.Options) {
		options.BaseEndpoint = aws.String(server.URL)
		options.UsePathStyle = true
	})
	return client, store
}

// TestRetentionCleanup_MongoPostgresS3FailSoft_CoverageGate proves the documented
// retention contract against real MongoDB, PostgreSQL and the S3 HTTP protocol.
func TestRetentionCleanup_MongoPostgresS3FailSoft_CoverageGate(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	mongoDB := mongotest.Connect(t, "audit_retention_task225_")
	postgres := openAuditPostgresTask225(t)
	s3Client, s3Store := newRetentionS3Task225(t)

	now := time.Now().UTC()
	oldMessage := bson.M{"marker": "old-message", "created_at": now.Add(-800 * 24 * time.Hour)}
	freshMessage := bson.M{"marker": "fresh-message", "created_at": now.Add(-24 * time.Hour)}
	_, err := mongoDB.Collection("messages").InsertMany(ctx, []any{oldMessage, freshMessage})
	require.NoError(t, err)

	oldAllure := bson.M{"marker": "old-allure", "created_at": now.Add(-400 * 24 * time.Hour)}
	freshAllure := bson.M{"marker": "fresh-allure", "created_at": now.Add(-24 * time.Hour)}
	_, err = mongoDB.Collection("allure_results").InsertMany(ctx, []any{oldAllure, freshAllure})
	require.NoError(t, err)

	deletedFromS3 := seedRecordingFixture(t, postgres, now.Add(-200*24*time.Hour))
	s3Failure := seedRecordingFixture(t, postgres, now.Add(-201*24*time.Hour))
	freshRecording := seedRecordingFixture(t, postgres, now.Add(-24*time.Hour))
	s3Store.setStatus(s3Failure.FilePath, http.StatusInternalServerError)

	service := audit.NewRetentionService(postgres, mongoDB, s3Client, task225S3Bucket)
	require.NoError(t, service.EnsureIndexes(ctx))
	require.NoError(t, service.RunRetentionCleanup(ctx), "S3 failure must not block retention")

	oldMessages, err := mongoDB.Collection("messages").CountDocuments(ctx, bson.M{"marker": "old-message"})
	require.NoError(t, err)
	freshMessages, err := mongoDB.Collection("messages").CountDocuments(ctx, bson.M{"marker": "fresh-message"})
	require.NoError(t, err)
	assert.Equal(t, int64(0), oldMessages)
	assert.Equal(t, int64(1), freshMessages)

	oldAllureResults, err := mongoDB.Collection("allure_results").CountDocuments(ctx, bson.M{"marker": "old-allure"})
	require.NoError(t, err)
	freshAllureResults, err := mongoDB.Collection("allure_results").CountDocuments(ctx, bson.M{"marker": "fresh-allure"})
	require.NoError(t, err)
	assert.Equal(t, int64(0), oldAllureResults)
	assert.Equal(t, int64(1), freshAllureResults)

	_, err = postgres.Recording.Get(ctx, deletedFromS3.ID)
	assert.True(t, ent.IsNotFound(err), "expired row must be deleted after S3 204")
	_, err = postgres.Recording.Get(ctx, s3Failure.ID)
	assert.True(t, ent.IsNotFound(err), "expired row must be deleted after S3 500 (fail-soft)")
	kept, err := postgres.Recording.Get(ctx, freshRecording.ID)
	require.NoError(t, err)
	assert.Equal(t, freshRecording.ID, kept.ID)

	assert.GreaterOrEqual(t, s3Store.deleteCount(deletedFromS3.FilePath), 1)
	assert.GreaterOrEqual(t, s3Store.deleteCount(s3Failure.FilePath), 1, "SDK retries are allowed")
	assert.Equal(t, 0, s3Store.deleteCount(freshRecording.FilePath))
}

type channelMessageEraserTask225 struct {
	mu      sync.Mutex
	events  chan uuid.UUID
	deleted []uuid.UUID
}

func (e *channelMessageEraserTask225) DeleteMessagesBySender(_ context.Context, senderID uuid.UUID) (int64, error) {
	e.mu.Lock()
	e.deleted = append(e.deleted, senderID)
	e.mu.Unlock()
	e.events <- senderID
	return 1, nil
}

func (e *channelMessageEraserTask225) calls() []uuid.UUID {
	e.mu.Lock()
	defer e.mu.Unlock()
	return append([]uuid.UUID(nil), e.deleted...)
}

// TestMongoEraseRetry_JetStreamLifecycle_CoverageGate exercises Start, durable
// enqueue, real consumer delivery and Stop without timing sleeps.
func TestMongoEraseRetry_JetStreamLifecycle_CoverageGate(t *testing.T) {
	nc := startAuditTestNATS(t)
	eraser := &channelMessageEraserTask225{events: make(chan uuid.UUID, 1)}
	consumer := audit.NewMongoEraseRetryConsumer(nc, eraser)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	require.NoError(t, consumer.Start(ctx))
	t.Cleanup(consumer.Stop)

	senderID := uuid.New()
	actorID := uuid.New()
	retry := audit.NewJetStreamMongoEraseRetry(nc)
	require.NoError(t, retry.EnqueueMongoErase(ctx, senderID, actorID))

	select {
	case delivered := <-eraser.events:
		assert.Equal(t, senderID, delivered)
	case <-ctx.Done():
		t.Fatalf("timed out waiting for JetStream retry delivery: %v", ctx.Err())
	}

	assert.Equal(t, []uuid.UUID{senderID}, eraser.calls())
	consumer.Stop()
}
