//go:build hubtest

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
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// errBatchCountsRepo — мутация W1-A4-012: GetSprintIssueCounts всегда err.
type errBatchCountsRepo struct {
	*mockRepository
}

func (r *errBatchCountsRepo) GetSprintIssueCounts(_ context.Context, _ []uuid.UUID) (map[uuid.UUID]int, error) {
	return nil, errors.New("counts db down")
}

// TestListSprints_CountsError_FailClosed_TASK057 — W1-A4-012 / P3-BE-01.
// Мутация: вернуть silent `counts = make(map...)` → ListSprints без err с нулями,
// тест краснеет (ожидаем Internal).
func TestListSprints_CountsError_FailClosed_TASK057(t *testing.T) {
	repo := &errBatchCountsRepo{mockRepository: newMockRepository()}
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	projectID := uuid.New()

	sprintID := uuid.New()
	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "S1", Status: sprintEnt.StatusPlanning,
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14),
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.sprintProject[sprintID] = projectID

	list, err := svc.ListSprints(ctx, projectID)
	require.Error(t, err)
	assert.Nil(t, list)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 500, appErr.Code)
	assert.Contains(t, appErr.Message, "счётчиков")
}
