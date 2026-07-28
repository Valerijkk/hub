package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GitLabIntegration хранит настройки интеграции проекта с GitLab-репозиторием.
type GitLabIntegration struct {
	ent.Schema
}

// Поля сущности GitLabIntegration.
func (GitLabIntegration) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("gitlab_url").
			Default("https://gitlab.com").
			MaxLen(500).
			Comment("URL инстанса GitLab (по умолчанию https://gitlab.com)"),
		field.String("project_path").
			NotEmpty().
			MaxLen(500).
			Comment("Путь к проекту в GitLab (например group/repo)"),
		field.String("access_token").
			Sensitive().
			Comment("GitLab Personal Access Token"),
		field.String("webhook_secret").
			Optional().
			Sensitive().
			Comment("Секрет для валидации GitLab webhook (X-Gitlab-Token)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (GitLabIntegration) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Рёбра (edges) сущности GitLabIntegration.
func (GitLabIntegration) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("gitlab_integrations").
			Unique().
			Required(),
		edge.To("gitlab_links", GitLabLink.Type),
	}
}

// Индексы сущности GitLabIntegration.
func (GitLabIntegration) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_path").
			Edges("project").
			Unique(),
	}
}
