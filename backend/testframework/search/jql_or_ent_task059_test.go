//go:build hubtest

package search_test

// TASK-059 / P3-BE-03 / W5-A4-012: OR JQL → Ent union count; AND → intersection.
// Mutation: игнор Logic (всегда And) → OR count == AND count → FAIL.

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/internal/search"
)

func TestToEntPredicates_ORUnionVsAND_EntCount_W5A4012(t *testing.T) {
	dsn := "file:jql_or_059_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	u, err := client.User.Create().SetEmail("jql-or-" + uuid.New().String()[:8] + "@t.com").Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().SetName("OR Proj").SetKey("ORP").Save(ctx)
	require.NoError(t, err)
	st, err := client.WorkflowStatus.Create().
		SetName("Todo").SetCategory(workflowstatus.CategoryTodo).SetProjectID(proj.ID).Save(ctx)
	require.NoError(t, err)

	// high+task, medium+bug, low+task — OR(high|bug)=2, AND(high&bug)=0
	_, err = client.Issue.Create().
		SetIssueNumber(1).SetIssueKey("ORP-1").SetTitle("High task").
		SetIssueType(issue.IssueTypeTask).SetPriority(issue.PriorityHigh).
		SetProjectID(proj.ID).SetStatusID(st.ID).SetReporterID(u.ID).Save(ctx)
	require.NoError(t, err)
	_, err = client.Issue.Create().
		SetIssueNumber(2).SetIssueKey("ORP-2").SetTitle("Med bug").
		SetIssueType(issue.IssueTypeBug).SetPriority(issue.PriorityMedium).
		SetProjectID(proj.ID).SetStatusID(st.ID).SetReporterID(u.ID).Save(ctx)
	require.NoError(t, err)
	_, err = client.Issue.Create().
		SetIssueNumber(3).SetIssueKey("ORP-3").SetTitle("Low task").
		SetIssueType(issue.IssueTypeTask).SetPriority(issue.PriorityLow).
		SetProjectID(proj.ID).SetStatusID(st.ID).SetReporterID(u.ID).Save(ctx)
	require.NoError(t, err)

	qOR, err := search.ParseJQL(`priority = high OR type = bug`)
	require.NoError(t, err)
	predsOR := search.ToEntPredicates(qOR, u.ID)
	require.Len(t, predsOR, 1)
	orCount, err := client.Issue.Query().Where(predsOR[0]).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 2, orCount, "OR must union high OR bug")

	qAND, err := search.ParseJQL(`priority = high AND type = bug`)
	require.NoError(t, err)
	predsAND := search.ToEntPredicates(qAND, u.ID)
	require.Len(t, predsAND, 1)
	andCount, err := client.Issue.Query().Where(predsAND[0]).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, andCount, "AND must require both high and bug")
	assert.NotEqual(t, orCount, andCount, "OR vs AND must differ — Logic guard")
}
