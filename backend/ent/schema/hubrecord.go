package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// HubRecord — строка HubTable; значения ячеек по UUID колонки.
type HubRecord struct {
	ent.Schema
}

func (HubRecord) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.JSON("cells", map[string]interface{}{}).
			Default(map[string]interface{}{}),
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
func (HubRecord) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


func (HubRecord) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("table", HubTable.Type).
			Ref("records").
			Unique().
			Required(),
		edge.From("created_by", User.Type).
			Ref("hub_records").
			Unique(),
	}
}

func (HubRecord) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
		// A2W8-2-009: ListRecords WHERE table AND deleted_at IS NULL
		index.Edges("table").Fields("deleted_at"),
	}
}
