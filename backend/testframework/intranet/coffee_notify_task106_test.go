package intranet_test

import (
	"context"
	"sync"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/intranet"
)

// recordingCoffeeNotifier — TASK-106: считает NotifyCoffeeMatch (mutation target).
type recordingCoffeeNotifier struct {
	mu    sync.Mutex
	calls []coffeeNotifyCall
}

type coffeeNotifyCall struct {
	UserID    uuid.UUID
	PartnerID uuid.UUID
}

func (s *recordingCoffeeNotifier) NotifyCoffeeMatch(_ context.Context, userID uuid.UUID, partner intranet.CoffeePartnerResponse) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, coffeeNotifyCall{UserID: userID, PartnerID: partner.UserID})
	return nil
}

func (s *recordingCoffeeNotifier) snapshot() []coffeeNotifyCall {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]coffeeNotifyCall, len(s.calls))
	copy(out, s.calls)
	return out
}

// TestCoffeeMatch_NotifyOnly_Task106 — FR-INTR-11 / ADR-INTRANET-U05:
// канон notify-only; пара → ровно 2 notify (по одному каждому участнику); без auto-DM.
// Mutation: убрать s.notifyCoffeeMatch в RunCoffeeMatching → len(calls)!=2 FAIL.
func TestCoffeeMatch_NotifyOnly_Task106(t *testing.T) {
	assert.Equal(t, "notify_only", intranet.CoffeeMatchDeliveryMode)

	client := enttest.Open(t, "sqlite3", "file:coffee_notify_t106_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("coffee-admin-t106-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		SetDisplayName("Admin T106").
		Save(ctx)
	require.NoError(t, err)

	member, err := client.User.Create().
		SetEmail("coffee-member-t106-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		SetDisplayName("Member T106").
		Save(ctx)
	require.NoError(t, err)

	notifier := &recordingCoffeeNotifier{}
	svc := intranet.NewService(intranet.NewEntRepository(client), nil, notifier)

	_, err = svc.SetCoffeeEnrollment(ctx, admin.ID, true)
	require.NoError(t, err)
	_, err = svc.SetCoffeeEnrollment(ctx, member.ID, true)
	require.NoError(t, err)

	round, err := svc.RunCoffeeMatching(ctx, admin.ID)
	require.NoError(t, err)
	require.Equal(t, 1, round.PairCount)
	require.Equal(t, 0, round.TripleCount)

	calls := notifier.snapshot()
	require.Len(t, calls, 2, "пара: ровно один notify каждому участнику (ADR-INTRANET-U05)")

	got := map[uuid.UUID]uuid.UUID{}
	for _, c := range calls {
		got[c.UserID] = c.PartnerID
	}
	assert.Equal(t, member.ID, got[admin.ID])
	assert.Equal(t, admin.ID, got[member.ID])
}

// TestCoffeeMatch_NotifyOnly_Triple_Task106 — тройка: 6 notify (каждый о двух партнёрах).
// Mutation: убрать notify в ветке triples → len(calls)!=6 FAIL.
func TestCoffeeMatch_NotifyOnly_Triple_Task106(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:coffee_triple_t106_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("coffee-admin-tri-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	ids := make([]uuid.UUID, 0, 3)
	ids = append(ids, admin.ID)
	for i := 0; i < 2; i++ {
		u, err := client.User.Create().
			SetEmail("coffee-m-t106-" + uuid.New().String()[:8] + "@hub.local").
			SetSystemRole(user.SystemRoleMember).
			SetStatus(user.StatusActive).
			Save(ctx)
		require.NoError(t, err)
		ids = append(ids, u.ID)
	}

	notifier := &recordingCoffeeNotifier{}
	svc := intranet.NewService(intranet.NewEntRepository(client), nil, notifier)
	for _, id := range ids {
		_, err := svc.SetCoffeeEnrollment(ctx, id, true)
		require.NoError(t, err)
	}

	round, err := svc.RunCoffeeMatching(ctx, admin.ID)
	require.NoError(t, err)
	require.Equal(t, 0, round.PairCount)
	require.Equal(t, 1, round.TripleCount)

	calls := notifier.snapshot()
	require.Len(t, calls, 6, "тройка: 3 участника × 2 партнёра = 6 notify")
}
