//go:build hubtest

package intranet_test

import (
	"bytes"
	"context"
	"errors"
	"log/slog"
	"strings"
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

// failingCoffeeNotifier — A3W7-2-001: всегда err на NotifyCoffeeMatch.
type failingCoffeeNotifier struct {
	mu    sync.Mutex
	calls int
}

func (n *failingCoffeeNotifier) NotifyCoffeeMatch(_ context.Context, _ uuid.UUID, _ intranet.CoffeePartnerResponse) error {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.calls++
	return errors.New("notify transport down")
}

func (n *failingCoffeeNotifier) callCount() int {
	n.mu.Lock()
	defer n.mu.Unlock()
	return n.calls
}

// TestCoffeeMatch_NotifyFail_WarnContext_Task163 — ADR-U05 §3 / A3W7-2-001:
// fail NotifyCoffeeMatch → slog.WarnContext (user_id/partner_id/error); раунд не откатывается.
// Mutation: вернуть `_ = s.coffeeNotifier.Notify…` / убрать WarnContext → лог без warn FAIL.
func TestCoffeeMatch_NotifyFail_WarnContext_Task163(t *testing.T) {
	assert.Equal(t, "notify_only", intranet.CoffeeMatchDeliveryMode)

	var buf bytes.Buffer
	prev := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelWarn})))
	t.Cleanup(func() { slog.SetDefault(prev) })

	client := enttest.Open(t, "sqlite3", "file:coffee_notify_warn_t163_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	admin, err := client.User.Create().
		SetEmail("coffee-admin-w163-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		SetStatus(user.StatusActive).
		SetDisplayName("Admin W163").
		Save(ctx)
	require.NoError(t, err)
	member, err := client.User.Create().
		SetEmail("coffee-member-w163-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		SetStatus(user.StatusActive).
		SetDisplayName("Member W163").
		Save(ctx)
	require.NoError(t, err)

	notifier := &failingCoffeeNotifier{}
	svc := intranet.NewService(intranet.NewEntRepository(client), nil, notifier)
	_, err = svc.SetCoffeeEnrollment(ctx, admin.ID, true)
	require.NoError(t, err)
	_, err = svc.SetCoffeeEnrollment(ctx, member.ID, true)
	require.NoError(t, err)

	round, err := svc.RunCoffeeMatching(ctx, admin.ID)
	require.NoError(t, err, "notify fail must not roll back matching")
	require.Equal(t, 1, round.PairCount)
	assert.Equal(t, 2, notifier.callCount())

	logOut := buf.String()
	require.Contains(t, logOut, "intranet: notify coffee match failed")
	require.Contains(t, logOut, "user_id=")
	require.Contains(t, logOut, "partner_id=")
	require.True(t, strings.Contains(logOut, "notify transport down") || strings.Contains(logOut, "error="),
		"warn must include error field, got: %s", logOut)

	match, err := svc.GetMyCoffeeMatch(ctx, member.ID)
	require.NoError(t, err)
	require.NotNil(t, match.Partner)
}
