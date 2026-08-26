package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type MailMutationReceiptMember struct {
	ent.Schema
}

func (MailMutationReceiptMember) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("receipt_id", uuid.UUID{}).Immutable(),
		field.UUID("target_id", uuid.UUID{}).Immutable(),
		field.UUID("message_id", uuid.UUID{}).Immutable(),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (MailMutationReceiptMember) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("receipt", MailMutationReceipt.Type).Ref("members").Field("receipt_id").Unique().Required().Immutable().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("target", MailMutationReceiptTarget.Type).Ref("members").Field("target_id").Unique().Required().Immutable().Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (MailMutationReceiptMember) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("receipt_id", "message_id").Unique(),
		index.Fields("target_id", "message_id").Unique(),
	}
}
