package sprint_test

// P1-BE-05 / TASK-014: CompleteSprintAtomic — all-or-nothing при mid-fail после move.
import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/internal/sprint"
)

// failStatusAfterMoveRepo: move OK, UpdateSprintStatus fails — CompleteSprintAtomic должен
// откатить move (mock: не оставлять movedIssueIDs при ошибке status).
type failStatusAfterMoveRepo struct {
	*mockRepository
	statusErr error
}

func (r *failStatusAfterMoveRepo) CompleteSprintAtomic(ctx context.Context, projectID, sprintID uuid.UUID, afterSprint *ent.Sprint, incompleteIDs []uuid.UUID) (*ent.Sprint, error) {
	// Имитация: move выполнен, status падает → tx rollback (не фиксируем move).
	if len(incompleteIDs) > 0 {
		if _, err := r.FindOrCreateNextPlanningSprint(ctx, projectID, afterSprint); err != nil {
			return nil, err
		}
		// Не пишем movedIssueIDs — rollback.
	}
	if r.statusErr != nil {
		return nil, r.statusErr
	}
	return r.mockRepository.UpdateSprintStatus(ctx, sprintID, sprintEnt.StatusCompleted)
}

func TestCompleteSprint_Atomic_FailAfterMove_NoPartialState(t *testing.T) {
	base := newMockRepository()
	repo := &failStatusAfterMoveRepo{mockRepository: base, statusErr: errors.New("status update failed")}
	svc := sprint.NewSprintService(repo, nil, nil)
	ctx := context.Background()
	projectID := uuid.New()
	sprintID := uuid.New()

	repo.sprints[sprintID] = &ent.Sprint{
		ID: sprintID, Name: "Active", Status: sprintEnt.StatusActive,
		StartDate: time.Now().AddDate(0, 0, -14), EndDate: time.Now(),
	}
	repo.sprintProject[sprintID] = projectID

	issueID := uuid.New()
	repo.issues[issueID] = &ent.Issue{
		ID: issueID,
		Edges: ent.IssueEdges{Sprint: repo.sprints[sprintID]},
	}

	_, err := svc.CompleteSprint(ctx, projectID, sprintID)
	require.Error(t, err)
	assert.Empty(t, repo.movedIssueIDs, "при fail mid CompleteSprint issues не должны быть зафиксированы в next sprint")
	assert.Equal(t, sprintEnt.StatusActive, repo.sprints[sprintID].Status, "status должен остаться active при rollback")
}

func TestCompleteSprintAtomic_Ent_MovesAndCompletes(t *testing.T) {
	env := newSprintRepoEnv(t)
	start := time.Now().UTC().Truncate(time.Second)
	end := start.AddDate(0, 0, 14)

	sp, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Atomic Sprint", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)
	_, err = env.repo.UpdateSprintStatus(env.ctx, sp.ID, sprintEnt.StatusActive)
	require.NoError(t, err)
	sp, err = env.repo.GetSprint(env.ctx, sp.ID)
	require.NoError(t, err)

	iss := env.createIssue(t, 1, env.todoID, &sp.ID)

	updated, err := env.repo.CompleteSprintAtomic(env.ctx, env.projectID, sp.ID, sp, []uuid.UUID{iss.ID})
	require.NoError(t, err)
	assert.Equal(t, sprintEnt.StatusCompleted, updated.Status)

	oldCount, err := env.repo.GetSprintIssueCount(env.ctx, sp.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, oldCount)

	next, err := env.repo.FindOrCreateNextPlanningSprint(env.ctx, env.projectID, sp)
	require.NoError(t, err)
	nextCount, err := env.repo.GetSprintIssueCount(env.ctx, next.ID)
	require.NoError(t, err)
	assert.Equal(t, 1, nextCount)
}
