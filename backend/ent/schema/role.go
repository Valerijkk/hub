package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	"time"
)

// Role holds the schema definition for the Role entity.
type Role struct {
	ent.Schema
}

// Fields of the Role.
func (Role) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty(),
		field.String("description").
			Optional(),
		field.Uint64("permissions").
			Default(0).
			Comment("Cached bit-projection of permission_names that fit in 64 bits (fast path for middleware)"),
		field.JSON("permission_names", []string{}).
			Optional().
			Comment("Authoritative list of permission names; permissions uint64 is a cached bit-projection of those that fit in 64 bits"),
		field.Bool("is_system").
			Default(false),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the Role.
func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("roles").
			Unique().
			Required(),
		edge.To("members", ProjectMember.Type),
	}
}
