package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Team — определение Ent-схемы сущности Team.
type Team struct {
	ent.Schema
}

// Поля сущности Team.
func (Team) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности Team.
func (Team) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("projects", Project.Type).
			Ref("teams"),
		edge.To("members", TeamMember.Type),
	}
}
