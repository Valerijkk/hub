package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// LabelGroup — определение Ent-схемы сущности LabelGroup.
type LabelGroup struct {
	ent.Schema
}

// Поля сущности LabelGroup.
func (LabelGroup) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(50),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		// A2W8-3-005 / ADR-ENT-SOFT-DELETE-GAPS-2026 (version → TASK-095).
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

// Рёбра (edges) сущности LabelGroup.
func (LabelGroup) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("label_groups").
			Unique().
			Required(),
		edge.To("labels", Label.Type),
	}
}

// Индексы сущности LabelGroup.
func (LabelGroup) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("project").
			Fields("name").
			Unique(),
		index.Fields("deleted_at"),
		index.Edges("project").Fields("deleted_at"),
	}
}
