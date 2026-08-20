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

type MailSubmission struct {
	ent.Schema
}

func (MailSubmission) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).Immutable(),
		field.UUID("mail_draft_id", uuid.UUID{}).Optional().Nillable(),
		field.Bytes("idempotency_key_hash").MinLen(sha256.Size).MaxLen(sha256.Size).Immutable(),
		field.Bytes("request_hash").MinLen(sha256.Size).MaxLen(sha256.Size).Immutable(),
		field.String("engine_submission_key").Validate(validateMailMessageOpaqueID).Immutable(),
		field.Enum("status").Values("queued", "submitting", "sent", "unknown", "failed"),
		field.String("engine_submission_id").Optional().Nillable().Validate(validateMailMessageOpaqueID),
		field.String("engine_result_message_id").Optional().Nillable().Validate(validateMailMessageOpaqueID),
		field.UUID("result_mail_message_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("last_error_code").Values("MAIL_SUBMISSION_OUTCOME_UNKNOWN", "MAIL_SUBMISSION_RECONCILIATION_REQUIRED", "MAILBOX_UNAVAILABLE", "MAIL_SUBMISSION_REJECTED", "MAIL_SUBMISSION_POLICY", "MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN", "MAIL_SUBMISSION_IDENTITY_UNAVAILABLE", "MAIL_SUBMISSION_RECIPIENT_LIMIT", "MAIL_SUBMISSION_SIZE_LIMIT").Optional().Nillable(),
		field.Int32("attempt_count").Default(0).Min(0),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (MailSubmission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_submissions_key":         "COALESCE(octet_length(engine_submission_key) BETWEEN 1 AND 1024, FALSE)",
			"mail_submissions_hashes":      "COALESCE(octet_length(idempotency_key_hash) = 32 AND octet_length(request_hash) = 32, FALSE)",
			"mail_submissions_opaque_pair": "COALESCE((engine_submission_id IS NULL AND engine_result_message_id IS NULL) OR (octet_length(engine_submission_id) BETWEEN 1 AND 1024 AND octet_length(engine_result_message_id) BETWEEN 1 AND 1024), FALSE)",
			"mail_submissions_attempts":    "COALESCE(attempt_count >= 0, FALSE)",
			"mail_submissions_status_enum": "COALESCE(status IN ('queued','submitting','sent','unknown','failed'), FALSE)",
			"mail_submissions_state":       "COALESCE((status IN ('queued','submitting') AND mail_draft_id IS NOT NULL AND engine_submission_id IS NULL AND engine_result_message_id IS NULL AND result_mail_message_id IS NULL AND last_error_code IS NULL) OR (status = 'sent' AND engine_submission_id IS NOT NULL AND engine_result_message_id IS NOT NULL AND result_mail_message_id IS NOT NULL AND last_error_code IS NULL) OR (status = 'unknown' AND mail_draft_id IS NOT NULL AND result_mail_message_id IS NULL AND last_error_code IN ('MAIL_SUBMISSION_OUTCOME_UNKNOWN','MAIL_SUBMISSION_RECONCILIATION_REQUIRED','MAILBOX_UNAVAILABLE') AND NOT (engine_submission_id IS NOT NULL AND last_error_code = 'MAIL_SUBMISSION_OUTCOME_UNKNOWN')) OR (status = 'failed' AND engine_submission_id IS NULL AND engine_result_message_id IS NULL AND result_mail_message_id IS NULL AND last_error_code IN ('MAIL_SUBMISSION_REJECTED','MAIL_SUBMISSION_POLICY','MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN','MAIL_SUBMISSION_IDENTITY_UNAVAILABLE','MAIL_SUBMISSION_RECIPIENT_LIMIT','MAIL_SUBMISSION_SIZE_LIMIT','MAILBOX_UNAVAILABLE')), FALSE)",
		}),
	}
}

func (MailSubmission) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("account", MailAccount.Type).Ref("submissions").Field("mail_account_id").Unique().Required().Annotations(entsql.OnDelete(entsql.Cascade)),
		edge.From("draft", MailDraft.Type).Ref("submissions").Field("mail_draft_id").Unique().Annotations(entsql.OnDelete(entsql.SetNull)),
	}
}

func (MailSubmission) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("mail_account_id", "idempotency_key_hash").Unique().StorageKey("mail_submissions_account_key_hash_uidx"),
		index.Fields("mail_account_id", "engine_submission_key").Unique().StorageKey("mail_submissions_account_engine_key_uidx"),
		index.Fields("mail_account_id", "engine_submission_id").Unique().StorageKey("mail_submissions_account_engine_submission_uidx").Annotations(entsql.IndexWhere("engine_submission_id IS NOT NULL")),
		index.Fields("mail_account_id", "engine_result_message_id").Unique().StorageKey("mail_submissions_account_engine_result_uidx").Annotations(entsql.IndexWhere("engine_result_message_id IS NOT NULL")),
		index.Fields("mail_account_id", "result_mail_message_id").Unique().StorageKey("mail_submissions_account_result_snapshot_uidx").Annotations(entsql.IndexWhere("result_mail_message_id IS NOT NULL")),
	}
}
