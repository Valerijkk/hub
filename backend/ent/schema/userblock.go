package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// UserBlock С…СЂР°РЅРёС‚ Р±Р»РѕРєРёСЂРѕРІРєСѓ РѕРґРЅРѕРіРѕ РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ РґСЂСѓРіРёРј.
// Р—Р°Р±Р»РѕРєРёСЂРѕРІР°РЅРЅС‹Р№ РїРѕР»СЊР·РѕРІР°С‚РµР»СЊ РЅРµ РјРѕР¶РµС‚ РґРѕР±Р°РІРёС‚СЊ Р±Р»РѕРєРёСЂСѓСЋС‰РµРіРѕ РІ РєР°РЅР°Р»С‹
// Рё РЅРµ РјРѕР¶РµС‚ РѕС‚РїСЂР°РІРёС‚СЊ РµРјСѓ DM.
type UserBlock struct {
	ent.Schema
}

// Fields of the UserBlock.
func (UserBlock) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("blocker_id", uuid.UUID{}).
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ, РєРѕС‚РѕСЂС‹Р№ Р±Р»РѕРєРёСЂСѓРµС‚"),
		field.UUID("blocked_id", uuid.UUID{}).
			Comment("ID Р·Р°Р±Р»РѕРєРёСЂРѕРІР°РЅРЅРѕРіРѕ РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges of the UserBlock.
func (UserBlock) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("blocker", User.Type).
			Ref("blocks_created").
			Field("blocker_id").
			Unique().
			Required(),
		edge.From("blocked", User.Type).
			Ref("blocks_received").
			Field("blocked_id").
			Unique().
			Required(),
	}
}

// Indexes of the UserBlock.
func (UserBlock) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("blocker_id", "blocked_id").
			Unique(),
	}
}
