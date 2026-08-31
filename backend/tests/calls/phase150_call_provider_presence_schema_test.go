//go:build hubtest

package calls_test

import (
	"testing"

	"entgo.io/ent/schema/field"
	callschema "github.com/valerii/hub/backend/ent/schema"
)

func TestPhase150CallProviderPresenceSchemaContract(t *testing.T) {
	fields := (callschema.Call{}).Fields()

	providerCount := 0
	presenceModeCount := 0
	for _, configuredField := range fields {
		descriptor := configuredField.Descriptor()
		if descriptor == nil {
			continue
		}

		switch descriptor.Name {
		case "provider":
			providerCount++
			requirePhase150EnumDescriptor(
				t,
				descriptor,
				[]phase150EnumPair{
					{N: "livekit", V: "livekit"},
					{N: "jitsi", V: "jitsi"},
				},
				"livekit",
				"provider",
			)
		case "presence_mode":
			presenceModeCount++
			requirePhase150EnumDescriptor(
				t,
				descriptor,
				[]phase150EnumPair{
					{N: "legacy", V: "legacy"},
					{N: "durable", V: "durable"},
				},
				"legacy",
				"presence_mode",
			)
		}
	}

	if providerCount != 1 {
		t.Fatal("provider must have exactly one descriptor")
	}
	if presenceModeCount != 1 {
		t.Fatal("presence_mode must have exactly one descriptor")
	}
}

type phase150EnumPair struct {
	N string
	V string
}

func requirePhase150EnumDescriptor(
	t *testing.T,
	descriptor *field.Descriptor,
	wantEnums []phase150EnumPair,
	wantDefault string,
	fieldName string,
) {
	t.Helper()

	if descriptor.Info == nil {
		t.Fatalf("%s must have type information", fieldName)
	}
	if descriptor.Info.Type != field.TypeEnum {
		t.Fatalf("%s must be an enum descriptor", fieldName)
	}
	if len(descriptor.Enums) != len(wantEnums) {
		t.Fatalf("%s must declare the frozen enum values in order", fieldName)
	}
	for index, wantEnum := range wantEnums {
		gotEnum := descriptor.Enums[index]
		if gotEnum.N != wantEnum.N || gotEnum.V != wantEnum.V {
			t.Fatalf("%s must declare the frozen enum values in order", fieldName)
		}
	}
	if descriptor.Default != wantDefault {
		t.Fatalf("%s must declare the frozen default", fieldName)
	}
	if !descriptor.Immutable {
		t.Fatalf("%s must be immutable", fieldName)
	}
	if descriptor.Optional {
		t.Fatalf("%s must be required", fieldName)
	}
	if descriptor.Nillable {
		t.Fatalf("%s must be non-nillable", fieldName)
	}
	if descriptor.Unique {
		t.Fatalf("%s must not be unique", fieldName)
	}
	if descriptor.UpdateDefault != nil {
		t.Fatalf("%s must not declare an update default", fieldName)
	}
	if descriptor.Err != nil {
		t.Fatalf("%s descriptor must not report an error", fieldName)
	}
}
