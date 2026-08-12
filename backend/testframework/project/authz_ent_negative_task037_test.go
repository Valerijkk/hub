//go:build hubtest

package project_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/project"
)

// TASK-037 / P2-BE-02 / W1-A5-012 — Ent stranger denial на teams (не mockRepo).
// Mutation: убрать membership check в AddTeamMember → stranger добавит участника.

func TestAddTeamMember_RejectsStranger_Ent_TASK037(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	team, err := svc.CreateTeam(env.ctx, env.owner.ID, "Ent Team", "owned")
	require.NoError(t, err)
	teamID, err := uuid.Parse(team.ID)
	require.NoError(t, err)

	err = svc.AddTeamMember(env.ctx, teamID, env.member.ID, env.member.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a team member")

	members, err := env.repo.ListTeamMembers(env.ctx, teamID)
	require.NoError(t, err)
	require.Len(t, members, 1)
	assert.Equal(t, env.owner.ID, members[0].Edges.User.ID)
}

func TestListTeams_HidesForeign_Ent_TASK037(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	_, err := svc.CreateTeam(env.ctx, env.owner.ID, "Secret Team", "")
	require.NoError(t, err)

	listed, err := svc.ListTeams(env.ctx, env.member.ID)
	require.NoError(t, err)
	assert.Empty(t, listed)

	mine, err := svc.ListTeams(env.ctx, env.owner.ID)
	require.NoError(t, err)
	require.Len(t, mine, 1)
}
