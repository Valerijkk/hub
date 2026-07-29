package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Sprint содержит определение схемы сущности Sprint.
type Sprint struct {
	ent.Schema
}

// Fields определяет поля сущности Sprint.
func (Sprint) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(200),
		field.String("goal").
			Optional().
			MaxLen(1000).
			Comment("Цель спринта"),
		field.Enum("status").
			Values("planning", "active", "completed").
			Default("planning"),
		field.Time("start_date"),
		field.Time("end_date"),
		field.JSON("retro_notes", map[string]interface{}{}).
			Optional().
			Comment("JSONB заметки ретроспективы: went_well, to_improve, actions"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		// A2W8-4-004 / ADR-ENT-SOFT-DELETE-GAPS-2026: sprint trash (version → TASK-095).
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (Sprint) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

// Edges определяет связи сущности Sprint.
func (Sprint) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("sprints").
			Unique().
			Required(),
		// A3W8-4-002 / ADR-ENT-ONDELETE optional_ref: Issue.sprint optional → SET NULL.
		edge.To("issues", Issue.Type).
			Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

// Indexes определяет индексы сущности Sprint.
func (Sprint) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("project").
			Fields("status"),
		index.Edges("project").
			Fields("start_date"),
		index.Fields("deleted_at"),
		index.Edges("project").Fields("deleted_at"),
		// A3W8-4-001: ≤1 non-deleted active sprint per project (app-guard + DB).
		index.Edges("project").
			Unique().
			Annotations(entsql.IndexWhere("status = 'active' AND deleted_at IS NULL")),
	}
}
