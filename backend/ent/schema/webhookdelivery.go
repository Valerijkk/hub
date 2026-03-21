package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// WebhookDelivery holds the schema definition for the WebhookDelivery entity.
type WebhookDelivery struct {
	ent.Schema
}

// Fields of the WebhookDelivery.
func (WebhookDelivery) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("event_type").
			NotEmpty().
			Comment("Тип события (issue.created, test_run.completed и т.д.)"),
		field.JSON("payload", map[string]interface{}{}).
			Comment("Тело отправленного запроса"),
		field.Int("status_code").
			Comment("HTTP-код ответа"),
		field.Text("response_body").
			Optional().
			Comment("Тело ответа"),
		field.Text("error").
			Optional().
			Nillable().
			Comment("Сообщение об ошибке при неудаче"),
		field.Time("delivered_at").
			Default(time.Now).
			Comment("Время доставки"),
	}
}

// Edges of the WebhookDelivery.
func (WebhookDelivery) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("endpoint", WebhookEndpoint.Type).
			Ref("deliveries").
			Unique().
			Required(),
	}
}

// Indexes of the WebhookDelivery.
func (WebhookDelivery) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("delivered_at"),
		index.Fields("event_type"),
	}
}
