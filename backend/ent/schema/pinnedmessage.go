package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PinnedMessage СЃРѕРґРµСЂР¶РёС‚ РѕРїСЂРµРґРµР»РµРЅРёРµ СЃС…РµРјС‹ РґР»СЏ Р·Р°РєСЂРµРїР»С‘РЅРЅРѕРіРѕ СЃРѕРѕР±С‰РµРЅРёСЏ РІ РєР°РЅР°Р»Рµ.
type PinnedMessage struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё PinnedMessage.
func (PinnedMessage) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("message_id").
			NotEmpty().
			Comment("ID СЃРѕРѕР±С‰РµРЅРёСЏ РІ MongoDB"),
		field.UUID("channel_id", uuid.UUID{}).
			Comment("ID РєР°РЅР°Р»Р°"),
		field.UUID("pinned_by", uuid.UUID{}).
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ, Р·Р°РєСЂРµРїРёРІС€РµРіРѕ СЃРѕРѕР±С‰РµРЅРёРµ"),
		field.Time("pinned_at").
			Default(nowUTC).
			Comment("Р’СЂРµРјСЏ Р·Р°РєСЂРµРїР»РµРЅРёСЏ"),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё PinnedMessage.
func (PinnedMessage) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("channel", Channel.Type).
			Ref("pinned_messages").
			Field("channel_id").
			Required().
			Unique(),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё PinnedMessage.
func (PinnedMessage) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("channel_id", "message_id").Unique(),
	}
}
