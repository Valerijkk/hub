package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// GitHubLink хранит связь PR или коммита из GitHub с задачей трекера.
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
			Comment("Тип связи: коммит или pull request"),
		field.String("external_id").
			NotEmpty().
			Comment("GitHub PR number или commit SHA"),
		field.String("title").
			MaxLen(500).
			Comment("Заголовок PR или сообщение коммита"),
		field.String("url").
			Comment("Полный URL на GitHub"),
		field.Enum("status").
			Values("open", "merged", "closed", "none").
			Default("none").
			Comment("Статус PR (для коммитов — none)"),
		field.String("author").
			Optional().
			Comment("Имя автора на GitHub"),
		field.String("branch").
			Optional().
			Comment("Имя ветки (для PR)"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
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
