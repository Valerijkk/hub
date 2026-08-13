package tracker_test

// Тесты сервиса трекера задач (волна 117): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStopTimer_NoActiveTimer проверяет сценарий: stop timer no active timer.
func TestStopTimer_NoActiveTimer(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	issue, err := svc.CreateIssue(ctx, projectID, userID, tracker.CreateIssueRequest{
		Title: "Timer issue", IssueType: "task", Priority: "medium",
	})
	require.NoError(t, err)

	_, err = svc.StopTimer(ctx, projectID, issue.IssueKey, userID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "таймер")
}

// TestLogTime_IssueNotFound проверяет сценарий: log time issue not found.
func TestLogTime_IssueNotFound(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	_, err := svc.LogTime(ctx, projectID, "MISSING-1", uuid.New(), tracker.LogTimeRequest{Minutes: 30})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}

// TestStartTimer_IssueNotFound проверяет сценарий: start timer issue not found.
func TestStartTimer_IssueNotFound(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	_, err := svc.StartTimer(ctx, projectID, "NOPE-99", uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}

// TestListTimeLogs_IssueNotFound проверяет сценарий: список time logs issue not found.
func TestListTimeLogs_IssueNotFound(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()

	_, err := svc.ListTimeLogs(ctx, projectID, "GHOST-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}
