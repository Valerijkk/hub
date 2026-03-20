package outbox

import (
	"context"
	"encoding/json"

	"github.com/valerii/hub/backend/ent"
)

// Publisher writes events to the outbox table within an existing database transaction.
// This ensures events are published atomically with the primary data changes.
type Publisher struct{}

// NewPublisher creates a new outbox publisher.
func NewPublisher() *Publisher {
	return &Publisher{}
}

// PublishInTx records an event in the outbox table inside an existing Ent transaction.
// The event will be picked up by the Worker and published to NATS.
func (p *Publisher) PublishInTx(ctx context.Context, tx *ent.Tx, eventType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = tx.OutboxEvent.Create().
		SetEventType(eventType).
		SetPayload(data).
		Save(ctx)
	return err
}
