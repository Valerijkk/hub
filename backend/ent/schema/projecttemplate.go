package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// ProjectTemplate — определение Ent-схемы сущности ProjectTemplate.
type ProjectTemplate struct {
	ent.Schema
}

// Поля сущности ProjectTemplate.
func (ProjectTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("РќР°Р·РІР°РЅРёРµ С€Р°Р±Р»РѕРЅР°"),
		field.Text("description").
			Optional().
			Comment("РћРїРёСЃР°РЅРёРµ С€Р°Р±Р»РѕРЅР°"),
		field.Enum("template_type").
			Values("empty", "software_dev", "qa", "kanban_simple", "custom").
			Default("custom").
			Comment("РўРёРї С€Р°Р±Р»РѕРЅР°"),
		field.JSON("config", map[string]interface{}{}).
			Comment("РЎРЅРёРјРѕРє РєРѕРЅС„РёРіСѓСЂР°С†РёРё: statuses, labels, boards, custom_fields"),
		field.Bool("is_system").
			Default(false).
			Comment("РЎРёСЃС‚РµРјРЅС‹Р№ С€Р°Р±Р»РѕРЅ (РЅРµР»СЊР·СЏ СѓРґР°Р»РёС‚СЊ)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности ProjectTemplate.
func (ProjectTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("created_by", User.Type).
			Ref("project_templates").
			Unique(),
	}
}

// Индексы сущности ProjectTemplate.
func (ProjectTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("template_type"),
		index.Fields("is_system"),
	}
}
