package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Notification С…СЂР°РЅРёС‚ СѓРІРµРґРѕРјР»РµРЅРёРµ РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ.
type Notification struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё Notification.
func (Notification) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID РїРѕР»СѓС‡Р°С‚РµР»СЏ СѓРІРµРґРѕРјР»РµРЅРёСЏ"),
		field.String("type").
			NotEmpty().
			Comment("РўРёРї СѓРІРµРґРѕРјР»РµРЅРёСЏ: assigned, mentioned, status_changed Рё С‚.Рґ."),
		field.String("title").
			NotEmpty().
			Comment("Р—Р°РіРѕР»РѕРІРѕРє СѓРІРµРґРѕРјР»РµРЅРёСЏ"),
		field.String("body").
			Optional().
			Comment("РўРµР»Рѕ СѓРІРµРґРѕРјР»РµРЅРёСЏ"),
		field.String("entity_type").
			Optional().
			Comment("РўРёРї СЃСѓС‰РЅРѕСЃС‚Рё: issue, channel, testrun Рё С‚.Рґ."),
		field.String("entity_id").
			Optional().
			Comment("ID СЃРІСЏР·Р°РЅРЅРѕР№ СЃСѓС‰РЅРѕСЃС‚Рё"),
		field.UUID("project_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID РїСЂРѕРµРєС‚Р° (РµСЃР»Рё РїСЂРёРјРµРЅРёРјРѕ)"),
		field.UUID("actor_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ-РёРЅРёС†РёР°С‚РѕСЂР°"),
		field.Bool("is_read").
			Default(false).
			Comment("РџСЂРѕС‡РёС‚Р°РЅРѕ Р»Рё СѓРІРµРґРѕРјР»РµРЅРёРµ"),
		field.Bool("is_urgent").
			Default(false).
			Comment("РЎСЂРѕС‡РЅРѕРµ СѓРІРµРґРѕРјР»РµРЅРёРµ (РїСЂРѕР±РёРІР°РµС‚ С‚РёС…РёРµ С‡Р°СЃС‹)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё Notification.
func (Notification) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("notifications").
			Field("user_id").
			Required().
			Unique(),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё Notification.
func (Notification) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "is_read"),
		index.Fields("user_id", "created_at"),
		index.Fields("user_id", "type", "entity_id", "created_at"),
	}
}
