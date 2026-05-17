package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GitHubLink С…СЂР°РЅРёС‚ СЃРІСЏР·СЊ PR РёР»Рё РєРѕРјРјРёС‚Р° РёР· GitHub СЃ Р·Р°РґР°С‡РµР№ С‚СЂРµРєРµСЂР°.
type GitHubLink struct {
	ent.Schema
}

// Fields of the GitHubLink.
func (GitHubLink) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Enum("link_type").
			Values("commit", "pull_request").
			Comment("РўРёРї СЃРІСЏР·Рё: РєРѕРјРјРёС‚ РёР»Рё pull request"),
		field.String("external_id").
			NotEmpty().
			Comment("GitHub PR number РёР»Рё commit SHA"),
		field.String("title").
			MaxLen(500).
			Comment("Р—Р°РіРѕР»РѕРІРѕРє PR РёР»Рё СЃРѕРѕР±С‰РµРЅРёРµ РєРѕРјРјРёС‚Р°"),
		field.String("url").
			Comment("РџРѕР»РЅС‹Р№ URL РЅР° GitHub"),
		field.Enum("status").
			Values("open", "merged", "closed", "none").
			Default("none").
			Comment("РЎС‚Р°С‚СѓСЃ PR (РґР»СЏ РєРѕРјРјРёС‚РѕРІ вЂ” none)"),
		field.String("author").
			Optional().
			Comment("РРјСЏ Р°РІС‚РѕСЂР° РЅР° GitHub"),
		field.String("branch").
			Optional().
			Comment("РРјСЏ РІРµС‚РєРё (РґР»СЏ PR)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Edges of the GitHubLink.
func (GitHubLink) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("issue", Issue.Type).
			Ref("github_links").
			Unique().
			Required(),
		edge.From("integration", GitHubIntegration.Type).
			Ref("github_links").
			Unique().
			Required(),
	}
}

// Indexes of the GitHubLink.
func (GitHubLink) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("issue"),
		index.Fields("external_id", "link_type"),
	}
}
