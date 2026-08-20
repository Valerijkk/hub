package calendar_test

import (
	"reflect"
	"sort"
	"testing"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	entschema "github.com/valerii/hub/backend/ent/schema"
)

func TestCalendarAttendee_RemovedAtIsOptionalTimestamptzTombstone(t *testing.T) {
	schemaFields := (entschema.CalendarAttendee{}).Fields()
	fields := calendarAttendeeFieldDescriptors(schemaFields)
	gotNames := make([]string, 0, len(fields))
	for name := range fields {
		gotNames = append(gotNames, name)
	}
	sort.Strings(gotNames)
	wantNames := []string{
		"calendar_event_id", "created_at", "display_name", "id", "normalized_email", "recurrence_id",
		"removed_at", "reply_delivery_status", "responded_at", "role", "rsvp_status", "updated_at", "user_id",
	}
	if len(schemaFields) != len(wantNames) || len(fields) != len(wantNames) || !reflect.DeepEqual(gotNames, wantNames) {
		t.Fatalf("CalendarAttendee fields = %v, want %v", gotNames, wantNames)
	}

	removedAt := fields["removed_at"]
	if removedAt == nil || removedAt.Info == nil || removedAt.Info.Type != field.TypeTime {
		t.Fatalf("removed_at type = %#v, want timestamp", removedAt)
	}
	if !removedAt.Optional || !removedAt.Nillable {
		t.Fatalf("removed_at optional/nillable = %t/%t, want true/true", removedAt.Optional, removedAt.Nillable)
	}
	if removedAt.Default != nil || removedAt.UpdateDefault != nil || removedAt.Immutable {
		t.Fatalf("removed_at default/update default/immutable = %#v/%#v/%t, want nil/nil/false", removedAt.Default, removedAt.UpdateDefault, removedAt.Immutable)
	}
	if removedAt.Unique {
		t.Fatal("removed_at must not be unique")
	}
	if removedAt.StorageKey != "" || removedAt.SchemaType != nil || len(removedAt.Validators) != 0 || len(removedAt.Annotations) != 0 {
		t.Fatalf("removed_at adds schema behavior: storage key=%q schema type=%#v validators=%d annotations=%d", removedAt.StorageKey, removedAt.SchemaType, len(removedAt.Validators), len(removedAt.Annotations))
	}
}

func TestCalendarAttendee_LiveUniquenessIndexesAreExact(t *testing.T) {
	want := []calendarAttendeeIndex{
		{
			fields:     []string{"calendar_event_id", "normalized_email"},
			unique:     true,
			storageKey: "",
			where:      "recurrence_id IS NULL AND removed_at IS NULL",
		},
		{
			fields:     []string{"calendar_event_id", "normalized_email", "recurrence_id"},
			unique:     true,
			storageKey: "",
			where:      "recurrence_id IS NOT NULL AND removed_at IS NULL",
		},
	}
	indexes := (entschema.CalendarAttendee{}).Indexes()
	if len(indexes) != len(want) {
		t.Fatalf("CalendarAttendee index count = %d, want %d", len(indexes), len(want))
	}
	for position, expected := range want {
		descriptor := indexes[position].Descriptor()
		where, indexWhereAnnotations := calendarAttendeeIndexWhere(descriptor.Annotations)
		if !reflect.DeepEqual(descriptor.Fields, expected.fields) || descriptor.Unique != expected.unique || descriptor.StorageKey != expected.storageKey || where != expected.where || len(descriptor.Edges) != 0 || len(descriptor.Annotations) != 1 || indexWhereAnnotations != 1 {
			t.Fatalf("CalendarAttendee index[%d] = fields:%v edges:%v unique:%t storage key:%q where:%q annotations:%d index-where:%d, want fields:%v empty edges unique:%t storage key:%q where:%q and one IndexWhere annotation", position, descriptor.Fields, descriptor.Edges, descriptor.Unique, descriptor.StorageKey, where, len(descriptor.Annotations), indexWhereAnnotations, expected.fields, expected.unique, expected.storageKey, expected.where)
		}
	}
}

type calendarAttendeeIndex struct {
	fields     []string
	unique     bool
	storageKey string
	where      string
}

func calendarAttendeeFieldDescriptors(fields []ent.Field) map[string]*field.Descriptor {
	descriptors := make(map[string]*field.Descriptor, len(fields))
	for _, schemaField := range fields {
		descriptor := schemaField.Descriptor()
		descriptors[descriptor.Name] = descriptor
	}
	return descriptors
}

func calendarAttendeeIndexWhere(annotations []schema.Annotation) (string, int) {
	for _, annotation := range annotations {
		if sqlAnnotation, ok := annotation.(*entsql.IndexAnnotation); ok {
			return sqlAnnotation.Where, 1
		}
	}
	return "", 0
}
