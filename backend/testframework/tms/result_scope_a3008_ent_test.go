package tms_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestEntRepository_UpdateTestResult_RejectsWrongRun_A3_008 — HasRunWith(runID) на Postgres.
// Mutation: снять testresult.HasRunWith(testrun.ID(runID)) в UpdateTestResult →
// обновление result чужого run проходит → FAIL.
func TestEntRepository_UpdateTestResult_RejectsWrongRun_A3_008(t *testing.T) {
	repo, ctx, projectID, userID := newTMSPostgresRepoEnv(t)

	runA, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Run A", RunType: "manual", CreatedByID: userID,
	})
	require.NoError(t, err)
	runB, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Run B", RunType: "manual", CreatedByID: userID,
	})
	require.NoError(t, err)

	resultB, err := repo.CreateTestResult(ctx, runB.ID, tms.CreateTestResultInput{
		Status: "unknown", Comment: "owned by B",
	})
	require.NoError(t, err)

	passed := "passed"
	_, err = repo.UpdateTestResult(ctx, runA.ID, resultB.ID, tms.UpdateTestResultInput{Status: &passed})
	require.True(t, ent.IsNotFound(err), "result of run B must not update via run A, got %v", err)

	got, err := repo.GetTestResult(ctx, resultB.ID)
	require.NoError(t, err)
	assert.Equal(t, "unknown", string(got.Status), "wrong-run update must not change status")

	updated, err := repo.UpdateTestResult(ctx, runB.ID, resultB.ID, tms.UpdateTestResultInput{Status: &passed})
	require.NoError(t, err)
	assert.Equal(t, "passed", string(updated.Status))
}
