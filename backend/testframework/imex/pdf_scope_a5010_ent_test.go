package imex_test

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
	"github.com/valerii/hub/backend/internal/imex"
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
		t.Fatalf("PostgreSQL недоступен для imex A5-010 ent-теста (POSTGRES_TEST_URL=%q): %v", uri, err)
	}
	return uri
}

func openImexPostgresEnt(t *testing.T) *ent.Client {
	t.Helper()
	adminURI := postgresTestURL(t)
	dbName := "hub_imex_" + strings.ReplaceAll(uuid.New().String(), "-", "")
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

// TestEntRepository_PDFExport_RejectsForeignRunAndSprint_A5_010 — run/sprint чужого
// проекта не находятся при PDF-export scope. Mutation: снять HasProjectWith в
// GetTestRunInProject / GetSprintInProject → тест краснеет.
func TestEntRepository_PDFExport_RejectsForeignRunAndSprint_A5_010(t *testing.T) {
	client := openImexPostgresEnt(t)
	ctx := context.Background()
	repo := imex.NewEntRepository(client)

	user, err := client.User.Create().
		SetEmail("a5010-" + uuid.New().String()[:8] + "@test.com").
		SetName("A5010").
		Save(ctx)
	require.NoError(t, err)

	projA, err := client.Project.Create().SetName("Imex PDF Project A").SetKey("IPA").Save(ctx)
	require.NoError(t, err)
	projB, err := client.Project.Create().SetName("Imex PDF Project B").SetKey("IPB").Save(ctx)
	require.NoError(t, err)

	runB, err := client.TestRun.Create().
		SetTitle("Foreign run").
		SetProjectID(projB.ID).
		SetCreatedByID(user.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = repo.GetTestRunInProject(ctx, projA.ID, runB.ID)
	require.True(t, ent.IsNotFound(err), "run проекта B не должен находиться через project A, got %v", err)

	gotRun, err := repo.GetTestRunInProject(ctx, projB.ID, runB.ID)
	require.NoError(t, err)
	assert.Equal(t, runB.ID, gotRun.ID)

	now := time.Now().UTC().Truncate(time.Second)
	spB, err := client.Sprint.Create().
		SetName("Foreign sprint").
		SetStartDate(now).
		SetEndDate(now.AddDate(0, 0, 14)).
		SetProjectID(projB.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = repo.GetSprintInProject(ctx, projA.ID, spB.ID)
	require.True(t, ent.IsNotFound(err), "sprint проекта B не должен находиться через project A, got %v", err)

	gotSprint, err := repo.GetSprintInProject(ctx, projB.ID, spB.ID)
	require.NoError(t, err)
	assert.Equal(t, spB.ID, gotSprint.ID)
}
