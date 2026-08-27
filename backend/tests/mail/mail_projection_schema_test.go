package mail_test

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"entgo.io/ent"
	"entgo.io/ent/dialect"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestMailProjectionSchemas_HaveFrozenFieldsMixinsAndJSONStorage(t *testing.T) {
	folder := entschema.MailFolder{}
	assertExactOrderedFieldNames(t, "MailFolder fields", folder.Fields(), []string{
		"id", "mail_account_id", "engine_folder_id", "parent_id", "role", "name", "normalized_name", "color", "sort_order", "unread_count", "total_count", "control_state", "control_receipt_id", "last_seen_sync_epoch", "created_at", "updated_at", "deleted_at",
	})
	folderFields := fieldDescriptors(folder.Fields())
	for _, name := range []string{"id", "mail_account_id", "parent_id", "control_receipt_id"} {
		assertFieldType(t, name, folderFields[name], field.TypeUUID)
	}
	for _, name := range []string{"engine_folder_id", "role", "color", "parent_id", "control_receipt_id", "deleted_at"} {
		assertOptionalNillable(t, name, folderFields[name])
	}
	for _, name := range []string{"id", "mail_account_id"} {
		if !folderFields[name].Immutable {
			t.Fatalf("MailFolder.%s must be immutable", name)
		}
	}
	if !folderFields["role"].Immutable {
		t.Fatal("MailFolder.role must be immutable")
	}
	assertStringValidators(t, "normalized_name", folderFields["normalized_name"], []struct {
		value string
		want  bool
	}{
		{"inbox", true},
		{" inbox", false},
		{"inbox  archive", false},
		{"inbox\u00a0archive", false},
		{"inbox\u2003archive", false},
		{"INBOX", false},
		{"cafe\u0301", false},
		{"caf\u00e9", true},
		{"stra\u00dfe", false},
		{"strasse", true},
	})
	assertExactDefaults(t, folderFields, map[string]any{
		"sort_order": int32(0), "unread_count": int32(0), "total_count": int32(0), "last_seen_sync_epoch": int64(0),
	})
	assertVersionAndTimestamps(t, folder, folderFields)

	thread := entschema.MailThread{}
	assertExactOrderedFieldNames(t, "MailThread fields", thread.Fields(), []string{
		"id", "mail_account_id", "thread_key_version", "thread_key", "normalized_subject", "participant_addresses", "last_message_at", "message_count", "unread_count", "is_pinned", "has_attachments", "snippet", "last_seen_sync_epoch", "created_at", "updated_at", "deleted_at",
	})
	threadFields := fieldDescriptors(thread.Fields())
	assertNoMixins(t, thread.Mixin())
	for _, name := range []string{"id", "mail_account_id"} {
		assertFieldType(t, name, threadFields[name], field.TypeUUID)
		if !threadFields[name].Immutable {
			t.Fatalf("MailThread.%s must be immutable", name)
		}
	}
	assertFieldType(t, "thread_key_version", threadFields["thread_key_version"], field.TypeInt32)
	assertFieldType(t, "thread_key", threadFields["thread_key"], field.TypeBytes)
	assertFieldType(t, "participant_addresses", threadFields["participant_addresses"], field.TypeJSON)
	assertRequired(t, "participant_addresses", threadFields["participant_addresses"])
	if got := threadFields["participant_addresses"].SchemaType[dialect.Postgres]; got != "jsonb" {
		t.Fatalf("MailThread.participant_addresses PostgreSQL type = %q, want jsonb", got)
	}
	participantFieldValidator := false
	for _, rawValidator := range threadFields["participant_addresses"].Validators {
		validator, ok := rawValidator.(func(entschema.MailThreadParticipants) error)
		if !ok {
			continue
		}
		participantFieldValidator = true
		if err := validator(entschema.MailThreadParticipants{"alice@example.test"}); err != nil {
			t.Fatalf("MailThread participant typed field validator rejected canonical value: %v", err)
		}
		if err := validator(entschema.MailThreadParticipants{"alice@example.test", "alice@example.test"}); err == nil {
			t.Fatal("MailThread participant typed field validator accepted duplicate address")
		}
	}
	if !participantFieldValidator {
		t.Fatal("MailThread.participant_addresses must attach its typed persistence validator")
	}
	assertOptionalNillable(t, "deleted_at", threadFields["deleted_at"])
	assertExactDefaults(t, threadFields, map[string]any{"thread_key_version": int32(1), "last_seen_sync_epoch": int64(0)})
	assertTimestampsWithoutMixin(t, threadFields)

	message := entschema.MailMessage{}
	assertExactOrderedFieldNames(t, "MailMessage fields", message.Fields(), []string{
		"id", "mail_account_id", "mail_thread_id", "engine_message_id", "engine_version", "internet_message_id", "in_reply_to", "references", "from", "to", "cc", "bcc", "reply_to", "subject", "snippet", "sent_at", "received_at", "size_bytes", "priority", "has_attachments", "is_draft", "is_seen", "is_answered", "is_flagged", "is_junk", "engine_state", "projection_origin", "flags_synced_at", "calendar_method", "calendar_event_id", "last_seen_sync_epoch", "received_event_id", "created_at", "updated_at", "deleted_at",
	})
	messageFields := fieldDescriptors(message.Fields())
	assertNoMixins(t, message.Mixin())
	for _, name := range []string{"id", "mail_account_id", "mail_thread_id", "calendar_event_id", "received_event_id"} {
		assertFieldType(t, name, messageFields[name], field.TypeUUID)
	}
	for _, name := range []string{"id", "mail_account_id", "engine_message_id"} {
		if !messageFields[name].Immutable {
			t.Fatalf("MailMessage.%s must be immutable", name)
		}
	}
	for _, name := range []string{"internet_message_id", "in_reply_to", "engine_state", "sent_at", "received_at", "flags_synced_at", "calendar_method", "calendar_event_id", "received_event_id", "deleted_at"} {
		assertOptionalNillable(t, name, messageFields[name])
	}
	internetMessageIDCollation := ""
	for _, annotation := range messageFields["internet_message_id"].Annotations {
		switch sqlAnnotation := annotation.(type) {
		case entsql.Annotation:
			if sqlAnnotation.Collation != "" {
				internetMessageIDCollation = sqlAnnotation.Collation
			}
		case *entsql.Annotation:
			if sqlAnnotation != nil && sqlAnnotation.Collation != "" {
				internetMessageIDCollation = sqlAnnotation.Collation
			}
		}
	}
	if internetMessageIDCollation != "C" {
		t.Fatalf("MailMessage.internet_message_id EntSQL collation = %q, want C", internetMessageIDCollation)
	}
	for _, name := range []string{"references", "from", "to", "cc", "bcc", "reply_to"} {
		assertFieldType(t, name, messageFields[name], field.TypeJSON)
		assertRequired(t, name, messageFields[name])
		if got := messageFields[name].SchemaType[dialect.Postgres]; got != "jsonb" {
			t.Fatalf("MailMessage.%s PostgreSQL type = %q, want jsonb", name, got)
		}
	}
	assertExactDefaults(t, messageFields, map[string]any{
		"references": entschema.MailMessageReferences{}, "to": entschema.MailStoredAddresses{}, "cc": entschema.MailStoredAddresses{}, "bcc": entschema.MailStoredAddresses{}, "reply_to": entschema.MailStoredAddresses{}, "last_seen_sync_epoch": int64(0),
	})
	if messageFields["from"].Default != nil {
		t.Fatalf("MailMessage.from default = %#v, want none", messageFields["from"].Default)
	}
	assertTimestampsWithoutMixin(t, messageFields)

	messageFolder := entschema.MailMessageFolder{}
	assertExactOrderedFieldNames(t, "MailMessageFolder fields", messageFolder.Fields(), []string{
		"id", "mail_message_id", "mail_folder_id", "engine_state", "last_seen_sync_epoch", "projected_at", "created_at", "updated_at",
	})
	messageFolderFields := fieldDescriptors(messageFolder.Fields())
	assertNoMixins(t, messageFolder.Mixin())
	for _, name := range []string{"id", "mail_message_id", "mail_folder_id"} {
		assertFieldType(t, name, messageFolderFields[name], field.TypeUUID)
		if !messageFolderFields[name].Immutable {
			t.Fatalf("MailMessageFolder.%s must be immutable", name)
		}
	}
	assertOptionalNillable(t, "engine_state", messageFolderFields["engine_state"])
	assertExactDefaults(t, messageFolderFields, map[string]any{"last_seen_sync_epoch": int64(0)})
	assertTimestampsWithoutMixin(t, messageFolderFields)

	label := entschema.MailLabel{}
	assertExactOrderedFieldNames(t, "MailLabel fields", label.Fields(), []string{
		"id", "mail_account_id", "name", "normalized_name", "color", "engine_keyword", "created_at", "updated_at", "deleted_at",
	})
	labelFields := fieldDescriptors(label.Fields())
	for _, name := range []string{"id", "mail_account_id"} {
		assertFieldType(t, name, labelFields[name], field.TypeUUID)
		if !labelFields[name].Immutable {
			t.Fatalf("MailLabel.%s must be immutable", name)
		}
	}
	if !labelFields["engine_keyword"].Immutable {
		t.Fatal("MailLabel.engine_keyword must be immutable")
	}
	assertStringValidators(t, "normalized_name", labelFields["normalized_name"], []struct {
		value string
		want  bool
	}{
		{"label", true},
		{" label", false},
		{"label  name", false},
		{"label\u00a0name", false},
		{"label\u2003name", false},
		{"LABEL", false},
		{"cafe\u0301", false},
		{"caf\u00e9", true},
		{"stra\u00dfe", false},
		{"strasse", true},
	})
	assertOptionalNillable(t, "deleted_at", labelFields["deleted_at"])
	assertVersionAndTimestamps(t, label, labelFields)

	messageLabel := entschema.MailMessageLabel{}
	assertExactOrderedFieldNames(t, "MailMessageLabel fields", messageLabel.Fields(), []string{
		"id", "mail_message_id", "mail_label_id", "manual", "engine_rule", "created_at", "updated_at",
	})
	messageLabelFields := fieldDescriptors(messageLabel.Fields())
	assertNoMixins(t, messageLabel.Mixin())
	for _, name := range []string{"id", "mail_message_id", "mail_label_id"} {
		assertFieldType(t, name, messageLabelFields[name], field.TypeUUID)
		if !messageLabelFields[name].Immutable {
			t.Fatalf("MailMessageLabel.%s must be immutable", name)
		}
	}
	for _, name := range []string{"manual", "engine_rule"} {
		assertFieldType(t, name, messageLabelFields[name], field.TypeBool)
		assertRequired(t, name, messageLabelFields[name])
	}
	assertTimestampsWithoutMixin(t, messageLabelFields)
}

