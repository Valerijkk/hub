package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// ProjectTemplate holds the schema definition for the ProjectTemplate entity.
type ProjectTemplate struct {
	ent.Schema
}

// Fields of the ProjectTemplate.
func (ProjectTemplate) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(255).
			Comment("Название шаблона"),
		field.Text("description").
			Optional().
			Comment("Описание шаблона"),
		field.Enum("template_type").
			Values("empty", "software_dev", "qa", "kanban_simple", "custom").
			Default("custom").
			Comment("Тип шаблона"),
		field.JSON("config", map[string]interface{}{}).
			Comment("Снимок конфигурации: statuses, labels, boards, custom_fields"),
		field.Bool("is_system").
			Default(false).
			Comment("Системный шаблон (нельзя удалить)"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the ProjectTemplate.
func (ProjectTemplate) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("created_by", User.Type).
			Ref("project_templates").
			Unique(),
	}
}

// Indexes of the ProjectTemplate.
func (ProjectTemplate) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("template_type"),
		index.Fields("is_system"),
	}
}
