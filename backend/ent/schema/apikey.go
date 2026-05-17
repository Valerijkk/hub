package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// APIKey holds the schema definition for the APIKey entity.
type APIKey struct {
	ent.Schema
}

// Fields of the APIKey.
func (APIKey) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("Р§РµР»РѕРІРµРєРѕС‡РёС‚Р°РµРјРѕРµ РёРјСЏ РєР»СЋС‡Р°"),
		field.String("key_hash").
			Unique().
			NotEmpty().
			Comment("SHA-256 С…РµС€ API-РєР»СЋС‡Р°"),
		field.String("prefix").
			NotEmpty().
			MaxLen(8).
			Comment("РџРµСЂРІС‹Рµ 8 СЃРёРјРІРѕР»РѕРІ РєР»СЋС‡Р° РґР»СЏ РѕС‚РѕР±СЂР°Р¶РµРЅРёСЏ"),
		field.Time("last_used_at").
			Optional().
			Nillable().
			Comment("Р’СЂРµРјСЏ РїРѕСЃР»РµРґРЅРµРіРѕ РёСЃРїРѕР»СЊР·РѕРІР°РЅРёСЏ"),
		field.Time("expires_at").
			Optional().
			Nillable().
			Comment("РЎСЂРѕРє РґРµР№СЃС‚РІРёСЏ РєР»СЋС‡Р°"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges of the APIKey.
func (APIKey) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("api_keys").
			Unique().
			Required(),
	}
}

// Indexes of the APIKey.
func (APIKey) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("key_hash"),
		index.Fields("expires_at"),
	}
}
