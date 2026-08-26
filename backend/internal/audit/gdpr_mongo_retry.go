package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// SubjectGDPRMongoErase — JetStream subject для compensating retry стирания
// Mongo messages по sender_id после GDPR Ent Commit (C-W19-002 / TASK-178).
// DeleteMany идемпотентен: sync erase + consumer / dual deliver безопасны.
const SubjectGDPRMongoErase = "audit.gdpr.mongo_erase"

const (
	gdprMongoEraseDurable   = "gdpr-mongo-erase"
	gdprMongoEraseMaxDeliver = 10
	gdprMongoEraseAckWait    = 30 * time.Second
)

// MongoEraseRetryEvent — payload retry-сообщения (только sender_id; user уже HardDelete).
type MongoEraseRetryEvent struct {
	SenderID uuid.UUID `json:"sender_id"`
	ActorID  uuid.UUID `json:"actor_id,omitempty"`
}

// MongoEraseRetryEnqueuer ставит compensating erase в JetStream (NATS only, не Redis).
// Опциональная зависимость AuditService: nil → fail eraser без durable retry (тесты).
type MongoEraseRetryEnqueuer interface {
	EnqueueMongoErase(ctx context.Context, senderID, actorID uuid.UUID) error
}

// JetStreamMongoEraseRetry публикует SubjectGDPRMongoErase через JetStream PubAck.
type JetStreamMongoEraseRetry struct {
	nc *natsclient.Client
}

// NewJetStreamMongoEraseRetry создаёт durable enqueue path.
// nc == nil → Enqueue вернёт error (fail-closed: не silent success).
func NewJetStreamMongoEraseRetry(nc *natsclient.Client) *JetStreamMongoEraseRetry {
	return &JetStreamMongoEraseRetry{nc: nc}
}

// EnqueueMongoErase пишет at-least-once событие в EVENTS (InterestPolicy + durable consumer).
func (r *JetStreamMongoEraseRetry) EnqueueMongoErase(ctx context.Context, senderID, actorID uuid.UUID) error {
	if r == nil || r.nc == nil {
		return fmt.Errorf("gdpr mongo erase retry: nats unavailable")
	}
	payload, err := json.Marshal(MongoEraseRetryEvent{SenderID: senderID, ActorID: actorID})
	if err != nil {
		return fmt.Errorf("marshal gdpr mongo erase retry: %w", err)
	}
	msg := nats.NewMsg(SubjectGDPRMongoErase)
	msg.Data = payload
	natsclient.InjectTraceContext(ctx, msg.Header)
	if err := r.nc.PublishMsgAck(ctx, msg); err != nil {
		return fmt.Errorf("publish gdpr mongo erase retry: %w", err)
	}
	return nil
}

// MongoEraseRetryConsumer — durable JetStream consumer: повторный DeleteMessagesBySender.
type MongoEraseRetryConsumer struct {
	nc      *natsclient.Client
	eraser  MessageEraser
	consCtx jetstream.ConsumeContext
}

// NewMongoEraseRetryConsumer создаёт consumer compensating erase.
func NewMongoEraseRetryConsumer(nc *natsclient.Client, eraser MessageEraser) *MongoEraseRetryConsumer {
	return &MongoEraseRetryConsumer{nc: nc, eraser: eraser}
}

// Start подписывается на SubjectGDPRMongoErase (MaxDeliver → Nak retry).
func (c *MongoEraseRetryConsumer) Start(ctx context.Context) error {
	if c == nil || c.nc == nil {
		slog.Warn("gdpr mongo erase consumer: nats nil, skip start")
		return nil
	}
	if c.eraser == nil {
		return fmt.Errorf("gdpr mongo erase consumer: eraser required")
	}

	stream, err := c.nc.JetStream.Stream(ctx, natsclient.StreamName)
	if err != nil {
		return fmt.Errorf("gdpr mongo erase stream: %w", err)
	}

	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:       gdprMongoEraseDurable,
		FilterSubject: SubjectGDPRMongoErase,
		AckPolicy:     jetstream.AckExplicitPolicy,
		MaxDeliver:    gdprMongoEraseMaxDeliver,
		AckWait:       gdprMongoEraseAckWait,
	})
	if err != nil {
		return fmt.Errorf("gdpr mongo erase consumer: %w", err)
	}

	cons, err := consumer.Consume(func(msg jetstream.Msg) {
		c.processMsg(ctx, msg)
	})
	if err != nil {
		return fmt.Errorf("gdpr mongo erase consume: %w", err)
	}
	c.consCtx = cons
	slog.Info("gdpr mongo erase retry consumer started", "subject", SubjectGDPRMongoErase)
	return nil
}

// Stop останавливает consumer.
func (c *MongoEraseRetryConsumer) Stop() {
	if c != nil && c.consCtx != nil {
		c.consCtx.Stop()
	}
}

func (c *MongoEraseRetryConsumer) processMsg(ctx context.Context, msg jetstream.Msg) {
	ctx = natsclient.ExtractTraceContext(ctx, msg.Headers())

	var evt MongoEraseRetryEvent
	if err := json.Unmarshal(msg.Data(), &evt); err != nil {
		slog.Error("gdpr mongo erase retry: bad payload", "error", err)
		_ = msg.Term() // poison — не крутить MaxDeliver
		return
	}
	if evt.SenderID == uuid.Nil {
		slog.Error("gdpr mongo erase retry: empty sender_id")
		_ = msg.Term()
		return
	}

	n, err := ExportPhase151DispatchLegacyMongoEraseRetry(ctx, evt, c.eraser)
	if err != nil {
		slog.Error("gdpr mongo erase retry failed",
			"error", err, "sender_id", evt.SenderID, "actor_id", evt.ActorID)
		if nakErr := msg.Nak(); nakErr != nil {
			slog.Error("gdpr mongo erase retry nak failed", "error", nakErr)
		}
		return
	}

	slog.Info("gdpr mongo erase retry ok",
		"sender_id", evt.SenderID, "deleted", n, "actor_id", evt.ActorID)
	if err := msg.Ack(); err != nil {
		slog.Error("gdpr mongo erase retry ack failed", "error", err)
	}
}
