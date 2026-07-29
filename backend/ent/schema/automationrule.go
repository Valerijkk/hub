package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// AutomationRule — пользовательское правило автоматизации (PROD-01, v2.6).
type AutomationRule struct {
	ent.Schema
}

// Поля сущности AutomationRule.
func (AutomationRule) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("project_id", uuid.UUID{}),
		field.UUID("created_by_id", uuid.UUID{}),
		field.String("name").
			NotEmpty().
			MaxLen(100),
		field.Bool("enabled").
			Default(true),
		field.Enum("trigger_type").
			Values("status_changed", "issue_created", "issue_assigned"),
		field.JSON("trigger_config", map[string]interface{}{}).
			Default(map[string]interface{}{}),
		field.Enum("action_type").
			Values("post_to_channel", "notify_user"),
		field.JSON("action_config", map[string]interface{}{}).
			Default(map[string]interface{}{}),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		// EA-001 / TASK-191 / ADR-ENT-SOFT-DELETE-GAPS: корзина 30д вместо hard Delete.
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (AutomationRule) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


// Рёбра (edges) сущности AutomationRule.
func (AutomationRule) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("automation_rules").
			Field("project_id").
			Required().
			Unique(),
		edge.From("created_by", User.Type).
			Ref("automation_rules").
			Field("created_by_id").
			Required().
			Unique(),
	}
}

// Индексы сущности AutomationRule.
func (AutomationRule) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("project_id", "enabled"),
		index.Fields("project_id", "trigger_type"),
		index.Fields("deleted_at"),
		index.Fields("project_id", "deleted_at"),
	}
}
