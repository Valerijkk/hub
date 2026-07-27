package imex_test
import (
	"github.com/valerii/hub/backend/internal/imex"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
)

func TestEntRepository_ListTestRunResultsAndSprint(t *testing.T) {
	dsn := "file:imexpdfrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("pdfrepo-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().SetName("PDF").SetKey("PDF").Save(ctx)
	require.NoError(t, err)

	repo := imex.NewEntRepository(client)

	run, err := client.TestRun.Create().
		SetTitle("Repo run").
		SetProjectID(proj.ID).
		SetCreatedByID(user.ID).
		Save(ctx)
	require.NoError(t, err)

	tc, err := client.TestCase.Create().
		SetCaseNumber(1).
		SetCaseKey("PDF-TC-1").
		SetTitle("Case").
		SetProjectID(proj.ID).
		SetCreatedByID(user.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.TestResult.Create().
		SetRunID(run.ID).
		SetTestCaseID(tc.ID).
		Save(ctx)
	require.NoError(t, err)

	gotRun, err := repo.GetTestRunInProject(ctx, proj.ID, run.ID)
	require.NoError(t, err)

	results, err := repo.ListTestRunResults(ctx, gotRun)
	require.NoError(t, err)
	assert.Len(t, results, 1)

	now := time.Now().UTC()
	sprint, err := client.Sprint.Create().
		SetName("Repo sprint").
		SetStartDate(now).
		SetEndDate(now).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	gotSprint, err := repo.GetSprintInProject(ctx, proj.ID, sprint.ID)
	require.NoError(t, err)
	issues, err := repo.ListSprintIssues(ctx, gotSprint)
	require.NoError(t, err)
	assert.Empty(t, issues)

	_, err = repo.GetTestRunInProject(ctx, proj.ID, uuid.New())
	assert.True(t, imex.IsNotFound(err))
}
