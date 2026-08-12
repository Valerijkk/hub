package tms_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	entissue "github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestLinkTestCaseToIssue_RejectsForeignIssue_A3_003 — LinkTestCaseToIssue не
// привязывает TC проекта A к issue проекта B.
// Mutation: снять HasProjectWith в FindIssueInProject → foreign link проходит → FAIL.
func TestLinkTestCaseToIssue_RejectsForeignIssue_A3_003(t *testing.T) {
	repo, ctx, projectAID, userID := newTMSPostgresRepoEnv(t)
	client := tms.ExportEntRepositoryClient(repo)
	svc := tms.NewTMSService(tms.NewCombinedRepository(repo, &tms.MongoRepository{}), nil)

	tc, err := svc.CreateTestCase(ctx, projectAID, userID, tms.CreateTestCaseRequest{
		Title: "TC in A", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	projB, err := client.Project.Create().
		SetName("TMS Link Project B").
		SetKey("TLB").
		Save(ctx)
	require.NoError(t, err)
	statusB, err := client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetIsDefault(true).
		SetProjectID(projB.ID).
		Save(ctx)
	require.NoError(t, err)

	foreignIssue, err := client.Issue.Create().
		SetIssueNumber(1).
		SetIssueKey("TLB-1").
		SetTitle("Foreign issue").
		SetIssueType(entissue.IssueTypeTask).
		SetPriority(entissue.PriorityMedium).
		SetProjectID(projB.ID).
		SetReporterID(userID).
		SetStatusID(statusB.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = repo.FindIssueInProject(ctx, projectAID, foreignIssue.ID)
	require.True(t, ent.IsNotFound(err), "foreign issue must not resolve in project A, got %v", err)

	err = svc.LinkTestCaseToIssue(ctx, projectAID, tc.CaseKey, foreignIssue.ID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 404, appErr.Code)

	links, err := svc.GetTestCaseLinks(ctx, projectAID, tc.CaseKey)
	require.NoError(t, err)
	assert.Empty(t, links, "cross-project link must not be created")

	statusA := mustDefaultStatusID(t, repo, ctx, projectAID)
	sameIssue, err := client.Issue.Create().
		SetIssueNumber(1).
		SetIssueKey("TMP-1").
		SetTitle("Same-project issue").
		SetIssueType(entissue.IssueTypeTask).
		SetPriority(entissue.PriorityMedium).
		SetProjectID(projectAID).
		SetReporterID(userID).
		SetStatusID(statusA).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.LinkTestCaseToIssue(ctx, projectAID, tc.CaseKey, sameIssue.ID))
	links, err = svc.GetTestCaseLinks(ctx, projectAID, tc.CaseKey)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, "TMP-1", links[0].IssueKey)

	require.NoError(t, svc.UnlinkTestCaseFromIssue(ctx, projectAID, tc.CaseKey, sameIssue.ID))
	err = svc.UnlinkTestCaseFromIssue(ctx, projectAID, tc.CaseKey, foreignIssue.ID)
	require.Error(t, err)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 404, appErr.Code)
}
