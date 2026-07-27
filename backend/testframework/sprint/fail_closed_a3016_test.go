package sprint_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/internal/sprint"
)

type errFindActiveRepo struct {
	*mockRepository
	findErr error
}

func (m *errFindActiveRepo) FindActiveSprint(context.Context, uuid.UUID) (*ent.Sprint, error) {
	return nil, m.findErr
}

// TestStartSprint_FailClosedOnFindError_A3_016 — DB error from FindActiveSprint
// must not skip the active-sprint conflict check (mutation: ignore err → FAIL).
func TestStartSprint_FailClosedOnFindError_A3_016(t *testing.T) {
	base := newMockRepository()
	sprintID := uuid.New()
	base.sprints[sprintID] = &ent.Sprint{
		ID:        sprintID,
		Name:      "Sprint 1",
		Status:    sprintEnt.StatusPlanning,
		StartDate: time.Now(),
		EndDate:   time.Now().AddDate(0, 0, 14),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	repo := &errFindActiveRepo{mockRepository: base, findErr: errors.New("db down")}
	svc := sprint.NewSprintService(repo, nil, nil)

	_, err := svc.StartSprint(context.Background(), uuid.New(), sprintID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "активного спринта")
	assert.Equal(t, sprintEnt.StatusPlanning, base.sprints[sprintID].Status,
		"sprint must not start when FindActiveSprint fails")
}
