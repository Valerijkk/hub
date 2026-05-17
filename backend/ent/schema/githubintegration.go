package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GitHubIntegration С…СЂР°РЅРёС‚ РЅР°СЃС‚СЂРѕР№РєРё РёРЅС‚РµРіСЂР°С†РёРё РїСЂРѕРµРєС‚Р° СЃ GitHub-СЂРµРїРѕР·РёС‚РѕСЂРёРµРј.
type GitHubIntegration struct {
	ent.Schema
}

// Fields of the GitHubIntegration.
func (GitHubIntegration) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("repo_owner").
			NotEmpty().
			MaxLen(255).
			Comment("Р’Р»Р°РґРµР»РµС† СЂРµРїРѕР·РёС‚РѕСЂРёСЏ (РѕСЂРіР°РЅРёР·Р°С†РёСЏ РёР»Рё РїРѕР»СЊР·РѕРІР°С‚РµР»СЊ)"),
		field.String("repo_name").
			NotEmpty().
			MaxLen(255).
			Comment("РРјСЏ СЂРµРїРѕР·РёС‚РѕСЂРёСЏ"),
		field.String("access_token").
			Sensitive().
			Comment("GitHub Personal Access Token"),
		field.String("webhook_secret").
			Comment("РЎРµРєСЂРµС‚ РґР»СЏ РІР°Р»РёРґР°С†РёРё webhook РїРѕРґРїРёСЃРµР№"),
		field.UUID("merge_status_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Р¦РµР»РµРІРѕР№ СЃС‚Р°С‚СѓСЃ Р·Р°РґР°С‡Рё РїСЂРё РјРµСЂР¶Рµ PR (GHUB-04)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges of the GitHubIntegration.
func (GitHubIntegration) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("github_integrations").
			Unique().
			Required(),
		edge.To("github_links", GitHubLink.Type),
	}
}

// Indexes of the GitHubIntegration.
func (GitHubIntegration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("repo_owner", "repo_name").
			Edges("project").
			Unique(),
	}
}
