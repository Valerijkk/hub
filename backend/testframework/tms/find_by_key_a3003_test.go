package tms_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/internal/tms"
)

// TestEntRepository_FindTestCaseByKey_RejectsForeignProject_A3_003 —
// TC чужого проекта не находится. Mutation: снять HasProjectWith → FAIL.
func TestEntRepository_FindTestCaseByKey_RejectsForeignProject_A3_003(t *testing.T) {
	dsn := "file:tms_a3003_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	repo := tms.NewEntRepository(client)

	user, err := client.User.Create().
		SetEmail("a3003-" + uuid.New().String()[:8] + "@test.com").
		SetName("A3003").
		Save(ctx)
	require.NoError(t, err)

	projA, err := client.Project.Create().SetName("TMS Project A").SetKey("TMA").Save(ctx)
	require.NoError(t, err)
	projB, err := client.Project.Create().SetName("TMS Project B").SetKey("TMB").Save(ctx)
	require.NoError(t, err)

	tcB, err := repo.CreateTestCase(ctx, projB.ID, tms.CreateTestCaseInput{
		Title: "Foreign TC", CaseType: "classic", Priority: "medium", Status: "draft", CreatedByID: user.ID,
	})
	require.NoError(t, err)

	_, err = repo.FindTestCaseByKey(ctx, projA.ID, tcB.CaseKey)
	require.True(t, ent.IsNotFound(err), "foreign TC must not be visible via project A, got %v", err)

	got, err := repo.FindTestCaseByKey(ctx, projB.ID, tcB.CaseKey)
	require.NoError(t, err)
	assert.Equal(t, tcB.ID, got.ID)
}
