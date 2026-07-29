//go:build hubtest

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

// TestGetMyCoffeeMatch_TriplePartners_Task163 — A3W7-2-003: N=3 → len(Partners)==2 для каждого.
// Mutation: вернуть First/один Partner → len(Partners)!=2 FAIL.
func TestGetMyCoffeeMatch_TriplePartners_Task163(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:coffee_partners_t163_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("coffee-admin-p163-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		SetDisplayName("Admin P163").
		SetDepartment("Eng").
		Save(ctx)
	require.NoError(t, err)

	ids := []uuid.UUID{admin.ID}
	for i := 0; i < 2; i++ {
		u, err := client.User.Create().
			SetEmail("coffee-m-p163-" + uuid.New().String()[:8] + "@hub.local").
			SetSystemRole(user.SystemRoleMember).
			SetStatus(user.StatusActive).
			SetDisplayName("Member P163").
			SetDepartment("QA").
			Save(ctx)
		require.NoError(t, err)
		ids = append(ids, u.ID)
	}

	svc := intranet.NewService(intranet.NewEntRepository(client), nil, &stubCoffeeNotifier{})
	for _, id := range ids {
		_, err := svc.SetCoffeeEnrollment(ctx, id, true)
		require.NoError(t, err)
	}

	round, err := svc.RunCoffeeMatching(ctx, admin.ID)
	require.NoError(t, err)
	require.Equal(t, 1, round.TripleCount)
	require.Equal(t, 0, round.PairCount)

	allIDs := map[uuid.UUID]struct{}{}
	for _, id := range ids {
		allIDs[id] = struct{}{}
	}

	for _, id := range ids {
		match, err := svc.GetMyCoffeeMatch(ctx, id)
		require.NoError(t, err)
		require.NotNil(t, match.RoundID)
		require.Len(t, match.Partners, 2, "triple participant must see both partners")
		require.NotNil(t, match.Partner)
		assert.Equal(t, match.Partners[0].UserID, match.Partner.UserID, "Partner BWC = Partners[0]")

		seen := map[uuid.UUID]struct{}{id: {}}
		for _, p := range match.Partners {
			assert.NotEqual(t, uuid.Nil, p.UserID)
			assert.NotEmpty(t, p.DisplayName)
			_, ok := allIDs[p.UserID]
			assert.True(t, ok, "partner must be round participant")
			_, self := seen[p.UserID]
			assert.False(t, self, "partner must not be self/dup")
			seen[p.UserID] = struct{}{}
		}
		assert.Len(t, seen, 3)
	}
}

// TestGetMyCoffeeMatch_PairOnePartner_Task163 — N=2 → ровно один партнёр.
func TestGetMyCoffeeMatch_PairOnePartner_Task163(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:coffee_pair_p163_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("coffee-admin-pair163-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)
	member, err := client.User.Create().
		SetEmail("coffee-member-pair163-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	svc := intranet.NewService(intranet.NewEntRepository(client), nil, &stubCoffeeNotifier{})
	_, err = svc.SetCoffeeEnrollment(ctx, admin.ID, true)
	require.NoError(t, err)
	_, err = svc.SetCoffeeEnrollment(ctx, member.ID, true)
	require.NoError(t, err)

	_, err = svc.RunCoffeeMatching(ctx, admin.ID)
	require.NoError(t, err)

	match, err := svc.GetMyCoffeeMatch(ctx, member.ID)
	require.NoError(t, err)
	require.Len(t, match.Partners, 1)
	require.NotNil(t, match.Partner)
	assert.Equal(t, admin.ID, match.Partner.UserID)
	assert.Equal(t, admin.ID, match.Partners[0].UserID)
}
