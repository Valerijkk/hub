package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Recording С…СЂР°РЅРёС‚ РјРµС‚Р°РґР°РЅРЅС‹Рµ Р·Р°РїРёСЃРё Р·РІРѕРЅРєР°.
// Р—Р°РїРёСЃСЊ СЃРѕР·РґР°С‘С‚СЃСЏ С‡РµСЂРµР· LiveKit Egress Рё СЃРѕС…СЂР°РЅСЏРµС‚СЃСЏ РІ SeaweedFS (S3).
// РџСЂРёРІСЏР·Р°РЅР° Рє Р·РІРѕРЅРєСѓ, РїСЂРѕРµРєС‚Сѓ Рё РїРѕР»СЊР·РѕРІР°С‚РµР»СЋ, РЅР°С‡Р°РІС€РµРјСѓ Р·Р°РїРёСЃСЊ.
type Recording struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё Recording.
func (Recording) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("egress_id").
			NotEmpty().
			Comment("ID egress-РїСЂРѕС†РµСЃСЃР° РІ LiveKit"),
		field.Enum("status").
			Values("recording", "processing", "ready", "failed").
			Default("recording").
			Comment("РЎС‚Р°С‚СѓСЃ Р·Р°РїРёСЃРё: recording -> processing -> ready/failed"),
		field.String("file_path").
			Optional().
			Comment("РљР»СЋС‡ С„Р°Р№Р»Р° РІ S3 (recordings/{projectID}/{callID}.mp4)"),
		field.Int64("file_size").
			Default(0).
			Comment("Р Р°Р·РјРµСЂ С„Р°Р№Р»Р° Р·Р°РїРёСЃРё РІ Р±Р°Р№С‚Р°С…"),
		field.Int("duration_seconds").
			Default(0).
			Comment("Р”Р»РёС‚РµР»СЊРЅРѕСЃС‚СЊ Р·Р°РїРёСЃРё РІ СЃРµРєСѓРЅРґР°С…"),
		field.Time("started_at").
			Default(nowUTC).
			Comment("Р’СЂРµРјСЏ РЅР°С‡Р°Р»Р° Р·Р°РїРёСЃРё"),
		field.Time("ended_at").
			Optional().
			Nillable().
			Comment("Р’СЂРµРјСЏ РѕРєРѕРЅС‡Р°РЅРёСЏ Р·Р°РїРёСЃРё"),
		field.UUID("call_id", uuid.UUID{}).
			Comment("ID Р·РІРѕРЅРєР°, Рє РєРѕС‚РѕСЂРѕРјСѓ РѕС‚РЅРѕСЃРёС‚СЃСЏ Р·Р°РїРёСЃСЊ"),
		field.UUID("project_id", uuid.UUID{}).
			Comment("ID РїСЂРѕРµРєС‚Р° РґР»СЏ РїСЂРёРІСЏР·РєРё Р·Р°РїРёСЃРµР№ Рє РїСЂРѕРµРєС‚Сѓ"),
		field.UUID("started_by", uuid.UUID{}).
			Comment("ID РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ, РЅР°С‡Р°РІС€РµРіРѕ Р·Р°РїРёСЃСЊ"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё Recording.
func (Recording) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("call", Call.Type).
			Ref("recordings").
			Field("call_id").
			Unique().
			Required(),
		edge.From("project", Project.Type).
			Ref("recordings").
			Field("project_id").
			Unique().
			Required(),
		edge.From("user", User.Type).
			Ref("recordings").
			Field("started_by").
			Unique().
			Required(),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё Recording.
func (Recording) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id", "created_at"),
		index.Fields("call_id"),
		index.Fields("egress_id"),
	}
}
