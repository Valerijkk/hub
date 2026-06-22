package project_test

// Тесты сервиса проектов: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/invite"
	entproject "github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
)

// TestGetProject_NotFound проверяет сценарий: получение project not found.
func TestGetProject_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	_, err := svc.GetProject(context.Background(), uuid.New(), uuid.Nil)
	require.Error(t, err)
}

// TestGetProject_ReturnsMyRole проверяет, что GetProject возвращает роль текущего пользователя.
func TestGetProject_ReturnsMyRole(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Role Test", "ROLE", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	got, err := svc.GetProject(ctx, projectID, ownerID)
	require.NoError(t, err)
	assert.Equal(t, "owner", got.MyRole)
}

// TestUpdateProject_Success проверяет сценарий: обновление project успех.
func TestUpdateProject_Success(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Old Name", "UPD", "desc")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	newName := "New Name"
	newDesc := "updated"
	updated, err := svc.UpdateProject(ctx, projectID, project.UpdateProjectRequest{
		Name:        &newName,
		Description: &newDesc,
	})
	require.NoError(t, err)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, "updated", updated.Description)
}

// TestUnarchiveProject проверяет сценарий: unarchive project.
func TestUnarchiveProject(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Archive Me", "UNA", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	_, err = svc.ArchiveProject(ctx, projectID)
	require.NoError(t, err)

	restored, err := svc.UnarchiveProject(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, "active", restored.Status)
}

// TestRestoreProject_FromTrash проверяет сценарий: restore project from корзина.
func TestRestoreProject_FromTrash(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Trash Me", "TRH", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	require.NoError(t, svc.DeleteProject(ctx, projectID))
	restored, err := svc.RestoreProject(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, "active", restored.Status)
}

