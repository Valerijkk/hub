package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GitLabLink С…СЂР°РЅРёС‚ СЃРІСЏР·СЊ MR РёР»Рё РєРѕРјРјРёС‚Р° РёР· GitLab СЃ Р·Р°РґР°С‡РµР№ С‚СЂРµРєРµСЂР°.
type GitLabLink struct {
	ent.Schema
}

// Поля сущности GitLabLink.
func (GitLabLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Enum("link_type").
			Values("commit", "merge_request").
			Comment("РўРёРї СЃРІСЏР·Рё: РєРѕРјРјРёС‚ РёР»Рё merge request"),
		field.String("external_id").
			NotEmpty().
			Comment("GitLab MR IID РёР»Рё commit SHA"),
		field.String("title").
			MaxLen(500).
			Comment("Р—Р°РіРѕР»РѕРІРѕРє MR РёР»Рё СЃРѕРѕР±С‰РµРЅРёРµ РєРѕРјРјРёС‚Р°"),
		field.String("url").
			Comment("РџРѕР»РЅС‹Р№ URL РЅР° GitLab"),
		field.Enum("status").
			Values("opened", "merged", "closed", "none").
			Default("none").
			Comment("РЎС‚Р°С‚СѓСЃ MR (РґР»СЏ РєРѕРјРјРёС‚РѕРІ вЂ” none)"),
		field.String("author").
			Optional().
			Comment("РРјСЏ Р°РІС‚РѕСЂР° РЅР° GitLab"),
		field.String("source_branch").
			Optional().
			Comment("РРјСЏ РёСЃС…РѕРґРЅРѕР№ РІРµС‚РєРё (РґР»СЏ MR)"),
		field.String("project_path").
			Optional().
			Comment("РџСѓС‚СЊ РїСЂРѕРµРєС‚Р° РІ GitLab (group/repo)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Рёбра (edges) сущности GitLabLink.
func (GitLabLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("issue", Issue.Type).
			Ref("gitlab_links").
			Unique().
			Required(),
		edge.From("integration", GitLabIntegration.Type).
			Ref("gitlab_links").
			Unique().
			Required(),
	}
}

// Индексы сущности GitLabLink.
func (GitLabLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("issue"),
		index.Fields("external_id", "link_type"),
	}
}
