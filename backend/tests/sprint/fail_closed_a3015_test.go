package sprint_test

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/sprint"
)

type errCountWIPRepo struct {
	*mockRepository
	countErr error
}

func (m *errCountWIPRepo) CountIssuesInStatus(context.Context, uuid.UUID) (int, error) {
	return 0, m.countErr
}

// TestCheckWIPLimit_FailClosed_A3_015 — missing status / count error must not
// silently allow the move (mutation: return nil on err → FAIL).
func TestCheckWIPLimit_FailClosed_A3_015(t *testing.T) {
	base := newMockRepository()
	svc := sprint.NewSprintService(base, nil, nil)

	err := svc.CheckWIPLimit(context.Background(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найден")

	statusID := uuid.New()
	base.statuses[statusID] = &ent.WorkflowStatus{
		ID:       statusID,
		Name:     "In Progress",
		WipLimit: 5,
	}
	repo := &errCountWIPRepo{mockRepository: base, countErr: errors.New("count failed")}
	svc2 := sprint.NewSprintService(repo, nil, nil)
	err = svc2.CheckWIPLimit(context.Background(), statusID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "WIP")
}
