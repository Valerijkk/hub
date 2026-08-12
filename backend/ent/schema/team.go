package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
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
func (Team) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

// Рёбра (edges) сущности Team.
// OnDelete Cascade на members — owned_junction (A3W8-5-003 / ADR-ENT-ONDELETE-CASCADE).
func (Team) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("projects", Project.Type).
			Ref("teams"),
		edge.To("members", TeamMember.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Индексы сущности Team.
func (Team) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
	}
}
