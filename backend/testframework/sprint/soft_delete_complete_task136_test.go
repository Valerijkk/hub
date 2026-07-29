//go:build hubtest

package sprint_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	sprintEnt "github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/internal/sprint"
)

// TestCompleteSprint_SkipsSoftDeletedIncomplete_A3W1A4001 — trash unfinished не уезжает в next planning.
// Mutation: убрать DeletedAtIsNil из FindIncompleteIssues (и/или move Where) → trash.sprint_id = next → RED.
func TestCompleteSprint_SkipsSoftDeletedIncomplete_A3W1A4001(t *testing.T) {
	env := newSprintRepoEnv(t)
	start := time.Now().UTC()
	end := start.AddDate(0, 0, 14)

	sp, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Active", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)
	_, err = env.repo.UpdateSprintStatus(env.ctx, sp.ID, sprintEnt.StatusActive)
	require.NoError(t, err)

	alive := env.createIssue(t, 1, env.todoID, &sp.ID)
	trash := env.createIssue(t, 2, env.todoID, &sp.ID)
	_, err = env.client.Issue.UpdateOneID(trash.ID).
		SetDeletedAt(time.Now().UTC()).
		Save(env.ctx)
	require.NoError(t, err)

	incomplete, err := env.repo.FindIncompleteIssues(env.ctx, sp.ID)
	require.NoError(t, err)
	require.Len(t, incomplete, 1)
	assert.Equal(t, alive.ID, incomplete[0].ID)

	svc := sprint.NewSprintService(env.repo, nil, nil)
	resp, err := svc.CompleteSprint(env.ctx, env.projectID, sp.ID)
	require.NoError(t, err)
	assert.Equal(t, string(sprintEnt.StatusCompleted), resp.Status)

	nextList, err := env.repo.ListSprints(env.ctx, env.projectID)
	require.NoError(t, err)
	var next *ent.Sprint
	for _, s := range nextList {
		if s.Status == sprintEnt.StatusPlanning {
			next = s
			break
		}
	}
	require.NotNil(t, next, "ожидается next planning после Complete с incomplete")

	aliveAfter, err := env.client.Issue.Query().Where(issue.IDEQ(alive.ID)).WithSprint().Only(env.ctx)
	require.NoError(t, err)
	require.NotNil(t, aliveAfter.Edges.Sprint)
	assert.Equal(t, next.ID, aliveAfter.Edges.Sprint.ID)

	trashAfter, err := env.client.Issue.Query().Where(issue.IDEQ(trash.ID)).WithSprint().Only(env.ctx)
	require.NoError(t, err)
	require.NotNil(t, trashAfter.DeletedAt)
	require.NotNil(t, trashAfter.Edges.Sprint)
	assert.Equal(t, sp.ID, trashAfter.Edges.Sprint.ID, "A3W1-A4-001: trash остаётся на completed sprint, не next")
}
