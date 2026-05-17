package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// SavedFilter С…СЂР°РЅРёС‚ СЃРѕС…СЂР°РЅС‘РЅРЅС‹Рµ С„РёР»СЊС‚СЂС‹ Р·Р°РґР°С‡ С‚СЂРµРєРµСЂР°.
// РџРѕР»СЊР·РѕРІР°С‚РµР»СЊ РјРѕР¶РµС‚ СЃРѕР·РґР°РІР°С‚СЊ, СЃРѕС…СЂР°РЅСЏС‚СЊ Рё СЂР°СЃС€Р°СЂРёРІР°С‚СЊ С„РёР»СЊС‚СЂС‹ РІ СЂР°РјРєР°С… РїСЂРѕРµРєС‚Р°.
type SavedFilter struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃРѕС…СЂР°РЅС‘РЅРЅРѕРіРѕ С„РёР»СЊС‚СЂР°.
func (SavedFilter) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(200).
			Comment("РќР°Р·РІР°РЅРёРµ С„РёР»СЊС‚СЂР°"),
		field.JSON("filter_config", map[string]interface{}{}).
			Comment("JSON-РєРѕРЅС„РёРіСѓСЂР°С†РёСЏ С„РёР»СЊС‚СЂР° (СЃС‚Р°С‚СѓСЃ, РїСЂРёРѕСЂРёС‚РµС‚, РёСЃРїРѕР»РЅРёС‚РµР»СЊ Рё С‚.Рґ.)"),
		field.Bool("shared").
			Default(false).
			Comment("Р Р°СЃС€Р°СЂРµРЅ Р»Рё С„РёР»СЊС‚СЂ РґР»СЏ РІСЃРµС… СѓС‡Р°СЃС‚РЅРёРєРѕРІ РїСЂРѕРµРєС‚Р°"),
		field.UUID("owner_id", uuid.UUID{}).
			Comment("ID РІР»Р°РґРµР»СЊС†Р° С„РёР»СЊС‚СЂР°"),
		field.UUID("project_id", uuid.UUID{}).
			Comment("ID РїСЂРѕРµРєС‚Р°, Рє РєРѕС‚РѕСЂРѕРјСѓ РїСЂРёРІСЏР·Р°РЅ С„РёР»СЊС‚СЂ"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃРѕС…СЂР°РЅС‘РЅРЅРѕРіРѕ С„РёР»СЊС‚СЂР°.
func (SavedFilter) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("saved_filters").
			Field("owner_id").
			Required().
			Unique(),
		edge.From("project", Project.Type).
			Ref("saved_filters").
			Field("project_id").
			Required().
			Unique(),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ РґР»СЏ СЃРѕС…СЂР°РЅС‘РЅРЅС‹С… С„РёР»СЊС‚СЂРѕРІ.
func (SavedFilter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id", "project_id"),
		index.Fields("project_id", "shared"),
	}
}
