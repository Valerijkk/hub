package search_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/search"
)

// TestPostgresSearch_ExcludesSoftDeleted_A4_002 — runtime proof that Ctrl+K
// shards omit soft-deleted issues/files/test_cases/projects (A2W5-A5-002).
// Mutation: drop deleted_at / status filters → FAIL.
func TestPostgresSearch_ExcludesSoftDeleted_A4_002(t *testing.T) {
	db := connectPostgresSearchTest(t)
	defer db.Close()
	resetPostgresSearchSchema(t, db)

	ctx := context.Background()
	projectID := uuid.New()
	activeIssueID := uuid.New()
	deletedIssueID := uuid.New()
	activeFileID := uuid.New()
	deletedFileID := uuid.New()
	pendingFileID := uuid.New()
	activeCaseID := uuid.New()
	deletedCaseID := uuid.New()
	activeProjectID := projectID
	deletedProjectID := uuid.New()
	deletedAt := time.Now().UTC()

	_, err := db.ExecContext(ctx,
		`INSERT INTO projects (id, key, name, description, deleted_at) VALUES
			($1, $2, $3, $4, NULL), ($5, $6, $7, $8, $9)`,
		activeProjectID, "SDF", "softdel active project", "live project for ctrlk",
		deletedProjectID, "SDX", "softdel deleted project", "trashed project must hide", deletedAt)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO issues (id, issue_key, title, description_text, project_issues, deleted_at)
		 VALUES ($1, $2, $3, $4, $5, NULL), ($6, $7, $8, $9, $5, $10)`,
		activeIssueID, "SDF-1", "softdel active issue", "live row for ctrlk", projectID,
		deletedIssueID, "SDF-2", "softdel deleted issue", "trashed row must hide", deletedAt)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO test_cases (id, case_key, title, description, project_test_cases, deleted_at)
		 VALUES ($1, $2, $3, $4, $5, NULL), ($6, $7, $8, $9, $5, $10)`,
		activeCaseID, "SDF-C1", "softdel active case", "live case", projectID,
		deletedCaseID, "SDF-C2", "softdel deleted case", "trashed case", deletedAt)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO file_metadata (id, filename, project_files, status, deleted_at)
		 VALUES ($1, $2, $3, 'confirmed', NULL),
		        ($4, $5, $3, 'confirmed', $6),
		        ($7, $8, $3, 'pending', NULL)`,
		activeFileID, "softdel active report.pdf", projectID,
		deletedFileID, "softdel deleted report.pdf", deletedAt,
		pendingFileID, "softdel pending report.pdf")
	require.NoError(t, err)

	svc := search.ExportNewSearchServiceSQL(db)

	issues, err := search.ExportSearchIssues(svc, ctx, "softdel", []uuid.UUID{projectID})
	require.NoError(t, err)
	require.Len(t, issues, 1, "soft-deleted issue must be excluded")
	assert.Equal(t, activeIssueID.String(), issues[0].ID)
	assert.Contains(t, issues[0].Title, "SDF-1")

	files, err := search.ExportSearchFiles(svc, ctx, "softdel", []uuid.UUID{projectID})
	require.NoError(t, err)
	require.Len(t, files, 1, "soft-deleted and pending files must be excluded")
	assert.Equal(t, activeFileID.String(), files[0].ID)
	assert.Equal(t, "softdel active report.pdf", files[0].Title)

	cases, err := search.ExportSearchTestCases(svc, ctx, "softdel", []uuid.UUID{projectID})
	require.NoError(t, err)
	require.Len(t, cases, 1, "soft-deleted test_case must be excluded")
	assert.Equal(t, activeCaseID.String(), cases[0].ID)

	projects, err := search.ExportSearchProjects(svc, ctx, "softdel", []uuid.UUID{activeProjectID, deletedProjectID})
	require.NoError(t, err)
	require.Len(t, projects, 1, "soft-deleted project must be excluded")
	assert.Equal(t, activeProjectID.String(), projects[0].ID)
}
