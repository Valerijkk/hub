package project_test

// Тесты сервиса проектов (волна 122): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/projectmember"
)

// Фаза 122 — service team/trash coverage..

func TestService_Wave122_UpdateMemberRole(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "RoleUp", "RLUP", "", env.owner.ID)
	require.NoError(t, err)
	_, err = env.repo.AddMember(env.ctx, created.ID, env.member.ID, projectmember.ProjectRoleDeveloper)
	require.NoError(t, err)

	updated, err := svc.UpdateMemberRole(env.ctx, created.ID, env.member.ID, "tester")
	require.NoError(t, err)
	assert.Equal(t, "tester", updated.Role)
}

// TestService_Wave122_ListTrash_AfterSoftDelete проверяет бизнес-логику: сервис wave 122 список корзина after soft удаление.
func TestService_Wave122_ListTrash_AfterSoftDelete(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "TrashMe", "TRSH", "", env.owner.ID)
	require.NoError(t, err)

	require.NoError(t, svc.DeleteProject(env.ctx, created.ID))

	trash, err := svc.ListTrash(env.ctx, env.owner.ID)
	require.NoError(t, err)
	found := false
	for _, p := range trash {
		if p.ID == created.ID.String() {
			found = true
			break
		}
	}
	assert.True(t, found)
}

// TestService_Wave122_RestoreProject проверяет бизнес-логику: сервис wave 122 restore project.
func TestService_Wave122_RestoreProject(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "Restore", "REST", "", env.owner.ID)
	require.NoError(t, err)

	require.NoError(t, svc.DeleteProject(env.ctx, created.ID))
	restored, err := svc.RestoreProject(env.ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", restored.Status)
}

// TestService_Wave122_GetProject_NotFound проверяет бизнес-логику: сервис wave 122 получение project not found.
func TestService_Wave122_GetProject_NotFound(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	_, err := svc.GetProject(env.ctx, uuid.New())
	require.Error(t, err)
}
