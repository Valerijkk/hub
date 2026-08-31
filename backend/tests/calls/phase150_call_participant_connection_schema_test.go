//go:build hubtest

package calls_test

import (
	"reflect"
	"testing"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
	callschema "github.com/valerii/hub/backend/ent/schema"
)

// TestPhase150CallParticipantConnectionSchemaDescriptorContract is the CALL-DUR-02A
// external descriptor oracle. It intentionally has no SQL, generated-Ent, filesystem,
// database, network, enttest, transition, retention-worker, or runtime assertion.
func TestPhase150CallParticipantConnectionSchemaDescriptorContract(t *testing.T) {
	connection := callschema.CallParticipantConnection{}
	fields := connection.Fields()

	wantNames := []string{
		"id",
		"call_id",
		"user_id",
		"epoch",
		"state",
		"connected_at",
		"left_at",
		"created_at",
		"updated_at",
	}
	if len(fields) != len(wantNames) {
		t.Fatalf("CallParticipantConnection must expose exactly %d fields, got %d", len(wantNames), len(fields))
	}

	descriptors := make(map[string]*field.Descriptor, len(fields))
	for position, configuredField := range fields {
		descriptor := configuredField.Descriptor()
		if descriptor == nil {
			t.Fatalf("CallParticipantConnection field %d must expose a descriptor", position)
		}
		if descriptor.Name != wantNames[position] {
			t.Fatalf("CallParticipantConnection field %d must be %q, got %q", position, wantNames[position], descriptor.Name)
		}
		if descriptor.Err != nil {
			t.Fatalf("CallParticipantConnection field %q descriptor must be valid: %v", descriptor.Name, descriptor.Err)
		}
		descriptors[descriptor.Name] = descriptor
	}

	for _, forbidden := range []string{
		"call_participant_id",
		"provider",
		"delivery_id",
		"token",
		"raw_event",
		"public_presence",
		"connected_count",
		"presence_sequence",
		"outbox_id",
	} {
		if _, found := descriptors[forbidden]; found {
			t.Fatalf("CallParticipantConnection must not declare %q", forbidden)
		}
	}

	requirePhase150ConnectionField(t, descriptors["id"], "id", field.TypeUUID, true, false, false, true, false)
	requirePhase150UUIDDefault(t, descriptors["id"], "id")

	for _, name := range []string{"call_id", "user_id", "epoch"} {
		requirePhase150ConnectionField(t, descriptors[name], name, field.TypeUUID, true, false, false, false, false)
	}

	state := descriptors["state"]
	requirePhase150ConnectionField(t, state, "state", field.TypeEnum, false, false, false, true, false)
	requirePhase150Enum(t, state, "state", []string{"unknown", "connected", "left"}, "unknown")

	for _, name := range []string{"connected_at", "left_at"} {
		requirePhase150ConnectionField(t, descriptors[name], name, field.TypeTime, false, true, true, false, false)
	}

	createdAt := descriptors["created_at"]
	requirePhase150ConnectionField(t, createdAt, "created_at", field.TypeTime, true, false, false, true, false)
	requirePhase150UTCTimeDefault(t, createdAt.Default, "created_at default")

	updatedAt := descriptors["updated_at"]
	requirePhase150ConnectionField(t, updatedAt, "updated_at", field.TypeTime, false, false, false, true, true)
	requirePhase150UTCTimeDefault(t, updatedAt.Default, "updated_at default")
	requirePhase150UTCTimeDefault(t, updatedAt.UpdateDefault, "updated_at update default")

	mixins := connection.Mixin()
	if len(mixins) != 1 {
		t.Fatalf("CallParticipantConnection must use exactly VersionMixin, got %d mixins", len(mixins))
	}
	if _, ok := mixins[0].(callschema.VersionMixin); !ok {
		t.Fatalf("CallParticipantConnection must use the existing VersionMixin, got %T", mixins[0])
	}
	requirePhase150VersionMixin(t, mixins[0])

	edges := connection.Edges()
	if len(edges) != 2 {
		t.Fatalf("CallParticipantConnection must expose exactly two child-owned edges, got %d", len(edges))
	}
	requirePhase150ConnectionEdge(t, edges[0], "call", "Call", "call_id", entsql.Cascade)
	requirePhase150ConnectionEdge(t, edges[1], "user", "User", "user_id", entsql.NoAction)

	indexes := connection.Indexes()
	if len(indexes) != 3 {
		t.Fatalf("CallParticipantConnection must expose exactly three indexes, got %d", len(indexes))
	}

	tuple := indexes[0].Descriptor()
	if tuple == nil || !tuple.Unique || !reflect.DeepEqual(tuple.Fields, []string{"call_id", "user_id", "epoch"}) || tuple.StorageKey != "" || len(tuple.Edges) != 0 || len(tuple.Annotations) != 0 {
		t.Fatal("CallParticipantConnection must first declare only the unique (call_id, user_id, epoch) tuple index")
	}

	retention := indexes[1].Descriptor()
	if retention == nil || retention.Unique || !reflect.DeepEqual(retention.Fields, []string{"left_at"}) || retention.StorageKey != "" || len(retention.Edges) != 0 || len(retention.Annotations) != 1 {
		t.Fatal("CallParticipantConnection must second declare only the partial left_at retention index")
	}
	partial, ok := retention.Annotations[0].(*entsql.IndexAnnotation)
	if !ok || !reflect.DeepEqual(partial, &entsql.IndexAnnotation{Where: "state = 'left'"}) {
		t.Fatal("CallParticipantConnection left_at retention index must be annotated exactly with state = 'left'")
	}

	byUser := indexes[2].Descriptor()
	if byUser == nil || byUser.Unique || !reflect.DeepEqual(byUser.Fields, []string{"user_id"}) || byUser.StorageKey != "" || len(byUser.Edges) != 0 || len(byUser.Annotations) != 0 {
		t.Fatal("CallParticipantConnection must third declare only the default-storage non-unique user_id index")
	}
}

