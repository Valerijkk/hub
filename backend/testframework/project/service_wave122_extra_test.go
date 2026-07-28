package project_test

// Тесты сервиса проектов (волна 122): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/invite"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// Фаза 122 — пробелы покрытия сервиса: cross-project delete, teams, invites..

func TestService_Wave122_DeleteProject_CleansCrossLinks(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	projA, err := env.repo.Create(env.ctx, "AlphaDel", "ALD", "", env.owner.ID)
	require.NoError(t, err)
	projB, err := env.repo.Create(env.ctx, "BetaDel", "BLD", "", env.owner.ID)
	require.NoError(t, err)

	statusA, err := env.client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(projA.ID).
		Save(env.ctx)
	require.NoError(t, err)

	statusB, err := env.client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(projB.ID).
		Save(env.ctx)
	require.NoError(t, err)

	src := createTestIssue(t, env, projA.ID, 1, "ALD-1", statusA.ID)
	tgt := createTestIssue(t, env, projB.ID, 1, "BLD-1", statusB.ID)

	link, err := env.client.IssueLink.Create().
		SetLinkType(issuelink.LinkTypeRelatesTo).
		SetSourceID(src.ID).
		SetTargetID(tgt.ID).
		Save(env.ctx)
	require.NoError(t, err)

	require.NoError(t, svc.DeleteProject(env.ctx, projA.ID))

	outgoing, err := env.repo.ListOutgoingCrossProjectLinks(env.ctx, projA.ID)
	require.NoError(t, err)
	assert.Empty(t, outgoing)

	_, err = env.client.IssueLink.Get(env.ctx, link.ID)
	require.Error(t, err)
}

// TestService_Wave122_ListTeams_WithMember проверяет бизнес-логику: сервис wave 122 список teams with участник.
func TestService_Wave122_ListTeams_WithMember(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	team, err := env.repo.CreateTeam(env.ctx, "Backend", "API team")
	require.NoError(t, err)
	_, err = env.repo.AddTeamMember(env.ctx, team.ID, env.owner.ID)
	require.NoError(t, err)

	teams, err := svc.ListTeams(env.ctx, env.owner.ID)
	require.NoError(t, err)
	require.NotEmpty(t, teams)

	found := false
	for _, tm := range teams {
		if tm.ID == team.ID.String() {
			found = true
			assert.GreaterOrEqual(t, tm.MemberCount, 1)
			break
		}
	}
	assert.True(t, found)
}

// TestService_Wave122_RemoveTeamMember проверяет бизнес-логику: сервис wave 122 remove team участник.
func TestService_Wave122_RemoveTeamMember(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	team, err := env.repo.CreateTeam(env.ctx, "Ops", "")
	require.NoError(t, err)
	_, err = env.repo.AddTeamMember(env.ctx, team.ID, env.owner.ID)
	require.NoError(t, err)

	require.NoError(t, svc.RemoveTeamMember(env.ctx, team.ID, env.owner.ID, env.owner.ID))

	members, err := env.repo.ListTeamMembers(env.ctx, team.ID)
	require.NoError(t, err)
	assert.Empty(t, members)
}

// TestService_Wave122_CreateInviteAndList проверяет бизнес-логику: сервис wave 122 создание приглашение and список.
func TestService_Wave122_CreateInviteAndList(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "InviteProj", "INV", "", env.owner.ID)
	require.NoError(t, err)

	inv, err := svc.CreateInvite(env.ctx, created.ID, env.owner.ID, "developer", 3, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, inv.Code)

	list, err := svc.ListInvites(env.ctx, created.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, inv.ID, list[0].ID)
}

// TestService_Wave122_RemoveMember_Ent проверяет слой репозитория: сервис wave 122 remove участник ent.
func TestService_Wave122_RemoveMember_Ent(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "RemMem", "RMM", "", env.owner.ID)
	require.NoError(t, err)
	_, err = env.repo.AddMember(env.ctx, created.ID, env.member.ID, projectmember.ProjectRoleDeveloper)
	require.NoError(t, err)

	require.NoError(t, svc.RemoveMember(env.ctx, created.ID, env.member.ID))

	members, err := svc.ListMembers(env.ctx, created.ID)
	require.NoError(t, err)
	for _, m := range members {
		assert.NotEqual(t, env.member.ID.String(), m.UserID)
	}
}

