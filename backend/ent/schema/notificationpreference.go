package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// NotificationPreference хранит настройки уведомлений пользователя.
type NotificationPreference struct {
	ent.Schema
}

// Fields описывает поля сущности NotificationPreference.
func (NotificationPreference) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID пользователя"),
		field.String("notification_type").
			NotEmpty().
			Comment("Тип уведомления: assigned, mentioned и т.д."),
		field.String("channel").
			NotEmpty().
			Comment("Канал доставки: in_app, telegram"),
		field.Bool("enabled").
			Default(true).
			Comment("Включено ли уведомление для данного типа и канала"),
		field.UUID("project_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID проекта для per-project настройки (nil = глобальная)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (NotificationPreference) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Edges описывает связи сущности NotificationPreference.
func (NotificationPreference) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("notification_preferences").
			Field("user_id").
			Required().
			Unique(),
	}
}

// Indexes описывает индексы сущности NotificationPreference.
func (NotificationPreference) Indexes() []ent.Index {
	return []ent.Index{
		// P1-BE-41 / W8-A3-001: Postgres/SQLite NULL≠NULL в обычном UNIQUE —
		// два partial unique закрывают глобальные (project_id IS NULL) и per-project prefs.
		index.Fields("user_id", "notification_type", "channel").
			Unique().
			Annotations(entsql.IndexWhere("project_id IS NULL")),
		index.Fields("user_id", "notification_type", "channel", "project_id").
			Unique().
			Annotations(entsql.IndexWhere("project_id IS NOT NULL")),
	}
}
