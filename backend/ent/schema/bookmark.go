package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Bookmark содержит определение схемы для закладки на сообщение.
type Bookmark struct {
	ent.Schema
}

// Fields описывает поля сущности Bookmark.
func (Bookmark) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID пользователя, создавшего закладку"),
		field.String("message_id").
			NotEmpty().
			Comment("ID сообщения в MongoDB"),
		field.UUID("channel_id", uuid.UUID{}).
			Comment("ID канала сообщения"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges описывает связи сущности Bookmark.
func (Bookmark) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("bookmarks").
			Field("user_id").
			Required().
			Unique(),
	}
}

// Indexes описывает индексы сущности Bookmark.
func (Bookmark) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "message_id").Unique(),
	}
}
