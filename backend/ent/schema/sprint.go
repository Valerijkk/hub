package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Sprint СЃРѕРґРµСЂР¶РёС‚ РѕРїСЂРµРґРµР»РµРЅРёРµ СЃС…РµРјС‹ СЃСѓС‰РЅРѕСЃС‚Рё Sprint.
type Sprint struct {
	ent.Schema
}

// Fields РѕРїСЂРµРґРµР»СЏРµС‚ РїРѕР»СЏ СЃСѓС‰РЅРѕСЃС‚Рё Sprint.
func (Sprint) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(200),
		field.String("goal").
			Optional().
			MaxLen(1000).
			Comment("Р¦РµР»СЊ СЃРїСЂРёРЅС‚Р°"),
		field.Enum("status").
			Values("planning", "active", "completed").
			Default("planning"),
		field.Time("start_date"),
		field.Time("end_date"),
		field.JSON("retro_notes", map[string]interface{}{}).
			Optional().
			Comment("JSONB Р·Р°РјРµС‚РєРё СЂРµС‚СЂРѕСЃРїРµРєС‚РёРІС‹: went_well, to_improve, actions"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Edges РѕРїСЂРµРґРµР»СЏРµС‚ СЃРІСЏР·Рё СЃСѓС‰РЅРѕСЃС‚Рё Sprint.
func (Sprint) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("sprints").
			Unique().
			Required(),
		edge.To("issues", Issue.Type),
	}
}

// Indexes РѕРїСЂРµРґРµР»СЏРµС‚ РёРЅРґРµРєСЃС‹ СЃСѓС‰РЅРѕСЃС‚Рё Sprint.
func (Sprint) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("project").
			Fields("status"),
		index.Edges("project").
			Fields("start_date"),
	}
}
