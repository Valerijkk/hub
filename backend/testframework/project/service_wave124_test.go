package project_test

// Тесты сервиса проектов (волна 124): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/invite"
)

// Фаза 124 — граничные сценарии приглашений + удаление из команды..

func TestService_Wave124_JoinByInvite_Expired(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "Expired", "EXP", "", env.owner.ID)
	require.NoError(t, err)

	expired := time.Now().UTC().Add(-1 * time.Hour)
	inv, err := env.repo.CreateInvite(env.ctx, created.ID, env.owner.ID, "EXPIRED1", invite.ProjectRoleDeveloper, 5, &expired)
	require.NoError(t, err)

	_, err = svc.JoinByInvite(env.ctx, inv.Code, env.member.ID)
	require.Error(t, err)
}

// TestService_Wave124_JoinByInvite_InvalidatesCache проверяет бизнес-логику: сервис wave 124 join by приглашение invalidates кэш.
func TestService_Wave124_JoinByInvite_InvalidatesCache(t *testing.T) {
	env := newProjectRepoEnv(t)
	cache := &mockCache{}
	svc := project.NewProjectService(env.repo, cache, nil)

	created, err := env.repo.Create(env.ctx, "CacheInv", "CIN", "", env.owner.ID)
	require.NoError(t, err)

	inv, err := svc.CreateInvite(env.ctx, created.ID, env.owner.ID, "developer", 3, nil)
	require.NoError(t, err)

	_, err = svc.JoinByInvite(env.ctx, inv.Code, env.member.ID)
	require.NoError(t, err)
	assert.Contains(t, cache.invalidated, created.ID.String())
}

// TestService_Wave124_RemoveTeamMember проверяет бизнес-логику: сервис wave 124 remove team участник.
func TestService_Wave124_RemoveTeamMember(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	proj, err := env.repo.Create(env.ctx, "TeamRm", "TRM", "", env.owner.ID)
	require.NoError(t, err)

	team, err := env.repo.CreateTeam(env.ctx, "Squad", "")
	require.NoError(t, err)
	require.NoError(t, env.repo.AddTeamToProject(env.ctx, team.ID, proj.ID))

	require.NoError(t, svc.AddTeamMember(env.ctx, team.ID, env.member.ID))
	require.NoError(t, svc.RemoveTeamMember(env.ctx, team.ID, env.member.ID))

	count, err := env.repo.TeamMemberCount(env.ctx, team.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestService_Wave124_DeleteProject_NotFound проверяет бизнес-логику: сервис wave 124 удаление project not found.
func TestService_Wave124_DeleteProject_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	err := svc.DeleteProject(context.Background(), uuid.New())
	require.Error(t, err)
}

// TestService_Wave124_JoinByInvite_MaxUsesFastPath проверяет бизнес-логику: сервис wave 124 join by приглашение max uses fast path.
func TestService_Wave124_JoinByInvite_MaxUsesFastPath(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "MaxUse", "MXU", "", env.owner.ID)
	require.NoError(t, err)

	inv, err := env.repo.CreateInvite(env.ctx, created.ID, env.owner.ID, "FULL", invite.ProjectRoleDeveloper, 1, nil)
	require.NoError(t, err)
	require.NoError(t, env.repo.IncrementInviteUseCount(env.ctx, inv.ID))

	_, err = svc.JoinByInvite(env.ctx, inv.Code, env.member.ID)
	require.Error(t, err)
}

// TestService_Wave124_RestoreProject_Ent проверяет слой репозитория: сервис wave 124 restore project ent.
func TestService_Wave124_RestoreProject_Ent(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "Restore", "RST", "", env.owner.ID)
	require.NoError(t, err)

	_, err = env.repo.UpdateStatus(env.ctx, created.ID, "deleted", true)
	require.NoError(t, err)

	restored, err := svc.RestoreProject(env.ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID.String(), restored.ID)
}
