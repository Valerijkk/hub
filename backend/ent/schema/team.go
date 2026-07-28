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
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (Team) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
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
