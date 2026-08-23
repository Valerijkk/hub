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

func (m *errFindActiveRepo) StartSprintAtomic(ctx context.Context, projectID, sprintID uuid.UUID) (*ent.Sprint, error) {
	if _, err := m.GetSprintInProject(ctx, projectID, sprintID); err != nil {
		return nil, err
	}
	if _, err := m.FindActiveSprint(ctx, projectID); err != nil {
		return nil, err
	}
	return m.UpdateSprintStatus(ctx, sprintID, sprintEnt.StatusActive)
}

// TestStartSprint_FailClosedOnFindError_A3_016 — DB error from FindActiveSprint
// must not skip the active-sprint conflict check (mutation: ignore err → FAIL).
func TestStartSprint_FailClosedOnFindError_A3_016(t *testing.T) {
	base := newMockRepository()
	projectID := uuid.New()
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
	base.sprintProject[sprintID] = projectID
	repo := &errFindActiveRepo{mockRepository: base, findErr: errors.New("db down")}
	svc := sprint.NewSprintService(repo, nil, nil)

	_, err := svc.StartSprint(context.Background(), projectID, sprintID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "запуска спринта")
	assert.Equal(t, sprintEnt.StatusPlanning, base.sprints[sprintID].Status,
		"sprint must not start when FindActiveSprint fails")
}
