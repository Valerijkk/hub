package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

// CrmContact — контакт CRM-lite (MOD-05, отдельно от User).
type CrmContact struct {
	ent.Schema
}

func (CrmContact) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.String("name").
			NotEmpty().
			MaxLen(200),
		field.String("email").
			Optional().
			MaxLen(320),
		field.String("phone").
			Optional().
			MaxLen(50),
		field.String("company").
			Optional().
			MaxLen(200),
		field.Text("notes").
			Optional(),
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

func (CrmContact) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("created_by", User.Type).
			Ref("crm_contacts_created").
			Unique().
			Required(),
		edge.From("linked_user", User.Type).
			Ref("crm_contact_link").
			Unique(),
		edge.To("deals", CrmDeal.Type),
	}
}

func (CrmContact) Indexes() []ent.Index {
	return []ent.Index{
		// A2W8-2-005: lookup по email — hot-path CRM.
		index.Fields("email"),
		index.Fields("deleted_at"),
		index.Fields("company"),
	}
}
