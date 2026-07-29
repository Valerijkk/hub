//go:build hubtest

package tms_test

// TASK-139 — A3W2-2-001…006: AssignTesters membership, FileIDs scope,
// ImportAllure ExistingRun tx/compensating, mock UpdateTestResult fail-closed,
// snapshot steps, Delete/Link/Unlink AppError wrap.

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/filemetadata"
	"github.com/valerii/hub/backend/ent/testresult"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestAssignTestersToRun_RejectsNonMember_A3W2_2_001 — non-member UUID → 400.
// Mutation: убрать CountProjectMembers check в AssignTestersToRun → PASS с чужим UUID.
func TestAssignTestersToRun_RejectsNonMember_A3W2_2_001(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	member := uuid.New()
	outsider := uuid.New()
	repo.projectMembers[projectID] = map[uuid.UUID]bool{member: true}

	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "assign-scope"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	_, err = svc.AssignTestersToRun(ctx, projectID, runID, tms.AssignTestersRequest{
		UserIDs: []string{outsider.String()},
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)

	_, err = svc.AssignTestersToRun(ctx, projectID, runID, tms.AssignTestersRequest{
		UserIDs: []string{member.String()},
	})
	require.NoError(t, err)
}

// TestAssignTestersToRun_EntNonMember_A3W2_2_001 — Ent CountProjectMembers denial.
func TestAssignTestersToRun_EntNonMember_A3W2_2_001(t *testing.T) {
	entRepo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	client := tms.ExportEntRepositoryClient(entRepo)
	_, err := client.ProjectMember.Create().
		SetUserID(userID).
		SetProjectID(projectID).
		SetProjectRole("owner").
		Save(ctx)
	require.NoError(t, err)

	outsider, err := client.User.Create().
		SetEmail("outsider-" + uuid.New().String()[:8] + "@test.com").
		SetName("Outsider").
		Save(ctx)
	require.NoError(t, err)

	repo := tms.NewCombinedRepository(entRepo, &tms.MongoRepository{})
	svc := tms.NewTMSService(repo, nil)
	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "ent-assign"})
	require.NoError(t, err)
	runID, err := uuid.Parse(run.ID)
	require.NoError(t, err)

	_, err = svc.AssignTestersToRun(ctx, projectID, runID, tms.AssignTestersRequest{
		UserIDs: []string{outsider.ID.String()},
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)

	_, err = svc.AssignTestersToRun(ctx, projectID, runID, tms.AssignTestersRequest{
		UserIDs: []string{userID.String()},
	})
	require.NoError(t, err)
}

// TestExecuteTestResult_RejectsForeignFile_A3W2_2_002 — чужой FileID → 404, M2M не создаётся.
// Mutation: убрать CountFilesInProject → foreign file attach OK.
func TestExecuteTestResult_RejectsForeignFile_A3W2_2_002(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	foreignProject := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, uuid.New(), tms.CreateTestRunRequest{Title: "files"})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)
	_, err = svc.UpdateRunStatus(ctx, projectID, runID, "in_progress")
	require.NoError(t, err)

	resultID := uuid.New()
	repo.results[resultID] = &ent.TestResult{
		ID: resultID, Status: testresult.StatusUnknown,
		Edges: ent.TestResultEdges{Run: &ent.TestRun{ID: runID}},
	}
	repo.runResults[runID] = []uuid.UUID{resultID}

	foreignFile := uuid.New()
	repo.files[foreignFile] = &ent.FileMetadata{
		ID: foreignFile,
		Edges: ent.FileMetadataEdges{Project: &ent.Project{ID: foreignProject}},
	}
	okFile := uuid.New()
	repo.files[okFile] = &ent.FileMetadata{
		ID: okFile,
		Edges: ent.FileMetadataEdges{Project: &ent.Project{ID: projectID}},
	}

	passed := "passed"
	_, err = svc.ExecuteTestResult(ctx, projectID, runID, resultID, uuid.New(), tms.UpdateTestResultRequest{
		Status:  &passed,
		FileIDs: []string{foreignFile.String()},
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 404, appErr.Code)

	_, err = svc.ExecuteTestResult(ctx, projectID, runID, resultID, uuid.New(), tms.UpdateTestResultRequest{
		Status:  &passed,
		FileIDs: []string{okFile.String()},
	})
	require.NoError(t, err)
}

