package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// NotificationPreference С…СЂР°РЅРёС‚ РЅР°СЃС‚СЂРѕР№РєРё СѓРІРµРґРѕРјР»РµРЅРёР№ РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ.
type NotificationPreference struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё NotificationPreference.
func (NotificationPreference) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ"),
		field.String("notification_type").
			NotEmpty().
			Comment("РўРёРї СѓРІРµРґРѕРјР»РµРЅРёСЏ: assigned, mentioned Рё С‚.Рґ."),
		field.String("channel").
			NotEmpty().
			Comment("РљР°РЅР°Р» РґРѕСЃС‚Р°РІРєРё: in_app, telegram"),
		field.Bool("enabled").
			Default(true).
			Comment("Р’РєР»СЋС‡РµРЅРѕ Р»Рё СѓРІРµРґРѕРјР»РµРЅРёРµ РґР»СЏ РґР°РЅРЅРѕРіРѕ С‚РёРїР° Рё РєР°РЅР°Р»Р°"),
		field.UUID("project_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID РїСЂРѕРµРєС‚Р° РґР»СЏ per-project РЅР°СЃС‚СЂРѕР№РєРё (nil = РіР»РѕР±Р°Р»СЊРЅР°СЏ)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё NotificationPreference.
func (NotificationPreference) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("notification_preferences").
			Field("user_id").
			Required().
			Unique(),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё NotificationPreference.
func (NotificationPreference) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "notification_type", "channel", "project_id").
			Unique(),
	}
}
