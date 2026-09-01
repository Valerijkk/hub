//go:build hubtest

package calls_test

import (
	"reflect"
	"strings"
	"testing"
	"time"

	"entgo.io/ent/schema/field"
	"entgo.io/ent/schema/index"
	callschema "github.com/valerii/hub/backend/ent/schema"
	"github.com/google/uuid"
)

func TestPhase150CallProviderDeliverySchemaContract(t *testing.T) {
	delivery := callschema.CallProviderDelivery{}
	fields := delivery.Fields()
	wantFieldNames := []string{
		"id",
		"provider",
		"delivery_id",
		"kind",
		"call_id",
		"user_id",
		"epoch",
		"result",
		"received_at",
		"terminal_at",
	}
	if len(fields) != len(wantFieldNames) {
		t.Fatalf("CallProviderDelivery must define exactly %d fields, got %d", len(wantFieldNames), len(fields))
	}

	descriptors := make(map[string]*field.Descriptor, len(fields))
	for position, configuredField := range fields {
		descriptor := configuredField.Descriptor()
		if descriptor == nil {
			t.Fatalf("CallProviderDelivery field at position %d must have a descriptor", position)
		}
		if descriptor.Name != wantFieldNames[position] {
			t.Fatalf("CallProviderDelivery field %d must be %q, got %q", position, wantFieldNames[position], descriptor.Name)
		}
		if _, duplicate := descriptors[descriptor.Name]; duplicate {
			t.Fatalf("CallProviderDelivery must not duplicate field %q", descriptor.Name)
		}
		phase150RequireCallProviderDeliverySafeFieldName(t, descriptor.Name)
		descriptors[descriptor.Name] = descriptor
	}

	phase150RequireCallProviderDeliveryGeneratedID(t, descriptors["id"])
	phase150RequireCallProviderDeliveryEnum(
		t,
		descriptors["provider"],
		"provider",
		[]phase150CallProviderDeliveryEnumPair{
			{N: "livekit", V: "livekit"},
			{N: "jitsi", V: "jitsi"},
		},
		nil,
		true,
	)
	phase150RequireCallProviderDeliveryID(t, descriptors["delivery_id"])
	phase150RequireCallProviderDeliveryEnum(
		t,
		descriptors["kind"],
		"kind",
		[]phase150CallProviderDeliveryEnumPair{
			{N: "connected", V: "connected"},
			{N: "disconnected", V: "disconnected"},
		},
		nil,
		true,
	)
	for _, fieldName := range []string{"call_id", "epoch"} {
		phase150RequireCallProviderDeliveryRequiredUUID(t, descriptors[fieldName], fieldName)
	}
	phase150RequireCallProviderDeliveryGDPRUserID(t, descriptors["user_id"])
	phase150RequireCallProviderDeliveryEnum(
		t,
		descriptors["result"],
		"result",
		[]phase150CallProviderDeliveryEnumPair{
			{N: "pending", V: "pending"},
			{N: "applied", V: "applied"},
			{N: "ignored", V: "ignored"},
			{N: "reconcile", V: "reconcile"},
		},
		"pending",
		false,
	)
	phase150RequireCallProviderDeliveryReceivedAt(t, descriptors["received_at"])
	phase150RequireCallProviderDeliveryTerminalAt(t, descriptors["terminal_at"])

	for fieldName, descriptor := range descriptors {
		if descriptor.Unique {
			t.Fatalf("%s must not create a field-level unique constraint; dedupe is only (provider, delivery_id)", fieldName)
		}
	}

	if mixins := delivery.Mixin(); len(mixins) != 0 {
		t.Fatalf("CallProviderDelivery must not define mixins, got %d", len(mixins))
	}
	if edges := delivery.Edges(); len(edges) != 0 {
		t.Fatalf("CallProviderDelivery must not define Ent edges, got %d", len(edges))
	}

	indexes := delivery.Indexes()
	if len(indexes) != 3 {
		t.Fatalf("CallProviderDelivery must define exactly three indexes, got %d", len(indexes))
	}
	phase150RequireCallProviderDeliveryIndex(t, indexes[0].Descriptor(), []string{"provider", "delivery_id"}, true, "dedupe")
	phase150RequireCallProviderDeliveryIndex(t, indexes[1].Descriptor(), []string{"result", "terminal_at"}, false, "terminal cleanup")
	phase150RequireCallProviderDeliveryIndex(t, indexes[2].Descriptor(), []string{"user_id"}, false, "GDPR tombstone lookup")
}

