package mail_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/mail"

	_ "github.com/mattn/go-sqlite3"
)

func TestAdminService_GetUpdateSettings(t *testing.T) {
	dsn := "file:mail_admin_test?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	defer client.Close()

	store := mail.NewEntSettingsStore(client)
	svc := mail.NewAdminService(store, "http://127.0.0.1:3001", "alerts@hub.local", false)

	ctx := context.Background()
	got, err := svc.GetSettings(ctx)
	require.NoError(t, err)
	assert.Equal(t, "http://127.0.0.1:3001", got.HubMailURL)
	assert.False(t, got.ResendReady)

	enabled := false
	limit := 15
	updated, err := svc.UpdateSettings(ctx, mail.UpdateSettingsRequest{
		HubMailURL:    "http://mail:3001",
		FromAlerts:    "Hub <alerts@test.local>",
		AlertsEnabled: &enabled,
		DailyLimit:    &limit,
	})
	require.NoError(t, err)
	assert.Equal(t, "http://mail:3001", updated.HubMailURL)
	assert.False(t, updated.AlertsEnabled)
	assert.Equal(t, 15, updated.DailyLimit)
}

func TestService_SendAlert_DevMode(t *testing.T) {
	svc := mail.NewService(mail.Config{})
	err := svc.SendAlert(context.Background(), mail.SendRequest{
		To: "user@test.local", Subject: "Test", Template: "run_failed",
	})
	require.NoError(t, err)
}

func TestService_SendAlert_Invalid(t *testing.T) {
	svc := mail.NewService(mail.Config{})
	err := svc.SendAlert(context.Background(), mail.SendRequest{})
	assert.ErrorIs(t, err, mail.ErrInvalidRequest)
}

func TestAdminService_Broadcast(t *testing.T) {
	dsn := "file:mail_broadcast_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	_, err := client.User.Create().SetEmail("a@test.local").SetStatus(user.StatusActive).Save(ctx)
	require.NoError(t, err)
	_, err = client.User.Create().SetEmail("b@test.local").SetStatus(user.StatusActive).Save(ctx)
	require.NoError(t, err)
	// A7-008: pending must not be counted/sent
	_, err = client.User.Create().SetEmail("pending@test.local").SetStatus(user.StatusPending).Save(ctx)
	require.NoError(t, err)

	store := mail.NewEntSettingsStore(client)
	svc := mail.NewAdminService(store, "", "", false)

	t.Run("invalid request", func(t *testing.T) {
		_, err := svc.Broadcast(ctx, client, nil, mail.BroadcastRequest{})
		require.ErrorIs(t, err, mail.ErrInvalidRequest)
	})

	t.Run("dry run", func(t *testing.T) {
		resp, err := svc.Broadcast(ctx, client, nil, mail.BroadcastRequest{
			Subject:  "Hi",
			HTMLBody: "<p>Hello</p>",
			DryRun:   true,
		})
		require.NoError(t, err)
		assert.Equal(t, 2, resp.Recipients, "A7-008: only active users")
		assert.True(t, resp.DryRun)
	})

	sender := &broadcastStub{}
	resp, err := svc.Broadcast(ctx, client, sender, mail.BroadcastRequest{
		Subject:  "Ship",
		HTMLBody: "<p>Go</p>",
	})
	require.NoError(t, err)
	assert.Equal(t, 2, resp.Recipients)
	assert.Equal(t, 2, resp.Sent)
	assert.Equal(t, 2, sender.calls)
}

type broadcastStub struct {
	calls int
}

func (b *broadcastStub) SendAlert(_ context.Context, _ mail.SendRequest) error {
	b.calls++
	return nil
}
