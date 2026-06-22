package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// FieldChange — изменение одного поля в журнале активности.
type FieldChange struct {
	Field    string `json:"field"`
	OldValue any    `json:"old_value"`
	NewValue any    `json:"new_value"`
}

// ActivityLog — определение Ent-схемы сущности ActivityLog.
type ActivityLog struct {
	ent.Schema
}

// Поля сущности ActivityLog.
func (ActivityLog) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("action").
			NotEmpty().
			Comment("created/updated/commented/status_changed/assigned/etc"),
		field.JSON("changes", []FieldChange{}).
			Optional().
			Comment("Field-level diffs"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности ActivityLog.
func (ActivityLog) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("issue", Issue.Type).
			Ref("activities").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("activity_logs").
			Unique().
			Required(),
	}
}

// Индексы сущности ActivityLog.
func (ActivityLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("issue").
			Fields("created_at"),
	}
}
