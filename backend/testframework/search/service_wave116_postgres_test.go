package search_test
import (
	"github.com/valerii/hub/backend/internal/search"
	"context"
	"database/sql"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var postgresSearchMu sync.Mutex

func connectPostgresSearchTest(t *testing.T) *sql.DB {
	t.Helper()

	uri := os.Getenv("POSTGRES_TEST_URL")
	if uri == "" {
		uri = "postgresql://hub:hub@postgres:5432/hub?sslmode=disable"
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := sql.Open("postgres", uri)
	require.NoError(t, err)

	if err := db.PingContext(ctx); err != nil {
		// Запасной вариант для локальных прогонов против опубликованного порта.
		fallback := "postgresql://hub:hub@localhost:5432/hub?sslmode=disable"
		if uri != fallback {
			db2, err2 := sql.Open("postgres", fallback)
			if err2 == nil && db2.PingContext(ctx) == nil {
				_ = db.Close()
				return db2
			}
			if db2 != nil {
				_ = db2.Close()
			}
		}
		t.Skip("PostgreSQL недоступен для search integration:", err)
	}
	return db
}

func resetPostgresSearchSchema(t *testing.T, db *sql.DB) {
	t.Helper()
	postgresSearchMu.Lock()
	t.Cleanup(func() { postgresSearchMu.Unlock() })

	ctx := context.Background()
	stmts := []string{
		`DROP TABLE IF EXISTS file_metadata, test_cases, issues, channel_members, channels, project_members, projects, users CASCADE`,
		`CREATE TABLE issues (
			id UUID PRIMARY KEY,
			issue_key TEXT NOT NULL,
			title TEXT NOT NULL,
			description_text TEXT,
			project_issues UUID NOT NULL,
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE test_cases (
			id UUID PRIMARY KEY,
			case_key TEXT NOT NULL,
			title TEXT NOT NULL,
			description TEXT,
			project_test_cases UUID NOT NULL,
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE users (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			email TEXT NOT NULL,
			display_name TEXT
		)`,
		`CREATE TABLE projects (
			id UUID PRIMARY KEY,
			key TEXT NOT NULL,
			name TEXT NOT NULL,
			description TEXT,
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE file_metadata (
			id UUID PRIMARY KEY,
			filename TEXT NOT NULL,
			project_files UUID NOT NULL,
			status TEXT NOT NULL DEFAULT 'confirmed',
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE project_members (
			project_members UUID NOT NULL,
			user_project_members UUID NOT NULL
		)`,
		`CREATE TABLE channels (
			id UUID PRIMARY KEY,
			name TEXT NOT NULL,
			channel_type TEXT NOT NULL,
			project_id UUID,
			deleted_at TIMESTAMPTZ
		)`,
		`CREATE TABLE channel_members (
			channel_id UUID NOT NULL,
			user_id UUID NOT NULL
		)`,
	}
	for _, stmt := range stmts {
		_, err := db.ExecContext(ctx, stmt)
		require.NoError(t, err, stmt)
	}
	require.NoError(t, search.EnsureSearchIndexes(ctx, db))
}

func TestPostgresSearch_IssuesAndProjects(t *testing.T) {
	db := connectPostgresSearchTest(t)
	defer db.Close()
	resetPostgresSearchSchema(t, db)

	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	issueID := uuid.New()

	_, err := db.ExecContext(ctx,
		`INSERT INTO projects (id, key, name, description) VALUES ($1, $2, $3, $4)`,
		projectID, "HUB", "Hub Platform", "dev team tool")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		`INSERT INTO project_members (project_members, user_project_members) VALUES ($1, $2)`,
		projectID, userID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		`INSERT INTO issues (id, issue_key, title, description_text, project_issues) VALUES ($1, $2, $3, $4, $5)`,
		issueID, "HUB-42", "Fix login flow", "authentication bug in magic link", projectID)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)

	issues, err := search.ExportSearchIssues(svc, ctx, "login", []uuid.UUID{projectID})
	require.NoError(t, err)
	require.NotEmpty(t, issues)
	assert.Equal(t, "issue", issues[0].Type)
	assert.Contains(t, issues[0].Title, "HUB-42")

	projects, err := search.ExportSearchProjects(svc, ctx, "Hub", []uuid.UUID{projectID})
	require.NoError(t, err)
	require.NotEmpty(t, projects)
	assert.Equal(t, "project", projects[0].Type)

	results, err := svc.Search(ctx, "login", userID, []string{"issue", "project"}, 10)
	require.NoError(t, err)
	assert.NotEmpty(t, results)
}

func TestPostgresSearch_TestCasesAndFiles(t *testing.T) {
	db := connectPostgresSearchTest(t)
	defer db.Close()
	resetPostgresSearchSchema(t, db)

	ctx := context.Background()
	projectID := uuid.New()

	_, err := db.ExecContext(ctx,
		`INSERT INTO test_cases (id, case_key, title, description, project_test_cases) VALUES ($1, $2, $3, $4, $5)`,
		uuid.New(), "TC-1", "Verify checkout", "payment regression case", projectID)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		`INSERT INTO file_metadata (id, filename, project_files) VALUES ($1, $2, $3)`,
		uuid.New(), "architecture diagram.pdf", projectID)
	require.NoError(t, err)

	svc := search.ExportNewSearchServiceSQL(db)

	cases, err := search.ExportSearchTestCases(svc, ctx, "Verify", []uuid.UUID{projectID})
	require.NoError(t, err)
	require.NotEmpty(t, cases)
	assert.Equal(t, "test_case", cases[0].Type)

	files, err := search.ExportSearchFiles(svc, ctx, "architecture", []uuid.UUID{projectID})
	require.NoError(t, err)
	require.NotEmpty(t, files)
	assert.Equal(t, "file", files[0].Type)
}

