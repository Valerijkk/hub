package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// LedgerAccount — план счетов (V4-03).
type LedgerAccount struct {
	ent.Schema
}

func (LedgerAccount) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("code").
			NotEmpty().
			MaxLen(20),
		field.String("name").
			NotEmpty().
			MaxLen(200),
		field.Enum("account_type").
			Values("asset", "liability", "equity", "revenue", "expense"),
		field.Bool("is_active").
			Default(true),
		field.Int("version").
			Default(1),
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

func (LedgerAccount) Edges() []ent.Edge {
	return []ent.Edge{
		edge.To("journal_lines", JournalLine.Type),
	}
}

func (LedgerAccount) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("code").Unique(),
		index.Fields("deleted_at"),
	}
}
