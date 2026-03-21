package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// Component holds the schema definition for the Component entity.
type Component struct {
	ent.Schema
}

// Fields of the Component.
func (Component) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(200),
		field.Text("description").
			Optional(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Component.
func (Component) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("components").
			Unique().
			Required(),
		edge.From("test_cases", TestCase.Type).
			Ref("components"),
	}
}

// Indexes of the Component.
func (Component) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
	}
}
