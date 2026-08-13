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
	"github.com/valerii/hub/backend/internal/sprint"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// TestUpdateSprint_VersionEQ_TASK095 — optimistic lock на UpdateSprint.
// Mutation: убрать VersionEQ из EntRepository.UpdateSprint → stale version проходит.
func TestUpdateSprint_VersionEQ_TASK095(t *testing.T) {
	env := newSprintRepoEnv(t)
	ctx := context.Background()
	svc := sprint.NewSprintService(env.repo, nil, nil)

	created, err := svc.CreateSprint(ctx, env.projectID, sprint.CreateSprintRequest{
		Name:      "OL Sprint",
		StartDate: time.Now().UTC(),
		EndDate:   time.Now().UTC().Add(7 * 24 * time.Hour),
	})
	require.NoError(t, err)
	require.Equal(t, 1, created.Version)

	sprintID := uuid.MustParse(created.ID)
	name := "renamed"
	_, err = svc.UpdateSprint(ctx, env.projectID, sprintID, sprint.UpdateSprintRequest{
		Name:    &name,
		Version: 0,
	})
	require.Error(t, err)
	var bad *apperrors.AppError
	require.True(t, errors.As(err, &bad))
	assert.Equal(t, 400, bad.Code)

	updated, err := svc.UpdateSprint(ctx, env.projectID, sprintID, sprint.UpdateSprintRequest{
		Name:    &name,
		Version: created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Name)
	assert.Equal(t, 2, updated.Version)

	_, err = svc.UpdateSprint(ctx, env.projectID, sprintID, sprint.UpdateSprintRequest{
		Name:    &name,
		Version: created.Version, // stale
	})
	require.Error(t, err)
	var conflict *apperrors.AppError
	require.True(t, errors.As(err, &conflict))
	assert.Equal(t, 409, conflict.Code)
}
