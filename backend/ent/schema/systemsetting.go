package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// SystemSetting — определение Ent-схемы сущности SystemSetting.
type SystemSetting struct {
	ent.Schema
}

// Поля сущности SystemSetting.
func (SystemSetting) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("key").
			Unique().
			NotEmpty(),
		field.String("value"),
		field.String("description").
			Optional(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (SystemSetting) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

