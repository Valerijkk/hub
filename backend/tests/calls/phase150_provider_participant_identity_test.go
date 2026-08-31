//go:build hubtest

package calls_test

import (
	"encoding"
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calls"
)

const (
	callParticipantConnectionKeyCanonical = "hub.call-connection.v1/11111111-1111-4111-8111-111111111111/22222222-2222-4222-8222-222222222222/33333333-3333-4333-8333-333333333333"
	providerParticipantIdentityCanonical  = "hub.call-participant.v1/11111111-1111-4111-8111-111111111111/22222222-2222-4222-8222-222222222222/33333333-3333-4333-8333-333333333333"
)

func TestCallParticipantConnectionKeyConstructsCanonicalizesAndExposesUUIDAccessors(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	epoch := uuid.MustParse("33333333-3333-4333-8333-333333333333")

	key, err := calls.NewCallParticipantConnectionKey(callID, userID, epoch)
	if err != nil {
		t.Fatal("expected valid neutral connection key construction")
	}
	if !key.IsValid() {
		t.Error("expected constructed neutral connection key to be valid")
	}
	if key.Canonical() != callParticipantConnectionKeyCanonical {
		t.Error("expected neutral connection key canonical form")
	}
	if key.CallID() != callID || key.UserID() != userID || key.Epoch() != epoch {
		t.Error("expected neutral connection key UUID accessors")
	}
}

func TestCallParticipantConnectionKeyRejectsEveryZeroComponent(t *testing.T) {
	callID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	userID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	epoch := uuid.MustParse("33333333-3333-4333-8333-333333333333")

	tests := []struct {
		name                 string
		callID, userID, epoch uuid.UUID
	}{
		{name: "all zero"},
		{name: "zero call", userID: userID, epoch: epoch},
		{name: "zero user", callID: callID, epoch: epoch},
		{name: "zero epoch", callID: callID, userID: userID},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			key, err := calls.NewCallParticipantConnectionKey(testCase.callID, testCase.userID, testCase.epoch)

			assertCallParticipantConnectionKeyInvalid(t, key, err)
		})
	}
}

func TestCallParticipantConnectionKeyAcceptsNonV4AndRepeatedUUIDComponents(t *testing.T) {
	tests := []struct {
		name   string
		callID string
		userID string
		epoch  string
		want   string
	}{
		{
			name:   "canonical non v4 UUID versions",
			callID: "f47ac10b-58cc-11cf-8f0f-08002be10318",
			userID: "3d813cbb-47fb-32ba-91df-831e1593ac29",
			epoch:  "74738ff5-5367-5958-9aee-98fffdcd1876",
			want:   "hub.call-connection.v1/f47ac10b-58cc-11cf-8f0f-08002be10318/3d813cbb-47fb-32ba-91df-831e1593ac29/74738ff5-5367-5958-9aee-98fffdcd1876",
		},
		{
			name:   "repeated UUID components",
			callID: "f47ac10b-58cc-11cf-8f0f-08002be10318",
			userID: "f47ac10b-58cc-11cf-8f0f-08002be10318",
			epoch:  "f47ac10b-58cc-11cf-8f0f-08002be10318",
			want:   "hub.call-connection.v1/f47ac10b-58cc-11cf-8f0f-08002be10318/f47ac10b-58cc-11cf-8f0f-08002be10318/f47ac10b-58cc-11cf-8f0f-08002be10318",
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			key := validCallParticipantConnectionKeyFromStrings(t, testCase.callID, testCase.userID, testCase.epoch)

			if key.Canonical() != testCase.want {
				t.Error("expected neutral connection key canonical form")
			}
		})
	}
}

func TestProviderParticipantIdentityRequiresSeparateNeutralConnectionKeyType(t *testing.T) {
	var neutralKey calls.CallParticipantConnectionKey
	var confirmedKey calls.ConfirmedParticipantConnectionKey
	neutralType := reflect.TypeOf(neutralKey)
	confirmedType := reflect.TypeOf(confirmedKey)
	if neutralType == confirmedType {
		t.Fatal("expected neutral and confirmed connection keys to remain separate types")
	}

	constructorType := reflect.TypeOf(calls.NewProviderParticipantIdentity)
	if constructorType.NumIn() != 1 || constructorType.In(0) != neutralType {
		t.Error("expected provider participant identity to accept only the neutral connection key")
	}
}

