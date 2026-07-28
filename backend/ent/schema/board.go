package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Board — определение Ent-схемы сущности Board.
type Board struct {
	ent.Schema
}

// Поля сущности Board.
func (Board) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(100),
		field.JSON("filters", map[string]interface{}{}).
			Optional().
			Comment("Saved filter config: status, assignee, priority, type"),
		field.Bool("is_default").
			Default(false),
		field.Enum("swimlane_mode").
			Values("none", "assignee", "priority", "type").
			Default("none").
			Comment("Режим swimlanes на доске"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (Board) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Рёбра (edges) сущности Board.
func (Board) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("boards").
			Unique().
			Required(),
	}
}
