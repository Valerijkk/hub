package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// WebhookEndpoint holds the schema definition for the WebhookEndpoint entity.
type WebhookEndpoint struct {
	ent.Schema
}

// Fields of the WebhookEndpoint.
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
			Comment("Секрет для HMAC-SHA256 подписи payload"),
		field.JSON("events", []string{}).
			Comment("Массив подписанных событий"),
		field.Bool("is_active").
			Default(true).
			Comment("Активен ли вебхук"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the WebhookEndpoint.
func (WebhookEndpoint) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("webhook_endpoints").
			Unique().
			Required(),
		edge.To("deliveries", WebhookDelivery.Type),
	}
}

// Indexes of the WebhookEndpoint.
func (WebhookEndpoint) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_active"),
	}
}
