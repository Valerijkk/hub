package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Comment — определение Ent-схемы сущности Comment.
type Comment struct {
	ent.Schema
}

// Поля сущности Comment.
func (Comment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.JSON("body_json", map[string]interface{}{}).
			Comment("Plate editor JSON"),
		field.Text("body_text").
			Comment("Plain text for search/notifications"),
		field.Bool("is_internal").
			Default(false).
			Comment("Visible only to certain roles"),
		field.JSON("mentions", []string{}).
			Optional().
			Comment("User IDs mentioned"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Рёбра (edges) сущности Comment.
func (Comment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("issue", Issue.Type).
			Ref("comments").
			Unique().
			Required(),
		edge.From("author", User.Type).
			Ref("comments").
			Unique().
			Required(),
		edge.To("reactions", Reaction.Type),
	}
}

// Индексы сущности Comment.
func (Comment) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("issue").
			Fields("created_at"),
	}
}
