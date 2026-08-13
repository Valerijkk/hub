package tms_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestEntRepository_GetTestCaseVersion_ScopedByTestCase_A3_007 — HasTestCaseWith on Postgres.
// Mutation: remove testcaseversion.HasTestCaseWith(testcase.ID(testCaseID)) in
// GetTestCaseVersion → foreign TC lookup returns OK → FAIL.
func TestEntRepository_GetTestCaseVersion_ScopedByTestCase_A3_007(t *testing.T) {
	repo, ctx, projectID, userID := newTMSPostgresRepoEnv(t)

	tcA, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "TC-A", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)
	tcB, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "TC-B", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	versionB, err := repo.CreateTestCaseVersion(ctx, tcB.ID, userID, 1, map[string]interface{}{
		"title": "TC-B snapshot",
	}, "initial B")
	require.NoError(t, err)

	_, err = repo.GetTestCaseVersion(ctx, tcA.ID, versionB.ID)
	require.True(t, ent.IsNotFound(err), "version of TC-B must not load via TC-A, got %v", err)

	got, err := repo.GetTestCaseVersion(ctx, tcB.ID, versionB.ID)
	require.NoError(t, err)
	require.Equal(t, versionB.ID, got.ID)
}
