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

// TestUpdateRetroNotes_VersionEQ_A3W1A4003 — optimistic lock на PUT /retro.
// Mutation: убрать VersionEQ/version из DTO → stale concurrent PUT оба 200 → RED.
func TestUpdateRetroNotes_VersionEQ_A3W1A4003(t *testing.T) {
	env := newSprintRepoEnv(t)
	ctx := context.Background()
	svc := sprint.NewSprintService(env.repo, nil, nil)

	created, err := svc.CreateSprint(ctx, env.projectID, sprint.CreateSprintRequest{
		Name:      "Retro OL",
		StartDate: time.Now().UTC(),
		EndDate:   time.Now().UTC().Add(7 * 24 * time.Hour),
	})
	require.NoError(t, err)
	require.Equal(t, 1, created.Version)

	sprintID := uuid.MustParse(created.ID)

	_, err = svc.UpdateRetroNotes(ctx, env.projectID, sprintID, sprint.RetroNotesRequest{
		WentWell: []string{"a"},
		Version:  0,
	})
	require.Error(t, err)
	var bad *apperrors.AppError
	require.True(t, errors.As(err, &bad))
	assert.Equal(t, 400, bad.Code)

	updated, err := svc.UpdateRetroNotes(ctx, env.projectID, sprintID, sprint.RetroNotesRequest{
		WentWell:  []string{"good"},
		ToImprove: []string{"est"},
		Actions:   []string{"poker"},
		Version:   created.Version,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, updated.Version)
	assert.NotEmpty(t, updated.RetroNotes["went_well"])

	_, err = svc.UpdateRetroNotes(ctx, env.projectID, sprintID, sprint.RetroNotesRequest{
		WentWell: []string{"stale"},
		Version:  created.Version, // stale
	})
	require.Error(t, err)
	var conflict *apperrors.AppError
	require.True(t, errors.As(err, &conflict))
	assert.Equal(t, 409, conflict.Code)
}