func TestProviderParticipantIdentityConstructsAndCanonicalizesNeutralConnectionKey(t *testing.T) {
	key := validCallParticipantConnectionKey(t)

	identity, err := calls.NewProviderParticipantIdentity(key)
	if err != nil {
		t.Fatal("expected valid provider participant identity construction")
	}
	if !identity.IsValid() {
		t.Error("expected constructed provider participant identity to be valid")
	}
	if identity.Canonical() != providerParticipantIdentityCanonical {
		t.Error("expected provider participant identity canonical form")
	}
	assertSameCallParticipantConnectionKey(t, identity.ConnectionKey(), key)
	if identity.ConnectionKey().Canonical() != callParticipantConnectionKeyCanonical {
		t.Error("expected provider identity to retain the distinct neutral canonical language")
	}
}

func TestProviderParticipantIdentityParsesExactCanonicalRoundTrip(t *testing.T) {
	if len(providerParticipantIdentityCanonical) != 134 {
		t.Fatal("expected provider participant identity fixture to use the exact 134-byte grammar")
	}

	identity, err := calls.ParseProviderParticipantIdentity(providerParticipantIdentityCanonical)
	if err != nil {
		t.Fatal("expected canonical provider participant identity to parse")
	}
	if !identity.IsValid() {
		t.Error("expected parsed provider participant identity to be valid")
	}
	if identity.Canonical() != providerParticipantIdentityCanonical {
		t.Error("expected parsed provider participant identity canonical round trip")
	}
	assertSameCallParticipantConnectionKey(t, identity.ConnectionKey(), validCallParticipantConnectionKey(t))
}

func TestProviderParticipantIdentityRoundTripsCanonicalUUIDsWithoutVersionVariantOrDistinctnessPolicy(t *testing.T) {
	tests := []struct {
		name      string
		canonical string
		key       calls.CallParticipantConnectionKey
	}{
		{
			name: "UUIDs outside RFC version and variant patterns",
			canonical: "hub.call-participant.v1/11111111-1111-e111-c111-111111111111/22222222-2222-0111-f111-222222222222/33333333-3333-b111-d111-333333333333",
			key: validCallParticipantConnectionKeyFromStrings(
				t,
				"11111111-1111-e111-c111-111111111111",
				"22222222-2222-0111-f111-222222222222",
				"33333333-3333-b111-d111-333333333333",
			),
		},
		{
			name: "repeated UUID components",
			canonical: "hub.call-participant.v1/f47ac10b-58cc-11cf-8f0f-08002be10318/f47ac10b-58cc-11cf-8f0f-08002be10318/f47ac10b-58cc-11cf-8f0f-08002be10318",
			key: validCallParticipantConnectionKeyFromStrings(
				t,
				"f47ac10b-58cc-11cf-8f0f-08002be10318",
				"f47ac10b-58cc-11cf-8f0f-08002be10318",
				"f47ac10b-58cc-11cf-8f0f-08002be10318",
			),
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			identity, err := calls.NewProviderParticipantIdentity(testCase.key)
			if err != nil {
				t.Fatal("expected valid provider participant identity construction")
			}
			if identity.Canonical() != testCase.canonical {
				t.Error("expected constructed provider participant identity canonical form")
			}

			parsed, err := calls.ParseProviderParticipantIdentity(testCase.canonical)
			if err != nil {
				t.Fatal("expected canonical provider participant identity to parse")
			}
			if parsed.Canonical() != testCase.canonical {
				t.Error("expected parsed provider participant identity round trip")
			}
			assertSameCallParticipantConnectionKey(t, parsed.ConnectionKey(), testCase.key)
		})
	}
}

func TestProviderParticipantIdentityRejectsZeroNeutralConnectionKey(t *testing.T) {
	var zeroKey calls.CallParticipantConnectionKey
	identity, err := calls.NewProviderParticipantIdentity(zeroKey)

	assertProviderParticipantIdentityInvalid(t, identity, err)
}

