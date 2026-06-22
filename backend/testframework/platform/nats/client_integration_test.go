//go:build !short

package nats_test

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

func startTestNATSServer(t *testing.T) string {
	t.Helper()
	opts := &server.Options{
		JetStream: true,
		StoreDir:  t.TempDir(),
		Port:      -1,
	}
	s := natsserver.RunServer(opts)
	t.Cleanup(s.Shutdown)
	return s.ClientURL()
}

func TestNewClient_ConnectAndClose(t *testing.T) {
	url := startTestNATSServer(t)

	client, err := natsclient.NewClient(context.Background(), url)
	require.NoError(t, err)
	require.NotNil(t, client)
	require.NotNil(t, client.Conn)
	require.NotNil(t, client.JetStream)

	assert.True(t, client.Conn.IsConnected())
	client.Close()
	assert.True(t, client.Conn.IsClosed())
}

func TestClient_PublishSubscribe(t *testing.T) {
	url := startTestNATSServer(t)
	client, err := natsclient.NewClient(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	subject := "hub.test.ping"
	payload := []byte(`{"event":"ping"}`)

	var wg sync.WaitGroup
	wg.Add(1)
	var got []byte

	_, err = client.Subscribe(subject, func(msg *nats.Msg) {
		got = append([]byte(nil), msg.Data...)
		wg.Done()
	})
	require.NoError(t, err)
	require.NoError(t, client.Conn.Flush())

	require.NoError(t, client.Publish(subject, payload))
	require.NoError(t, client.Conn.Flush())

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()

	select {
	case <-waitDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for NATS message")
	}

	assert.Equal(t, payload, got)
}

func TestClient_EnsureStreams_CreatesEVENTS(t *testing.T) {
	url := startTestNATSServer(t)
	client, err := natsclient.NewClient(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, client.EnsureStreams(ctx))

	info, err := client.JetStream.Stream(ctx, natsclient.StreamName)
	require.NoError(t, err)
	assert.Equal(t, natsclient.StreamName, info.CachedInfo().Config.Name)
	assert.Contains(t, info.CachedInfo().Config.Subjects, "messenger.channel.>")
	assert.Contains(t, info.CachedInfo().Config.Subjects, "notification.>")
}

func TestClient_PublishJSON_NilClient_NoPanic(t *testing.T) {
	var c *natsclient.Client
	assert.NotPanics(t, func() {
		c.PublishJSON("any.subject", map[string]string{"k": "v"})
	})
}
