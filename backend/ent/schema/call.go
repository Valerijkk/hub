package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Call С…СЂР°РЅРёС‚ РјРµС‚Р°РґР°РЅРЅС‹Рµ РіРѕР»РѕСЃРѕРІРѕРіРѕ Р·РІРѕРЅРєР°.
// РљР°Р¶РґС‹Р№ Р·РІРѕРЅРѕРє РїСЂРёРІСЏР·Р°РЅ Рє РєР°РЅР°Р»Сѓ РјРµСЃСЃРµРЅРґР¶РµСЂР° Рё СЃРѕРґРµСЂР¶РёС‚ РёРЅС„РѕСЂРјР°С†РёСЋ
// Рѕ РєРѕРјРЅР°С‚Рµ LiveKit, СЃС‚Р°С‚СѓСЃРµ Рё РєРѕР»РёС‡РµСЃС‚РІРµ СѓС‡Р°СЃС‚РЅРёРєРѕРІ.
type Call struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё Call.
func (Call) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("room_name").
			Unique().
			NotEmpty().
			Comment("РЈРЅРёРєР°Р»СЊРЅРѕРµ РёРјСЏ РєРѕРјРЅР°С‚С‹ LiveKit"),
		field.UUID("channel_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID РєР°РЅР°Р»Р° РјРµСЃСЃРµРЅРґР¶РµСЂР° (nil РґР»СЏ instant-Р·РІРѕРЅРєРѕРІ)"),
		field.String("invite_code").
			Optional().
			Unique().
			MaxLen(12).
			Comment("РљРѕРґ РїСЂРёРіР»Р°С€РµРЅРёСЏ РґР»СЏ instant-Р·РІРѕРЅРєР° (РЅР°РїСЂРёРјРµСЂ hub-abc123)"),
		field.String("title").
			Optional().
			MaxLen(255).
			Comment("РќР°Р·РІР°РЅРёРµ Р·РІРѕРЅРєР° (РґР»СЏ instant-Р·РІРѕРЅРєРѕРІ Р±РµР· РєР°РЅР°Р»Р°)"),
		field.UUID("created_by", uuid.UUID{}).
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ, СЃРѕР·РґР°РІС€РµРіРѕ Р·РІРѕРЅРѕРє"),
		field.Enum("status").
			Values("active", "ended").
			Default("active").
			Comment("РЎС‚Р°С‚СѓСЃ Р·РІРѕРЅРєР°: Р°РєС‚РёРІРµРЅ РёР»Рё Р·Р°РІРµСЂС€С‘РЅ"),
		field.Time("started_at").
			Optional().
			Nillable().
			Comment("Р’СЂРµРјСЏ РЅР°С‡Р°Р»Р° Р·РІРѕРЅРєР° (СѓСЃС‚Р°РЅР°РІР»РёРІР°РµС‚СЃСЏ РїРѕ webhook room_started)"),
		field.Time("ended_at").
			Optional().
			Nillable().
			Comment("Р’СЂРµРјСЏ Р·Р°РІРµСЂС€РµРЅРёСЏ Р·РІРѕРЅРєР°"),
		field.Int("participant_count").
			Default(0).
			Comment("РўРµРєСѓС‰РµРµ РєРѕР»РёС‡РµСЃС‚РІРѕ СѓС‡Р°СЃС‚РЅРёРєРѕРІ РІ Р·РІРѕРЅРєРµ"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё Call.
func (Call) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", Channel.Type).
			Ref("calls").
			Field("channel_id").
			Unique(),
		edge.From("creator", User.Type).
			Ref("created_calls").
			Field("created_by").
			Unique().
			Required(),
		edge.To("participants", CallParticipant.Type),
		edge.To("recordings", Recording.Type),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё Call.
func (Call) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "status"),
		index.Fields("room_name"),
		index.Fields("invite_code"),
	}
}
