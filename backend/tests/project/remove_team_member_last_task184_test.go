package project_test

// TASK-184 / C-W19-TT-001: RemoveTeamMember last-member → 409; orphan team запрещён.
// Mutation: убрать TeamMemberCount≤1 guard в service.RemoveTeamMember → этот тест RED.
import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/project"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// TestRemoveTeamMember_RejectsLastMember_TASK184 — last remaining member → Conflict 409;
// команда остаётся с ≥1 участником (requireTeamMember не orphan).
func TestRemoveTeamMember_RejectsLastMember_TASK184(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	team, err := env.repo.CreateTeam(env.ctx, "Solo", "")
	require.NoError(t, err)
	_, err = env.repo.AddTeamMember(env.ctx, team.ID, env.owner.ID)
	require.NoError(t, err)

	err = svc.RemoveTeamMember(env.ctx, team.ID, env.owner.ID, env.owner.ID)
	require.Error(t, err)

	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr), "expected AppError, got %T", err)
	assert.Equal(t, 409, appErr.Code)
	assert.Contains(t, appErr.Message, "cannot remove the last member")

	count, err := env.repo.TeamMemberCount(env.ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "team must keep the last member (no orphan)")
}

// TestRemoveTeamMember_AllowsWhenMultiple_TASK184 — при ≥2 участниках remove OK;
// после сведения к 1 — повторный remove last → 409.
func TestRemoveTeamMember_AllowsWhenMultiple_TASK184(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	team, err := env.repo.CreateTeam(env.ctx, "Duo", "")
	require.NoError(t, err)
	_, err = env.repo.AddTeamMember(env.ctx, team.ID, env.owner.ID)
	require.NoError(t, err)
	_, err = env.repo.AddTeamMember(env.ctx, team.ID, env.member.ID)
	require.NoError(t, err)

	require.NoError(t, svc.RemoveTeamMember(env.ctx, team.ID, env.owner.ID, env.member.ID))

	count, err := env.repo.TeamMemberCount(env.ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	err = svc.RemoveTeamMember(env.ctx, team.ID, env.owner.ID, env.owner.ID)
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.True(t, errors.As(err, &appErr))
	assert.Equal(t, 409, appErr.Code)

	count, err = env.repo.TeamMemberCount(env.ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}
