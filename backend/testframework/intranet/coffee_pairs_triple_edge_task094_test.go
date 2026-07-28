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

// TestStripTripleEdges_NoDuplicateInPairs_A2W7_2_N01 — storage-рёбра тройки
// не остаются в Pairs[]; PairCount == len(Pairs).
// Mutation: вернуть append triple-edge в Pairs → len(Pairs)!=PairCount FAIL.
func TestStripTripleEdges_NoDuplicateInPairs_A2W7_2_N01(t *testing.T) {
	hub := uuid.New()
	b := uuid.New()
	c := uuid.New()
	d := uuid.New()
	e := uuid.New()

	resp := intranet.CoffeeRoundResponse{
		PairCount: 3, // намеренно «грязный» счётчик до нормализации
		Pairs: []intranet.CoffeePairResponse{
			{ID: uuid.New(), ParticipantA: intranet.CoffeePartnerResponse{UserID: d}, ParticipantB: intranet.CoffeePartnerResponse{UserID: e}},
			{ID: uuid.New(), ParticipantA: intranet.CoffeePartnerResponse{UserID: hub}, ParticipantB: intranet.CoffeePartnerResponse{UserID: b}},
			{ID: uuid.New(), ParticipantA: intranet.CoffeePartnerResponse{UserID: hub}, ParticipantB: intranet.CoffeePartnerResponse{UserID: c}},
		},
	}
	intranet.ExportStripTripleEdgesFromRoundResponse(&resp)

	assert.Equal(t, 1, resp.PairCount)
	require.Len(t, resp.Pairs, 1)
	assert.Equal(t, resp.PairCount, len(resp.Pairs))
	assert.Equal(t, 1, resp.TripleCount)
	require.Len(t, resp.Triples, 1)
	ids := map[uuid.UUID]struct{}{
		resp.Triples[0].ParticipantA.UserID: {},
		resp.Triples[0].ParticipantB.UserID: {},
		resp.Triples[0].ParticipantC.UserID: {},
	}
	assert.Contains(t, ids, hub)
	assert.Contains(t, ids, b)
	assert.Contains(t, ids, c)
}

// TestRunCoffeeMatching_PairsExcludeTripleEdges_A2W7_2_N01 — N=3 и N=5:
// Pairs[] без рёбер троек, PairCount согласован с len(Pairs).
func TestRunCoffeeMatching_PairsExcludeTripleEdges_A2W7_2_N01(t *testing.T) {
	t.Run("N3_only_triple", func(t *testing.T) {
		svc, adminID := setupCoffeeEnrolled(t, 3)
		round, err := svc.RunCoffeeMatching(context.Background(), adminID)
		require.NoError(t, err)

		assert.Equal(t, 0, round.UnpairedCount)
		assert.Equal(t, 0, round.PairCount)
		assert.Empty(t, round.Pairs)
		assert.Equal(t, round.PairCount, len(round.Pairs))
		assert.Equal(t, 1, round.TripleCount)
		require.Len(t, round.Triples, 1)

		listed, err := svc.ListCoffeeRounds(context.Background(), adminID)
		require.NoError(t, err)
		require.Len(t, listed, 1)
		assert.Equal(t, 0, listed[0].PairCount)
		assert.Empty(t, listed[0].Pairs)
		assert.Equal(t, listed[0].PairCount, len(listed[0].Pairs))
		assert.Equal(t, 1, listed[0].TripleCount)
		require.Len(t, listed[0].Triples, 1)
	})

	t.Run("N5_one_pair_one_triple", func(t *testing.T) {
		svc, adminID := setupCoffeeEnrolled(t, 5)
		round, err := svc.RunCoffeeMatching(context.Background(), adminID)
		require.NoError(t, err)

		assert.Equal(t, 0, round.UnpairedCount)
		assert.Equal(t, 1, round.PairCount)
		require.Len(t, round.Pairs, 1)
		assert.Equal(t, round.PairCount, len(round.Pairs))
		assert.Equal(t, 1, round.TripleCount)
		require.Len(t, round.Triples, 1)

		listed, err := svc.ListCoffeeRounds(context.Background(), adminID)
		require.NoError(t, err)
		require.Len(t, listed, 1)
		assert.Equal(t, 1, listed[0].PairCount)
		require.Len(t, listed[0].Pairs, 1)
		assert.Equal(t, listed[0].PairCount, len(listed[0].Pairs))
		assert.Equal(t, 1, listed[0].TripleCount)
		require.Len(t, listed[0].Triples, 1)
	})
}

// setupCoffeeEnrolled создаёт admin + (n-1) member, всех enroll в coffee.
func setupCoffeeEnrolled(t *testing.T, n int) (*intranet.Service, uuid.UUID) {
	t.Helper()
	require.GreaterOrEqual(t, n, 2)

	client := enttest.Open(t, "sqlite3", "file:coffee_t094_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("coffee-admin-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	svc := intranet.NewService(intranet.NewEntRepository(client), nil, &stubCoffeeNotifier{})
	_, err = svc.SetCoffeeEnrollment(ctx, admin.ID, true)
	require.NoError(t, err)

	for i := 0; i < n-1; i++ {
		m, err := client.User.Create().
			SetEmail("coffee-m-" + uuid.New().String()[:8] + "@hub.local").
			SetSystemRole(user.SystemRoleMember).
			SetStatus(user.StatusActive).
			Save(ctx)
		require.NoError(t, err)
		_, err = svc.SetCoffeeEnrollment(ctx, m.ID, true)
		require.NoError(t, err)
	}
	return svc, admin.ID
}
