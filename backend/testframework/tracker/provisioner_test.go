package tracker_test

// Тесты модуля трекера задач: интеграционные и unit-сценарии testframework.
import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestTrackerProjectProvisioner_SeedAndBoard проверяет сценарий: tracker project провизионинг seed and board.
func TestTrackerProjectProvisioner_SeedAndBoard(t *testing.T) {
	_, repo := newTestService()
	p := tracker.NewTrackerProjectProvisioner(repo)
	ctx := context.Background()
	projectID := uuid.New()

	require.NoError(t, p.SeedDefaultWorkflow(ctx, projectID))
	require.NoError(t, p.CreateDefaultBoard(ctx, projectID))

	boards, err := repo.ListBoards(ctx, projectID)
	require.NoError(t, err)
	require.NotEmpty(t, boards)
	assert.Equal(t, "Main Board", boards[0].Name)
}
