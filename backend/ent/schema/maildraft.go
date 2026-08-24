package schema

import (
	"crypto/sha256"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	"github.com/google/uuid"
)

type MailDraft struct {
	ent.Schema
}

func (MailDraft) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).Immutable(),
		field.UUID("from_identity_id", uuid.UUID{}),
		field.UUID("draft_client_id", uuid.UUID{}).Immutable(),
		field.String("engine_draft_key").Validate(validateMailMessageOpaqueID).Immutable(),
		field.String("engine_message_id").Optional().Nillable().Validate(validateMailMessageOpaqueID),
		field.String("engine_version").Optional().Nillable().Validate(validateMailMessageOpaqueID),
		field.Int64("revision").Default(1).Min(1),
		field.Enum("status").Values("active", "saving", "deleting", "unknown", "submitting", "submitted", "deleted"),
		field.Bytes("create_idempotency_key_hash").MinLen(sha256.Size).MaxLen(sha256.Size).Immutable(),
		field.Bytes("create_request_hash").MinLen(sha256.Size).MaxLen(sha256.Size).Immutable(),
		field.UUID("reply_to_message_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("forward_message_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("pending_from_identity_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("pending_reply_to_message_id", uuid.UUID{}).Optional().Nillable(),
		field.UUID("pending_forward_message_id", uuid.UUID{}).Optional().Nillable(),
		field.Bytes("pending_save_request_hash").Optional().Nillable().MinLen(sha256.Size).MaxLen(sha256.Size),
		field.UUID("last_mutation_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("pending_mutation_kind").Values("create", "save", "delete").Optional().Nillable(),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (MailDraft) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_drafts_key":                "COALESCE(octet_length(engine_draft_key) BETWEEN 1 AND 1024, FALSE)",
			"mail_drafts_engine_pair":        "COALESCE((engine_message_id IS NULL AND engine_version IS NULL) OR (octet_length(engine_message_id) BETWEEN 1 AND 1024 AND octet_length(engine_version) BETWEEN 1 AND 1024), FALSE)",
			"mail_drafts_revision":           "COALESCE(revision >= 1, FALSE)",
			"mail_drafts_hashes":             "COALESCE(octet_length(create_idempotency_key_hash) = 32 AND octet_length(create_request_hash) = 32, FALSE)",
			"mail_drafts_status_enum":        "COALESCE(status IN ('active','saving','deleting','unknown','submitting','submitted','deleted'), FALSE)",
			"mail_drafts_state":              "COALESCE((status IN ('active','submitting','submitted') AND engine_message_id IS NOT NULL AND last_mutation_id IS NULL AND pending_mutation_kind IS NULL) OR (status = 'saving' AND last_mutation_id IS NOT NULL AND pending_mutation_kind IN ('create','save') AND ((pending_mutation_kind = 'create' AND engine_message_id IS NULL) OR (pending_mutation_kind = 'save' AND engine_message_id IS NOT NULL))) OR (status = 'deleting' AND last_mutation_id IS NOT NULL AND pending_mutation_kind = 'delete' AND engine_message_id IS NOT NULL) OR (status = 'unknown' AND last_mutation_id IS NOT NULL AND ((pending_mutation_kind = 'create' AND engine_message_id IS NULL) OR (pending_mutation_kind IN ('save','delete') AND engine_message_id IS NOT NULL))) OR (status = 'deleted' AND last_mutation_id IS NULL AND pending_mutation_kind IS NULL), FALSE)",
			"mail_drafts_pending_save_state": "COALESCE((pending_mutation_kind = 'save' AND pending_from_identity_id IS NOT NULL AND octet_length(pending_save_request_hash) = 32) OR ((pending_mutation_kind IS NULL OR pending_mutation_kind <> 'save') AND pending_from_identity_id IS NULL AND pending_reply_to_message_id IS NULL AND pending_forward_message_id IS NULL AND pending_save_request_hash IS NULL), FALSE)",
		}),
	}
}

func (MailDraft) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", MailAccount.Type).Ref("drafts").Field("mail_account_id").Unique().Required().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("from_identity", MailIdentity.Type).Ref("drafts").Field("from_identity_id").Unique().Required().Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.From("reply_to_message", MailMessage.Type).Ref("reply_drafts").Field("reply_to_message_id").Unique().Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.From("forward_message", MailMessage.Type).Ref("forward_drafts").Field("forward_message_id").Unique().Annotations(entsql.OnDelete(entsql.SetNull)),
		edge.From("pending_from_identity", MailIdentity.Type).Ref("pending_drafts").Field("pending_from_identity_id").Unique().Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.From("pending_reply_to_message", MailMessage.Type).Ref("pending_reply_drafts").Field("pending_reply_to_message_id").Unique().Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.From("pending_forward_message", MailMessage.Type).Ref("pending_forward_drafts").Field("pending_forward_message_id").Unique().Annotations(entsql.OnDelete(entsql.Restrict)),
		edge.To("attachments", MailAttachment.Type),
		edge.To("pending_attachments", MailAttachment.Type),
		edge.To("submissions", MailSubmission.Type),
	}
}

func (MailDraft) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mail_account_id", "engine_draft_key").Unique().StorageKey("mail_drafts_account_engine_key_uidx"),
		index.Fields("mail_account_id", "draft_client_id").Unique().StorageKey("mail_drafts_account_client_uidx"),
		index.Fields("mail_account_id", "create_idempotency_key_hash").Unique().StorageKey("mail_drafts_account_create_key_uidx"),
		index.Fields("mail_account_id", "engine_message_id").Unique().StorageKey("mail_drafts_account_engine_message_uidx").Annotations(entsql.IndexWhere("engine_message_id IS NOT NULL")),
		index.Fields("from_identity_id").StorageKey("mail_drafts_from_identity_idx"),
	}
}
