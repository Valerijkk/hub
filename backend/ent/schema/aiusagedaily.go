package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AIUsageDaily — дневные счётчики использования AI на пользователя.
type AIUsageDaily struct {
	ent.Schema
}

// Поля сущности AIUsageDaily.
func (AIUsageDaily) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("usage_date").
			Comment("UTC date bucket"),
		field.Int("tokens_in").
			Default(0),
		field.Int("tokens_out").
			Default(0),
		field.Int("request_count").
			Default(0),
	}
}

// Рёбра (edges) сущности AIUsageDaily.
func (AIUsageDaily) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("ai_usage_daily").
			Unique().
			Required(),
	}
}

// Индексы сущности AIUsageDaily.
func (AIUsageDaily) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("user").
			Fields("usage_date").
			Unique(),
	}
}
