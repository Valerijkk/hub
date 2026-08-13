package tracker_test

// Тесты сервиса трекера задач (волна 112): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// TestUpdateIssue_InvalidStatusID проверяет отклонение при обновление issue некорректный status id.
func TestUpdateIssue_InvalidStatusID(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Bad status", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	badStatus := "not-a-uuid"
	_, err = svc.UpdateIssue(ctx, projectID, created.IssueKey, userID, tracker.UpdateIssueRequest{
		StatusID: &badStatus,
		Version:  created.Version,
	})
	require.Error(t, err)
}

// TestUpdateIssue_InvalidAssigneeID проверяет отклонение при обновление issue некорректный assignee id.
func TestUpdateIssue_InvalidAssigneeID(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Bad assignee", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	badAssignee := "invalid"
	_, err = svc.UpdateIssue(ctx, projectID, created.IssueKey, userID, tracker.UpdateIssueRequest{
		AssigneeID: &badAssignee,
		Version:    created.Version,
	})
	require.Error(t, err)
}

// TestUpdateIssue_InvalidLabelID проверяет отклонение при обновление issue некорректный label id.
func TestUpdateIssue_InvalidLabelID(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Bad label", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	labels := []string{"not-uuid"}
	_, err = svc.UpdateIssue(ctx, projectID, created.IssueKey, userID, tracker.UpdateIssueRequest{
		Labels:  &labels,
		Version: created.Version,
	})
	require.Error(t, err)
}

// TestUpdateIssue_ForbiddenStatusTransition проверяет отклонение при обновление issue запрещённый status transition.
func TestUpdateIssue_ForbiddenStatusTransition(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Transition test", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	issue, err := repo.FindIssueByKey(ctx, projectID, created.IssueKey)
	require.NoError(t, err)
	require.NotNil(t, issue.Edges.Status)

	unreachable := uuid.New()
	repo.statuses[unreachable] = &ent.WorkflowStatus{
		ID: unreachable, Name: "Blocked", Category: workflowstatus.CategoryInProgress,
	}
	statusStr := unreachable.String()
	_, err = svc.UpdateIssue(ctx, projectID, created.IssueKey, userID, tracker.UpdateIssueRequest{
		StatusID: &statusStr,
		Version:  created.Version,
	})
	require.Error(t, err)
}

// TestUpdateIssue_ValidStatusTransition проверяет сценарий: обновление issue корректный status transition.
func TestUpdateIssue_ValidStatusTransition(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Move forward", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	issue, err := repo.FindIssueByKey(ctx, projectID, created.IssueKey)
	require.NoError(t, err)
	newStatusID := repo.addStatusWithTransition("In Progress", workflowstatus.CategoryInProgress, issue.Edges.Status.ID)
	statusStr := newStatusID.String()

	updated, err := svc.UpdateIssue(ctx, projectID, created.IssueKey, userID, tracker.UpdateIssueRequest{
		StatusID: &statusStr,
		Version:  created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, statusStr, updated.Status.ID)
}

// TestUpdateIssue_WithDescriptionAndChecklist проверяет сценарий: обновление issue with description and checklist.
func TestUpdateIssue_WithDescriptionAndChecklist(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Rich update", IssueType: "task", Priority: "low",
	})
	require.NoError(t, err)

	desc := map[string]interface{}{
		"type": "doc",
		"children": []interface{}{
			map[string]interface{}{
				"type": "p",
				"children": []interface{}{
					map[string]interface{}{"text": "Updated body"},
				},
			},
		},
	}
	checklist := []tracker.ChecklistItem{{ID: "1", Text: "Step one", Checked: false}}
	newTitle := "Rich update v2"
	updated, err := svc.UpdateIssue(ctx, projectID, created.IssueKey, userID, tracker.UpdateIssueRequest{
		Title:           &newTitle,
		DescriptionJSON: &desc,
		Checklist:       &checklist,
		Version:         created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, newTitle, updated.Title)
}

// TestDeleteComment_Success проверяет сценарий: удаление comment успех.
func TestDeleteComment_Success(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	created, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Comment host", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	comment, err := svc.CreateComment(ctx, projectID, created.IssueKey, userID, tracker.CreateCommentRequest{BodyText: "remove me"})
	require.NoError(t, err)

	commentID, err := uuid.Parse(comment.ID)
	require.NoError(t, err)
	require.NoError(t, svc.DeleteComment(ctx, projectID, created.IssueKey, userID, commentID))
	_, ok := repo.comments[commentID]
	assert.False(t, ok)
}

// TestDeleteCustomField_Success проверяет сценарий: удаление custom field успех.
func TestDeleteCustomField_Success(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	field, err := svc.CreateCustomField(ctx, projectID, tracker.CreateCustomFieldRequest{
		Name: "Severity", FieldType: "select", Options: []string{"low", "high"},
	})
	require.NoError(t, err)

	fieldID, err := uuid.Parse(field.ID)
	require.NoError(t, err)
	require.NoError(t, svc.DeleteCustomField(ctx, projectID, fieldID))
	assert.Empty(t, repo.customFields)
}
