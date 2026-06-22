package outbox_test
import (
	"github.com/valerii/hub/backend/internal/platform/outbox"
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/outboxevent"
)

func TestProcessEvent_PublishFail_IncrementsAttempts(t *testing.T) {
	client := newTestEntClient(t)
	nats := startTestNATSServer(t)
	nats.Close()

	w := outbox.NewWorker(client, nats)
	ctx := context.Background()
	evt := insertOutboxEvent(t, client, "fail.event", []byte(`{"x":1}`))

	outbox.ExportWorkerProcessEvent(w, ctx, evt)

	reloaded, err := client.OutboxEvent.Get(ctx, evt.ID)
	require.NoError(t, err)
	assert.Equal(t, outboxevent.StatusPending, reloaded.Status)
	assert.Equal(t, 1, reloaded.Attempts)
	assert.NotEmpty(t, reloaded.LastError)
}

func TestProcessBatch_ClosedClient_NoPanic(t *testing.T) {
	client := newTestEntClient(t)
	client.Close()
	w := outbox.NewWorker(client, startTestNATSServer(t))

	require.NotPanics(t, func() {
		outbox.ExportWorkerProcessBatch(w, context.Background())
	})
}
