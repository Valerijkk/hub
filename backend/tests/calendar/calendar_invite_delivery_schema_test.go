package calendar_test

import (
	"os"
	"reflect"
	"strings"
	"testing"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestCalendarInviteDeliverySchemaContract(t *testing.T) {
	fields := mailBridgeFieldDescriptors((entschema.CalendarInviteDelivery{}).Fields())
	mailBridgeAssertExactNames(t, "CalendarInviteDelivery fields", mailBridgeMapKeys(fields), []string{
		"attempts", "calendar_attendee_id", "calendar_event_id", "claim_generation", "created_at", "event_sequence", "id", "kind",
		"last_error_code", "lease_token", "lease_until", "mail_submission_id", "next_attempt_at", "organization_id", "recurrence_id",
		"source_envelope_id", "state", "terminal_at", "updated_at",
	})
	for _, name := range []string{"id", "organization_id", "calendar_event_id", "calendar_attendee_id", "source_envelope_id"} {
		mailBridgeAssertRequiredUUID(t, name, fields[name])
		mailBridgeAssertImmutable(t, name, fields[name])
	}
	for _, name := range []string{"kind", "event_sequence", "recurrence_id"} {
		mailBridgeAssertImmutable(t, name, fields[name])
	}
	mailBridgeAssertIntMinimum(t, "event_sequence", fields["event_sequence"], 0)
	calendarInviteAssertIntBounds(t, "attempts", fields["attempts"], 0, 8)
	calendarInviteAssertInt64Bounds(t, "claim_generation", fields["claim_generation"], 0, 2147483647)
	mailBridgeAssertStringMax(t, "recurrence_id", fields["recurrence_id"], 255, true)
	mailBridgeAssertEnumValues(t, "kind", fields["kind"], []string{"created", "updated", "cancelled"})
	mailBridgeAssertEnumValues(t, "state", fields["state"], []string{"pending", "submitting", "unknown", "sent", "superseded", "failed_permanent"})
	mailBridgeAssertEnumValues(t, "last_error_code", fields["last_error_code"], []string{
		"mailbox_unavailable", "engine_unavailable", "submission_rejected", "submission_outcome_unknown", "invalid_recipient", "materialization_invalid", "policy_denied", "retry_exhausted", "internal_error",
	})
	for _, name := range []string{"mail_submission_id", "lease_token", "lease_until", "next_attempt_at", "last_error_code", "terminal_at"} {
		mailBridgeAssertOptionalNillable(t, name, fields[name])
	}
	mailBridgeAssertImmutable(t, "created_at", fields["created_at"])

	mailBridgeAssertChecks(t, (entschema.CalendarInviteDelivery{}).Annotations(), map[string]string{
		"calendar_invite_deliveries_event_sequence_non_negative": "event_sequence >= 0",
		"calendar_invite_deliveries_attempts_bounds":        "attempts >= 0 AND attempts <= 8",
		"calendar_invite_deliveries_lease_pair":              "COALESCE((lease_token IS NULL) = (lease_until IS NULL), FALSE)",
		"calendar_invite_deliveries_claim_generation_bounds": "claim_generation >= 0 AND claim_generation <= 2147483647",
		"calendar_invite_deliveries_state_shape": "COALESCE((state = 'pending' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND terminal_at IS NULL AND next_attempt_at IS NOT NULL AND ((attempts = 0 AND last_error_code IS NULL) OR (attempts BETWEEN 1 AND 8 AND last_error_code IN ('mailbox_unavailable','engine_unavailable','internal_error')))) OR (state = 'submitting' AND lease_token IS NOT NULL AND lease_until IS NOT NULL AND claim_generation >= 1 AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NULL) OR (state = 'unknown' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NOT NULL AND terminal_at IS NULL AND last_error_code IN ('mailbox_unavailable','engine_unavailable','submission_outcome_unknown')) OR (state = 'sent' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NOT NULL AND attempts BETWEEN 1 AND 8 AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NOT NULL) OR (state = 'superseded' AND lease_token IS NULL AND lease_until IS NULL AND mail_submission_id IS NULL AND next_attempt_at IS NULL AND last_error_code IS NULL AND terminal_at IS NOT NULL) OR (state = 'failed_permanent' AND lease_token IS NULL AND lease_until IS NULL AND next_attempt_at IS NULL AND last_error_code IN ('submission_rejected','invalid_recipient','materialization_invalid','policy_denied','retry_exhausted') AND terminal_at IS NOT NULL), FALSE)",
	})
	mailBridgeAssertUniqueIndex(t, (entschema.CalendarInviteDelivery{}).Indexes(), []string{"source_envelope_id", "calendar_attendee_id"})
	mailBridgeAssertUniquePartialIndex(t, (entschema.CalendarInviteDelivery{}).Indexes(), []string{"calendar_event_id", "calendar_attendee_id", "event_sequence", "kind"}, "recurrence_id IS NULL")
	mailBridgeAssertUniquePartialIndex(t, (entschema.CalendarInviteDelivery{}).Indexes(), []string{"calendar_event_id", "calendar_attendee_id", "event_sequence", "kind", "recurrence_id"}, "recurrence_id IS NOT NULL")
	calendarInviteAssertPartialIndex(t, (entschema.CalendarInviteDelivery{}).Indexes(), []string{"next_attempt_at", "id"}, "state IN ('pending','unknown') AND next_attempt_at IS NOT NULL")
	mailBridgeAssertNonUniquePartialIndex(t, (entschema.CalendarInviteDelivery{}).Indexes(), []string{"lease_until", "id"}, "calendar_invite_deliveries_stale_lease_idx", "state = 'submitting' AND lease_until IS NOT NULL")

	edges := mailBridgeEdgeDescriptors((entschema.CalendarInviteDelivery{}).Edges())
	mailBridgeAssertExactNames(t, "CalendarInviteDelivery edges", mailBridgeMapKeys(edges), []string{"attendee", "event", "mail_submission", "organization"})
	mailBridgeAssertEdge(t, "organization", edges["organization"], false, true, true, "organization_id", "", entsql.NoAction)
	mailBridgeAssertEdge(t, "event", edges["event"], false, true, true, "calendar_event_id", "", entsql.NoAction)
	mailBridgeAssertEdge(t, "attendee", edges["attendee"], false, true, true, "calendar_attendee_id", "", entsql.NoAction)
	mailBridgeAssertEdge(t, "mail_submission", edges["mail_submission"], false, true, false, "mail_submission_id", "", entsql.NoAction)

	source, err := os.ReadFile("../../ent/schema/calendarinvitedelivery.go")
	if err != nil {
		t.Fatalf("read CalendarInviteDelivery schema: %v", err)
	}
	for _, forbidden := range []string{
		"itip", "field.string(\"body\")", "field.bytes(\"body\")", "field.string(\"title\")", "field.string(\"description\")",
		"field.string(\"normalized_email\")", "field.string(\"address\")", "provider_", "engine_version", "field.string(\"last_error\")", "result_mail_message_id", "edge.to(\"source_envelope\"",
	} {
		if strings.Contains(strings.ToLower(string(source)), forbidden) {
			t.Fatalf("CalendarInviteDelivery must not persist %q", forbidden)
		}
	}
}

func calendarInviteAssertInt64Bounds(t *testing.T, name string, descriptor *field.Descriptor, minimum, maximum int64) {
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

func calendarInviteAssertIntBounds(t *testing.T, name string, descriptor *field.Descriptor, minimum, maximum int) {
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

func calendarInviteAssertPartialIndex(t *testing.T, indexes []ent.Index, wantFields []string, wantWhere string) {
	t.Helper()
	for _, schemaIndex := range indexes {
		descriptor := schemaIndex.Descriptor()
		if !reflect.DeepEqual(descriptor.Fields, wantFields) {
			continue
		}
		for _, annotation := range descriptor.Annotations {
			if sqlAnnotation, ok := annotation.(*entsql.IndexAnnotation); ok && sqlAnnotation.Where == wantWhere {
				return
			}
		}
		t.Fatalf("partial index over %v missing WHERE %q", wantFields, wantWhere)
	}
	t.Fatalf("missing partial index over %v", wantFields)
}
