package outbox

import (
	"context"
	"encoding/json"

	"github.com/valerii/hub/backend/ent"
)

// Publisher записывает события в outbox-таблицу внутри существующей транзакции.
// Гарантирует атомарность публикации событий с основными изменениями данных.
type Publisher struct{}

// NewPublisher создаёт новый outbox-публикатор.
func NewPublisher() *Publisher {
	return &Publisher{}
}

// PublishInTx записывает событие в outbox-таблицу внутри существующей Ent-транзакции.
// Событие будет подхвачено Worker и опубликовано в NATS.
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

// Publish записывает событие в outbox-таблицу в автономной транзакции.
// Используется когда вызывающий код не управляет транзакцией явно.
func (p *Publisher) Publish(ctx context.Context, client *ent.Client, eventType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	_, err = client.OutboxEvent.Create().
		SetEventType(eventType).
		SetPayload(data).
		Save(ctx)
	return err
}
