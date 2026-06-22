package project_test

// Тесты репозитория проектов: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/invite"
	entproject "github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
)

type projectRepoEnv struct {
	client *ent.Client
	repo   *project.EntRepository
	ctx    context.Context
	owner  *ent.User
	member *ent.User
}

func newProjectRepoEnv(t *testing.T) projectRepoEnv {
	t.Helper()
	dsn := "file:projectrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	owner, err := client.User.Create().
		SetEmail("owner-" + uuid.New().String()[:8] + "@test.com").
		SetName("Owner").
		Save(ctx)
	require.NoError(t, err)

	member, err := client.User.Create().
		SetEmail("member-" + uuid.New().String()[:8] + "@test.com").
		SetName("Member").
		Save(ctx)
	require.NoError(t, err)

	return projectRepoEnv{
		client: client,
		repo:   project.NewEntRepository(client, nil),
		ctx:    ctx,
		owner:  owner,
		member: member,
	}
}

// TestEntRepository_CreateProject_AddsOwnerMember проверяет слой репозитория: ent репозиторий создание project adds owner участник.
func TestEntRepository_CreateProject_AddsOwnerMember(t *testing.T) {
	env := newProjectRepoEnv(t)

	created, err := env.repo.Create(env.ctx, "Alpha", "ALP", "desc", env.owner.ID)
	require.NoError(t, err)
	assert.Equal(t, "Alpha", created.Name)
	assert.Equal(t, entproject.StatusActive, created.Status)

	count, err := env.repo.MemberCount(env.ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	members, err := env.repo.ListMembers(env.ctx, created.ID)
	require.NoError(t, err)
	require.Len(t, members, 1)
	assert.Equal(t, projectmember.ProjectRoleOwner, members[0].ProjectRole)
}

// TestEntRepository_FindByKey проверяет слой репозитория: ent репозиторий find by key.
func TestEntRepository_FindByKey(t *testing.T) {
	env := newProjectRepoEnv(t)

	created, err := env.repo.Create(env.ctx, "Beta", "BET", "", env.owner.ID)
	require.NoError(t, err)

	found, err := env.repo.FindByKey(env.ctx, "BET")
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)

	_, err = env.repo.FindByKey(env.ctx, "MISSING")
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

// TestEntRepository_ListByUserID_ExcludesDeleted проверяет слой репозитория: ent репозиторий список by user id excludes deleted.
func TestEntRepository_ListByUserID_ExcludesDeleted(t *testing.T) {
	env := newProjectRepoEnv(t)

	active, err := env.repo.Create(env.ctx, "Active", "ACT", "", env.owner.ID)
	require.NoError(t, err)

	trashed, err := env.repo.Create(env.ctx, "Trash", "TRH", "", env.owner.ID)
	require.NoError(t, err)
	_, err = env.repo.UpdateStatus(env.ctx, trashed.ID, entproject.StatusDeleted, true)
	require.NoError(t, err)

	list, err := env.repo.ListByUserID(env.ctx, env.owner.ID, false)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, active.ID, list[0].ID)

	trashList, err := env.repo.ListTrash(env.ctx, env.owner.ID)
	require.NoError(t, err)
	require.Len(t, trashList, 1)
	assert.Equal(t, trashed.ID, trashList[0].ID)
}

// TestEntRepository_UpdateAndStatus проверяет слой репозитория: ent репозиторий обновление and status.
func TestEntRepository_UpdateAndStatus(t *testing.T) {
	env := newProjectRepoEnv(t)

	created, err := env.repo.Create(env.ctx, "Gamma", "GAM", "old", env.owner.ID)
	require.NoError(t, err)

	name := "Gamma Renamed"
	desc := "new desc"
	updated, err := env.repo.Update(env.ctx, created.ID, &name, &desc, nil)
	require.NoError(t, err)
	assert.Equal(t, name, updated.Name)
	assert.Equal(t, desc, updated.Description)

	archived, err := env.repo.UpdateStatus(env.ctx, created.ID, entproject.StatusArchived, false)
	require.NoError(t, err)
	assert.Equal(t, entproject.StatusArchived, archived.Status)

	restored, err := env.repo.UpdateStatus(env.ctx, created.ID, entproject.StatusActive, false)
	require.NoError(t, err)
	assert.Equal(t, entproject.StatusActive, restored.Status)
}

