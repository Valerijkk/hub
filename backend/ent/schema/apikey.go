package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// APIKey holds the schema definition for the APIKey entity.
type APIKey struct {
	ent.Schema
}

// Fields of the APIKey.
func (APIKey) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("Человекочитаемое имя ключа"),
		field.String("key_hash").
			Unique().
			NotEmpty().
			Comment("SHA-256 хеш API-ключа"),
		field.String("prefix").
			NotEmpty().
			MaxLen(8).
			Comment("Первые 8 символов ключа для отображения"),
		field.Time("last_used_at").
			Optional().
			Nillable().
			Comment("Время последнего использования"),
		field.Time("expires_at").
			Optional().
			Nillable().
			Comment("Срок действия ключа"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the APIKey.
func (APIKey) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("api_keys").
			Unique().
			Required(),
	}
}

// Indexes of the APIKey.
func (APIKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("key_hash"),
		index.Fields("expires_at"),
	}
}
