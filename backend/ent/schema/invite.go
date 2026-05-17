package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Invite holds the schema definition for the Invite entity.
type Invite struct {
	ent.Schema
}

// Fields of the Invite.
func (Invite) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("code").
			Unique().
			NotEmpty(),
		field.Int("max_uses").
			Default(0).
			Comment("0 means unlimited"),
		field.Int("use_count").
			Default(0),
		field.Enum("project_role").
			Values("developer", "tester", "viewer").
			Default("developer"),
		field.Time("expires_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges of the Invite.
func (Invite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("invites").
			Unique().
			Required(),
		edge.From("created_by", User.Type).
			Ref("invites").
			Unique().
			Required(),
	}
}

// Indexes of the Invite.
func (Invite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expires_at"),
	}
}
