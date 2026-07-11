package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// OutboxEvent — определение Ent-схемы сущности OutboxEvent.
type OutboxEvent struct {
	ent.Schema
}

// Поля сущности OutboxEvent.
func (OutboxEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("event_type").
			NotEmpty(),
		field.Bytes("payload"),
		// W3C trace-context (traceparent/tracestate/baggage) REST-запроса,
		// породившего событие. Отдельная nullable-колонка вместо envelope в
		// payload: envelope сломал бы json.Unmarshal payload'а во всех
		// существующих consumers (notification, audit, webhook).
		field.JSON("trace_context", map[string]string{}).
			Optional(),
		field.Enum("status").
			Values("pending", "published", "failed").
			Default("pending"),
		field.Int("attempts").
			Default(0),
		field.String("last_error").
			Optional(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("published_at").
			Optional().
			Nillable(),
	}
}

// Индексы сущности OutboxEvent.
func (OutboxEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "created_at"),
	}
}
