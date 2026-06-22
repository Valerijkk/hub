package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// StepResult РѕРїРёСЃС‹РІР°РµС‚ СЂРµР·СѓР»СЊС‚Р°С‚ РІС‹РїРѕР»РЅРµРЅРёСЏ РѕРґРЅРѕРіРѕ С€Р°РіР° С‚РµСЃС‚-РєРµР№СЃР°.
type StepResult struct {
	StepOrder int      `json:"step_order"`
	Status    string   `json:"status"`
	Comment   string   `json:"comment,omitempty"`
	FileIDs   []string `json:"file_ids,omitempty"`
}

// TestResult — определение Ent-схемы сущности TestResult.
type TestResult struct {
	ent.Schema
}

// Поля сущности TestResult.
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
			Comment("Р РµР·СѓР»СЊС‚Р°С‚С‹ РІС‹РїРѕР»РЅРµРЅРёСЏ С€Р°РіРѕРІ"),
		field.Int64("duration_ms").
			Optional().
			Nillable().
			Comment("Р”Р»РёС‚РµР»СЊРЅРѕСЃС‚СЊ РІС‹РїРѕР»РЅРµРЅРёСЏ РІ РјРёР»Р»РёСЃРµРєСѓРЅРґР°С…"),
		field.String("allure_uuid").
			Optional().
			Comment("UUID СЂРµР·СѓР»СЊС‚Р°С‚Р° РёР· Allure"),
		field.Text("error_message").
			Optional(),
		field.Text("error_trace").
			Optional(),
		field.Bool("is_flaky").
			Default(false).
			Comment("РќРµСЃС‚Р°Р±РёР»СЊРЅС‹Р№ С‚РµСЃС‚"),
		field.Enum("match_strategy").
			Values("annotation", "fullname", "classmethod", "unmapped").
			Default("unmapped").
			Comment("РЎС‚СЂР°С‚РµРіРёСЏ РїСЂРёРІСЏР·РєРё Рє С‚РµСЃС‚-РєРµР№СЃСѓ"),
		field.Time("started_at").
			Optional().
			Nillable(),
		field.Time("completed_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Рёбра (edges) сущности TestResult.
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

// Индексы сущности TestResult.
func (TestResult) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("status"),
		index.Fields("allure_uuid"),
		index.Fields("is_flaky"),
	}
}
