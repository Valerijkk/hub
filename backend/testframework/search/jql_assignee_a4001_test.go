package search_test

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

// TestJQL_AssigneeFiltersSpecificUser_A4_001 — assignee = UUID must match only
// that assignee (HasAssigneeWith). Mutation: both branches HasAssignee() → FAIL.
func TestJQL_AssigneeFiltersSpecificUser_A4_001(t *testing.T) {
	dsn := "file:jql_a4001_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	alice, err := client.User.Create().SetEmail("alice-" + uuid.New().String()[:8] + "@test.com").Save(ctx)
	require.NoError(t, err)
	bob, err := client.User.Create().SetEmail("bob-" + uuid.New().String()[:8] + "@test.com").Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().SetName("JQL Proj").SetKey("JQL").Save(ctx)
	require.NoError(t, err)
	st, err := client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.Issue.Create().
		SetIssueNumber(1).SetIssueKey("JQL-1").SetTitle("Alice task").
		SetIssueType(issue.IssueTypeTask).SetProjectID(proj.ID).SetStatusID(st.ID).
		SetAssigneeID(alice.ID).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.Issue.Create().
		SetIssueNumber(2).SetIssueKey("JQL-2").SetTitle("Bob task").
		SetIssueType(issue.IssueTypeTask).SetProjectID(proj.ID).SetStatusID(st.ID).
		SetAssigneeID(bob.ID).
		Save(ctx)
	require.NoError(t, err)

	q, err := search.ParseJQL(`assignee = ` + alice.ID.String())
	require.NoError(t, err)
	preds := search.ToEntPredicates(q, bob.ID)
	require.Len(t, preds, 1)

	got, err := client.Issue.Query().Where(preds[0]).All(ctx)
	require.NoError(t, err)
	require.Len(t, got, 1, "must return only Alice's issue")
	assert.Equal(t, "JQL-1", got[0].IssueKey)

	qCur, err := search.ParseJQL(`assignee = currentUser()`)
	require.NoError(t, err)
	predsCur := search.ToEntPredicates(qCur, alice.ID)
	require.Len(t, predsCur, 1)
	gotCur, err := client.Issue.Query().Where(predsCur[0]).All(ctx)
	require.NoError(t, err)
	require.Len(t, gotCur, 1)
	assert.Equal(t, "JQL-1", gotCur[0].IssueKey)
}
