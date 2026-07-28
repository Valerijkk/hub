package project_test

// FA-11 — закрытие micro-gap project coverage 81.8% → 82%.
import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/invite"
	"github.com/valerii/hub/backend/internal/project"
)

func TestService_FA11_JoinByInvite_InviteWithoutProject(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()

	repo.invites["ORPHAN"] = &ent.Invite{
		ID:          uuid.New(),
		Code:        "ORPHAN",
		MaxUses:     5,
		ProjectRole: invite.ProjectRoleDeveloper,
		CreatedAt:   time.Now(),
	}

	_, err := svc.JoinByInvite(ctx, "ORPHAN", uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invite has no associated project")
}

func TestService_FA11_JoinByInvite_AlreadyMemberConflict(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()

	resp, err := svc.CreateProject(ctx, ownerID, "MemberDup", "MDP", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(resp.ID)
	require.NoError(t, err)

	repo.invites["DUP"] = &ent.Invite{
		ID:          uuid.New(),
		Code:        "DUP",
		MaxUses:     5,
		ProjectRole: invite.ProjectRoleDeveloper,
		CreatedAt:   time.Now(),
		Edges: ent.InviteEdges{
			Project: repo.projects[projectID],
		},
	}
	_, err = svc.AddMember(ctx, projectID, memberID, "developer")
	require.NoError(t, err)

	_, err = svc.JoinByInvite(ctx, "DUP", memberID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already a member")
}

func TestService_FA11_JoinByInvite_FindInviteInternalError(t *testing.T) {
	svc, repo, _ := newTestService()
	repo.findInviteByCodeErr = errors.New("db unavailable")

	_, err := svc.JoinByInvite(context.Background(), "ANY", uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to find invite")
}

func TestService_FA11_JoinByInvite_FindMemberInternalError(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	resp, err := svc.CreateProject(ctx, ownerID, "MemberErr", "MER", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(resp.ID)
	require.NoError(t, err)

	repo.invites["MEMERR"] = &ent.Invite{
		ID:          uuid.New(),
		Code:        "MEMERR",
		MaxUses:     5,
		ProjectRole: invite.ProjectRoleDeveloper,
		CreatedAt:   time.Now(),
		Edges: ent.InviteEdges{
			Project: repo.projects[projectID],
		},
	}
	repo.findMemberErr = errors.New("membership lookup failed")

	_, err = svc.JoinByInvite(ctx, "MEMERR", uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to check membership")
}

func TestService_FA11_RemoveTeamMember_RepoError(t *testing.T) {
	svc, repo, _ := newTestService()
	repo.removeTeamMemberErr = errors.New("delete failed")

	ctx := context.Background()
	actorID := uuid.New()
	team, err := svc.CreateTeam(ctx, actorID, "FA11", "")
	require.NoError(t, err)
	teamID, err := uuid.Parse(team.ID)
	require.NoError(t, err)

	err = svc.RemoveTeamMember(ctx, teamID, actorID, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to remove team member")
}

func TestService_FA11_AddTeamToProject_NotFound(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	team, err := env.repo.CreateTeam(env.ctx, "FA11 Team", "")
	require.NoError(t, err)

	err = svc.AddTeamToProject(env.ctx, team.ID, uuid.New())
	require.Error(t, err)
}
