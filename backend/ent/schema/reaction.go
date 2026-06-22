package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Reaction — определение Ent-схемы сущности Reaction.
type Reaction struct {
	ent.Schema
}

// Поля сущности Reaction.
func (Reaction) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("emoji").
			NotEmpty().
			MaxLen(32),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности Reaction.
func (Reaction) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("comment", Comment.Type).
			Ref("reactions").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("reactions").
			Unique().
			Required(),
	}
}

// Индексы сущности Reaction.
func (Reaction) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("comment", "user").
			Fields("emoji").
			Unique(),
	}
}
