package sprint_test

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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/sprint"
)

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
		t.Fatalf("PostgreSQL недоступен для sprint ent-теста: %v", err)
	}
	return uri
}

func openSprintPostgresEnt(t *testing.T) *ent.Client {
	t.Helper()
	adminURI := postgresTestURL(t)
	dbName := "hub_sprint_" + strings.ReplaceAll(uuid.New().String(), "-", "")
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

// TestEntRepository_GetSprintInProject_RejectsForeign_A3_001 — sprint чужого проекта
// не находится. Mutation: снять HasProjectWith в GetSprintInProject → тест краснеет.
func TestEntRepository_GetSprintInProject_RejectsForeign_A3_001(t *testing.T) {
	client := openSprintPostgresEnt(t)
	ctx := context.Background()
	repo := sprint.NewEntRepository(client)

	projA, err := client.Project.Create().SetName("Sprint Project A").SetKey("SPA").Save(ctx)
	require.NoError(t, err)
	projB, err := client.Project.Create().SetName("Sprint Project B").SetKey("SPB").Save(ctx)
	require.NoError(t, err)

	start := time.Now().UTC().Truncate(time.Second)
	end := start.AddDate(0, 0, 14)
	spB, err := repo.CreateSprint(ctx, projB.ID, sprint.CreateSprintRequest{
		Name: "Foreign", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)

	_, err = repo.GetSprintInProject(ctx, projA.ID, spB.ID)
	require.True(t, ent.IsNotFound(err), "sprint проекта B не должен находиться через project A, got %v", err)

	got, err := repo.GetSprintInProject(ctx, projB.ID, spB.ID)
	require.NoError(t, err)
	assert.Equal(t, spB.ID, got.ID)
}
