package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// WebhookDelivery — определение Ent-схемы сущности WebhookDelivery.
type WebhookDelivery struct {
	ent.Schema
}

// Поля сущности WebhookDelivery.
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
			Default(nowUTC).
			Comment("Время доставки"),
	}
}

// Рёбра (edges) сущности WebhookDelivery.
func (WebhookDelivery) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("endpoint", WebhookEndpoint.Type).
			Ref("deliveries").
			Unique().
			Required(),
	}
}

// Индексы сущности WebhookDelivery.
func (WebhookDelivery) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("delivered_at"),
		index.Fields("event_type"),
	}
}
