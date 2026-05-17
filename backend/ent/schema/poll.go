package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Poll СЃРѕРґРµСЂР¶РёС‚ РѕРїСЂРµРґРµР»РµРЅРёРµ СЃС…РµРјС‹ РґР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё РіРѕР»РѕСЃРѕРІР°РЅРёСЏ (РѕРїСЂРѕСЃР°).
type Poll struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё Poll.
func (Poll) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("channel_id", uuid.UUID{}).
			Comment("ID РєР°РЅР°Р»Р°, РІ РєРѕС‚РѕСЂРѕРј СЃРѕР·РґР°РЅ РѕРїСЂРѕСЃ"),
		field.UUID("creator_id", uuid.UUID{}).
			Comment("ID СЃРѕР·РґР°С‚РµР»СЏ РѕРїСЂРѕСЃР°"),
		field.String("question").
			NotEmpty().
			MaxLen(500).
			Comment("РўРµРєСЃС‚ РІРѕРїСЂРѕСЃР° РѕРїСЂРѕСЃР°"),
		field.Bool("is_anonymous").
			Default(false).
			Comment("РђРЅРѕРЅРёРјРЅРѕРµ РіРѕР»РѕСЃРѕРІР°РЅРёРµ (РЅРµ РїРѕРєР°Р·С‹РІР°С‚СЊ РєС‚Рѕ РіРѕР»РѕСЃРѕРІР°Р»)"),
		field.Bool("is_multi_choice").
			Default(false).
			Comment("РњРЅРѕР¶РµСЃС‚РІРµРЅРЅС‹Р№ РІС‹Р±РѕСЂ (РјРѕР¶РЅРѕ РІС‹Р±СЂР°С‚СЊ РЅРµСЃРєРѕР»СЊРєРѕ РІР°СЂРёР°РЅС‚РѕРІ)"),
		field.Bool("is_closed").
			Default(false).
			Comment("Р—Р°РєСЂС‹С‚ Р»Рё РѕРїСЂРѕСЃ РґР»СЏ РіРѕР»РѕСЃРѕРІР°РЅРёСЏ"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё Poll.
func (Poll) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", Channel.Type).
			Ref("polls").
			Field("channel_id").
			Unique().
			Required(),
		edge.From("creator", User.Type).
			Ref("polls").
			Field("creator_id").
			Unique().
			Required(),
		edge.To("options", PollOption.Type),
		edge.To("votes", PollVote.Type),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё Poll.
func (Poll) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id"),
		index.Fields("creator_id"),
	}
}