// TestExecuteTestResult_EntForeignFile_A3W2_2_002 — Ent CountFilesInProject denial.
func TestExecuteTestResult_EntForeignFile_A3W2_2_002(t *testing.T) {
	entRepo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	client := tms.ExportEntRepositoryClient(entRepo)

	foreign, err := client.Project.Create().SetName("Foreign").SetKey("FRX").Save(ctx)
	require.NoError(t, err)

	foreignFile, err := client.FileMetadata.Create().
		SetFilename("x.png").
		SetContentType("image/png").
		SetSize(1).
		SetS3Key("k").
		SetS3Bucket("b").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(userID).
		SetProjectID(foreign.ID).
		Save(ctx)
	require.NoError(t, err)

	okFile, err := client.FileMetadata.Create().
		SetFilename("ok.png").
		SetContentType("image/png").
		SetSize(1).
		SetS3Key("k2").
		SetS3Bucket("b").
		SetStatus(filemetadata.StatusConfirmed).
		SetUploaderID(userID).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	n, err := entRepo.CountFilesInProject(ctx, projectID, []uuid.UUID{foreignFile.ID})
	require.NoError(t, err)
	assert.Equal(t, 0, n)

	n, err = entRepo.CountFilesInProject(ctx, projectID, []uuid.UUID{okFile.ID})
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	repo := tms.NewCombinedRepository(entRepo, &tms.MongoRepository{})
	svc := tms.NewTMSService(repo, nil)
	tc, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "f", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)
	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "r"})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)
	detail, err := svc.AddTestCasesToRun(ctx, projectID, runID, tms.AddTestCasesToRunRequest{CaseIDs: []string{tc.ID}})
	require.NoError(t, err)
	require.Len(t, detail.Results, 1)
	resultID, _ := uuid.Parse(detail.Results[0].ID)
	_, err = svc.StartTestRun(ctx, projectID, runID)
	require.NoError(t, err)

	passed := "passed"
	_, err = svc.ExecuteTestResult(ctx, projectID, runID, resultID, userID, tms.UpdateTestResultRequest{
		Status:  &passed,
		FileIDs: []string{foreignFile.ID.String()},
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 404, appErr.Code)

	_, err = svc.ExecuteTestResult(ctx, projectID, runID, resultID, userID, tms.UpdateTestResultRequest{
		Status:  &passed,
		FileIDs: []string{okFile.ID.String()},
	})
	require.NoError(t, err)
}

// TestImportAllureToExistingRun_CompensatingCleanup_A3W2_2_003 —
// fail recompute → partial results удалены.
// Mutation: убрать DeleteTestResults после fail → остаются orphan results.
func TestImportAllureToExistingRun_CompensatingCleanup_A3W2_2_003(t *testing.T) {
	repo := newMockRepo()
	repo.failUpdateStats = true
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	run, err := svc.CreateTestRun(ctx, projectID, userID, tms.CreateTestRunRequest{Title: "import-existing"})
	require.NoError(t, err)
	runID, _ := uuid.Parse(run.ID)
	before := len(repo.results)

	payload, err := json.Marshal([]tms.AllureResult{{
		UUID: "u1", Name: "t", Status: "passed",
	}})
	require.NoError(t, err)

	_, err = svc.ImportAllureToExistingRun(ctx, projectID, runID, userID, tms.ImportAllureRequest{Results: payload})
	require.Error(t, err)
	assert.Equal(t, before, len(repo.results), "partial results must be compensated")
	assert.Empty(t, repo.runResults[runID])
}

// TestImportAllureToExistingRun_EntTx_A3W2_2_003 — Ent ImportResultsToExistingRun happy path.
// Без Mongo CombinedRepository: StoreAllureResults на nil-DB паникует (вне scope A3W2-2-003).
func TestImportAllureToExistingRun_EntTx_A3W2_2_003(t *testing.T) {
	entRepo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	run, err := entRepo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "existing", CreatedByID: userID,
	})
	require.NoError(t, err)

	err = entRepo.ImportResultsToExistingRun(ctx, run.ID, []tms.CreateTestResultInput{
		{Status: "passed", AllureUUID: "e1"},
	})
	require.NoError(t, err)

	got, err := entRepo.GetTestRun(ctx, run.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, got.TotalCount, 1)
	assert.GreaterOrEqual(t, got.PassedCount, 1)

	results, err := entRepo.ListTestResults(ctx, run.ID)
	require.NoError(t, err)
	assert.Len(t, results, 1)
}

