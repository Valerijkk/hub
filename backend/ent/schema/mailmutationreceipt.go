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

type MailMutationReceipt struct {
	ent.Schema
}

func (MailMutationReceipt) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Default(uuid.New).Immutable(),
		field.UUID("organization_id", uuid.UUID{}).Immutable(),
		field.UUID("actor_user_id", uuid.UUID{}).Immutable(),
		field.UUID("mail_account_id", uuid.UUID{}).Optional().Nillable().Immutable(),
		field.Enum("operation").Values("folder_create", "folder_update", "folder_delete", "label_create", "signature_create", "rule_create", "attachment_presign", "domain_create", "alias_create", "message_bulk", "draft_restart", "draft_abandon").Immutable(),
		field.Enum("key_source").Values("client", "internal").Immutable(),
		field.Enum("state").Values("in_progress", "unknown", "completed", "failed").Default("in_progress"),
		field.Bytes("idempotency_key_hash").MinLen(sha256.Size).MaxLen(sha256.Size).Immutable(),
		field.Bytes("request_hash").MinLen(sha256.Size).MaxLen(sha256.Size).Immutable(),
		field.Int32("fingerprint_version").Default(1).Min(1).Immutable(),
		field.Enum("target_kind").Values("folder", "label", "signature", "rule", "attachment", "domain", "identity", "draft").Optional().Nillable(),
		field.UUID("target_id", uuid.UUID{}).Optional().Nillable(),
		field.Enum("bulk_action").Values("move", "add_label", "remove_label", "mark_read", "mark_unread", "flag", "unflag", "archive", "junk", "not_junk", "trash", "restore", "pin", "unpin").Optional().Nillable(),
		field.Enum("argument_kind").Values("folder", "label").Optional().Nillable(),
		field.UUID("argument_id", uuid.UUID{}).Optional().Nillable(),
		field.Int16("original_http_status").Optional().Nillable().Min(100).Max(599),
		field.Enum("safe_code").Values("MAIL_FOLDER_NAME_CONFLICT", "MAIL_FOLDER_TREE_CONFLICT", "MAIL_FOLDER_LIMIT_EXCEEDED", "MAIL_FOLDER_RECONCILIATION_REQUIRED", "MAIL_RECONCILIATION_EXHAUSTED", "MAIL_LABEL_NAME_CONFLICT", "MAIL_LABEL_LIMIT_EXCEEDED", "MAIL_SIGNATURE_LIMIT_EXCEEDED", "MAIL_RULE_LIMIT_EXCEEDED", "MAIL_DOMAIN_UNAVAILABLE", "MAIL_DOMAIN_LIMIT_EXCEEDED", "MAIL_ALIAS_ADDRESS_UNAVAILABLE", "MAIL_ALIAS_ACCOUNT_NOT_ACTIVE", "MAIL_IDENTITY_LIMIT_EXCEEDED", "MAIL_BULK_EXPANSION_LIMIT", "MAIL_ENGINE_OUTCOME_UNKNOWN").Optional().Nillable(),
		field.Int32("reconcile_attempts").Default(0).Min(0).Max(10),
		field.Time("next_reconcile_at").Optional().Nillable(),
		field.UUID("reconcile_lease_token", uuid.UUID{}).Optional().Nillable(),
		field.Time("reconcile_lease_until").Optional().Nillable(),
		field.Time("terminal_at").Optional().Nillable(),
		field.Time("replay_until").Optional().Nillable(),
		field.Time("compacted_at").Optional().Nillable(),
		field.Time("created_at").Default(nowUTC).Immutable(),
		field.Time("updated_at").Default(nowUTC).UpdateDefault(nowUTC),
	}
}

