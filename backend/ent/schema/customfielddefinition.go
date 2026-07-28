package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CustomFieldDefinition — определение Ent-схемы сущности CustomFieldDefinition.
type CustomFieldDefinition struct {
	ent.Schema
}

// Поля сущности CustomFieldDefinition.
func (CustomFieldDefinition) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(100),
		field.Enum("field_type").
			Values("text", "number", "date", "checkbox", "select", "multi_select", "user", "url"),
		field.JSON("options", []string{}).
			Optional().
			Comment("For select/multi_select"),
		field.Bool("required").
			Default(false),
		field.Int("sort_order").
			Default(0),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		// A2W8-2-004 / ADR-ENT-SOFT-DELETE-GAPS-2026 (version → TASK-095).
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}

// Рёбра (edges) сущности CustomFieldDefinition.
func (CustomFieldDefinition) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("project", Project.Type).
			Ref("custom_fields").
			Unique().
			Required(),
	}
}

// Индексы сущности CustomFieldDefinition.
func (CustomFieldDefinition) Indexes() []ent.Index {
	return []ent.Index{
		index.Edges("project").
			Fields("name").
			Unique(),
		index.Fields("deleted_at"),
		index.Edges("project").Fields("deleted_at"),
	}
}
