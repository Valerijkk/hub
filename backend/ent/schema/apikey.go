package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// APIKey — определение Ent-схемы сущности APIKey.
type APIKey struct {
	ent.Schema
}

// Поля сущности APIKey.
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
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Время мягкого удаления / revoke (корзина 30 дней)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности APIKey.
func (APIKey) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("api_keys").
			Unique().
			Required(),
	}
}

// Индексы сущности APIKey.
// key_hash уже Unique() на field — отдельный индекс не дублируем (P3-BE-17 / W8-A1-009).
func (APIKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("expires_at"),
		index.Fields("deleted_at"),
	}
}
