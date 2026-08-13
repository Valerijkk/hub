package project_test

// Тесты сервиса проектов (волна 120): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/projectmember"
)

// TestService_Wave120_ListProjectsMemberCount проверяет бизнес-логику: сервис wave 120 список projects участник count.
func TestService_Wave120_ListProjectsMemberCount(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "Members", "MEM", "", env.owner.ID)
	require.NoError(t, err)
	_, err = env.repo.AddMember(env.ctx, created.ID, env.member.ID, projectmember.ProjectRoleDeveloper)
	require.NoError(t, err)

	list, err := svc.ListProjects(env.ctx, env.owner.ID, false)
	require.NoError(t, err)
	require.NotEmpty(t, list)

	var found *project.ProjectResponse
	for i := range list {
		if list[i].ID == created.ID.String() {
			found = &list[i]
			break
		}
	}
	require.NotNil(t, found)
	assert.GreaterOrEqual(t, found.MemberCount, 2)
}

// TestService_Wave120_UnarchiveAfterArchive проверяет бизнес-логику: сервис wave 120 unarchive after archive.
func TestService_Wave120_UnarchiveAfterArchive(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	created, err := env.repo.Create(env.ctx, "Archive Flow", "ARCF", "", env.owner.ID)
	require.NoError(t, err)

	_, err = svc.ArchiveProject(env.ctx, created.ID)
	require.NoError(t, err)

	restored, err := svc.UnarchiveProject(env.ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "active", restored.Status)
}

// TestService_Wave120_DeleteProject_NotFound проверяет бизнес-логику: сервис wave 120 удаление project not found.
func TestService_Wave120_DeleteProject_NotFound(t *testing.T) {
	env := newProjectRepoEnv(t)
	svc := project.NewProjectService(env.repo, nil, nil)

	err := svc.DeleteProject(env.ctx, uuid.New())
	require.Error(t, err)
}
