package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// DashboardLayout хранит раскладку виджетов дашборда пользователя.
type DashboardLayout struct {
	ent.Schema
}

// Fields описывает поля сущности DashboardLayout.
func (DashboardLayout) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("user_id", uuid.UUID{}).
			Comment("ID владельца раскладки"),
		field.Enum("dashboard_type").
			Values("personal", "project", "admin").
			Comment("Тип дашборда: personal, project, admin"),
		field.UUID("project_id", uuid.UUID{}).
			Optional().
			Nillable().
			Comment("ID проекта (только для project-дашборда)"),
		field.JSON("items", []map[string]interface{}{}).
			Comment("JSON-массив виджетов с позицией и размерами"),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
		field.Time("updated_at").
			Default(time.Now).
			UpdateDefault(time.Now),
	}
}

// Edges описывает связи сущности DashboardLayout.
func (DashboardLayout) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("user", User.Type).
			Ref("dashboard_layouts").
			Field("user_id").
			Required().
			Unique(),
	}
}

// Indexes описывает индексы сущности DashboardLayout.
func (DashboardLayout) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("user_id", "dashboard_type", "project_id").
			Unique(),
	}
}
