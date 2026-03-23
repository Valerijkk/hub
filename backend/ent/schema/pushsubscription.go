package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PushSubscription хранит подписку браузера на Web Push уведомления.
// Один пользователь может иметь несколько подписок (разные браузеры/устройства).
type PushSubscription struct {
	ent.Schema
}

// Fields описывает поля сущности PushSubscription.
func (PushSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID пользователя-владельца подписки"),
		field.String("endpoint").
			NotEmpty().
			Comment("Push endpoint URL (уникальный для каждого браузера)"),
		field.String("p256dh_key").
			NotEmpty().
			Comment("Публичный ключ клиента (base64url, RFC 8291)"),
		field.String("auth_key").
			NotEmpty().
			Comment("Shared auth secret (base64url, RFC 8291)"),
		field.String("user_agent").
			Optional().
			Comment("User-Agent браузера для идентификации устройства"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges описывает связи сущности PushSubscription.
func (PushSubscription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("push_subscriptions").
			Field("user_id").
			Required().
			Unique(),
	}
}

// Indexes описывает индексы сущности PushSubscription.
func (PushSubscription) Indexes() []ent.Index {
	return []ent.Index{
		// Уникальный endpoint — один браузер = одна подписка.
		index.Fields("endpoint").Unique(),
		// Быстрый поиск подписок по пользователю.
		index.Fields("user_id"),
	}
}
