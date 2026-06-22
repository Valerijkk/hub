package audit

import (
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// AuditConsumer — JetStream consumer для записи событий аудита из всех модулей.
type AuditConsumer struct {
	nc   *natsclient.Client
	repo Repository
	cons jetstream.ConsumeContext
}

// NewAuditConsumer создаёт новый consumer аудита.
func NewAuditConsumer(nc *natsclient.Client, repo Repository) *AuditConsumer {
	return &AuditConsumer{nc: nc, repo: repo}
}

// Start запускает JetStream durable consumer для событий аудита.
func (c *AuditConsumer) Start(ctx context.Context) error {
	js := c.nc.JetStream

	// Получаем общий стрим EVENTS (создаётся при старте сервера).
	stream, err := js.Stream(ctx, natsclient.StreamName)
	if err != nil {
		return err
	}

	// Durable consumer на общем стриме.
	consumer, err := stream.CreateOrUpdateConsumer(ctx, jetstream.ConsumerConfig{
		Durable:    "audit-logger",
		AckPolicy:  jetstream.AckExplicitPolicy,
		MaxDeliver: 5,
		AckWait:    30 * time.Second,
	})
	if err != nil {
		return err
	}

	// Обработка сообщений
	cons, err := consumer.Consume(func(msg jetstream.Msg) {
		c.processEvent(ctx, msg)
		if err := msg.Ack(); err != nil {
			slog.Error("ack audit event failed", "error", err)
		}
	})
	if err != nil {
		return err
	}
	c.cons = cons

	slog.Info("audit consumer started")
	return nil
}

// Stop останавливает consumer аудита.
func (c *AuditConsumer) Stop() {
	if c.cons != nil {
		c.cons.Stop()
	}
}

// processEvent обрабатывает событие и записывает в журнал аудита.
func (c *AuditConsumer) processEvent(ctx context.Context, msg jetstream.Msg) {
	subject := msg.Subject()
	action, entityType := c.mapSubjectToAction(subject)
	if action == "" {
		return
	}

	// Разбираем payload для получения ID
	var payload map[string]interface{}
	if err := json.Unmarshal(msg.Data(), &payload); err != nil {
		slog.Error("unmarshal audit event", "error", err, "subject", subject)
		return
	}

	entityID := extractStringField(payload, "id", "issue_id", "channel_id", "test_case_id", "call_id")
	userID := extractUUIDField(payload, "user_id", "created_by_id", "sender_id")
	projectID := extractUUIDField(payload, "project_id")

	_, err := c.repo.CreateAuditLog(ctx, action, entityType, entityID, userID, projectID, payload, "", "")
	if err != nil {
		slog.Error("create audit log", "error", err, "subject", subject)
	}
}

// mapSubjectToAction маршрутизирует NATS subject в пару (действие, тип сущности).
func (c *AuditConsumer) mapSubjectToAction(subject string) (string, string) {
	parts := strings.Split(subject, ".")

	// auth.session.{event} -> login/logout, session
	if len(parts) >= 3 && parts[0] == "auth" && parts[1] == "session" {
		switch parts[2] {
		case "created":
			return "login", "session"
		case "revoked":
			return "logout", "session"
		}
	}

	// tracker.{projID}.{entity}.{event} -> CRUD по типу сущности
	if len(parts) >= 2 && parts[0] == "tracker" {
		if len(parts) >= 4 {
			entityType := parts[2]
			action := mapCRUDSuffix(parts[3])
			return action, entityType
		}
	}

	// project.{created|updated|deleted}
	if len(parts) >= 2 && parts[0] == "project" {
		action := mapCRUDSuffix(parts[len(parts)-1])
		return action, "project"
	}

	// tms.{projID}.{entity}.{event}
	if len(parts) >= 2 && parts[0] == "tms" {
		if len(parts) >= 4 {
			entityType := parts[2]
			action := mapCRUDSuffix(parts[3])
			return action, entityType
		}
	}

	// messenger.channel.{id}.{entity}.{event}
	if len(parts) >= 2 && parts[0] == "messenger" {
		if len(parts) >= 5 {
			entityType := parts[3]
			action := mapCRUDSuffix(parts[4])
			return action, entityType
		}
	}

	// calls.channel.{id}.{event}
	if len(parts) >= 2 && parts[0] == "calls" {
		if len(parts) >= 4 {
			action := mapCRUDSuffix(parts[3])
			return action, "call"
		}
	}

	return "", ""
}

// mapCRUDSuffix конвертирует суффикс события в действие аудита.
func mapCRUDSuffix(suffix string) string {
	switch suffix {
	case "created", "sent":
		return "create"
	case "updated", "edited":
		return "update"
	case "deleted", "removed":
		return "delete"
	case "started":
		return "create"
	case "ended", "finished":
		return "update"
	default:
		return suffix
	}
}

// extractStringField извлекает первое непустое строковое поле из payload.
func extractStringField(payload map[string]interface{}, keys ...string) string {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				return s
			}
		}
	}
	return ""
}

// extractUUIDField извлекает первый валидный UUID из payload.
func extractUUIDField(payload map[string]interface{}, keys ...string) *uuid.UUID {
	for _, key := range keys {
		if v, ok := payload[key]; ok {
			if s, ok := v.(string); ok && s != "" {
				if id, err := uuid.Parse(s); err == nil {
					return &id
				}
			}
		}
	}
	return nil
}
