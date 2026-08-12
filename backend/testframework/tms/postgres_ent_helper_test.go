package tms_test

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
	"github.com/valerii/hub/backend/internal/tms"
)

// postgresTestURL — DSN для ent-тестов TMS, которые должны бить в реальный Postgres
// (не sqlite: VersionEQ/HasRunWith обязаны проверяться на том же движке, что prod).
func postgresTestURL(t *testing.T) string {
	t.Helper()
	uri := os.Getenv("POSTGRES_TEST_URL")
	if uri == "" {
		uri = "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
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
		t.Fatalf("PostgreSQL недоступен для TMS ent-теста (POSTGRES_TEST_URL=%q): %v", uri, err)
	}
	return uri
}

// openTMSPostgresEnt создаёт изолированную БД + Ent-клиент на Postgres.
func openTMSPostgresEnt(t *testing.T) *ent.Client {
	t.Helper()
	adminURI := postgresTestURL(t)
	dbName := "hub_tms_" + strings.ReplaceAll(uuid.New().String(), "-", "")
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

// newTMSPostgresRepoEnv — project + user + default workflow (для CreateBugIssue и т.п.).
func newTMSPostgresRepoEnv(t *testing.T) (*tms.EntRepository, context.Context, uuid.UUID, uuid.UUID) {
	t.Helper()
	client := openTMSPostgresEnt(t)
	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("tms-pg-" + uuid.New().String()[:8] + "@test.com").
		SetName("TMS PG").
		Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().
		SetName("TMS PG Project").
		SetKey("TMP").
		Save(ctx)
	require.NoError(t, err)
	_, err = client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)
	return tms.NewEntRepository(client), ctx, proj.ID, user.ID
}