func TestMailProjectionSchemas_HaveExactChecksAndIndexes(t *testing.T) {
	assertExactChecks(t, (entschema.MailFolder{}).Annotations(), map[string]string{
		"mail_folders_engine_folder_id":  "COALESCE(engine_folder_id IS NULL OR octet_length(engine_folder_id) BETWEEN 1 AND 1024, FALSE)",
		"mail_folders_text":              "COALESCE(char_length(name) BETWEEN 1 AND 120 AND octet_length(name) BETWEEN 1 AND 480 AND octet_length(normalized_name) BETWEEN 1 AND 480 AND (color IS NULL OR (char_length(color) BETWEEN 1 AND 32 AND octet_length(color) BETWEEN 1 AND 128)), FALSE)",
		"mail_folders_role":              "COALESCE(role IS NULL OR role IN ('inbox','sent','drafts','archive','trash','junk'), FALSE)",
		"mail_folders_control_shape":     "COALESCE((control_state = 'active' AND engine_folder_id IS NOT NULL AND control_receipt_id IS NULL AND deleted_at IS NULL) OR (control_state = 'creating' AND engine_folder_id IS NULL AND control_receipt_id IS NOT NULL AND deleted_at IS NULL) OR (control_state IN ('unknown','error') AND control_receipt_id IS NOT NULL AND deleted_at IS NULL) OR (control_state = 'deleting' AND engine_folder_id IS NOT NULL AND ((deleted_at IS NULL AND control_receipt_id IS NOT NULL) OR (deleted_at IS NOT NULL AND control_receipt_id IS NULL))), FALSE)",
		"mail_folders_counters":          "COALESCE(unread_count >= 0 AND total_count >= 0 AND unread_count <= total_count, FALSE)",
		"mail_folders_non_active_counts": "COALESCE(control_state = 'active' OR (unread_count = 0 AND total_count = 0), FALSE)",
		"mail_folders_role_parent":       "COALESCE(role IS NULL OR parent_id IS NULL, FALSE)",
		"mail_folders_not_self_parent":   "COALESCE(parent_id IS NULL OR parent_id <> id, FALSE)",
		"mail_folders_epoch":             "COALESCE(last_seen_sync_epoch BETWEEN 0 AND 2147483647, FALSE)",
		"mail_folders_version":           "COALESCE(version >= 1, FALSE)",
	})
	assertExactChecks(t, (entschema.MailThread{}).Annotations(), map[string]string{
		"mail_threads_key":          "COALESCE(thread_key_version = 1 AND octet_length(thread_key) = 32, FALSE)",
		"mail_threads_subject":      "COALESCE(octet_length(normalized_subject) <= 998, FALSE)",
		"mail_threads_participants": "CASE WHEN jsonb_typeof(participant_addresses) = 'array' THEN COALESCE(jsonb_array_length(participant_addresses) BETWEEN 1 AND 401 AND NOT jsonb_path_exists(participant_addresses, '$[*] ? (@.type() != \"string\")'), FALSE) ELSE FALSE END",
		"mail_threads_counts":       "COALESCE(message_count >= 1 AND unread_count BETWEEN 0 AND message_count, FALSE)",
		"mail_threads_snippet":      "COALESCE(char_length(snippet) <= 1200 AND octet_length(snippet) <= 4096, FALSE)",
		"mail_threads_epoch":        "COALESCE(last_seen_sync_epoch BETWEEN 0 AND 2147483647, FALSE)",
	})
	assertExactChecks(t, (entschema.MailMessage{}).Annotations(), map[string]string{
		"mail_messages_engine_ids":      "COALESCE(octet_length(engine_message_id) BETWEEN 1 AND 1024 AND octet_length(engine_version) BETWEEN 1 AND 1024, FALSE)",
		"mail_messages_message_ids":     "COALESCE((internet_message_id IS NULL OR octet_length(internet_message_id) BETWEEN 1 AND 998) AND (in_reply_to IS NULL OR octet_length(in_reply_to) BETWEEN 1 AND 998), FALSE)",
		"mail_messages_references":      "CASE WHEN jsonb_typeof(references) = 'array' THEN COALESCE(jsonb_array_length(references) <= 100 AND NOT jsonb_path_exists(references, '$[*] ? (@.type() != \"string\")'), FALSE) ELSE FALSE END",
		"mail_messages_from_shape":      "COALESCE(jsonb_typeof(\"from\") = 'object' AND jsonb_typeof(\"from\"->'email') = 'string' AND jsonb_typeof(\"from\"->'name') = 'string', FALSE)",
		"mail_messages_address_arrays":  "CASE WHEN jsonb_typeof(\"to\") = 'array' AND jsonb_typeof(cc) = 'array' AND jsonb_typeof(bcc) = 'array' AND jsonb_typeof(reply_to) = 'array' THEN COALESCE(jsonb_array_length(\"to\") <= 200 AND jsonb_array_length(cc) <= 200 AND jsonb_array_length(bcc) <= 200 AND jsonb_array_length(reply_to) <= 200 AND NOT jsonb_path_exists(\"to\", '$[*] ? (@.type() != \"object\" || !exists(@.email) || !exists(@.name) || @.email.type() != \"string\" || @.name.type() != \"string\")') AND NOT jsonb_path_exists(cc, '$[*] ? (@.type() != \"object\" || !exists(@.email) || !exists(@.name) || @.email.type() != \"string\" || @.name.type() != \"string\")') AND NOT jsonb_path_exists(bcc, '$[*] ? (@.type() != \"object\" || !exists(@.email) || !exists(@.name) || @.email.type() != \"string\" || @.name.type() != \"string\")') AND NOT jsonb_path_exists(reply_to, '$[*] ? (@.type() != \"object\" || !exists(@.email) || !exists(@.name) || @.email.type() != \"string\" || @.name.type() != \"string\")'), FALSE) ELSE FALSE END",
		"mail_messages_text":            "COALESCE(octet_length(subject) <= 998 AND char_length(snippet) <= 1200 AND octet_length(snippet) <= 4096 AND (engine_state IS NULL OR octet_length(engine_state) <= 4096), FALSE)",
		"mail_messages_size":            "COALESCE(size_bytes BETWEEN 0 AND 52428800, FALSE)",
		"mail_messages_epoch":           "COALESCE(last_seen_sync_epoch BETWEEN 0 AND 2147483647, FALSE)",
		"mail_messages_priority":        "COALESCE(priority IN ('low','normal','high'), FALSE)",
		"mail_messages_projection_origin": "COALESCE(projection_origin IN ('inbound_or_unknown','confirmed_owner_submission','confirmed_owner_draft'), FALSE)",
		"mail_messages_calendar_method": "COALESCE(calendar_method IS NULL OR calendar_method IN ('request','cancel','reply'), FALSE)",
		"mail_messages_bcc_origin":      "CASE WHEN jsonb_typeof(bcc) = 'array' THEN COALESCE(projection_origin IN ('confirmed_owner_submission','confirmed_owner_draft') OR jsonb_array_length(bcc) = 0, FALSE) ELSE FALSE END",
	})
	assertExactChecks(t, (entschema.MailMessageFolder{}).Annotations(), map[string]string{
		"mail_message_folders_engine_state": "COALESCE(engine_state IS NULL OR octet_length(engine_state) <= 4096, FALSE)",
		"mail_message_folders_epoch":        "COALESCE(last_seen_sync_epoch BETWEEN 0 AND 2147483647, FALSE)",
	})
	assertExactChecks(t, (entschema.MailLabel{}).Annotations(), map[string]string{
		"mail_labels_text":    "COALESCE(char_length(name) BETWEEN 1 AND 60 AND octet_length(name) BETWEEN 1 AND 240 AND octet_length(normalized_name) BETWEEN 1 AND 240 AND char_length(color) BETWEEN 1 AND 32 AND octet_length(color) BETWEEN 1 AND 128 AND octet_length(engine_keyword) = 57 AND engine_keyword ~ '^hubk-[a-z2-7]{52}$', FALSE)",
		"mail_labels_version": "COALESCE(version >= 1, FALSE)",
	})
	assertExactChecks(t, (entschema.MailMessageLabel{}).Annotations(), map[string]string{
		"mail_message_labels_manual_or_engine_rule": "COALESCE(manual OR engine_rule, FALSE)",
	})

	folderIndexes := (entschema.MailFolder{}).Indexes()
	assertExactIndexes(t, folderIndexes, []wantedIndex{
		{fields: []string{"mail_account_id", "engine_folder_id"}, unique: true, where: "engine_folder_id IS NOT NULL"},
		{fields: []string{"mail_account_id", "role"}, unique: true, where: "deleted_at IS NULL AND role IS NOT NULL"},
		{fields: []string{"mail_account_id", "normalized_name"}, unique: true, where: "deleted_at IS NULL AND role IS NULL AND parent_id IS NULL"},
		{fields: []string{"mail_account_id", "parent_id", "normalized_name"}, unique: true, where: "deleted_at IS NULL AND role IS NULL AND parent_id IS NOT NULL"},
	})
	assertProjectionIndexStorageKeys(t, "MailFolder", folderIndexes, []string{"mail_folders_account_engine_folder_uidx", "mail_folders_live_role_uidx", "mail_folders_live_custom_root_uidx", "mail_folders_live_custom_child_uidx"})

	threadIndexes := (entschema.MailThread{}).Indexes()
	assertExactIndexes(t, threadIndexes, []wantedIndex{
		{fields: []string{"mail_account_id", "thread_key_version", "thread_key"}, unique: true},
		{fields: []string{"mail_account_id", "is_pinned", "last_message_at", "id"}, where: "deleted_at IS NULL"},
	})
	assertProjectionIndexStorageKeys(t, "MailThread", threadIndexes, []string{"mail_threads_account_version_key_uidx", "mail_threads_live_pinned_idx"})

	messageIndexes := (entschema.MailMessage{}).Indexes()
	assertExactIndexes(t, messageIndexes, []wantedIndex{
		{fields: []string{"mail_account_id", "engine_message_id"}, unique: true},
		{fields: []string{"mail_account_id", "internet_message_id"}, where: "deleted_at IS NULL"},
		{fields: []string{"received_event_id"}, unique: true},
		{fields: []string{"mail_account_id", "received_at", "sent_at", "id"}, where: "deleted_at IS NULL"},
		{fields: []string{"mail_thread_id", "received_at", "sent_at", "id"}, where: "deleted_at IS NULL"},
	})
	assertProjectionIndexStorageKeys(t, "MailMessage", messageIndexes, []string{"mail_messages_account_engine_uidx", "mail_messages_live_internet_message_id_idx", "mail_messages_received_event_uidx", "mail_messages_live_account_date_idx", "mail_messages_live_thread_date_idx"})

	messageFolderIndexes := (entschema.MailMessageFolder{}).Indexes()
	assertExactIndexes(t, messageFolderIndexes, []wantedIndex{{fields: []string{"mail_message_id", "mail_folder_id"}, unique: true}})
	assertProjectionIndexStorageKeys(t, "MailMessageFolder", messageFolderIndexes, []string{"mail_message_folders_message_folder_uidx"})

	labelIndexes := (entschema.MailLabel{}).Indexes()
	assertExactIndexes(t, labelIndexes, []wantedIndex{
		{fields: []string{"mail_account_id", "normalized_name"}, unique: true, where: "deleted_at IS NULL"},
		{fields: []string{"mail_account_id", "engine_keyword"}, unique: true},
	})
	assertProjectionIndexStorageKeys(t, "MailLabel", labelIndexes, []string{"mail_labels_live_normalized_name_uidx", "mail_labels_account_keyword_uidx"})

	messageLabelIndexes := (entschema.MailMessageLabel{}).Indexes()
	assertExactIndexes(t, messageLabelIndexes, []wantedIndex{{fields: []string{"mail_message_id", "mail_label_id"}, unique: true}})
	assertProjectionIndexStorageKeys(t, "MailMessageLabel", messageLabelIndexes, []string{"mail_message_labels_message_label_uidx"})
}

