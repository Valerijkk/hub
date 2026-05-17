package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GitLabIntegration С…СЂР°РЅРёС‚ РЅР°СЃС‚СЂРѕР№РєРё РёРЅС‚РµРіСЂР°С†РёРё РїСЂРѕРµРєС‚Р° СЃ GitLab-СЂРµРїРѕР·РёС‚РѕСЂРёРµРј.
type GitLabIntegration struct {
	ent.Schema
}

// Fields of the GitLabIntegration.
func (GitLabIntegration) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("gitlab_url").
			Default("https://gitlab.com").
			MaxLen(500).
			Comment("URL РёРЅСЃС‚Р°РЅСЃР° GitLab (РїРѕ СѓРјРѕР»С‡Р°РЅРёСЋ https://gitlab.com)"),
		field.String("project_path").
			NotEmpty().
			MaxLen(500).
			Comment("РџСѓС‚СЊ Рє РїСЂРѕРµРєС‚Сѓ РІ GitLab (РЅР°РїСЂРёРјРµСЂ group/repo)"),
		field.String("access_token").
			Sensitive().
			Comment("GitLab Personal Access Token"),
		field.String("webhook_secret").
			Optional().
			Comment("РЎРµРєСЂРµС‚ РґР»СЏ РІР°Р»РёРґР°С†РёРё GitLab webhook (X-Gitlab-Token)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges of the GitLabIntegration.
func (GitLabIntegration) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("gitlab_integrations").
			Unique().
			Required(),
		edge.To("gitlab_links", GitLabLink.Type),
	}
}

// Indexes of the GitLabIntegration.
func (GitLabIntegration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_path").
			Edges("project").
			Unique(),
	}
}
