package tms

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateSuite_InvalidParentID(t *testing.T) {
	svc := NewTMSService(newMockRepo(), nil)
	_, err := svc.CreateSuite(context.Background(), uuid.New(), CreateSuiteRequest{
		Name:     "Folder",
		ParentID: "not-a-uuid",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "parent_id")
}

func TestSuiteAndComponentCRUD(t *testing.T) {
	repo := newMockRepo()
	svc := NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()

	suite, err := svc.CreateSuite(ctx, projectID, CreateSuiteRequest{
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
	updated, err := svc.UpdateSuite(ctx, suiteID, UpdateSuiteRequest{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "Regression v2", updated.Name)

	comp, err := svc.CreateComponent(ctx, projectID, CreateComponentRequest{
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

func TestGetTestCase_NotFound(t *testing.T) {
	svc := NewTMSService(newMockRepo(), nil)
	_, err := svc.GetTestCase(context.Background(), "HUB-TC-999")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

func TestListTestCases_InvalidSuiteID(t *testing.T) {
	svc := NewTMSService(newMockRepo(), nil)
	_, err := svc.ListTestCases(context.Background(), uuid.New(), ListTestCasesRequest{
		SuiteID: "bad-id",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "suite_id")
}

func TestUpdateSuite_NotFound(t *testing.T) {
	svc := NewTMSService(newMockRepo(), nil)
	name := "x"
	_, err := svc.UpdateSuite(context.Background(), uuid.New(), UpdateSuiteRequest{Name: &name})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}
