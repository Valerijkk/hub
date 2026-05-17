package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// FieldChange represents a single field-level diff in an activity log entry.
type FieldChange struct {
	Field    string `json:"field"`
	OldValue any    `json:"old_value"`
	NewValue any    `json:"new_value"`
}

// ActivityLog holds the schema definition for the ActivityLog entity.
type ActivityLog struct {
	ent.Schema
}

// Fields of the ActivityLog.
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

// Edges of the ActivityLog.
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

// Indexes of the ActivityLog.
func (ActivityLog) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("issue").
			Fields("created_at"),
	}
}