func TestPostgresSearch_UsersAdminAndScoped(t *testing.T) {
	db := connectPostgresSearchTest(t)
	defer db.Close()
	resetPostgresSearchSchema(t, db)

	ctx := context.Background()
	projectID := uuid.New()
	callerID := uuid.New()
	peerID := uuid.New()
	outsiderID := uuid.New()

	for _, u := range []struct {
		id    uuid.UUID
		name  string
		email string
	}{
		{callerID, "Alice Caller", "alice@hub.local"},
		{peerID, "Bob Peer", "bob@hub.local"},
		{outsiderID, "Carol Outsider", "carol@hub.local"},
	} {
		_, err := db.ExecContext(ctx,
			`INSERT INTO users (id, name, email, display_name) VALUES ($1, $2, $3, $4)`,
			u.id, u.name, u.email, u.name)
		require.NoError(t, err)
	}
	_, err := db.ExecContext(ctx,
		`INSERT INTO project_members (project_members, user_project_members) VALUES ($1, $2), ($1, $3)`,
		projectID, callerID, peerID)
	require.NoError(t, err)

	adminSvc := search.NewSearchService(&adminSearchRepo{admin: true}, db, nil)
	all, err := search.ExportSearchUsers(adminSvc, ctx, "Alice", callerID)
	require.NoError(t, err)
	require.NotEmpty(t, all)

	memberSvc := search.NewSearchService(&mockSearchRepo{isAdmin: false}, db, nil)
	scoped, err := search.ExportSearchUsers(memberSvc, ctx, "Bob", callerID)
	require.NoError(t, err)
	require.NotEmpty(t, scoped)
	for _, r := range scoped {
		assert.NotEqual(t, outsiderID.String(), r.ID)
	}

	unscoped, err := search.ExportSearchUsersUnscoped(memberSvc, ctx, "Carol")
	require.NoError(t, err)
	require.NotEmpty(t, unscoped)
}

func TestPostgresSearch_Channels(t *testing.T) {
	db := connectPostgresSearchTest(t)
	defer db.Close()
	resetPostgresSearchSchema(t, db)

	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	channelID := uuid.New()

	_, err := db.ExecContext(ctx,
		`INSERT INTO channels (id, name, channel_type, project_id) VALUES ($1, $2, $3, $4)`,
		channelID, "general-dev", "public", projectID)
	require.NoError(t, err)

	svc := search.ExportNewSearchServiceSQL(db)
	channels, err := search.ExportSearchChannels(svc, ctx, "general", []uuid.UUID{projectID}, userID)
	require.NoError(t, err)
	require.NotEmpty(t, channels)
	assert.Equal(t, "channel", channels[0].Type)
	assert.Contains(t, channels[0].Title, "general")
}

// TestPostgresSearch_Channels_SoftDeleteAndPrivateScope_P1BE32 — mutation-ready
// guard for P1-BE-32: soft-deleted → 0 hits; private non-member → не в выдаче.
func TestPostgresSearch_Channels_SoftDeleteAndPrivateScope_P1BE32(t *testing.T) {
	db := connectPostgresSearchTest(t)
	defer db.Close()
	resetPostgresSearchSchema(t, db)

	ctx := context.Background()
	projectID := uuid.New()
	memberID := uuid.New()
	outsiderID := uuid.New()
	publicLive := uuid.New()
	publicDeleted := uuid.New()
	privateMember := uuid.New()
	privateSecret := uuid.New()
	deletedAt := time.Now().UTC()

	_, err := db.ExecContext(ctx,
		`INSERT INTO channels (id, name, channel_type, project_id, deleted_at) VALUES
			($1, 'p1be32 public live', 'public', $5, NULL),
			($2, 'p1be32 public deleted', 'public', $5, $6),
			($3, 'p1be32 private member', 'private', $5, NULL),
			($4, 'p1be32 private secret', 'private', $5, NULL)`,
		publicLive, publicDeleted, privateMember, privateSecret, projectID, deletedAt)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		`INSERT INTO channel_members (channel_id, user_id) VALUES ($1, $2)`,
		privateMember, memberID)
	require.NoError(t, err)

	svc := search.ExportNewSearchServiceSQL(db)

	asMember, err := search.ExportSearchChannels(svc, ctx, "p1be32", []uuid.UUID{projectID}, memberID)
	require.NoError(t, err)
	require.Len(t, asMember, 2, "member sees public live + own private; not deleted/secret")
	got := map[string]string{}
	for _, r := range asMember {
		got[r.ID] = r.Subtitle
	}
	assert.Equal(t, "public", got[publicLive.String()])
	assert.Equal(t, "private", got[privateMember.String()])
	assert.NotContains(t, got, publicDeleted.String())
	assert.NotContains(t, got, privateSecret.String())

	asOutsider, err := search.ExportSearchChannels(svc, ctx, "p1be32", []uuid.UUID{projectID}, outsiderID)
	require.NoError(t, err)
	require.Len(t, asOutsider, 1, "project peer without channel membership sees only public")
	assert.Equal(t, publicLive.String(), asOutsider[0].ID)
	assert.Equal(t, "public", asOutsider[0].Subtitle)
}

func TestPostgresEnsureSearchIndexes_Idempotent(t *testing.T) {
	db := connectPostgresSearchTest(t)
	defer db.Close()
	resetPostgresSearchSchema(t, db)

	ctx := context.Background()
	require.NoError(t, search.EnsureSearchIndexes(ctx, db))
	require.NoError(t, search.EnsureSearchIndexes(ctx, db))
}
