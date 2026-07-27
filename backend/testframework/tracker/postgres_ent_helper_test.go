package tracker_test

import (
	"context"
	"database/sql"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/internal/tracker"
)

// postgresTestURL — DSN для Ent-тестов tracker, которые должны бить в реальный Postgres
// (не sqlite: HasSourceWith / HasIssueWith / issueProjectIDEQ обязаны проверяться на prod-движке).
// Без POSTGRES_TEST_URL или при недоступном Postgres — Skip (не Fatal), чтобы локальный
// прогон без БД компилировался и проходил чисто.
func postgresTestURL(t *testing.T) string {
	t.Helper()
	uri := os.Getenv("POSTGRES_TEST_URL")
	if uri == "" {
		t.Skip("POSTGRES_TEST_URL not set; skipping tracker Postgres Ent twin")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	db, err := sql.Open("postgres", uri)
	if err == nil {
		err = db.PingContext(ctx)
		_ = db.Close()
	}
	if err != nil {
		fallback := "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
		if uri != fallback {
			db2, err2 := sql.Open("postgres", fallback)
			if err2 == nil && db2.PingContext(ctx) == nil {
				_ = db2.Close()
				return fallback
			}
			if db2 != nil {
				_ = db2.Close()
			}
		}
		t.Skipf("PostgreSQL недоступен для tracker Postgres Ent twin (POSTGRES_TEST_URL=%q): %v", uri, err)
	}
	return uri
}

// openTrackerPostgresEnt создаёт изолированную БД + Ent-клиент на Postgres.
func openTrackerPostgresEnt(t *testing.T) *ent.Client {
	t.Helper()
	adminURI := postgresTestURL(t)
	dbName := "hub_tracker_" + strings.ReplaceAll(uuid.New().String(), "-", "")
	admin, err := sql.Open("postgres", adminURI)
	require.NoError(t, err)
	_, err = admin.Exec("CREATE DATABASE " + dbName)
	require.NoError(t, err)
	t.Cleanup(func() {
		_, _ = admin.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $1 AND pid <> pg_backend_pid()", dbName)
		_, _ = admin.Exec("DROP DATABASE IF EXISTS " + dbName)
		_ = admin.Close()
	})
	u, err := url.Parse(adminURI)
	require.NoError(t, err)
	u.Path = "/" + dbName
	client := enttest.Open(t, "postgres", u.String())
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// newTrackerPostgresRepoEnv — project + user + default workflow на Postgres (зеркало sqlite env.New).
func newTrackerPostgresRepoEnv(t *testing.T) trackerTestEnv {
	t.Helper()
	client := openTrackerPostgresEnt(t)
	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("tracker-pg-" + uuid.New().String()[:8] + "@test.com").
		SetName("Tracker PG").
		Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().
		SetName("Tracker PG Project").
		SetKey("TPG").
		Save(ctx)
	require.NoError(t, err)
	status, err := client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)
	return trackerTestEnv{
		client:    client,
		repo:      tracker.NewEntRepository(client, nil),
		ctx:       ctx,
		projectID: proj.ID,
		userID:    user.ID,
		statusID:  status.ID,
	}
}
