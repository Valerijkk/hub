package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
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
		field.String("origin_kind").
			Optional().
			Nillable().
			Immutable(),
		field.UUID("origin_id", uuid.UUID{}).
			Optional().
			Nillable().
			Immutable(),
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
// Annotations keeps the nullable origin pair structurally complete.
func (OutboxEvent) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"outbox_events_origin_pair": "(origin_kind IS NULL AND origin_id IS NULL) OR (origin_kind IS NOT NULL AND origin_id IS NOT NULL)",
		}),
	}
}

func (OutboxEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "created_at"),
		index.Fields("origin_kind", "origin_id", "event_type"),
	}
}
