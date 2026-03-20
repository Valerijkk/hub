package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// OutboxEvent holds the schema definition for the OutboxEvent entity.
type OutboxEvent struct {
	ent.Schema
}

// Fields of the OutboxEvent.
func (OutboxEvent) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("event_type").
			NotEmpty(),
		field.Bytes("payload"),
		field.Enum("status").
			Values("pending", "published", "failed").
			Default("pending"),
		field.Int("attempts").
			Default(0),
		field.String("last_error").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("published_at").
			Optional().
			Nillable(),
	}
}

// Indexes of the OutboxEvent.
func (OutboxEvent) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status", "created_at"),
	}
}
