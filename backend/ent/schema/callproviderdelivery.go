package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type CallProviderDelivery struct {
	ent.Schema
}

func (CallProviderDelivery) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.Enum("provider").Values("livekit", "jitsi").Immutable(),
		field.String("delivery_id").NotEmpty().MaxLen(255).Immutable(),
		field.Enum("kind").Values("connected", "disconnected").Immutable(),
		field.UUID("call_id", uuid.UUID{}).Immutable(),
		field.UUID("user_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("epoch", uuid.UUID{}).Immutable(),
		field.Enum("result").Values("pending", "applied", "ignored", "reconcile").Default("pending"),
		field.Time("received_at").Default(nowUTC).Immutable(),
		field.Time("terminal_at").Optional().Nillable(),
	}
}

func (CallProviderDelivery) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("provider", "delivery_id").Unique(),
		index.Fields("result", "terminal_at"),
		index.Fields("user_id"),
	}
}
