package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
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
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges определяет связи сущности Sprint.
func (Sprint) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("sprints").
			Unique().
			Required(),
		edge.To("issues", Issue.Type),
	}
}

// Indexes определяет индексы сущности Sprint.
func (Sprint) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("project").
			Fields("status"),
		index.Edges("project").
			Fields("start_date"),
	}
}
