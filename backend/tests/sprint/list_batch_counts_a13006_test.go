//go:build hubtest

package sprint_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/internal/sprint"
)

// countingCountsRepo wraps mock and counts GetSprintIssueCount vs GetSprintIssueCounts calls.
type countingCountsRepo struct {
	*mockRepository
	single atomic.Int32
	batch  atomic.Int32
	counts map[uuid.UUID]int
}

func (r *countingCountsRepo) GetSprintIssueCount(ctx context.Context, sprintID uuid.UUID) (int, error) {
	r.single.Add(1)
	if r.counts != nil {
		return r.counts[sprintID], nil
	}
	return r.mockRepository.GetSprintIssueCount(ctx, sprintID)
}

func (r *countingCountsRepo) GetSprintIssueCounts(ctx context.Context, sprintIDs []uuid.UUID) (map[uuid.UUID]int, error) {
	r.batch.Add(1)
	out := make(map[uuid.UUID]int, len(sprintIDs))
	for _, id := range sprintIDs {
		if r.counts != nil {
			out[id] = r.counts[id]
			continue
		}
		out[id] = 0
	}
	return out, nil
}

// TestListSprints_BatchIssueCounts_A13_006 — ListSprints использует ≤1 batch Count,
// не N× GetSprintIssueCount. Mutation: вернуть N× single Count → single>0 FAIL.
func TestListSprints_BatchIssueCounts_A13_006(t *testing.T) {
	base := newMockRepository()
	ctx := context.Background()
	projectID := uuid.New()
	now := time.Now().UTC().Truncate(time.Second)

	const n = 5
	counts := make(map[uuid.UUID]int, n)
	for i := 0; i < n; i++ {
		id := uuid.New()
		base.sprints[id] = &ent.Sprint{
			ID: id, Name: "S", Status: sprintEnt.StatusPlanning,
			StartDate: now, EndDate: now.AddDate(0, 0, 14),
			CreatedAt: now, UpdatedAt: now,
		}
		base.sprintProject[id] = projectID
		counts[id] = i + 1
	}

	repo := &countingCountsRepo{mockRepository: base, counts: counts}
	svc := sprint.NewSprintService(repo, nil, nil)

	list, err := svc.ListSprints(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, list, n)

	assert.LessOrEqual(t, int(repo.batch.Load()), 2, "expected ≤1–2 batch count queries")
	assert.GreaterOrEqual(t, int(repo.batch.Load()), 1, "expected at least one batch call")
	assert.Equal(t, int32(0), repo.single.Load(), "ListSprints must not call GetSprintIssueCount per sprint")

	byID := make(map[string]int, n)
	for _, sp := range list {
		byID[sp.ID] = sp.IssueCount
	}
	for id, want := range counts {
		assert.Equal(t, want, byID[id.String()], "issue_count for %s", id)
	}
}

// TestGetSprintIssueCounts_Ent_A13_006 — Ent GROUP BY batch matches per-sprint Count.
func TestGetSprintIssueCounts_Ent_A13_006(t *testing.T) {
	env := newSprintRepoEnv(t)
	start := time.Now().UTC().Truncate(time.Second)
	end := start.AddDate(0, 0, 14)

	sp1, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Batch A", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)
	sp2, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Batch B", StartDate: start.AddDate(0, 0, -14), EndDate: start,
	})
	require.NoError(t, err)
	spEmpty, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Empty", StartDate: start.AddDate(0, 0, -28), EndDate: start.AddDate(0, 0, -14),
	})
	require.NoError(t, err)

	for i := 1; i <= 3; i++ {
		_, err := env.client.Issue.Create().
			SetProjectID(env.projectID).
			SetStatusID(env.todoID).
			SetReporterID(env.userID).
			SetIssueNumber(i).
			SetIssueKey(fmt.Sprintf("SP-A%d", i)).
			SetTitle("A").
			SetIssueType(issue.IssueTypeTask).
			SetPriority(issue.PriorityMedium).
			SetSprintID(sp1.ID).
			Save(env.ctx)
		require.NoError(t, err)
	}
	_, err = env.client.Issue.Create().
		SetProjectID(env.projectID).
		SetStatusID(env.todoID).
		SetReporterID(env.userID).
		SetIssueNumber(4).
		SetIssueKey("SP-B1").
		SetTitle("B").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		SetSprintID(sp2.ID).
		Save(env.ctx)
	require.NoError(t, err)

	counts, err := env.repo.GetSprintIssueCounts(env.ctx, []uuid.UUID{sp1.ID, sp2.ID, spEmpty.ID})
	require.NoError(t, err)
	assert.Equal(t, 3, counts[sp1.ID])
	assert.Equal(t, 1, counts[sp2.ID])
	assert.Equal(t, 0, counts[spEmpty.ID], "empty sprint absent from GROUP BY → 0")

	svc := sprint.NewSprintService(env.repo, nil, nil)
	resp, err := svc.ListSprints(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, resp, 3)
	byName := map[string]int{}
	for _, r := range resp {
		byName[r.Name] = r.IssueCount
	}
	assert.Equal(t, 3, byName["Batch A"])
	assert.Equal(t, 1, byName["Batch B"])
	assert.Equal(t, 0, byName["Empty"])
}
