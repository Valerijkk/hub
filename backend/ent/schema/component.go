package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Component — определение Ent-схемы сущности Component.
type Component struct {
	ent.Schema
}

// Поля сущности Component.
func (Component) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(200),
		field.Text("description").
			Optional(),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Время мягкого удаления (корзина 30 дней)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (Component) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Рёбра (edges) сущности Component.
func (Component) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("components").
			Unique().
			Required(),
		edge.From("test_cases", TestCase.Type).
			Ref("components"),
	}
}

// Индексы сущности Component.
func (Component) Indexes() []ent.Index {
	return []ent.Index{
		// P2-BE-22 / W8-A2-006: unique (project, name).
		index.Edges("project").Fields("name").Unique(),
		index.Fields("deleted_at"),
		index.Edges("project").Fields("deleted_at"),
	}
}