func (MailMutationReceipt) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Checks(map[string]string{
			"mail_mutation_receipts_digest_version": "COALESCE(octet_length(idempotency_key_hash) = 32 AND octet_length(request_hash) = 32 AND fingerprint_version >= 1, FALSE)",
			"mail_mutation_receipts_target_pair": "((target_kind IS NULL AND target_id IS NULL) OR (target_kind IS NOT NULL AND target_id IS NOT NULL))",
			"mail_mutation_receipts_argument_pair": "((argument_kind IS NULL AND argument_id IS NULL) OR (argument_kind IS NOT NULL AND argument_id IS NOT NULL))",
			"mail_mutation_receipts_http_status_range": "(original_http_status IS NULL OR original_http_status BETWEEN 100 AND 599)",
			"mail_mutation_receipts_safe_code_closed_union": "(safe_code IS NULL OR safe_code IN ('MAIL_FOLDER_NAME_CONFLICT','MAIL_FOLDER_TREE_CONFLICT','MAIL_FOLDER_LIMIT_EXCEEDED','MAIL_FOLDER_RECONCILIATION_REQUIRED','MAIL_RECONCILIATION_EXHAUSTED','MAIL_LABEL_NAME_CONFLICT','MAIL_LABEL_LIMIT_EXCEEDED','MAIL_SIGNATURE_LIMIT_EXCEEDED','MAIL_RULE_LIMIT_EXCEEDED','MAIL_DOMAIN_UNAVAILABLE','MAIL_DOMAIN_LIMIT_EXCEEDED','MAIL_ALIAS_ADDRESS_UNAVAILABLE','MAIL_ALIAS_ACCOUNT_NOT_ACTIVE','MAIL_IDENTITY_LIMIT_EXCEEDED','MAIL_BULK_EXPANSION_LIMIT','MAIL_ENGINE_OUTCOME_UNKNOWN'))",
			"mail_mutation_receipts_reconcile_attempts_bounds": "(reconcile_attempts BETWEEN 0 AND 10)",
			"mail_mutation_receipts_reconcile_lease_pair": "((reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL) OR (reconcile_lease_token IS NOT NULL AND reconcile_lease_until IS NOT NULL))",
			"mail_mutation_receipts_terminal_replay_pair": "COALESCE((terminal_at IS NULL AND replay_until IS NULL) OR (terminal_at IS NOT NULL AND replay_until = terminal_at + INTERVAL '30 days'), FALSE)",
			"mail_mutation_receipts_compaction": "COALESCE(compacted_at IS NULL OR (terminal_at IS NOT NULL AND replay_until IS NOT NULL AND compacted_at >= replay_until), FALSE)",
			"mail_mutation_receipts_state_shape": "COALESCE(CASE state WHEN 'in_progress' THEN safe_code IS NULL AND original_http_status IS NULL AND terminal_at IS NULL AND replay_until IS NULL AND compacted_at IS NULL WHEN 'unknown' THEN original_http_status IS NULL AND terminal_at IS NULL AND replay_until IS NULL AND compacted_at IS NULL AND ((reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL) OR (reconcile_lease_token IS NOT NULL AND reconcile_lease_until IS NOT NULL)) AND ((operation IN ('folder_create','folder_update','folder_delete') AND safe_code='MAIL_FOLDER_RECONCILIATION_REQUIRED' AND reconcile_attempts BETWEEN 1 AND 10 AND next_reconcile_at IS NOT NULL) OR (operation='message_bulk' AND safe_code='MAIL_ENGINE_OUTCOME_UNKNOWN' AND reconcile_attempts BETWEEN 1 AND 9 AND next_reconcile_at IS NOT NULL) OR (operation='message_bulk' AND safe_code='MAIL_ENGINE_OUTCOME_UNKNOWN' AND reconcile_attempts=10 AND next_reconcile_at IS NULL AND reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL)) WHEN 'completed' THEN safe_code IS NULL AND original_http_status BETWEEN 200 AND 299 AND terminal_at IS NOT NULL AND replay_until = terminal_at + INTERVAL '30 days' AND next_reconcile_at IS NULL AND reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL WHEN 'failed' THEN safe_code IS NOT NULL AND original_http_status BETWEEN 400 AND 599 AND terminal_at IS NOT NULL AND replay_until = terminal_at + INTERVAL '30 days' AND next_reconcile_at IS NULL AND reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL ELSE FALSE END, FALSE)",
			"mail_mutation_receipts_reconciliation_exhausted_terminal": "COALESCE(safe_code IS DISTINCT FROM 'MAIL_RECONCILIATION_EXHAUSTED' OR (state='failed' AND original_http_status=409 AND reconcile_attempts=10 AND terminal_at IS NOT NULL AND replay_until=terminal_at + INTERVAL '30 days' AND compacted_at IS NULL AND next_reconcile_at IS NULL AND reconcile_lease_token IS NULL AND reconcile_lease_until IS NULL), FALSE)",
			"mail_mutation_receipts_operation_bulk_action": "COALESCE((operation='message_bulk' AND bulk_action IN ('move','add_label','remove_label','mark_read','mark_unread','flag','unflag','archive','junk','not_junk','trash','restore','pin','unpin')) OR (operation<>'message_bulk' AND bulk_action IS NULL), FALSE)",
			"mail_mutation_receipts_operation_target_status_safe_code": "COALESCE(CASE operation WHEN 'folder_create' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND argument_kind IS NULL AND argument_id IS NULL AND (target_kind IS NULL OR target_kind = 'folder')) OR (state = 'unknown' AND target_kind = 'folder' AND safe_code = 'MAIL_FOLDER_RECONCILIATION_REQUIRED' AND original_http_status IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'folder' AND safe_code IS NULL AND original_http_status = 201 AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'failed' AND target_kind = 'folder' AND argument_kind IS NULL AND argument_id IS NULL AND ((safe_code IN ('MAIL_FOLDER_NAME_CONFLICT','MAIL_FOLDER_TREE_CONFLICT','MAIL_RECONCILIATION_EXHAUSTED') AND original_http_status = 409) OR (safe_code = 'MAIL_FOLDER_LIMIT_EXCEEDED' AND original_http_status = 422)))) WHEN 'folder_update' THEN (target_kind = 'folder' AND argument_kind IS NULL AND argument_id IS NULL AND ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL) OR (state = 'unknown' AND safe_code = 'MAIL_FOLDER_RECONCILIATION_REQUIRED' AND original_http_status IS NULL) OR (state = 'completed' AND safe_code IS NULL AND original_http_status = 200) OR (state = 'failed' AND safe_code = 'MAIL_RECONCILIATION_EXHAUSTED' AND original_http_status = 409))) WHEN 'folder_delete' THEN (target_kind = 'folder' AND argument_kind IS NULL AND argument_id IS NULL AND ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL) OR (state = 'unknown' AND safe_code = 'MAIL_FOLDER_RECONCILIATION_REQUIRED' AND original_http_status IS NULL) OR (state = 'completed' AND safe_code IS NULL AND original_http_status = 204) OR (state = 'failed' AND safe_code = 'MAIL_RECONCILIATION_EXHAUSTED' AND original_http_status = 409))) WHEN 'label_create' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'label' AND safe_code IS NULL AND original_http_status = 201 AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'failed' AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL AND ((safe_code = 'MAIL_LABEL_NAME_CONFLICT' AND original_http_status = 409) OR (safe_code = 'MAIL_LABEL_LIMIT_EXCEEDED' AND original_http_status = 422)))) WHEN 'signature_create' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'signature' AND safe_code IS NULL AND original_http_status = 201 AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'failed' AND target_kind IS NULL AND safe_code = 'MAIL_SIGNATURE_LIMIT_EXCEEDED' AND original_http_status = 422 AND argument_kind IS NULL AND argument_id IS NULL)) WHEN 'rule_create' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'rule' AND safe_code IS NULL AND original_http_status = 201 AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'failed' AND target_kind IS NULL AND safe_code = 'MAIL_RULE_LIMIT_EXCEEDED' AND original_http_status = 422 AND argument_kind IS NULL AND argument_id IS NULL)) WHEN 'attachment_presign' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'attachment' AND safe_code IS NULL AND original_http_status = 201 AND argument_kind IS NULL AND argument_id IS NULL)) WHEN 'domain_create' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'domain' AND safe_code IS NULL AND original_http_status = 202 AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'failed' AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL AND ((safe_code = 'MAIL_DOMAIN_UNAVAILABLE' AND original_http_status = 409) OR (safe_code = 'MAIL_DOMAIN_LIMIT_EXCEEDED' AND original_http_status = 422)))) WHEN 'alias_create' THEN ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'completed' AND target_kind = 'identity' AND safe_code IS NULL AND original_http_status = 202 AND argument_kind IS NULL AND argument_id IS NULL) OR (state = 'failed' AND target_kind IS NULL AND argument_kind IS NULL AND argument_id IS NULL AND ((safe_code IN ('MAIL_ALIAS_ADDRESS_UNAVAILABLE','MAIL_ALIAS_ACCOUNT_NOT_ACTIVE') AND original_http_status = 409) OR (safe_code = 'MAIL_IDENTITY_LIMIT_EXCEEDED' AND original_http_status = 422)))) WHEN 'message_bulk' THEN (target_kind IS NULL AND target_id IS NULL AND ((bulk_action = 'move' AND argument_kind = 'folder' AND argument_id IS NOT NULL) OR (bulk_action IN ('add_label','remove_label') AND argument_kind = 'label' AND argument_id IS NOT NULL) OR (bulk_action IN ('mark_read','mark_unread','flag','unflag','archive','junk','not_junk','trash','restore','pin','unpin') AND argument_kind IS NULL AND argument_id IS NULL)) AND ((state = 'in_progress' AND safe_code IS NULL AND original_http_status IS NULL) OR (state = 'unknown' AND safe_code = 'MAIL_ENGINE_OUTCOME_UNKNOWN' AND original_http_status IS NULL) OR (state = 'completed' AND safe_code IS NULL AND original_http_status = 200) OR (state = 'failed' AND safe_code = 'MAIL_BULK_EXPANSION_LIMIT' AND original_http_status = 422))) WHEN 'draft_restart' THEN (state = 'completed' AND target_kind = 'draft' AND target_id IS NOT NULL AND bulk_action IS NULL AND argument_kind IS NULL AND argument_id IS NULL AND safe_code IS NULL AND original_http_status = 200 AND reconcile_attempts = 0) WHEN 'draft_abandon' THEN (state = 'completed' AND target_kind = 'draft' AND target_id IS NOT NULL AND bulk_action IS NULL AND argument_kind IS NULL AND argument_id IS NULL AND safe_code IS NULL AND original_http_status = 200 AND reconcile_attempts = 0) ELSE FALSE END, FALSE)",
			"mail_mutation_receipts_operation_key_source_account": "COALESCE((operation IN ('folder_update','folder_delete') AND key_source='internal' AND mail_account_id IS NOT NULL) OR (operation='domain_create' AND key_source='client' AND mail_account_id IS NULL) OR (operation IN ('folder_create','label_create','signature_create','rule_create','attachment_presign','alias_create','message_bulk','draft_restart','draft_abandon') AND key_source='client' AND mail_account_id IS NOT NULL), FALSE)",
		}),
	}
}

func (MailMutationReceipt) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("organization", Organization.Type).Ref("mail_mutation_receipts").Field("organization_id").Unique().Required().Immutable().Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.From("account", MailAccount.Type).Ref("mutation_receipts").Field("mail_account_id").Unique().Immutable().Annotations(entsql.OnDelete(entsql.NoAction)),
		edge.To("targets", MailMutationReceiptTarget.Type),
		edge.To("members", MailMutationReceiptMember.Type),
		edge.To("folders", MailFolder.Type),
	}
}

func (MailMutationReceipt) Indexes() []ent.Index {
	return []ent.Index{
		index.Fields("organization_id", "actor_user_id", "operation", "idempotency_key_hash").Unique(),
	}
}