func TestMailProjectionSchemas_HaveOnlyFrozenRelations(t *testing.T) {
	folderEdges := edgeDescriptors((entschema.MailFolder{}).Edges())
	assertExactNames(t, "MailFolder edges", edgeNames(folderEdges), []string{"account", "children", "control_receipt", "message_folders", "parent"})
	assertReceiptParentEdge(t, "MailFolder.account", folderEdges["account"], "mail_account_id", "folders", true, entsql.Cascade)
	assertReceiptParentEdge(t, "MailFolder.parent", folderEdges["parent"], "parent_id", "children", false, entsql.Restrict)
	assertReceiptParentEdge(t, "MailFolder.control_receipt", folderEdges["control_receipt"], "control_receipt_id", "folders", false, entsql.Restrict)
	assertReceiptChildEdge(t, "MailFolder.children", folderEdges["children"])
	assertReceiptChildEdge(t, "MailFolder.message_folders", folderEdges["message_folders"])

	threadEdges := edgeDescriptors((entschema.MailThread{}).Edges())
	assertExactNames(t, "MailThread edges", edgeNames(threadEdges), []string{"account", "messages"})
	assertReceiptParentEdge(t, "MailThread.account", threadEdges["account"], "mail_account_id", "threads", true, entsql.Cascade)
	assertReceiptChildEdge(t, "MailThread.messages", threadEdges["messages"])

	messageEdges := edgeDescriptors((entschema.MailMessage{}).Edges())
	assertExactNames(t, "MailMessage edges", edgeNames(messageEdges), []string{"account", "attachment_records", "calendar_event", "folders", "forward_drafts", "labels", "reply_drafts", "thread"})
	assertReceiptParentEdge(t, "MailMessage.account", messageEdges["account"], "mail_account_id", "messages", true, entsql.Cascade)
	assertReceiptParentEdge(t, "MailMessage.thread", messageEdges["thread"], "mail_thread_id", "messages", true, entsql.Cascade)
	assertReceiptParentEdge(t, "MailMessage.calendar_event", messageEdges["calendar_event"], "calendar_event_id", "mail_messages", false, entsql.SetNull)
	assertReceiptChildEdge(t, "MailMessage.folders", messageEdges["folders"])
	assertReceiptChildEdge(t, "MailMessage.labels", messageEdges["labels"])
	assertReceiptChildEdge(t, "MailMessage.attachment_records", messageEdges["attachment_records"])
	assertReceiptChildEdge(t, "MailMessage.reply_drafts", messageEdges["reply_drafts"])
	assertReceiptChildEdge(t, "MailMessage.forward_drafts", messageEdges["forward_drafts"])

	messageFolderEdges := edgeDescriptors((entschema.MailMessageFolder{}).Edges())
	assertExactNames(t, "MailMessageFolder edges", edgeNames(messageFolderEdges), []string{"folder", "message"})
	assertReceiptParentEdge(t, "MailMessageFolder.message", messageFolderEdges["message"], "mail_message_id", "folders", true, entsql.Cascade)
	assertReceiptParentEdge(t, "MailMessageFolder.folder", messageFolderEdges["folder"], "mail_folder_id", "message_folders", true, entsql.Cascade)

	labelEdges := edgeDescriptors((entschema.MailLabel{}).Edges())
	assertExactNames(t, "MailLabel edges", edgeNames(labelEdges), []string{"account", "message_labels"})
	assertReceiptParentEdge(t, "MailLabel.account", labelEdges["account"], "mail_account_id", "labels", true, entsql.Cascade)
	assertReceiptChildEdge(t, "MailLabel.message_labels", labelEdges["message_labels"])

	messageLabelEdges := edgeDescriptors((entschema.MailMessageLabel{}).Edges())
	assertExactNames(t, "MailMessageLabel edges", edgeNames(messageLabelEdges), []string{"label", "message"})
	assertReceiptParentEdge(t, "MailMessageLabel.message", messageLabelEdges["message"], "mail_message_id", "labels", true, entsql.Cascade)
	assertReceiptParentEdge(t, "MailMessageLabel.label", messageLabelEdges["label"], "mail_label_id", "message_labels", true, entsql.Cascade)
}

