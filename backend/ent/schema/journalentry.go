package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// JournalEntry — заголовок журнала с двойной записью (V4-03).
type JournalEntry struct {
	ent.Schema
}

func (JournalEntry) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Time("entry_date").
			Comment("UTC posting date"),
		field.String("description").
			NotEmpty().
			MaxLen(500),
		field.String("reference").
			Optional().
			MaxLen(100),
		field.Enum("status").
			Values("draft", "posted").
			Default("draft"),
		field.Time("posted_at").
			Optional().
			Nillable(),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
		field.Time("deleted_at").
			Optional().
			Nillable(),
	}
}
// Mixin подключает VersionMixin (NFR-24 / ADR-ENT-VERSION-MIXIN-2026).
func (JournalEntry) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


func (JournalEntry) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("created_by", User.Type).
			Ref("journal_entries").
			Unique().
			Required(),
		edge.To("lines", JournalLine.Type),
		// P1-BE-42 / W8-A3-003: O2O с Invoice (From уже Unique).
		edge.To("invoice", Invoice.Type).Unique(),
	}
}

func (JournalEntry) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("entry_date"),
		index.Fields("status"),
		index.Fields("deleted_at"),
	}
}
