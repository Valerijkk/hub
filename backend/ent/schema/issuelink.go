package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// IssueLink — определение Ent-схемы сущности IssueLink.
type IssueLink struct {
	ent.Schema
}

// Поля сущности IssueLink.
func (IssueLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Enum("link_type").
			Values(
				"blocks", "blocked_by",
				"relates_to",
				"duplicates", "duplicated_by",
				"clones", "cloned_from",
				"causes", "caused_by",
				"parent", "child",
			),
		field.String("custom_type").
			Optional().
			Comment("For user-defined link types"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности IssueLink.
func (IssueLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("source", Issue.Type).
			Ref("outgoing_links").
			Unique().
			Required(),
		edge.To("target", Issue.Type).
			Unique().
			Required(),
	}
}

// Индексы сущности IssueLink.
func (IssueLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("source", "target").
			Unique(),
	}
}
