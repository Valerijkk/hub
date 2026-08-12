package sprint_test
import (
	"github.com/valerii/hub/backend/internal/sprint"
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/ent"
)

func TestGetSprint_NotFound(t *testing.T) {
	svc := sprint.NewSprintService(newMockRepository(), nil, nil)
	_, err := svc.GetSprint(context.Background(), uuid.New(), uuid.New())
	require.Error(t, err)
}

func TestUpdateSprint_Success(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	projectID := uuid.New()

	sprintID := uuid.New()
	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "Sprint 1", Status: sprintEnt.StatusPlanning,
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14),
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.sprintProject[sprintID] = projectID

	newName := "Sprint 1 updated"
	resp, err := svc.UpdateSprint(ctx, projectID, sprintID, sprint.UpdateSprintRequest{Name: &newName, Version: 1})
	require.NoError(t, err)
	assert.Equal(t, "Sprint 1 updated", resp.Name)
}

func TestListSprints(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	projectID := uuid.New()

	sprintID := uuid.New()
	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "S1", Status: sprintEnt.StatusPlanning,
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14),
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}

	list, err := svc.ListSprints(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, list, 1)
}

func TestCompleteSprint_RejectsNonActive(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	projectID := uuid.New()

	sprintID := uuid.New()
	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "Planning sprint", Status: sprintEnt.StatusPlanning,
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14),
	}
	repo.sprintProject[sprintID] = projectID

	_, err := svc.CompleteSprint(ctx, projectID, sprintID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "активный")
}

func TestUpdateRetroNotes(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	projectID := uuid.New()

	sprintID := uuid.New()
	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "Retro", Status: sprintEnt.StatusCompleted,
		StartDate: time.Now().AddDate(0, 0, -14), EndDate: time.Now(),
		Version: 1, CreatedAt: time.Now(), UpdatedAt: time.Now(),
	}
	repo.sprintProject[sprintID] = projectID

	resp, err := svc.UpdateRetroNotes(ctx, projectID, sprintID, sprint.RetroNotesRequest{
		WentWell:  []string{"good deploy"},
		ToImprove: []string{"estimates"},
		Actions:   []string{"refine poker"},
		Version:   1,
	})
	require.NoError(t, err)
	require.NotNil(t, resp.RetroNotes)
	assert.NotEmpty(t, resp.RetroNotes["went_well"])
}

func TestAddAndRemoveIssuesFromSprint(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	projectID := uuid.New()

	sprintID := uuid.New()
	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "S", Status: sprintEnt.StatusPlanning,
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 14),
	}
	repo.sprintProject[sprintID] = projectID
	issueID := uuid.New()
	repo.issueProject[issueID] = projectID // W1-A4-015: mock guard требует project mapping

	require.NoError(t, svc.AddIssuesToSprint(ctx, projectID, sprintID, []uuid.UUID{issueID}))
	assert.Equal(t, sprintID, repo.movedToSprint)
	assert.Equal(t, []uuid.UUID{issueID}, repo.movedIssueIDs)

	require.NoError(t, svc.RemoveIssuesFromSprint(ctx, projectID, []uuid.UUID{issueID}))
	assert.Equal(t, []uuid.UUID{issueID}, repo.clearedIssueIDs)
}

func TestToSprintResponse_WithIssueEdges(t *testing.T) {
	repo := newMockRepository()
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()

	sp := &ent.Sprint{
		ID: uuid.New(), Name: "S", Goal: "G", Status: sprintEnt.StatusActive,
		StartDate: time.Now(), EndDate: time.Now().AddDate(0, 0, 7),
		CreatedAt: time.Now(), UpdatedAt: time.Now(),
		Edges: ent.SprintEdges{
			Issues: []*ent.Issue{{ID: uuid.New()}, {ID: uuid.New()}},
		},
	}

	resp := sprint.ExportSprintServiceToSprintResponse(svc, ctx, sp)
	assert.Equal(t, 2, resp.IssueCount)
}

func newPokerTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	t.Helper()
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	return redis.NewClient(&redis.Options{Addr: mr.Addr()}), mr
}

func TestPokerService_StartCastRevealGet(t *testing.T) {
	rdb, _ := newPokerTestRedis(t)
	poker := sprint.NewPokerService(rdb, nil)
	ctx := context.Background()
	sprintID := uuid.New()
	issueID := uuid.New()
	userID := uuid.New()

	session, err := poker.StartPokerSession(ctx, uuid.New(), sprintID, issueID)
	require.NoError(t, err)
	assert.False(t, session.Revealed)

	require.NoError(t, poker.CastVote(ctx, session.SessionID, userID, "8"))

	err = poker.CastVote(ctx, session.SessionID, userID, "invalid")
	require.Error(t, err)

	revealed, err := poker.RevealVotes(ctx, session.SessionID)
	require.NoError(t, err)
	assert.True(t, revealed.Revealed)
	assert.Equal(t, "8", revealed.Votes[userID.String()])

	err = poker.CastVote(ctx, session.SessionID, userID, "5")
	require.Error(t, err)

	got, err := poker.GetSession(ctx, session.SessionID)
	require.NoError(t, err)
	assert.True(t, got.Revealed)
	assert.Equal(t, "8", got.Votes[userID.String()])
}

func TestPokerService_GetSession_HiddenVotes(t *testing.T) {
	rdb, _ := newPokerTestRedis(t)
	poker := sprint.NewPokerService(rdb, nil)
	ctx := context.Background()

	session, err := poker.StartPokerSession(ctx, uuid.New(), uuid.New(), uuid.New())
	require.NoError(t, err)

	userID := uuid.New()
	require.NoError(t, poker.CastVote(ctx, session.SessionID, userID, "3"))

	got, err := poker.GetSession(ctx, session.SessionID)
	require.NoError(t, err)
	assert.False(t, got.Revealed)
	assert.Contains(t, got.Voters, userID.String())
	assert.Empty(t, got.Votes)
}

func TestPokerService_SessionNotFound(t *testing.T) {
	rdb, _ := newPokerTestRedis(t)
	poker := sprint.NewPokerService(rdb, nil)

	_, err := poker.RevealVotes(context.Background(), uuid.New().String())
	require.Error(t, err)
}
