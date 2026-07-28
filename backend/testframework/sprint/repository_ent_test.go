package sprint_test
import (
	"github.com/valerii/hub/backend/internal/sprint"
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

type sprintRepoEnv struct {
	repo      *sprint.EntRepository
	ctx       context.Context
	client    *ent.Client
	projectID uuid.UUID
	userID    uuid.UUID
	todoID    uuid.UUID
	doneID    uuid.UUID
}

func newSprintRepoEnv(t *testing.T) sprintRepoEnv {
	t.Helper()
	dsn := "file:sprintrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("sprint-" + uuid.New().String()[:8] + "@test.com").
		SetName("Sprint User").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Sprint Project").
		SetKey("SPR").
		Save(ctx)
	require.NoError(t, err)

	todo, err := client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	done, err := client.WorkflowStatus.Create().
		SetName("Done").
		SetCategory(workflowstatus.CategoryDone).
		SetColor("#10B981").
		SetSortOrder(99).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	return sprintRepoEnv{
		repo:      sprint.NewEntRepository(client),
		ctx:       ctx,
		client:    client,
		projectID: proj.ID,
		userID:    user.ID,
		todoID:    todo.ID,
		doneID:    done.ID,
	}
}

func (e sprintRepoEnv) createIssue(t *testing.T, num int, statusID uuid.UUID, sprintID *uuid.UUID) *ent.Issue {
	t.Helper()
	b := e.client.Issue.Create().
		SetIssueNumber(num).
		SetIssueKey(fmt.Sprintf("SPR-%d", num)).
		SetTitle("Issue").
		SetIssueType(issue.IssueTypeTask).
		SetProjectID(e.projectID).
		SetReporterID(e.userID).
		SetStatusID(statusID)
	if sprintID != nil {
		b.SetSprintID(*sprintID)
	}
	iss, err := b.Save(e.ctx)
	require.NoError(t, err)
	return iss
}

func TestEntRepository_SprintCRUD(t *testing.T) {
	env := newSprintRepoEnv(t)
	start := time.Now().UTC().Truncate(time.Second)
	end := start.AddDate(0, 0, 14)

	created, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Sprint 1", Goal: "Ship", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)
	assert.Equal(t, "Sprint 1", created.Name)
	assert.Nil(t, sprint.ExportNilIfEmpty(""))
	assert.NotNil(t, sprint.ExportNilIfEmpty("goal"))

	got, err := env.repo.GetSprint(env.ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)

	list, err := env.repo.ListSprints(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	newName := "Renamed"
	updated, err := env.repo.UpdateSprint(env.ctx, created.ID, sprint.UpdateSprintRequest{Name: &newName})
	require.NoError(t, err)
	assert.Equal(t, "Renamed", updated.Name)

	active, err := env.repo.UpdateSprintStatus(env.ctx, created.ID, sprintEnt.StatusActive)
	require.NoError(t, err)
	assert.Equal(t, sprintEnt.StatusActive, active.Status)

	found, err := env.repo.FindActiveSprint(env.ctx, env.projectID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)

	notes := map[string]interface{}{"went_well": []string{"tests"}}
	withNotes, err := env.repo.UpdateRetroNotes(env.ctx, created.ID, notes)
	require.NoError(t, err)
	require.NotNil(t, withNotes.RetroNotes)
	assert.NotEmpty(t, withNotes.RetroNotes["went_well"])
}

func TestEntRepository_FindOrCreateNextPlanningSprint(t *testing.T) {
	env := newSprintRepoEnv(t)
	start := time.Now().UTC()
	end := start.AddDate(0, 0, 14)

	first, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Active sprint", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)
	_, err = env.repo.UpdateSprintStatus(env.ctx, first.ID, sprintEnt.StatusActive)
	require.NoError(t, err)

	next, err := env.repo.FindOrCreateNextPlanningSprint(env.ctx, env.projectID, first)
	require.NoError(t, err)
	assert.Equal(t, sprintEnt.StatusPlanning, next.Status)
	assert.Contains(t, next.Name, "следующий")

	again, err := env.repo.FindOrCreateNextPlanningSprint(env.ctx, env.projectID, first)
	require.NoError(t, err)
	assert.Equal(t, next.ID, again.ID)
}

func TestEntRepository_IssueSprintOperations(t *testing.T) {
	env := newSprintRepoEnv(t)
	start := time.Now().UTC()
	end := start.AddDate(0, 0, 7)

	sp, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Issues sprint", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)

	open := env.createIssue(t, 1, env.todoID, nil)
	done := env.createIssue(t, 2, env.doneID, nil)

	require.NoError(t, env.repo.SetIssuesSprint(env.ctx, env.projectID, []uuid.UUID{open.ID, done.ID}, sp.ID))

	count, err := env.repo.GetSprintIssueCount(env.ctx, sp.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, count)

	total, doneCount, err := env.repo.GetBurndownData(env.ctx, sp.ID)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, 1, doneCount)

	incomplete, err := env.repo.FindIncompleteIssues(env.ctx, sp.ID)
	require.NoError(t, err)
	require.Len(t, incomplete, 1)
	assert.Equal(t, open.ID, incomplete[0].ID)

	sprint2, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Target", StartDate: end, EndDate: end.AddDate(0, 0, 7),
	})
	require.NoError(t, err)
	require.NoError(t, env.repo.MoveIssuesToSprint(env.ctx, []uuid.UUID{open.ID}, sprint2.ID))

	require.NoError(t, env.repo.ClearIssuesSprint(env.ctx, env.projectID, []uuid.UUID{done.ID}))
	count, err = env.repo.GetSprintIssueCount(env.ctx, sp.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestEntRepository_VelocityAndStatusCounts(t *testing.T) {
	env := newSprintRepoEnv(t)
	start := time.Now().UTC().AddDate(0, -1, 0)
	end := start.AddDate(0, 0, 14)

	completed, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Done sprint", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)
	_, err = env.repo.UpdateSprintStatus(env.ctx, completed.ID, sprintEnt.StatusCompleted)
	require.NoError(t, err)

	env.createIssue(t, 1, env.doneID, &completed.ID)
	env.createIssue(t, 2, env.todoID, &completed.ID)

	points, err := env.repo.GetVelocityData(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, points, 1)
	assert.Equal(t, "Done sprint", points[0].SprintName)
	assert.Equal(t, 2, points[0].Committed)
	assert.Equal(t, 1, points[0].Completed)

	n, err := env.repo.CountIssuesInStatus(env.ctx, env.doneID)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	st, err := env.repo.GetWorkflowStatus(env.ctx, env.todoID)
	require.NoError(t, err)
	assert.Equal(t, workflowstatus.CategoryTodo, st.Category)
}
