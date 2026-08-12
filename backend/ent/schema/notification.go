package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Notification хранит уведомление пользователя.
type Notification struct {
	ent.Schema
}

// Fields описывает поля сущности Notification.
func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID получателя уведомления"),
		field.String("type").
			NotEmpty().
			Comment("Тип уведомления: assigned, mentioned, status_changed и т.д."),
		field.String("title").
			NotEmpty().
			Comment("Заголовок уведомления"),
		field.String("body").
			Optional().
			Comment("Тело уведомления"),
		field.String("entity_type").
			Optional().
			Comment("Тип сущности: issue, channel, testrun и т.д."),
		// A2W8-3-006: polymorphic entity_id — UUID (без FK на разные таблицы).
		field.UUID("entity_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID связанной сущности (UUID; полиморф без FK)"),
		field.UUID("project_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID проекта (если применимо)"),
		field.UUID("actor_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID пользователя-инициатора"),
		field.Bool("is_read").
			Default(false).
			Comment("Прочитано ли уведомление"),
		field.Bool("is_urgent").
			Default(false).
			Comment("Срочное уведомление (пробивает тихие часы)"),
		// A2W8-3-006: OL для mark-read race.
		field.Int("version").
			Default(1),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges описывает связи сущности Notification.
func (Notification) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("notifications").
			Field("user_id").
			Required().
			Unique(),
	}
}

// Indexes описывает индексы сущности Notification.
func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "is_read"),
		index.Fields("user_id", "created_at"),
		index.Fields("user_id", "type", "entity_id", "created_at"),
	}
}
