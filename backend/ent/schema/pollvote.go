package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// PollVote СЃРѕРґРµСЂР¶РёС‚ РѕРїСЂРµРґРµР»РµРЅРёРµ СЃС…РµРјС‹ РґР»СЏ РіРѕР»РѕСЃР° РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ РІ РѕРїСЂРѕСЃРµ.
type PollVote struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё PollVote.
func (PollVote) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("poll_id", uuid.UUID{}).
			Comment("ID РѕРїСЂРѕСЃР°"),
		field.UUID("option_id", uuid.UUID{}).
			Comment("ID РІС‹Р±СЂР°РЅРЅРѕРіРѕ РІР°СЂРёР°РЅС‚Р°"),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID РїСЂРѕРіРѕР»РѕСЃРѕРІР°РІС€РµРіРѕ РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё PollVote.
func (PollVote) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("poll", Poll.Type).
			Ref("votes").
			Field("poll_id").
			Unique().
			Required(),
		edge.From("option", PollOption.Type).
			Ref("votes").
			Field("option_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("poll_votes").
			Field("user_id").
			Unique().
			Required(),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё PollVote.
func (PollVote) Indexes() []ent.Index {
	return []ent.Index{
		// РЈРЅРёРєР°Р»СЊРЅРѕРµ РѕРіСЂР°РЅРёС‡РµРЅРёРµ: РѕРґРёРЅ РіРѕР»РѕСЃ РЅР° РІР°СЂРёР°РЅС‚ РѕС‚ РѕРґРЅРѕРіРѕ РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ.
		index.Fields("poll_id", "option_id", "user_id").Unique(),
		index.Fields("poll_id", "user_id"),
	}
}
