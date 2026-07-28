package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GitHubIntegration хранит настройки интеграции проекта с GitHub-репозиторием.
type GitHubIntegration struct {
	ent.Schema
}

// Поля сущности GitHubIntegration.
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
			Sensitive().
			NotEmpty().
			Comment("Секрет для валидации webhook подписей"),
		field.UUID("merge_status_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("Целевой статус задачи при мерже PR (GHUB-04)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (GitHubIntegration) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Рёбра (edges) сущности GitHubIntegration.
func (GitHubIntegration) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("github_integrations").
			Unique().
			Required(),
		edge.To("github_links", GitHubLink.Type),
	}
}

// Индексы сущности GitHubIntegration.
func (GitHubIntegration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("repo_owner", "repo_name").
			Edges("project").
			Unique(),
	}
}
