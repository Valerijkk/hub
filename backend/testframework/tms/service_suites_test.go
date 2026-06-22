package tms_test

// Тесты сервиса TMS: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCreateSuite_InvalidParentID проверяет отклонение при создание suite некорректный parent id.
func TestCreateSuite_InvalidParentID(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	_, err := svc.CreateSuite(context.Background(), uuid.New(), tms.CreateSuiteRequest{
		Name:     "Folder",
		ParentID: "not-a-uuid",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parent_id")
}

// TestSuiteAndComponentCRUD проверяет сценарий: suite and component crud.
func TestSuiteAndComponentCRUD(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	suite, err := svc.CreateSuite(ctx, projectID, tms.CreateSuiteRequest{
		Name:        "Regression",
		Description: "regression cases",
	})
	require.NoError(t, err)
	assert.Equal(t, "Regression", suite.Name)

	tree, err := svc.ListSuites(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, tree, 1)

	suiteID, err := uuid.Parse(suite.ID)
	require.NoError(t, err)
	newName := "Regression v2"
	updated, err := svc.UpdateSuite(ctx, suiteID, tms.UpdateSuiteRequest{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "Regression v2", updated.Name)

	comp, err := svc.CreateComponent(ctx, projectID, tms.CreateComponentRequest{
		Name:        "Payments",
		Description: "payment module",
	})
	require.NoError(t, err)
	assert.Equal(t, "Payments", comp.Name)

	comps, err := svc.ListComponents(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, comps, 1)

	compID, err := uuid.Parse(comp.ID)
	require.NoError(t, err)
	require.NoError(t, svc.DeleteComponent(ctx, compID))
	require.NoError(t, svc.DeleteSuite(ctx, suiteID))

	comps, err = svc.ListComponents(ctx, projectID)
	require.NoError(t, err)
	assert.Empty(t, comps)
}

// TestGetTestCase_NotFound проверяет сценарий: получение test case not found.
func TestGetTestCase_NotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	_, err := svc.GetTestCase(context.Background(), "HUB-TC-999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

// TestListTestCases_InvalidSuiteID проверяет отклонение при список test cases некорректный suite id.
func TestListTestCases_InvalidSuiteID(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	_, err := svc.ListTestCases(context.Background(), uuid.New(), tms.ListTestCasesRequest{
		SuiteID: "bad-id",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "suite_id")
}

// TestUpdateSuite_NotFound проверяет сценарий: обновление suite not found.
func TestUpdateSuite_NotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	name := "x"
	_, err := svc.UpdateSuite(context.Background(), uuid.New(), tms.UpdateSuiteRequest{Name: &name})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}
