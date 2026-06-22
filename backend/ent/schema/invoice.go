package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// Invoice — счёт на оплату (V4-03.5, Phase 52).
type Invoice struct {
	ent.Schema
}

func (Invoice) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("number").
			NotEmpty().
			MaxLen(32).
			Comment("INV-YYYY-NNN"),
		field.String("customer_name").
			NotEmpty().
			MaxLen(200),
		field.Text("description").
			Optional(),
		field.Int64("amount_cents").
			Positive(),
		field.String("currency").
			Default("RUB").
			MaxLen(3),
		field.Enum("status").
			Values("draft", "issued", "paid", "cancelled").
			Default("draft"),
		field.Time("due_date").
			Optional().
			Nillable(),
		field.Time("issued_at").
			Optional().
			Nillable(),
		field.Time("paid_at").
			Optional().
			Nillable(),
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

func (Invoice) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("created_by", User.Type).
			Ref("invoices_created").
			Unique().
			Required(),
		edge.From("crm_deal", CrmDeal.Type).
			Ref("invoices").
			Unique(),
		edge.From("journal_entry", JournalEntry.Type).
			Ref("invoice").
			Unique(),
	}
}

func (Invoice) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("number").Unique(),
		index.Fields("status"),
		index.Fields("deleted_at"),
	}
}
