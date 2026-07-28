package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AIProvider — настроенный LLM-провайдер (AI-01, v3.0).
type AIProvider struct {
	ent.Schema
}

// Поля сущности AIProvider.
func (AIProvider) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(100),
		field.Enum("provider_type").
			Values("openai_compatible", "ollama", "gigachat", "anthropic"),
		field.String("base_url").
			NotEmpty().
			MaxLen(500),
		field.Text("api_key_enc").
			Optional().
			Sensitive().
			Comment("AES-256-GCM encrypted API key"),
		field.JSON("models", []string{}).
			Default([]string{}).
			Comment("Allowlisted model IDs"),
		field.Bool("enabled").
			Default(true),
		field.Int("priority").
			Default(0).
			Comment("Fallback order — lower first"),
		field.Int("max_rpm").
			Default(60),
		field.Bool("ru_resident").
			Default(false),
		field.Enum("scope").
			Values("system", "user").
			Default("system"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Рёбра (edges) сущности AIProvider.
func (AIProvider) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("ai_providers").
			Unique(),
		// A2W8-1-005: обратная сторона FK из AIUserPreference.default_provider_id.
		edge.From("user_preferences", AIUserPreference.Type).
			Ref("default_provider"),
	}
}

// Индексы сущности AIProvider.
func (AIProvider) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("enabled", "priority"),
		index.Fields("scope"),
	}
}
