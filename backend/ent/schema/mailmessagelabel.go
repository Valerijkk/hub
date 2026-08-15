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

type MailMessageLabel struct {
	ent.Schema
}

func (MailMessageLabel) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("mail_message_id", uuid.UUID{}).Immutable(),
		field.UUID("mail_label_id", uuid.UUID{}).Immutable(),
		field.Bool("manual"),
		field.Bool("engine_rule"),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (MailMessageLabel) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_message_labels_manual_or_engine_rule": "COALESCE(manual OR engine_rule, FALSE)",
		}),
	}
}

func (MailMessageLabel) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("message", MailMessage.Type).Ref("labels").Field("mail_message_id").Unique().Required().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("label", MailLabel.Type).Ref("message_labels").Field("mail_label_id").Unique().Required().Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (MailMessageLabel) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mail_message_id", "mail_label_id").Unique().StorageKey("mail_message_labels_message_label_uidx"),
	}
}