type phase150CallProviderDeliveryEnumPair struct {
	N string
	V string
}

func phase150RequireCallProviderDeliveryGeneratedID(t *testing.T, descriptor *field.Descriptor) {
	t.Helper()
	phase150RequireCallProviderDeliveryBaseField(t, descriptor, "id", field.TypeUUID, true, false, false, true)
	if !phase150CallProviderDeliverySameFunction(descriptor.Default, uuid.New) {
		t.Fatal("id must use uuid.New as its generated UUID default")
	}
}

func phase150RequireCallProviderDeliveryRequiredUUID(t *testing.T, descriptor *field.Descriptor, fieldName string) {
	t.Helper()
	phase150RequireCallProviderDeliveryBaseField(t, descriptor, fieldName, field.TypeUUID, true, false, false, false)
	if descriptor.Default != nil {
		t.Fatalf("%s must not have a default", fieldName)
	}
}

func phase150RequireCallProviderDeliveryGDPRUserID(t *testing.T, descriptor *field.Descriptor) {
	t.Helper()
	phase150RequireCallProviderDeliveryBaseField(t, descriptor, "user_id", field.TypeUUID, false, true, true, false)
	if descriptor.Default != nil {
		t.Fatal("user_id must not have a default; NULL is reserved for a terminal GDPR tombstone")
	}
}

func phase150RequireCallProviderDeliveryID(t *testing.T, descriptor *field.Descriptor) {
	t.Helper()
	phase150RequireCallProviderDeliveryBaseField(t, descriptor, "delivery_id", field.TypeString, true, false, false, false)
	if descriptor.Default != nil {
		t.Fatal("delivery_id must not have a default")
	}
	if descriptor.Size != 255 {
		t.Fatalf("delivery_id must have MaxLen(255), got size %d", descriptor.Size)
	}
	if len(descriptor.Validators) != 2 {
		t.Fatalf("delivery_id must define exactly NotEmpty and MaxLen validators, got %d", len(descriptor.Validators))
	}
	for _, value := range []string{"", strings.Repeat("x", 256)} {
		if phase150CallProviderDeliveryValidatesString(descriptor, value) {
			t.Fatalf("delivery_id must reject %q", value)
		}
	}
	if !phase150CallProviderDeliveryValidatesString(descriptor, strings.Repeat("x", 255)) {
		t.Fatal("delivery_id must accept a nonempty 255-character value")
	}
}

func phase150RequireCallProviderDeliveryEnum(
	t *testing.T,
	descriptor *field.Descriptor,
	fieldName string,
	wantEnums []phase150CallProviderDeliveryEnumPair,
	wantDefault any,
	wantImmutable bool,
) {
	t.Helper()
	phase150RequireCallProviderDeliveryBaseField(t, descriptor, fieldName, field.TypeEnum, wantImmutable, false, false, wantDefault != nil)
	if !reflect.DeepEqual(descriptor.Default, wantDefault) {
		t.Fatalf("%s default must be %#v, got %#v", fieldName, wantDefault, descriptor.Default)
	}
	if len(descriptor.Enums) != len(wantEnums) {
		t.Fatalf("%s must define exactly %d enum values, got %d", fieldName, len(wantEnums), len(descriptor.Enums))
	}
	for position, wantEnum := range wantEnums {
		gotEnum := descriptor.Enums[position]
		if gotEnum.N != wantEnum.N || gotEnum.V != wantEnum.V {
			t.Fatalf("%s enum value %d must be (%q, %q), got (%q, %q)", fieldName, position, wantEnum.N, wantEnum.V, gotEnum.N, gotEnum.V)
		}
	}
}

func phase150RequireCallProviderDeliveryReceivedAt(t *testing.T, descriptor *field.Descriptor) {
	t.Helper()
	phase150RequireCallProviderDeliveryBaseField(t, descriptor, "received_at", field.TypeTime, true, false, false, true)
	defaultFactory, ok := descriptor.Default.(func() time.Time)
	if !ok {
		t.Fatal("received_at must have a time-valued default function")
	}
	defaultValue := defaultFactory()
	if defaultValue.IsZero() || defaultValue.Location() != time.UTC {
		t.Fatal("received_at default must produce a non-zero UTC time")
	}
}

