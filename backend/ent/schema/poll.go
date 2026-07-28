package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Poll содержит определение схемы для сущности голосования (опроса).
type Poll struct {
	ent.Schema
}

// Fields описывает поля сущности Poll.
func (Poll) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}).
			Comment("ID канала, в котором создан опрос"),
		field.UUID("creator_id", uuid.UUID{}).
			Comment("ID создателя опроса"),
		field.String("question").
			NotEmpty().
			MaxLen(500).
			Comment("Текст вопроса опроса"),
		field.Bool("is_anonymous").
			Default(false).
			Comment("Анонимное голосование (не показывать кто голосовал)"),
		field.Bool("is_multi_choice").
			Default(false).
			Comment("Множественный выбор (можно выбрать несколько вариантов)"),
		field.Bool("is_closed").
			Default(false).
			Comment("Закрыт ли опрос для голосования"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (Poll) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Edges описывает связи сущности Poll.
func (Poll) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", Channel.Type).
			Ref("polls").
			Field("channel_id").
			Unique().
			Required(),
		edge.From("creator", User.Type).
			Ref("polls").
			Field("creator_id").
			Unique().
			Required(),
		edge.To("options", PollOption.Type),
		edge.To("votes", PollVote.Type),
	}
}

// Indexes описывает индексы сущности Poll.
func (Poll) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id"),
		index.Fields("creator_id"),
	}
}
