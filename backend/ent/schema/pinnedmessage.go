package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PinnedMessage содержит определение схемы для закреплённого сообщения в канале.
type PinnedMessage struct {
	ent.Schema
}

// Fields описывает поля сущности PinnedMessage.
func (PinnedMessage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("message_id").
			NotEmpty().
			Comment("ID сообщения в MongoDB"),
		field.UUID("channel_id", uuid.UUID{}).
			Comment("ID канала"),
		field.UUID("pinned_by", uuid.UUID{}).
			Comment("ID пользователя, закрепившего сообщение"),
		field.Time("pinned_at").
			Default(nowUTC).
			Comment("Время закрепления"),
	}
}

// Edges описывает связи сущности PinnedMessage.
func (PinnedMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", Channel.Type).
			Ref("pinned_messages").
			Field("channel_id").
			Required().
			Unique(),
	}
}

// Indexes описывает индексы сущности PinnedMessage.
func (PinnedMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "message_id").Unique(),
	}
}
