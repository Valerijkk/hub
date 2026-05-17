package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// TestCaseVersion holds the schema definition for the TestCaseVersion entity.
type TestCaseVersion struct {
	ent.Schema
}

// Fields of the TestCaseVersion.
func (TestCaseVersion) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Int("version_number").
			Positive().
			Comment("РќРѕРјРµСЂ РІРµСЂСЃРёРё СЃРЅРёРјРєР°"),
		field.JSON("snapshot", map[string]interface{}{}).
			Comment("РџРѕР»РЅС‹Р№ СЃРЅРёРјРѕРє С‚РµСЃС‚-РєРµР№СЃР°"),
		field.Text("change_description").
			Optional().
			Comment("РћРїРёСЃР°РЅРёРµ РёР·РјРµРЅРµРЅРёР№"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges of the TestCaseVersion.
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

// Indexes of the TestCaseVersion.
func (TestCaseVersion) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("version_number"),
	}
}
