package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// SavedFilter хранит сохранённые фильтры задач трекера.
// Пользователь может создавать, сохранять и расшаривать фильтры в рамках проекта.
type SavedFilter struct {
	ent.Schema
}

// Fields описывает поля сохранённого фильтра.
func (SavedFilter) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(200).
			Comment("Название фильтра"),
		field.JSON("filter_config", map[string]interface{}{}).
			Comment("JSON-конфигурация фильтра (статус, приоритет, исполнитель и т.д.)"),
		field.Bool("shared").
			Default(false).
			Comment("Расшарен ли фильтр для всех участников проекта"),
		field.UUID("owner_id", uuid.UUID{}).
			Comment("ID владельца фильтра"),
		field.UUID("project_id", uuid.UUID{}).
			Comment("ID проекта, к которому привязан фильтр"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges описывает связи сохранённого фильтра.
func (SavedFilter) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("saved_filters").
			Field("owner_id").
			Required().
			Unique(),
		edge.From("project", Project.Type).
			Ref("saved_filters").
			Field("project_id").
			Required().
			Unique(),
	}
}

// Indexes описывает индексы для сохранённых фильтров.
func (SavedFilter) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("owner_id", "project_id"),
		index.Fields("project_id", "shared"),
	}
}
