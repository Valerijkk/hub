package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
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
// parent/child убраны из link_type: иерархия только через Issue.self-edge
// (ADR-ENT-ONDELETE-CASCADE-2026 / A2W8-3-003).
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
// OnDelete Cascade: связь не переживает удаление source/target issue.
func (IssueLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("source", Issue.Type).
			Ref("outgoing_links").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.To("target", Issue.Type).
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

// Индексы сущности IssueLink.
// Unique (source, target, link_type) — blocks+relates на одной паре допустимы (A2W8-3-003).
func (IssueLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("source", "target").
			Fields("link_type").
			Unique(),
	}
}
