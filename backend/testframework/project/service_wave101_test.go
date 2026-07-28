package project_test

// Тесты сервиса проектов (волна 101): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/project"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"

	"github.com/valerii/hub/backend/ent"
	entproject "github.com/valerii/hub/backend/ent/project"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

type crossLinkRepo struct {
	mockRepo
	outgoing []*ent.IssueLink
	incoming []*ent.IssueLink
	deleted  []uuid.UUID
	teamErr  error
}

func (r *crossLinkRepo) ListOutgoingCrossProjectLinks(_ context.Context, _ uuid.UUID) ([]*ent.IssueLink, error) {
	return r.outgoing, nil
}

func (r *crossLinkRepo) ListIncomingCrossProjectLinks(_ context.Context, _ uuid.UUID) ([]*ent.IssueLink, error) {
	return r.incoming, nil
}

func (r *crossLinkRepo) DeleteIssueLink(_ context.Context, linkID uuid.UUID) error {
	r.deleted = append(r.deleted, linkID)
	return nil
}

func (r *crossLinkRepo) Delete(_ context.Context, id uuid.UUID) error {
	if _, ok := r.projects[id]; !ok {
		return notFoundErr()
	}
	delete(r.projects, id)
	return nil
}

func (r *crossLinkRepo) AddTeamToProject(_ context.Context, _, _ uuid.UUID) error {
	return r.teamErr
}

func startProjectNATSServer(t *testing.T) *natsclient.Client {
	t.Helper()
	opts := &server.Options{Port: -1}
	s := natsserver.RunServer(opts)
	t.Cleanup(s.Shutdown)
	client, err := natsclient.NewClient(context.Background(), s.ClientURL())
	require.NoError(t, err)
	t.Cleanup(client.Close)
	return client
}

// TestPermanentDelete_Success проверяет сценарий: permanent удаление успех.
func TestPermanentDelete_Success(t *testing.T) {
	svc, repo, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Gone", "GONE", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	require.NoError(t, svc.PermanentDelete(ctx, projectID))
	_, ok := repo.projects[projectID]
	assert.False(t, ok)
}

// TestPermanentDelete_NotFound проверяет сценарий: permanent удаление not found.
func TestPermanentDelete_NotFound(t *testing.T) {
	repo := &crossLinkRepo{mockRepo: *newMockRepo()}
	svc := project.NewProjectService(repo, &mockCache{}, nil)
	err := svc.PermanentDelete(context.Background(), uuid.New())
	require.Error(t, err)
}

// TestDeleteProject_CleansCrossProjectLinks проверяет сценарий: удаление project cleans cross project links.
func TestDeleteProject_CleansCrossProjectLinks(t *testing.T) {
	repo := &crossLinkRepo{mockRepo: *newMockRepo()}
	nats := startProjectNATSServer(t)
	svc := project.NewProjectService(repo, &mockCache{}, nats)
	ctx := context.Background()

	projectID := uuid.New()
	assigneeID := uuid.New()
	targetIssueID := uuid.New()
	sourceIssueID := uuid.New()
	linkID := uuid.New()

	repo.projects[projectID] = &ent.Project{
		ID: projectID, Name: "Deleted Proj", Key: "DEL", Status: entproject.StatusActive,
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.outgoing = []*ent.IssueLink{{
		ID: linkID,
		Edges: ent.IssueLinkEdges{
			Source: &ent.Issue{ID: sourceIssueID, IssueKey: "DEL-1"},
			Target: &ent.Issue{
				ID: targetIssueID, IssueKey: "OTH-2",
				Edges: ent.IssueEdges{Assignee: &ent.User{ID: assigneeID}},
			},
		},
	}}

	require.NoError(t, svc.DeleteProject(ctx, projectID))
	assert.Contains(t, repo.deleted, linkID)
}

// TestAddTeamToProject_NotFound проверяет сценарий: add team to project not found.
func TestAddTeamToProject_NotFound(t *testing.T) {
	repo := &crossLinkRepo{mockRepo: *newMockRepo(), teamErr: notFoundErr()}
	svc := project.NewProjectService(repo, &mockCache{}, nil)

	err := svc.AddTeamToProject(context.Background(), uuid.New(), uuid.New(), uuid.New())
	require.Error(t, err)
}

// TestAddTeamToProject_Success проверяет сценарий: add team to project успех.
func TestAddTeamToProject_Success(t *testing.T) {
	svc, _, _ := newTestService()
	ctx := context.Background()
	ownerID := uuid.New()

	created, err := svc.CreateProject(ctx, ownerID, "Team Link", "TLK", "")
	require.NoError(t, err)
	projectID, err := uuid.Parse(created.ID)
	require.NoError(t, err)

	team, err := svc.CreateTeam(ctx, ownerID, "Squad", "devs")
	require.NoError(t, err)
	teamID, err := uuid.Parse(team.ID)
	require.NoError(t, err)

	require.NoError(t, svc.AddTeamToProject(ctx, teamID, projectID, ownerID))
}

// TestNotifyLinkDeleted_NoNatsNoOp проверяет сценарий: notify link deleted no nats no op.
func TestNotifyLinkDeleted_NoNatsNoOp(t *testing.T) {
	svc, _, _ := newTestService()
	project.ExportProjectServiceNotifyLinkDeleted(svc, context.Background(), &ent.Issue{ID: uuid.New()}, "X-1", "Proj")
}
