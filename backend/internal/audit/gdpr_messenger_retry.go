package audit

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

const (
	SubjectGDPRMessengerErase       = "audit.gdpr.messenger_erase_v1"
	GDPRMessengerEraseEventVersion = 1
	GDPRMessengerEraseOperation    = "messenger_full_v1"

	gdprMessengerEraseDurable    = "gdpr-messenger-erase-v1"
	gdprMessengerEraseMaxDeliver = 10
	gdprMessengerEraseAckWait    = 30 * time.Second
)

var ErrInvalidGDPRMessengerEraseEvent = errors.New("invalid GDPR messenger erase event")

// GDPRMessengerEraseRetryEvent identifies one complete, idempotent Messenger
// cleanup. Tenant authority is intentionally not accepted from the event: the
// injected eraser resolves the locked singleton Organization.ID on every try.
type GDPRMessengerEraseRetryEvent struct {
	Version   int       `json:"version"`
	Operation string    `json:"operation"`
	SubjectID uuid.UUID `json:"subject_id"`
	ActorID   uuid.UUID `json:"actor_id,omitempty"`
}

type MessengerEraseRetryEnqueuer interface {
	EnqueueMessengerErase(context.Context, uuid.UUID, uuid.UUID) error
}

func ExportPhase151DispatchLegacyMongoEraseRetry(
	ctx context.Context,
	event MongoEraseRetryEvent,
	eraser MessageEraser,
) (int64, error) {
	if event.SenderID == uuid.Nil {
		return 0, fmt.Errorf("legacy GDPR mongo erase retry: empty sender_id")
	}
	if eraser == nil {
		return 0, fmt.Errorf("legacy GDPR mongo erase retry: eraser required")
	}
	return eraser.DeleteMessagesBySender(ctx, event.SenderID)
}

func (r *JetStreamMongoEraseRetry) EnqueueMessengerErase(
	ctx context.Context,
	subjectID uuid.UUID,
	actorID uuid.UUID,
) error {
	if r == nil || r.nc == nil {
		return fmt.Errorf("GDPR messenger erase retry: nats unavailable")
	}
	if subjectID == uuid.Nil {
		return fmt.Errorf("%w: empty subject_id", ErrInvalidGDPRMessengerEraseEvent)
	}
	payload, err := json.Marshal(GDPRMessengerEraseRetryEvent{
		Version: GDPRMessengerEraseEventVersion, Operation: GDPRMessengerEraseOperation,
		SubjectID: subjectID, ActorID: actorID,
	})
	if err != nil {
		return fmt.Errorf("marshal GDPR messenger erase retry: %w", err)
	}
	msg := nats.NewMsg(SubjectGDPRMessengerErase)
	msg.Data = payload
	natsclient.InjectTraceContext(ctx, msg.Header)
	if err := r.nc.PublishMsgAck(ctx, msg); err != nil {
		return fmt.Errorf("publish GDPR messenger erase retry: %w", err)
	}
	return nil
}

type MessengerEraseRetryConsumer struct {
	nc      *natsclient.Client
	eraser  MessengerDataEraser
	consCtx jetstream.ConsumeContext
}

func NewMessengerEraseRetryConsumer(
	nc *natsclient.Client,
	eraser MessengerDataEraser,
) *MessengerEraseRetryConsumer {
	return &MessengerEraseRetryConsumer{nc: nc, eraser: eraser}
}

func (c *MessengerEraseRetryConsumer) Start(ctx context.Context) error {
	if c == nil || c.nc == nil {
		return fmt.Errorf("GDPR messenger erase consumer: nats required")
	}
	if c.eraser == nil {
		return fmt.Errorf("GDPR messenger erase consumer: full eraser required")
	}
	stream, err := c.nc.JetStream.Stream(ctx, natsclient.StreamName)
	if err != nil {
		return fmt.Errorf("GDPR messenger erase stream: %w", err)
	}
	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       gdprMessengerEraseDurable,
		FilterSubject: SubjectGDPRMessengerErase,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    gdprMessengerEraseMaxDeliver,
		AckWait:       gdprMessengerEraseAckWait,
	})
	if err != nil {
		return fmt.Errorf("GDPR messenger erase consumer: %w", err)
	}
	cons, err := consumer.Consume(func(msg jetstream.Msg) {
		c.processMsg(ctx, msg)
	})
	if err != nil {
		return fmt.Errorf("GDPR messenger erase consume: %w", err)
	}
	c.consCtx = cons
	slog.Info("GDPR messenger erase retry consumer started", "subject", SubjectGDPRMessengerErase)
	return nil
}

func (c *MessengerEraseRetryConsumer) Stop() {
	if c != nil && c.consCtx != nil {
		c.consCtx.Stop()
	}
}

func (c *MessengerEraseRetryConsumer) processMsg(ctx context.Context, msg jetstream.Msg) {
	ctx = natsclient.ExtractTraceContext(ctx, msg.Headers())
	var event GDPRMessengerEraseRetryEvent
	if err := json.Unmarshal(msg.Data(), &event); err != nil {
		slog.Error("GDPR messenger erase retry: bad payload", "error", err, "subject", SubjectGDPRMessengerErase)
		_ = msg.Term()
		return
	}
	if err := ExportPhase151DispatchMessengerEraseRetry(ctx, event, c.eraser); err != nil {
		if errors.Is(err, ErrInvalidGDPRMessengerEraseEvent) {
			slog.Error("GDPR messenger erase retry: invalid event", "error", err,
				"version", event.Version, "operation", event.Operation, "subject_id", event.SubjectID,
				"subject", SubjectGDPRMessengerErase)
			_ = msg.Term()
			return
		}
		slog.Error("GDPR messenger erase retry failed", "error", err,
			"operation", event.Operation, "subject_id", event.SubjectID, "actor_id", event.ActorID)
		if nakErr := msg.Nak(); nakErr != nil {
			slog.Error("GDPR messenger erase retry nak failed", "error", nakErr)
		}
		return
	}
	if err := msg.Ack(); err != nil {
		slog.Error("GDPR messenger erase retry ack failed", "error", err)
	}
}

// ExportPhase151DispatchMessengerEraseRetry is the pure dispatch boundary used
// by the consumer and external-package acceptance tests.
func ExportPhase151DispatchMessengerEraseRetry(
	ctx context.Context,
	event GDPRMessengerEraseRetryEvent,
	eraser MessengerDataEraser,
) error {
	if event.Version != GDPRMessengerEraseEventVersion ||
		event.Operation != GDPRMessengerEraseOperation || event.SubjectID == uuid.Nil {
		return fmt.Errorf("%w: version=%d operation=%q subject_id=%s",
			ErrInvalidGDPRMessengerEraseEvent, event.Version, event.Operation, event.SubjectID)
	}
	if eraser == nil {
		return fmt.Errorf("GDPR messenger erase retry: full eraser required")
	}
	if _, err := eraser.EraseMessengerUserData(ctx, event.SubjectID); err != nil {
		return fmt.Errorf("erase complete Messenger data: %w", err)
	}
	return nil
}