func TestProviderParticipantIdentityRejectsMalformedCanonicalInput(t *testing.T) {
	valid := providerParticipantIdentityCanonical
	callID := "11111111-1111-4111-8111-111111111111"
	userID := "22222222-2222-4222-8222-222222222222"
	epoch := "33333333-3333-4333-8333-333333333333"
	prefix := "hub.call-participant.v1/"

	tests := []struct {
		name string
		raw  string
	}{
		{name: "zero call", raw: prefix + "00000000-0000-0000-0000-000000000000/" + userID + "/" + epoch},
		{name: "zero user", raw: prefix + callID + "/00000000-0000-0000-0000-000000000000/" + epoch},
		{name: "zero epoch", raw: prefix + callID + "/" + userID + "/00000000-0000-0000-0000-000000000000"},
		{name: "empty input", raw: ""},
		{name: "invalid call UUID", raw: prefix + "x1111111-1111-4111-8111-111111111111/" + userID + "/" + epoch},
		{name: "invalid user UUID", raw: prefix + callID + "/x2222222-2222-4222-8222-222222222222/" + epoch},
		{name: "invalid epoch UUID", raw: prefix + callID + "/" + userID + "/x3333333-3333-4333-8333-333333333333"},
		{name: "uppercase UUID", raw: prefix + "11111111-1111-4111-8111-11111111111A/" + userID + "/" + epoch},
		{name: "braced UUID", raw: prefix + "{11111111-1111-4111-8111-111111111111}/" + userID + "/" + epoch},
		{name: "UUID URN", raw: prefix + "urn:uuid:11111111-1111-4111-8111-111111111111/" + userID + "/" + epoch},
		{name: "compact UUID", raw: prefix + "11111111111141118111111111111111/" + userID + "/" + epoch},
		{name: "leading whitespace", raw: " " + valid},
		{name: "trailing whitespace", raw: valid + " "},
		{name: "newline", raw: valid + "\n"},
		{name: "NUL", raw: valid + "\x00"},
		{name: "non ASCII", raw: valid + "\u00e9"},
		{name: "connection key namespace", raw: callParticipantConnectionKeyCanonical},
		{name: "uppercase namespace", raw: "HUB.CALL-PARTICIPANT.V1/" + callID + "/" + userID + "/" + epoch},
		{name: "mixed case namespace", raw: "hub.Call-participant.v1/" + callID + "/" + userID + "/" + epoch},
		{name: "wrong namespace", raw: "hub.call-observer.v1/" + callID + "/" + userID + "/" + epoch},
		{name: "wrong version", raw: "hub.call-participant.v2/" + callID + "/" + userID + "/" + epoch},
		{name: "empty segment", raw: prefix + "/" + userID + "/" + epoch},
		{name: "empty user", raw: prefix + callID + "//" + epoch},
		{name: "empty epoch", raw: prefix + callID + "/" + userID + "/"},
		{name: "missing segment", raw: prefix + callID + "/" + userID},
		{name: "extra segment", raw: valid + "/extra"},
		{name: "query suffix", raw: valid + "?next=1"},
		{name: "fragment suffix", raw: valid + "#fragment"},
		{name: "percent escaped UUID", raw: prefix + "%31111111-1111-4111-8111-111111111111/" + userID + "/" + epoch},
		{name: "133 bytes", raw: valid[:133]},
		{name: "135 bytes", raw: valid + "x"},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			identity, err := calls.ParseProviderParticipantIdentity(testCase.raw)

			assertProviderParticipantIdentityInvalid(t, identity, err)
		})
	}
}

func TestCallParticipantConnectionKeyAndProviderIdentityDoNotExposeStringTextOrJSONInterfaces(t *testing.T) {
	var key calls.CallParticipantConnectionKey
	var identity calls.ProviderParticipantIdentity
	assertNoLeakageInterfaces(t, key)
	assertNoLeakageInterfaces(t, identity)
}

func TestCallParticipantConnectionKeyAndProviderIdentityAreOpaqueWithoutRepresentationAssumptions(t *testing.T) {
	key := validCallParticipantConnectionKey(t)
	identity, err := calls.NewProviderParticipantIdentity(key)
	if err != nil {
		t.Fatal("expected valid provider participant identity construction")
	}

	for _, value := range []any{key, identity} {
		valueType := reflect.TypeOf(value)
		if valueType.Kind() != reflect.Struct {
			continue
		}
		exportedFieldCount := 0
		for index := 0; index < valueType.NumField(); index++ {
			if valueType.Field(index).IsExported() {
				exportedFieldCount++
			}
		}
		if exportedFieldCount != 0 {
			t.Error("expected opaque value to expose zero exported fields")
		}
	}
}

func TestCallParticipantConnectionKeyAndProviderIdentityZeroValuesAreSafeAndInvalid(t *testing.T) {
	var key calls.CallParticipantConnectionKey
	if key.IsValid() {
		t.Error("expected zero neutral connection key to be invalid")
	}
	if key.Canonical() != "" {
		t.Error("expected zero neutral connection key canonical form to be empty")
	}
	if key.CallID() != uuid.Nil || key.UserID() != uuid.Nil || key.Epoch() != uuid.Nil {
		t.Error("expected zero neutral connection key accessors to be zero")
	}

	var identity calls.ProviderParticipantIdentity
	if identity.IsValid() {
		t.Error("expected zero provider participant identity to be invalid")
	}
	if identity.Canonical() != "" {
		t.Error("expected zero provider participant identity canonical form to be empty")
	}
	assertZeroCallParticipantConnectionKey(t, identity.ConnectionKey())
}

