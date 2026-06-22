package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TestStep РѕРїРёСЃС‹РІР°РµС‚ РѕРґРёРЅ С€Р°Рі С‚РµСЃС‚-РєРµР№СЃР° (РєР»Р°СЃСЃРёС‡РµСЃРєРёР№ РёР»Рё BDD).
type TestStep struct {
	Order          int    `json:"order"`
	Action         string `json:"action"`
	ExpectedResult string `json:"expected_result"`
	Keyword        string `json:"keyword,omitempty"` // Given/When/Then РґР»СЏ BDD
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
			Comment("РџРѕСЂСЏРґРєРѕРІС‹Р№ РЅРѕРјРµСЂ РІ РїСЂРѕРµРєС‚Рµ"),
		field.String("case_key").
			Unique().
			Immutable().
			Comment("РљР»СЋС‡ С‚РµСЃС‚-РєРµР№СЃР°: HUB-TC-42"),
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
			Comment("РЁР°РіРё С‚РµСЃС‚-РєРµР№СЃР° РІ С„РѕСЂРјР°С‚Рµ JSON"),
		field.String("full_name").
			Optional().
			Comment("РџРѕР»РЅРѕРµ РёРјСЏ РґР»СЏ Allure matching"),
		field.String("class_name").
			Optional().
			Comment("РРјСЏ РєР»Р°СЃСЃР° РґР»СЏ Allure matching"),
		field.String("method_name").
			Optional().
			Comment("РРјСЏ РјРµС‚РѕРґР° РґР»СЏ Allure matching"),
		field.String("external_id").
			Optional().
			Comment("Р’РЅРµС€РЅРёР№ ID РґР»СЏ @TestCase Р°РЅРЅРѕС‚Р°С†РёРё"),
		field.String("automation_id").
			Optional().
			Comment("РЇРІРЅС‹Р№ ID РґР»СЏ РїСЂРёРІСЏР·РєРё Р°РІС‚РѕС‚РµСЃС‚РѕРІ. QA-РёРЅР¶РµРЅРµСЂ Р·Р°РґР°С‘С‚ РІСЂСѓС‡РЅСѓСЋ, Р°РІС‚РѕС‚РµСЃС‚ РІРєР»СЋС‡Р°РµС‚ СЌС‚РѕС‚ ID РІ РёРјСЏ РёР»Рё Р°РЅРЅРѕС‚Р°С†РёСЋ РґР»СЏ Р°РІС‚РѕРјР°С‚РёС‡РµСЃРєРѕРіРѕ СЃРѕРїРѕСЃС‚Р°РІР»РµРЅРёСЏ СЂРµР·СѓР»СЊС‚Р°С‚РѕРІ"),
		field.Int("version").
			Default(1).
			Comment("Р’РµСЂСЃРёСЏ С‚РµСЃС‚-РєРµР№СЃР°"),
		field.Enum("sync_status").
			Values("synced", "outdated", "unlinked").
			Default("unlinked").
			Comment("РЎС‚Р°С‚СѓСЃ СЃРёРЅС…СЂРѕРЅРёР·Р°С†РёРё СЃ GitHub: synced/outdated/unlinked"),
		field.Time("deleted_at").
			Optional().
			Nillable().
			Comment("Р’СЂРµРјСЏ РјСЏРіРєРѕРіРѕ СѓРґР°Р»РµРЅРёСЏ (РєРѕСЂР·РёРЅР° 30 РґРЅРµР№)"),
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
	}
}
