package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Invite — определение Ent-схемы сущности Invite.
type Invite struct {
	ent.Schema
}

// Поля сущности Invite.
func (Invite) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("code").
			Unique().
			NotEmpty().
			MaxLen(64).
			Comment("Invite code (generator: 8 chars base32; MaxLen guard, DB varchar migrate needs owner approval)"),
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
		// A2W8-3-002: OL для use_count + soft-revoke вместо hard delete.
		field.Int("version").
			Default(1),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("revoked_at").
			Optional().
			Nillable().
			Comment("Мягкий revoke приглашения (аудит кода сохраняется)"),
	}
}

// Рёбра (edges) сущности Invite.
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

// Индексы сущности Invite.
func (Invite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expires_at"),
		index.Fields("revoked_at"),
	}
}
