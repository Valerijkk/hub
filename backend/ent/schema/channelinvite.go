package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ChannelInvite СЃРѕРґРµСЂР¶РёС‚ РѕРїСЂРµРґРµР»РµРЅРёРµ СЃС…РµРјС‹ РґР»СЏ РїСЂРёРіР»Р°СЃРёС‚РµР»СЊРЅРѕР№ СЃСЃС‹Р»РєРё РєР°РЅР°Р»Р°.
// РђРЅР°Р»РѕРі invite-СЃСЃС‹Р»РѕРє Telegram: РіРµРЅРµСЂРёСЂСѓРµС‚СЃСЏ С‚РѕРєРµРЅ, РїРѕ РєРѕС‚РѕСЂРѕРјСѓ РјРѕР¶РЅРѕ РІСЃС‚СѓРїРёС‚СЊ РІ РєР°РЅР°Р».
type ChannelInvite struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё ChannelInvite.
func (ChannelInvite) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}).
			Comment("ID РєР°РЅР°Р»Р°, Рє РєРѕС‚РѕСЂРѕРјСѓ РѕС‚РЅРѕСЃРёС‚СЃСЏ РїСЂРёРіР»Р°С€РµРЅРёРµ"),
		field.UUID("creator_id", uuid.UUID{}).
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ, СЃРѕР·РґР°РІС€РµРіРѕ РїСЂРёРіР»Р°С€РµРЅРёРµ"),
		field.String("token").
			Unique().
			NotEmpty().
			MaxLen(32).
			Comment("РЈРЅРёРєР°Р»СЊРЅС‹Р№ С‚РѕРєРµРЅ РїСЂРёРіР»Р°С€РµРЅРёСЏ (base62, 22 СЃРёРјРІРѕР»Р°)"),
		field.Int("max_uses").
			Default(0).
			Comment("РњР°РєСЃРёРјР°Р»СЊРЅРѕРµ С‡РёСЃР»Рѕ РёСЃРїРѕР»СЊР·РѕРІР°РЅРёР№ (0 = Р±РµР· Р»РёРјРёС‚Р°)"),
		field.Int("use_count").
			Default(0).
			Comment("РўРµРєСѓС‰РµРµ С‡РёСЃР»Рѕ РёСЃРїРѕР»СЊР·РѕРІР°РЅРёР№"),
		field.Time("expires_at").
			Optional().
			Nillable().
			Comment("Р’СЂРµРјСЏ РёСЃС‚РµС‡РµРЅРёСЏ СЃСЃС‹Р»РєРё (nil = Р±РµСЃСЃСЂРѕС‡РЅР°СЏ)"),
		field.Bool("is_revoked").
			Default(false).
			Comment("Р¤Р»Р°Рі РѕС‚Р·С‹РІР° РїСЂРёРіР»Р°С€РµРЅРёСЏ"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё ChannelInvite.
func (ChannelInvite) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", Channel.Type).
			Ref("invites").
			Field("channel_id").
			Required().
			Unique(),
		edge.From("creator", User.Type).
			Ref("channel_invites").
			Field("creator_id").
			Required().
			Unique(),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё ChannelInvite.
func (ChannelInvite) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id"),
		index.Fields("token").Unique(),
		index.Fields("expires_at"),
	}
}
