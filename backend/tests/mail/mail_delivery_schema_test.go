package mail_test

import (
	"crypto/sha256"
	"reflect"
	"strings"
	"testing"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestMailDeliverySchemas_RejectWrongFieldsDefaultsAndImmutability(t *testing.T) {
	object := entschema.MailObject{}
	assertExactOrderedFieldNames(t, "MailObject fields", object.Fields(), []string{
		"id", "deletion_id", "organization_id", "mail_account_id", "uploader_user_id", "draft_client_id", "bucket", "object_key", "expected_size_bytes", "actual_size_bytes", "filename", "media_type", "sha256", "upload_state", "expires_at", "upload_drain_until", "delete_lease_token", "delete_lease_generation", "delete_lease_until", "delete_attempts", "next_delete_at", "delete_terminal_at", "created_at", "updated_at",
	})
	objectFields := fieldDescriptors(object.Fields())
	for _, name := range []string{"id", "deletion_id", "organization_id", "mail_account_id", "uploader_user_id", "draft_client_id", "delete_lease_token"} {
		assertFieldType(t, name, objectFields[name], field.TypeUUID)
	}
	for _, name := range []string{"deletion_id", "organization_id", "mail_account_id", "uploader_user_id", "draft_client_id", "bucket", "object_key", "expected_size_bytes", "filename", "media_type", "expires_at", "upload_drain_until"} {
		if !objectFields[name].Immutable {
			t.Fatalf("MailObject.%s must be immutable", name)
		}
	}
	for _, name := range []string{"actual_size_bytes", "sha256", "delete_lease_token", "delete_lease_until", "next_delete_at", "delete_terminal_at"} {
		assertOptionalNillable(t, name, objectFields[name])
	}
	assertExactDefaults(t, objectFields, map[string]any{"upload_state": "pending", "delete_lease_generation": int64(0), "delete_attempts": int32(0)})
	assertOnlyDefaults(t, objectFields, map[string]bool{"id": true, "deletion_id": true, "upload_state": true, "delete_lease_generation": true, "delete_attempts": true, "created_at": true, "updated_at": true})
	assertExactEnumValues(t, "upload_state", objectFields["upload_state"], []string{"pending", "confirmed", "orphaned", "deleted"})
	for _, name := range []string{"expected_size_bytes", "actual_size_bytes"} {
		assertFieldType(t, name, objectFields[name], field.TypeInt64)
	}
	assertInt64Validators(t, "expected_size_bytes", objectFields["expected_size_bytes"], []struct{ value int64; want bool }{{0, false}, {1, true}, {26_214_400, true}, {26_214_401, false}})
	assertInt64Validators(t, "actual_size_bytes", objectFields["actual_size_bytes"], []struct{ value int64; want bool }{{0, false}, {1, true}, {26_214_400, true}, {26_214_401, false}})
	assertInt64Validators(t, "delete_lease_generation", objectFields["delete_lease_generation"], []struct{ value int64; want bool }{{-1, false}, {0, true}})
	assertInt32Validators(t, "delete_attempts", objectFields["delete_attempts"], []struct{ value int32; want bool }{{-1, false}, {0, true}, {10, true}, {11, false}})
	assertBytesValidators(t, "sha256", objectFields["sha256"], []struct{ value []byte; want bool }{{make([]byte, sha256.Size-1), false}, {make([]byte, sha256.Size), true}, {make([]byte, sha256.Size+1), false}})
	assertTimestampsWithoutMixin(t, objectFields)
	for _, forbidden := range []string{"scan_status", "file_metadata_id", "project_id"} {
		if _, found := objectFields[forbidden]; found {
			t.Fatalf("MailObject must not persist forbidden %q", forbidden)
		}
	}

	attachment := entschema.MailAttachment{}
	assertExactOrderedFieldNames(t, "MailAttachment fields", attachment.Fields(), []string{
		"id", "mail_message_id", "mail_draft_id", "pending_mail_draft_id", "draft_client_id", "storage_object_id", "engine_blob_id", "content_id", "filename", "media_type", "size_bytes", "disposition", "scan_status", "sha256", "scan_lease_token", "scan_lease_generation", "scan_lease_until", "scan_attempts", "next_scan_at", "scan_terminal_at", "created_at", "updated_at",
	})
	attachmentFields := fieldDescriptors(attachment.Fields())
	for _, name := range []string{"mail_message_id", "mail_draft_id", "pending_mail_draft_id", "draft_client_id", "storage_object_id", "scan_lease_token"} {
		assertOptionalNillable(t, name, attachmentFields[name])
	}
	for _, name := range []string{"draft_client_id", "storage_object_id", "engine_blob_id", "content_id", "filename", "media_type", "size_bytes", "disposition"} {
		if !attachmentFields[name].Immutable {
			t.Fatalf("MailAttachment.%s must be immutable", name)
		}
	}
	assertInt64Validators(t, "size_bytes", attachmentFields["size_bytes"], []struct{ value int64; want bool }{{-1, false}, {0, true}, {26_214_400, true}, {26_214_401, false}})
	assertExactEnumValues(t, "disposition", attachmentFields["disposition"], []string{"attachment", "inline"})
	assertExactEnumValues(t, "scan_status", attachmentFields["scan_status"], []string{"pending", "clean", "rejected", "error"})
	assertBytesValidators(t, "sha256", attachmentFields["sha256"], []struct{ value []byte; want bool }{{make([]byte, sha256.Size-1), false}, {make([]byte, sha256.Size), true}, {make([]byte, sha256.Size+1), false}})
	for _, name := range []string{"scan_lease_until", "next_scan_at", "scan_terminal_at"} {
		assertOptionalNillable(t, name, attachmentFields[name])
	}
	assertExactDefaults(t, attachmentFields, map[string]any{"scan_lease_generation": int64(0), "scan_attempts": int32(0)})
	assertOnlyDefaults(t, attachmentFields, map[string]bool{"id": true, "scan_lease_generation": true, "scan_attempts": true, "created_at": true, "updated_at": true})
	assertInt64Validators(t, "scan_lease_generation", attachmentFields["scan_lease_generation"], []struct{ value int64; want bool }{{-1, false}, {0, true}})
	assertInt32Validators(t, "scan_attempts", attachmentFields["scan_attempts"], []struct{ value int32; want bool }{{-1, false}, {0, true}, {10, true}, {11, false}})
	assertTimestampsWithoutMixin(t, attachmentFields)
	if _, found := attachmentFields["project_id"]; found {
		t.Fatal("MailAttachment must not persist project_id")
	}

	draft := entschema.MailDraft{}
	assertExactOrderedFieldNames(t, "MailDraft fields", draft.Fields(), []string{
		"id", "mail_account_id", "from_identity_id", "draft_client_id", "engine_draft_key", "engine_message_id", "engine_version", "revision", "status", "create_idempotency_key_hash", "create_request_hash", "reply_to_message_id", "forward_message_id", "pending_from_identity_id", "pending_reply_to_message_id", "pending_forward_message_id", "pending_save_request_hash", "last_mutation_id", "confirmed_mutation_id", "pending_mutation_kind", "pending_create_attachment_phase", "pending_create_upload_blob_id", "pending_create_upload_size_bytes", "pending_create_upload_attempt", "pending_create_upload_retry_at", "pending_create_upload_exhausted_at", "pending_create_mime_kind", "pending_create_mime_phase", "pending_create_mime_blob_id", "pending_create_mime_size_bytes", "pending_create_mime_binding_digest", "pending_create_mime_attempt", "pending_create_mime_retry_at", "pending_create_mime_exhausted_at", "abandoned_at", "abandoned_mutation_id", "created_at", "updated_at",
	})
	draftFields := fieldDescriptors(draft.Fields())
	for _, name := range []string{"id", "mail_account_id", "from_identity_id", "draft_client_id", "reply_to_message_id", "forward_message_id", "pending_from_identity_id", "pending_reply_to_message_id", "pending_forward_message_id", "last_mutation_id", "confirmed_mutation_id", "abandoned_mutation_id"} {
		assertFieldType(t, name, draftFields[name], field.TypeUUID)
	}
	for _, name := range []string{"mail_account_id", "draft_client_id", "engine_draft_key", "create_idempotency_key_hash", "create_request_hash"} {
		if !draftFields[name].Immutable {
			t.Fatalf("MailDraft.%s must be immutable", name)
		}
	}
	for _, name := range []string{"engine_message_id", "engine_version", "reply_to_message_id", "forward_message_id", "pending_from_identity_id", "pending_reply_to_message_id", "pending_forward_message_id", "pending_save_request_hash", "last_mutation_id", "confirmed_mutation_id", "pending_mutation_kind", "pending_create_attachment_phase", "pending_create_upload_blob_id", "pending_create_upload_size_bytes", "pending_create_upload_retry_at", "pending_create_upload_exhausted_at", "pending_create_mime_kind", "pending_create_mime_phase", "pending_create_mime_blob_id", "pending_create_mime_size_bytes", "pending_create_mime_binding_digest", "pending_create_mime_retry_at", "pending_create_mime_exhausted_at", "abandoned_at", "abandoned_mutation_id"} {
		assertOptionalNillable(t, name, draftFields[name])
	}
	assertExactDefaults(t, draftFields, map[string]any{"revision": int64(1), "pending_create_upload_attempt": 0, "pending_create_mime_attempt": 0})
	assertOnlyDefaults(t, draftFields, map[string]bool{"id": true, "revision": true, "pending_create_upload_attempt": true, "pending_create_mime_attempt": true, "created_at": true, "updated_at": true})
	assertInt64Validators(t, "revision", draftFields["revision"], []struct{ value int64; want bool }{{0, false}, {1, true}})
	assertExactEnumValues(t, "status", draftFields["status"], []string{"active", "saving", "deleting", "unknown", "submitting", "submitted", "deleted", "abandoned"})
	assertExactEnumValues(t, "pending_mutation_kind", draftFields["pending_mutation_kind"], []string{"create", "save", "delete"})
	assertExactEnumValues(t, "pending_create_attachment_phase", draftFields["pending_create_attachment_phase"], []string{"upload", "import"})
	assertExactEnumValues(t, "pending_create_mime_kind", draftFields["pending_create_mime_kind"], []string{"attachments", "calendar"})
	assertExactEnumValues(t, "pending_create_mime_phase", draftFields["pending_create_mime_phase"], []string{"upload", "import"})
	for _, name := range []string{"create_idempotency_key_hash", "create_request_hash", "pending_save_request_hash", "pending_create_mime_binding_digest"} {
		assertBytesValidators(t, name, draftFields[name], []struct{ value []byte; want bool }{{make([]byte, sha256.Size-1), false}, {make([]byte, sha256.Size), true}, {make([]byte, sha256.Size+1), false}})
	}
	for _, name := range []string{"pending_create_upload_size_bytes", "pending_create_mime_size_bytes"} {
		assertInt64Validators(t, name, draftFields[name], []struct{ value int64; want bool }{{0, false}, {1, true}, {52_428_800, true}, {52_428_801, false}})
	}
	for _, name := range []string{"pending_create_upload_attempt", "pending_create_mime_attempt"} {
		assertDeliveryIntValidators(t, name, draftFields[name], []struct{ value int; want bool }{{-1, false}, {0, true}, {3, true}, {4, false}})
	}
	assertTimestampsWithoutMixin(t, draftFields)

	submission := entschema.MailSubmission{}
	assertExactOrderedFieldNames(t, "MailSubmission fields", submission.Fields(), []string{
		"id", "mail_account_id", "mail_draft_id", "idempotency_key_hash", "request_hash", "engine_submission_key", "engine_binding_version", "engine_account_id", "engine_draft_key", "engine_message_id", "engine_expected_version", "engine_max_message_bytes", "engine_binding_hash", "status", "command_lease_token", "command_fence", "command_lease_expires_at", "dispatch_authorized_at", "dispatch_authorization_hash", "provider_submission_id", "provider_message_id", "result_mail_message_id", "last_error_code", "next_attempt_at", "attempt_count", "reconcile_attempts", "next_reconcile_at", "created_at", "updated_at",
	})
	submissionFields := fieldDescriptors(submission.Fields())
	for _, name := range []string{"id", "mail_account_id", "mail_draft_id", "result_mail_message_id"} {
		assertFieldType(t, name, submissionFields[name], field.TypeUUID)
	}
	for _, name := range []string{"mail_account_id", "idempotency_key_hash", "request_hash", "engine_submission_key", "engine_binding_version", "engine_account_id", "engine_draft_key", "engine_message_id", "engine_expected_version", "engine_max_message_bytes", "engine_binding_hash"} {
		if !submissionFields[name].Immutable {
			t.Fatalf("MailSubmission.%s must be immutable", name)
		}
	}
	for _, name := range []string{"mail_draft_id", "command_lease_token", "command_lease_expires_at", "dispatch_authorized_at", "dispatch_authorization_hash", "provider_submission_id", "provider_message_id", "result_mail_message_id", "last_error_code", "next_attempt_at", "next_reconcile_at"} {
		assertOptionalNillable(t, name, submissionFields[name])
	}
	for _, name := range []string{"idempotency_key_hash", "request_hash", "engine_binding_hash"} {
		assertBytesValidators(t, name, submissionFields[name], []struct{ value []byte; want bool }{{make([]byte, sha256.Size-1), false}, {make([]byte, sha256.Size), true}, {make([]byte, sha256.Size+1), false}})
	}
	assertExactDefaults(t, submissionFields, map[string]any{"engine_binding_version": 1, "command_fence": int64(0), "attempt_count": int32(0), "reconcile_attempts": int32(0)})
	assertOnlyDefaults(t, submissionFields, map[string]bool{"id": true, "engine_binding_version": true, "command_fence": true, "attempt_count": true, "reconcile_attempts": true, "created_at": true, "updated_at": true})
	assertInt64Validators(t, "engine_max_message_bytes", submissionFields["engine_max_message_bytes"], []struct{ value int64; want bool }{{0, false}, {1, true}, {52_428_800, true}, {52_428_801, false}})
	assertInt32Validators(t, "attempt_count", submissionFields["attempt_count"], []struct{ value int32; want bool }{{-1, false}, {0, true}, {7, true}, {8, true}, {9, false}})
	assertInt32Validators(t, "reconcile_attempts", submissionFields["reconcile_attempts"], []struct{ value int32; want bool }{{-1, false}, {0, true}, {9, true}, {10, true}, {11, false}})
	assertExactEnumValues(t, "status", submissionFields["status"], []string{"queued", "authorizing", "submitting", "unknown", "accepted", "sent", "failed"})
	assertExactEnumValues(t, "last_error_code", submissionFields["last_error_code"], []string{"MAIL_SUBMISSION_OUTCOME_UNKNOWN", "MAIL_SUBMISSION_RECONCILIATION_REQUIRED", "MAIL_SUBMISSION_RECONCILIATION_EXHAUSTED", "MAILBOX_UNAVAILABLE", "MAIL_SUBMISSION_REJECTED", "MAIL_SUBMISSION_POLICY", "MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN", "MAIL_SUBMISSION_IDENTITY_UNAVAILABLE", "MAIL_SUBMISSION_RECIPIENT_LIMIT", "MAIL_SUBMISSION_SIZE_LIMIT", "MAIL_SUBMISSION_RETRY_EXHAUSTED"})
	assertTimestampsWithoutMixin(t, submissionFields)
	for _, forbidden := range []string{"idempotency_key", "raw_idempotency_key", "body", "mail_message_id"} {
		if _, found := submissionFields[forbidden]; found {
			t.Fatalf("MailSubmission must not persist forbidden %q", forbidden)
		}
	}
}

func TestMailDeliverySchemas_RejectWrongPerFieldMetadata(t *testing.T) {
	for _, subject := range []struct {
		name     string
		fields   []ent.Field
		expected []mailDeliveryFieldMetadata
	}{
		{
			name:   "MailObject",
			fields: (entschema.MailObject{}).Fields(),
			expected: []mailDeliveryFieldMetadata{
				{"id", field.TypeUUID, false, false, true, deliveryUUIDDefault, false},
				{"deletion_id", field.TypeUUID, false, false, true, deliveryUUIDDefault, false},
				{"organization_id", field.TypeUUID, false, false, true, deliveryNoDefault, false},
				{"mail_account_id", field.TypeUUID, false, false, true, deliveryNoDefault, false},
				{"uploader_user_id", field.TypeUUID, false, false, true, deliveryNoDefault, false},
				{"draft_client_id", field.TypeUUID, false, false, true, deliveryNoDefault, false},
				{"bucket", field.TypeString, false, false, true, deliveryNoDefault, false},
				{"object_key", field.TypeString, false, false, true, deliveryNoDefault, false},
				{"expected_size_bytes", field.TypeInt64, false, false, true, deliveryNoDefault, false},
				{"actual_size_bytes", field.TypeInt64, true, true, false, deliveryNoDefault, false},
				{"filename", field.TypeString, false, false, true, deliveryNoDefault, false},
				{"media_type", field.TypeString, false, false, true, deliveryNoDefault, false},
				{"sha256", field.TypeBytes, true, true, false, deliveryNoDefault, false},
				{"upload_state", field.TypeEnum, false, false, false, "pending", false},
				{"expires_at", field.TypeTime, false, false, true, deliveryNoDefault, false},
				{"upload_drain_until", field.TypeTime, false, false, true, deliveryNoDefault, false},
				{"delete_lease_token", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"delete_lease_generation", field.TypeInt64, false, false, false, int64(0), false},
				{"delete_lease_until", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"delete_attempts", field.TypeInt32, false, false, false, int32(0), false},
				{"next_delete_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"delete_terminal_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"created_at", field.TypeTime, false, false, true, deliveryNowDefault, false},
				{"updated_at", field.TypeTime, false, false, false, deliveryNowDefault, true},
			},
		},
		{
			name:   "MailAttachment",
			fields: (entschema.MailAttachment{}).Fields(),
			expected: []mailDeliveryFieldMetadata{
				{"id", field.TypeUUID, false, false, true, deliveryUUIDDefault, false},
				{"mail_message_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"mail_draft_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"pending_mail_draft_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"draft_client_id", field.TypeUUID, true, true, true, deliveryNoDefault, false},
				{"storage_object_id", field.TypeUUID, true, true, true, deliveryNoDefault, false},
				{"engine_blob_id", field.TypeString, true, true, true, deliveryNoDefault, false},
				{"content_id", field.TypeString, true, true, true, deliveryNoDefault, false},
				{"filename", field.TypeString, false, false, true, deliveryNoDefault, false},
				{"media_type", field.TypeString, false, false, true, deliveryNoDefault, false},
				{"size_bytes", field.TypeInt64, false, false, true, deliveryNoDefault, false},
				{"disposition", field.TypeEnum, false, false, true, deliveryNoDefault, false},
				{"scan_status", field.TypeEnum, false, false, false, deliveryNoDefault, false},
				{"sha256", field.TypeBytes, true, true, false, deliveryNoDefault, false},
				{"scan_lease_token", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"scan_lease_generation", field.TypeInt64, false, false, false, int64(0), false},
				{"scan_lease_until", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"scan_attempts", field.TypeInt32, false, false, false, int32(0), false},
				{"next_scan_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"scan_terminal_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"created_at", field.TypeTime, false, false, true, deliveryNowDefault, false},
				{"updated_at", field.TypeTime, false, false, false, deliveryNowDefault, true},
			},
		},
		{
			name:   "MailDraft",
			fields: (entschema.MailDraft{}).Fields(),
			expected: []mailDeliveryFieldMetadata{
				{"id", field.TypeUUID, false, false, true, deliveryUUIDDefault, false},
				{"mail_account_id", field.TypeUUID, false, false, true, deliveryNoDefault, false},
				{"from_identity_id", field.TypeUUID, false, false, false, deliveryNoDefault, false},
				{"draft_client_id", field.TypeUUID, false, false, true, deliveryNoDefault, false},
				{"engine_draft_key", field.TypeString, false, false, true, deliveryNoDefault, false},
				{"engine_message_id", field.TypeString, true, true, false, deliveryNoDefault, false},
				{"engine_version", field.TypeString, true, true, false, deliveryNoDefault, false},
				{"revision", field.TypeInt64, false, false, false, int64(1), false},
				{"status", field.TypeEnum, false, false, false, deliveryNoDefault, false},
				{"create_idempotency_key_hash", field.TypeBytes, false, false, true, deliveryNoDefault, false},
				{"create_request_hash", field.TypeBytes, false, false, true, deliveryNoDefault, false},
				{"reply_to_message_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"forward_message_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"pending_from_identity_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"pending_reply_to_message_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"pending_forward_message_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"pending_save_request_hash", field.TypeBytes, true, true, false, deliveryNoDefault, false},
				{"last_mutation_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"confirmed_mutation_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"pending_mutation_kind", field.TypeEnum, true, true, false, deliveryNoDefault, false},
				{"pending_create_attachment_phase", field.TypeEnum, true, true, false, deliveryNoDefault, false},
				{"pending_create_upload_blob_id", field.TypeString, true, true, false, deliveryNoDefault, false},
				{"pending_create_upload_size_bytes", field.TypeInt64, true, true, false, deliveryNoDefault, false},
				{"pending_create_upload_attempt", field.TypeInt, false, false, false, 0, false},
				{"pending_create_upload_retry_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"pending_create_upload_exhausted_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"pending_create_mime_kind", field.TypeEnum, true, true, false, deliveryNoDefault, false},
				{"pending_create_mime_phase", field.TypeEnum, true, true, false, deliveryNoDefault, false},
				{"pending_create_mime_blob_id", field.TypeString, true, true, false, deliveryNoDefault, false},
				{"pending_create_mime_size_bytes", field.TypeInt64, true, true, false, deliveryNoDefault, false},
				{"pending_create_mime_binding_digest", field.TypeBytes, true, true, false, deliveryNoDefault, false},
				{"pending_create_mime_attempt", field.TypeInt, false, false, false, 0, false},
				{"pending_create_mime_retry_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"pending_create_mime_exhausted_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"abandoned_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"abandoned_mutation_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"created_at", field.TypeTime, false, false, true, deliveryNowDefault, false},
				{"updated_at", field.TypeTime, false, false, false, deliveryNowDefault, true},
			},
		},
		{
			name:   "MailSubmission",
			fields: (entschema.MailSubmission{}).Fields(),
			expected: []mailDeliveryFieldMetadata{
				{"id", field.TypeUUID, false, false, true, deliveryUUIDDefault, false},
				{"mail_account_id", field.TypeUUID, false, false, true, deliveryNoDefault, false},
				{"mail_draft_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"idempotency_key_hash", field.TypeBytes, false, false, true, deliveryNoDefault, false},
				{"request_hash", field.TypeBytes, false, false, true, deliveryNoDefault, false},
				{"engine_submission_key", field.TypeString, false, false, true, deliveryNoDefault, false},
				{"engine_binding_version", field.TypeInt, false, false, true, 1, false},
				{"engine_account_id", field.TypeString, false, false, true, deliveryNoDefault, false},
				{"engine_draft_key", field.TypeString, false, false, true, deliveryNoDefault, false},
				{"engine_message_id", field.TypeString, false, false, true, deliveryNoDefault, false},
				{"engine_expected_version", field.TypeString, false, false, true, deliveryNoDefault, false},
				{"engine_max_message_bytes", field.TypeInt64, false, false, true, deliveryNoDefault, false},
				{"engine_binding_hash", field.TypeBytes, false, false, true, deliveryNoDefault, false},
				{"status", field.TypeEnum, false, false, false, deliveryNoDefault, false},
				{"command_lease_token", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"command_fence", field.TypeInt64, false, false, false, int64(0), false},
				{"command_lease_expires_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"dispatch_authorized_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"dispatch_authorization_hash", field.TypeBytes, true, true, false, deliveryNoDefault, false},
				{"provider_submission_id", field.TypeString, true, true, false, deliveryNoDefault, false},
				{"provider_message_id", field.TypeString, true, true, false, deliveryNoDefault, false},
				{"result_mail_message_id", field.TypeUUID, true, true, false, deliveryNoDefault, false},
				{"last_error_code", field.TypeEnum, true, true, false, deliveryNoDefault, false},
				{"next_attempt_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"attempt_count", field.TypeInt32, false, false, false, int32(0), false},
				{"reconcile_attempts", field.TypeInt32, false, false, false, int32(0), false},
				{"next_reconcile_at", field.TypeTime, true, true, false, deliveryNoDefault, false},
				{"created_at", field.TypeTime, false, false, true, deliveryNowDefault, false},
				{"updated_at", field.TypeTime, false, false, false, deliveryNowDefault, true},
			},
		},
	} {
		t.Run(subject.name, func(t *testing.T) {
			assertMailDeliveryFieldMetadata(t, subject.name, subject.fields, subject.expected)
		})
	}
}

type mailDeliveryDefaultKind uint8

const (
	deliveryNoDefault mailDeliveryDefaultKind = iota
	deliveryUUIDDefault
	deliveryNowDefault
)

type mailDeliveryFieldMetadata struct {
	name          string
	fieldType     field.Type
	optional      bool
	nillable      bool
	immutable     bool
	defaultValue  any
	updateDefault bool
}

func assertMailDeliveryFieldMetadata(t *testing.T, schemaName string, fields []ent.Field, want []mailDeliveryFieldMetadata) {
	t.Helper()
	if len(fields) != len(want) {
		t.Fatalf("%s field count = %d, want %d", schemaName, len(fields), len(want))
	}
	for position, wanted := range want {
		descriptor := fields[position].Descriptor()
		if descriptor.Name != wanted.name || descriptor.Info == nil || descriptor.Info.Type != wanted.fieldType || descriptor.Optional != wanted.optional || descriptor.Nillable != wanted.nillable || descriptor.Immutable != wanted.immutable {
			t.Fatalf("%s.%s metadata = %#v, want type:%v optional:%t nillable:%t immutable:%t", schemaName, wanted.name, descriptor, wanted.fieldType, wanted.optional, wanted.nillable, wanted.immutable)
		}
		assertMailDeliveryDefault(t, schemaName+"."+wanted.name, descriptor.Default, wanted.defaultValue)
		if got := descriptor.UpdateDefault != nil; got != wanted.updateDefault {
			t.Fatalf("%s.%s update default present = %t, want %t", schemaName, wanted.name, got, wanted.updateDefault)
		}
		if wanted.updateDefault && reflect.TypeOf(descriptor.UpdateDefault) != reflect.TypeOf(func() time.Time { return time.Time{} }) {
			t.Fatalf("%s.%s update default type = %T, want func() time.Time", schemaName, wanted.name, descriptor.UpdateDefault)
		}
	}
}

func assertMailDeliveryDefault(t *testing.T, name string, got, want any) {
	t.Helper()
	switch wanted := want.(type) {
	case mailDeliveryDefaultKind:
		switch wanted {
		case deliveryNoDefault:
			if got != nil {
				t.Fatalf("%s default = %#v, want none", name, got)
			}
		case deliveryUUIDDefault:
			if reflect.TypeOf(got) != reflect.TypeOf(uuid.New) {
				t.Fatalf("%s default type = %T, want func() uuid.UUID", name, got)
			}
		case deliveryNowDefault:
			if reflect.TypeOf(got) != reflect.TypeOf(func() time.Time { return time.Time{} }) {
				t.Fatalf("%s default type = %T, want func() time.Time", name, got)
			}
		default:
			t.Fatalf("%s has unknown default expectation %d", name, wanted)
		}
	default:
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s default = %#v, want %#v", name, got, want)
		}
	}
}

func assertDeliveryIntValidators(t *testing.T, name string, descriptor *field.Descriptor, cases []struct {
	value int
	want  bool
}) {
	t.Helper()
	validators := make([]func(int) error, 0, len(descriptor.Validators))
	for _, rawValidator := range descriptor.Validators {
		if validator, ok := rawValidator.(func(int) error); ok {
			validators = append(validators, validator)
		}
	}
	if len(validators) == 0 {
		t.Fatalf("%s lacks an int validator", name)
	}
	for _, testCase := range cases {
		accepted := true
		for _, validator := range validators {
			if validator(testCase.value) != nil {
				accepted = false
				break
			}
		}
		if accepted != testCase.want {
			t.Fatalf("%s accepts %d = %t, want %t", name, testCase.value, accepted, testCase.want)
		}
	}
}

func TestMailDeliverySchemas_RejectWrongChecksIndexesAndRelations(t *testing.T) {
	assertExactChecks(t, (entschema.MailObject{}).Annotations(), map[string]string{
		"mail_objects_text":                  "COALESCE(char_length(bucket) BETWEEN 1 AND 63 AND octet_length(bucket) BETWEEN 1 AND 63 AND char_length(object_key) BETWEEN 1 AND 512 AND octet_length(object_key) BETWEEN 1 AND 512 AND char_length(filename) BETWEEN 1 AND 255 AND octet_length(filename) BETWEEN 1 AND 1020 AND char_length(media_type) BETWEEN 1 AND 255 AND octet_length(media_type) BETWEEN 1 AND 1020, FALSE)",
		"mail_objects_sizes":                 "COALESCE(expected_size_bytes BETWEEN 1 AND 26214400 AND (actual_size_bytes IS NULL OR actual_size_bytes BETWEEN 1 AND 26214400), FALSE)",
		"mail_objects_actual_sha_pair":       "COALESCE((actual_size_bytes IS NULL) = (sha256 IS NULL) AND (sha256 IS NULL OR octet_length(sha256) = 32), FALSE)",
		"mail_objects_upload_enum":           "COALESCE(upload_state IN ('pending','confirmed','orphaned','deleted'), FALSE)",
		"mail_objects_delete_lease_pair":     "COALESCE((delete_lease_token IS NULL) = (delete_lease_until IS NULL), FALSE)",
		"mail_objects_delete_retry_bounds":   "COALESCE(delete_lease_generation >= 0 AND delete_attempts BETWEEN 0 AND 10, FALSE)",
		"mail_objects_upload_state":          "COALESCE((upload_state = 'pending' AND actual_size_bytes IS NULL AND sha256 IS NULL AND next_delete_at IS NULL AND delete_lease_token IS NULL AND delete_lease_until IS NULL AND delete_lease_generation = 0 AND delete_terminal_at IS NULL AND delete_attempts = 0) OR (upload_state = 'confirmed' AND actual_size_bytes = expected_size_bytes AND sha256 IS NOT NULL AND next_delete_at IS NULL AND delete_lease_token IS NULL AND delete_lease_until IS NULL AND delete_lease_generation = 0 AND delete_terminal_at IS NULL AND delete_attempts = 0) OR (upload_state IN ('orphaned','deleted') AND ((delete_terminal_at IS NOT NULL AND next_delete_at IS NULL AND delete_lease_token IS NULL AND delete_lease_until IS NULL AND delete_attempts = 10) OR (delete_terminal_at IS NULL AND next_delete_at IS NOT NULL AND ((delete_attempts BETWEEN 0 AND 9 AND (delete_lease_token IS NULL OR delete_lease_generation >= 1)) OR (delete_attempts = 10 AND delete_lease_token IS NOT NULL AND delete_lease_generation >= 1))))), FALSE)",
		"mail_objects_upload_drain_boundary": "COALESCE(upload_drain_until = expires_at + INTERVAL '11 minutes' AND (next_delete_at IS NULL OR next_delete_at >= upload_drain_until), FALSE)",
	})
	assertExactChecks(t, (entschema.MailAttachment{}).Annotations(), map[string]string{
		"mail_attachments_text":                  "COALESCE((engine_blob_id IS NULL OR octet_length(engine_blob_id) BETWEEN 1 AND 1024) AND (content_id IS NULL OR (char_length(content_id) BETWEEN 1 AND 255 AND octet_length(content_id) BETWEEN 1 AND 1020)) AND char_length(filename) BETWEEN 1 AND 255 AND octet_length(filename) BETWEEN 1 AND 1020 AND char_length(media_type) BETWEEN 1 AND 255 AND octet_length(media_type) BETWEEN 1 AND 1020 AND size_bytes BETWEEN 0 AND 26214400 AND (sha256 IS NULL OR octet_length(sha256) = 32), FALSE)",
		"mail_attachments_backing_xor":           "COALESCE((storage_object_id IS NOT NULL) <> (engine_blob_id IS NOT NULL), FALSE)",
		"mail_attachments_owner_exclusive":       "COALESCE(NOT (mail_message_id IS NOT NULL AND (mail_draft_id IS NOT NULL OR pending_mail_draft_id IS NOT NULL)) AND (mail_draft_id IS NULL OR pending_mail_draft_id IS NULL OR mail_draft_id = pending_mail_draft_id), FALSE)",
		"mail_attachments_storage_shape":         "COALESCE(storage_object_id IS NULL OR (draft_client_id IS NOT NULL AND size_bytes >= 1), FALSE)",
		"mail_attachments_pending_storage_shape": "COALESCE(pending_mail_draft_id IS NULL OR (storage_object_id IS NOT NULL AND draft_client_id IS NOT NULL AND size_bytes >= 1), FALSE)",
		"mail_attachments_engine_shape":          "COALESCE(engine_blob_id IS NULL OR (mail_message_id IS NOT NULL AND mail_draft_id IS NULL AND pending_mail_draft_id IS NULL AND draft_client_id IS NULL), FALSE)",
		"mail_attachments_inline_content":        "COALESCE(disposition <> 'inline' OR content_id IS NOT NULL, FALSE)",
		"mail_attachments_disposition":           "COALESCE(disposition IN ('attachment','inline'), FALSE)",
		"mail_attachments_scan_enum":             "COALESCE(scan_status IN ('pending','clean','rejected','error'), FALSE)",
		"mail_attachments_hash_scan":             "COALESCE(scan_status NOT IN ('clean','rejected') OR sha256 IS NOT NULL, FALSE)",
		"mail_attachments_scan_lease_pair":       "COALESCE((scan_lease_token IS NULL) = (scan_lease_until IS NULL), FALSE)",
		"mail_attachments_scan_retry_bounds":     "COALESCE(scan_lease_generation >= 0 AND scan_attempts BETWEEN 0 AND 10, FALSE)",
		"mail_attachments_scan_schedule":         "COALESCE((storage_object_id IS NULL AND scan_lease_token IS NULL AND scan_lease_generation = 0 AND scan_lease_until IS NULL AND scan_attempts = 0 AND next_scan_at IS NULL AND scan_terminal_at IS NULL) OR (storage_object_id IS NOT NULL AND ((scan_status = 'pending' AND sha256 IS NULL AND scan_lease_token IS NULL AND scan_lease_generation = 0 AND scan_lease_until IS NULL AND scan_attempts = 0 AND next_scan_at IS NULL AND scan_terminal_at IS NULL) OR (scan_status = 'pending' AND sha256 IS NOT NULL AND scan_terminal_at IS NULL AND next_scan_at IS NOT NULL AND ((scan_attempts BETWEEN 0 AND 9 AND (scan_lease_token IS NULL OR scan_lease_generation >= 1)) OR (scan_attempts = 10 AND scan_lease_token IS NOT NULL AND scan_lease_generation >= 1))) OR (scan_status IN ('clean','rejected') AND sha256 IS NOT NULL AND scan_lease_token IS NULL AND scan_lease_until IS NULL AND next_scan_at IS NULL AND scan_terminal_at IS NOT NULL AND scan_attempts BETWEEN 1 AND 10 AND scan_lease_generation >= 1) OR (scan_status = 'error' AND sha256 IS NOT NULL AND scan_lease_token IS NULL AND scan_lease_until IS NULL AND next_scan_at IS NULL AND scan_terminal_at IS NOT NULL AND scan_attempts = 10 AND scan_lease_generation >= 1))), FALSE)",
	})
	assertExactChecks(t, (entschema.MailDraft{}).Annotations(), map[string]string{
		"mail_drafts_key":                "COALESCE(octet_length(engine_draft_key) BETWEEN 1 AND 1024, FALSE)",
		"mail_drafts_engine_pair":        "COALESCE((engine_message_id IS NULL AND engine_version IS NULL) OR (octet_length(engine_message_id) BETWEEN 1 AND 1024 AND octet_length(engine_version) BETWEEN 1 AND 1024), FALSE)",
		"mail_drafts_revision":           "COALESCE(revision >= 1, FALSE)",
		"mail_drafts_hashes":             "COALESCE(octet_length(create_idempotency_key_hash) = 32 AND octet_length(create_request_hash) = 32, FALSE)",
		"mail_drafts_status_enum":        "COALESCE(status IN ('active','saving','deleting','unknown','submitting','submitted','deleted','abandoned'), FALSE)",
		"mail_drafts_state":              "COALESCE((status IN ('active','submitting','submitted') AND engine_message_id IS NOT NULL AND last_mutation_id IS NULL AND pending_mutation_kind IS NULL) OR (status = 'saving' AND last_mutation_id IS NOT NULL AND pending_mutation_kind IN ('create','save') AND ((pending_mutation_kind = 'create' AND engine_message_id IS NULL) OR (pending_mutation_kind = 'save' AND engine_message_id IS NOT NULL))) OR (status = 'deleting' AND last_mutation_id IS NOT NULL AND pending_mutation_kind = 'delete' AND engine_message_id IS NOT NULL) OR (status = 'unknown' AND last_mutation_id IS NOT NULL AND ((pending_mutation_kind = 'create' AND engine_message_id IS NULL) OR (pending_mutation_kind IN ('save','delete') AND engine_message_id IS NOT NULL))) OR (status = 'deleted' AND last_mutation_id IS NULL AND pending_mutation_kind IS NULL) OR (status = 'abandoned' AND engine_message_id IS NULL AND engine_version IS NULL AND last_mutation_id IS NULL AND pending_mutation_kind IS NULL), FALSE)",
		"mail_drafts_pending_save_state": "COALESCE((pending_mutation_kind = 'save' AND pending_from_identity_id IS NOT NULL AND octet_length(pending_save_request_hash) = 32) OR ((pending_mutation_kind IS NULL OR pending_mutation_kind <> 'save') AND pending_from_identity_id IS NULL AND pending_reply_to_message_id IS NULL AND pending_forward_message_id IS NULL AND pending_save_request_hash IS NULL), FALSE)",
		"mail_drafts_pending_create_attachment_phase": "COALESCE(pending_create_attachment_phase IS NULL OR (pending_mutation_kind = 'create' AND status IN ('saving','unknown') AND pending_create_attachment_phase IN ('upload','import')), FALSE)",
		"mail_drafts_pending_create_upload_handoff":   "COALESCE((pending_create_attachment_phase = 'import' AND octet_length(pending_create_upload_blob_id) BETWEEN 1 AND 1024 AND pending_create_upload_size_bytes BETWEEN 1 AND 52428800) OR (pending_create_attachment_phase IS DISTINCT FROM 'import' AND pending_create_upload_blob_id IS NULL AND pending_create_upload_size_bytes IS NULL), FALSE)",
		"mail_drafts_pending_create_upload_recovery": "COALESCE((pending_create_attachment_phase IS NULL AND pending_create_upload_attempt = 0 AND pending_create_upload_retry_at IS NULL AND pending_create_upload_exhausted_at IS NULL) OR (pending_create_attachment_phase = 'import' AND pending_mutation_kind = 'create' AND status IN ('saving','unknown') AND pending_create_upload_attempt BETWEEN 1 AND 3 AND pending_create_upload_retry_at IS NULL AND pending_create_upload_exhausted_at IS NULL) OR (pending_create_attachment_phase = 'upload' AND pending_mutation_kind = 'create' AND status = 'saving' AND pending_create_upload_attempt BETWEEN 1 AND 3 AND pending_create_upload_retry_at IS NULL AND pending_create_upload_exhausted_at IS NULL) OR (pending_create_attachment_phase = 'upload' AND pending_mutation_kind = 'create' AND status = 'unknown' AND ((pending_create_upload_attempt IN (1,2) AND pending_create_upload_retry_at IS NOT NULL AND pending_create_upload_exhausted_at IS NULL) OR (pending_create_upload_attempt = 3 AND pending_create_upload_retry_at IS NULL AND pending_create_upload_exhausted_at IS NOT NULL))), FALSE)",
		"mail_drafts_pending_create_mime_idle": "COALESCE((pending_create_mime_kind IS NULL AND pending_create_mime_phase IS NULL AND pending_create_mime_blob_id IS NULL AND pending_create_mime_size_bytes IS NULL AND pending_create_mime_binding_digest IS NULL AND pending_create_mime_attempt = 0 AND pending_create_mime_retry_at IS NULL AND pending_create_mime_exhausted_at IS NULL) OR (pending_create_mime_kind IS NOT NULL AND pending_create_mime_phase IS NOT NULL), FALSE)",
		"mail_drafts_pending_create_mime_active": "COALESCE((pending_create_mime_kind IS NULL AND pending_create_mime_phase IS NULL) OR (pending_mutation_kind = 'create' AND status IN ('saving','unknown') AND pending_create_mime_kind IN ('attachments','calendar') AND pending_create_mime_phase IN ('upload','import') AND octet_length(pending_create_mime_binding_digest) = 32), FALSE)",
		"mail_drafts_pending_create_mime_handoff": "COALESCE((pending_create_mime_phase = 'import' AND octet_length(pending_create_mime_blob_id) BETWEEN 1 AND 1024 AND pending_create_mime_size_bytes BETWEEN 1 AND 52428800) OR (pending_create_mime_phase IS DISTINCT FROM 'import' AND pending_create_mime_blob_id IS NULL AND pending_create_mime_size_bytes IS NULL), FALSE)",
		"mail_drafts_pending_create_mime_recovery": "COALESCE((pending_create_mime_phase IS NULL AND pending_create_mime_attempt = 0 AND pending_create_mime_retry_at IS NULL AND pending_create_mime_exhausted_at IS NULL) OR (pending_create_mime_phase = 'import' AND pending_mutation_kind = 'create' AND status IN ('saving','unknown') AND pending_create_mime_attempt BETWEEN 1 AND 3 AND pending_create_mime_retry_at IS NULL AND pending_create_mime_exhausted_at IS NULL) OR (pending_create_mime_phase = 'upload' AND pending_mutation_kind = 'create' AND status = 'saving' AND pending_create_mime_attempt BETWEEN 1 AND 3 AND pending_create_mime_retry_at IS NULL AND pending_create_mime_exhausted_at IS NULL) OR (pending_create_mime_phase = 'upload' AND pending_mutation_kind = 'create' AND status = 'unknown' AND ((pending_create_mime_attempt IN (1,2) AND pending_create_mime_retry_at IS NOT NULL AND pending_create_mime_exhausted_at IS NULL) OR (pending_create_mime_attempt = 3 AND pending_create_mime_retry_at IS NULL AND pending_create_mime_exhausted_at IS NOT NULL))), FALSE)",
		"mail_drafts_pending_create_mime_legacy_bridge": "COALESCE(pending_create_mime_kind IS NULL OR (pending_create_mime_kind = 'calendar' AND pending_create_attachment_phase IS NULL AND pending_create_upload_blob_id IS NULL AND pending_create_upload_size_bytes IS NULL AND pending_create_upload_attempt = 0 AND pending_create_upload_retry_at IS NULL AND pending_create_upload_exhausted_at IS NULL) OR (pending_create_mime_kind = 'attachments' AND pending_create_attachment_phase IS NOT NULL AND pending_create_attachment_phase = pending_create_mime_phase AND pending_create_upload_blob_id IS NOT DISTINCT FROM pending_create_mime_blob_id AND pending_create_upload_size_bytes IS NOT DISTINCT FROM pending_create_mime_size_bytes AND pending_create_upload_attempt = pending_create_mime_attempt AND pending_create_upload_retry_at IS NOT DISTINCT FROM pending_create_mime_retry_at AND pending_create_upload_exhausted_at IS NOT DISTINCT FROM pending_create_mime_exhausted_at), FALSE)",
		"mail_drafts_abandoned_state": "COALESCE((status = 'abandoned' AND abandoned_at IS NOT NULL AND abandoned_mutation_id IS NOT NULL) OR (status <> 'abandoned' AND abandoned_at IS NULL AND abandoned_mutation_id IS NULL), FALSE)",
		"mail_drafts_confirmed_mutation":  "COALESCE(confirmed_mutation_id IS NULL OR (engine_message_id IS NOT NULL AND engine_version IS NOT NULL AND (last_mutation_id IS NULL OR confirmed_mutation_id <> last_mutation_id)), FALSE)",
	})
	assertExactChecks(t, (entschema.MailSubmission{}).Annotations(), map[string]string{
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
		// RETRY_EXHAUSTED implies dispatch attempt 8; an eighth definitive failure may keep its exact code.
		"mail_submissions_state":                       "COALESCE((status = 'queued' AND mail_draft_id IS NOT NULL AND command_lease_token IS NULL AND command_lease_expires_at IS NULL AND dispatch_authorized_at IS NULL AND dispatch_authorization_hash IS NULL AND provider_submission_id IS NULL AND provider_message_id IS NULL AND result_mail_message_id IS NULL AND last_error_code IS NULL AND next_attempt_at IS NOT NULL AND attempt_count BETWEEN 0 AND 7 AND reconcile_attempts = 0 AND next_reconcile_at IS NULL) OR (status = 'authorizing' AND mail_draft_id IS NOT NULL AND command_lease_token IS NOT NULL AND command_lease_expires_at IS NOT NULL AND dispatch_authorized_at IS NULL AND dispatch_authorization_hash IS NULL AND provider_submission_id IS NULL AND provider_message_id IS NULL AND result_mail_message_id IS NULL AND last_error_code IS NULL AND next_attempt_at IS NULL AND attempt_count BETWEEN 1 AND 8 AND command_fence >= 1 AND reconcile_attempts = 0 AND next_reconcile_at IS NULL) OR (status = 'submitting' AND mail_draft_id IS NOT NULL AND command_lease_token IS NOT NULL AND command_lease_expires_at IS NOT NULL AND dispatch_authorized_at IS NOT NULL AND dispatch_authorization_hash IS NOT NULL AND provider_submission_id IS NULL AND provider_message_id IS NULL AND result_mail_message_id IS NULL AND last_error_code IS NULL AND next_attempt_at IS NULL AND attempt_count BETWEEN 1 AND 8 AND command_fence >= 1 AND reconcile_attempts = 0 AND next_reconcile_at IS NULL) OR (status = 'unknown' AND mail_draft_id IS NOT NULL AND dispatch_authorized_at IS NOT NULL AND dispatch_authorization_hash IS NOT NULL AND provider_submission_id IS NULL AND provider_message_id IS NULL AND result_mail_message_id IS NULL AND next_attempt_at IS NULL AND attempt_count BETWEEN 1 AND 8 AND command_fence >= 1 AND ((command_lease_token IS NULL AND command_lease_expires_at IS NULL AND reconcile_attempts BETWEEN 0 AND 9 AND next_reconcile_at IS NOT NULL AND last_error_code IN ('MAIL_SUBMISSION_OUTCOME_UNKNOWN','MAIL_SUBMISSION_RECONCILIATION_REQUIRED','MAILBOX_UNAVAILABLE')) OR (command_lease_token IS NOT NULL AND command_lease_expires_at IS NOT NULL AND reconcile_attempts BETWEEN 1 AND 10 AND next_reconcile_at IS NULL AND last_error_code IN ('MAIL_SUBMISSION_OUTCOME_UNKNOWN','MAIL_SUBMISSION_RECONCILIATION_REQUIRED','MAILBOX_UNAVAILABLE')) OR (command_lease_token IS NULL AND command_lease_expires_at IS NULL AND reconcile_attempts = 10 AND next_reconcile_at IS NULL AND last_error_code = 'MAIL_SUBMISSION_RECONCILIATION_EXHAUSTED'))) OR (status = 'accepted' AND mail_draft_id IS NOT NULL AND dispatch_authorized_at IS NOT NULL AND dispatch_authorization_hash IS NOT NULL AND provider_submission_id IS NOT NULL AND provider_message_id IS NOT NULL AND result_mail_message_id IS NULL AND last_error_code IS NULL AND next_attempt_at IS NULL AND attempt_count BETWEEN 1 AND 8 AND command_fence >= 1 AND reconcile_attempts BETWEEN 0 AND 10 AND next_reconcile_at IS NULL) OR (status = 'sent' AND command_lease_token IS NULL AND command_lease_expires_at IS NULL AND dispatch_authorized_at IS NOT NULL AND dispatch_authorization_hash IS NOT NULL AND provider_submission_id IS NOT NULL AND provider_message_id IS NOT NULL AND result_mail_message_id IS NOT NULL AND last_error_code IS NULL AND next_attempt_at IS NULL AND attempt_count BETWEEN 1 AND 8 AND command_fence >= 1 AND reconcile_attempts BETWEEN 0 AND 10 AND next_reconcile_at IS NULL) OR (status = 'failed' AND command_lease_token IS NULL AND command_lease_expires_at IS NULL AND dispatch_authorized_at IS NULL AND dispatch_authorization_hash IS NULL AND provider_submission_id IS NULL AND provider_message_id IS NULL AND result_mail_message_id IS NULL AND last_error_code IN ('MAIL_SUBMISSION_REJECTED','MAIL_SUBMISSION_POLICY','MAIL_SUBMISSION_ATTACHMENT_NOT_CLEAN','MAIL_SUBMISSION_IDENTITY_UNAVAILABLE','MAIL_SUBMISSION_RECIPIENT_LIMIT','MAIL_SUBMISSION_SIZE_LIMIT','MAIL_SUBMISSION_RETRY_EXHAUSTED') AND next_attempt_at IS NULL AND attempt_count BETWEEN 1 AND 8 AND (last_error_code <> 'MAIL_SUBMISSION_RETRY_EXHAUSTED' OR attempt_count = 8) AND reconcile_attempts = 0 AND next_reconcile_at IS NULL), FALSE)",
	})

	assertExactIndexes(t, (entschema.MailObject{}).Indexes(), []wantedIndex{{fields: []string{"bucket", "object_key"}, unique: true}, {fields: []string{"next_delete_at", "id"}, where: "upload_state IN ('orphaned','deleted') AND delete_terminal_at IS NULL AND next_delete_at IS NOT NULL"}})
	assertDeliveryIndexStorageKeys(t, "MailObject", (entschema.MailObject{}).Indexes(), []string{"mail_objects_bucket_object_key_uidx", "mail_objects_delete_due_idx"})
	assertExactIndexes(t, (entschema.MailAttachment{}).Indexes(), []wantedIndex{{fields: []string{"storage_object_id"}, unique: true, where: "storage_object_id IS NOT NULL"}, {fields: []string{"mail_message_id"}}, {fields: []string{"mail_draft_id"}}, {fields: []string{"pending_mail_draft_id"}}, {fields: []string{"next_scan_at", "id"}, where: "storage_object_id IS NOT NULL AND scan_status = 'pending' AND scan_terminal_at IS NULL AND next_scan_at IS NOT NULL"}})
	assertDeliveryIndexStorageKeys(t, "MailAttachment", (entschema.MailAttachment{}).Indexes(), []string{"mail_attachments_storage_object_uidx", "mail_attachments_message_idx", "mail_attachments_draft_idx", "mail_attachments_pending_draft_idx", "mail_attachments_scan_due_idx"})
	assertExactIndexes(t, (entschema.MailDraft{}).Indexes(), []wantedIndex{{fields: []string{"mail_account_id", "engine_draft_key"}, unique: true}, {fields: []string{"mail_account_id", "draft_client_id"}, unique: true}, {fields: []string{"mail_account_id", "create_idempotency_key_hash"}, unique: true}, {fields: []string{"mail_account_id", "engine_message_id"}, unique: true, where: "engine_message_id IS NOT NULL"}, {fields: []string{"from_identity_id"}}, {fields: []string{"pending_from_identity_id"}, where: "pending_from_identity_id IS NOT NULL"}})
	assertDeliveryIndexStorageKeys(t, "MailDraft", (entschema.MailDraft{}).Indexes(), []string{"mail_drafts_account_engine_key_uidx", "mail_drafts_account_client_uidx", "mail_drafts_account_create_key_uidx", "mail_drafts_account_engine_message_uidx", "mail_drafts_from_identity_idx", "mail_drafts_pending_from_identity_idx"})
	assertExactIndexes(t, (entschema.MailSubmission{}).Indexes(), []wantedIndex{
		{fields: []string{"mail_account_id", "idempotency_key_hash"}, unique: true},
		{fields: []string{"mail_account_id", "engine_submission_key"}, unique: true},
		{fields: []string{"mail_account_id", "provider_submission_id"}, unique: true, where: "provider_submission_id IS NOT NULL"},
		{fields: []string{"mail_account_id", "provider_message_id"}, unique: true, where: "provider_message_id IS NOT NULL"},
		{fields: []string{"mail_account_id", "result_mail_message_id"}, unique: true, where: "result_mail_message_id IS NOT NULL"},
		{fields: []string{"mail_draft_id"}, unique: true, where: "status IN ('queued','authorizing','submitting','unknown','accepted')"},
		{fields: []string{"next_attempt_at", "id"}, where: "status = 'queued'"},
		{fields: []string{"command_lease_expires_at", "id"}, where: "status IN ('authorizing','submitting','unknown','accepted')"},
		{fields: []string{"next_reconcile_at", "id"}, where: "status = 'unknown' AND next_reconcile_at IS NOT NULL"},
		{fields: []string{"updated_at", "id"}, where: "status = 'accepted'"},
	})
	assertDeliveryIndexStorageKeys(t, "MailSubmission", (entschema.MailSubmission{}).Indexes(), []string{"mail_submissions_account_key_hash_uidx", "mail_submissions_account_engine_key_uidx", "mail_submissions_account_provider_submission_uidx", "mail_submissions_account_provider_message_uidx", "mail_submissions_account_result_snapshot_uidx", "mail_submissions_active_draft_uidx", "mail_submissions_queued_due_idx", "mail_submissions_stale_lease_idx", "mail_submissions_unknown_reconcile_idx", "mail_submissions_accepted_finalizer_due_idx"})

	objectEdges := edgeDescriptors((entschema.MailObject{}).Edges())
	assertExactNames(t, "MailObject edges", edgeNames(objectEdges), []string{"account", "attachment", "organization", "uploader"})
	assertRequiredInverseEdge(t, "MailObject.organization", objectEdges["organization"], "organization_id", "mail_objects", entsql.NoAction)
	assertRequiredInverseEdge(t, "MailObject.account", objectEdges["account"], "mail_account_id", "objects", entsql.NoAction)
	assertRequiredInverseEdge(t, "MailObject.uploader", objectEdges["uploader"], "uploader_user_id", "mail_objects", entsql.NoAction)
	assertDeliveryPlainUniqueEdge(t, "MailObject.attachment", objectEdges["attachment"])

	attachmentEdges := edgeDescriptors((entschema.MailAttachment{}).Edges())
	assertExactNames(t, "MailAttachment edges", edgeNames(attachmentEdges), []string{"draft", "message", "pending_draft", "storage_object"})
	assertDeliveryParentEdge(t, "MailAttachment.storage_object", attachmentEdges["storage_object"], "storage_object_id", "attachment", false, entsql.Restrict)
	assertDeliveryParentEdge(t, "MailAttachment.message", attachmentEdges["message"], "mail_message_id", "attachment_records", false, entsql.Restrict)
	assertDeliveryParentEdge(t, "MailAttachment.draft", attachmentEdges["draft"], "mail_draft_id", "attachments", false, entsql.Restrict)
	assertDeliveryParentEdge(t, "MailAttachment.pending_draft", attachmentEdges["pending_draft"], "pending_mail_draft_id", "pending_attachments", false, entsql.Restrict)

	draftEdges := edgeDescriptors((entschema.MailDraft{}).Edges())
	assertExactNames(t, "MailDraft edges", edgeNames(draftEdges), []string{"account", "attachments", "forward_message", "from_identity", "pending_attachments", "pending_forward_message", "pending_from_identity", "pending_reply_to_message", "reply_to_message", "submissions"})
	assertDeliveryParentEdge(t, "MailDraft.account", draftEdges["account"], "mail_account_id", "drafts", true, entsql.Cascade)
	assertDeliveryParentEdge(t, "MailDraft.from_identity", draftEdges["from_identity"], "from_identity_id", "drafts", true, entsql.Restrict)
	assertDeliveryParentEdge(t, "MailDraft.reply_to_message", draftEdges["reply_to_message"], "reply_to_message_id", "reply_drafts", false, entsql.SetNull)
	assertDeliveryParentEdge(t, "MailDraft.forward_message", draftEdges["forward_message"], "forward_message_id", "forward_drafts", false, entsql.SetNull)
	assertDeliveryParentEdge(t, "MailDraft.pending_from_identity", draftEdges["pending_from_identity"], "pending_from_identity_id", "pending_drafts", false, entsql.Restrict)
	assertDeliveryParentEdge(t, "MailDraft.pending_reply_to_message", draftEdges["pending_reply_to_message"], "pending_reply_to_message_id", "pending_reply_drafts", false, entsql.Restrict)
	assertDeliveryParentEdge(t, "MailDraft.pending_forward_message", draftEdges["pending_forward_message"], "pending_forward_message_id", "pending_forward_drafts", false, entsql.Restrict)
	assertPlainToManyEdge(t, "MailDraft.attachments", draftEdges["attachments"])
	assertPlainToManyEdge(t, "MailDraft.pending_attachments", draftEdges["pending_attachments"])
	assertPlainToManyEdge(t, "MailDraft.submissions", draftEdges["submissions"])

	submissionEdges := edgeDescriptors((entschema.MailSubmission{}).Edges())
	assertExactNames(t, "MailSubmission edges", edgeNames(submissionEdges), []string{"account", "draft"})
	assertDeliveryParentEdge(t, "MailSubmission.account", submissionEdges["account"], "mail_account_id", "submissions", true, entsql.Cascade)
	assertDeliveryParentEdge(t, "MailSubmission.draft", submissionEdges["draft"], "mail_draft_id", "submissions", false, entsql.SetNull)
}

func assertDeliveryIndexStorageKeys(t *testing.T, entity string, indexes []ent.Index, want []string) {
	t.Helper()
	if len(indexes) != len(want) {
		t.Fatalf("%s index count = %d, want %d", entity, len(indexes), len(want))
	}
	for position, storageKey := range want {
		if got := indexes[position].Descriptor().StorageKey; got != storageKey {
			t.Fatalf("%s index[%d] storage key = %q, want %q", entity, position, got, storageKey)
		}
	}
}

func assertDeliveryParentEdge(t *testing.T, name string, descriptor *edge.Descriptor, fieldName, ref string, required bool, onDelete entsql.ReferenceOption) {
	t.Helper()
	if descriptor == nil || !descriptor.Inverse || !descriptor.Unique || descriptor.Required != required || descriptor.Field != fieldName || descriptor.RefName != ref {
		t.Fatalf("%s = %#v, want inverse unique field %q ref %q required %t", name, descriptor, fieldName, ref, required)
	}
	assertOnDelete(t, name, descriptor.Annotations, onDelete)
}

func assertDeliveryPlainUniqueEdge(t *testing.T, name string, descriptor *edge.Descriptor) {
	t.Helper()
	if descriptor == nil || descriptor.Inverse || !descriptor.Unique || descriptor.Required || descriptor.RefName != "" {
		t.Fatalf("%s = %#v, want plain unique target edge", name, descriptor)
	}
}

func TestMailDeliverySchemas_RejectWrongOpaqueLengths(t *testing.T) {
	for _, subject := range []struct {
		name  string
		field *field.Descriptor
	}{
		{"MailAttachment.engine_blob_id", fieldDescriptors((entschema.MailAttachment{}).Fields())["engine_blob_id"]},
		{"MailDraft.engine_draft_key", fieldDescriptors((entschema.MailDraft{}).Fields())["engine_draft_key"]},
		{"MailDraft.engine_message_id", fieldDescriptors((entschema.MailDraft{}).Fields())["engine_message_id"]},
		{"MailDraft.engine_version", fieldDescriptors((entschema.MailDraft{}).Fields())["engine_version"]},
		{"MailDraft.pending_create_upload_blob_id", fieldDescriptors((entschema.MailDraft{}).Fields())["pending_create_upload_blob_id"]},
		{"MailDraft.pending_create_mime_blob_id", fieldDescriptors((entschema.MailDraft{}).Fields())["pending_create_mime_blob_id"]},
		{"MailSubmission.engine_submission_key", fieldDescriptors((entschema.MailSubmission{}).Fields())["engine_submission_key"]},
		{"MailSubmission.engine_account_id", fieldDescriptors((entschema.MailSubmission{}).Fields())["engine_account_id"]},
		{"MailSubmission.engine_draft_key", fieldDescriptors((entschema.MailSubmission{}).Fields())["engine_draft_key"]},
		{"MailSubmission.engine_message_id", fieldDescriptors((entschema.MailSubmission{}).Fields())["engine_message_id"]},
		{"MailSubmission.engine_expected_version", fieldDescriptors((entschema.MailSubmission{}).Fields())["engine_expected_version"]},
		{"MailSubmission.provider_submission_id", fieldDescriptors((entschema.MailSubmission{}).Fields())["provider_submission_id"]},
		{"MailSubmission.provider_message_id", fieldDescriptors((entschema.MailSubmission{}).Fields())["provider_message_id"]},
	} {
		t.Run(subject.name, func(t *testing.T) {
			assertStringValidators(t, subject.name, subject.field, []struct{ value string; want bool }{{"opaque", true}, {"", false}, {strings.Repeat("x", 1024), true}, {strings.Repeat("x", 1025), false}})
		})
	}
}
