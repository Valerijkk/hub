package tracker_test

// Тесты сервиса трекера задач (волна 114): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

type notFoundWorkflowStatusDeleteRepo struct {
	mockRepo
}

func (r *notFoundWorkflowStatusDeleteRepo) DeleteWorkflowStatus(_ context.Context, _, _ uuid.UUID) error {
	return &ent.NotFoundError{}
}

type notFoundWorkflowTransitionDeleteRepo struct {
	mockRepo
}

func (r *notFoundWorkflowTransitionDeleteRepo) DeleteWorkflowTransition(_ context.Context, _, _ uuid.UUID) error {
	return &ent.NotFoundError{}
}

// TestUpdateWorkflowStatus_NotFound проверяет сценарий: обновление workflow status not found.
func TestUpdateWorkflowStatus_NotFound(t *testing.T) {
	svc, _ := newTestService()
	name := "Updated"
	_, err := svc.UpdateWorkflowStatus(context.Background(), uuid.New(), uuid.New(), tracker.CreateWorkflowStatusRequest{
		Name: name, Category: "todo", Color: "#000000",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

// TestDeleteWorkflowStatus_NotFound проверяет сценарий: удаление workflow status not found.
func TestDeleteWorkflowStatus_NotFound(t *testing.T) {
	repo := &notFoundWorkflowStatusDeleteRepo{mockRepo: *newMockRepo()}
	svc := tracker.NewTrackerService(repo, nil)
	err := svc.DeleteWorkflowStatus(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

// TestDeleteWorkflowTransition_NotFound проверяет сценарий: удаление workflow transition not found.
func TestDeleteWorkflowTransition_NotFound(t *testing.T) {
	repo := &notFoundWorkflowTransitionDeleteRepo{mockRepo: *newMockRepo()}
	svc := tracker.NewTrackerService(repo, nil)
	err := svc.DeleteWorkflowTransition(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

// TestCreateWorkflowTransition_InvalidFromStatusID проверяет отклонение при создание workflow transition некорректный from status id.
func TestCreateWorkflowTransition_InvalidFromStatusID(t *testing.T) {
	svc, _ := newTestService()
	_, err := svc.CreateWorkflowTransition(context.Background(), uuid.New(), tracker.CreateWorkflowTransitionRequest{
		FromStatusID: "bad",
		ToStatusID:   uuid.New().String(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "from_status_id")
}

// TestCreateWorkflowTransition_InvalidToStatusID проверяет отклонение при создание workflow transition некорректный to status id.
func TestCreateWorkflowTransition_InvalidToStatusID(t *testing.T) {
	svc, _ := newTestService()
	_, err := svc.CreateWorkflowTransition(context.Background(), uuid.New(), tracker.CreateWorkflowTransitionRequest{
		FromStatusID: uuid.New().String(),
		ToStatusID:   "bad",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "to_status_id")
}

// TestDeleteLabel_NotFound проверяет сценарий: удаление label not found.
func TestDeleteLabel_NotFound(t *testing.T) {
	repo := &notFoundLabelDeleteRepo{mockRepo: *newMockRepo()}
	svc := tracker.NewTrackerService(repo, nil)
	err := svc.DeleteLabel(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}

type notFoundLabelDeleteRepo struct {
	mockRepo
}

func (r *notFoundLabelDeleteRepo) DeleteLabel(_ context.Context, _, _ uuid.UUID) error {
	return &ent.NotFoundError{}
}

// TestDeleteIssueMessageLink_IssueNotFound проверяет сценарий: удаление issue message link issue not found.
func TestDeleteIssueMessageLink_IssueNotFound(t *testing.T) {
	svc, _ := newTestService()
	err := svc.DeleteIssueMessageLink(context.Background(), uuid.New(), "HUB-404", uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}

// TestDeleteIssueMessageLink_LinkNotFound проверяет сценарий: удаление issue message link link not found.
func TestDeleteIssueMessageLink_LinkNotFound(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Link host", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	issueID, err := uuid.Parse(created.ID)
	require.NoError(t, err)
	repo.issues[issueID].Edges.Project = &ent.Project{ID: projectID}

	err = svc.DeleteIssueMessageLink(ctx, projectID, created.IssueKey, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "связь не найдена")
}
