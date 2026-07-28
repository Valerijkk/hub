package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Role — определение Ent-схемы сущности Role.
type Role struct {
	ent.Schema
}

// Поля сущности Role.
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
			Default(nowUTC).
			Immutable(),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (Role) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Рёбра (edges) сущности Role.
func (Role) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("roles").
			Unique().
			Required(),
		edge.To("members", ProjectMember.Type),
	}
}

// Indexes — P2-BE-22 / W8-A4-005: unique (project, name).
func (Role) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("project").Fields("name").Unique(),
	}
}
