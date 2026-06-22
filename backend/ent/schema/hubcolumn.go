package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// HubColumn — типизированная колонка HubTable.
type HubColumn struct {
	ent.Schema
}

func (HubColumn) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(200),
		field.Enum("col_type").
			Values("text", "number", "date", "select", "checkbox", "url", "user_ref").
			Default("text"),
		field.JSON("options", map[string]interface{}{}).
			Optional().
			Comment("select: {choices: []string}"),
		field.Int("sort_order").
			Default(0),
		field.Int("version").
			Default(1),
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

func (HubColumn) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("table", HubTable.Type).
			Ref("columns").
			Unique().
			Required(),
	}
}

func (HubColumn) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
		index.Fields("sort_order"),
	}
}
