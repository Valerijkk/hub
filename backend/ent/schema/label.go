package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Label — определение Ent-схемы сущности Label.
type Label struct {
	ent.Schema
}

// Поля сущности Label.
func (Label) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(50),
		field.String("color").
			NotEmpty().
			Default("#3B82F6"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности Label.
func (Label) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("labels").
			Unique().
			Required(),
		edge.From("group", LabelGroup.Type).
			Ref("labels").
			Unique(),
		edge.From("issues", Issue.Type).
			Ref("labels"),
	}
}

// Индексы сущности Label.
func (Label) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("project").
			Fields("name").
			Unique(),
	}
}
