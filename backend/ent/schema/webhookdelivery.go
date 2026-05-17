package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
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
			Comment("РўРёРї СЃРѕР±С‹С‚РёСЏ (issue.created, test_run.completed Рё С‚.Рґ.)"),
		field.JSON("payload", map[string]interface{}{}).
			Comment("РўРµР»Рѕ РѕС‚РїСЂР°РІР»РµРЅРЅРѕРіРѕ Р·Р°РїСЂРѕСЃР°"),
		field.Int("status_code").
			Comment("HTTP-РєРѕРґ РѕС‚РІРµС‚Р°"),
		field.Text("response_body").
			Optional().
			Comment("РўРµР»Рѕ РѕС‚РІРµС‚Р°"),
		field.Text("error").
			Optional().
			Nillable().
			Comment("РЎРѕРѕР±С‰РµРЅРёРµ РѕР± РѕС€РёР±РєРµ РїСЂРё РЅРµСѓРґР°С‡Рµ"),
		field.Time("delivered_at").
			Default(nowUTC).
			Comment("Р’СЂРµРјСЏ РґРѕСЃС‚Р°РІРєРё"),
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
