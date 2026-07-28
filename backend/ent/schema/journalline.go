package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// JournalLine — строка дебета/кредита в проводке журнала.
type JournalLine struct {
	ent.Schema
}

func (JournalLine) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.Int64("debit_cents").
			Default(0).
			NonNegative(),
		field.Int64("credit_cents").
			Default(0).
			NonNegative(),
		field.String("memo").
			Optional().
			MaxLen(300),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
	}
}

func (JournalLine) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("entry", JournalEntry.Type).
			Ref("lines").
			Unique().
			Required(),
		edge.From("account", LedgerAccount.Type).
			Ref("journal_lines").
			Unique().
			Required(),
	}
}

func (JournalLine) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("debit_cents", "credit_cents"),
	}
}

// Annotations — P2-BE-22: XOR debit/credit CHECK.
// Выражение без OR — совместимо с SQLite (enttest) и Postgres.
func (JournalLine) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"journalline_debit_xor_credit": "(debit_cents > 0) = (credit_cents = 0) AND (debit_cents + credit_cents) > 0",
		}),
	}
}
