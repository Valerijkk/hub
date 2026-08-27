package calendar_test

import (
	"crypto/sha256"
	"reflect"
	"sort"
	"testing"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

const calendarEventResponseMaxSerializedUTF8Bytes = 1_048_576

func TestCalendarMutationReceipt_HasExactFieldsAndEnumValues(t *testing.T) {
	receipt := entschema.CalendarMutationReceipt{}
	fields := map[string]*field.Descriptor{}
	for _, receiptField := range receipt.Fields() {
		descriptor := receiptField.Descriptor()
		fields[descriptor.Name] = descriptor
	}

	wantFieldNames := []string{
		"actor_user_id",
		"calendar_event_id",
		"created_at",
		"expires_at",
		"id",
		"idempotency_key_hash",
		"operation",
		"organization_id",
		"request_hash",
		"response_json",
		"response_status",
		"state",
		"updated_at",
	}
	assertExactNames(t, "CalendarMutationReceipt fields", mapsKeys(fields), wantFieldNames)

	assertExactEnumValues(t, fields["operation"], []string{"event_create", "legacy_call_create", "rsvp_series", "rsvp_occurrence"})
	assertExactEnumValues(t, fields["state"], []string{"in_progress", "completed"})

	for _, uuidFieldName := range []string{"id", "organization_id", "actor_user_id"} {
		assertRequiredUUIDField(t, uuidFieldName, fields[uuidFieldName])
	}
	for _, hashFieldName := range []string{"idempotency_key_hash", "request_hash"} {
		assertExactSHA256Field(t, hashFieldName, fields[hashFieldName])
	}
	assertByteFieldMaxSize(t, "response_json", fields["response_json"], calendarEventResponseMaxSerializedUTF8Bytes)
	assertIntFieldBounds(t, "response_status", fields["response_status"], 100, 599)
	assertResponseJSONValidation(t, fields["response_json"])

	for _, fieldName := range []string{"response_status", "response_json", "calendar_event_id"} {
		if descriptor := fields[fieldName]; !descriptor.Optional || !descriptor.Nillable {
			t.Fatalf("%s optional/nillable = %t/%t, want true/true", fieldName, descriptor.Optional, descriptor.Nillable)
		}
	}
	if descriptor := fields["expires_at"]; descriptor.Optional || descriptor.Nillable {
		t.Fatalf("expires_at optional/nillable = %t/%t, want false/false", descriptor.Optional, descriptor.Nillable)
	}
	if descriptor := fields["created_at"]; !descriptor.Immutable {
		t.Fatal("created_at must be immutable")
	}
	for _, fieldName := range []string{"operation", "idempotency_key_hash", "request_hash"} {
		if descriptor := fields[fieldName]; !descriptor.Immutable {
			t.Fatalf("%s must be immutable", fieldName)
		}
	}

	for _, forbidden := range []string{"idempotency_key", "raw_idempotency_key"} {
		if _, found := fields[forbidden]; found {
			t.Fatalf("must not persist raw idempotency key field %q", forbidden)
		}
	}
}

func TestCalendarMutationReceipt_HasNamedDatabaseChecks(t *testing.T) {
	wantChecks := map[string]string{
		"calendar_mutation_receipts_idempotency_key_hash_length": "octet_length(idempotency_key_hash) = 32",
		"calendar_mutation_receipts_request_hash_length":         "octet_length(request_hash) = 32",
		"calendar_mutation_receipts_response_json_length":        "octet_length(response_json) <= 1048576",
		"calendar_mutation_receipts_response_status_range":       "response_status IS NULL OR (response_status >= 100 AND response_status <= 599)",
	}

	gotChecks := map[string]string{}
	for _, annotation := range (entschema.CalendarMutationReceipt{}).Annotations() {
		sqlAnnotation, ok := annotation.(*entsql.Annotation)
		if !ok {
			continue
		}
		for name, predicate := range sqlAnnotation.Checks {
			gotChecks[name] = predicate
		}
	}

	if !reflect.DeepEqual(gotChecks, wantChecks) {
		t.Fatalf("named CHECKs = %#v, want %#v", gotChecks, wantChecks)
	}
}

func TestCalendarMutationReceipt_UniqueIndexScopesHashedIdempotencyKey(t *testing.T) {
	wantFields := []string{"organization_id", "actor_user_id", "operation", "idempotency_key_hash"}
	for _, receiptIndex := range (entschema.CalendarMutationReceipt{}).Indexes() {
		descriptor := receiptIndex.Descriptor()
		if reflect.DeepEqual(descriptor.Fields, wantFields) {
			if !descriptor.Unique {
				t.Fatal("receipt idempotency scope index must be unique")
			}
			return
		}
	}

	t.Fatalf("missing unique receipt idempotency index over %v", wantFields)
}

func TestCalendarMutationReceipt_EdgesEncodeActorOwnershipAndOptionalEventReference(t *testing.T) {
	edges := edgeDescriptors((entschema.CalendarMutationReceipt{}).Edges())
	assertExactNames(t, "CalendarMutationReceipt edges", mapsKeys(edges), []string{"actor", "event", "organization"})

	organization := edges["organization"]
	if organization.Inverse || !organization.Unique || !organization.Required || organization.Field != "organization_id" {
		t.Fatalf("organization edge = inverse:%t unique:%t required:%t field:%q, want false/true/true/%q", organization.Inverse, organization.Unique, organization.Required, organization.Field, "organization_id")
	}
	assertOnDelete(t, "organization", organization.Annotations, entsql.NoAction)

	actor := edges["actor"]
	if !actor.Inverse || !actor.Unique || !actor.Required || actor.Field != "actor_user_id" || actor.RefName != "calendar_mutation_receipts" {
		t.Fatalf("actor edge = inverse:%t unique:%t required:%t field:%q ref:%q, want true/true/true/%q/%q", actor.Inverse, actor.Unique, actor.Required, actor.Field, actor.RefName, "actor_user_id", "calendar_mutation_receipts")
	}
	assertOnDelete(t, "actor", actor.Annotations, entsql.Cascade)

	event := edges["event"]
	if !event.Inverse || !event.Unique || event.Required || event.Field != "calendar_event_id" || event.RefName != "mutation_receipts" {
		t.Fatalf("event edge = inverse:%t unique:%t required:%t field:%q ref:%q, want true/true/false/%q/%q", event.Inverse, event.Unique, event.Required, event.Field, event.RefName, "calendar_event_id", "mutation_receipts")
	}
	assertOnDelete(t, "event", event.Annotations, entsql.SetNull)
}

func TestUserAndScheduledCall_PreserveExistingEdgesAndAddOnlyCalendarInverses(t *testing.T) {
	assertExactNames(t, "User edges", edgeNames((entschema.User{}).Edges()), []string{
		"activity_logs", "ai_preference", "ai_providers", "ai_usage_daily", "api_keys",
		"automation_rules", "blocks_created", "blocks_received", "bookmarks", "calendar_attendances",
		"calendar_events_owned", "calendar_mutation_receipts", "call_participations", "channel_invites",
		"channel_memberships", "coffee_enrollment", "coffee_pairs_a", "coffee_pairs_b", "coffee_rounds_triggered",
		"comments", "created_calls", "created_channels", "created_runs", "created_test_cases", "crm_contact_link",
		"crm_contacts_created", "crm_deals_owned", "dashboard_layouts", "executed_results", "files", "hub_bases",
		"hub_records", "invoices_created", "invites", "issue_message_links", "journal_entries", "kudos_received",
		"kudos_sent", "leave_requests", "leave_reviews", "marketplace_listings", "marketplace_reviews", "news_poll_votes",
		"news_posts", "notification_preferences", "notifications", "organization", "organized_calendar_events", "poll_votes",
		"polls", "project_members", "project_templates", "push_subscriptions", "reactions", "recordings", "saved_filters",
		"scheduled_calls", "sessions", "team_members", "test_case_versions", "time_logs", "mail_objects",
	})
	assertExactNames(t, "ScheduledCall edges", edgeNames((entschema.ScheduledCall{}).Edges()), []string{"calendar_event_link", "calendar_event_reference", "channel", "creator", "participants", "project"})

	userEdges := edgeDescriptors((entschema.User{}).Edges())
	for _, name := range []string{"calendar_events_owned", "organized_calendar_events", "calendar_attendances"} {
		if !userEdges[name].Inverse {
			t.Fatalf("User.%s must be an inverse calendar edge", name)
		}
	}
	if userEdges["calendar_mutation_receipts"].Inverse {
		t.Fatal("User.calendar_mutation_receipts must own the actor receipt edge")
	}

	callEdges := edgeDescriptors((entschema.ScheduledCall{}).Edges())
	callEdge := callEdges["calendar_event_link"]
	if callEdge == nil {
		t.Fatal("missing ScheduledCall.calendar_event_link")
	}
	if callEdge.Inverse || !callEdge.Unique || callEdge.Required || callEdge.Field != "" || callEdge.RefName != "" {
		t.Fatalf("ScheduledCall.calendar_event_link = inverse:%t unique:%t required:%t field:%q ref:%q, want false/true/false/empty/empty", callEdge.Inverse, callEdge.Unique, callEdge.Required, callEdge.Field, callEdge.RefName)
	}
	callReferenceEdge := callEdges["calendar_event_reference"]
	if callReferenceEdge == nil {
		t.Fatal("missing ScheduledCall.calendar_event_reference")
	}
	if !callReferenceEdge.Inverse || !callReferenceEdge.Unique || callReferenceEdge.Required || !callReferenceEdge.Immutable || callReferenceEdge.Field != "calendar_event_id" || callReferenceEdge.RefName != "scheduled_call_reference" {
		t.Fatalf("ScheduledCall.calendar_event_reference = inverse:%t unique:%t required:%t immutable:%t field:%q ref:%q, want true/true/false/true/%q/%q", callReferenceEdge.Inverse, callReferenceEdge.Unique, callReferenceEdge.Required, callReferenceEdge.Immutable, callReferenceEdge.Field, callReferenceEdge.RefName, "calendar_event_id", "scheduled_call_reference")
	}
	assertOnDelete(t, "calendar_event_reference", callReferenceEdge.Annotations, entsql.SetNull)

	referenceEdge := edgeDescriptors((entschema.CalendarEvent{}).Edges())["scheduled_call_reference"]
	if referenceEdge == nil {
		t.Fatal("missing CalendarEvent.scheduled_call_reference")
	}
	if referenceEdge.Inverse || !referenceEdge.Unique || referenceEdge.Required || referenceEdge.Field != "" || referenceEdge.RefName != "" {
		t.Fatalf("CalendarEvent.scheduled_call_reference = inverse:%t unique:%t required:%t field:%q ref:%q, want false/true/false/empty/empty", referenceEdge.Inverse, referenceEdge.Unique, referenceEdge.Required, referenceEdge.Field, referenceEdge.RefName)
	}
}

func TestCalendarEvent_ScheduledCallLinkRemainsOptionalImmutableOneToOne(t *testing.T) {
	fields := map[string]*field.Descriptor{}
	for _, eventField := range (entschema.CalendarEvent{}).Fields() {
		descriptor := eventField.Descriptor()
		fields[descriptor.Name] = descriptor
	}
	callField := fields["scheduled_call_id"]
	if callField == nil {
		t.Fatal("missing scheduled_call_id")
	}
	if !callField.Optional || !callField.Nillable || !callField.Unique || !callField.Immutable {
		t.Fatalf("scheduled_call_id = optional:%t nillable:%t unique:%t immutable:%t, want true/true/true/true", callField.Optional, callField.Nillable, callField.Unique, callField.Immutable)
	}

	callEdge := edgeDescriptors((entschema.CalendarEvent{}).Edges())["scheduled_call"]
	if callEdge == nil {
		t.Fatal("missing CalendarEvent.scheduled_call")
	}
	if !callEdge.Inverse || !callEdge.Unique || !callEdge.Immutable || callEdge.Field != "scheduled_call_id" || callEdge.RefName != "calendar_event_link" {
		t.Fatalf("CalendarEvent.scheduled_call = inverse:%t unique:%t immutable:%t field:%q ref:%q, want true/true/true/%q/%q", callEdge.Inverse, callEdge.Unique, callEdge.Immutable, callEdge.Field, callEdge.RefName, "scheduled_call_id", "calendar_event_link")
	}
}

func assertExactSHA256Field(t *testing.T, name string, descriptor *field.Descriptor) {
	t.Helper()
	if descriptor == nil {
		t.Fatalf("missing %s field", name)
	}
	if descriptor.Optional || descriptor.Nillable {
		t.Fatalf("%s optional/nillable = %t/%t, want false/false", name, descriptor.Optional, descriptor.Nillable)
	}
	assertByteFieldBounds(t, name, descriptor, sha256.Size, sha256.Size)
}

func assertRequiredUUIDField(t *testing.T, name string, descriptor *field.Descriptor) {
	t.Helper()
	if descriptor == nil {
		t.Fatalf("missing %s field", name)
	}
	if descriptor.Info == nil || descriptor.Info.Type != field.TypeUUID {
		t.Fatalf("%s type = %#v, want UUID", name, descriptor.Info)
	}
	if descriptor.Optional || descriptor.Nillable {
		t.Fatalf("%s optional/nillable = %t/%t, want false/false", name, descriptor.Optional, descriptor.Nillable)
	}
	if name == "id" && !descriptor.Immutable {
		t.Fatal("id must be immutable")
	}
}

func assertResponseJSONValidation(t *testing.T, descriptor *field.Descriptor) {
	t.Helper()
	for _, testCase := range []struct {
		name  string
		value []byte
		want  bool
	}{
		{name: "small object", value: []byte(`{"ok":true}`), want: true},
		{name: "valid multibyte utf8 string", value: []byte(`"Привет"`), want: true},
		{name: "truncated object", value: []byte(`{`), want: false},
		{name: "invalid utf8 json string", value: []byte{'"', 0xff, '"'}, want: false},
		{name: "exact maximum", value: validJSONString(calendarEventResponseMaxSerializedUTF8Bytes), want: true},
		{name: "oversize", value: validJSONString(calendarEventResponseMaxSerializedUTF8Bytes + 1), want: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if got := byteValidatorsAccept(descriptor.Validators, testCase.value); got != testCase.want {
				t.Fatalf("response_json accepted = %t, want %t", got, testCase.want)
			}
		})
	}
}