// TestService_Wave122_AddTeamToProject проверяет бизнес-логику: сервис wave 122 add team to project.
func TestService_Wave122_AddTeamToProject(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "TeamBind", "TMB", "", env.owner.ID)
	require.NoError(t, err)

	team, err := env.repo.CreateTeam(env.ctx, "Platform", "")
	require.NoError(t, err)

	require.NoError(t, svc.AddTeamToProject(env.ctx, team.ID, created.ID))

	members, err := svc.ListMembers(env.ctx, created.ID)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(members), 1)
}

// TestService_Wave122_DeleteProject_NotifyCrossLinkAssignee проверяет бизнес-логику: сервис wave 122 удаление project notify cross link assignee.
func TestService_Wave122_DeleteProject_NotifyCrossLinkAssignee(t *testing.T) {
	nats := startProjectNATSServer(t)
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, &mockCache{}, nats)

	projA, err := env.repo.Create(env.ctx, "NotifyDel", "NTD", "", env.owner.ID)
	require.NoError(t, err)
	projB, err := env.repo.Create(env.ctx, "OtherProj", "OTP", "", env.owner.ID)
	require.NoError(t, err)

	statusA, err := env.client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(projA.ID).
		Save(env.ctx)
	require.NoError(t, err)

	statusB, err := env.client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(projB.ID).
		Save(env.ctx)
	require.NoError(t, err)

	src := createTestIssue(t, env, projA.ID, 1, "NTD-1", statusA.ID)
	tgt := createTestIssue(t, env, projB.ID, 1, "OTP-1", statusB.ID)
	_, err = env.client.Issue.UpdateOneID(tgt.ID).SetAssigneeID(env.member.ID).Save(env.ctx)
	require.NoError(t, err)

	_, err = env.client.IssueLink.Create().
		SetLinkType(issuelink.LinkTypeRelatesTo).
		SetSourceID(src.ID).
		SetTargetID(tgt.ID).
		Save(env.ctx)
	require.NoError(t, err)

	require.NoError(t, svc.DeleteProject(env.ctx, projA.ID))
}

// TestService_Wave122_TryConsumeInvite проверяет бизнес-логику: сервис wave 122 try consume приглашение.
func TestService_Wave122_TryConsumeInvite(t *testing.T) {
	env := newProjectRepoEnv(t)

	created, err := env.repo.Create(env.ctx, "Consume", "CNS", "", env.owner.ID)
	require.NoError(t, err)

	inv, err := env.repo.CreateInvite(env.ctx, created.ID, env.owner.ID, "CODE123", invite.ProjectRoleDeveloper, 2, nil)
	require.NoError(t, err)

	ok, err := env.repo.TryConsumeInvite(env.ctx, inv.ID, 2)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = env.repo.TryConsumeInvite(env.ctx, inv.ID, 2)
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = env.repo.TryConsumeInvite(env.ctx, inv.ID, 2)
	require.NoError(t, err)
	assert.False(t, ok)
}

// TestService_Wave122_JoinByInvite_Ent проверяет слой репозитория: сервис wave 122 join by приглашение ent.
func TestService_Wave122_JoinByInvite_Ent(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "JoinMe", "JME", "", env.owner.ID)
	require.NoError(t, err)

	inv, err := svc.CreateInvite(env.ctx, created.ID, env.owner.ID, "developer", 5, nil)
	require.NoError(t, err)

	joined, err := svc.JoinByInvite(env.ctx, inv.Code, env.member.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID.String(), joined.ID)

	_, err = svc.JoinByInvite(env.ctx, inv.Code, env.member.ID)
	require.Error(t, err)
}
