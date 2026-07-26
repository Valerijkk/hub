package tms_test

// Тесты репозитория TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/schema"
)

// TestEntRepository_ListTestCases_Filters проверяет слой репозитория: ent репозиторий список test cases filters.
func TestEntRepository_ListTestCases_Filters(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	suite, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "Regression"})
	require.NoError(t, err)

	_, err = repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Alpha case", CaseType: "classic", Priority: "high", Status: "draft",
		CreatedByID: userID, SuiteID: &suite.ID,
	})
	require.NoError(t, err)
	_, err = repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Beta login", CaseType: "classic", Priority: "low", Status: "active",
		CreatedByID: userID,
	})
	require.NoError(t, err)

	items, total, err := repo.ListTestCases(ctx, projectID, tms.ListTestCasesParams{
		Status: "draft", Priority: "high", SuiteID: &suite.ID,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, "Alpha case", items[0].Title)

	items, total, err = repo.ListTestCases(ctx, projectID, tms.ListTestCasesParams{Search: "login"})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
}

// TestEntRepository_UpdateTestCase_AllFields проверяет слой репозитория: ent репозиторий обновление test case all fields.
func TestEntRepository_UpdateTestCase_AllFields(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	suite, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "Suite A"})
	require.NoError(t, err)

	label, err := tms.ExportEntRepositoryClient(repo).Label.Create().
		SetName("smoke").
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	comp, err := repo.CreateComponent(ctx, projectID, "Auth", "auth module")
	require.NoError(t, err)

	created, err := repo.CreateTestCase(ctx, projectID, tms.CreateTestCaseInput{
		Title: "Original", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: userID,
	})
	require.NoError(t, err)

	newTitle := "Updated title"
	newDesc := "Steps here"
	newPre := "Pre"
	newPost := "Post"
	newType := "bdd"
	newPriority := "critical"
	newStatus := "active"
	steps := []schema.TestStep{{Order: 1, Action: "click", ExpectedResult: "ok"}}
	fullName := "com.example.TestLogin"
	className := "com.example.TestLogin"
	methodName := "testLogin"
	extID := "EXT-1"
	autoID := "auto-1"

	updated, err := repo.UpdateTestCase(ctx, created.ID, tms.UpdateTestCaseInput{
		Title:         &newTitle,
		Description:   &newDesc,
		Precondition:  &newPre,
		Postcondition: &newPost,
		CaseType:      &newType,
		Priority:      &newPriority,
		Status:        &newStatus,
		Steps:         &steps,
		FullName:      &fullName,
		ClassName:     &className,
		MethodName:    &methodName,
		ExternalID:    &extID,
		AutomationID:  &autoID,
		SuiteID:       &suite.ID,
		LabelIDs:      &[]uuid.UUID{label.ID},
		ComponentIDs:  &[]uuid.UUID{comp.ID},
		Version:       1,
	})
	require.NoError(t, err)
	assert.Equal(t, newTitle, updated.Title)
	assert.Equal(t, newPriority, string(updated.Priority))
	assert.Equal(t, 2, updated.Version)
	require.NotNil(t, updated.Edges.Suite)
	assert.Equal(t, suite.ID, updated.Edges.Suite.ID)

	cleared, err := repo.UpdateTestCase(ctx, created.ID, tms.UpdateTestCaseInput{ClearSuite: true, Version: 2})
	require.NoError(t, err)
	assert.Nil(t, cleared.Edges.Suite)
}
