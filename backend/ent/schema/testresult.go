package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
	"time"
)

// StepResult описывает результат выполнения одного шага тест-кейса.
type StepResult struct {
	StepOrder int      `json:"step_order"`
	Status    string   `json:"status"`
	Comment   string   `json:"comment,omitempty"`
	FileIDs   []string `json:"file_ids,omitempty"`
}

// TestResult holds the schema definition for the TestResult entity.
type TestResult struct {
	ent.Schema
}

// Fields of the TestResult.
func (TestResult) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Enum("status").
			Values("passed", "failed", "broken", "skipped", "unknown").
			Default("unknown"),
		field.Text("comment").
			Optional(),
		field.JSON("step_results", []StepResult{}).
			Optional().
			Comment("Результаты выполнения шагов"),
		field.Int64("duration_ms").
			Optional().
			Nillable().
			Comment("Длительность выполнения в миллисекундах"),
		field.String("allure_uuid").
			Optional().
			Comment("UUID результата из Allure"),
		field.Text("error_message").
			Optional(),
		field.Text("error_trace").
			Optional(),
		field.Bool("is_flaky").
			Default(false).
			Comment("Нестабильный тест"),
		field.Enum("match_strategy").
			Values("annotation", "fullname", "classmethod", "unmapped").
			Default("unmapped").
			Comment("Стратегия привязки к тест-кейсу"),
		field.Time("started_at").
			Optional().
			Nillable(),
		field.Time("completed_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(time.Now).
			Immutable(),
	}
}

// Edges of the TestResult.
func (TestResult) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("run", TestRun.Type).
			Ref("results").
			Unique().
			Required(),
		edge.From("test_case", TestCase.Type).
			Ref("results").
			Unique(),
		edge.From("executed_by", User.Type).
			Ref("executed_results").
			Unique(),
		edge.To("files", FileMetadata.Type),
	}
}

// Indexes of the TestResult.
func (TestResult) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("allure_uuid"),
		index.Fields("is_flaky"),
	}
}
