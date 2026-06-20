package audit

import (
	"context"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/stretchr/testify/require"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func startAuditTestNATS(t *testing.T) *natsclient.Client {
	t.Helper()
	opts := &server.Options{
		JetStream: true,
		StoreDir:  t.TempDir(),
		Port:      -1,
	}
	s := natsserver.RunServer(opts)
	t.Cleanup(s.Shutdown)

	nc, err := natsclient.NewClient(context.Background(), s.ClientURL())
	require.NoError(t, err)
	require.NoError(t, nc.EnsureStreams(context.Background()))
	t.Cleanup(nc.Close)
	return nc
}

func TestAuditConsumer_StartStop(t *testing.T) {
	nc := startAuditTestNATS(t)
	repo := newAuditRepoTestEnv(t)
	consumer := NewAuditConsumer(nc, repo)

	ctx, cancel := context.WithCancel(context.Background())
	require.NoError(t, consumer.Start(ctx))

	time.Sleep(50 * time.Millisecond)
	cancel()
	consumer.Stop()
}

func TestAuditConsumer_Stop_NilSafe(t *testing.T) {
	consumer := &AuditConsumer{}
	assertNotPanics := func() {
		consumer.Stop()
	}
	assertNotPanics()
}
