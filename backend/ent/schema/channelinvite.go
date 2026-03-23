package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ChannelInvite содержит определение схемы для пригласительной ссылки канала.
// Аналог invite-ссылок Telegram: генерируется токен, по которому можно вступить в канал.
type ChannelInvite struct {
	ent.Schema
}

// Fields описывает поля сущности ChannelInvite.
func (ChannelInvite) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}).
			Comment("ID канала, к которому относится приглашение"),
		field.UUID("creator_id", uuid.UUID{}).
			Comment("ID пользователя, создавшего приглашение"),
		field.String("token").
			Unique().
			NotEmpty().
			MaxLen(32).
			Comment("Уникальный токен приглашения (base62, 22 символа)"),
		field.Int("max_uses").
			Default(0).
			Comment("Максимальное число использований (0 = без лимита)"),
		field.Int("use_count").
			Default(0).
			Comment("Текущее число использований"),
		field.Time("expires_at").
			Optional().
			Nillable().
			Comment("Время истечения ссылки (nil = бессрочная)"),
		field.Bool("is_revoked").
			Default(false).
			Comment("Флаг отзыва приглашения"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges описывает связи сущности ChannelInvite.
func (ChannelInvite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", Channel.Type).
			Ref("invites").
			Field("channel_id").
			Required().
			Unique(),
		edge.From("creator", User.Type).
			Ref("channel_invites").
			Field("creator_id").
			Required().
			Unique(),
	}
}

// Indexes описывает индексы сущности ChannelInvite.
func (ChannelInvite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id"),
		index.Fields("token").Unique(),
		index.Fields("expires_at"),
	}
}
