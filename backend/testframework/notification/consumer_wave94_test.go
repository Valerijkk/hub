package notification_test

// Тесты NATS consumer уведомлений (волна 94): интеграционные и unit-сценарии testframework.

import (
	"context"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/notification"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/stretchr/testify/require"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func startJetStreamForConsumer(t *testing.T) *natsclient.Client {
	t.Helper()
	opts := &server.Options{
		JetStream: true,
		StoreDir:  t.TempDir(),
		Port:      -1,
	}
	s := natsserver.RunServer(opts)
	t.Cleanup(s.Shutdown)
	client, err := natsclient.NewClient(context.Background(), s.ClientURL())
	require.NoError(t, err)
	t.Cleanup(client.Close)
	require.NoError(t, client.EnsureStreams(context.Background()))
	return client
}

// TestConsumer_StartStop_WithJetStream проверяет сценарий: consumer start stop with jet stream.
func TestConsumer_StartStop_WithJetStream(t *testing.T) {
	nc := startJetStreamForConsumer(t)
	svc, repo, _, _ := newNotificationTestEnv(t)
	c := notification.NewConsumer(nc, repo, svc)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	require.NoError(t, c.Start(ctx))
	time.Sleep(100 * time.Millisecond)
	c.Stop()
}
