package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CallParticipant хранит информацию об участнике голосового звонка.
// Содержит время присоединения и ухода участника из звонка.
type CallParticipant struct {
	ent.Schema
}

// Fields описывает поля сущности CallParticipant.
func (CallParticipant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("call_id", uuid.UUID{}).
			Comment("ID звонка"),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID пользователя-участника"),
		field.Time("joined_at").
			Default(nowUTC).
			Comment("Время присоединения к звонку"),
		field.Time("left_at").
			Optional().
			Nillable().
			Comment("Время выхода из звонка (nil если ещё в звонке)"),
	}
}

// Edges описывает связи сущности CallParticipant.
func (CallParticipant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("call", Call.Type).
			Ref("participants").
			Field("call_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("call_participations").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes описывает индексы сущности CallParticipant.
func (CallParticipant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("call_id", "user_id").
			Unique(),
	}
}
