package project_test

// Тесты сервиса проектов (волна 113): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

type teamConflictRepo struct {
	mockRepo
}

func (r *teamConflictRepo) AddTeamMember(_ context.Context, _, _ uuid.UUID) (*ent.TeamMember, error) {
	return nil, constraintErr()
}

// TestAddTeamMember_AlreadyMember проверяет сценарий: add team участник already участник.
func TestAddTeamMember_AlreadyMember(t *testing.T) {
	repo := &teamConflictRepo{mockRepo: *newMockRepo()}
	svc := project.NewProjectService(repo, &mockCache{}, nil)

	teamID := uuid.New()
	actorID := uuid.New()
	repo.teams[teamID] = &ent.Team{ID: teamID}
	repo.teamMembers[teamID] = map[uuid.UUID]struct{}{actorID: {}}

	err := svc.AddTeamMember(context.Background(), teamID, actorID, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already a member")
}

// TestRemoveTeamMember_Success проверяет сценарий: remove team участник успех.
func TestRemoveTeamMember_Success(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	actorID := uuid.New()
	team, err := svc.CreateTeam(ctx, actorID, "T", "")
	require.NoError(t, err)
	teamID, err := uuid.Parse(team.ID)
	require.NoError(t, err)
	memberID := uuid.New()
	require.NoError(t, svc.AddTeamMember(ctx, teamID, actorID, memberID))
	require.NoError(t, svc.RemoveTeamMember(ctx, teamID, actorID, memberID))
	_ = repo
}

// TestUpdateMemberRole_NotFound проверяет сценарий: обновление участник role not found.
func TestUpdateMemberRole_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Role Test", "ROLE", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	_, err = svc.UpdateMemberRole(ctx, projectID, uuid.New(), "tester")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestRemoveMember_NotFound проверяет сценарий: remove участник not found.
func TestRemoveMember_NotFound(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Remove Test", "RMV", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	err = svc.RemoveMember(ctx, projectID, uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestAddMember_AlreadyMember проверяет сценарий: add участник already участник.
func TestAddMember_AlreadyMember(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()
	memberID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Dup Member", "DUP", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	_, err = svc.AddMember(ctx, projectID, memberID, "developer")
	require.NoError(t, err)

	_, err = svc.AddMember(ctx, projectID, memberID, "tester")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already a member")
}
