package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// HubBase — no-code рабочая область (V4-04, аналог Lark Base).
type HubBase struct {
	ent.Schema
}

func (HubBase) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(200),
		field.Text("description").
			Optional(),
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

func (HubBase) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("hub_bases").
			Unique().
			Required(),
		edge.To("tables", HubTable.Type),
	}
}

func (HubBase) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("deleted_at"),
	}
}
