package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ScheduledCall хранит метаданные запланированного или повторяющегося звонка.
// Поддерживает RFC 5545 RRULE для повторяющихся звонков (ежедневный стендап и т.д.).
type ScheduledCall struct {
	ent.Schema
}

// Fields описывает поля сущности ScheduledCall.
func (ScheduledCall) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("title").
			NotEmpty().
			MaxLen(255).
			Comment("Название запланированного звонка"),
		field.String("description").
			Optional().
			MaxLen(2000).
			Comment("Описание звонка"),
		field.Time("scheduled_at").
			Comment("Следующее запланированное время звонка"),
		field.Int("duration_minutes").
			Optional().
			Default(60).
			Comment("Длительность звонка в минутах"),
		field.Int("reminder_minutes").
			Optional().
			Nillable().
			Comment("За сколько минут до звонка отправить напоминание (null = без напоминания)"),
		field.String("rrule").
			Optional().
			Comment("RFC 5545 RRULE для повторяющихся звонков"),
		field.UUID("channel_id", uuid.UUID{}).
			Comment("ID канала, в котором будет звонок"),
		field.UUID("project_id", uuid.UUID{}).
			Comment("ID проекта"),
		field.UUID("created_by", uuid.UUID{}).
			Comment("ID пользователя, создавшего запланированный звонок"),
		field.Bool("is_active").
			Default(true).
			Comment("Активен ли запланированный звонок"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges описывает связи сущности ScheduledCall.
func (ScheduledCall) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("scheduled_calls").
			Field("project_id").
			Unique().
			Required(),
		edge.From("channel", Channel.Type).
			Ref("scheduled_calls").
			Field("channel_id").
			Unique().
			Required(),
		edge.From("creator", User.Type).
			Ref("scheduled_calls").
			Field("created_by").
			Unique().
			Required(),
		edge.To("participants", User.Type).
			Comment("Приглашённые участники звонка"),
	}
}

// Indexes описывает индексы сущности ScheduledCall.
func (ScheduledCall) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "scheduled_at"),
		index.Fields("is_active", "scheduled_at"),
	}
}
