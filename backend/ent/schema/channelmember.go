package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ChannelMember содержит определение схемы для участника канала.
type ChannelMember struct {
	ent.Schema
}

// Fields описывает поля сущности ChannelMember.
func (ChannelMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}).
			Comment("ID канала"),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID пользователя"),
		field.Enum("role").
			Values("owner", "admin", "member").
			Default("member").
			Comment("Роль участника в канале"),
		field.Time("joined_at").
			Default(nowUTC).
			Comment("Время вступления в канал"),
		field.Bool("is_starred").
			Default(false).
			Comment("Флаг закрепления (starred) канала пользователем"),
		field.Time("muted_until").
			Optional().
			Nillable().
			Comment("Канал замьючен до этого времени (nil = не замьючен)"),
		field.Enum("notify_mode").
			Values("all", "mentions", "nothing").
			Default("all").
			Comment("Режим уведомлений: all=всё, mentions=только @упоминания, nothing=ничего"),
	}
}

// Edges описывает связи сущности ChannelMember.
func (ChannelMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", Channel.Type).
			Ref("members").
			Field("channel_id").
			Required().
			Unique(),
		edge.From("user", User.Type).
			Ref("channel_memberships").
			Field("user_id").
			Required().
			Unique(),
	}
}

// Indexes описывает индексы сущности ChannelMember.
func (ChannelMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "user_id").Unique(),
	}
}
