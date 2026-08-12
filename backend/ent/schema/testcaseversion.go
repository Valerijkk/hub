package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TestCaseVersion — определение Ent-схемы сущности TestCaseVersion.
type TestCaseVersion struct {
	ent.Schema
}

// Поля сущности TestCaseVersion.
func (TestCaseVersion) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Int("version_number").
			Positive().
			Comment("Номер версии снимка"),
		field.JSON("snapshot", map[string]interface{}{}).
			Comment("Полный снимок тест-кейса"),
		field.Text("change_description").
			Optional().
			Comment("Описание изменений"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности TestCaseVersion.
func (TestCaseVersion) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("test_case", TestCase.Type).
			Ref("versions").
			Unique().
			Required(),
		edge.From("created_by", User.Type).
			Ref("test_case_versions").
			Unique(),
	}
}

// Индексы сущности TestCaseVersion.
func (TestCaseVersion) Indexes() []ent.Index {
	return []ent.Index{
		// P2-BE-22: unique (test_case, version_number).
		index.Edges("test_case").Fields("version_number").Unique(),
	}
}
