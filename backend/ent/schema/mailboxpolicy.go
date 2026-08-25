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

type MailboxPolicy struct {
	ent.Schema
}

func (MailboxPolicy) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).
			Default(uuid.New).
			Immutable(),
		field.UUID("organization_id", uuid.UUID{}).
			Immutable(),
		field.Bool("external_mail_enabled"),
		field.Int64("max_message_bytes").
			Min(1_048_576).
			Max(52_428_800),
		field.Int64("max_attachment_bytes").
			Min(1).
			Max(26_214_400),
		field.Int32("max_recipients").
			Min(1).
			Max(300),
		field.Int32("send_rate_per_minute").
			Min(1).
			Max(1_000),
		field.Int32("instance_send_rate_per_minute").
			Default(1_000).
			Min(1).
			Max(1_000),
		field.Int32("trash_retention_days").
			Min(1).
			Max(3_650),
		field.Int32("junk_retention_days").
			Min(1).
			Max(3_650),
		field.Time("created_at").
			Default(nowUTC).
			Immutable(),
		field.Time("updated_at").
			Default(nowUTC).
			UpdateDefault(nowUTC),
	}
}

func (MailboxPolicy) Mixin() []ent.Mixin {
	return []ent.Mixin{
		VersionMixin{},
	}
}

func (MailboxPolicy) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mailbox_policies_bounds": "max_message_bytes BETWEEN 1048576 AND 52428800 AND max_attachment_bytes BETWEEN 1 AND 26214400 AND max_attachment_bytes <= max_message_bytes AND max_recipients BETWEEN 1 AND 300 AND send_rate_per_minute BETWEEN 1 AND 1000 AND instance_send_rate_per_minute BETWEEN 1 AND 1000 AND trash_retention_days BETWEEN 1 AND 3650 AND junk_retention_days BETWEEN 1 AND 3650",
		}),
	}
}

func (MailboxPolicy) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).
			Ref("mailbox_policy").
			Field("organization_id").
			Unique().
			Required().
			Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (MailboxPolicy) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id").
			Unique(),
	}
}
