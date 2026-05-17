package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Bookmark СЃРѕРґРµСЂР¶РёС‚ РѕРїСЂРµРґРµР»РµРЅРёРµ СЃС…РµРјС‹ РґР»СЏ Р·Р°РєР»Р°РґРєРё РЅР° СЃРѕРѕР±С‰РµРЅРёРµ.
type Bookmark struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё Bookmark.
func (Bookmark) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ, СЃРѕР·РґР°РІС€РµРіРѕ Р·Р°РєР»Р°РґРєСѓ"),
		field.String("message_id").
			NotEmpty().
			Comment("ID СЃРѕРѕР±С‰РµРЅРёСЏ РІ MongoDB"),
		field.UUID("channel_id", uuid.UUID{}).
			Comment("ID РєР°РЅР°Р»Р° СЃРѕРѕР±С‰РµРЅРёСЏ"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё Bookmark.
func (Bookmark) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("bookmarks").
			Field("user_id").
			Required().
			Unique(),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё Bookmark.
func (Bookmark) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "message_id").Unique(),
	}
}
