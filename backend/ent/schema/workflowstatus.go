package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// WorkflowStatus holds the schema definition for the WorkflowStatus entity.
type WorkflowStatus struct {
	ent.Schema
}

// Fields of the WorkflowStatus.
func (WorkflowStatus) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(100),
		field.Enum("category").
			Values("todo", "in_progress", "done"),
		field.String("color").
			Optional().
			Default("#6B7280"),
		field.Int("sort_order").
			Default(0),
		field.Bool("is_default").
			Default(false),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the WorkflowStatus.
func (WorkflowStatus) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("workflow_statuses").
			Unique().
			Required(),
		edge.To("issues", Issue.Type),
		edge.To("from_transitions", WorkflowTransition.Type),
		edge.From("to_transitions", WorkflowTransition.Type).
			Ref("to_status"),
	}
}

// Indexes of the WorkflowStatus.
func (WorkflowStatus) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("project").
			Fields("sort_order"),
		index.Edges("project").
			Fields("name").
			Unique(),
	}
}