// TestEntRepository_MemberCRUD проверяет слой репозитория: ent репозиторий участник crud.
func TestEntRepository_MemberCRUD(t *testing.T) {
	env := newProjectRepoEnv(t)

	proj, err := env.repo.Create(env.ctx, "Members", "MEM", "", env.owner.ID)
	require.NoError(t, err)

	added, err := env.repo.AddMember(env.ctx, proj.ID, env.member.ID, projectmember.ProjectRoleDeveloper)
	require.NoError(t, err)
	assert.Equal(t, env.member.ID, added.Edges.User.ID)

	updated, err := env.repo.UpdateMemberRole(env.ctx, proj.ID, env.member.ID, projectmember.ProjectRoleTester)
	require.NoError(t, err)
	assert.Equal(t, projectmember.ProjectRoleTester, updated.ProjectRole)

	found, err := env.repo.FindMember(env.ctx, proj.ID, env.member.ID)
	require.NoError(t, err)
	assert.Equal(t, updated.ID, found.ID)

	owners, err := env.repo.CountOwners(env.ctx, proj.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, owners)

	require.NoError(t, env.repo.RemoveMember(env.ctx, proj.ID, env.member.ID))
	_, err = env.repo.FindMember(env.ctx, proj.ID, env.member.ID)
	require.Error(t, err)
}

// TestEntRepository_MemberCountBatch проверяет слой репозитория: ent репозиторий участник count batch.
func TestEntRepository_MemberCountBatch(t *testing.T) {
	env := newProjectRepoEnv(t)

	p1, err := env.repo.Create(env.ctx, "One", "ONE", "", env.owner.ID)
	require.NoError(t, err)
	p2, err := env.repo.Create(env.ctx, "Two", "TWO", "", env.owner.ID)
	require.NoError(t, err)
	_, err = env.repo.AddMember(env.ctx, p1.ID, env.member.ID, projectmember.ProjectRoleDeveloper)
	require.NoError(t, err)

	counts, err := env.repo.MemberCountBatch(env.ctx, []uuid.UUID{p1.ID, p2.ID})
	require.NoError(t, err)
	assert.Equal(t, 2, counts[p1.ID])
	assert.Equal(t, 1, counts[p2.ID])
}

// TestEntRepository_InviteFlow проверяет слой репозитория: ent репозиторий приглашение поток.
func TestEntRepository_InviteFlow(t *testing.T) {
	env := newProjectRepoEnv(t)

	proj, err := env.repo.Create(env.ctx, "Invite", "INV", "", env.owner.ID)
	require.NoError(t, err)

	expires := time.Now().UTC().Add(24 * time.Hour)
	inv, err := env.repo.CreateInvite(env.ctx, proj.ID, env.owner.ID, "JOIN123", invite.ProjectRoleDeveloper, 2, &expires)
	require.NoError(t, err)
	assert.Equal(t, "JOIN123", inv.Code)

	found, err := env.repo.FindInviteByCode(env.ctx, "JOIN123")
	require.NoError(t, err)
	assert.Equal(t, inv.ID, found.ID)

	list, err := env.repo.ListInvites(env.ctx, proj.ID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	consumed, err := env.repo.TryConsumeInvite(env.ctx, inv.ID, 2)
	require.NoError(t, err)
	assert.True(t, consumed)

	require.NoError(t, env.repo.IncrementInviteUseCount(env.ctx, inv.ID))

	require.NoError(t, env.repo.DeleteInvite(env.ctx, inv.ID))
	list, err = env.repo.ListInvites(env.ctx, proj.ID)
	require.NoError(t, err)
	assert.Empty(t, list)
}

// TestEntRepository_TeamFlow проверяет слой репозитория: ent репозиторий team поток.
func TestEntRepository_TeamFlow(t *testing.T) {
	env := newProjectRepoEnv(t)

	team, err := env.repo.CreateTeam(env.ctx, "Platform", "infra team")
	require.NoError(t, err)

	_, err = env.repo.AddTeamMember(env.ctx, team.ID, env.owner.ID)
	require.NoError(t, err)

	count, err := env.repo.TeamMemberCount(env.ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, count)

	members, err := env.repo.ListTeamMembers(env.ctx, team.ID)
	require.NoError(t, err)
	require.Len(t, members, 1)

	teams, err := env.repo.ListTeams(env.ctx)
	require.NoError(t, err)
	require.Len(t, teams, 1)

	found, err := env.repo.FindTeamByID(env.ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, team.ID, found.ID)

	proj, err := env.repo.Create(env.ctx, "TeamMap", "TMAP", "", env.owner.ID)
	require.NoError(t, err)
	require.NoError(t, env.repo.AddTeamToProject(env.ctx, team.ID, proj.ID))

	require.NoError(t, env.repo.RemoveTeamMember(env.ctx, team.ID, env.owner.ID))
	count, err = env.repo.TeamMemberCount(env.ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestEntRepository_DeleteHard_RequiresNoMembers проверяет слой репозитория: ent репозиторий удаление hard requires no members.
func TestEntRepository_DeleteHard_RequiresNoMembers(t *testing.T) {
	env := newProjectRepoEnv(t)

	created, err := env.repo.Create(env.ctx, "DeleteMe", "DEL", "", env.owner.ID)
	require.NoError(t, err)

	err = env.repo.Delete(env.ctx, created.ID)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err))
}
