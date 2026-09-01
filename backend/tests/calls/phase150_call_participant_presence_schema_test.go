//go:build hubtest

package calls_test

import (
	"reflect"
	"testing"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"

	callschema "github.com/valerii/hub/backend/ent/schema"
)

func TestPhase150CallParticipantPresenceSchemaContract(t *testing.T) {
	presence := callschema.CallParticipantPresence{}

	if mixins := presence.Mixin(); len(mixins) != 0 {
		t.Fatal("call participant presence must not declare mixins")
	}
	requirePhase150PresenceEdges(t, presence.Edges())

	fields := presence.Fields()
	wantNames := []string{
		"id",
		"call_id",
		"user_id",
		"connected_count",
		"presence_sequence",
		"created_at",
		"updated_at",
	}
	if len(fields) != len(wantNames) {
		t.Fatal("call participant presence must declare exactly the frozen fields")
	}

	for position, configuredField := range fields {
		descriptor := configuredField.Descriptor()
		if descriptor == nil {
			t.Fatal("call participant presence fields must expose descriptors")
		}
		if descriptor.Name != wantNames[position] {
			t.Fatal("call participant presence fields must retain the frozen order")
		}
		if descriptor.Err != nil {
			t.Fatal("call participant presence field descriptor must not report an error")
		}
		if descriptor.Unique {
			t.Fatal("call participant presence fields must not add scalar unique constraints")
		}

		switch descriptor.Name {
		case "id":
			requirePhase150PresenceGeneratedUUID(t, descriptor)
		case "call_id", "user_id":
			requirePhase150PresenceRequiredUUID(t, descriptor)
		case "connected_count", "presence_sequence":
			requirePhase150PresenceNonnegativeCount(t, descriptor)
		case "created_at":
			requirePhase150PresenceTimestamp(t, descriptor, true, false)
		case "updated_at":
			requirePhase150PresenceTimestamp(t, descriptor, false, true)
		}
	}

	for _, forbiddenName := range []string{
		"state",
		"epoch",
		"provider",
		"delivery",
		"delivery_id",
		"legacy_call_participant",
		"call_participant_id",
		"outbox",
		"outbox_id",
		"event_id",
		"public_event",
		"public_event_id",
	} {
		for _, configuredField := range fields {
			if configuredField.Descriptor().Name == forbiddenName {
				t.Fatal("call participant presence must not store provider, delivery, legacy, or outbox state")
			}
		}
	}

	requirePhase150PresenceIndexes(t, presence.Indexes())
}

func requirePhase150PresenceEdges(t *testing.T, edges []ent.Edge) {
	t.Helper()
	if len(edges) != 2 {
		t.Fatal("call participant presence must declare exactly call and user child-owned edges")
	}
	requirePhase150PresenceEdge(t, edges[0].Descriptor(), "call", "Call", "call_id", entsql.Cascade)
	requirePhase150PresenceEdge(t, edges[1].Descriptor(), "user", "User", "user_id", entsql.NoAction)
}

func requirePhase150PresenceEdge(
	t *testing.T,
	descriptor *edge.Descriptor,
	wantName string,
	wantType string,
	wantField string,
	wantOnDelete entsql.ReferenceOption,
) {
	t.Helper()
	if descriptor == nil {
		t.Fatal("call participant presence edge must expose a descriptor")
	}
	if descriptor.Name != wantName || descriptor.Type != wantType || descriptor.Field != wantField {
		t.Fatal("call participant presence edge identity must match the frozen contract")
	}
	if !descriptor.Required || !descriptor.Unique || !descriptor.Immutable {
		t.Fatal("call participant presence child-owned edges must be required, unique, and immutable")
	}
	if descriptor.Inverse || descriptor.Ref != nil || descriptor.RefName != "" || descriptor.StorageKey != nil || descriptor.Through != nil {
		t.Fatal("call participant presence child-owned edges must be non-inverse without a ref or storage key")
	}
	if len(descriptor.Annotations) != 1 {
		t.Fatal("call participant presence child-owned edge must have exactly one SQL on-delete annotation")
	}
	annotation, ok := descriptor.Annotations[0].(*entsql.Annotation)
	if !ok || !reflect.DeepEqual(annotation, &entsql.Annotation{OnDelete: wantOnDelete}) {
		t.Fatal("call participant presence child-owned edge must use only the frozen SQL on-delete annotation")
	}
}

func requirePhase150PresenceIndexes(t *testing.T, indexes []ent.Index) {
	t.Helper()
	if len(indexes) != 2 {
		t.Fatal("call participant presence must declare exactly call/user and user indexes")
	}
	requirePhase150PresenceIndex(t, indexes[0], true, "call_id", "user_id")
	requirePhase150PresenceIndex(t, indexes[1], false, "user_id")
}