func requirePhase150ConnectionEdge(
	t *testing.T,
	configuredEdge ent.Edge,
	wantName string,
	wantType string,
	wantField string,
	wantOnDelete entsql.ReferenceOption,
) {
	t.Helper()

	descriptor := configuredEdge.Descriptor()
	if descriptor == nil {
		t.Fatalf("CallParticipantConnection edge %q must expose a descriptor", wantName)
	}
	if descriptor.Name != wantName || descriptor.Type != wantType || descriptor.Field != wantField {
		t.Fatalf("CallParticipantConnection edge %q must be bound to %s.%s, got name=%q type=%q field=%q", wantName, wantType, wantField, descriptor.Name, descriptor.Type, descriptor.Field)
	}
	if !descriptor.Required || !descriptor.Unique || !descriptor.Immutable {
		t.Fatalf("CallParticipantConnection edge %q must be required, unique, and immutable", wantName)
	}
	if descriptor.Inverse || descriptor.RefName != "" || descriptor.Ref != nil || descriptor.StorageKey != nil || descriptor.Through != nil {
		t.Fatalf("CallParticipantConnection edge %q must be a direct non-inverse child-owned edge without ref, storage key, or through", wantName)
	}
	if len(descriptor.Annotations) != 1 {
		t.Fatalf("CallParticipantConnection edge %q must expose exactly one OnDelete annotation, got %d", wantName, len(descriptor.Annotations))
	}
	onDelete, ok := descriptor.Annotations[0].(*entsql.Annotation)
	if !ok || !reflect.DeepEqual(onDelete, &entsql.Annotation{OnDelete: wantOnDelete}) {
		t.Fatalf("CallParticipantConnection edge %q must use exactly OnDelete(%s)", wantName, wantOnDelete)
	}
}

func requirePhase150ConnectionField(
	t *testing.T,
	descriptor *field.Descriptor,
	name string,
	wantType field.Type,
	wantImmutable bool,
	wantOptional bool,
	wantNillable bool,
	wantDefault bool,
	wantUpdateDefault bool,
) {
	t.Helper()

	if descriptor == nil || descriptor.Info == nil || descriptor.Info.Type != wantType {
		t.Fatalf("%s must be a %s field", name, wantType)
	}
	if descriptor.Immutable != wantImmutable {
		t.Fatalf("%s immutable must be %t", name, wantImmutable)
	}
	if descriptor.Optional != wantOptional {
		t.Fatalf("%s optional must be %t", name, wantOptional)
	}
	if descriptor.Nillable != wantNillable {
		t.Fatalf("%s nillable must be %t", name, wantNillable)
	}
	if descriptor.Unique {
		t.Fatalf("%s must not create a field-level unique constraint", name)
	}
	if (descriptor.Default != nil) != wantDefault {
		t.Fatalf("%s default presence must be %t", name, wantDefault)
	}
	if (descriptor.UpdateDefault != nil) != wantUpdateDefault {
		t.Fatalf("%s update-default presence must be %t", name, wantUpdateDefault)
	}
	if descriptor.StorageKey != "" {
		t.Fatalf("%s must use the default storage key", name)
	}
}

func requirePhase150UUIDDefault(t *testing.T, descriptor *field.Descriptor, name string) {
	t.Helper()

	newUUID, ok := descriptor.Default.(func() uuid.UUID)
	if !ok || newUUID() == uuid.Nil {
		t.Fatalf("%s must generate a non-nil UUID default", name)
	}
}

func requirePhase150Enum(t *testing.T, descriptor *field.Descriptor, name string, wantValues []string, wantDefault string) {
	t.Helper()

	if len(descriptor.Enums) != len(wantValues) {
		t.Fatalf("%s must declare exactly %d enum values", name, len(wantValues))
	}
	for position, want := range wantValues {
		got := descriptor.Enums[position]
		if got.N != want || got.V != want {
			t.Fatalf("%s enum value %d must be %q", name, position, want)
		}
	}
	if descriptor.Default != wantDefault || descriptor.UpdateDefault != nil {
		t.Fatalf("%s must default to %q without an update default", name, wantDefault)
	}
}

func requirePhase150UTCTimeDefault(t *testing.T, configuredDefault any, label string) {
	t.Helper()

	now, ok := configuredDefault.(func() time.Time)
	if !ok {
		t.Fatalf("%s must be a func() time.Time", label)
	}
	value := now()
	if value.IsZero() || value.Location() != time.UTC {
		t.Fatalf("%s must produce a non-zero UTC time", label)
	}
}

func requirePhase150VersionMixin(t *testing.T, configuredMixin ent.Mixin) {
	t.Helper()

	fields := configuredMixin.Fields()
	if len(fields) != 1 {
		t.Fatalf("VersionMixin must expose exactly version, got %d fields", len(fields))
	}
	version := fields[0].Descriptor()
	if version == nil || version.Info == nil || version.Name != "version" || version.Info.Type != field.TypeInt || version.Default != 1 || version.UpdateDefault != nil || version.Immutable || version.Optional || version.Nillable || version.Err != nil {
		t.Fatal("VersionMixin must expose mutable required int version with default one and no update default")
	}
	if len(version.Validators) != 1 {
		t.Fatalf("VersionMixin version must expose exactly one positive validator, got %d", len(version.Validators))
	}
	positive, ok := version.Validators[0].(func(int) error)
	if !ok || positive(0) == nil || positive(1) != nil {
		t.Fatal("VersionMixin version validator must reject zero and accept one")
	}
}