func phase150RequireCallProviderDeliveryTerminalAt(t *testing.T, descriptor *field.Descriptor) {
	t.Helper()
	phase150RequireCallProviderDeliveryBaseField(t, descriptor, "terminal_at", field.TypeTime, false, true, true, false)
	if descriptor.Default != nil {
		t.Fatal("terminal_at must not have a default")
	}
}

func phase150RequireCallProviderDeliveryBaseField(
	t *testing.T,
	descriptor *field.Descriptor,
	fieldName string,
	wantType field.Type,
	wantImmutable bool,
	wantOptional bool,
	wantNillable bool,
	wantDefault bool,
) {
	t.Helper()
	if descriptor == nil {
		t.Fatalf("%s must have a descriptor", fieldName)
	}
	if descriptor.Info == nil || descriptor.Info.Type != wantType {
		t.Fatalf("%s must have type %s", fieldName, wantType)
	}
	if descriptor.Immutable != wantImmutable {
		t.Fatalf("%s immutable must be %t", fieldName, wantImmutable)
	}
	if descriptor.Optional != wantOptional {
		t.Fatalf("%s optional must be %t", fieldName, wantOptional)
	}
	if descriptor.Nillable != wantNillable {
		t.Fatalf("%s nillable must be %t", fieldName, wantNillable)
	}
	if (descriptor.Default != nil) != wantDefault {
		t.Fatalf("%s default presence must be %t", fieldName, wantDefault)
	}
	if descriptor.UpdateDefault != nil {
		t.Fatalf("%s must not have an update default", fieldName)
	}
	if descriptor.Err != nil {
		t.Fatalf("%s descriptor must not report an error: %v", fieldName, descriptor.Err)
	}
	if descriptor.StorageKey != "" {
		t.Fatalf("%s must use the default storage key", fieldName)
	}
}

func phase150RequireCallProviderDeliverySafeFieldName(t *testing.T, fieldName string) {
	t.Helper()
	for _, forbidden := range []string{
		"raw",
		"payload",
		"digest",
		"hmac",
		"mac",
		"key",
		"cert",
		"room",
		"identity",
		"jwt",
		"token",
		"error",
		"trace",
		"outbox",
	} {
		if strings.Contains(strings.ToLower(fieldName), forbidden) {
			t.Fatalf("CallProviderDelivery must not store %s-derived field %q", forbidden, fieldName)
		}
	}
}

func phase150RequireCallProviderDeliveryIndex(
	t *testing.T,
	descriptor *index.Descriptor,
	wantFields []string,
	wantUnique bool,
	indexPurpose string,
) {
	t.Helper()
	if descriptor == nil {
		t.Fatalf("%s index must have a descriptor", indexPurpose)
	}
	if !reflect.DeepEqual(descriptor.Fields, wantFields) {
		t.Fatalf("%s index fields must be %v, got %v", indexPurpose, wantFields, descriptor.Fields)
	}
	if descriptor.Unique != wantUnique {
		t.Fatalf("%s index unique must be %t", indexPurpose, wantUnique)
	}
	if len(descriptor.Edges) != 0 {
		t.Fatalf("%s index must not reference Ent edges, got %v", indexPurpose, descriptor.Edges)
	}
	if descriptor.StorageKey != "" {
		t.Fatalf("%s index must use the default storage key", indexPurpose)
	}
	if len(descriptor.Annotations) != 0 {
		t.Fatalf("%s index must not define annotations", indexPurpose)
	}
}

func phase150CallProviderDeliveryValidatesString(descriptor *field.Descriptor, value string) bool {
	for _, configuredValidator := range descriptor.Validators {
		validator, ok := configuredValidator.(func(string) error)
		if !ok || validator(value) != nil {
			return false
		}
	}
	return true
}

func phase150CallProviderDeliverySameFunction(got any, want any) bool {
	gotValue := reflect.ValueOf(got)
	wantValue := reflect.ValueOf(want)
	return gotValue.IsValid() &&
		wantValue.IsValid() &&
		gotValue.Kind() == reflect.Func &&
		wantValue.Kind() == reflect.Func &&
		gotValue.Type() == wantValue.Type() &&
		gotValue.Pointer() == wantValue.Pointer()
}
