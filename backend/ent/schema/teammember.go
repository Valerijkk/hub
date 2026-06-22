package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TeamMember — определение Ent-схемы сущности TeamMember.
type TeamMember struct {
	ent.Schema
}

// Поля сущности TeamMember.
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

// Рёбра (edges) сущности TeamMember.
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

// Индексы сущности TeamMember.
func (TeamMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("team", "user").
			Unique(),
	}
}