func TestMailProjectionJSON_IsStrictCanonicalAndBounded(t *testing.T) {
	var participants entschema.MailThreadParticipants
	if err := json.Unmarshal([]byte(`["alice@example.test","bob@example.test"]`), &participants); err != nil {
		t.Fatalf("unmarshal canonical participants: %v", err)
	}
	for _, raw := range []string{
		`[]`,
		`["alice@example.test","alice@example.test"]`,
		`["Alice@example.test"]`,
		`["alice@example.test",1]`,
	} {
		if err := json.Unmarshal([]byte(raw), &participants); err == nil {
			t.Fatalf("participants %s must be rejected", raw)
		}
	}
	canonicalParticipantsJSON := `["alice@example.test"]`
	paddedParticipantsJSON := []byte(strings.Repeat(" ", 1_048_577-len(canonicalParticipantsJSON)) + canonicalParticipantsJSON)
	if err := json.Unmarshal(paddedParticipantsJSON, &participants); err != nil {
		t.Fatalf("unmarshal whitespace-padded canonical participants: %v", err)
	}
	if len(participants) != 1 || participants[0] != "alice@example.test" {
		t.Fatalf("whitespace-padded participants = %#v, want canonical value", participants)
	}
	if err := json.Unmarshal([]byte(`["alice\u0040example.test"]`), &participants); err != nil {
		t.Fatalf("unmarshal escaped canonical participant: %v", err)
	}
	if len(participants) != 1 || participants[0] != "alice@example.test" {
		t.Fatalf("escaped participant = %#v, want alice@example.test", participants)
	}
	lexicallyOversizedParticipantJSON := []byte("[\"" + strings.Repeat("a", 6*320+1) + "\"]")
	if err := json.Unmarshal(lexicallyOversizedParticipantJSON, &participants); err == nil || !strings.Contains(err.Error(), "lexical budget") {
		t.Fatalf("lexically oversized participant error = %v, want lexical budget rejection", err)
	}
	tooManyParticipants := make([]string, 402)
	for index := range tooManyParticipants {
		tooManyParticipants[index] = fmt.Sprintf("member%03d@example.test", index)
	}
	tooManyParticipantsJSON, err := json.Marshal(tooManyParticipants)
	if err != nil {
		t.Fatalf("marshal 402 canonical participants: %v", err)
	}
	if err := json.Unmarshal(tooManyParticipantsJSON, &participants); err == nil {
		t.Fatal("402 canonical participants must be rejected")
	}
	if err := (entschema.MailThreadParticipants{strings.Repeat("a", 65) + "@example.test"}).Validate(); err == nil {
		t.Fatal("participant local part longer than the canonical address limit must be rejected")
	}

	var address entschema.MailStoredAddress
	if err := json.Unmarshal([]byte(`{"email":"alice@example.test","name":"Alice"}`), &address); err != nil {
		t.Fatalf("unmarshal canonical stored address: %v", err)
	}
	for _, raw := range []string{
		`{"email":"Alice@example.test","name":"Alice"}`,
		`{"email":"alice@example.test","name":"Alice","extra":true}`,
		`{"email":"alice@example.test","email":"bob@example.test","name":"Alice"}`,
		`{"email":"alice@example.test"}`,
	} {
		if err := json.Unmarshal([]byte(raw), &address); err == nil {
			t.Fatalf("stored address %s must be rejected", raw)
		}
	}
	canonicalStoredAddressJSON := `{"email":"alice@example.test","name":"Alice"}`
	paddedStoredAddressJSON := []byte(strings.Repeat(" ", 65_537-len(canonicalStoredAddressJSON)) + canonicalStoredAddressJSON)
	if err := json.Unmarshal(paddedStoredAddressJSON, &address); err != nil {
		t.Fatalf("unmarshal whitespace-padded canonical stored address: %v", err)
	}
	if address.Email != "alice@example.test" || address.Name != "Alice" {
		t.Fatalf("whitespace-padded stored address = %#v, want canonical value", address)
	}
	escapedEmojiName := strings.Repeat(`\ud83d\ude00`, 160)
	if err := json.Unmarshal([]byte(`{"email":"alice\u0040example.test","name":"`+escapedEmojiName+`"}`), &address); err != nil {
		t.Fatalf("unmarshal escaped email and 160-emoji name: %v", err)
	}
	if address.Email != "alice@example.test" || address.Name != strings.Repeat("😀", 160) {
		t.Fatalf("escaped stored address = %#v, want canonical email and 160-emoji name", address)
	}
	lexicallyOversizedStoredAddressJSON := []byte(`{"email":"` + strings.Repeat("a", 6*640+1) + `"}`)
	if err := json.Unmarshal(lexicallyOversizedStoredAddressJSON, &address); err == nil || !strings.Contains(err.Error(), "lexical budget") {
		t.Fatalf("lexically oversized stored address error = %v, want lexical budget rejection", err)
	}

	var addresses entschema.MailStoredAddresses
	if err := json.Unmarshal([]byte(`[{"email":"alice@example.test","name":"Alice"}]`), &addresses); err != nil {
		t.Fatalf("unmarshal canonical stored addresses: %v", err)
	}
	if err := json.Unmarshal([]byte(`[{"email":"alice@example.test","name":"Alice"},{"email":"alice@example.test","name":"Alice A."}]`), &addresses); err == nil {
		t.Fatal("duplicate stored address email must be rejected")
	}
	canonicalStoredAddressesJSON := `[{"email":"alice@example.test","name":"Alice"}]`
	paddedStoredAddressesJSON := []byte(strings.Repeat(" ", 65_537-len(canonicalStoredAddressesJSON)) + canonicalStoredAddressesJSON)
	if err := json.Unmarshal(paddedStoredAddressesJSON, &addresses); err != nil {
		t.Fatalf("unmarshal whitespace-padded canonical stored addresses: %v", err)
	}
	if len(addresses) != 1 || addresses[0].Email != "alice@example.test" || addresses[0].Name != "Alice" {
		t.Fatalf("whitespace-padded stored addresses = %#v, want canonical value", addresses)
	}

	var references entschema.MailMessageReferences
	if err := json.Unmarshal([]byte(`["<one@example.test>","<two@example.test>"]`), &references); err != nil {
		t.Fatalf("unmarshal strict message-id references: %v", err)
	}
	for _, raw := range []string{
		`["<one@example.test>","<one@example.test>"]`,
		`["one@example.test"]`,
		`["<one@example.test>",1]`,
	} {
		if err := json.Unmarshal([]byte(raw), &references); err == nil {
			t.Fatalf("references %s must be rejected", raw)
		}
	}
	canonicalReferencesJSON := `["<one@example.test>"]`
	paddedReferencesJSON := []byte(strings.Repeat(" ", 1_048_577-len(canonicalReferencesJSON)) + canonicalReferencesJSON)
	if err := json.Unmarshal(paddedReferencesJSON, &references); err != nil {
		t.Fatalf("unmarshal whitespace-padded canonical references: %v", err)
	}
	if len(references) != 1 || references[0] != "<one@example.test>" {
		t.Fatalf("whitespace-padded references = %#v, want canonical value", references)
	}
	if err := json.Unmarshal([]byte(`["<one\u0040example.test>"]`), &references); err != nil {
		t.Fatalf("unmarshal escaped canonical reference: %v", err)
	}
	if len(references) != 1 || references[0] != "<one@example.test>" {
		t.Fatalf("escaped reference = %#v, want <one@example.test>", references)
	}
	lexicallyOversizedReferencesJSON := []byte("[\"" + strings.Repeat("a", 6*998+1) + "\"]")
	if err := json.Unmarshal(lexicallyOversizedReferencesJSON, &references); err == nil || !strings.Contains(err.Error(), "lexical budget") {
		t.Fatalf("lexically oversized references error = %v, want lexical budget rejection", err)
	}
}

func assertProjectionIndexStorageKeys(t *testing.T, entity string, indexes []ent.Index, want []string) {
	t.Helper()
	if len(indexes) != len(want) {
		t.Fatalf("%s index count = %d, want %d", entity, len(indexes), len(want))
	}
	for index, storageKey := range want {
		if got := indexes[index].Descriptor().StorageKey; got != storageKey {
			t.Fatalf("%s index[%d] storage key = %q, want %q", entity, index, got, storageKey)
		}
	}
}
