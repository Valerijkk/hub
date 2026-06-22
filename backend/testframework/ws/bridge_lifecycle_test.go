package ws_test

// Тесты модуля WebSocket: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/ws"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func startEmbeddedNATS(t *testing.T) *natsclient.Client {
	t.Helper()
	opts := &server.Options{Port: -1}
	s := natsserver.RunServer(opts)
	t.Cleanup(s.Shutdown)

	nc, err := natsclient.NewClient(context.Background(), s.ClientURL())
	require.NoError(t, err)
	t.Cleanup(nc.Close)
	return nc
}

// TestBridge_StartStop проверяет сценарий: bridge start stop.
func TestBridge_StartStop(t *testing.T) {
	hub, _ := startTestHub(t)
	nc := startEmbeddedNATS(t)

	bridge := ws.NewBridge(hub, nc)
	require.NoError(t, bridge.Start())
	assert.Equal(t, 6, ws.ExportBridgeSubsLen(bridge))
	bridge.Stop()
	assert.Nil(t, ws.ExportBridgeSubs(bridge))
}

// TestBridge_StartRoutesLiveNATSMessage проверяет сценарий: bridge start маршруты live nats message.
func TestBridge_StartRoutesLiveNATSMessage(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "bridge-live", "project:live-1")
	nc := startEmbeddedNATS(t)

	bridge := ws.NewBridge(hub, nc)
	require.NoError(t, bridge.Start())
	t.Cleanup(bridge.Stop)

	require.NoError(t, nc.Publish("project.live-1.issue.created", []byte(`{"key":"LIVE-1"}`)))

	select {
	case raw := <-ws.ExportClientSendCh(client):
		var msg ws.ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, "issue.created", msg.Event)
		assert.Equal(t, "project:live-1", msg.Room)
	case <-time.After(2 * time.Second):
		t.Fatal("bridge did not route NATS message to websocket client")
	}
}