// TestMockUpdateTestResult_RequiresRunScope_A3W2_2_004 —
// без Edges.Run и без runResults → NotFound.
// Mutation: ослабить mock (skip без Run) → тест краснеет.
func TestMockUpdateTestResult_RequiresRunScope_A3W2_2_004(t *testing.T) {
	repo := newMockRepo()
	resultID := uuid.New()
	runID := uuid.New()
	repo.results[resultID] = &ent.TestResult{ID: resultID, Status: testresult.StatusUnknown}

	status := "passed"
	_, err := repo.UpdateTestResult(context.Background(), runID, resultID, tms.UpdateTestResultInput{Status: &status})
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	repo.results[resultID].Edges.Run = &ent.TestRun{ID: runID}
	updated, err := repo.UpdateTestResult(context.Background(), runID, resultID, tms.UpdateTestResultInput{Status: &status})
	require.NoError(t, err)
	assert.Equal(t, testresult.StatusPassed, updated.Status)
}

// TestSnapshotToUpdateInput_CorruptSteps_A3W2_2_005 — битый steps → error.
// Mutation: вернуть `_ = json.Marshal` / ignore Unmarshal → silent empty steps.
func TestSnapshotToUpdateInput_CorruptSteps_A3W2_2_005(t *testing.T) {
	_, err := tms.ExportSnapshotToUpdateInput(map[string]interface{}{
		"title": "t",
		"steps": "not-an-array",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "steps")

	input, err := tms.ExportSnapshotToUpdateInput(map[string]interface{}{
		"title": "ok",
		"steps": []map[string]interface{}{{"order": 1, "action": "click"}},
	})
	require.NoError(t, err)
	require.NotNil(t, input.Steps)
	assert.Len(t, *input.Steps, 1)
}

// TestRollbackTestCase_CorruptSnapshotSteps_A3W2_2_005 — rollback path fail-closed.
func TestRollbackTestCase_CorruptSnapshotSteps_A3W2_2_005(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "rb", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)
	tcID, _ := uuid.Parse(created.ID)

	versionID := uuid.New()
	repo.versions[versionID] = &ent.TestCaseVersion{
		ID:       versionID,
		Snapshot: map[string]interface{}{"title": "old", "steps": 42},
		Edges:    ent.TestCaseVersionEdges{TestCase: &ent.TestCase{ID: tcID}},
	}

	_, err = svc.RollbackTestCase(ctx, projectID, created.CaseKey, versionID, userID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 500, appErr.Code)
	assert.Contains(t, appErr.Error(), "снимка")
}

// TestDeleteLinkUnlink_WrapRawErr_A3W2_2_006 — raw repo err → AppError.
// Mutation: вернуть raw `return err` → тест ловит сырой text / non-AppError.
func TestDeleteLinkUnlink_WrapRawErr_A3W2_2_006(t *testing.T) {
	raw := errors.New("ent: constraint UNIQUE_xyz leaked")
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "wrap", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	repo.deleteCaseErr = raw
	err = svc.DeleteTestCase(ctx, projectID, created.CaseKey)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 500, appErr.Code)
	assert.NotContains(t, appErr.Error(), "UNIQUE_xyz")

	issueID := uuid.New()
	repo.issues[issueID] = &ent.Issue{
		ID: issueID,
		Edges: ent.IssueEdges{Project: &ent.Project{ID: projectID}},
	}
	repo.deleteCaseErr = nil
	repo.linkIssueErr = raw
	err = svc.LinkTestCaseToIssue(ctx, projectID, created.CaseKey, issueID)
	require.Error(t, err)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 500, appErr.Code)
	assert.NotContains(t, appErr.Error(), "UNIQUE_xyz")

	repo.linkIssueErr = nil
	repo.unlinkIssueErr = raw
	err = svc.UnlinkTestCaseFromIssue(ctx, projectID, created.CaseKey, issueID)
	require.Error(t, err)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 500, appErr.Code)
	assert.NotContains(t, appErr.Error(), "UNIQUE_xyz")

	// AppError пробрасывается без переобёртки.
	repo.unlinkIssueErr = apperrors.NewConflict("уже отвязан")
	err = svc.UnlinkTestCaseFromIssue(ctx, projectID, created.CaseKey, issueID)
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 409, appErr.Code)
}
