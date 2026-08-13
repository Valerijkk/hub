package tms_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestEntRepository_GetTestRunInProject_RejectsForeign_A3_012 — run чужого проекта
// не находится. Mutation: снять testrun.HasProjectWith в GetTestRunInProject → FAIL.
func TestEntRepository_GetTestRunInProject_RejectsForeign_A3_012(t *testing.T) {
	client := openTMSPostgresEnt(t)
	ctx := context.Background()
	repo := tms.NewEntRepository(client)

	user, err := client.User.Create().
		SetEmail("a3012-" + uuid.New().String()[:8] + "@test.com").
		SetName("A3012").
		Save(ctx)
	require.NoError(t, err)

	projA, err := client.Project.Create().SetName("TMS Run Project A").SetKey("TRA").Save(ctx)
	require.NoError(t, err)
	projB, err := client.Project.Create().SetName("TMS Run Project B").SetKey("TRB").Save(ctx)
	require.NoError(t, err)

	runB, err := repo.CreateTestRun(ctx, projB.ID, tms.CreateTestRunInput{
		Title: "Foreign run", RunType: "manual", CreatedByID: user.ID,
	})
	require.NoError(t, err)

	_, err = repo.GetTestRunInProject(ctx, projA.ID, runB.ID)
	require.True(t, ent.IsNotFound(err), "run проекта B не должен находиться через project A, got %v", err)

	_, err = repo.GetTestRunWithResultsInProject(ctx, projA.ID, runB.ID)
	require.True(t, ent.IsNotFound(err), "detail чужого run → NotFound, got %v", err)

	got, err := repo.GetTestRunInProject(ctx, projB.ID, runB.ID)
	require.NoError(t, err)
	assert.Equal(t, runB.ID, got.ID)

	// Unscoped GetTestRun остаётся для imex/post-create.
	unscoped, err := repo.GetTestRun(ctx, runB.ID)
	require.NoError(t, err)
	assert.Equal(t, runB.ID, unscoped.ID)
}
