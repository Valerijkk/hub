package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// HubTable — таблица в рабочей области HubBase.
type HubTable struct {
	ent.Schema
}

func (HubTable) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(200),
		field.Int("sort_order").
			Default(0),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (HubTable) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

// Edges — owned children CASCADE на PermanentDelete/GDPR hard-purge
// (ADR-ENT-ONDELETE-CASCADE-2026 / A3W8-3-001). Soft-delete OnDelete не вызывает.
func (HubTable) Edges() []ent.Edge {
	cascade := entsql.OnDelete(entsql.Cascade)
	return []ent.Edge{
		edge.From("base", HubBase.Type).
			Ref("tables").
			Unique().
			Required(),
		edge.To("columns", HubColumn.Type).Annotations(cascade),
		edge.To("records", HubRecord.Type).Annotations(cascade),
	}
}

func (HubTable) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
		index.Fields("sort_order"),
	}
}
