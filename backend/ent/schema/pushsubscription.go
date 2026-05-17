package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PushSubscription С…СЂР°РЅРёС‚ РїРѕРґРїРёСЃРєСѓ Р±СЂР°СѓР·РµСЂР° РЅР° Web Push СѓРІРµРґРѕРјР»РµРЅРёСЏ.
// РћРґРёРЅ РїРѕР»СЊР·РѕРІР°С‚РµР»СЊ РјРѕР¶РµС‚ РёРјРµС‚СЊ РЅРµСЃРєРѕР»СЊРєРѕ РїРѕРґРїРёСЃРѕРє (СЂР°Р·РЅС‹Рµ Р±СЂР°СѓР·РµСЂС‹/СѓСЃС‚СЂРѕР№СЃС‚РІР°).
type PushSubscription struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё PushSubscription.
func (PushSubscription) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ-РІР»Р°РґРµР»СЊС†Р° РїРѕРґРїРёСЃРєРё"),
		field.String("endpoint").
			NotEmpty().
			Comment("Push endpoint URL (СѓРЅРёРєР°Р»СЊРЅС‹Р№ РґР»СЏ РєР°Р¶РґРѕРіРѕ Р±СЂР°СѓР·РµСЂР°)"),
		field.String("p256dh_key").
			NotEmpty().
			Comment("РџСѓР±Р»РёС‡РЅС‹Р№ РєР»СЋС‡ РєР»РёРµРЅС‚Р° (base64url, RFC 8291)"),
		field.String("auth_key").
			NotEmpty().
			Comment("Shared auth secret (base64url, RFC 8291)"),
		field.String("user_agent").
			Optional().
			Comment("User-Agent Р±СЂР°СѓР·РµСЂР° РґР»СЏ РёРґРµРЅС‚РёС„РёРєР°С†РёРё СѓСЃС‚СЂРѕР№СЃС‚РІР°"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё PushSubscription.
func (PushSubscription) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("push_subscriptions").
			Field("user_id").
			Required().
			Unique(),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё PushSubscription.
func (PushSubscription) Indexes() []ent.Index {
	return []ent.Index{
		// РЈРЅРёРєР°Р»СЊРЅС‹Р№ endpoint вЂ” РѕРґРёРЅ Р±СЂР°СѓР·РµСЂ = РѕРґРЅР° РїРѕРґРїРёСЃРєР°.
		index.Fields("endpoint").Unique(),
		// Р‘С‹СЃС‚СЂС‹Р№ РїРѕРёСЃРє РїРѕРґРїРёСЃРѕРє РїРѕ РїРѕР»СЊР·РѕРІР°С‚РµР»СЋ.
		index.Fields("user_id"),
	}
}
