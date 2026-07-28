//go:build hubtest

package tms_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/tms"
)

// TASK-037 / P2-BE-02 — Ent-negative для TMS CreateBug / suite update (не mock handler_flow).
// Mutation: снять GetTestRunInProject / HasProjectWith → cross-project 404 краснеет.

func TestCreateBugFromFailedResult_RejectsForeignProject_Ent_TASK037(t *testing.T) {
	entRepo, ctx, projectA, userID := newEntRepositoryTestEnv(t)
	repo := tms.NewCombinedRepository(entRepo, &tms.MongoRepository{})
	svc := tms.NewTMSService(repo, nil)
	client := tms.ExportEntRepositoryClient(entRepo)

	run, err := repo.CreateTestRun(ctx, projectA, tms.CreateTestRunInput{
		Title: "Owned run", CreatedByID: userID,
	})
	require.NoError(t, err)

	tc, err := repo.CreateTestCase(ctx, projectA, tms.CreateTestCaseInput{
		Title: "Failing", CaseType: "classic", Priority: "medium", Status: "active", CreatedByID: userID,
	})
	require.NoError(t, err)

	result, err := repo.CreateTestResult(ctx, run.ID, tms.CreateTestResultInput{
		TestCaseID: &tc.ID, Status: "failed", Comment: "boom",
	})
	require.NoError(t, err)

	foreign, err := client.Project.Create().
		SetName("Foreign Bug Proj").
		SetKey("FBP").
		Save(ctx)
	require.NoError(t, err)

	_, err = svc.CreateBugFromFailedResult(ctx, foreign.ID, userID, result.ID, tms.CreateBugRequest{
		Title: "IDOR bug",
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "не найден")

	bug, err := svc.CreateBugFromFailedResult(ctx, projectA, userID, result.ID, tms.CreateBugRequest{
		Title: "Real bug",
	})
	require.NoError(t, err)
	require.NotEmpty(t, bug.ID)
}

func TestUpdateSuite_ServiceRejectsForeignProject_Ent_TASK037(t *testing.T) {
	entRepo, ctx, projectID, _ := newEntRepositoryTestEnv(t)
	repo := tms.NewCombinedRepository(entRepo, &tms.MongoRepository{})
	svc := tms.NewTMSService(repo, nil)

	suite, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "Owned"})
	require.NoError(t, err)

	foreign, err := tms.ExportEntRepositoryClient(entRepo).Project.Create().
		SetName("Foreign Suite Svc").
		SetKey("FSS").
		Save(ctx)
	require.NoError(t, err)

	name := "Hacked via service"
	_, err = svc.UpdateSuite(ctx, foreign.ID, suite.ID, tms.UpdateSuiteRequest{Name: &name})
	require.Error(t, err)
	require.True(t, strings.Contains(err.Error(), "не найден"),
		"cross-project suite update must 404, got %v", err)

	got, err := repo.GetSuite(ctx, suite.ID)
	require.NoError(t, err)
	require.Equal(t, "Owned", got.Name)
}
