package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserBlock хранит блокировку одного пользователя другим.
// Заблокированный пользователь не может добавить блокирующего в каналы
// и не может отправить ему DM.
type UserBlock struct {
	ent.Schema
}

// Поля сущности UserBlock.
func (UserBlock) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("blocker_id", uuid.UUID{}).
			Comment("ID пользователя, который блокирует"),
		field.UUID("blocked_id", uuid.UUID{}).
			Comment("ID заблокированного пользователя"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// ибра (edges) сущности UserBlock.
func (UserBlock) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("blocker", User.Type).
			Ref("blocks_created").
			Field("blocker_id").
			Unique().
			Required(),
		edge.From("blocked", User.Type).
			Ref("blocks_received").
			Field("blocked_id").
			Unique().
			Required(),
	}
}

// Индексы сущности UserBlock.
func (UserBlock) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("blocker_id", "blocked_id").
			Unique(),
	}
}
