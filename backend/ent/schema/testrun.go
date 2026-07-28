package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TestRun — определение Ent-схемы сущности TestRun.
type TestRun struct {
	ent.Schema
}

// Поля сущности TestRun.
func (TestRun) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("title").
			NotEmpty().
			MaxLen(500),
		field.Text("description").
			Optional(),
		field.Enum("run_type").
			Values("manual", "automatic", "mixed").
			Default("manual"),
		field.Enum("status").
			Values("created", "in_progress", "completed", "aborted").
			Default("created"),
		field.Int("total_count").
			Default(0),
		field.Int("passed_count").
			Default(0),
		field.Int("failed_count").
			Default(0),
		field.Int("skipped_count").
			Default(0),
		field.Int("broken_count").
			Default(0),
		field.Int("unknown_count").
			Default(0),
		field.String("allure_run_id").
			Optional().
			Comment("Идентификатор прогона Allure"),
		field.Time("started_at").
			Optional().
			Nillable(),
		field.Time("completed_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Рёбра (edges) сущности TestRun.
func (TestRun) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("test_runs").
			Unique().
			Required(),
		edge.From("created_by", User.Type).
			Ref("created_runs").
			Unique().
			Required(),
		edge.To("results", TestResult.Type),
		edge.To("assignees", User.Type),
		edge.From("environment", Environment.Type).
			Ref("test_runs").
			Unique(),
	}
}

// Индексы сущности TestRun.
func (TestRun) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("run_type"),
		index.Fields("allure_run_id"),
		index.Fields("created_at"),
	}
}
