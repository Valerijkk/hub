package tracker_test

// Тесты сервиса трекера задач (волна 113): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

type notFoundCommentDeleteRepo struct {
	mockRepo
}

func (r *notFoundCommentDeleteRepo) DeleteComment(_ context.Context, _, _, _ uuid.UUID) error {
	return &ent.NotFoundError{}
}

type notFoundFieldDeleteRepo struct {
	mockRepo
}

func (r *notFoundFieldDeleteRepo) DeleteCustomField(_ context.Context, _ uuid.UUID) error {
	return &ent.NotFoundError{}
}

// TestDeleteComment_NotFound проверяет сценарий: удаление comment not found.
func TestDeleteComment_NotFound(t *testing.T) {
	repo := &notFoundCommentDeleteRepo{mockRepo: *newMockRepo()}
	svc := tracker.NewTrackerService(repo, nil)

	err := svc.DeleteComment(context.Background(), uuid.New(), "TST-1", uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")
}

// TestDeleteCustomField_NotFound проверяет сценарий: удаление custom field not found.
func TestDeleteCustomField_NotFound(t *testing.T) {
	repo := &notFoundFieldDeleteRepo{mockRepo: *newMockRepo()}
	svc := tracker.NewTrackerService(repo, nil)

	err := svc.DeleteCustomField(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдено")
}

// TestListCustomFields_EmptyProject проверяет сценарий: список custom fields пустой project.
func TestListCustomFields_EmptyProject(t *testing.T) {
	svc, _ := newTestService()
	fields, err := svc.ListCustomFields(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Empty(t, fields)
}

// TestGetActivity_EmptyIssue проверяет сценарий: получение activity пустой issue.
func TestGetActivity_EmptyIssue(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Activity host", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	logs, err := svc.GetActivity(ctx, projectID, created.IssueKey)
	require.NoError(t, err)
	assert.NotNil(t, logs)
}
