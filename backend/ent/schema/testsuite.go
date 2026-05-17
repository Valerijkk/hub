package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TestSuite holds the schema definition for the TestSuite entity (folders/sections).
type TestSuite struct {
	ent.Schema
}

// Fields of the TestSuite.
func (TestSuite) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(200),
		field.Text("description").
			Optional(),
		field.Int("sort_order").
			Default(0),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Edges of the TestSuite.
func (TestSuite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("test_suites").
			Unique().
			Required(),
		edge.To("test_cases", TestCase.Type),
		edge.To("children", TestSuite.Type).
			From("parent").
			Unique(),
	}
}

// Indexes of the TestSuite.
func (TestSuite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("name"),
	}
}
