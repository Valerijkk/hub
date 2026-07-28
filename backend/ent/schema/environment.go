package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Environment — определение Ent-схемы сущности Environment.
type Environment struct {
	ent.Schema
}

// Поля сущности Environment.
func (Environment) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(100).
			Comment("Название окружения"),
		field.Enum("env_label").
			Values("dev", "staging", "prod", "custom").
			Default("custom").
			Comment("Метка окружения"),
		field.JSON("config", map[string]interface{}{}).
			Optional().
			Comment("Конфигурация: {browser, os, version}"),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

// ибра (edges) сущности Environment.
func (Environment) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("environments").
			Unique().
			Required(),
		edge.To("test_runs", TestRun.Type),
	}
}

// Индексы сущности Environment.
func (Environment) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("env_label"),
	}
}
