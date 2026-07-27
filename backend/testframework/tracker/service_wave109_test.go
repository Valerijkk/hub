package tracker_test

// Тесты сервиса трекера задач (волна 109): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestGetBoardData_SwimlaneByAssignee проверяет сценарий: получение board data swimlane by assignee.
func TestGetBoardData_SwimlaneByAssignee(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	assigneeID := uuid.New()

	board, err := svc.CreateBoard(ctx, projectID, tracker.CreateBoardRequest{Name: "Swim Board"})
	require.NoError(t, err)

	_, err = svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title:      "Assigned task",
		IssueType:  "task",
		Priority:   "high",
		AssigneeID: assigneeID.String(),
	})
	require.NoError(t, err)

	_, err = svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title:     "Unassigned task",
		IssueType: "bug",
		Priority:  "low",
	})
	require.NoError(t, err)

	boardID, err := uuid.Parse(board.ID)
	require.NoError(t, err)

	data, err := svc.GetBoardData(ctx, projectID, boardID, "assignee")
	require.NoError(t, err)
	require.NotEmpty(t, data.Swimlanes)
	assert.NotEmpty(t, data.Columns)
	_ = repo
}

// TestGetBoardData_SwimlaneByPriority проверяет сценарий: получение board data swimlane by priority.
func TestGetBoardData_SwimlaneByPriority(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	board, err := svc.CreateBoard(ctx, projectID, tracker.CreateBoardRequest{Name: "Priority Board"})
	require.NoError(t, err)

	_, err = svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "High prio", IssueType: "task", Priority: "high",
	})
	require.NoError(t, err)

	boardID, err := uuid.Parse(board.ID)
	require.NoError(t, err)

	data, err := svc.GetBoardData(ctx, projectID, boardID, "priority")
	require.NoError(t, err)
	require.NotEmpty(t, data.Swimlanes)
}

// TestGetBoardData_NotFound проверяет сценарий: получение board data not found.
func TestGetBoardData_NotFound(t *testing.T) {
	svc, _ := newTestService()
	_, err := svc.GetBoardData(context.Background(), uuid.New(), uuid.New(), "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}
