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

type MailMessageFolder struct {
	ent.Schema
}

func (MailMessageFolder) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("mail_message_id", uuid.UUID{}).Immutable(),
		field.UUID("mail_folder_id", uuid.UUID{}).Immutable(),
		field.String("engine_state").Optional().Nillable().Validate(validateMailMessageEngineState),
		field.Int64("last_seen_sync_epoch").Default(0).Validate(validateMailProjectionEpoch),
		field.Time("projected_at"),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (MailMessageFolder) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_message_folders_engine_state": "COALESCE(engine_state IS NULL OR octet_length(engine_state) <= 4096, FALSE)",
			"mail_message_folders_epoch":        "COALESCE(last_seen_sync_epoch BETWEEN 0 AND 2147483647, FALSE)",
		}),
	}
}

func (MailMessageFolder) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("message", MailMessage.Type).Ref("folders").Field("mail_message_id").Unique().Required().Immutable().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("folder", MailFolder.Type).Ref("message_folders").Field("mail_folder_id").Unique().Required().Immutable().Annotations(entsql.OnDelete(entsql.Cascade)),
	}
}

func (MailMessageFolder) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mail_message_id", "mail_folder_id").Unique().StorageKey("mail_message_folders_message_folder_uidx"),
	}
}
