//go:build hubtest

package nats_test

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// TestPublishWithTraceContext_FlushAfterPublish_A3W4_2_003 — после Publish
// вызывается FlushTimeout (через PublishMsg). Hook гасит сервер до Flush →
// ошибка наружу (не silent success).
//
// Мутация «PublishWithTraceContext → Conn.PublishMsg без Flush»: hook не
// участвует / Flush не вызывается → err==nil → тест краснеет.
// ForceClose клиента недостаточен: FlushTimeout иногда успевает до CLOSED.
func TestPublishWithTraceContext_FlushAfterPublish_A3W4_2_003(t *testing.T) {
	opts := &server.Options{
		JetStream: true,
		StoreDir:  t.TempDir(),
		Port:      -1,
	}
	s := natsserver.RunServer(opts)
	t.Cleanup(s.Shutdown)

	client, err := natsclient.NewClient(context.Background(), s.ClientURL())
	require.NoError(t, err)
	t.Cleanup(client.ForceClose)

	var hookFired atomic.Bool
	natsclient.ExportSetAfterCorePublishHook(func() {
		hookFired.Store(true)
		s.Shutdown()
	})
	t.Cleanup(func() { natsclient.ExportSetAfterCorePublishHook(nil) })

	err = client.PublishWithTraceContext(context.Background(), "hub.test.trace.flush", []byte(`{"ok":true}`))
	assert.True(t, hookFired.Load(), "PublishMsg path must invoke afterCorePublishHook before Flush")
	assert.Error(t, err, "drop before Flush must surface error (not silent success)")
}

// TestPublishJSONContext_DropBeforeFlush_NoPanic_A3W4_2_003 —
// PublishJSONContext наследует Flush через PublishWithTraceContext;
// drop до Flush не паникует (fire-and-forget лог ошибки).
func TestPublishJSONContext_DropBeforeFlush_NoPanic_A3W4_2_003(t *testing.T) {
	opts := &server.Options{
		JetStream: true,
		StoreDir:  t.TempDir(),
		Port:      -1,
	}
	s := natsserver.RunServer(opts)
	t.Cleanup(s.Shutdown)

	client, err := natsclient.NewClient(context.Background(), s.ClientURL())
	require.NoError(t, err)
	t.Cleanup(client.ForceClose)

	natsclient.ExportSetAfterCorePublishHook(func() {
		s.Shutdown()
	})
	t.Cleanup(func() { natsclient.ExportSetAfterCorePublishHook(nil) })

	assert.NotPanics(t, func() {
		client.PublishJSONContext(context.Background(), "hub.test.trace.json", map[string]string{"k": "v"})
	})
}
