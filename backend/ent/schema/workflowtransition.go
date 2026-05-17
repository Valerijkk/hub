package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// WorkflowTransition holds the schema definition for the WorkflowTransition entity.
type WorkflowTransition struct {
	ent.Schema
}

// Fields of the WorkflowTransition.
func (WorkflowTransition) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges of the WorkflowTransition.
func (WorkflowTransition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("from_status", WorkflowStatus.Type).
			Ref("from_transitions").
			Unique().
			Required(),
		edge.To("to_status", WorkflowStatus.Type).
			Unique().
			Required(),
	}
}

// Indexes of the WorkflowTransition.
func (WorkflowTransition) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("from_status", "to_status").
			Unique(),
	}
}
