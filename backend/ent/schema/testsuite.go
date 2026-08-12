package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TestSuite — Ent-схема набора тест-кейсов (папки/секции TMS).
type TestSuite struct {
	ent.Schema
}

// Поля сущности TestSuite.
func (TestSuite) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(200),
		field.Text("description").
			Optional(),
		field.Int("sort_order").
			Default(0),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Время мягкого удаления (корзина 30 дней)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (TestSuite) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Рёбра (edges) сущности TestSuite.
func (TestSuite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("test_suites").
			Unique().
			Required(),
		edge.To("test_cases", TestCase.Type),
		edge.To("children", TestSuite.Type).
			From("parent").
			Unique(),
	}
}

// Индексы сущности TestSuite.
func (TestSuite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
		index.Fields("deleted_at"),
		index.Edges("project").Fields("deleted_at"),
	}
}
