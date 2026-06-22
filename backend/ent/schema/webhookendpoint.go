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
			Comment("URL РґР»СЏ РѕС‚РїСЂР°РІРєРё POST-Р·Р°РїСЂРѕСЃРѕРІ"),
		field.String("secret").
			NotEmpty().
			Comment("РЎРµРєСЂРµС‚ РґР»СЏ HMAC-SHA256 РїРѕРґРїРёСЃРё payload"),
		field.JSON("events", []string{}).
			Comment("РњР°СЃСЃРёРІ РїРѕРґРїРёСЃР°РЅРЅС‹С… СЃРѕР±С‹С‚РёР№"),
		field.Bool("is_active").
			Default(true).
			Comment("РђРєС‚РёРІРµРЅ Р»Рё РІРµР±С…СѓРє"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
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
	}
}
