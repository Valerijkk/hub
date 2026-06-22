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
	}
}
