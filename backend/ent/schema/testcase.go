package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TestStep описывает один шаг тест-кейса (классический или BDD).
type TestStep struct {
	Order          int    `json:"order"`
	Action         string `json:"action"`
	ExpectedResult string `json:"expected_result"`
	Keyword        string `json:"keyword,omitempty"` // Given/When/Then для BDD
}

// TestCase — определение Ent-схемы сущности TestCase.
type TestCase struct {
	ent.Schema
}

// Поля сущности TestCase.
func (TestCase) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Int("case_number").
			Positive().
			Immutable().
			Comment("Порядковый номер в проекте"),
		field.String("case_key").
			Unique().
			Immutable().
			Comment("Ключ тест-кейса: HUB-TC-42"),
		field.String("title").
			NotEmpty().
			MaxLen(500),
		field.Text("description").
			Optional(),
		field.Text("precondition").
			Optional(),
		field.Text("postcondition").
			Optional(),
		field.Enum("case_type").
			Values("classic", "bdd").
			Default("classic"),
		field.Enum("priority").
			Values("critical", "high", "medium", "low", "trivial").
			Default("medium"),
		field.Enum("status").
			Values("draft", "active", "deprecated").
			Default("draft"),
		field.JSON("steps", []TestStep{}).
			Optional().
			Comment("Шаги тест-кейса в формате JSON"),
		field.String("full_name").
			Optional().
			Comment("Полное имя для Allure matching"),
		field.String("class_name").
			Optional().
			Comment("Имя класса для Allure matching"),
		field.String("method_name").
			Optional().
			Comment("Имя метода для Allure matching"),
		field.String("external_id").
			Optional().
			Comment("Внешний ID для @TestCase аннотации"),
		field.String("automation_id").
			Optional().
			Comment("Явный ID для привязки автотестов. QA-инженер задаёт вручную, автотест включает этот ID в имя или аннотацию для автоматического сопоставления результатов"),
		field.Int("version").
			Default(1).
			Comment("Версия тест-кейса"),
		field.Enum("sync_status").
			Values("synced", "outdated", "unlinked").
			Default("unlinked").
			Comment("Статус синхронизации с GitHub: synced/outdated/unlinked"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Время мягкого удаления (корзина 30 дней)"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

// Рёбра (edges) сущности TestCase.
func (TestCase) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("test_cases").
			Unique().
			Required(),
		edge.From("suite", TestSuite.Type).
			Ref("test_cases").
			Unique(),
		edge.From("created_by", User.Type).
			Ref("created_test_cases").
			Unique().
			Required(),
		edge.To("versions", TestCaseVersion.Type),
		edge.To("results", TestResult.Type),
		edge.To("labels", Label.Type),
		edge.To("linked_issues", Issue.Type),
		edge.To("components", Component.Type),
	}
}

// Индексы сущности TestCase.
func (TestCase) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("case_key"),
		index.Fields("external_id"),
		index.Fields("automation_id"),
		index.Fields("full_name"),
		index.Fields("class_name", "method_name"),
		index.Fields("status"),
		index.Fields("priority"),
		index.Fields("deleted_at"),
		// A8-002: ListTestCases WHERE project_test_cases = $1 AND deleted_at IS NULL
		index.Edges("project").Fields("deleted_at"),
	}
}
