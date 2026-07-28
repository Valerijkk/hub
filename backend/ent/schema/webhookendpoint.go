package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// WebhookEndpoint — определение Ent-схемы сущности WebhookEndpoint.
type WebhookEndpoint struct {
	ent.Schema
}

// Поля сущности WebhookEndpoint.
func (WebhookEndpoint) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("url").
			NotEmpty().
			Comment("URL для отправки POST-запросов"),
		field.String("secret").
			NotEmpty().
			Sensitive().
			Comment("Секрет для HMAC-SHA256 подписи payload (at-rest encrypted)"),
		field.JSON("events", []string{}).
			Comment("Массив подписанных событий"),
		field.Bool("is_active").
			Default(true).
			Comment("Активен ли вебхук"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Время мягкого удаления (корзина 30 дней)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (WebhookEndpoint) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Рёбра (edges) сущности WebhookEndpoint.
func (WebhookEndpoint) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("webhook_endpoints").
			Unique().
			Required(),
		edge.To("deliveries", WebhookDelivery.Type),
	}
}

// Индексы сущности WebhookEndpoint.
func (WebhookEndpoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
		index.Fields("deleted_at"),
	}
}