func validCallParticipantConnectionKey(t *testing.T) calls.CallParticipantConnectionKey {
	return validCallParticipantConnectionKeyFromStrings(
		t,
		"11111111-1111-4111-8111-111111111111",
		"22222222-2222-4222-8222-222222222222",
		"33333333-3333-4333-8333-333333333333",
	)
}

func validCallParticipantConnectionKeyFromStrings(t *testing.T, callID, userID, epoch string) calls.CallParticipantConnectionKey {
	t.Helper()

	key, err := calls.NewCallParticipantConnectionKey(
		uuid.MustParse(callID),
		uuid.MustParse(userID),
		uuid.MustParse(epoch),
	)
	if err != nil {
		t.Fatal("expected valid neutral connection key fixture")
	}

	return key
}

func assertCallParticipantConnectionKeyInvalid(t *testing.T, key calls.CallParticipantConnectionKey, err error) {
	t.Helper()

	if err != calls.ErrCallParticipantConnectionKeyInvalid {
		t.Error("expected exact invalid neutral connection key sentinel")
	}
	if key.IsValid() {
		t.Error("expected rejected neutral connection key to be invalid")
	}
	if key.Canonical() != "" {
		t.Error("expected rejected neutral connection key canonical form to be empty")
	}
	assertZeroCallParticipantConnectionKey(t, key)
}

func assertProviderParticipantIdentityInvalid(t *testing.T, identity calls.ProviderParticipantIdentity, err error) {
	t.Helper()

	if err != calls.ErrProviderParticipantIdentityInvalid {
		t.Error("expected exact invalid provider participant identity sentinel")
	}
	if identity.IsValid() {
		t.Error("expected rejected provider participant identity to be invalid")
	}
	if identity.Canonical() != "" {
		t.Error("expected rejected provider participant identity canonical form to be empty")
	}
	assertZeroCallParticipantConnectionKey(t, identity.ConnectionKey())
}

func assertSameCallParticipantConnectionKey(t *testing.T, got, want calls.CallParticipantConnectionKey) {
	t.Helper()

	if !got.IsValid() || !want.IsValid() {
		t.Error("expected compared neutral connection keys to be valid")
	}
	if got.Canonical() != want.Canonical() {
		t.Error("expected provider participant identity to retain its neutral connection key")
	}
	if got.CallID() != want.CallID() || got.UserID() != want.UserID() || got.Epoch() != want.Epoch() {
		t.Error("expected compared neutral connection key UUID accessors to match")
	}
}

func assertZeroCallParticipantConnectionKey(t *testing.T, key calls.CallParticipantConnectionKey) {
	t.Helper()

	if key.IsValid() {
		t.Error("expected neutral connection key to be invalid")
	}
	if key.Canonical() != "" {
		t.Error("expected neutral connection key canonical form to be empty")
	}
	if key.CallID() != uuid.Nil || key.UserID() != uuid.Nil || key.Epoch() != uuid.Nil {
		t.Error("expected neutral connection key accessors to be zero")
	}
}

func assertNoLeakageInterfaces(t *testing.T, value any) {
	t.Helper()

	checks := []struct {
		name string
		value any
	}{
		{name: "value fmt.Stringer", value: value},
		{name: "pointer fmt.Stringer", value: reflect.New(reflect.TypeOf(value)).Interface()},
	}
	for _, check := range checks {
		if _, ok := check.value.(fmt.Stringer); ok {
			t.Errorf("expected %s not to implement fmt.Stringer", check.name)
		}
	}

	if _, ok := value.(encoding.TextMarshaler); ok {
		t.Error("expected value not to implement encoding.TextMarshaler")
	}
	if _, ok := reflect.New(reflect.TypeOf(value)).Interface().(encoding.TextMarshaler); ok {
		t.Error("expected pointer not to implement encoding.TextMarshaler")
	}
	if _, ok := value.(encoding.TextUnmarshaler); ok {
		t.Error("expected value not to implement encoding.TextUnmarshaler")
	}
	if _, ok := reflect.New(reflect.TypeOf(value)).Interface().(encoding.TextUnmarshaler); ok {
		t.Error("expected pointer not to implement encoding.TextUnmarshaler")
	}
	if _, ok := value.(json.Marshaler); ok {
		t.Error("expected value not to implement json.Marshaler")
	}
	if _, ok := reflect.New(reflect.TypeOf(value)).Interface().(json.Marshaler); ok {
		t.Error("expected pointer not to implement json.Marshaler")
	}
	if _, ok := value.(json.Unmarshaler); ok {
		t.Error("expected value not to implement json.Unmarshaler")
	}
	if _, ok := reflect.New(reflect.TypeOf(value)).Interface().(json.Unmarshaler); ok {
		t.Error("expected pointer not to implement json.Unmarshaler")
	}
}
