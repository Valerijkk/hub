package intranet_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/coffeeround"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/intranet"
)

// TestCreateCoffeeRoundWithPairs_RollbackOnPairFail_W7A2004 — mid-fail пары
// откатывает весь round (нет orphan). Mutation: убрать tx → Count(rounds)>0.
func TestCreateCoffeeRoundWithPairs_RollbackOnPairFail_W7A2004(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:coffee_tx_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	repo := intranet.NewEntRepository(client)

	admin, err := client.User.Create().
		SetEmail("coffee-tx-admin-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	member, err := client.User.Create().
		SetEmail("coffee-tx-m-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	// Второе ребро ссылается на несуществующего user → FK fail после create round.
	ghost := uuid.New()
	_, _, err = repo.CreateCoffeeRoundWithPairs(ctx, admin.ID, 2, 0, [][2]uuid.UUID{
		{admin.ID, member.ID},
		{admin.ID, ghost},
	})
	require.Error(t, err)

	n, err := client.CoffeeRound.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n, "W7-A2-004: orphan coffee_round must not remain after pair fail")

	nPairs, err := client.CoffeePair.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, nPairs)
}

// TestCreateCoffeeRoundWithPairs_AtomicOK_W7A2004 — happy path: round+pairs committed.
func TestCreateCoffeeRoundWithPairs_AtomicOK_W7A2004(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:coffee_tx_ok_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	repo := intranet.NewEntRepository(client)

	admin, err := client.User.Create().
		SetEmail("coffee-ok-a-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	m1, err := client.User.Create().
		SetEmail("coffee-ok-1-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	m2, err := client.User.Create().
		SetEmail("coffee-ok-2-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	rnd, pairs, err := repo.CreateCoffeeRoundWithPairs(ctx, admin.ID, 1, 0, [][2]uuid.UUID{
		{m1.ID, m2.ID},
	})
	require.NoError(t, err)
	require.NotNil(t, rnd)
	require.Len(t, pairs, 1)

	got, err := client.CoffeeRound.Query().Where(coffeeround.IDEQ(rnd.ID)).Only(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, got.PairCount)
}