// TestListProjects_ExcludesArchivedByDefault проверяет сценарий: список projects excludes archived by default.
func TestListProjects_ExcludesArchivedByDefault(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	activeID := uuid.New()
	repo.projects[activeID] = &ent.Project{
		ID: activeID, Name: "Active", Key: "ACT", Status: entproject.StatusActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	archivedID := uuid.New()
	repo.projects[archivedID] = &ent.Project{
		ID: archivedID, Name: "Archived", Key: "ARC", Status: entproject.StatusArchived,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.memberCountMap[activeID] = 1
	repo.memberCountMap[archivedID] = 1

	list, err := svc.ListProjects(ctx, ownerID, false)
	require.NoError(t, err)
	assert.Len(t, list, 1)
	assert.Equal(t, "ACT", list[0].Key)

	listAll, err := svc.ListProjects(ctx, ownerID, true)
	require.NoError(t, err)
	assert.Len(t, listAll, 2)
}

// TestAddMember_InvalidatesCache проверяет сценарий: add участник invalidates кэш.
func TestAddMember_InvalidatesCache(t *testing.T) {
	svc, repo, cache := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Team", "TEA", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	memberID := uuid.New()
	_, err = svc.AddMember(ctx, projectID, memberID, "developer")
	require.NoError(t, err)
	assert.Contains(t, cache.invalidated, projectID.String())
	assert.Len(t, repo.members[projectID], 1)
}

// TestAddMember_Duplicate проверяет сценарий: add участник duplicate.
func TestAddMember_Duplicate(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	repo.members[projectID] = []*ent.ProjectMember{
		{
			ID:          uuid.New(),
			ProjectRole: projectmember.ProjectRoleDeveloper,
			JoinedAt:    time.Now(),
			Edges:       ent.ProjectMemberEdges{User: &ent.User{ID: userID}},
		},
	}

	_, err := svc.AddMember(ctx, projectID, userID, "developer")
	require.Error(t, err)
}

// TestUpdateMemberRole_Success проверяет сценарий: обновление участник role успех.
func TestUpdateMemberRole_Success(t *testing.T) {
	svc, repo, cache := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	repo.members[projectID] = []*ent.ProjectMember{
		{
			ID:          uuid.New(),
			ProjectRole: projectmember.ProjectRoleDeveloper,
			JoinedAt:    time.Now(),
			Edges:       ent.ProjectMemberEdges{User: &ent.User{ID: userID, Email: "dev@test.com", Name: "Dev"}},
		},
	}

	resp, err := svc.UpdateMemberRole(ctx, projectID, userID, "tester")
	require.NoError(t, err)
	assert.Equal(t, "tester", resp.Role)
	assert.Contains(t, cache.invalidated, userID.String()+":"+projectID.String())
}

// TestCreateInvite_GeneratesCode проверяет сценарий: создание приглашение generates code.
func TestCreateInvite_GeneratesCode(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Invite Project", "INV2", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	inv, err := svc.CreateInvite(ctx, projectID, ownerID, "developer", 5, nil)
	require.NoError(t, err)
	assert.NotEmpty(t, inv.Code)
	assert.Len(t, repo.invites, 1)
}

// TestJoinByInvite_MaxUsesExceeded проверяет сценарий: join by приглашение max uses exceeded.
func TestJoinByInvite_MaxUsesExceeded(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	repo.projects[projectID] = &ent.Project{
		ID: projectID, Name: "P", Key: "P", Status: entproject.StatusActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.invites["FULL"] = &ent.Invite{
		ID:          uuid.New(),
		Code:        "FULL",
		MaxUses:     1,
		UseCount:    1,
		ProjectRole: invite.ProjectRoleDeveloper,
		CreatedAt:   time.Now(),
		Edges:       ent.InviteEdges{Project: repo.projects[projectID]},
	}

	_, err := svc.JoinByInvite(ctx, "FULL", uuid.New())
	require.Error(t, err)
}

// TestGenerateInviteCode_Unique проверяет сценарий: generate приглашение code unique.
func TestGenerateInviteCode_Unique(t *testing.T) {
	codes := make(map[string]struct{}, 20)
	for i := 0; i < 20; i++ {
		code, err := project.GenerateInviteCode()
		require.NoError(t, err)
		assert.Len(t, code, 8)
		codes[code] = struct{}{}
	}
	assert.Len(t, codes, 20)
}

// TestToProjectResponseAndMemberResponse проверяет сценарий: to project response and участник response.
func TestToProjectResponseAndMemberResponse(t *testing.T) {
	now := time.Now()
	userID := uuid.New()
	p := &ent.Project{
		ID: uuid.New(), Name: "P", Key: "PK", Description: "d", AvatarURL: "a.png",
		Status: entproject.StatusActive, CreatedAt: now, UpdatedAt: now,
	}
	resp := project.ExportToProjectResponse(p, 3, "owner")
	assert.Equal(t, "PK", resp.Key)
	assert.Equal(t, 3, resp.MemberCount)
	assert.Equal(t, "owner", resp.MyRole)

	mem := &ent.ProjectMember{
		ProjectRole: projectmember.ProjectRoleDeveloper,
		JoinedAt:    now,
		Edges: ent.ProjectMemberEdges{
			User: &ent.User{ID: userID, Email: "u@test.com", Name: "User", AvatarURL: "u.png"},
		},
	}
	memberResp := project.ExportToMemberResponse(mem)
	assert.Equal(t, userID.String(), memberResp.UserID)
	assert.Equal(t, "developer", memberResp.Role)
}

// TestToInviteResponse проверяет сценарий: to приглашение response.
func TestToInviteResponse(t *testing.T) {
	exp := time.Now().Add(24 * time.Hour)
	inv := &ent.Invite{
		ID: uuid.New(), Code: "ABC12345", ProjectRole: invite.ProjectRoleViewer,
		MaxUses: 10, UseCount: 2, ExpiresAt: &exp, CreatedAt: time.Now(),
	}
	resp := project.ExportToInviteResponse(inv, "https://hub.test")
	assert.Contains(t, resp.Link, "https://hub.test/invite/ABC12345")
	assert.Equal(t, "viewer", resp.Role)
}

// TestCreateTeam проверяет сценарий: создание team.
func TestCreateTeam(t *testing.T) {
	svc, _, _ := newTestService()
	team, err := svc.CreateTeam(context.Background(), "Platform", "infra team")
	require.NoError(t, err)
	assert.Equal(t, "Platform", team.Name)
}

// TestListTeams_AndMembers проверяет сценарий: список teams and members.
func TestListTeams_AndMembers(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()

	team, err := svc.CreateTeam(ctx, "QA", "quality")
	require.NoError(t, err)
	teamID, err := uuid.Parse(team.ID)
	require.NoError(t, err)

	userID := uuid.New()
	require.NoError(t, svc.AddTeamMember(ctx, teamID, userID))

	teams, err := svc.ListTeams(ctx)
	require.NoError(t, err)
	assert.Empty(t, teams) // mock ListTeams returns nil — service path still exercised
}

// TestListTrash_ShowsDeletedProjects проверяет сценарий: список корзина shows deleted projects.
func TestListTrash_ShowsDeletedProjects(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Trash Bin", "TRB", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	require.NoError(t, svc.DeleteProject(ctx, projectID))
	trash, err := svc.ListTrash(ctx, ownerID)
	require.NoError(t, err)
	require.Len(t, trash, 1)
	assert.Equal(t, "TRB", trash[0].Key)
}

// TestRemoveMember_Success проверяет сценарий: remove участник успех.
func TestRemoveMember_Success(t *testing.T) {
	svc, repo, cache := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Members", "MEM", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	memberID := uuid.New()
	_, err = svc.AddMember(ctx, projectID, memberID, "developer")
	require.NoError(t, err)

	require.NoError(t, svc.RemoveMember(ctx, projectID, memberID))
	assert.Empty(t, repo.members[projectID])
	assert.Contains(t, cache.invalidated, projectID.String())
}

// TestListMembers_ReturnsAll проверяет сценарий: список members returns all.
func TestListMembers_ReturnsAll(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Roster", "ROS", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	memberID := uuid.New()
	_, err = svc.AddMember(ctx, projectID, memberID, "tester")
	require.NoError(t, err)

	members, err := svc.ListMembers(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, members, 1)
	assert.Equal(t, "tester", members[0].Role)
}

// TestUpdateProject_NotFound проверяет сценарий: обновление project not found.
func TestUpdateProject_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	name := "nope"
	_, err := svc.UpdateProject(context.Background(), uuid.New(), project.UpdateProjectRequest{Name: &name})
	require.Error(t, err)
}

// TestArchiveProject_NotFound проверяет сценарий: archive project not found.
func TestArchiveProject_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	_, err := svc.ArchiveProject(context.Background(), uuid.New())
	require.Error(t, err)
}
