package tms_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/tms"
	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"context"
)

func newTMSSoftDeleteEnv(t *testing.T) (*tms.EntRepository, *ent.Client, context.Context, uuid.UUID) {
	t.Helper()
	dsn := "file:tmssoft_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("soft-" + uuid.New().String()[:8] + "@test.com").
		SetName("Soft Delete").
		Save(ctx)
	require.NoError(t, err)
	_ = user

	proj, err := client.Project.Create().
		SetName("Soft Project").
		SetKey("SFT").
		Save(ctx)
	require.NoError(t, err)

	_, err = client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	return tms.NewEntRepository(client), client, ctx, proj.ID
}

// TestSoftDeleteSuite_KeepsRowWithDeletedAt — W2-A3-008 / P2-BE-08:
// DeleteSuite → deleted_at; Get/List скрывают; row остаётся (hard Delete не используется).
func TestSoftDeleteSuite_KeepsRowWithDeletedAt(t *testing.T) {
	repo, client, ctx, projectID := newTMSSoftDeleteEnv(t)

	suite, err := repo.CreateSuite(ctx, projectID, tms.CreateSuiteInput{Name: "TrashMe"})
	require.NoError(t, err)

	require.NoError(t, repo.DeleteSuite(ctx, projectID, suite.ID))

	_, err = repo.GetSuite(ctx, suite.ID)
	require.True(t, ent.IsNotFound(err))

	list, err := repo.ListSuites(ctx, projectID)
	require.NoError(t, err)
	for _, s := range list {
		assert.NotEqual(t, suite.ID, s.ID)
	}

	raw, err := client.TestSuite.Get(ctx, suite.ID)
	require.NoError(t, err)
	require.NotNil(t, raw.DeletedAt, "row must remain with deleted_at set")
}

func TestSoftDeleteComponent_KeepsRowWithDeletedAt(t *testing.T) {
	repo, client, ctx, projectID := newTMSSoftDeleteEnv(t)

	comp, err := repo.CreateComponent(ctx, projectID, "CompTrash", "")
	require.NoError(t, err)
	require.NoError(t, repo.DeleteComponent(ctx, projectID, comp.ID))

	list, err := repo.ListComponents(ctx, projectID)
	require.NoError(t, err)
	for _, c := range list {
		assert.NotEqual(t, comp.ID, c.ID)
	}

	raw, err := client.Component.Get(ctx, comp.ID)
	require.NoError(t, err)
	require.NotNil(t, raw.DeletedAt)
}

func TestSoftDeleteEnvironment_KeepsRowWithDeletedAt(t *testing.T) {
	repo, client, ctx, projectID := newTMSSoftDeleteEnv(t)

	env, err := repo.CreateEnvironment(ctx, projectID, tms.CreateEnvironmentInput{
		Name:  "EnvTrash",
		Label: "dev",
	})
	require.NoError(t, err)
	require.NoError(t, repo.DeleteEnvironment(ctx, projectID, env.ID))

	_, err = repo.GetEnvironment(ctx, env.ID)
	require.True(t, ent.IsNotFound(err))

	raw, err := client.Environment.Get(ctx, env.ID)
	require.NoError(t, err)
	require.NotNil(t, raw.DeletedAt)
}
