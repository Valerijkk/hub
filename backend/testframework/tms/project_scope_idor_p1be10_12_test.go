//go:build hubtest

package tms_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestDeleteSuite_RejectsForeignProject_P1BE10 — suite UUID проекта A не удаляется
// под projectID B (mutation: DeleteOneID без HasProjectWith → FAIL).
func TestDeleteSuite_RejectsForeignProject_P1BE10(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	suite, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "Owned"})
	require.NoError(t, err)

	foreign, err := tms.ExportEntRepositoryClient(repo).Project.Create().
		SetName("Foreign Suite Proj").
		SetKey("FSP").
		Save(ctx)
	require.NoError(t, err)

	err = repo.DeleteSuite(ctx, foreign.ID, suite.ID)
	require.True(t, ent.IsNotFound(err), "cross-project suite delete must 404, got %v", err)

	still, err := repo.GetSuite(ctx, suite.ID)
	require.NoError(t, err)
	require.Equal(t, suite.ID, still.ID)

	require.NoError(t, repo.DeleteSuite(ctx, projectID, suite.ID))
}

// TestUpdateSuite_RejectsForeignProject_P1BE10 — переименование чужой папки → NotFound.
func TestUpdateSuite_RejectsForeignProject_P1BE10(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	suite, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "Owned"})
	require.NoError(t, err)

	foreign, err := tms.ExportEntRepositoryClient(repo).Project.Create().
		SetName("Foreign Suite Upd").
		SetKey("FSU").
		Save(ctx)
	require.NoError(t, err)

	name := "Hacked"
	_, err = repo.UpdateSuite(ctx, foreign.ID, suite.ID, tms.UpdateSuiteInput{Name: &name, Version: suite.Version})
	require.True(t, ent.IsNotFound(err), "cross-project suite update must 404, got %v", err)

	got, err := repo.GetSuite(ctx, suite.ID)
	require.NoError(t, err)
	require.Equal(t, "Owned", got.Name)
}

// TestDeleteComponent_RejectsForeignProject_P1BE10 — component чужого проекта → NotFound.
func TestDeleteComponent_RejectsForeignProject_P1BE10(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	comp, err := repo.CreateComponent(ctx, projectID, "UI", "frontend")
	require.NoError(t, err)

	foreign, err := tms.ExportEntRepositoryClient(repo).Project.Create().
		SetName("Foreign Comp Proj").
		SetKey("FCP").
		Save(ctx)
	require.NoError(t, err)

	err = repo.DeleteComponent(ctx, foreign.ID, comp.ID)
	require.True(t, ent.IsNotFound(err), "cross-project component delete must 404, got %v", err)

	comps, err := repo.ListComponents(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, comps, 1)
}

// TestDeleteEnvironment_RejectsForeignProject_P1BE11 — env чужого проекта → NotFound.
func TestDeleteEnvironment_RejectsForeignProject_P1BE11(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	env, err := repo.CreateEnvironment(ctx, projectID, tms.CreateEnvironmentInput{
		Name: "Staging", Label: "staging",
	})
	require.NoError(t, err)

	foreign, err := tms.ExportEntRepositoryClient(repo).Project.Create().
		SetName("Foreign Env Proj").
		SetKey("FEP").
		Save(ctx)
	require.NoError(t, err)

	err = repo.DeleteEnvironment(ctx, foreign.ID, env.ID)
	require.True(t, ent.IsNotFound(err), "cross-project env delete must 404, got %v", err)

	still, err := repo.GetEnvironmentInProject(ctx, projectID, env.ID)
	require.NoError(t, err)
	require.Equal(t, env.ID, still.ID)
}

// TestGetEnvironmentInProject_RejectsForeign_P1BE11 — CompareEnvironments path: чужой env → NotFound.
func TestGetEnvironmentInProject_RejectsForeign_P1BE11(t *testing.T) {
	repo, ctx, projectID, _ := newEntRepositoryTestEnv(t)

	env, err := repo.CreateEnvironment(ctx, projectID, tms.CreateEnvironmentInput{
		Name: "Dev", Label: "dev",
	})
	require.NoError(t, err)

	foreign, err := tms.ExportEntRepositoryClient(repo).Project.Create().
		SetName("Foreign Compare Proj").
		SetKey("FCM").
		Save(ctx)
	require.NoError(t, err)

	_, err = repo.GetEnvironmentInProject(ctx, foreign.ID, env.ID)
	require.True(t, ent.IsNotFound(err), "cross-project env get must 404, got %v", err)
}

// TestCreateTestCase_RejectsForeignSuite_P1BE12 — suite_id чужого проекта → 404 на service.
func TestCreateTestCase_RejectsForeignSuite_P1BE12(t *testing.T) {
	entRepo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	repo := tms.NewCombinedRepository(entRepo, &tms.MongoRepository{})
	svc := tms.NewTMSService(repo, nil)

	foreignProj, err := tms.ExportEntRepositoryClient(entRepo).Project.Create().
		SetName("Foreign TC Suite").
		SetKey("FTS").
		Save(ctx)
	require.NoError(t, err)

	foreignSuite, err := repo.CreateSuite(ctx, foreignProj.ID, tms.CreateSuiteInput{Name: "Alien"})
	require.NoError(t, err)

	_, err = svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title:   "Cross suite",
		SuiteID: foreignSuite.ID.String(),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "не найдена")
}

// TestCreateSuite_RejectsForeignParent_P1BE12 — ParentID чужого проекта → 404.
func TestCreateSuite_RejectsForeignParent_P1BE12(t *testing.T) {
	entRepo, ctx, projectID, _ := newEntRepositoryTestEnv(t)
	repo := tms.NewCombinedRepository(entRepo, &tms.MongoRepository{})
	svc := tms.NewTMSService(repo, nil)

	foreignProj, err := tms.ExportEntRepositoryClient(entRepo).Project.Create().
		SetName("Foreign Parent Proj").
		SetKey("FPP").
		Save(ctx)
	require.NoError(t, err)

	foreignParent, err := repo.CreateSuite(ctx, foreignProj.ID, tms.CreateSuiteInput{Name: "Alien Parent"})
	require.NoError(t, err)

	_, err = svc.CreateSuite(ctx, projectID, tms.CreateSuiteRequest{
		Name:     "Child",
		ParentID: foreignParent.ID.String(),
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "не найдена")
}

// TestAddTestCasesToRun_RejectsForeignCase_P1BE12 — CaseID чужого проекта → 404.
func TestAddTestCasesToRun_RejectsForeignCase_P1BE12(t *testing.T) {
	entRepo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	repo := tms.NewCombinedRepository(entRepo, &tms.MongoRepository{})
	svc := tms.NewTMSService(repo, nil)

	run, err := repo.CreateTestRun(ctx, projectID, tms.CreateTestRunInput{
		Title: "Local run", CreatedByID: userID,
	})
	require.NoError(t, err)

	foreignProj, err := tms.ExportEntRepositoryClient(entRepo).Project.Create().
		SetName("Foreign Case Proj").
		SetKey("FCA").
		Save(ctx)
	require.NoError(t, err)

	foreignTC, err := repo.CreateTestCase(ctx, foreignProj.ID, tms.CreateTestCaseInput{
		Title: "Alien", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	_, err = svc.AddTestCasesToRun(ctx, projectID, run.ID, tms.AddTestCasesToRunRequest{
		CaseIDs: []string{foreignTC.ID.String()},
	})
	require.Error(t, err)
	require.Contains(t, err.Error(), "не найден")
}
