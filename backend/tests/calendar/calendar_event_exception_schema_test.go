package calendar_test

import (
	"testing"

	"entgo.io/ent/dialect/entsql"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestCalendarEventException_PayloadHashColumnsArePaired(t *testing.T) {
	const (
		checkName = "calendar_event_exceptions_payload_hash_pair"
		predicate = "(payload_hash_version IS NULL) = (payload_hash IS NULL)"
	)

	for _, annotation := range (entschema.CalendarEventException{}).Annotations() {
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
