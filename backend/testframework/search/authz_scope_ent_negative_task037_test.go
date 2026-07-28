//go:build hubtest

package search_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/search"
)

// TASK-037 / P2-BE-02 / W5-A5-001 — Postgres project-scope: foreign issue не в выдаче.
// Mutation: убрать project_issues = ANY($2) → Len==2 / foreign ID в results.

func TestPostgresSearch_RejectsForeignProjectIssues_TASK037(t *testing.T) {
	db := connectPostgresSearchTest(t)
	defer db.Close()
	resetPostgresSearchSchema(t, db)

	ctx := context.Background()
	userID := uuid.New()
	allowedProject := uuid.New()
	foreignProject := uuid.New()
	allowedIssue := uuid.New()
	foreignIssue := uuid.New()

	_, err := db.ExecContext(ctx,
		`INSERT INTO projects (id, key, name, description) VALUES ($1, $2, $3, $4), ($5, $6, $7, $8)`,
		allowedProject, "ALW", "Allowed", "scope allowed",
		foreignProject, "FRN", "Foreign", "scope foreign",
	)
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		`INSERT INTO project_members (project_members, user_project_members) VALUES ($1, $2)`,
		allowedProject, userID)
	require.NoError(t, err)

	_, err = db.ExecContext(ctx,
		`INSERT INTO issues (id, issue_key, title, description_text, project_issues) VALUES
			($1, $2, $3, $4, $5),
			($6, $7, $8, $9, $10)`,
		allowedIssue, "ALW-1", "shared-query login allowed", "auth magic", allowedProject,
		foreignIssue, "FRN-1", "shared-query login foreign", "auth magic", foreignProject,
	)
	require.NoError(t, err)

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)

	// Direct shard with only allowed projectIDs.
	issues, err := search.ExportSearchIssues(svc, ctx, "shared-query", []uuid.UUID{allowedProject})
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Contains(t, issues[0].Title, "ALW-1")
	assert.NotContains(t, issues[0].Title, "FRN-1")

	// Full Search resolves membership → same scope.
	results, _, err := svc.Search(ctx, "shared-query", userID, []string{"issue"}, 20)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Contains(t, results[0].Title, "ALW-1")
	for _, r := range results {
		assert.NotEqual(t, foreignIssue.String(), r.ID)
	}
}
