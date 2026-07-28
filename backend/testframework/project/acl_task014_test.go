package project_test

// TASK-014 / P1-BE-06..09: project ACL denial/regression на слое service/repo.
import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/invite"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/projecttemplate"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/project"
)

// TestRevokeInvite_ForeignProject_IDOR — P1-BE-06: revoke invite чужого project → 404.
func TestRevokeInvite_ForeignProject_IDOR(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()

	ownerA := uuid.New()
	ownerB := uuid.New()
	projA, err := svc.CreateProject(ctx, ownerA, "A", "PRA", "")
	require.NoError(t, err)
	projB, err := svc.CreateProject(ctx, ownerB, "B", "PRB", "")
	require.NoError(t, err)
	projectAID, err := uuid.Parse(projA.ID)
	require.NoError(t, err)
	projectBID, err := uuid.Parse(projB.ID)
	require.NoError(t, err)

	inv, err := svc.CreateInvite(ctx, projectAID, ownerA, "developer", 5, nil)
	require.NoError(t, err)
	inviteID, err := uuid.Parse(inv.ID)
	require.NoError(t, err)

	// Attacker с доступом к projB пытается revoke invite из projA.
	err = svc.RevokeInvite(ctx, projectBID, inviteID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
	assert.Len(t, repo.invites, 1, "invite должен остаться")

	require.NoError(t, svc.RevokeInvite(ctx, projectAID, inviteID))
	assert.Empty(t, repo.invites)
}

// TestUpdateMemberRole_LastOwnerDemote_Forbidden — P1-BE-07.
func TestUpdateMemberRole_LastOwnerDemote_Forbidden(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	projectID := uuid.New()

	repo.projects[projectID] = &ent.Project{
		ID: projectID, Name: "Solo", Key: "SLO", Status: "active",
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.members[projectID] = []*ent.ProjectMember{
		{
			ID: uuid.New(), ProjectRole: projectmember.ProjectRoleOwner, JoinedAt: time.Now(),
			Edges: ent.ProjectMemberEdges{User: &ent.User{ID: ownerID}},
		},
	}

	_, err := svc.UpdateMemberRole(ctx, projectID, ownerID, "developer")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "last owner")
}

// TestTeams_StrangerCannotAddMember — P1-BE-08 ownership.
func TestTeams_StrangerCannotAddMember(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	creatorID := uuid.New()
	strangerID := uuid.New()
	team, err := svc.CreateTeam(ctx, creatorID, "Owned", "")
	require.NoError(t, err)
	teamID, err := uuid.Parse(team.ID)
	require.NoError(t, err)

	err = svc.AddTeamMember(ctx, teamID, strangerID, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not a team member")

	teams, err := svc.ListTeams(ctx, strangerID)
	require.NoError(t, err)
	assert.Empty(t, teams)

	mine, err := svc.ListTeams(ctx, creatorID)
	require.NoError(t, err)
	require.Len(t, mine, 1)
}

// TestDeleteTemplate_ForeignUser_Forbidden — P1-BE-09.
func TestDeleteTemplate_ForeignUser_Forbidden(t *testing.T) {
	repo := newTemplateMockRepo()
	svc := project.NewTemplateService(repo)
	ctx := context.Background()

	authorID := uuid.New()
	strangerID := uuid.New()
	tplID := uuid.New()
	repo.templates[tplID] = &ent.ProjectTemplate{
		ID: tplID, Name: "Mine", IsSystem: false, CreatedAt: time.Now(),
		Edges: ent.ProjectTemplateEdges{CreatedBy: &ent.User{ID: authorID}},
	}

	err := svc.DeleteTemplate(ctx, tplID, strangerID)
	require.Error(t, err)
	_, ok := repo.templates[tplID]
	assert.True(t, ok, "шаблон не должен быть удалён чужим")

	require.NoError(t, svc.DeleteTemplate(ctx, tplID, authorID))
	_, ok = repo.templates[tplID]
	assert.False(t, ok)
}

// TestDeleteTemplate_AdminCanDeleteForeign — P1-BE-09 admin path.
func TestDeleteTemplate_AdminCanDeleteForeign(t *testing.T) {
	repo := newTemplateMockRepo()
	svc := project.NewTemplateService(repo)
	ctx := context.Background()

	authorID := uuid.New()
	adminID := uuid.New()
	repo.admins[adminID] = true
	tplID := uuid.New()
	repo.templates[tplID] = &ent.ProjectTemplate{
		ID: tplID, Name: "AdminDel", IsSystem: false, CreatedAt: time.Now(),
		Edges: ent.ProjectTemplateEdges{CreatedBy: &ent.User{ID: authorID}},
	}

	require.NoError(t, svc.DeleteTemplate(ctx, tplID, adminID))
	_, ok := repo.templates[tplID]
	assert.False(t, ok)
}

// TestRevokeInvite_Ent_ScopedByProject — P1-BE-06 на реальном Ent.
func TestRevokeInvite_Ent_ScopedByProject(t *testing.T) {
	env := newProjectRepoEnv(t)

	projA, err := env.repo.Create(env.ctx, "Project Alpha", "IDA", "", env.owner.ID)
	require.NoError(t, err)
	projB, err := env.repo.Create(env.ctx, "Project Beta", "IDB", "", env.member.ID)
	require.NoError(t, err)

	inv, err := env.repo.CreateInvite(env.ctx, projA.ID, env.owner.ID, "SCOPED1", invite.ProjectRoleDeveloper, 1, nil)
	require.NoError(t, err)

	err = env.repo.DeleteInvite(env.ctx, projB.ID, inv.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	list, err := env.repo.ListInvites(env.ctx, projA.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	require.NoError(t, env.repo.DeleteInvite(env.ctx, projA.ID, inv.ID))
}

// TestUpdateMemberRole_Ent_LastOwnerGuard — P1-BE-07 на Ent+service.
func TestUpdateMemberRole_Ent_LastOwnerGuard(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "Own", "OWN", "", env.owner.ID)
	require.NoError(t, err)

	_, err = svc.UpdateMemberRole(env.ctx, created.ID, env.owner.ID, string(projectmember.ProjectRoleDeveloper))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "last owner")
}

// TestDeleteTemplate_Ent_Ownership — P1-BE-09 на Ent.
func TestDeleteTemplate_Ent_Ownership(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewTemplateService(env.repo)

	tpl, err := env.repo.CreateTemplate(env.ctx, "EntTpl", "", projecttemplate.TemplateTypeCustom,
		map[string]interface{}{"statuses": []string{"Open"}}, &env.owner.ID, false)
	require.NoError(t, err)

	err = svc.DeleteTemplate(env.ctx, tpl.ID, env.member.ID)
	require.Error(t, err)

	// Promote member to admin — then delete OK.
	_, err = env.client.User.UpdateOneID(env.member.ID).SetSystemRole(user.SystemRoleAdmin).Save(env.ctx)
	require.NoError(t, err)
	require.NoError(t, svc.DeleteTemplate(env.ctx, tpl.ID, env.member.ID))
}
