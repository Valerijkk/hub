package tms_test

// Тесты сервиса TMS (волна 113): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tms"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

type notFoundSuiteDeleteRepo struct {
	*mockRepository
}

func (r *notFoundSuiteDeleteRepo) DeleteSuite(_ context.Context, _ uuid.UUID) error {
	return &ent.NotFoundError{}
}

type notFoundComponentDeleteRepo struct {
	*mockRepository
}

func (r *notFoundComponentDeleteRepo) DeleteComponent(_ context.Context, _ uuid.UUID) error {
	return &ent.NotFoundError{}
}

// TestCreateTestCase_InvalidSuiteID проверяет отклонение при создание test case некорректный suite id.
func TestCreateTestCase_InvalidSuiteID(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	_, err := svc.CreateTestCase(context.Background(), uuid.New(), uuid.New(), tms.CreateTestCaseRequest{
		Title:   "TC",
		SuiteID: "not-a-uuid",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "suite_id")
}

// TestCreateTestCase_InvalidLabelID проверяет отклонение при создание test case некорректный label id.
func TestCreateTestCase_InvalidLabelID(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	_, err := svc.CreateTestCase(context.Background(), uuid.New(), uuid.New(), tms.CreateTestCaseRequest{
		Title:  "TC",
		Labels: []string{"bad-label"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "label")
}

// TestCreateTestCase_InvalidComponentID проверяет отклонение при создание test case некорректный component id.
func TestCreateTestCase_InvalidComponentID(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	_, err := svc.CreateTestCase(context.Background(), uuid.New(), uuid.New(), tms.CreateTestCaseRequest{
		Title:      "TC",
		Components: []string{"bad-component"},
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "component")
}

// TestUpdateTestCase_InvalidSuiteID проверяет отклонение при обновление test case некорректный suite id.
func TestUpdateTestCase_InvalidSuiteID(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	tc, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Update me", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	badSuite := "not-uuid"
	_, err = svc.UpdateTestCase(ctx, projectID, tc.CaseKey, userID, tms.UpdateTestCaseRequest{
		SuiteID: &badSuite,
		Version: tc.Version,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "suite_id")
}

// TestUpdateTestCase_ClearSuite проверяет сценарий: обновление test case clear suite.
func TestUpdateTestCase_ClearSuite(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	suite, err := svc.CreateSuite(ctx, projectID, tms.CreateSuiteRequest{Name: "Folder"})
	require.NoError(t, err)
	suiteID, err := uuid.Parse(suite.ID)
	require.NoError(t, err)

	tc, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title:   "With suite",
		SuiteID: suite.ID,
	})
	require.NoError(t, err)

	empty := ""
	updated, err := svc.UpdateTestCase(ctx, projectID, tc.CaseKey, userID, tms.UpdateTestCaseRequest{
		SuiteID: &empty,
		Version: tc.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, tc.CaseKey, updated.CaseKey)
	_ = suiteID
}

// TestUpdateTestCase_InvalidLabels проверяет отклонение при обновление test case некорректный метки.
func TestUpdateTestCase_InvalidLabels(t *testing.T) {
	repo := newMockRepo()
	svc := tms.NewTMSService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	tc, err := svc.CreateTestCase(ctx, projectID, userID, tms.CreateTestCaseRequest{
		Title: "Label update", CaseType: "classic", Priority: "medium",
	})
	require.NoError(t, err)

	labels := []string{"bad"}
	_, err = svc.UpdateTestCase(ctx, projectID, tc.CaseKey, userID, tms.UpdateTestCaseRequest{
		Labels:  &labels,
		Version: tc.Version,
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "label")
}

// TestDeleteSuite_NotFound проверяет сценарий: удаление suite not found.
func TestDeleteSuite_NotFound(t *testing.T) {
	base := newMockRepo()
	svc := tms.NewTMSService(&notFoundSuiteDeleteRepo{mockRepository: base}, nil)
	err := svc.DeleteSuite(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}

// TestDeleteComponent_NotFound проверяет сценарий: удаление component not found.
func TestDeleteComponent_NotFound(t *testing.T) {
	base := newMockRepo()
	svc := tms.NewTMSService(&notFoundComponentDeleteRepo{mockRepository: base}, nil)
	err := svc.DeleteComponent(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

// TestDeleteTestCase_NotFound проверяет сценарий: удаление test case not found.
func TestDeleteTestCase_NotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	err := svc.DeleteTestCase(context.Background(), uuid.New(), "HUB-TC-404")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

// TestLinkTestCaseToIssue_CaseNotFound проверяет сценарий: link test case to issue case not found.
func TestLinkTestCaseToIssue_CaseNotFound(t *testing.T) {
	svc := tms.NewTMSService(newMockRepo(), nil)
	err := svc.LinkTestCaseToIssue(context.Background(), uuid.New(), "HUB-TC-missing", uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}
