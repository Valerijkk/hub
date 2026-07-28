package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CrmDeal — сделка в pipeline CRM-lite (MOD-05).
type CrmDeal struct {
	ent.Schema
}

func (CrmDeal) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("title").
			NotEmpty().
			MaxLen(300),
		field.Text("description").
			Optional(),
		field.Enum("stage").
			Values("lead", "qualified", "proposal", "negotiation", "won", "lost").
			Default("lead"),
		field.Int64("amount_cents").
			Optional().
			Nillable().
			Comment("Сумма сделки в минорных единицах (копейки/центы)"),
		field.String("currency").
			Default("RUB").
			MaxLen(3),
		field.String("rank").
			Optional().
			Comment("LexoRank-подобный порядок внутри колонки"),
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
func (CrmDeal) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}


func (CrmDeal) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("owner", User.Type).
			Ref("crm_deals_owned").
			Unique().
			Required(),
		edge.From("contact", CrmContact.Type).
			Ref("deals").
			Unique(),
		edge.From("project", Project.Type).
			Ref("crm_deals").
			Unique(),
		edge.To("invoices", Invoice.Type),
	}
}

func (CrmDeal) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("stage", "rank"),
		index.Fields("deleted_at"),
		// A2W8-2-005: ListDeals WHERE owner AND deleted_at IS NULL
		index.Edges("owner").Fields("deleted_at"),
	}
}
