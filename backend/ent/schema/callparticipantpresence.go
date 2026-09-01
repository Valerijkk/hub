package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type CallParticipantPresence struct {
	ent.Schema
}

func (CallParticipantPresence) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("call_id", uuid.UUID{}).Immutable(),
		field.UUID("user_id", uuid.UUID{}).Immutable(),
		field.Int64("connected_count").Default(0).NonNegative(),
		field.Int64("presence_sequence").Default(0).NonNegative(),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (CallParticipantPresence) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("call", Call.Type).Field("call_id").Unique().Required().Immutable().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("user", User.Type).Field("user_id").Unique().Required().Immutable().Annotations(entsql.OnDelete(entsql.NoAction)),
	}
}

func (CallParticipantPresence) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("call_id", "user_id").Unique(),
		index.Fields("user_id"),
	}
}
