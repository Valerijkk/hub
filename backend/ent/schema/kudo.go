package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Kudo — благодарность коллеге (MOD-03, Hub One v4).
type Kudo struct {
	ent.Schema
}

func (Kudo) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Text("message").
			NotEmpty().
			MaxLen(500),
		field.Bool("is_public").
			Default(true),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

func (Kudo) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("sender", User.Type).
			Ref("kudos_sent").
			Unique().
			Required(),
		edge.From("recipient", User.Type).
			Ref("kudos_received").
			Unique().
			Required(),
	}
}

func (Kudo) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("is_public", "created_at"),
		index.Fields("created_at"),
	}
}
