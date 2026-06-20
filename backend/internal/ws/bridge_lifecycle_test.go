package ws

import (
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

func TestBridge_StartStop(t *testing.T) {
	hub, _ := startTestHub(t)
	nc := startEmbeddedNATS(t)

	bridge := NewBridge(hub, nc)
	require.NoError(t, bridge.Start())
	assert.Len(t, bridge.subs, 6)
	bridge.Stop()
	assert.Nil(t, bridge.subs)
}

func TestBridge_StartRoutesLiveNATSMessage(t *testing.T) {
	hub, _ := startTestHub(t)
	client := registerAndSubscribe(t, hub, "bridge-live", "project:live-1")
	nc := startEmbeddedNATS(t)

	bridge := NewBridge(hub, nc)
	require.NoError(t, bridge.Start())
	t.Cleanup(bridge.Stop)

	require.NoError(t, nc.Publish("project.live-1.issue.created", []byte(`{"key":"LIVE-1"}`)))

	select {
	case raw := <-client.send:
		var msg ServerMessage
		require.NoError(t, json.Unmarshal(raw, &msg))
		assert.Equal(t, "issue.created", msg.Event)
		assert.Equal(t, "project:live-1", msg.Room)
	case <-time.After(2 * time.Second):
		t.Fatal("bridge did not route NATS message to websocket client")
	}
}
