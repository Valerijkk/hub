package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type CallParticipantConnection struct {
	ent.Schema
}

func (CallParticipantConnection) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("call_id", uuid.UUID{}).Immutable(),
		field.UUID("user_id", uuid.UUID{}).Immutable(),
		field.UUID("epoch", uuid.UUID{}).Immutable(),
		field.Enum("state").Values("unknown", "connected", "left").Default("unknown"),
		field.Time("connected_at").Optional().Nillable(),
		field.Time("left_at").Optional().Nillable(),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (CallParticipantConnection) Mixin() []ent.Mixin {
	return []ent.Mixin{VersionMixin{}}
}

func (CallParticipantConnection) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("call", Call.Type).Field("call_id").Unique().Required().Immutable().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("user", User.Type).Field("user_id").Unique().Required().Immutable().Annotations(entsql.OnDelete(entsql.NoAction)),
	}
}

func (CallParticipantConnection) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("call_id", "user_id", "epoch").Unique(),
		index.Fields("left_at").Annotations(entsql.IndexWhere("state = 'left'")),
		index.Fields("user_id"),
	}
}
