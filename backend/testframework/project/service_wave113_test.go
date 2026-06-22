package project_test

// Тесты сервиса проектов (волна 113): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	entproject "github.com/valerii/hub/backend/ent/project"
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

	err := svc.AddTeamMember(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already a member")
}

// TestRemoveTeamMember_Success проверяет сценарий: remove team участник успех.
func TestRemoveTeamMember_Success(t *testing.T) {
	svc, _, _ := newTestService()
	require.NoError(t, svc.RemoveTeamMember(context.Background(), uuid.New(), uuid.New()))
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

// TestDeleteProject_IncomingCrossProjectLinks проверяет сценарий: удаление project incoming cross project links.
func TestDeleteProject_IncomingCrossProjectLinks(t *testing.T) {
	repo := &crossLinkRepo{mockRepo: *newMockRepo()}
	nats := startProjectNATSServer(t)
	svc := project.NewProjectService(repo, &mockCache{}, nats)
	ctx := context.Background()

	projectID := uuid.New()
	assigneeID := uuid.New()
	sourceIssueID := uuid.New()
	targetIssueID := uuid.New()
	linkID := uuid.New()

	repo.projects[projectID] = &ent.Project{
		ID: projectID, Name: "Incoming Del", Key: "IN", Status: entproject.StatusActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.incoming = []*ent.IssueLink{{
		ID: linkID,
		Edges: ent.IssueLinkEdges{
			Source: &ent.Issue{
				ID: sourceIssueID, IssueKey: "SRC-1",
				Edges: ent.IssueEdges{Assignee: &ent.User{ID: assigneeID}},
			},
			Target: &ent.Issue{ID: targetIssueID, IssueKey: "IN-2"},
		},
	}}

	require.NoError(t, svc.DeleteProject(ctx, projectID))
	assert.Contains(t, repo.deleted, linkID)
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
