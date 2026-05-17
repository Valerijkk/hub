package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ChannelMember СЃРѕРґРµСЂР¶РёС‚ РѕРїСЂРµРґРµР»РµРЅРёРµ СЃС…РµРјС‹ РґР»СЏ СѓС‡Р°СЃС‚РЅРёРєР° РєР°РЅР°Р»Р°.
type ChannelMember struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё ChannelMember.
func (ChannelMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}).
			Comment("ID РєР°РЅР°Р»Р°"),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ"),
		field.Enum("role").
			Values("owner", "admin", "member").
			Default("member").
			Comment("Р РѕР»СЊ СѓС‡Р°СЃС‚РЅРёРєР° РІ РєР°РЅР°Р»Рµ"),
		field.Time("joined_at").
			Default(nowUTC).
			Comment("Р’СЂРµРјСЏ РІСЃС‚СѓРїР»РµРЅРёСЏ РІ РєР°РЅР°Р»"),
		field.Bool("is_starred").
			Default(false).
			Comment("Р¤Р»Р°Рі Р·Р°РєСЂРµРїР»РµРЅРёСЏ (starred) РєР°РЅР°Р»Р° РїРѕР»СЊР·РѕРІР°С‚РµР»РµРј"),
		field.Time("muted_until").
			Optional().
			Nillable().
			Comment("РљР°РЅР°Р» Р·Р°РјСЊСЋС‡РµРЅ РґРѕ СЌС‚РѕРіРѕ РІСЂРµРјРµРЅРё (nil = РЅРµ Р·Р°РјСЊСЋС‡РµРЅ)"),
		field.Enum("notify_mode").
			Values("all", "mentions", "nothing").
			Default("all").
			Comment("Р РµР¶РёРј СѓРІРµРґРѕРјР»РµРЅРёР№: all=РІСЃС‘, mentions=С‚РѕР»СЊРєРѕ @СѓРїРѕРјРёРЅР°РЅРёСЏ, nothing=РЅРёС‡РµРіРѕ"),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё ChannelMember.
func (ChannelMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", Channel.Type).
			Ref("members").
			Field("channel_id").
			Required().
			Unique(),
		edge.From("user", User.Type).
			Ref("channel_memberships").
			Field("user_id").
			Required().
			Unique(),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё ChannelMember.
func (ChannelMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "user_id").Unique(),
	}
}
