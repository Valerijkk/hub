package mail_test

import (
	"errors"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestNormalizeAddress_CanonicalizesASCIIIDNAAndPreservesDisplayName(t *testing.T) {
	name := "\u00a0Cafe\u0301\u2003Example\u00a0"
	got, err := mail.NormalizeAddress(mail.EngineAddress{
		Email: "ALICE@B\u00dcCHER.Example",
		Name:  name,
	})
	if err != nil {
		t.Fatalf("NormalizeAddress() unexpected error = %v", err)
	}
	want := mail.EngineAddress{Email: "alice@xn--bcher-kva.example", Name: name}
	if got != want {
		t.Fatalf("NormalizeAddress() = %#v, want %#v", got, want)
	}
}

func TestNormalizeAddress_PreservesFormatCharactersInDisplayName(t *testing.T) {
	want := mail.EngineAddress{Email: "alice@example.test", Name: " \u200fOwner\u2003Name\u00a0"}
	got, err := mail.NormalizeAddress(want)
	if err != nil {
		t.Fatalf("NormalizeAddress() unexpected error = %v", err)
	}
	if got != want {
		t.Fatalf("NormalizeAddress() = %#v, want %#v", got, want)
	}
}

func TestNormalizeAddress_AcceptsReservedSyntheticFixturesWithoutDNS(t *testing.T) {
	for _, input := range []mail.EngineAddress{
		{Email: "u-0123456789ab@example.test"},
		{Email: "fixture@domain.invalid"},
		{Email: "first.last+tag@example.test"},
	} {
		got, err := mail.NormalizeAddress(input)
		if err != nil {
			t.Fatalf("NormalizeAddress(%#v) unexpected error = %v", input, err)
		}
		if got != input {
			t.Fatalf("NormalizeAddress(%#v) = %#v, want %#v", input, got, input)
		}
	}
}

func TestNormalizeAddress_EnforcesLocalDomainAndEmailBounds(t *testing.T) {
	validLocal := strings.Repeat("a", 64) + "@example.test"
	if _, err := mail.NormalizeAddress(mail.EngineAddress{Email: validLocal}); err != nil {
		t.Fatalf("NormalizeAddress(%q) unexpected error = %v", validLocal, err)
	}
	requireAddressError(t, mail.EngineAddress{Email: strings.Repeat("a", 65) + "@example.test"}, mail.AddressErrorTooLarge, mail.ErrAddressTooLarge)

	label63 := strings.Repeat("a", 63) + ".example.test"
	if _, err := mail.NormalizeAddress(mail.EngineAddress{Email: "a@" + label63}); err != nil {
		t.Fatalf("NormalizeAddress(label63) unexpected error = %v", err)
	}
	requireAddressError(t, mail.EngineAddress{Email: "a@" + strings.Repeat("a", 64) + ".example.test"}, mail.AddressErrorTooLarge, mail.ErrAddressTooLarge)

	domain253 := strings.Join([]string{strings.Repeat("a", 63), strings.Repeat("b", 63), strings.Repeat("c", 63), strings.Repeat("d", 61)}, ".")
	if len(domain253) != 253 {
		t.Fatalf("fixture domain length = %d, want 253", len(domain253))
	}
	if _, err := mail.NormalizeAddress(mail.EngineAddress{Email: "a@" + domain253}); err != nil {
		t.Fatalf("NormalizeAddress(domain253) unexpected error = %v", err)
	}
	domain254 := strings.Join([]string{strings.Repeat("a", 63), strings.Repeat("b", 63), strings.Repeat("c", 63), strings.Repeat("d", 62)}, ".")
	if len(domain254) != 254 {
		t.Fatalf("fixture domain length = %d, want 254", len(domain254))
	}
	requireAddressError(t, mail.EngineAddress{Email: "a@" + domain254}, mail.AddressErrorTooLarge, mail.ErrAddressTooLarge)

	if mail.MaxEngineAddressEmailBytes != 320 {
		t.Fatalf("MaxEngineAddressEmailBytes = %d, want 320", mail.MaxEngineAddressEmailBytes)
	}
	rawOverLimit := "a@" + strings.Repeat("\u00fc", 159) + "a"
	if len(rawOverLimit) != mail.MaxEngineAddressEmailBytes+1 {
		t.Fatalf("raw UTF-8 email fixture length = %d, want %d", len(rawOverLimit), mail.MaxEngineAddressEmailBytes+1)
	}
	rawLimitErr := normalizeAddressError(t, mail.EngineAddress{Email: rawOverLimit}, mail.AddressErrorTooLarge, mail.ErrAddressTooLarge)
	if got := rawLimitErr.Error(); got != "mail address exceeds bounds" {
		t.Fatalf("raw UTF-8 email overflow error = %q, want fixed safe text", got)
	}
	if strings.Contains(rawLimitErr.Error(), rawOverLimit) {
		t.Fatalf("raw UTF-8 email overflow error leaked input: %q", rawLimitErr.Error())
	}
	if 64+1+253 > mail.MaxEngineAddressEmailBytes {
		t.Fatal("local and domain ceilings must fit within the address hard bound")
	}
}

func TestNormalizeAddress_EnforcesDisplayNameRuneAndByteBounds(t *testing.T) {
	name640Bytes := strings.Repeat("\U0001f600", 160)
	if len(name640Bytes) != mail.MaxEngineAddressNameBytes {
		t.Fatalf("display-name fixture byte length = %d, want %d", len(name640Bytes), mail.MaxEngineAddressNameBytes)
	}
	if _, err := mail.NormalizeAddress(mail.EngineAddress{Email: "alice@example.test", Name: name640Bytes}); err != nil {
		t.Fatalf("NormalizeAddress(max display name) unexpected error = %v", err)
	}
	requireAddressError(t, mail.EngineAddress{Email: "alice@example.test", Name: strings.Repeat("\U0001f600", 161)}, mail.AddressErrorTooLarge, mail.ErrAddressTooLarge)
	requireAddressError(t, mail.EngineAddress{Email: "alice@example.test", Name: string([]byte{0xff})}, mail.AddressErrorInvalid, mail.ErrAddressInvalid)
}

func TestNormalizeAddress_RejectsUntrustedMailboxSyntaxAndInjection(t *testing.T) {
	for _, input := range []mail.EngineAddress{
		{Email: ""},
		{Email: string([]byte{0xff}) + "@example.test"},
		{Email: " alice@example.test"},
		{Email: "alice@example.test "},
		{Email: "alice@\u00a0example.test"},
		{Email: "\u00fc@example.test"},
		{Email: "\"alice\"@example.test"},
		{Email: "alice(comment)@example.test"},
		{Email: "Alice <alice@example.test>"},
		{Email: ".alice@example.test"},
		{Email: "alice.@example.test"},
		{Email: "alice..smith@example.test"},
		{Email: "alice@example..test"},
		{Email: "alice@example.test."},
		{Email: "alice@-example.test"},
		{Email: "alice@example-.test"},
		{Email: "alice@exa_mple.test"},
		{Email: "alice@example.test\r\nBcc: victim@example.test"},
		{Email: "alice@example.test\x00"},
		{Email: "alice@example.test\x7f"},
		{Email: "alice@example.test\u0085"},
		{Email: "alice@@example.test"},
		{Email: "alice example@example.test"},
		{Email: "alice@example.test", Name: "Alice\r\nBcc: victim@example.test"},
		{Email: "alice@example.test", Name: "Alice\x00"},
		{Email: "alice@example.test", Name: "Alice\u0001"},
		{Email: "alice@example.test", Name: "Alice\x7f"},
		{Email: "alice@example.test", Name: "Alice\u0085"},
	} {
		requireAddressError(t, input, mail.AddressErrorInvalid, mail.ErrAddressInvalid)
	}
}

func TestNormalizeAddress_ErrorValuesAreTypedStableAndContentFree(t *testing.T) {
	for _, assertion := range []struct {
		input mail.EngineAddress
		kind  mail.AddressErrorKind
		sent  error
		text  string
	}{
		{mail.EngineAddress{Email: "invalid\r\nsecret@example.test"}, mail.AddressErrorInvalid, mail.ErrAddressInvalid, "mail address invalid"},
		{mail.EngineAddress{Email: strings.Repeat("a", 65) + "@example.test"}, mail.AddressErrorTooLarge, mail.ErrAddressTooLarge, "mail address exceeds bounds"},
	} {
		err := normalizeAddressError(t, assertion.input, assertion.kind, assertion.sent)
		if got := err.Error(); got != assertion.text {
			t.Errorf("NormalizeAddress(%#v).Error() = %q, want %q", assertion.input, got, assertion.text)
		}
		if strings.Contains(err.Error(), assertion.input.Email) || strings.Contains(err.Error(), assertion.input.Name) {
			t.Errorf("NormalizeAddress(%#v).Error() leaked input: %q", assertion.input, err.Error())
		}
	}

	unknown := &mail.AddressPolicyError{Kind: mail.AddressErrorKind("unknown synthetic-secret@example.test")}
	if got := unknown.Error(); got != "mail address policy error" {
		t.Fatalf("unknown AddressPolicyError.Error() = %q, want fixed generic error", got)
	}
	if errors.Unwrap(unknown) != nil {
		t.Fatal("unknown AddressPolicyError must not unwrap")
	}
}

func TestAddressPolicyError_LiteralsSentinelsAndSafeTextAreStable(t *testing.T) {
	for _, assertion := range []struct {
		kind     mail.AddressErrorKind
		literal  string
		sentinel error
		text     string
	}{
		{mail.AddressErrorInvalid, "invalid", mail.ErrAddressInvalid, "mail address invalid"},
		{mail.AddressErrorTooLarge, "too_large", mail.ErrAddressTooLarge, "mail address exceeds bounds"},
		{mail.AddressErrorDuplicate, "duplicate", mail.ErrAddressDuplicate, "mail address duplicate"},
	} {
		if got := string(assertion.kind); got != assertion.literal {
			t.Errorf("address error kind = %q, want external literal %q", got, assertion.literal)
		}
		err := &mail.AddressPolicyError{Kind: mail.AddressErrorKind(assertion.literal)}
		if !errors.Is(err, assertion.sentinel) {
			t.Errorf("AddressPolicyError{%q} does not match expected sentinel", assertion.literal)
		}
		if got := err.Error(); got != assertion.text {
			t.Errorf("AddressPolicyError{%q}.Error() = %q, want %q", assertion.literal, got, assertion.text)
		}
	}
}

func TestAddressPolicyError_NilReceiverIsSafeAndDoesNotUnwrap(t *testing.T) {
	var nilErr *mail.AddressPolicyError
	if got := nilErr.Error(); got != "mail address policy error" {
		t.Fatalf("nil AddressPolicyError.Error() = %q, want fixed generic error", got)
	}
	if got := nilErr.Unwrap(); got != nil {
		t.Fatalf("nil AddressPolicyError.Unwrap() = %v, want nil", got)
	}
	if got := errors.Unwrap(nilErr); got != nil {
		t.Fatalf("errors.Unwrap(nil AddressPolicyError) = %v, want nil", got)
	}
}

func TestNormalizeAddressSet_NormalizesInOrderWithoutMutatingInput(t *testing.T) {
	input := []mail.EngineAddress{
		{Email: "SECOND@EXAMPLE.TEST", Name: "Second"},
		{Email: "first@B\u00dcCHER.example", Name: " First "},
	}
	want := []mail.EngineAddress{
		{Email: "second@example.test", Name: "Second"},
		{Email: "first@xn--bcher-kva.example", Name: " First "},
	}
	got, err := mail.NormalizeAddressSet(input, 2)
	if err != nil {
		t.Fatalf("NormalizeAddressSet() unexpected error = %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("NormalizeAddressSet() length = %d, want %d", len(got), len(want))
	}
	for index := range want {
		if got[index] != want[index] {
			t.Fatalf("NormalizeAddressSet()[%d] = %#v, want %#v", index, got[index], want[index])
		}
	}
	if input[0].Email != "SECOND@EXAMPLE.TEST" || input[1].Email != "first@B\u00dcCHER.example" {
		t.Fatal("NormalizeAddressSet() must not mutate caller input")
	}
}

func TestNormalizeAddressSet_ReturnsConcreteEmptySet(t *testing.T) {
	got, err := mail.NormalizeAddressSet(nil, 1)
	if err != nil {
		t.Fatalf("NormalizeAddressSet(nil) unexpected error = %v", err)
	}
	if got == nil || len(got) != 0 {
		t.Fatalf("NormalizeAddressSet(nil) = %#v, want concrete empty set", got)
	}
}

func TestNormalizeAddressSet_RejectsInvalidLimitsAndOvercountWithoutPartialResult(t *testing.T) {
	input := []mail.EngineAddress{{Email: "one@example.test"}, {Email: "two@example.test"}}
	for _, maximum := range []int{0, -1, mail.MaxEngineAddressItems + 1} {
		got, err := mail.NormalizeAddressSet(input, maximum)
		if !errors.Is(err, mail.ErrAddressTooLarge) {
			t.Errorf("NormalizeAddressSet(maximum=%d) error = %v, want ErrAddressTooLarge", maximum, err)
		}
		var typed *mail.AddressPolicyError
		if !errors.As(err, &typed) || typed.Kind != mail.AddressErrorTooLarge {
			t.Errorf("NormalizeAddressSet(maximum=%d) typed error = %#v, want too-large AddressPolicyError", maximum, err)
		}
		if err != nil && err.Error() != "mail address exceeds bounds" {
			t.Errorf("NormalizeAddressSet(maximum=%d) error = %q, want fixed safe text", maximum, err.Error())
		}
		if got != nil {
			t.Errorf("NormalizeAddressSet(maximum=%d) result = %#v, want nil", maximum, got)
		}
	}

	got, err := mail.NormalizeAddressSet(input, 1)
	if !errors.Is(err, mail.ErrAddressTooLarge) {
		t.Fatalf("NormalizeAddressSet(overcount) error = %v, want ErrAddressTooLarge", err)
	}
	if got != nil {
		t.Fatalf("NormalizeAddressSet(overcount) result = %#v, want nil", got)
	}

	got, err = mail.NormalizeAddressSet([]mail.EngineAddress{{Email: "one@example.test"}, {Email: "bad@@example.test"}}, 2)
	if !errors.Is(err, mail.ErrAddressInvalid) {
		t.Fatalf("NormalizeAddressSet(invalid element) error = %v, want ErrAddressInvalid", err)
	}
	if got != nil {
		t.Fatalf("NormalizeAddressSet(invalid element) result = %#v, want nil", got)
	}
	if got := err.Error(); got != "mail address invalid" {
		t.Fatalf("NormalizeAddressSet(invalid element) error = %q, want fixed safe text", got)
	}
	if strings.Contains(err.Error(), "bad@@example.test") {
		t.Fatalf("NormalizeAddressSet(invalid element) error leaked input: %q", err.Error())
	}
	var typed *mail.AddressPolicyError
	if !errors.As(err, &typed) || typed.Kind != mail.AddressErrorInvalid {
		t.Fatalf("NormalizeAddressSet(invalid element) typed error = %#v, want invalid AddressPolicyError", err)
	}
}

func TestNormalizeAddressSet_RejectsCanonicalDuplicatesWithoutPartialResult(t *testing.T) {
	input := []mail.EngineAddress{
		{Email: "Alice@B\u00dcCHER.example", Name: "First"},
		{Email: "alice@xn--bcher-kva.example", Name: "Second"},
	}
	got, err := mail.NormalizeAddressSet(input, mail.MaxEngineAddressItems)
	if !errors.Is(err, mail.ErrAddressDuplicate) {
		t.Fatalf("NormalizeAddressSet() error = %v, want ErrAddressDuplicate", err)
	}
	var typed *mail.AddressPolicyError
	if !errors.As(err, &typed) || typed.Kind != mail.AddressErrorDuplicate {
		t.Fatalf("NormalizeAddressSet() typed error = %#v, want duplicate AddressPolicyError", err)
	}
	if got != nil {
		t.Fatalf("NormalizeAddressSet() result = %#v, want nil", got)
	}
	if got := err.Error(); got != "mail address duplicate" {
		t.Fatalf("duplicate error = %q, want fixed safe text", got)
	}
	if strings.Contains(err.Error(), "Alice@") || strings.Contains(err.Error(), "Second") {
		t.Fatalf("duplicate error leaked input: %q", err.Error())
	}
}

func TestNormalizeAddressSet_EnforcesExactCollectionBoundary(t *testing.T) {
	accepted := make([]mail.EngineAddress, mail.MaxEngineAddressItems)
	for index := range accepted {
		accepted[index] = mail.EngineAddress{Email: "u-" + decimalAddressIndex(index) + "@example.test"}
	}
	got, err := mail.NormalizeAddressSet(accepted, mail.MaxEngineAddressItems)
	if err != nil {
		t.Fatalf("NormalizeAddressSet(max items) unexpected error = %v", err)
	}
	if len(got) != mail.MaxEngineAddressItems {
		t.Fatalf("NormalizeAddressSet(max items) length = %d, want %d", len(got), mail.MaxEngineAddressItems)
	}
	for index := range got {
		if got[index].Email != accepted[index].Email {
			t.Fatalf("NormalizeAddressSet(max items)[%d] = %#v, want order-preserved %#v", index, got[index], accepted[index])
		}
	}

	over := append(append([]mail.EngineAddress(nil), accepted...), mail.EngineAddress{Email: "u-200@example.test"})
	got, err = mail.NormalizeAddressSet(over, mail.MaxEngineAddressItems)
	if !errors.Is(err, mail.ErrAddressTooLarge) {
		t.Fatalf("NormalizeAddressSet(201 items) error = %v, want ErrAddressTooLarge", err)
	}
	if got != nil {
		t.Fatalf("NormalizeAddressSet(201 items) result = %#v, want nil", got)
	}
	if got := err.Error(); got != "mail address exceeds bounds" {
		t.Fatalf("201-item error = %q, want fixed safe text", got)
	}
	var typed *mail.AddressPolicyError
	if !errors.As(err, &typed) || typed.Kind != mail.AddressErrorTooLarge {
		t.Fatalf("NormalizeAddressSet(201 items) typed error = %#v, want too-large AddressPolicyError", err)
	}
}

func requireAddressError(t *testing.T, input mail.EngineAddress, wantKind mail.AddressErrorKind, wantSentinel error) {
	t.Helper()
	_ = normalizeAddressError(t, input, wantKind, wantSentinel)
}

func normalizeAddressError(t *testing.T, input mail.EngineAddress, wantKind mail.AddressErrorKind, wantSentinel error) error {
	t.Helper()
	got, err := mail.NormalizeAddress(input)
	if err == nil {
		t.Fatalf("NormalizeAddress(%#v) returned nil error; want %q", input, wantKind)
	}
	if got != (mail.EngineAddress{}) {
		t.Fatalf("NormalizeAddress(%#v) result = %#v, want zero address on error", input, got)
	}
	if !errors.Is(err, wantSentinel) {
		t.Fatalf("NormalizeAddress(%#v) error = %v, want errors.Is(%v)", input, err, wantSentinel)
	}
	var typed *mail.AddressPolicyError
	if !errors.As(err, &typed) || typed.Kind != wantKind {
		t.Fatalf("NormalizeAddress(%#v) typed error = %#v, want kind %q", input, err, wantKind)
	}
	if wantKind == mail.AddressErrorTooLarge && err.Error() != "mail address exceeds bounds" {
		t.Fatalf("NormalizeAddress(%#v) error = %q, want fixed too-large text", input, err.Error())
	}
	return err
}

func decimalAddressIndex(value int) string {
	if value == 0 {
		return "0"
	}
	var digits [3]byte
	position := len(digits)
	for value > 0 {
		position--
		digits[position] = byte('0' + value%10)
		value /= 10
	}
	return string(digits[position:])
}
