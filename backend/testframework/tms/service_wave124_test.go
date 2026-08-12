package tms_test

// Тесты сервиса TMS (волна 124): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	entissue "github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/schema"
)

// Фаза 124 — ent-backed service paths for version rollback and TestRail import..

func wave124Repo(t *testing.T) (tms.Repository, context.Context, uuid.UUID, uuid.UUID) {
	t.Helper()
	entRepo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	return tms.NewCombinedRepository(entRepo, &tms.MongoRepository{}), ctx, projectID, userID
}

// TestService_Wave124_UpdateTestCase_WithComponents_Ent проверяет слой репозитория: сервис wave 124 обновление test case with components ent.
func TestService_Wave124_UpdateTestCase_WithComponents_Ent(t *testing.T) {
	repo, ctx, projectID, userID := wave124Repo(t)
	svc := tms.NewTMSService(repo, nil)

	comp, err := repo.CreateComponent(ctx, projectID, "API", "")
	require.NoError(t, err)

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Case with comp", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	compIDs := []string{comp.ID.String()}
	title := "Updated comp case"
	updated, err := svc.UpdateTestCase(ctx, projectID, created.CaseKey, userID, tms.UpdateTestCaseRequest{
		Title:      &title,
		Components: &compIDs,
		Version:    1,
	})
	require.NoError(t, err)
	assert.Equal(t, "Updated comp case", updated.Title)
}

// TestService_Wave124_RollbackTestCase_Success_Ent проверяет слой репозитория: сервис wave 124 rollback test case успех ent.
func TestService_Wave124_RollbackTestCase_Success_Ent(t *testing.T) {
	repo, ctx, projectID, userID := wave124Repo(t)
	svc := tms.NewTMSService(repo, nil)

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Rollback me", CaseType: "classic", Priority: "low",
	})
	require.NoError(t, err)

	v1Title := "Rollback me"
	_, err = svc.UpdateTestCase(ctx, projectID, created.CaseKey, userID, tms.UpdateTestCaseRequest{Title: &v1Title, Version: 1})
	require.NoError(t, err)

	versions, err := svc.ListTestCaseVersions(ctx, projectID, created.CaseKey)
	require.NoError(t, err)
	require.NotEmpty(t, versions)

	versionID, err := uuid.Parse(versions[0].ID)
	require.NoError(t, err)

	rolled, err := svc.RollbackTestCase(ctx, projectID, created.CaseKey, versionID, userID)
	require.NoError(t, err)
	assert.Equal(t, "Rollback me", rolled.Title)
}

// TestService_Wave124_GetTestCaseLinks_Ent проверяет слой репозитория: сервис wave 124 получение test case links ent.
func TestService_Wave124_GetTestCaseLinks_Ent(t *testing.T) {
	repo, ctx, projectID, userID := wave124Repo(t)
	svc := tms.NewTMSService(repo, nil)

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Linked", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	links, err := svc.GetTestCaseLinks(ctx, projectID, created.CaseKey)
	require.NoError(t, err)
	assert.Empty(t, links)
}

// TestService_Wave124_LinkTestCaseToIssue_Ent проверяет слой репозитория: сервис wave 124 link test case to issue ent.
func TestService_Wave124_LinkTestCaseToIssue_Ent(t *testing.T) {
	entRepo, ctx, projectID, userID := newEntRepositoryTestEnv(t)
	repo := tms.NewCombinedRepository(entRepo, &tms.MongoRepository{})
	svc := tms.NewTMSService(repo, nil)

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Link target", CaseType: "classic", Priority: "high",
	})
	require.NoError(t, err)

	statusID := mustDefaultStatusID(t, entRepo, ctx, projectID)
	issue, err := tms.ExportEntRepositoryClient(entRepo).Issue.Create().
		SetIssueNumber(42).
		SetIssueKey("TMS-42").
		SetTitle("Issue for link").
		SetIssueType(entissue.IssueTypeTask).
		SetPriority(entissue.PriorityMedium).
		SetProjectID(projectID).
		SetReporterID(userID).
		SetStatusID(statusID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, svc.LinkTestCaseToIssue(ctx, projectID, created.CaseKey, issue.ID))
	links, err := svc.GetTestCaseLinks(ctx, projectID, created.CaseKey)
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, "TMS-42", links[0].IssueKey)
}

// TestService_Wave124_UpdateTestCase_AllFields_Ent проверяет слой репозитория: сервис wave 124 обновление test case all fields ent.
func TestService_Wave124_UpdateTestCase_AllFields_Ent(t *testing.T) {
	repo, ctx, projectID, userID := wave124Repo(t)
	svc := tms.NewTMSService(repo, nil)

	created, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Full update", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	title := "Full update v2"
	desc := "Description"
	pre := "Pre"
	post := "Post"
	status := "active"
	steps := []schema.TestStep{{Order: 1, Action: "click", ExpectedResult: "ok"}}
	updated, err := svc.UpdateTestCase(ctx, projectID, created.CaseKey, userID, tms.UpdateTestCaseRequest{
		Title:         &title,
		Description:   &desc,
		Precondition:  &pre,
		Postcondition: &post,
		Status:        &status,
		Steps:         &steps,
		Version:       1,
	})
	require.NoError(t, err)
	assert.Equal(t, "Full update v2", updated.Title)
	assert.Equal(t, "active", updated.Status)
}

// TestService_Wave124_CreateComponent_Ent проверяет слой репозитория: сервис wave 124 создание component ent.
func TestService_Wave124_CreateComponent_Ent(t *testing.T) {
	repo, ctx, projectID, _ := wave124Repo(t)
	svc := tms.NewTMSService(repo, nil)

	comp, err := svc.CreateComponent(ctx, projectID, tms.CreateComponentRequest{Name: "UI"})
	require.NoError(t, err)
	assert.Equal(t, "UI", comp.Name)

	comps, err := svc.ListComponents(ctx, projectID)
	require.NoError(t, err)
	assert.NotEmpty(t, comps)
}
