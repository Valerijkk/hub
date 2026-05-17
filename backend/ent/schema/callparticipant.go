package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CallParticipant С…СЂР°РЅРёС‚ РёРЅС„РѕСЂРјР°С†РёСЋ РѕР± СѓС‡Р°СЃС‚РЅРёРєРµ РіРѕР»РѕСЃРѕРІРѕРіРѕ Р·РІРѕРЅРєР°.
// РЎРѕРґРµСЂР¶РёС‚ РІСЂРµРјСЏ РїСЂРёСЃРѕРµРґРёРЅРµРЅРёСЏ Рё СѓС…РѕРґР° СѓС‡Р°СЃС‚РЅРёРєР° РёР· Р·РІРѕРЅРєР°.
type CallParticipant struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё CallParticipant.
func (CallParticipant) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("call_id", uuid.UUID{}).
			Comment("ID Р·РІРѕРЅРєР°"),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ-СѓС‡Р°СЃС‚РЅРёРєР°"),
		field.Time("joined_at").
			Default(nowUTC).
			Comment("Р’СЂРµРјСЏ РїСЂРёСЃРѕРµРґРёРЅРµРЅРёСЏ Рє Р·РІРѕРЅРєСѓ"),
		field.Time("left_at").
			Optional().
			Nillable().
			Comment("Р’СЂРµРјСЏ РІС‹С…РѕРґР° РёР· Р·РІРѕРЅРєР° (nil РµСЃР»Рё РµС‰С‘ РІ Р·РІРѕРЅРєРµ)"),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё CallParticipant.
func (CallParticipant) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("call", Call.Type).
			Ref("participants").
			Field("call_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("call_participations").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё CallParticipant.
func (CallParticipant) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("call_id", "user_id").
			Unique(),
	}
}
