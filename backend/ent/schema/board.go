package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
		// A2W8-1-004: timestamps + soft-delete для list/trash hot-path.
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("deleted_at").
			Optional().
			Nillable(),
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

// Индексы сущности Board.
func (Board) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
		// ListBoards WHERE project_boards = $1 AND deleted_at IS NULL
		index.Edges("project").Fields("deleted_at"),
	}
}
