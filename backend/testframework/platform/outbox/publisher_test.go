package outbox_test
import (
	"github.com/valerii/hub/backend/internal/platform/outbox"
	"context"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/outboxevent"
)

func TestPublisher_Publish(t *testing.T) {
	client := newTestEntClient(t)
	ctx := context.Background()
	p := outbox.NewPublisher()

	err := p.Publish(ctx, client, "user.created", map[string]any{"user_id": "abc"})
	require.NoError(t, err)

	events, err := client.OutboxEvent.Query().All(ctx)
	require.NoError(t, err)
	require.Len(t, events, 1)
	assert.Equal(t, "user.created", events[0].EventType)
	assert.Equal(t, outboxevent.StatusPending, events[0].Status)

	var payload map[string]any
	require.NoError(t, json.Unmarshal(events[0].Payload, &payload))
	assert.Equal(t, "abc", payload["user_id"])
}

func TestPublisher_PublishInTx(t *testing.T) {
	client := newTestEntClient(t)
	ctx := context.Background()
	p := outbox.NewPublisher()

	tx, err := client.Tx(ctx)
	require.NoError(t, err)

	err = p.PublishInTx(ctx, tx, "project.updated", map[string]string{"key": "HUB"})
	require.NoError(t, err)
	require.NoError(t, tx.Commit())

	cnt, err := client.OutboxEvent.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, cnt)
}

func TestPublisher_Publish_InvalidPayload(t *testing.T) {
	client := newTestEntClient(t)
	p := outbox.NewPublisher()
	ch := make(chan int)
	err := p.Publish(context.Background(), client, "bad", ch)
	require.Error(t, err)
}
