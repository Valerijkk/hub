package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// DashboardLayout С…СЂР°РЅРёС‚ СЂР°СЃРєР»Р°РґРєСѓ РІРёРґР¶РµС‚РѕРІ РґР°С€Р±РѕСЂРґР° РїРѕР»СЊР·РѕРІР°С‚РµР»СЏ.
type DashboardLayout struct {
	ent.Schema
}

// Fields РѕРїРёСЃС‹РІР°РµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё DashboardLayout.
func (DashboardLayout) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID РІР»Р°РґРµР»СЊС†Р° СЂР°СЃРєР»Р°РґРєРё"),
		field.Enum("dashboard_type").
			Values("personal", "project", "admin").
			Comment("РўРёРї РґР°С€Р±РѕСЂРґР°: personal, project, admin"),
		field.UUID("project_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID РїСЂРѕРµРєС‚Р° (С‚РѕР»СЊРєРѕ РґР»СЏ project-РґР°С€Р±РѕСЂРґР°)"),
		field.JSON("items", []map[string]interface{}{}).
			Comment("JSON-РјР°СЃСЃРёРІ РІРёРґР¶РµС‚РѕРІ СЃ РїРѕР·РёС†РёРµР№ Рё СЂР°Р·РјРµСЂР°РјРё"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Edges РѕРїРёСЃС‹РІР°РµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё DashboardLayout.
func (DashboardLayout) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("dashboard_layouts").
			Field("user_id").
			Required().
			Unique(),
	}
}

// Indexes РѕРїРёСЃС‹РІР°РµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё DashboardLayout.
func (DashboardLayout) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "dashboard_type", "project_id").
			Unique(),
	}
}
