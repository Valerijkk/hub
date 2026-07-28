package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/mixin"
)

// VersionMixin добавляет optimistic-lock поле version (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
type VersionMixin struct {
	mixin.Schema
}

// Fields — единый source of truth для version на mutable_needs_version схемах.
func (VersionMixin) Fields() []ent.Field {
	return []ent.Field{
		field.Int("version").
			Default(1).
			Positive().
			Comment("Optimistic locking (NFR-24 / ADR-ENT-VERSION-MIXIN-2026)"),
	}
}
