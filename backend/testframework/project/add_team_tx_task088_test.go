package project_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	entproject "github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/internal/project"
)

// TestAddTeamToProject_TxRollbackOnMidFail_A2W1_A5_002 — mid-fail после AddTeamIDs
// откатывает связь команды (A2W1-A5-002 / TASK-088).
// Mutation: убрать tx / commit сразу после AddTeamIDs → после inject-fail команда остаётся привязанной.
func TestAddTeamToProject_TxRollbackOnMidFail_A2W1_A5_002(t *testing.T) {
	env := newProjectRepoEnv(t)

	team, err := env.repo.CreateTeam(env.ctx, "Ops", "ops")
	require.NoError(t, err)
	_, err = env.repo.AddTeamMember(env.ctx, team.ID, env.member.ID)
	require.NoError(t, err)

	proj, err := env.repo.Create(env.ctx, "TeamTx", "TTX", "", env.owner.ID)
	require.NoError(t, err)

	project.ExportSetAfterAddTeamLinkHook(func(context.Context) error {
		return errors.New("injected mid-fail")
	})
	t.Cleanup(func() { project.ExportSetAfterAddTeamLinkHook(nil) })

	err = env.repo.AddTeamToProject(env.ctx, team.ID, proj.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "injected mid-fail")

	reloaded, err := env.client.Project.Query().
		Where(entproject.IDEQ(proj.ID)).
		WithTeams().
		Only(env.ctx)
	require.NoError(t, err)
	assert.Empty(t, reloaded.Edges.Teams, "team link must roll back on mid-fail")
}

// TestAddTeamToProject_AtomicSuccess_A2W1_A5_002 — happy path: команда + members атомарно.
func TestAddTeamToProject_AtomicSuccess_A2W1_A5_002(t *testing.T) {
	env := newProjectRepoEnv(t)

	team, err := env.repo.CreateTeam(env.ctx, "Devs", "devs")
	require.NoError(t, err)
	_, err = env.repo.AddTeamMember(env.ctx, team.ID, env.member.ID)
	require.NoError(t, err)

	proj, err := env.repo.Create(env.ctx, "TeamOK", "TOK", "", env.owner.ID)
	require.NoError(t, err)

	require.NoError(t, env.repo.AddTeamToProject(env.ctx, team.ID, proj.ID))

	reloaded, err := env.client.Project.Query().
		Where(entproject.IDEQ(proj.ID)).
		WithTeams().
		Only(env.ctx)
	require.NoError(t, err)
	require.Len(t, reloaded.Edges.Teams, 1)
	assert.Equal(t, team.ID, reloaded.Edges.Teams[0].ID)

	mem, err := env.repo.FindMember(env.ctx, proj.ID, env.member.ID)
	require.NoError(t, err)
	assert.Equal(t, env.member.ID, mem.Edges.User.ID)
}
