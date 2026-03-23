package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Channel содержит определение схемы для сущности Channel (канал мессенджера).
type Channel struct {
	ent.Schema
}

// Fields описывает поля сущности Channel.
func (Channel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Enum("channel_type").
			Values("dm", "group", "public", "private").
			Comment("Тип канала: личный чат, групповой, публичный или приватный"),
		field.String("name").
			Optional().
			MaxLen(100).
			Comment("Название канала (пусто для DM)"),
		field.String("description").
			Optional().
			MaxLen(500).
			Comment("Описание канала"),
		field.Bool("is_readonly").
			Default(false).
			Comment("Флаг read-only канала (например, CI-уведомления)"),
		field.UUID("project_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID проекта (nil для DM и глобальных каналов)"),
		field.UUID("created_by", uuid.UUID{}).
			Comment("ID создателя канала"),
		field.String("dm_key").
			Optional().
			Unique().
			Nillable().
			Comment("Ключ дедупликации DM: отсортированные user IDs через '-'"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges описывает связи сущности Channel.
func (Channel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", ChannelMember.Type),
		edge.From("project", Project.Type).
			Ref("channels").
			Field("project_id").
			Unique(),
		edge.From("creator", User.Type).
			Ref("created_channels").
			Field("created_by").
			Unique().
			Required(),
		edge.To("pinned_messages", PinnedMessage.Type),
		edge.To("calls", Call.Type),
		edge.To("scheduled_calls", ScheduledCall.Type),
		edge.To("polls", Poll.Type),
		edge.To("invites", ChannelInvite.Type),
	}
}

// Indexes описывает индексы сущности Channel.
func (Channel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_type"),
		index.Fields("project_id"),
	}
}