func validJSONString(size int) []byte {
	value := make([]byte, size)
	value[0] = '"'
	for index := 1; index < size-1; index++ {
		value[index] = 'a'
	}
	value[size-1] = '"'
	return value
}

func assertByteFieldBounds(t *testing.T, name string, descriptor *field.Descriptor, min, max int) {
	t.Helper()
	if descriptor == nil {
		t.Fatalf("missing %s field", name)
	}
	if descriptor.Info == nil || descriptor.Info.Type != field.TypeBytes {
		t.Fatalf("%s type = %#v, want bytes", name, descriptor.Info)
	}
	if got := descriptor.Size; got != max {
		t.Fatalf("%s max size = %d, want %d", name, got, max)
	}
	for _, length := range []int{min - 1, min, max, max + 1} {
		if length < 0 {
			continue
		}
		wantAccepted := length >= min && length <= max
		gotAccepted := byteValidatorsAccept(descriptor.Validators, make([]byte, length))
		if gotAccepted != wantAccepted {
			t.Fatalf("%s accepts %d-byte value = %t, want %t", name, length, gotAccepted, wantAccepted)
		}
	}
}

func assertByteFieldMaxSize(t *testing.T, name string, descriptor *field.Descriptor, max int) {
	t.Helper()
	if descriptor == nil {
		t.Fatalf("missing %s field", name)
	}
	if descriptor.Info == nil || descriptor.Info.Type != field.TypeBytes {
		t.Fatalf("%s type = %#v, want bytes", name, descriptor.Info)
	}
	if got := descriptor.Size; got != max {
		t.Fatalf("%s max size = %d, want %d", name, got, max)
	}
}

