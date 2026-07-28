package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// GitLabLink хранит связь MR или коммита из GitLab с задачей трекера.
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
			Comment("Тип связи: коммит или merge request"),
		field.String("external_id").
			NotEmpty().
			Comment("GitLab MR IID или commit SHA"),
		field.String("title").
			MaxLen(500).
			Comment("Заголовок MR или сообщение коммита"),
		field.String("url").
			Comment("Полный URL на GitLab"),
		field.Enum("status").
			Values("opened", "merged", "closed", "none").
			Default("none").
			Comment("Статус MR (для коммитов — none)"),
		field.String("author").
			Optional().
			Comment("Рмя автора на GitLab"),
		field.String("source_branch").
			Optional().
			Comment("Рмя исходной ветки (для MR)"),
		field.String("project_path").
			Optional().
			Comment("Путь проекта в GitLab (group/repo)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// ибра (edges) сущности GitLabLink.
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
