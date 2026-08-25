package calendar_test

import (
	"crypto/sha256"
	"os"
	"reflect"
	"sort"
	"strings"
	"testing"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestCalendarInboundObjectState_SchemaContract(t *testing.T) {
	fields := mailBridgeFieldDescriptors((entschema.CalendarInboundObjectState{}).Fields())
	mailBridgeAssertExactNames(t, "CalendarInboundObjectState fields", mailBridgeMapKeys(fields), []string{
		"calendar_event_id", "calendar_owner_id", "created_at", "expires_at", "id", "method",
		"organization_id", "organizer_email", "payload_hash", "payload_hash_version", "recurrence_id",
		"sequence", "state", "uid", "updated_at",
	})

	for _, name := range []string{"id", "organization_id", "calendar_owner_id"} {
		mailBridgeAssertRequiredUUID(t, name, fields[name])
	}
	for _, name := range []string{"id", "organization_id", "calendar_owner_id", "uid", "recurrence_id", "organizer_email"} {
		mailBridgeAssertImmutable(t, name, fields[name])
	}
	mailBridgeAssertStringMax(t, "uid", fields["uid"], 255, false)
	mailBridgeAssertStringMax(t, "recurrence_id", fields["recurrence_id"], 255, true)
	mailBridgeAssertStringMax(t, "organizer_email", fields["organizer_email"], 320, false)
	mailBridgeAssertIntMinimum(t, "sequence", fields["sequence"], 0)
	mailBridgeAssertUint8Minimum(t, "payload_hash_version", fields["payload_hash_version"], 1)
	mailBridgeAssertExactSHA256(t, "payload_hash", fields["payload_hash"])
	mailBridgeAssertEnumValues(t, "method", fields["method"], []string{"request", "cancel"})
	mailBridgeAssertEnumValues(t, "state", fields["state"], []string{"applied", "tombstone", "deferred_missing_master"})
	mailBridgeAssertOptionalNillable(t, "calendar_event_id", fields["calendar_event_id"])
	mailBridgeAssertOptionalNillable(t, "expires_at", fields["expires_at"])
	mailBridgeAssertImmutable(t, "created_at", fields["created_at"])

	mailBridgeAssertChecks(t, (entschema.CalendarInboundObjectState{}).Annotations(), map[string]string{
		"calendar_inbound_object_states_payload_hash_length": "octet_length(payload_hash) = 32",
		"calendar_inbound_object_states_state_expiry":       "(state = 'deferred_missing_master') = (expires_at IS NOT NULL)",
	})
	mailBridgeAssertUniquePartialIndex(t, (entschema.CalendarInboundObjectState{}).Indexes(), []string{"calendar_owner_id", "uid"}, "recurrence_id IS NULL")
	mailBridgeAssertUniquePartialIndex(t, (entschema.CalendarInboundObjectState{}).Indexes(), []string{"calendar_owner_id", "uid", "recurrence_id"}, "recurrence_id IS NOT NULL")

	edges := mailBridgeEdgeDescriptors((entschema.CalendarInboundObjectState{}).Edges())
	mailBridgeAssertExactNames(t, "CalendarInboundObjectState edges", mailBridgeMapKeys(edges), []string{"calendar_owner", "event", "organization"})
	mailBridgeAssertEdge(t, "organization", edges["organization"], false, true, true, "organization_id", "", entsql.NoAction)
	mailBridgeAssertEdge(t, "calendar_owner", edges["calendar_owner"], false, true, true, "calendar_owner_id", "", entsql.NoAction)
	mailBridgeAssertEdge(t, "event", edges["event"], false, true, false, "calendar_event_id", "", entsql.SetNull)
}

func TestCalendarInboundMessageReceipt_SchemaContract(t *testing.T) {
	fields := mailBridgeFieldDescriptors((entschema.CalendarInboundMessageReceipt{}).Fields())
	mailBridgeAssertExactNames(t, "CalendarInboundMessageReceipt fields", mailBridgeMapKeys(fields), []string{
		"attempts", "calendar_event_id", "created_at", "error_code", "event_envelope_id", "id", "lease_until",
		"mail_message_id", "next_attempt_at", "part_locator_hash", "state", "updated_at",
	})

	for _, name := range []string{"id", "mail_message_id", "event_envelope_id"} {
		mailBridgeAssertRequiredUUID(t, name, fields[name])
		mailBridgeAssertImmutable(t, name, fields[name])
	}
	mailBridgeAssertExactSHA256(t, "part_locator_hash", fields["part_locator_hash"])
	mailBridgeAssertImmutable(t, "part_locator_hash", fields["part_locator_hash"])
	mailBridgeAssertEnumValues(t, "state", fields["state"], []string{
		"pending", "processing", "applied", "ignored_stale", "duplicate", "deferred_missing_master", "quarantined", "dead_letter",
	})
	mailBridgeAssertEnumValues(t, "error_code", fields["error_code"], []string{
		"message_unavailable", "part_unavailable", "malformed_itip", "unsupported_method", "unsupported_version", "untrusted_sender",
		"organizer_mismatch", "equal_sequence_conflict", "missing_master_expired", "resource_limit", "retry_exhausted",
		"engine_unavailable", "database_unavailable", "internal_error",
	})
	mailBridgeAssertIntMinimum(t, "attempts", fields["attempts"], 0)
	for _, name := range []string{"lease_until", "next_attempt_at", "error_code", "calendar_event_id"} {
		mailBridgeAssertOptionalNillable(t, name, fields[name])
	}
	mailBridgeAssertImmutable(t, "created_at", fields["created_at"])

	mailBridgeAssertChecks(t, (entschema.CalendarInboundMessageReceipt{}).Annotations(), map[string]string{
		"calendar_inbound_message_receipts_part_locator_hash_length": "octet_length(part_locator_hash) = 32",
	})
	mailBridgeAssertUniqueIndex(t, (entschema.CalendarInboundMessageReceipt{}).Indexes(), []string{"mail_message_id", "part_locator_hash"})

	edges := mailBridgeEdgeDescriptors((entschema.CalendarInboundMessageReceipt{}).Edges())
	mailBridgeAssertExactNames(t, "CalendarInboundMessageReceipt edges", mailBridgeMapKeys(edges), []string{"event"})
	mailBridgeAssertEdge(t, "event", edges["event"], false, true, false, "calendar_event_id", "", entsql.SetNull)
}

func TestCalendarReplyDelivery_SchemaContract(t *testing.T) {
	fields := mailBridgeFieldDescriptors((entschema.CalendarReplyDelivery{}).Fields())
	mailBridgeAssertExactNames(t, "CalendarReplyDelivery fields", mailBridgeMapKeys(fields), []string{
		"attempts", "attendee_version", "calendar_attendee_id", "calendar_event_id", "claim_generation", "created_at", "event_sequence", "id",
		"last_error_code", "lease_token", "lease_until", "mail_submission_id", "next_attempt_at", "organization_id", "recurrence_id", "rsvp_status", "state", "updated_at",
	})

	for _, name := range []string{"id", "organization_id", "calendar_event_id", "calendar_attendee_id"} {
		mailBridgeAssertRequiredUUID(t, name, fields[name])
		mailBridgeAssertImmutable(t, name, fields[name])
	}
	for _, name := range []string{"event_sequence", "attendee_version"} {
		mailBridgeAssertImmutable(t, name, fields[name])
	}
	mailBridgeAssertIntMinimum(t, "event_sequence", fields["event_sequence"], 0)
	mailBridgeAssertIntMinimum(t, "attendee_version", fields["attendee_version"], 1)
	mailBridgeAssertIntBounds(t, "attempts", fields["attempts"], 0, 8)
	mailBridgeAssertInt64Bounds(t, "claim_generation", fields["claim_generation"], 0, 2147483647)
	mailBridgeAssertDefault(t, "claim_generation", fields["claim_generation"], int64(0))
	mailBridgeAssertFieldType(t, "lease_token", fields["lease_token"], field.TypeUUID)
	mailBridgeAssertFieldType(t, "lease_until", fields["lease_until"], field.TypeTime)
	mailBridgeAssertStringMax(t, "recurrence_id", fields["recurrence_id"], 255, true)
	mailBridgeAssertImmutable(t, "recurrence_id", fields["recurrence_id"])
	mailBridgeAssertEnumValues(t, "rsvp_status", fields["rsvp_status"], []string{"accepted", "declined", "tentative"})
	mailBridgeAssertImmutable(t, "rsvp_status", fields["rsvp_status"])
	mailBridgeAssertEnumValues(t, "state", fields["state"], []string{"pending", "submitting", "unknown", "sent", "superseded", "failed_permanent"})
	mailBridgeAssertEnumValues(t, "last_error_code", fields["last_error_code"], []string{
		"mailbox_unavailable", "engine_unavailable", "submission_rejected", "submission_outcome_unknown", "invalid_recipient",
		"policy_denied", "retry_exhausted", "internal_error",
	})
	for _, name := range []string{"mail_submission_id", "lease_token", "lease_until", "next_attempt_at", "last_error_code"} {
		mailBridgeAssertOptionalNillable(t, name, fields[name])
	}
	mailBridgeAssertImmutable(t, "created_at", fields["created_at"])
	mailBridgeAssertChecks(t, (entschema.CalendarReplyDelivery{}).Annotations(), map[string]string{
		"calendar_reply_deliveries_attempts_bounds":        "attempts >= 0 AND attempts <= 8",
		"calendar_reply_deliveries_lease_pair":              "COALESCE((lease_token IS NULL) = (lease_until IS NULL), FALSE)",
		"calendar_reply_deliveries_claim_generation_bounds": "claim_generation >= 0 AND claim_generation <= 2147483647",
		"calendar_reply_deliveries_state_shape": "COALESCE((state = 'pending' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND ((attempts = 0 AND last_error_code IS NULL) OR (attempts BETWEEN 1 AND 8 AND next_attempt_at IS NOT NULL AND last_error_code IN ('mailbox_unavailable','engine_unavailable','internal_error')))) OR (state = 'submitting' AND lease_token IS NOT NULL AND lease_until IS NOT NULL AND mail_submission_id IS NULL AND claim_generation >= 1 AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL) OR (state = 'unknown' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NOT NULL AND last_error_code IN ('mailbox_unavailable','engine_unavailable','submission_outcome_unknown')) OR (state = 'sent' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL) OR (state = 'superseded' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND next_attempt_at IS NULL AND last_error_code IS NULL) OR (state = 'failed_permanent' AND lease_token IS NULL AND lease_until IS NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IN ('submission_rejected','invalid_recipient','policy_denied','retry_exhausted') AND (last_error_code <> 'retry_exhausted' OR attempts = 8)), FALSE)",
	})
	mailBridgeAssertUniqueIndex(t, (entschema.CalendarReplyDelivery{}).Indexes(), []string{"calendar_attendee_id", "event_sequence", "attendee_version"})
	mailBridgeAssertPartialIndex(t, (entschema.CalendarReplyDelivery{}).Indexes(), []string{"next_attempt_at", "id"}, "state IN ('pending','unknown') AND next_attempt_at IS NOT NULL")
	mailBridgeAssertNonUniquePartialIndex(t, (entschema.CalendarReplyDelivery{}).Indexes(), []string{"lease_until", "id"}, "calendar_reply_deliveries_stale_lease_idx", "state = 'submitting' AND lease_until IS NOT NULL")

	edges := mailBridgeEdgeDescriptors((entschema.CalendarReplyDelivery{}).Edges())
	mailBridgeAssertExactNames(t, "CalendarReplyDelivery edges", mailBridgeMapKeys(edges), []string{"attendee", "event", "mail_submission", "organization"})
	mailBridgeAssertEdge(t, "organization", edges["organization"], false, true, true, "organization_id", "", entsql.NoAction)
	mailBridgeAssertEdge(t, "event", edges["event"], false, true, true, "calendar_event_id", "", entsql.NoAction)
	mailBridgeAssertEdge(t, "attendee", edges["attendee"], false, true, true, "calendar_attendee_id", "", entsql.NoAction)
	mailBridgeAssertEdge(t, "mail_submission", edges["mail_submission"], false, true, false, "mail_submission_id", "", entsql.NoAction)

	source, err := os.ReadFile("../../ent/schema/calendarreplydelivery.go")
	if err != nil {
		t.Fatalf("read CalendarReplyDelivery schema: %v", err)
	}
	for _, forbidden := range []string{
		"itip", "field.string(\"body\")", "field.bytes(\"body\")", "field.string(\"payload\")", "field.bytes(\"payload\")",
		"field.string(\"title\")", "field.string(\"description\")", "field.string(\"normalized_email\")", "field.string(\"address\")",
		"field.string(\"last_error\")", "provider_", "engine_version", "field.string(\"organizer\")", "field.string(\"from\")",
	} {
		if strings.Contains(strings.ToLower(string(source)), forbidden) {
			t.Fatalf("CalendarReplyDelivery must not persist %q", forbidden)
		}
	}
}

func TestCalendarMailBridge_InverseEdgesPreserveDurableRecords(t *testing.T) {
	mailBridgeAssertExistingInverseEdge(t, (entschema.Organization{}).Edges(), "calendar_inbound_object_states", "organization")
	mailBridgeAssertExistingInverseEdge(t, (entschema.User{}).Edges(), "calendar_inbound_object_states", "calendar_owner")
	mailBridgeAssertExistingInverseEdge(t, (entschema.CalendarEvent{}).Edges(), "inbound_object_states", "event")
	mailBridgeAssertExistingInverseEdge(t, (entschema.CalendarAttendee{}).Edges(), "reply_deliveries", "attendee")
}

func mailBridgeFieldDescriptors(fields []ent.Field) map[string]*field.Descriptor {
	descriptors := make(map[string]*field.Descriptor, len(fields))
	for _, schemaField := range fields {
		descriptor := schemaField.Descriptor()
		descriptors[descriptor.Name] = descriptor
	}
	return descriptors
}

func mailBridgeEdgeDescriptors(edges []ent.Edge) map[string]*edge.Descriptor {
	descriptors := make(map[string]*edge.Descriptor, len(edges))
	for _, schemaEdge := range edges {
		descriptor := schemaEdge.Descriptor()
		descriptors[descriptor.Name] = descriptor
	}
	return descriptors
}

func mailBridgeAssertRequiredUUID(t *testing.T, name string, descriptor *field.Descriptor) {
	t.Helper()
	if descriptor == nil || descriptor.Info == nil || descriptor.Info.Type != field.TypeUUID {
		t.Fatalf("%s must be a UUID field", name)
	}
	if descriptor.Optional || descriptor.Nillable {
		t.Fatalf("%s optional/nillable = %t/%t, want false/false", name, descriptor.Optional, descriptor.Nillable)
	}
}

func mailBridgeAssertExactSHA256(t *testing.T, name string, descriptor *field.Descriptor) {
	t.Helper()
	if descriptor == nil || descriptor.Info == nil || descriptor.Info.Type != field.TypeBytes {
		t.Fatalf("%s must be a bytes field", name)
	}
	if descriptor.Size != sha256.Size {
		t.Fatalf("%s max size = %d, want %d", name, descriptor.Size, sha256.Size)
	}
	for _, size := range []int{sha256.Size - 1, sha256.Size, sha256.Size + 1} {
		if got, want := mailBridgeByteValidatorsAccept(descriptor.Validators, make([]byte, size)), size == sha256.Size; got != want {
			t.Fatalf("%s accepts %d-byte value = %t, want %t", name, size, got, want)
		}
	}
}

func mailBridgeAssertStringMax(t *testing.T, name string, descriptor *field.Descriptor, max int, optional bool) {
	t.Helper()
	if descriptor == nil || descriptor.Info == nil || descriptor.Info.Type != field.TypeString {
		t.Fatalf("%s must be a string field", name)
	}
	if descriptor.Size != max {
		t.Fatalf("%s max size = %d, want %d", name, descriptor.Size, max)
	}
	if descriptor.Optional != optional || descriptor.Nillable != optional {
		t.Fatalf("%s optional/nillable = %t/%t, want %t/%t", name, descriptor.Optional, descriptor.Nillable, optional, optional)
	}
}

func mailBridgeAssertIntMinimum(t *testing.T, name string, descriptor *field.Descriptor, minimum int) {
	t.Helper()
	if descriptor == nil || descriptor.Info == nil || descriptor.Info.Type != field.TypeInt {
		t.Fatalf("%s must be an int field", name)
	}
	for _, value := range []int{minimum - 1, minimum} {
		if got, want := mailBridgeIntValidatorsAccept(descriptor.Validators, value), value >= minimum; got != want {
			t.Fatalf("%s accepts %d = %t, want %t", name, value, got, want)
		}
	}
}

func mailBridgeAssertIntBounds(t *testing.T, name string, descriptor *field.Descriptor, minimum, maximum int) {
	t.Helper()
	if descriptor == nil || descriptor.Info == nil || descriptor.Info.Type != field.TypeInt {
		t.Fatalf("%s must be an int field", name)
	}
	for _, value := range []int{minimum - 1, minimum, maximum, maximum + 1} {
		if got, want := mailBridgeIntValidatorsAccept(descriptor.Validators, value), value >= minimum && value <= maximum; got != want {
			t.Fatalf("%s accepts %d = %t, want %t", name, value, got, want)
		}
	}
}

func mailBridgeAssertInt64Bounds(t *testing.T, name string, descriptor *field.Descriptor, minimum, maximum int64) {
	t.Helper()
	if descriptor == nil || descriptor.Info == nil || descriptor.Info.Type != field.TypeInt64 {
		t.Fatalf("%s must be an int64 field", name)
	}
	for _, value := range []int64{minimum - 1, minimum, maximum, maximum + 1} {
		valid := true
		for _, rawValidator := range descriptor.Validators {
			if validator, ok := rawValidator.(func(int64) error); ok && validator(value) != nil {
				valid = false
			}
		}
		if want := value >= minimum && value <= maximum; valid != want {
			t.Fatalf("%s accepts %d = %t, want %t", name, value, valid, want)
		}
	}
}

func mailBridgeAssertDefault(t *testing.T, name string, descriptor *field.Descriptor, want any) {
	t.Helper()
	if descriptor == nil || !reflect.DeepEqual(descriptor.Default, want) {
		t.Fatalf("%s default = %#v, want %#v", name, descriptor.Default, want)
	}
}

func mailBridgeAssertFieldType(t *testing.T, name string, descriptor *field.Descriptor, want field.Type) {
	t.Helper()
	if descriptor == nil || descriptor.Info == nil || descriptor.Info.Type != want {
		t.Fatalf("%s type = %#v, want %v", name, descriptor, want)
	}
}

func mailBridgeAssertUint8Minimum(t *testing.T, name string, descriptor *field.Descriptor, minimum uint8) {
	t.Helper()
	if descriptor == nil || descriptor.Info == nil || descriptor.Info.Type != field.TypeUint8 {
		t.Fatalf("%s must be a uint8 field", name)
	}
	for _, value := range []uint8{minimum - 1, minimum} {
		if got, want := mailBridgeUint8ValidatorsAccept(descriptor.Validators, value), value >= minimum; got != want {
			t.Fatalf("%s accepts %d = %t, want %t", name, value, got, want)
		}
	}
}

func mailBridgeAssertEnumValues(t *testing.T, name string, descriptor *field.Descriptor, want []string) {
	t.Helper()
	if descriptor == nil {
		t.Fatalf("missing %s", name)
	}
	got := make([]string, 0, len(descriptor.Enums))
	for _, enum := range descriptor.Enums {
		got = append(got, enum.V)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s enum values = %v, want %v", name, got, want)
	}
}

func mailBridgeAssertOptionalNillable(t *testing.T, name string, descriptor *field.Descriptor) {
	t.Helper()
	if descriptor == nil || !descriptor.Optional || !descriptor.Nillable {
		t.Fatalf("%s must be optional and nillable", name)
	}
}

func mailBridgeAssertImmutable(t *testing.T, name string, descriptor *field.Descriptor) {
	t.Helper()
	if descriptor == nil || !descriptor.Immutable {
		t.Fatalf("%s must be immutable", name)
	}
}

func mailBridgeAssertChecks(t *testing.T, annotations []schema.Annotation, want map[string]string) {
	t.Helper()
	got := map[string]string{}
	for _, annotation := range annotations {
		if sqlAnnotation, ok := annotation.(*entsql.Annotation); ok {
			for name, predicate := range sqlAnnotation.Checks {
				got[name] = predicate
			}
		}
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("named CHECKs = %#v, want %#v", got, want)
	}
}

func mailBridgeAssertUniqueIndex(t *testing.T, indexes []ent.Index, wantFields []string) {
	t.Helper()
	for _, schemaIndex := range indexes {
		descriptor := schemaIndex.Descriptor()
		if reflect.DeepEqual(descriptor.Fields, wantFields) {
			if !descriptor.Unique {
				t.Fatalf("index over %v must be unique", wantFields)
			}
			return
		}
	}
	t.Fatalf("missing unique index over %v", wantFields)
}

func mailBridgeAssertUniquePartialIndex(t *testing.T, indexes []ent.Index, wantFields []string, wantWhere string) {
	t.Helper()
	for _, schemaIndex := range indexes {
		descriptor := schemaIndex.Descriptor()
		if !reflect.DeepEqual(descriptor.Fields, wantFields) {
			continue
		}
		if !descriptor.Unique {
			t.Fatalf("partial index over %v must be unique", wantFields)
		}
		for _, annotation := range descriptor.Annotations {
			if sqlAnnotation, ok := annotation.(*entsql.Annotation); ok && sqlAnnotation.IndexWhere == wantWhere {
				return
			}
		}
		t.Fatalf("partial index over %v missing WHERE %q", wantFields, wantWhere)
	}
	t.Fatalf("missing partial unique index over %v", wantFields)
}

func mailBridgeAssertPartialIndex(t *testing.T, indexes []ent.Index, wantFields []string, wantWhere string) {
	t.Helper()
	for _, schemaIndex := range indexes {
		descriptor := schemaIndex.Descriptor()
		if !reflect.DeepEqual(descriptor.Fields, wantFields) {
			continue
		}
		for _, annotation := range descriptor.Annotations {
			if sqlAnnotation, ok := annotation.(*entsql.Annotation); ok && sqlAnnotation.IndexWhere == wantWhere {
				return
			}
		}
		t.Fatalf("partial index over %v missing WHERE %q", wantFields, wantWhere)
	}
	t.Fatalf("missing partial index over %v", wantFields)
}

func mailBridgeAssertNonUniquePartialIndex(t *testing.T, indexes []ent.Index, wantFields []string, wantStorageKey, wantWhere string) {
	t.Helper()
	for _, schemaIndex := range indexes {
		descriptor := schemaIndex.Descriptor()
		if !reflect.DeepEqual(descriptor.Fields, wantFields) {
			continue
		}
		if descriptor.Unique {
			t.Fatalf("partial index over %v must be non-unique", wantFields)
		}
		if descriptor.StorageKey != wantStorageKey {
			t.Fatalf("partial index over %v storage key = %q, want %q", wantFields, descriptor.StorageKey, wantStorageKey)
		}
		for _, annotation := range descriptor.Annotations {
			if sqlAnnotation, ok := annotation.(*entsql.Annotation); ok && sqlAnnotation.IndexWhere == wantWhere {
				return
			}
		}
		t.Fatalf("partial index over %v missing WHERE %q", wantFields, wantWhere)
	}
	t.Fatalf("missing non-unique partial index over %v", wantFields)
}

func mailBridgeAssertEdge(t *testing.T, name string, descriptor *edge.Descriptor, inverse, unique, required bool, fieldName, refName string, onDelete any) {
	t.Helper()
	if descriptor == nil {
		t.Fatalf("missing %s edge", name)
	}
	if descriptor.Inverse != inverse || descriptor.Unique != unique || descriptor.Required != required || descriptor.Field != fieldName || descriptor.RefName != refName {
		t.Fatalf("%s edge = inverse:%t unique:%t required:%t field:%q ref:%q", name, descriptor.Inverse, descriptor.Unique, descriptor.Required, descriptor.Field, descriptor.RefName)
	}
	mailBridgeAssertOnDelete(t, name, descriptor.Annotations, onDelete)
}

func mailBridgeAssertExistingInverseEdge(t *testing.T, edges []ent.Edge, name, refName string) {
	t.Helper()
	for _, schemaEdge := range edges {
		descriptor := schemaEdge.Descriptor()
		if descriptor.Name != name {
			continue
		}
		if !descriptor.Inverse || descriptor.RefName != refName {
			t.Fatalf("%s = inverse:%t ref:%q, want true/%q", name, descriptor.Inverse, descriptor.RefName, refName)
		}
		return
	}
	t.Fatalf("missing %s edge", name)
}

func mailBridgeAssertOnDelete(t *testing.T, name string, annotations []schema.Annotation, want any) {
	t.Helper()
	for _, annotation := range annotations {
		if sqlAnnotation, ok := annotation.(*entsql.Annotation); ok && reflect.DeepEqual(sqlAnnotation.OnDelete, want) {
			return
		}
	}
	t.Fatalf("%s OnDelete = missing, want %q", name, want)
}

func mailBridgeByteValidatorsAccept(validators []interface{}, value []byte) bool {
	for _, rawValidator := range validators {
		if validator, ok := rawValidator.(func([]byte) error); ok {
			if err := validator(value); err != nil {
				return false
			}
		}
	}
	return true
}

func mailBridgeIntValidatorsAccept(validators []interface{}, value int) bool {
	for _, rawValidator := range validators {
		if validator, ok := rawValidator.(func(int) error); ok {
			if err := validator(value); err != nil {
				return false
			}
		}
	}
	return true
}

func mailBridgeUint8ValidatorsAccept(validators []interface{}, value uint8) bool {
	for _, rawValidator := range validators {
		if validator, ok := rawValidator.(func(uint8) error); ok {
			if err := validator(value); err != nil {
				return false
			}
		}
	}
	return true
}

func mailBridgeAssertExactNames(t *testing.T, label string, got, want []string) {
	t.Helper()
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}

func mailBridgeMapKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}
