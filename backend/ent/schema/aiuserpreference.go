package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AIUserPreference — выбор модели по умолчанию.
type AIUserPreference struct {
	ent.Schema
}

// Поля сущности AIUserPreference.
func (AIUserPreference) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		// A2W8-1-005: FK edge на AIProvider (orphan UUID запрещён).
		field.UUID("default_provider_id", uuid.UUID{}).
			Optional().
			Nillable(),
		field.String("default_model").
			Optional().
			MaxLen(100),
		field.Bool("allow_external_ai").
			Default(true).
			Comment("M3-030: user consent to send prompts to external LLM APIs"),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (AIUserPreference) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Рёбра (edges) сущности AIUserPreference.
func (AIUserPreference) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("ai_preference").
			Unique().
			Required(),
		edge.To("default_provider", AIProvider.Type).
			Field("default_provider_id").
			Unique(),
	}
}

// Индексы сущности AIUserPreference.
func (AIUserPreference) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("user").
			Unique(),
	}
}
