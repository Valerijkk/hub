package nats_test
import (
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
	"context"
	"sync"
	"testing"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startEmbeddedNATSServer(t *testing.T) string {
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

func TestClient_PublishJSON_Success(t *testing.T) {
	url := startEmbeddedNATSServer(t)
	client, err := natsclient.NewClient(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	subject := "hub.test.event"
	var wg sync.WaitGroup
	wg.Add(1)
	var got []byte
	_, err = client.Subscribe(subject, func(msg *nats.Msg) {
		got = append([]byte(nil), msg.Data...)
		wg.Done()
	})
	require.NoError(t, err)
	require.NoError(t, client.Conn.Flush())

	client.PublishJSON(subject, map[string]string{"event": "ping"})
	require.NoError(t, client.Conn.Flush())

	waitDone := make(chan struct{})
	go func() {
		wg.Wait()
		close(waitDone)
	}()
	select {
	case <-waitDone:
	case <-time.After(2 * time.Second):
		t.Fatal("timed out waiting for nats.PublishJSON message")
	}
	assert.Contains(t, string(got), "ping")
}

func TestClient_PublishJSON_InvalidPayload(t *testing.T) {
	url := startEmbeddedNATSServer(t)
	client, err := natsclient.NewClient(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	assert.NotPanics(t, func() {
		client.PublishJSON("hub.test.bad", make(chan int))
	})
}

func TestClient_Publish_CoreNATS(t *testing.T) {
	url := startEmbeddedNATSServer(t)
	client, err := natsclient.NewClient(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	require.NoError(t, client.Publish("hub.test.raw", []byte(`{"k":"v"}`)))
}

func TestClient_EnsureStreams_Idempotent(t *testing.T) {
	url := startEmbeddedNATSServer(t)
	client, err := natsclient.NewClient(context.Background(), url)
	require.NoError(t, err)
	t.Cleanup(client.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	require.NoError(t, client.EnsureStreams(ctx))
	require.NoError(t, client.EnsureStreams(ctx))

	info, err := client.JetStream.Stream(ctx, natsclient.StreamName)
	require.NoError(t, err)
	assert.Contains(t, info.CachedInfo().Config.Subjects, "notification.>")
}
