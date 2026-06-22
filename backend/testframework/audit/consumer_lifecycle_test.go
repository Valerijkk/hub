package audit_test

// Тесты NATS consumer аудита: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/audit"
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

// TestAuditConsumer_StartStop проверяет сценарий: audit consumer start stop.
func TestAuditConsumer_StartStop(t *testing.T) {
	nc := startAuditTestNATS(t)
	repo := newAuditRepoTestEnv(t)
	consumer := audit.NewAuditConsumer(nc, repo)

	ctx, cancel := context.WithCancel(context.Background())
	require.NoError(t, consumer.Start(ctx))

	time.Sleep(50 * time.Millisecond)
	cancel()
	consumer.Stop()
}

// TestAuditConsumer_Stop_NilSafe проверяет сценарий: audit consumer stop nil safe.
func TestAuditConsumer_Stop_NilSafe(t *testing.T) {
	consumer := audit.ExportNewAuditConsumerForTest(nil)
	assertNotPanics := func() {
		consumer.Stop()
	}
	assertNotPanics()
}