func requirePhase150PresenceIndex(t *testing.T, configuredIndex ent.Index, wantUnique bool, wantFields ...string) {
	t.Helper()
	descriptor := configuredIndex.Descriptor()
	if descriptor == nil {
		t.Fatal("call participant presence index must expose a descriptor")
	}
	if descriptor.Unique != wantUnique {
		t.Fatal("call participant presence index uniqueness must match the frozen contract")
	}
	if len(descriptor.Fields) != len(wantFields) {
		t.Fatal("call participant presence index fields must retain the frozen order")
	}
	for position, wantField := range wantFields {
		if descriptor.Fields[position] != wantField {
			t.Fatal("call participant presence index fields must retain the frozen order")
		}
	}
	if len(descriptor.Edges) != 0 || len(descriptor.Annotations) != 0 || descriptor.StorageKey != "" {
		t.Fatal("call participant presence indexes must use only default field storage")
	}
}

func requirePhase150PresenceGeneratedUUID(t *testing.T, descriptor *field.Descriptor) {
	t.Helper()

	requirePhase150PresenceFieldType(t, descriptor, field.TypeUUID)
	if !descriptor.Immutable {
		t.Fatal("id must be immutable")
	}
	if descriptor.Optional || descriptor.Nillable {
		t.Fatal("id must be required and non-nillable")
	}
	if descriptor.UpdateDefault != nil {
		t.Fatal("id must not declare an update default")
	}

	factory, ok := descriptor.Default.(func() uuid.UUID)
	if !ok || factory() == uuid.Nil {
		t.Fatal("id must use a non-nil generated UUID default")
	}
}

func requirePhase150PresenceRequiredUUID(t *testing.T, descriptor *field.Descriptor) {
	t.Helper()

	requirePhase150PresenceFieldType(t, descriptor, field.TypeUUID)
	if !descriptor.Immutable {
		t.Fatal("call_id and user_id must be immutable")
	}
	if descriptor.Optional || descriptor.Nillable {
		t.Fatal("call_id and user_id must be required and non-nillable")
	}
	if descriptor.Default != nil || descriptor.UpdateDefault != nil {
		t.Fatal("call_id and user_id must not declare defaults")
	}
}

func requirePhase150PresenceNonnegativeCount(t *testing.T, descriptor *field.Descriptor) {
	t.Helper()

	requirePhase150PresenceFieldType(t, descriptor, field.TypeInt64)
	if descriptor.Immutable {
		t.Fatal("presence counters must be mutable")
	}
	if descriptor.Optional || descriptor.Nillable {
		t.Fatal("presence counters must be required and non-nillable")
	}
	if defaultValue, ok := descriptor.Default.(int64); !ok || defaultValue != 0 {
		t.Fatal("presence counters must default to signed int64 zero")
	}
	if descriptor.UpdateDefault != nil {
		t.Fatal("presence counters must not declare update defaults")
	}
	if len(descriptor.Validators) == 0 {
		t.Fatal("presence counters must reject negative values")
	}

	rejectsNegative := false
	for _, configuredValidator := range descriptor.Validators {
		validator, ok := configuredValidator.(func(int64) error)
		if !ok {
			t.Fatal("presence counter validators must be func(int64) error")
		}
		if validator(int64(0)) != nil || validator(int64(1)) != nil || validator(int64(^uint64(0)>>1)) != nil {
			t.Fatal("presence counter validators must accept nonnegative signed int64 values")
		}
		if validator(int64(-1)) != nil {
			rejectsNegative = true
		}
	}
	if !rejectsNegative {
		t.Fatal("presence counters must reject negative signed int64 values")
	}
}

func requirePhase150PresenceTimestamp(
	t *testing.T,
	descriptor *field.Descriptor,
	wantImmutable bool,
	wantUpdateDefault bool,
) {
	t.Helper()

	requirePhase150PresenceFieldType(t, descriptor, field.TypeTime)
	if descriptor.Immutable != wantImmutable {
		t.Fatal("call participant presence timestamp immutability must match the frozen contract")
	}
	if descriptor.Optional || descriptor.Nillable {
		t.Fatal("call participant presence timestamps must be required and non-nillable")
	}
	if descriptor.Default == nil {
		t.Fatal("call participant presence timestamps must declare defaults")
	}
	if (descriptor.UpdateDefault != nil) != wantUpdateDefault {
		t.Fatal("call participant presence timestamp update defaults must match the frozen contract")
	}
	requirePhase150PresenceUTCTimeDefault(t, descriptor.Default, "timestamp default")
	if wantUpdateDefault {
		requirePhase150PresenceUTCTimeDefault(t, descriptor.UpdateDefault, "timestamp update default")
	}
}

func requirePhase150PresenceFieldType(t *testing.T, descriptor *field.Descriptor, wantType field.Type) {
	t.Helper()

	if descriptor.Info == nil || descriptor.Info.Type != wantType {
		t.Fatal("call participant presence field type must match the frozen contract")
	}
	if descriptor.StorageKey != "" {
		t.Fatal("call participant presence fields must use default storage keys")
	}
}

func requirePhase150PresenceUTCTimeDefault(t *testing.T, configuredDefault any, label string) {
	t.Helper()
	factory, ok := configuredDefault.(func() time.Time)
	if !ok {
		t.Fatalf("%s must be func() time.Time", label)
	}
	value := factory()
	if value.IsZero() || value.Location() != time.UTC {
		t.Fatalf("%s must produce a non-zero UTC time", label)
	}
}
