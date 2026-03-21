package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// GitHubIntegration хранит настройки интеграции проекта с GitHub-репозиторием.
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
			Comment("Владелец репозитория (организация или пользователь)"),
		field.String("repo_name").
			NotEmpty().
			MaxLen(255).
			Comment("Имя репозитория"),
		field.String("access_token").
			Sensitive().
			Comment("GitHub Personal Access Token"),
		field.String("webhook_secret").
			Comment("Секрет для валидации webhook подписей"),
		field.UUID("merge_status_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Целевой статус задачи при мерже PR (GHUB-04)"),
		field.Time("created_at").
			Default(time.Now).
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
