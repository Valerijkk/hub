package calendar_test

import (
	"testing"

	"entgo.io/ent/dialect/entsql"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestCalendarEvent_PayloadHashColumnsArePaired(t *testing.T) {
	const (
		checkName = "calendar_events_payload_hash_pair"
		predicate = "(payload_hash_version IS NULL) = (payload_hash IS NULL)"
	)

	for _, annotation := range (entschema.CalendarEvent{}).Annotations() {
		sqlAnnotation, ok := annotation.(*entsql.Annotation)
		if !ok {
			continue
		}
		if got, found := sqlAnnotation.Checks[checkName]; found {
			if got != predicate {
				t.Fatalf("%s = %q, want %q", checkName, got, predicate)
			}
			return
		}
	}

	t.Fatalf("missing named CHECK %s", checkName)
}

func TestCalendarEvent_MailMessagesInverseIsPresent(t *testing.T) {
	for _, schemaEdge := range (entschema.CalendarEvent{}).Edges() {
		descriptor := schemaEdge.Descriptor()
		if descriptor.Name != "mail_messages" {
			continue
		}
		if descriptor.Inverse || descriptor.Unique || descriptor.Required || descriptor.RefName != "" {
			t.Fatalf("CalendarEvent.mail_messages = %#v, want plain to-many inverse target", descriptor)
		}
		return
	}

	t.Fatal("CalendarEvent must expose mail_messages")
}
