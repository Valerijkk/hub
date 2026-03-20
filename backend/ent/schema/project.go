package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// Project holds the schema definition for the Project entity.
type Project struct {
	ent.Schema
}

// Fields of the Project.
func (Project) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MinLen(2).
			MaxLen(100),
		field.String("key").
			Unique().
			NotEmpty().
			Match(keyRegexp),
		field.String("description").
			Optional().
			MaxLen(2000),
		field.String("avatar_url").
			Optional(),
		field.Enum("status").
			Values("active", "archived", "deleted").
			Default("active"),
		field.Int("task_counter").
			Default(0),
		field.Time("deleted_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges of the Project.
func (Project) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", ProjectMember.Type),
		edge.To("roles", Role.Type),
		edge.To("teams", Team.Type),
		edge.To("invites", Invite.Type),
		edge.To("files", FileMetadata.Type),
	}
}

// Indexes of the Project.
func (Project) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("deleted_at"),
	}
}
