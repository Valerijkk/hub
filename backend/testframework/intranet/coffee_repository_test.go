package intranet_test
import (
	"github.com/valerii/hub/backend/internal/intranet"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
)

func TestEntRepository_CoffeeUpsertAndPreviousPairs(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:coffee_repo?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().SetEmail("admin@hub.local").SetSystemRole(user.SystemRoleAdmin).Save(ctx)
	require.NoError(t, err)
	u1, err := client.User.Create().SetEmail("a@hub.local").SetSystemRole(user.SystemRoleMember).SetStatus(user.StatusActive).Save(ctx)
	require.NoError(t, err)
	u2, err := client.User.Create().SetEmail("b@hub.local").SetSystemRole(user.SystemRoleMember).SetStatus(user.StatusActive).Save(ctx)
	require.NoError(t, err)
	u3, err := client.User.Create().SetEmail("c@hub.local").SetSystemRole(user.SystemRoleMember).SetStatus(user.StatusActive).Save(ctx)
	require.NoError(t, err)

	repo := intranet.NewEntRepository(client)

	created, err := repo.UpsertCoffeeEnrollment(ctx, u1.ID, true)
	require.NoError(t, err)
	assert.True(t, created.IsActive)

	updated, err := repo.UpsertCoffeeEnrollment(ctx, u1.ID, false)
	require.NoError(t, err)
	assert.False(t, updated.IsActive)

	_, err = repo.UpsertCoffeeEnrollment(ctx, u2.ID, true)
	require.NoError(t, err)
	_, err = repo.UpsertCoffeeEnrollment(ctx, u3.ID, true)
	require.NoError(t, err)

	active, err := repo.ListActiveCoffeeEnrollments(ctx)
	require.NoError(t, err)
	assert.Len(t, active, 2)

	round1, err := repo.CreateCoffeeRound(ctx, admin.ID, 1, 0)
	require.NoError(t, err)
	_, err = repo.CreateCoffeePair(ctx, round1.ID, u2.ID, u3.ID)
	require.NoError(t, err)

	// A7-006: до CreateRound следующего — previous = latest (round1), не Offset(1).
	keys, err := repo.GetPreviousRoundPairKeys(ctx)
	require.NoError(t, err)
	assert.Contains(t, keys, pairKey(u2.ID, u3.ID))

	round2, err := repo.CreateCoffeeRound(ctx, admin.ID, 1, 0)
	require.NoError(t, err)
	_, err = repo.CreateCoffeePair(ctx, round2.ID, u1.ID, u2.ID)
	require.NoError(t, err)

	keys, err = repo.GetPreviousRoundPairKeys(ctx)
	require.NoError(t, err)
	assert.Contains(t, keys, pairKey(u1.ID, u2.ID), "latest round pairs")
	assert.NotContains(t, keys, pairKey(u2.ID, u3.ID), "must not skip to round-before-last")

	rounds, err := repo.ListCoffeeRounds(ctx, 0)
	require.NoError(t, err)
	assert.Len(t, rounds, 2)

	pair, err := repo.FindUserPairsInRound(ctx, round2.ID, u2.ID)
	require.NoError(t, err)
	require.Len(t, pair, 1)
	assert.NotNil(t, pair[0])
}
