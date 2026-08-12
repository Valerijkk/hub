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
		// A2W8-3-005 / ADR-ENT-SOFT-DELETE-GAPS-2026: catalog soft-delete (version → TASK-095).
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (Label) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
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
		index.Fields("deleted_at"),
		index.Edges("project").Fields("deleted_at"),
	}
}
