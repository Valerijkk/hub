package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Channel СЃРѕРґРµСЂР¶РёС‚ РѕРїСЂРµРґРµР»РµРЅРёРµ СЃС…РµРјС‹ РґР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё Channel (РєР°РЅР°Р» РјРµСЃСЃРµРЅРґР¶РµСЂР°).
type Channel struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё Channel.
func (Channel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Enum("channel_type").
			Values("dm", "group", "public", "private").
			Comment("РўРёРї РєР°РЅР°Р»Р°: Р»РёС‡РЅС‹Р№ С‡Р°С‚, РіСЂСѓРїРїРѕРІРѕР№, РїСѓР±Р»РёС‡РЅС‹Р№ РёР»Рё РїСЂРёРІР°С‚РЅС‹Р№"),
		field.String("name").
			Optional().
			MaxLen(100).
			Comment("РќР°Р·РІР°РЅРёРµ РєР°РЅР°Р»Р° (РїСѓСЃС‚Рѕ РґР»СЏ DM)"),
		field.String("description").
			Optional().
			MaxLen(500).
			Comment("РћРїРёСЃР°РЅРёРµ РєР°РЅР°Р»Р°"),
		field.Bool("is_readonly").
			Default(false).
			Comment("Р¤Р»Р°Рі read-only РєР°РЅР°Р»Р° (РЅР°РїСЂРёРјРµСЂ, CI-СѓРІРµРґРѕРјР»РµРЅРёСЏ)"),
		field.UUID("project_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID РїСЂРѕРµРєС‚Р° (nil РґР»СЏ DM Рё РіР»РѕР±Р°Р»СЊРЅС‹С… РєР°РЅР°Р»РѕРІ)"),
		field.UUID("created_by", uuid.UUID{}).
			Comment("ID СЃРѕР·РґР°С‚РµР»СЏ РєР°РЅР°Р»Р°"),
		field.String("dm_key").
			Optional().
			Unique().
			Nillable().
			Comment("РљР»СЋС‡ РґРµРґСѓРїР»РёРєР°С†РёРё DM: РѕС‚СЃРѕСЂС‚РёСЂРѕРІР°РЅРЅС‹Рµ user IDs С‡РµСЂРµР· '-'"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё Channel.
func (Channel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("members", ChannelMember.Type),
		edge.From("project", Project.Type).
			Ref("channels").
			Field("project_id").
			Unique(),
		edge.From("creator", User.Type).
			Ref("created_channels").
			Field("created_by").
			Unique().
			Required(),
		edge.To("pinned_messages", PinnedMessage.Type),
		edge.To("calls", Call.Type),
		edge.To("scheduled_calls", ScheduledCall.Type),
		edge.To("polls", Poll.Type),
		edge.To("invites", ChannelInvite.Type),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё Channel.
func (Channel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_type"),
		index.Fields("project_id"),
	}
}
