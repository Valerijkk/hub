package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Environment holds the schema definition for the Environment entity.
type Environment struct {
	ent.Schema
}

// Fields of the Environment.
func (Environment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(100).
			Comment("РќР°Р·РІР°РЅРёРµ РѕРєСЂСѓР¶РµРЅРёСЏ"),
		field.Enum("env_label").
			Values("dev", "staging", "prod", "custom").
			Default("custom").
			Comment("РњРµС‚РєР° РѕРєСЂСѓР¶РµРЅРёСЏ"),
		field.JSON("config", map[string]interface{}{}).
			Optional().
			Comment("РљРѕРЅС„РёРіСѓСЂР°С†РёСЏ: {browser, os, version}"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// Edges of the Environment.
func (Environment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("environments").
			Unique().
			Required(),
		edge.To("test_runs", TestRun.Type),
	}
}

// Indexes of the Environment.
func (Environment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("env_label"),
	}
}
