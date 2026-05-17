package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TeamMember holds the schema definition for the TeamMember entity.
type TeamMember struct {
	ent.Schema
}

// Fields of the TeamMember.
func (TeamMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("joined_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges of the TeamMember.
func (TeamMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("team", Team.Type).
			Ref("members").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("team_members").
			Unique().
			Required(),
	}
}

// Indexes of the TeamMember.
func (TeamMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("team", "user").
			Unique(),
	}
}
