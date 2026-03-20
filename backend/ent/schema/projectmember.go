package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// ProjectMember holds the schema definition for the ProjectMember entity.
type ProjectMember struct {
	ent.Schema
}

// Fields of the ProjectMember.
func (ProjectMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Enum("project_role").
			Values("owner", "developer", "tester", "viewer").
			Default("developer"),
		field.Time("joined_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the ProjectMember.
func (ProjectMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("project_members").
			Unique().
			Required(),
		edge.From("project", Project.Type).
			Ref("members").
			Unique().
			Required(),
		edge.From("custom_role", Role.Type).
			Ref("members").
			Unique(),
	}
}

// Indexes of the ProjectMember.
func (ProjectMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("user", "project").
			Unique(),
	}
}