func assertIntFieldBounds(t *testing.T, name string, descriptor *field.Descriptor, min, max int) {
	t.Helper()
	if descriptor == nil {
		t.Fatalf("missing %s field", name)
	}
	for _, value := range []int{min - 1, min, max, max + 1} {
		wantAccepted := value >= min && value <= max
		gotAccepted := intValidatorsAccept(descriptor.Validators, value)
		if gotAccepted != wantAccepted {
			t.Fatalf("%s accepts %d = %t, want %t", name, value, gotAccepted, wantAccepted)
		}
	}
}

func assertExactEnumValues(t *testing.T, descriptor *field.Descriptor, want []string) {
	t.Helper()
	if descriptor == nil {
		t.Fatal("missing enum field")
	}
	got := make([]string, 0, len(descriptor.Enums))
	for _, value := range descriptor.Enums {
		got = append(got, value.V)
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("enum values = %v, want %v", got, want)
	}
}

func byteValidatorsAccept(validators []interface{}, value []byte) bool {
	for _, rawValidator := range validators {
		validator, ok := rawValidator.(func([]byte) error)
		if !ok {
			continue
		}
		if err := validator(value); err != nil {
			return false
		}
	}
	return true
}

func intValidatorsAccept(validators []interface{}, value int) bool {
	for _, rawValidator := range validators {
		validator, ok := rawValidator.(func(int) error)
		if !ok {
			continue
		}
		if err := validator(value); err != nil {
			return false
		}
	}
	return true
}

func edgeDescriptors(edges []ent.Edge) map[string]*edge.Descriptor {
	descriptors := map[string]*edge.Descriptor{}
	for _, schemaEdge := range edges {
		descriptor := schemaEdge.Descriptor()
		descriptors[descriptor.Name] = descriptor
	}
	return descriptors
}

func edgeNames(edges []ent.Edge) []string {
	return mapsKeys(edgeDescriptors(edges))
}

func assertOnDelete(t *testing.T, name string, annotations []schema.Annotation, want any) {
	t.Helper()
	for _, annotation := range annotations {
		sqlAnnotation, ok := annotation.(*entsql.Annotation)
		if ok && reflect.DeepEqual(sqlAnnotation.OnDelete, want) {
			return
		}
	}
	t.Fatalf("%s OnDelete = missing, want %q", name, want)
}

func assertExactNames(t *testing.T, label string, got, want []string) {
	t.Helper()
	sort.Strings(got)
	sort.Strings(want)
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("%s = %v, want %v", label, got, want)
	}
}

func mapsKeys[T any](values map[string]T) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	return keys
}
