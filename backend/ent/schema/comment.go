package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
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
		field.JSON("mentions", []uuid.UUID{}).
			Optional().
			Comment("User IDs mentioned (UUID)"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Soft-delete timestamp (30-day trash)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (Comment) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
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
		// A3W8-2-005: owned_child Cascade comment→reactions.
		edge.To("reactions", Reaction.Type).
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Индексы сущности Comment.
func (Comment) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("issue").
			Fields("created_at"),
		index.Fields("deleted_at"),
	}
}
