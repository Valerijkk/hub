//go:build hubtest

package tms_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	entissue "github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestEntRepository_DeleteTestCase_IdempotentNotFound_P3BE05 — повторный soft-delete → NotFound.
// Mutation: UpdateOneID без DeletedAtIsNil → второй delete «успешен» → FAIL.
func TestEntRepository_DeleteTestCase_IdempotentNotFound_P3BE05(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Once", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)
	require.NoError(t, repo.DeleteTestCase(ctx, tc.ID))
	err = repo.DeleteTestCase(ctx, tc.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

// TestEntRepository_MoveSuite_RequiresProject_P3BE05 — MoveSuite без project scope.
// Mutation: UpdateOneID без HasProjectWith → foreign project move succeeds → FAIL.
func TestEntRepository_MoveSuite_RequiresProject_P3BE05(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)
	parent, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "Parent"})
	require.NoError(t, err)
	child, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{
		Name: "Child", ParentID: &parent.ID,
	})
	require.NoError(t, err)

	foreign, err := tms.ExportEntRepositoryClient(repo).Project.Create().
		SetName("Foreign TMS").
		SetKey("FTM").
		Save(ctx)
	require.NoError(t, err)

	_, err = repo.MoveSuite(ctx, foreign.ID, child.ID, nil)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	moved, err := repo.MoveSuite(ctx, projectID, child.ID, nil)
	require.NoError(t, err)
	assert.Nil(t, moved.Edges.Parent)
}

// TestEntRepository_AddFilesToResult_RequiresRun_P3BE05 — AddFiles без HasRunWith.
func TestEntRepository_AddFilesToResult_RequiresRun_P3BE05(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Files", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)
	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Run", CreatedByID: userID,
	})
	require.NoError(t, err)
	result, err := repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
		TestCaseID: &tc.ID, Status: "passed",
	})
	require.NoError(t, err)

	fileID := uuid.New()
	_, err = tms.ExportEntRepositoryClient(repo).FileMetadata.Create().
		SetID(fileID).
		SetFilename("a.png").
		SetContentType("image/png").
		SetSize(1).
		SetS3Key("k").
		SetS3Bucket("b").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	err = repo.AddFilesToResult(ctx, uuid.New(), result.ID, []uuid.UUID{fileID})
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	require.NoError(t, repo.AddFilesToResult(ctx, run.ID, result.ID, []uuid.UUID{fileID}))
}

// TestEntRepository_GetProjectIssuesWithTestCases_SkipsSoftDeleted_P3BE05 —
// coverage включал задачи из корзины. Mutation: снять DeletedAtIsNil → FAIL.
func TestEntRepository_GetProjectIssuesWithTestCases_SkipsSoftDeleted_P3BE05(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	tc, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Linked", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	statusID := mustDefaultStatusID(t, repo, ctx, projectID)
	issue, err := tms.ExportEntRepositoryClient(repo).Issue.Create().
		SetIssueNumber(1).
		SetIssueKey("SOFT-1").
		SetTitle("Trash me").
		SetIssueType(entissue.IssueTypeTask).
		SetProjectID(projectID).
		SetReporterID(userID).
		SetStatusID(statusID).
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, repo.LinkTestCaseToIssue(ctx, tc.ID, issue.ID))

	now := time.Now().UTC()
	_, err = tms.ExportEntRepositoryClient(repo).Issue.UpdateOneID(issue.ID).SetDeletedAt(now).Save(ctx)
	require.NoError(t, err)

	issues, err := repo.GetProjectIssuesWithTestCases(ctx, projectID)
	require.NoError(t, err)
	assert.Empty(t, issues)
}
