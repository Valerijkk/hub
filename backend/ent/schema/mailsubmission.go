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
		field.Int("engine_binding_version").Default(1).Immutable(),
		field.String("engine_account_id").Validate(validateMailMessageOpaqueID).Immutable(),
		field.String("engine_draft_key").Validate(validateMailMessageOpaqueID).Immutable(),
		field.String("engine_message_id").Validate(validateMailMessageOpaqueID).Immutable(),
		field.String("engine_expected_version").Validate(validateMailMessageOpaqueID).Immutable(),
		field.Int64("engine_max_message_bytes").Min(1).Max(52_428_800).Immutable(),
		field.Bytes("engine_binding_hash").MinLen(sha256.Size).MaxLen(sha256.Size).Immutable(),
		field.Enum("status").Values("queued", "authorizing", "submitting", "unknown", "accepted", "sent", "failed"),
		field.UUID("command_lease_token", uuid.UUID{}).Optional().Nillable(),
		field.Int64("command_fence").Default(0).NonNegative(),
		field.Time("command_lease_expires_at").Optional().Nillable(),
		field.Time("dispatch_authorized_at").Optional().Nillable(),
		field.Bytes("dispatch_authorization_hash").Optional().Nillable().MinLen(sha256.Size).MaxLen(sha256.Size),
		field.String("provider_submission_id").Optional().Nillable().Validate(validateMailMessageOpaqueID),
		field.String("provider_message_id").Optional().Nillable().Validate(validateMailMessageOpaqueID),
		field.UUID("result_mail_message_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("last_error_code").Values("MAIL_SUBMISSION_OUTCOME_UNKNOWN", "MAIL_SUBMISSION_RECONCILIATION_REQUIRED", "MAIL_SUBMISSION_RECONCILIATION_EXHAUSTED", "MAILBOX_UNAVAILABLE", "MAIL_SUBMISSION_REJECTED", "MAIL_SUBMISSION_POLICY", "MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN", "MAIL_SUBMISSION_IDENTITY_UNAVAILABLE", "MAIL_SUBMISSION_RECIPIENT_LIMIT", "MAIL_SUBMISSION_SIZE_LIMIT", "MAIL_SUBMISSION_RETRY_EXHAUSTED").Optional().Nillable(),
		field.Time("next_attempt_at").Optional().Nillable(),
		field.Int32("attempt_count").Default(0).Min(0).Max(8),
		field.Int32("reconcile_attempts").Default(0).Min(0).Max(10),
		field.Time("next_reconcile_at").Optional().Nillable(),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (MailSubmission) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_submissions_key":                         "COALESCE(octet_length(engine_submission_key) BETWEEN 1 AND 1024, FALSE)",
			"mail_submissions_hashes":                      "COALESCE(octet_length(idempotency_key_hash) = 32 AND octet_length(request_hash) = 32, FALSE)",
			"mail_submissions_binding":                     "COALESCE(engine_binding_version = 1 AND octet_length(engine_submission_key) BETWEEN 1 AND 1024 AND octet_length(engine_account_id) BETWEEN 1 AND 1024 AND octet_length(engine_draft_key) BETWEEN 1 AND 1024 AND octet_length(engine_message_id) BETWEEN 1 AND 1024 AND octet_length(engine_expected_version) BETWEEN 1 AND 1024 AND engine_max_message_bytes BETWEEN 1 AND 52428800 AND octet_length(engine_binding_hash) = 32, FALSE)",
			"mail_submissions_provider_pair":               "COALESCE((provider_submission_id IS NULL AND provider_message_id IS NULL) OR (octet_length(provider_submission_id) BETWEEN 1 AND 1024 AND octet_length(provider_message_id) BETWEEN 1 AND 1024), FALSE)",
			"mail_submissions_command_lease_pair":          "COALESCE((command_lease_token IS NULL) = (command_lease_expires_at IS NULL), FALSE)",
			"mail_submissions_command_fence":               "COALESCE(command_fence >= 0, FALSE)",
			"mail_submissions_dispatch_authorization_pair": "COALESCE((dispatch_authorized_at IS NULL) = (dispatch_authorization_hash IS NULL) AND (dispatch_authorization_hash IS NULL OR octet_length(dispatch_authorization_hash) = 32), FALSE)",
			"mail_submissions_attempts":                    "COALESCE(attempt_count BETWEEN 0 AND 8, FALSE)",
			"mail_submissions_reconcile_attempts":          "COALESCE(reconcile_attempts BETWEEN 0 AND 10, FALSE)",
			"mail_submissions_status_enum":                 "COALESCE(status IN ('queued','authorizing','submitting','unknown','accepted','sent','failed'), FALSE)",
			"mail_submissions_state":                       "COALESCE((status = 'queued' AND mail_draft_id IS NOT NULL AND command_lease_token IS NULL AND command_lease_expires_at IS NULL AND dispatch_authorized_at IS NULL AND dispatch_authorization_hash IS NULL AND provider_submission_id IS NULL AND provider_message_id IS NULL AND result_mail_message_id IS NULL AND last_error_code IS NULL AND next_attempt_at IS NOT NULL AND attempt_count BETWEEN 0 AND 7 AND reconcile_attempts = 0 AND next_reconcile_at IS NULL) OR (status = 'authorizing' AND mail_draft_id IS NOT NULL AND command_lease_token IS NOT NULL AND command_lease_expires_at IS NOT NULL AND dispatch_authorized_at IS NULL AND dispatch_authorization_hash IS NULL AND provider_submission_id IS NULL AND provider_message_id IS NULL AND result_mail_message_id IS NULL AND last_error_code IS NULL AND next_attempt_at IS NULL AND attempt_count BETWEEN 1 AND 8 AND command_fence >= 1 AND reconcile_attempts = 0 AND next_reconcile_at IS NULL) OR (status = 'submitting' AND mail_draft_id IS NOT NULL AND command_lease_token IS NOT NULL AND command_lease_expires_at IS NOT NULL AND dispatch_authorized_at IS NOT NULL AND dispatch_authorization_hash IS NOT NULL AND provider_submission_id IS NULL AND provider_message_id IS NULL AND result_mail_message_id IS NULL AND last_error_code IS NULL AND next_attempt_at IS NULL AND attempt_count BETWEEN 1 AND 8 AND command_fence >= 1 AND reconcile_attempts = 0 AND next_reconcile_at IS NULL) OR (status = 'unknown' AND mail_draft_id IS NOT NULL AND dispatch_authorized_at IS NOT NULL AND dispatch_authorization_hash IS NOT NULL AND provider_submission_id IS NULL AND provider_message_id IS NULL AND result_mail_message_id IS NULL AND next_attempt_at IS NULL AND attempt_count BETWEEN 1 AND 8 AND command_fence >= 1 AND ((command_lease_token IS NULL AND command_lease_expires_at IS NULL AND reconcile_attempts BETWEEN 0 AND 9 AND next_reconcile_at IS NOT NULL AND last_error_code IN ('MAIL_SUBMISSION_OUTCOME_UNKNOWN','MAIL_SUBMISSION_RECONCILIATION_REQUIRED','MAILBOX_UNAVAILABLE')) OR (command_lease_token IS NOT NULL AND command_lease_expires_at IS NOT NULL AND reconcile_attempts BETWEEN 1 AND 10 AND next_reconcile_at IS NULL AND last_error_code IN ('MAIL_SUBMISSION_OUTCOME_UNKNOWN','MAIL_SUBMISSION_RECONCILIATION_REQUIRED','MAILBOX_UNAVAILABLE')) OR (command_lease_token IS NULL AND command_lease_expires_at IS NULL AND reconcile_attempts = 10 AND next_reconcile_at IS NULL AND last_error_code = 'MAIL_SUBMISSION_RECONCILIATION_EXHAUSTED'))) OR (status = 'accepted' AND mail_draft_id IS NOT NULL AND dispatch_authorized_at IS NOT NULL AND dispatch_authorization_hash IS NOT NULL AND provider_submission_id IS NOT NULL AND provider_message_id IS NOT NULL AND result_mail_message_id IS NULL AND last_error_code IS NULL AND next_attempt_at IS NULL AND attempt_count BETWEEN 1 AND 8 AND command_fence >= 1 AND reconcile_attempts BETWEEN 0 AND 10 AND next_reconcile_at IS NULL) OR (status = 'sent' AND command_lease_token IS NULL AND command_lease_expires_at IS NULL AND dispatch_authorized_at IS NOT NULL AND dispatch_authorization_hash IS NOT NULL AND provider_submission_id IS NOT NULL AND provider_message_id IS NOT NULL AND result_mail_message_id IS NOT NULL AND last_error_code IS NULL AND next_attempt_at IS NULL AND attempt_count BETWEEN 1 AND 8 AND command_fence >= 1 AND reconcile_attempts BETWEEN 0 AND 10 AND next_reconcile_at IS NULL) OR (status = 'failed' AND command_lease_token IS NULL AND command_lease_expires_at IS NULL AND dispatch_authorized_at IS NULL AND dispatch_authorization_hash IS NULL AND provider_submission_id IS NULL AND provider_message_id IS NULL AND result_mail_message_id IS NULL AND last_error_code IN ('MAIL_SUBMISSION_REJECTED','MAIL_SUBMISSION_POLICY','MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN','MAIL_SUBMISSION_IDENTITY_UNAVAILABLE','MAIL_SUBMISSION_RECIPIENT_LIMIT','MAIL_SUBMISSION_SIZE_LIMIT','MAIL_SUBMISSION_RETRY_EXHAUSTED') AND next_attempt_at IS NULL AND attempt_count BETWEEN 1 AND 8 AND (last_error_code <> 'MAIL_SUBMISSION_RETRY_EXHAUSTED' OR attempt_count = 8) AND reconcile_attempts = 0 AND next_reconcile_at IS NULL), FALSE)",
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
		index.Fields("mail_account_id", "provider_submission_id").Unique().StorageKey("mail_submissions_account_provider_submission_uidx").Annotations(entsql.IndexWhere("provider_submission_id IS NOT NULL")),
		index.Fields("mail_account_id", "provider_message_id").Unique().StorageKey("mail_submissions_account_provider_message_uidx").Annotations(entsql.IndexWhere("provider_message_id IS NOT NULL")),
		index.Fields("mail_account_id", "result_mail_message_id").Unique().StorageKey("mail_submissions_account_result_snapshot_uidx").Annotations(entsql.IndexWhere("result_mail_message_id IS NOT NULL")),
		index.Fields("mail_draft_id").Unique().StorageKey("mail_submissions_active_draft_uidx").Annotations(entsql.IndexWhere("status IN ('queued','authorizing','submitting','unknown','accepted')")),
		index.Fields("next_attempt_at", "id").StorageKey("mail_submissions_queued_due_idx").Annotations(entsql.IndexWhere("status = 'queued'")),
		index.Fields("command_lease_expires_at", "id").StorageKey("mail_submissions_stale_lease_idx").Annotations(entsql.IndexWhere("status IN ('authorizing','submitting','unknown','accepted')")),
		index.Fields("next_reconcile_at", "id").StorageKey("mail_submissions_unknown_reconcile_idx").Annotations(entsql.IndexWhere("status = 'unknown' AND next_reconcile_at IS NOT NULL")),
		index.Fields("updated_at", "id").StorageKey("mail_submissions_accepted_finalizer_due_idx").Annotations(entsql.IndexWhere("status = 'accepted'")),
	}
}
