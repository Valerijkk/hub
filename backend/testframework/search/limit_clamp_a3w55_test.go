//go:build hubtest

package search_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/search"
)

// A3W5-5-001 / TASK-157: clamp limit в Search доказывается по Len результата.
// Канон HEAD: limit <= 0 || limit > 50 → 20.
// Per-shard defaultSearchLimit=10, поэтому сейдим ≥3 шарда (issues/TC/files),
// иначе суммарно ≤10 и Len==20 flaky/недостижим.
// Mutation: service.go clamp (`limit > 50` → `limit > 500` / убрать / default≠20) → Len ≠ 20.
func TestPostgresSearch_LimitClamp_A3W55_001(t *testing.T) {
	db := connectPostgresSearchTest(t)
	defer db.Close()
	resetPostgresSearchSchema(t, db)

	ctx := context.Background()
	userID := uuid.New()
	projectID := uuid.New()
	const perShard = 10
	const wantCap = 20

	_, err := db.ExecContext(ctx,
		`INSERT INTO projects (id, key, name, description) VALUES ($1, $2, $3, $4)`,
		projectID, "LMT", "Limit Clamp Project", "seed for search limit")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx,
		`INSERT INTO project_members (project_members, user_project_members) VALUES ($1, $2)`,
		projectID, userID)
	require.NoError(t, err)

	for i := 0; i < perShard; i++ {
		n := i + 1
		_, err = db.ExecContext(ctx,
			`INSERT INTO issues (id, issue_key, title, description_text, project_issues)
			 VALUES ($1, $2, $3, $4, $5)`,
			uuid.New(),
			fmt.Sprintf("LMT-%d", n),
			fmt.Sprintf("clampseed issue hit %d", n),
			"limit clamp mutation seed row",
			projectID,
		)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx,
			`INSERT INTO test_cases (id, case_key, title, description, project_test_cases)
			 VALUES ($1, $2, $3, $4, $5)`,
			uuid.New(),
			fmt.Sprintf("TC-LMT-%d", n),
			fmt.Sprintf("clampseed case hit %d", n),
			"limit clamp mutation seed row",
			projectID,
		)
		require.NoError(t, err)

		_, err = db.ExecContext(ctx,
			`INSERT INTO file_metadata (id, filename, project_files, status)
			 VALUES ($1, $2, $3, 'confirmed')`,
			uuid.New(),
			fmt.Sprintf("clampseed file %d.pdf", n),
			projectID,
		)
		require.NoError(t, err)
	}

	svc := search.NewSearchService(&mockSearchRepo{}, db, nil)
	projectIDs := []uuid.UUID{projectID}

	issues, err := search.ExportSearchIssues(svc, ctx, "clampseed", projectIDs)
	require.NoError(t, err)
	cases, err := search.ExportSearchTestCases(svc, ctx, "clampseed", projectIDs)
	require.NoError(t, err)
	files, err := search.ExportSearchFiles(svc, ctx, "clampseed", projectIDs)
	require.NoError(t, err)
	preClamp := len(issues) + len(cases) + len(files)
	require.Greater(t, preClamp, wantCap,
		"seed must yield >%d searchable hits across shards before Search clamp (got issues=%d cases=%d files=%d)",
		wantCap, len(issues), len(cases), len(files))

	types := []string{"issue", "test_case", "file"}
	for _, limit := range []int{100, 0, -1} {
		results, _, err := svc.Search(ctx, "clampseed", userID, types, limit)
		require.NoError(t, err, "limit=%d", limit)
		require.Len(t, results, wantCap, "limit=%d must clamp to default %d", limit, wantCap)
	}

	// Контроль: валидный limit внутри диапазона не подменяется default'ом.
	results, _, err := svc.Search(ctx, "clampseed", userID, types, 15)
	require.NoError(t, err)
	require.Len(t, results, 15)
}
