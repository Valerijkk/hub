package mail_test

import (
	"crypto/sha256"
	"testing"

	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestMailMutationReceiptSchemas_HaveExactFieldsDefaultsAndImmutability(t *testing.T) {
	receipt := entschema.MailMutationReceipt{}
	target := entschema.MailMutationReceiptTarget{}
	member := entschema.MailMutationReceiptMember{}
	assertNoMixins(t, receipt.Mixin())
	assertNoMixins(t, target.Mixin())
	assertNoMixins(t, member.Mixin())

	receiptFields := fieldDescriptors(receipt.Fields())
	assertExactNames(t, "MailMutationReceipt fields", fieldNames(receiptFields), []string{
		"actor_user_id", "argument_id", "argument_kind", "bulk_action", "compacted_at", "created_at",
		"fingerprint_version", "id", "idempotency_key_hash", "key_source", "mail_account_id",
		"next_reconcile_at", "operation", "organization_id", "original_http_status", "reconcile_attempts",
		"reconcile_lease_token", "reconcile_lease_until", "replay_until", "request_hash", "safe_code",
		"state", "target_id", "target_kind", "terminal_at", "updated_at",
	})
	for _, name := range []string{"id", "organization_id", "actor_user_id", "mail_account_id", "target_id", "argument_id", "reconcile_lease_token"} {
		assertFieldType(t, name, receiptFields[name], field.TypeUUID)
	}
	for _, name := range []string{"idempotency_key_hash", "request_hash"} {
		assertFieldType(t, name, receiptFields[name], field.TypeBytes)
		assertBytesValidators(t, name, receiptFields[name], []struct {
			value []byte
			want  bool
		}{{make([]byte, sha256.Size-1), false}, {make([]byte, sha256.Size), true}, {make([]byte, sha256.Size+1), false}})
	}
	for _, name := range []string{"fingerprint_version", "reconcile_attempts"} {
		assertFieldType(t, name, receiptFields[name], field.TypeInt32)
	}
	assertFieldType(t, "original_http_status", receiptFields["original_http_status"], field.TypeInt16)
	assertInt32Validators(t, "fingerprint_version", receiptFields["fingerprint_version"], []struct {
		value int32
		want  bool
	}{{0, false}, {1, true}})
	assertInt32Validators(t, "reconcile_attempts", receiptFields["reconcile_attempts"], []struct {
		value int32
		want  bool
	}{{-1, false}, {0, true}, {10, true}, {11, false}})
	assertInt16Validators(t, "original_http_status", receiptFields["original_http_status"], []struct {
		value int16
		want  bool
	}{{99, false}, {100, true}, {599, true}, {600, false}})
	for _, name := range []string{"operation", "key_source", "state", "target_kind", "bulk_action", "argument_kind", "safe_code"} {
		assertFieldType(t, name, receiptFields[name], field.TypeEnum)
	}
	assertExactEnumValues(t, "operation", receiptFields["operation"], []string{"folder_create", "folder_update", "folder_delete", "label_create", "signature_create", "rule_create", "attachment_presign", "domain_create", "alias_create", "message_bulk", "draft_restart", "draft_abandon"})
	assertExactEnumValues(t, "key_source", receiptFields["key_source"], []string{"client", "internal"})
	assertExactEnumValues(t, "state", receiptFields["state"], []string{"in_progress", "unknown", "completed", "failed"})
	assertExactEnumValues(t, "target_kind", receiptFields["target_kind"], []string{"folder", "label", "signature", "rule", "attachment", "domain", "identity", "draft"})
	assertExactEnumValues(t, "bulk_action", receiptFields["bulk_action"], []string{"move", "add_label", "remove_label", "mark_read", "mark_unread", "flag", "unflag", "archive", "junk", "not_junk", "trash", "restore", "pin", "unpin"})
	assertExactEnumValues(t, "argument_kind", receiptFields["argument_kind"], []string{"folder", "label"})
	assertExactEnumValues(t, "safe_code", receiptFields["safe_code"], []string{"MAIL_FOLDER_NAME_CONFLICT", "MAIL_FOLDER_TREE_CONFLICT", "MAIL_FOLDER_LIMIT_EXCEEDED", "MAIL_FOLDER_RECONCILIATION_REQUIRED", "MAIL_RECONCILIATION_EXHAUSTED", "MAIL_LABEL_NAME_CONFLICT", "MAIL_LABEL_LIMIT_EXCEEDED", "MAIL_SIGNATURE_LIMIT_EXCEEDED", "MAIL_RULE_LIMIT_EXCEEDED", "MAIL_DOMAIN_UNAVAILABLE", "MAIL_DOMAIN_LIMIT_EXCEEDED", "MAIL_ALIAS_ADDRESS_UNAVAILABLE", "MAIL_ALIAS_ACCOUNT_NOT_ACTIVE", "MAIL_IDENTITY_LIMIT_EXCEEDED", "MAIL_BULK_EXPANSION_LIMIT", "MAIL_ENGINE_OUTCOME_UNKNOWN"})
	for _, name := range []string{"mail_account_id", "target_kind", "target_id", "bulk_action", "argument_kind", "argument_id", "original_http_status", "safe_code", "next_reconcile_at", "reconcile_lease_token", "reconcile_lease_until", "terminal_at", "replay_until", "compacted_at"} {
		assertOptionalNillable(t, name, receiptFields[name])
	}
	for _, name := range []string{"id", "organization_id", "actor_user_id", "mail_account_id", "operation", "key_source", "idempotency_key_hash", "request_hash", "fingerprint_version"} {
		if !receiptFields[name].Immutable {
			t.Fatalf("MailMutationReceipt.%s must be immutable", name)
		}
	}
	assertExactDefaults(t, receiptFields, map[string]any{"fingerprint_version": int32(1), "reconcile_attempts": int32(0), "state": "in_progress"})
	assertReceiptTimestamps(t, receiptFields)

	targetFields := fieldDescriptors(target.Fields())
	assertExactNames(t, "MailMutationReceiptTarget fields", fieldNames(targetFields), []string{
		"affected_count", "created_at", "failed_count", "id", "receipt_id", "request_index", "safe_code",
		"status", "succeeded_count", "target_id", "target_type", "training_status", "updated_at",
	})
	for _, name := range []string{"id", "receipt_id", "target_id"} {
		assertFieldType(t, name, targetFields[name], field.TypeUUID)
	}
	for _, name := range []string{"request_index", "affected_count", "succeeded_count", "failed_count"} {
		assertFieldType(t, name, targetFields[name], field.TypeInt32)
	}
	assertExactEnumValues(t, "target_type", targetFields["target_type"], []string{"message", "thread"})
	assertExactEnumValues(t, "status", targetFields["status"], []string{"succeeded", "partial", "failed", "unknown"})
	assertExactEnumValues(t, "training_status", targetFields["training_status"], []string{"not_requested", "accepted", "rejected", "unknown"})
	assertExactEnumValues(t, "safe_code", targetFields["safe_code"], []string{"MAIL_ACTION_REJECTED", "MAIL_CLASSIFIER_REJECTED", "MAIL_TARGET_UNAVAILABLE", "MAIL_FOLDER_UNAVAILABLE", "MAIL_LABEL_UNAVAILABLE", "MAIL_ENGINE_OUTCOME_UNKNOWN", "MAIL_CLASSIFIER_OUTCOME_UNKNOWN"})
	for _, name := range []string{"request_index"} {
		assertInt32Validators(t, name, targetFields[name], []struct {
			value int32
			want  bool
		}{{-1, false}, {0, true}, {199, true}, {200, false}})
	}
	for _, name := range []string{"affected_count", "succeeded_count", "failed_count"} {
		assertInt32Validators(t, name, targetFields[name], []struct {
			value int32
			want  bool
		}{{-1, false}, {0, true}, {200, true}, {201, false}})
	}
	assertOptionalNillable(t, "safe_code", targetFields["safe_code"])
	for _, name := range []string{"id", "receipt_id", "request_index", "target_type", "target_id"} {
		if !targetFields[name].Immutable {
			t.Fatalf("MailMutationReceiptTarget.%s must be immutable", name)
		}
	}
	assertReceiptTimestamps(t, targetFields)

	memberFields := fieldDescriptors(member.Fields())
	assertExactNames(t, "MailMutationReceiptMember fields", fieldNames(memberFields), []string{"created_at", "id", "message_id", "receipt_id", "target_id", "updated_at"})
	for _, name := range []string{"id", "receipt_id", "target_id", "message_id"} {
		assertFieldType(t, name, memberFields[name], field.TypeUUID)
		if !memberFields[name].Immutable {
			t.Fatalf("MailMutationReceiptMember.%s must be immutable", name)
		}
	}
	assertReceiptTimestamps(t, memberFields)
}

func TestMailMutationReceiptSchemas_HaveExactChecksAndIndexes(t *testing.T) {
	assertExactChecks(t, (entschema.MailMutationReceipt{}).Annotations(), map[string]string{
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
	})
	assertExactChecks(t, (entschema.MailMutationReceiptTarget{}).Annotations(), map[string]string{
		"mail_mutation_receipt_targets_request_index_bounds": "(request_index BETWEEN 0 AND 199)",
		"mail_mutation_receipt_targets_counts_bounds": "(affected_count BETWEEN 0 AND 200 AND succeeded_count BETWEEN 0 AND 200 AND failed_count BETWEEN 0 AND 200)",
		"mail_mutation_receipt_targets_counts_sum": "(succeeded_count + failed_count <= affected_count)",
		"mail_mutation_receipt_targets_outcome_training": "COALESCE(CASE status WHEN 'succeeded' THEN affected_count >= 1 AND succeeded_count = affected_count AND failed_count = 0 AND safe_code IS NULL AND training_status IN ('not_requested','accepted') WHEN 'partial' THEN affected_count >= 2 AND succeeded_count >= 1 AND failed_count >= 1 AND succeeded_count + failed_count = affected_count AND ((safe_code = 'MAIL_ACTION_REJECTED' AND training_status = 'not_requested') OR (safe_code = 'MAIL_CLASSIFIER_REJECTED' AND training_status = 'rejected')) WHEN 'failed' THEN ((affected_count = 0 AND succeeded_count = 0 AND failed_count = 0 AND safe_code IN ('MAIL_TARGET_UNAVAILABLE','MAIL_FOLDER_UNAVAILABLE','MAIL_LABEL_UNAVAILABLE') AND training_status = 'not_requested') OR (affected_count >= 1 AND succeeded_count = 0 AND failed_count = affected_count AND ((safe_code = 'MAIL_ACTION_REJECTED' AND training_status = 'not_requested') OR (safe_code = 'MAIL_CLASSIFIER_REJECTED' AND training_status = 'rejected')))) WHEN 'unknown' THEN affected_count >= 1 AND succeeded_count + failed_count < affected_count AND ((safe_code = 'MAIL_ENGINE_OUTCOME_UNKNOWN' AND training_status = 'not_requested') OR (safe_code = 'MAIL_CLASSIFIER_OUTCOME_UNKNOWN' AND training_status = 'unknown')) ELSE FALSE END, FALSE)",
	})
	assertExactChecks(t, (entschema.MailMutationReceiptMember{}).Annotations(), map[string]string{})

	assertExactIndexes(t, (entschema.MailMutationReceipt{}).Indexes(), []wantedIndex{{fields: []string{"organization_id", "actor_user_id", "operation", "idempotency_key_hash"}, unique: true}})
	assertExactIndexes(t, (entschema.MailMutationReceiptTarget{}).Indexes(), []wantedIndex{{fields: []string{"receipt_id", "request_index"}, unique: true}, {fields: []string{"receipt_id", "target_type", "target_id"}, unique: true}})
	assertExactIndexes(t, (entschema.MailMutationReceiptMember{}).Indexes(), []wantedIndex{{fields: []string{"receipt_id", "message_id"}, unique: true}, {fields: []string{"target_id", "message_id"}, unique: true}})
}

func TestMailMutationReceiptSchemas_HaveOnlyAuthorizedRelations(t *testing.T) {
	receiptEdges := edgeDescriptors((entschema.MailMutationReceipt{}).Edges())
	assertExactNames(t, "MailMutationReceipt edges", edgeNames(receiptEdges), []string{"account", "folders", "members", "organization", "targets"})
	assertReceiptParentEdge(t, "MailMutationReceipt.organization", receiptEdges["organization"], "organization_id", "mail_mutation_receipts", true, entsql.NoAction)
	assertReceiptParentEdge(t, "MailMutationReceipt.account", receiptEdges["account"], "mail_account_id", "mutation_receipts", false, entsql.NoAction)
	assertReceiptChildEdge(t, "MailMutationReceipt.targets", receiptEdges["targets"])
	assertReceiptChildEdge(t, "MailMutationReceipt.members", receiptEdges["members"])
	assertReceiptChildEdge(t, "MailMutationReceipt.folders", receiptEdges["folders"])

	targetEdges := edgeDescriptors((entschema.MailMutationReceiptTarget{}).Edges())
	assertExactNames(t, "MailMutationReceiptTarget edges", edgeNames(targetEdges), []string{"members", "receipt"})
	assertReceiptParentEdge(t, "MailMutationReceiptTarget.receipt", targetEdges["receipt"], "receipt_id", "targets", true, entsql.Cascade)
	assertReceiptChildEdge(t, "MailMutationReceiptTarget.members", targetEdges["members"])

	memberEdges := edgeDescriptors((entschema.MailMutationReceiptMember{}).Edges())
	assertExactNames(t, "MailMutationReceiptMember edges", edgeNames(memberEdges), []string{"receipt", "target"})
	assertReceiptParentEdge(t, "MailMutationReceiptMember.receipt", memberEdges["receipt"], "receipt_id", "members", true, entsql.Cascade)
	assertReceiptParentEdge(t, "MailMutationReceiptMember.target", memberEdges["target"], "target_id", "members", true, entsql.Cascade)
}

func assertReceiptTimestamps(t *testing.T, fields map[string]*field.Descriptor) {
	t.Helper()
	for _, name := range []string{"created_at", "updated_at"} {
		assertFieldType(t, name, fields[name], field.TypeTime)
		assertRequired(t, name, fields[name])
	}
	if fields["created_at"].Default == nil || !fields["created_at"].Immutable {
		t.Fatal("created_at must have an immutable creation default")
	}
	if fields["updated_at"].Default == nil || fields["updated_at"].UpdateDefault == nil {
		t.Fatal("updated_at must have creation and update defaults")
	}
}

func assertInt16Validators(t *testing.T, name string, descriptor *field.Descriptor, cases []struct {
	value int16
	want  bool
}) {
	t.Helper()
	for _, testCase := range cases {
		accepted := true
		for _, rawValidator := range descriptor.Validators {
			if validator, ok := rawValidator.(func(int16) error); ok && validator(testCase.value) != nil {
				accepted = false
			}
		}
		if accepted != testCase.want {
			t.Fatalf("%s accepts %d = %t, want %t", name, testCase.value, accepted, testCase.want)
		}
	}
}

func assertReceiptParentEdge(t *testing.T, name string, descriptor *edge.Descriptor, fieldName, ref string, required bool, onDelete entsql.ReferenceOption) {
	t.Helper()
	if descriptor == nil || !descriptor.Inverse || !descriptor.Unique || descriptor.Required != required || descriptor.Field != fieldName || descriptor.RefName != ref {
		t.Fatalf("%s = %#v, want inverse unique required:%t field:%q ref:%q", name, descriptor, required, fieldName, ref)
	}
	assertOnDelete(t, name, descriptor.Annotations, onDelete)
}

func assertReceiptChildEdge(t *testing.T, name string, descriptor *edge.Descriptor) {
	t.Helper()
	if descriptor == nil || descriptor.Inverse || descriptor.Unique || descriptor.Required || descriptor.RefName != "" {
		t.Fatalf("%s = %#v, want owning non-unique optional child edge", name, descriptor)
	}
}
