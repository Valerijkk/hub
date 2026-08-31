package outbox

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/platform/events"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// Publisher записывает события в outbox-таблицу внутри существующей транзакции.
// Гарантирует атомарность публикации событий с основными изменениями данных.
type Publisher struct{}

type OutboxOriginKind string

const (
	OutboxOriginKindMailMessage      OutboxOriginKind = "mail_message"
	OutboxOriginKindCalendarEvent    OutboxOriginKind = "calendar_event"
	OutboxOriginKindCalendarAttendee OutboxOriginKind = "calendar_attendee"
)

type OutboxOrigin struct {
	Kind OutboxOriginKind
	ID   uuid.UUID
}

type OutboxPublicationReceipt struct {
	OutboxEventID   uuid.UUID
	EnvelopeEventID uuid.UUID
	EventType       string
	Origin          *OutboxOrigin
}

var (
	errInvalidOutboxOrigin         = errors.New("outbox: invalid origin")
	errInvalidOutboxOriginEnvelope = errors.New("outbox: invalid origin envelope")
	errOutboxPublicationReceipt    = errors.New("outbox: publication receipt mismatch")
)

// NewPublisher создаёт новый outbox-публикатор.
func NewPublisher() *Publisher {
	return &Publisher{}
}

// PublishInTx записывает событие в outbox-таблицу внутри существующей Ent-транзакции.
// Событие будет подхвачено Worker и опубликовано в NATS.
// Trace-context запроса сохраняется вместе с событием, чтобы Worker мог
// продолжить трейс на publish (async-граница REST -> NATS).
func (p *Publisher) PublishInTx(ctx context.Context, tx *ent.Tx, eventType string, payload any) error {
	_, err := p.PublishInTxWithReceipt(ctx, tx, eventType, payload, nil)
	return err
}

// Publish записывает событие в outbox-таблицу в автономной транзакции.
// Используется когда вызывающий код не управляет транзакцией явно.
func (p *Publisher) Publish(ctx context.Context, client *ent.Client, eventType string, payload any) error {
	data, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	create := client.OutboxEvent.Create().
		SetEventType(eventType).
		SetPayload(data)
	if tc := natsclient.TraceContextMap(ctx); tc != nil {
		create.SetTraceContext(tc)
	}
	_, err = create.Save(ctx)
	return err
}

// PublishInTxWithReceipt persists one event in the caller-owned transaction and returns its identity.
func (p *Publisher) PublishInTxWithReceipt(
	ctx context.Context,
	tx *ent.Tx,
	eventType string,
	payload any,
	origin *OutboxOrigin,
) (OutboxPublicationReceipt, error) {
	var (
		envelopeEventID uuid.UUID
		acceptedOrigin  *OutboxOrigin
	)
	if origin != nil {
		validatedOrigin, err := validateOutboxOrigin(origin)
		if err != nil {
			return OutboxPublicationReceipt{}, err
		}
		envelope, ok := payload.(events.EnvelopeV1)
		if !ok || envelope.EventID() == uuid.Nil || envelope.Subject() != eventType {
			return OutboxPublicationReceipt{}, errInvalidOutboxOriginEnvelope
		}
		envelopeEventID = envelope.EventID()
		acceptedOrigin = &validatedOrigin
	}

	data, err := json.Marshal(payload)
	if err != nil {
		return OutboxPublicationReceipt{}, err
	}
	traceContext := natsclient.TraceContextMap(ctx)
	create := tx.OutboxEvent.Create().
		SetEventType(eventType).
		SetPayload(data)
	if acceptedOrigin != nil {
		create.SetOriginKind(string(acceptedOrigin.Kind)).
			SetOriginID(acceptedOrigin.ID)
	}
	if traceContext != nil {
		create.SetTraceContext(traceContext)
	}
	saved, err := create.Save(ctx)
	if err != nil {
		return OutboxPublicationReceipt{}, err
	}

	receipt := OutboxPublicationReceipt{
		OutboxEventID:   saved.ID,
		EnvelopeEventID: envelopeEventID,
		EventType:       saved.EventType,
	}
	if saved.ID == uuid.Nil || saved.EventType != eventType {
		return OutboxPublicationReceipt{}, errOutboxPublicationReceipt
	}
	if acceptedOrigin == nil {
		if saved.OriginKind != nil || saved.OriginID != nil {
			return OutboxPublicationReceipt{}, errOutboxPublicationReceipt
		}
		return receipt, nil
	}
	if saved.OriginKind == nil || saved.OriginID == nil {
		return OutboxPublicationReceipt{}, errOutboxPublicationReceipt
	}
	savedOrigin := OutboxOrigin{
		Kind: OutboxOriginKind(*saved.OriginKind),
		ID:   *saved.OriginID,
	}
	if savedOrigin != *acceptedOrigin {
		return OutboxPublicationReceipt{}, errOutboxPublicationReceipt
	}
	receipt.Origin = &savedOrigin
	return receipt, nil
}

func validateOutboxOrigin(origin *OutboxOrigin) (OutboxOrigin, error) {
	if origin == nil {
		return OutboxOrigin{}, errInvalidOutboxOrigin
	}
	copied := *origin
	if copied.ID == uuid.Nil {
		return OutboxOrigin{}, errInvalidOutboxOrigin
	}
	switch copied.Kind {
	case OutboxOriginKindMailMessage, OutboxOriginKindCalendarEvent, OutboxOriginKindCalendarAttendee:
		return copied, nil
	default:
		return OutboxOrigin{}, errInvalidOutboxOrigin
	}
}
