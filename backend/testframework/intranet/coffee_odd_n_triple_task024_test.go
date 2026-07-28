package intranet_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/intranet"
)

// TestBuildCoffeeMatches_OddNTriple_P1_BE_43 — N=3 → 0 unpaired / 1 triple.
// Mutation: unpaired на нечётном N → TripleCount!=1 или Unpaired!=0 FAIL.
func TestBuildCoffeeMatches_OddNTriple_P1_BE_43(t *testing.T) {
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New()}
	pairs, triples := intranet.ExportBuildCoffeeMatches(ids, nil)
	assert.Empty(t, pairs)
	require.Len(t, triples, 1)
	assert.Len(t, triples[0], 3)
}

func TestBuildCoffeeMatches_EvenNPairs_P1_BE_43(t *testing.T) {
	ids := []uuid.UUID{uuid.New(), uuid.New(), uuid.New(), uuid.New()}
	pairs, triples := intranet.ExportBuildCoffeeMatches(ids, nil)
	assert.Len(t, pairs, 2)
	assert.Empty(t, triples)
}

// TestRunCoffeeMatching_OddN_NoUnpaired_P1_BE_43 — интеграция: 3 enroll → 1 triple, 0 unpaired.
func TestRunCoffeeMatching_OddN_NoUnpaired_P1_BE_43(t *testing.T) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:coffee_odd_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("coffee-admin-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	var memberIDs []uuid.UUID
	for i := 0; i < 2; i++ {
		m, err := client.User.Create().
			SetEmail("coffee-m-" + uuid.New().String()[:8] + "@hub.local").
			SetSystemRole(user.SystemRoleMember).
			SetStatus(user.StatusActive).
			Save(ctx)
		require.NoError(t, err)
		memberIDs = append(memberIDs, m.ID)
	}

	svc := intranet.NewService(intranet.NewEntRepository(client), nil, &stubCoffeeNotifier{})
	_, err = svc.SetCoffeeEnrollment(ctx, admin.ID, true)
	require.NoError(t, err)
	for _, id := range memberIDs {
		_, err = svc.SetCoffeeEnrollment(ctx, id, true)
		require.NoError(t, err)
	}

	round, err := svc.RunCoffeeMatching(ctx, admin.ID)
	require.NoError(t, err)
	assert.Equal(t, 0, round.UnpairedCount)
	assert.Equal(t, 1, round.TripleCount)
	assert.Equal(t, 0, round.PairCount)
	assert.Empty(t, round.Pairs) // A2W7-2-N01: рёбра тройки не в Pairs[]
	assert.Equal(t, round.PairCount, len(round.Pairs))
	require.Len(t, round.Triples, 1)
}
