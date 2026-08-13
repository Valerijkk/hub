//go:build hubtest

package sprint_test

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/internal/sprint"
)

// TestEntRepository_ClearIssuesSprint_RequiresProject_P3BE05 — clear без project scope
// чистил sprint_id чужих задач. Mutation: Where только IDIn → FAIL.
func TestEntRepository_ClearIssuesSprint_RequiresProject_P3BE05(t *testing.T) {
	env := newSprintRepoEnv(t)
	start := time.Now().UTC()
	end := start.AddDate(0, 0, 7)

	sp, err := env.repo.CreateSprint(env.ctx, env.projectID, sprint.CreateSprintRequest{
		Name: "Clear scope", StartDate: start, EndDate: end,
	})
	require.NoError(t, err)

	iss := env.createIssue(t, 1, env.todoID, &sp.ID)

	foreignProj, err := env.client.Project.Create().
		SetName("Foreign Sprint").
		SetKey("FSP").
		Save(env.ctx)
	require.NoError(t, err)

	err = env.repo.ClearIssuesSprint(env.ctx, foreignProj.ID, []uuid.UUID{iss.ID})
	require.ErrorIs(t, err, sprint.ErrIssueNotInProject)

	still, err := env.client.Issue.Query().Where(issue.IDEQ(iss.ID)).WithSprint().Only(env.ctx)
	require.NoError(t, err)
	require.NotNil(t, still.Edges.Sprint)
	assert.Equal(t, sp.ID, still.Edges.Sprint.ID)

	require.NoError(t, env.repo.ClearIssuesSprint(env.ctx, env.projectID, []uuid.UUID{iss.ID}))
	cleared, err := env.client.Issue.Query().Where(issue.IDEQ(iss.ID)).WithSprint().Only(env.ctx)
	require.NoError(t, err)
	assert.Nil(t, cleared.Edges.Sprint)
}
