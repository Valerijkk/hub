package mail_test

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestThreadPublicContract_ExportsExactShapesAndLiterals(t *testing.T) {
	keyType := reflect.TypeOf(mail.ThreadKey{})
	if keyType.Kind() != reflect.Array || keyType.Len() != 32 || keyType.Elem().Kind() != reflect.Uint8 {
		t.Fatalf("ThreadKey type = %v, want [32]byte", keyType)
	}
	if versionType := reflect.TypeOf(mail.ThreadKeyVersionV1); versionType.Kind() != reflect.Uint8 {
		t.Fatalf("ThreadKeyVersion underlying kind = %v, want uint8", versionType.Kind())
	}
	if got := uint8(mail.ThreadKeyVersionV1); got != 1 {
		t.Fatalf("ThreadKeyVersionV1 = %d, want external literal 1", got)
	}

	requireThreadStructFields(t, reflect.TypeOf(mail.ThreadAncestor{}), []threadField{
		{"MessageID", reflect.TypeOf("")},
		{"KeyVersion", reflect.TypeOf(mail.ThreadKeyVersion(0))},
		{"Key", reflect.TypeOf(mail.ThreadKey{})},
	})
	requireThreadStructFields(t, reflect.TypeOf(mail.ThreadingInput{}), []threadField{
		{"References", reflect.TypeOf([]string(nil))},
		{"InReplyTo", reflect.TypeOf("")},
		{"Subject", reflect.TypeOf("")},
		{"From", reflect.TypeOf(mail.EngineAddress{})},
		{"To", reflect.TypeOf([]mail.EngineAddress(nil))},
		{"CC", reflect.TypeOf([]mail.EngineAddress(nil))},
		{"Ancestors", reflect.TypeOf([]mail.ThreadAncestor(nil))},
		{"AncestorLookupComplete", reflect.TypeOf(false)},
	})
	requireThreadStructFields(t, reflect.TypeOf(mail.ThreadIdentity{}), []threadField{
		{"KeyVersion", reflect.TypeOf(mail.ThreadKeyVersion(0))},
		{"Key", reflect.TypeOf(mail.ThreadKey{})},
		{"Source", reflect.TypeOf(mail.ThreadIdentitySource(""))},
		{"NormalizedSubject", reflect.TypeOf("")},
		{"CanonicalParticipants", reflect.TypeOf([]string(nil))},
	})
	requireThreadStructFields(t, reflect.TypeOf(mail.ThreadPolicyError{}), []threadField{
		{"Kind", reflect.TypeOf(mail.ThreadErrorKind(""))},
	})

	for _, assertion := range []struct {
		got  string
		want string
	}{
		{string(mail.ThreadIdentitySourceAncestor), "ancestor"},
		{string(mail.ThreadIdentitySourceFallback), "fallback"},
		{string(mail.ThreadErrorInvalid), "invalid"},
		{string(mail.ThreadErrorTooLarge), "too_large"},
	} {
		if assertion.got != assertion.want {
			t.Errorf("public literal = %q, want %q", assertion.got, assertion.want)
		}
	}
	if got := mail.ErrThreadInvalid.Error(); got != "mail thread invalid" {
		t.Fatalf("ErrThreadInvalid.Error() = %q, want fixed literal", got)
	}
	if got := mail.ErrThreadTooLarge.Error(); got != "mail thread exceeds bounds" {
		t.Fatalf("ErrThreadTooLarge.Error() = %q, want fixed literal", got)
	}
	if _, exists := keyType.MethodByName("Base64"); exists {
		t.Fatal("ThreadKey must not expose a Base64 API")
	}
	if _, exists := reflect.PointerTo(keyType).MethodByName("Base64"); exists {
		t.Fatal("*ThreadKey must not expose a Base64 API")
	}
}

func TestThreadPolicyError_IsSafeTypedAndContentFree(t *testing.T) {
	for _, assertion := range []struct {
		kind       mail.ThreadErrorKind
		wantText   string
		want       error
		notWanted  error
		literal    string
	}{
		{mail.ThreadErrorInvalid, "mail thread invalid", mail.ErrThreadInvalid, mail.ErrThreadTooLarge, "invalid"},
		{mail.ThreadErrorTooLarge, "mail thread exceeds bounds", mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "too_large"},
	} {
		t.Run(assertion.literal, func(t *testing.T) {
			if got := string(assertion.kind); got != assertion.literal {
				t.Fatalf("ThreadErrorKind = %q, want %q", got, assertion.literal)
			}
			err := &mail.ThreadPolicyError{Kind: assertion.kind}
			if got := err.Error(); got != assertion.wantText {
				t.Fatalf("ThreadPolicyError{%q}.Error() = %q, want %q", assertion.literal, got, assertion.wantText)
			}
			if got := err.Unwrap(); got != assertion.want {
				t.Fatalf("ThreadPolicyError{%q}.Unwrap() = %v, want exact sentinel %v", assertion.literal, got, assertion.want)
			}
			if !errors.Is(err, assertion.want) || errors.Is(err, assertion.notWanted) {
				t.Fatalf("ThreadPolicyError{%q} errors.Is mapping is not exclusive", assertion.literal)
			}
		})
	}

	unknown := &mail.ThreadPolicyError{Kind: mail.ThreadErrorKind("unknown secret@example.test")}
	if got := unknown.Error(); got != "mail thread policy error" {
		t.Fatalf("unknown ThreadPolicyError.Error() = %q, want fixed generic text", got)
	}
	if got := unknown.Unwrap(); got != nil {
		t.Fatalf("unknown ThreadPolicyError.Unwrap() = %v, want nil", got)
	}
	if errors.Is(unknown, mail.ErrThreadInvalid) || errors.Is(unknown, mail.ErrThreadTooLarge) {
		t.Fatal("unknown ThreadPolicyError must not match a thread sentinel")
	}
	if strings.Contains(unknown.Error(), "secret@example.test") {
		t.Fatalf("unknown ThreadPolicyError leaked detail: %q", unknown.Error())
	}

	var nilError *mail.ThreadPolicyError
	if got := nilError.Error(); got != "mail thread policy error" {
		t.Fatalf("nil ThreadPolicyError.Error() = %q, want fixed generic text", got)
	}
	if got := nilError.Unwrap(); got != nil {
		t.Fatalf("nil ThreadPolicyError.Unwrap() = %v, want nil", got)
	}
	if got := errors.Unwrap(nilError); got != nil {
		t.Fatalf("errors.Unwrap(nil ThreadPolicyError) = %v, want nil", got)
	}
}

func TestDeriveThreadIdentity_RequiresDefinitiveAncestorLookup(t *testing.T) {
	input := fallbackThreadingInput(" RE: alpha ", "ALICE@example.test")
	input.AncestorLookupComplete = false
	requireThreadFailure(t, input, mail.ThreadErrorInvalid, mail.ErrThreadInvalid, mail.ErrThreadTooLarge, "mail thread invalid")

	input.AncestorLookupComplete = true
	got, err := deriveThreadIdentity(t, input)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(complete empty lookup) error = %v", err)
	}
	assertThreadIdentity(t, got,
		threadKeyFromHex(t, "abd293687e172e7a1df63bde199adf75d910a8e231a1c4374996e69cb582f279"),
		mail.ThreadKeyVersion(1),
		mail.ThreadIdentitySource("fallback"),
		"alpha",
		[]string{"alice@example.test"},
	)
}

func TestDeriveThreadIdentity_PreflightsExactRawCountsAndByteBounds(t *testing.T) {
	base := fallbackThreadingInput("subject", "from@example.test")

	input := cloneThreadingInput(base)
	input.References = referenceIDs(100)
	if _, err := deriveThreadIdentity(t, input); err != nil {
		t.Fatalf("DeriveThreadIdentity(100 references) error = %v", err)
	}
	input.References = referenceIDs(101)
	requireThreadFailure(t, input, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")
	for _, references := range [][]string{
		make([]string, 101),
		func() []string {
			values := make([]string, 101)
			for index := range values {
				values[index] = "<malformed"
			}
			return values
		}(),
		func() []string {
			values := make([]string, 101)
			for index := range values {
				values[index] = "<duplicate@example.test>"
			}
			return values
		}(),
	} {
		input = cloneThreadingInput(base)
		input.References = references
		requireThreadFailure(t, input, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")
	}

	for _, field := range []string{"To", "CC"} {
		t.Run(field, func(t *testing.T) {
			atLimit := cloneThreadingInput(base)
			if field == "To" {
				atLimit.To = recipientAddresses("to", 200)
			} else {
				atLimit.CC = recipientAddresses("cc", 200)
			}
			if _, err := deriveThreadIdentity(t, atLimit); err != nil {
				t.Fatalf("DeriveThreadIdentity(%s=200) error = %v", field, err)
			}

			overLimit := cloneThreadingInput(base)
			if field == "To" {
				overLimit.To = recipientAddresses("to", 201)
			} else {
				overLimit.CC = recipientAddresses("cc", 201)
			}
			requireThreadFailure(t, overLimit, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")

			duplicateInvalid := cloneThreadingInput(base)
			invalidAddresses := make([]mail.EngineAddress, 201)
			for index := range invalidAddresses {
				invalidAddresses[index] = mail.EngineAddress{Email: "bad@@example.test"}
			}
			if field == "To" {
				duplicateInvalid.To = invalidAddresses
			} else {
				duplicateInvalid.CC = invalidAddresses
			}
			requireThreadFailure(t, duplicateInvalid, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")
		})
	}

	input = cloneThreadingInput(base)
	input.Ancestors = validAncestors(202)
	if _, err := deriveThreadIdentity(t, input); err != nil {
		t.Fatalf("DeriveThreadIdentity(202 ancestors) error = %v", err)
	}
	input.Ancestors = validAncestors(203)
	requireThreadFailure(t, input, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")

	input = cloneThreadingInput(base)
	input.Subject = strings.Repeat("A", 998)
	got, err := deriveThreadIdentity(t, input)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(998-byte subject) error = %v", err)
	}
	if got.NormalizedSubject != strings.Repeat("a", 998) {
		t.Fatalf("998-byte normalized subject = %q, want full-folded ASCII", got.NormalizedSubject)
	}
	input.Subject = strings.Repeat("A", 999)
	requireThreadFailure(t, input, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")
	input.Subject = strings.Repeat("a", 998) + string([]byte{0xff})
	if len(input.Subject) != 999 {
		t.Fatalf("oversize invalid UTF-8 subject fixture bytes = %d, want 999", len(input.Subject))
	}
	requireThreadFailure(t, input, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")

	input = cloneThreadingInput(base)
	input.Ancestors = []mail.ThreadAncestor{{
		MessageID:  strictMessageIDAtBytes(t, 998),
		KeyVersion: mail.ThreadKeyVersion(1),
		Key:        nonzeroThreadKey(1),
	}}
	if _, err := deriveThreadIdentity(t, input); err != nil {
		t.Fatalf("DeriveThreadIdentity(998-byte ancestor ID) error = %v", err)
	}
	input.Ancestors[0].MessageID = strictMessageIDAtBytes(t, 999)
	requireThreadFailure(t, input, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")

	input = cloneThreadingInput(base)
	input.From.Email = strings.Repeat("a", 320)
	requireAddressFailure(t, input, mail.AddressErrorInvalid, mail.ErrAddressInvalid, "mail address invalid")
	input.From.Email = strings.Repeat("a", 321)
	requireAddressFailure(t, input, mail.AddressErrorTooLarge, mail.ErrAddressTooLarge, "mail address exceeds bounds")

	input = cloneThreadingInput(base)
	input.From.Name = strings.Repeat("a", 640)
	if _, err := deriveThreadIdentity(t, input); err != nil {
		t.Fatalf("DeriveThreadIdentity(640-byte name) error = %v", err)
	}
	input.From.Name = strings.Repeat("a", 641)
	requireAddressFailure(t, input, mail.AddressErrorTooLarge, mail.ErrAddressTooLarge, "mail address exceeds bounds")
}

func TestDeriveThreadIdentity_RejectsInvalidSubjectBeforeNormalization(t *testing.T) {
	for _, subject := range []string{
		string([]byte{'a', 0xff}),
		"subject\x00with-control",
		"subject\u0085with-control",
	} {
		t.Run(strconv.Quote(subject), func(t *testing.T) {
			input := fallbackThreadingInput(subject, "alice@example.test")
			requireThreadFailure(t, input, mail.ThreadErrorInvalid, mail.ErrThreadInvalid, mail.ErrThreadTooLarge, "mail thread invalid")
		})
	}

	postFoldOverflow := fallbackThreadingInput(strings.Repeat("\u0130", 499), "alice@example.test")
	if len(postFoldOverflow.Subject) != 998 {
		t.Fatalf("post-fold fixture raw bytes = %d, want 998", len(postFoldOverflow.Subject))
	}
	requireThreadFailure(t, postFoldOverflow, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")
}

func TestDeriveThreadIdentity_StrictMessageIDGrammarPreservesBytesAndCase(t *testing.T) {
	exactID := "<Left.Part+tag@Right.Domain>"
	ancestor := mail.ThreadAncestor{MessageID: exactID, KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(7)}
	input := fallbackThreadingInput("misleading subject", "alice@example.test")
	input.References = []string{exactID}
	input.Ancestors = []mail.ThreadAncestor{ancestor}
	got, err := deriveThreadIdentity(t, input)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(valid strict Message-ID) error = %v", err)
	}
	assertAncestorIdentity(t, got, ancestor)

	caseChanged := cloneThreadingInput(input)
	caseChanged.References = []string{"<left.part+tag@right.domain>"}
	got, err = deriveThreadIdentity(t, caseChanged)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(case-changed Message-ID) error = %v", err)
	}
	assertFallbackIdentity(t, got, "misleading subject", []string{"alice@example.test"})

	atLimit := cloneThreadingInput(input)
	atLimitID := strictMessageIDAtBytes(t, 998)
	atLimitAncestor := mail.ThreadAncestor{MessageID: atLimitID, KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(8)}
	atLimit.References = []string{atLimitID}
	atLimit.Ancestors = []mail.ThreadAncestor{atLimitAncestor}
	got, err = deriveThreadIdentity(t, atLimit)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(998-byte Message-ID) error = %v", err)
	}
	assertAncestorIdentity(t, got, atLimitAncestor)

	oversize := strictMessageIDAtBytes(t, 999)
	oversizeInput := cloneThreadingInput(input)
	oversizeInput.References = []string{oversize}
	oversizeInput.Ancestors = nil
	got, err = deriveThreadIdentity(t, oversizeInput)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(oversize candidate) error = %v", err)
	}
	assertFallbackIdentity(t, got, "misleading subject", []string{"alice@example.test"})

	for _, candidate := range []string{
		"",
		"<left @right>",
		"<left@right\r\n.example>",
		"<(comment)left@right>",
		"<\"left\"@right>",
		"<left@[127.0.0.1]>",
		"<left@b\u00fccher.test>",
		"<left@right\x00.test>",
		"<left@@right>",
		"left@right",
		"<.left@right>",
		"<left.@right>",
		"<left..dot@right>",
		"<left@.right>",
		"<left@right.>",
		"<left@right..dot>",
	} {
		t.Run(strconv.Quote(candidate), func(t *testing.T) {
			invalid := fallbackThreadingInput("misleading subject", "alice@example.test")
			invalid.References = []string{candidate}
			got, err := deriveThreadIdentity(t, invalid)
			if err != nil {
				t.Fatalf("DeriveThreadIdentity(%q) error = %v", candidate, err)
			}
			assertFallbackIdentity(t, got, "misleading subject", []string{"alice@example.test"})
		})
	}
}

func TestDeriveThreadIdentity_AcceptsStrictMessageIDATextDotsAndExactCase(t *testing.T) {
	specials := "!#$%&'*+-/=?^_\u0060{|}~"
	for _, assertion := range []struct {
		name      string
		messageID string
		key       mail.ThreadKey
	}{
		{
			"all RFC5322 atext specials in both dot-atoms",
			"<A" + specials + ".Z9@Q7." + specials + "R>",
			nonzeroThreadKey(15),
		},
		{
			"ordinary case-sensitive dotted atoms",
			"<First.Last+Tag@Right.Domain>",
			nonzeroThreadKey(16),
		},
	} {
		t.Run(assertion.name, func(t *testing.T) {
			ancestor := mail.ThreadAncestor{MessageID: assertion.messageID, KeyVersion: mail.ThreadKeyVersion(1), Key: assertion.key}
			input := fallbackThreadingInput("RE: case preserved", "alice@example.test")
			input.References = []string{assertion.messageID}
			input.Ancestors = []mail.ThreadAncestor{ancestor}
			got, err := deriveThreadIdentity(t, input)
			if err != nil {
				t.Fatalf("DeriveThreadIdentity(%q) error = %v", assertion.messageID, err)
			}
			assertAncestorIdentity(t, got, ancestor)

			caseChanged := cloneThreadingInput(input)
			caseChanged.References = []string{strings.ToLower(assertion.messageID)}
			got, err = deriveThreadIdentity(t, caseChanged)
			if err != nil {
				t.Fatalf("DeriveThreadIdentity(case-changed %q) error = %v", assertion.messageID, err)
			}
			assertFallbackIdentity(t, got, "case preserved", []string{"alice@example.test"})
		})
	}
}

func TestDeriveThreadIdentity_SkipsMalformedOrUnresolvedInReplyToCandidates(t *testing.T) {
	ambiguousID := "<ambiguous-reply@example.test>"
	invalidUTF8 := string([]byte{'<', 'a', '@', 0xff, '>'})
	invalidUTF8Oversize := strings.Repeat("a", 998) + string([]byte{0xff})
	if len(invalidUTF8Oversize) != 999 {
		t.Fatalf("oversize invalid UTF-8 InReplyTo fixture bytes = %d, want 999", len(invalidUTF8Oversize))
	}
	for _, assertion := range []struct {
		name      string
		inReplyTo string
		ancestors []mail.ThreadAncestor
	}{
		{"malformed", "<malformed", nil},
		{"999 bytes", strictMessageIDAtBytes(t, 999), nil},
		{"invalid UTF-8", invalidUTF8, nil},
		{"999-byte invalid UTF-8", invalidUTF8Oversize, nil},
		{"unknown", "<unknown-reply@example.test>", nil},
		{
			"ambiguous",
			ambiguousID,
			[]mail.ThreadAncestor{
				{MessageID: ambiguousID, KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(21)},
				{MessageID: ambiguousID, KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(22)},
			},
		},
	} {
		t.Run(assertion.name, func(t *testing.T) {
			input := fallbackThreadingInput("RE: fallback subject", "alice@example.test")
			input.InReplyTo = assertion.inReplyTo
			input.Ancestors = assertion.ancestors
			got, err := deriveThreadIdentity(t, input)
			if err != nil {
				t.Fatalf("DeriveThreadIdentity(InReplyTo=%q) error = %v", assertion.inReplyTo, err)
			}
			assertFallbackIdentity(t, got, "fallback subject", []string{"alice@example.test"})
		})
	}
}

func TestDeriveThreadIdentity_UsesNearestUniqueReferenceThenInReplyToThenFallback(t *testing.T) {
	old := mail.ThreadAncestor{MessageID: "<old@example.test>", KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(1)}
	nearest := mail.ThreadAncestor{MessageID: "<nearest@example.test>", KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(2)}
	inReplyTo := mail.ThreadAncestor{MessageID: "<reply@example.test>", KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(3)}
	input := fallbackThreadingInput("RE: misleading subject", "alice@example.test")
	input.References = []string{old.MessageID, nearest.MessageID}
	input.InReplyTo = inReplyTo.MessageID
	input.Ancestors = []mail.ThreadAncestor{old, nearest, inReplyTo}
	got, err := deriveThreadIdentity(t, input)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(nearest reference) error = %v", err)
	}
	assertAncestorIdentity(t, got, nearest)

	input = fallbackThreadingInput("RE: misleading subject", "alice@example.test")
	input.References = []string{"<unknown@example.test>"}
	input.InReplyTo = inReplyTo.MessageID
	input.Ancestors = []mail.ThreadAncestor{inReplyTo}
	got, err = deriveThreadIdentity(t, input)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(InReplyTo fallback) error = %v", err)
	}
	assertAncestorIdentity(t, got, inReplyTo)

	input = fallbackThreadingInput("RE: misleading subject", "alice@example.test")
	input.References = []string{"<malformed", strictMessageIDAtBytes(t, 999), "", "<unknown@example.test>"}
	input.InReplyTo = inReplyTo.MessageID
	input.Ancestors = []mail.ThreadAncestor{inReplyTo}
	got, err = deriveThreadIdentity(t, input)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(skippable references) error = %v", err)
	}
	assertAncestorIdentity(t, got, inReplyTo)

	input = fallbackThreadingInput("RE: misleading subject", "alice@example.test")
	duplicate := mail.ThreadAncestor{MessageID: "<duplicate@example.test>", KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(4)}
	input.References = []string{duplicate.MessageID}
	input.Ancestors = []mail.ThreadAncestor{duplicate, duplicate}
	got, err = deriveThreadIdentity(t, input)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(identical duplicate ancestors) error = %v", err)
	}
	assertAncestorIdentity(t, got, duplicate)

	input = fallbackThreadingInput("RE: misleading subject", "alice@example.test")
	ambiguousID := "<ambiguous@example.test>"
	input.References = []string{ambiguousID}
	input.InReplyTo = inReplyTo.MessageID
	input.Ancestors = []mail.ThreadAncestor{
		{MessageID: ambiguousID, KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(5)},
		{MessageID: ambiguousID, KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(6)},
		inReplyTo,
	}
	got, err = deriveThreadIdentity(t, input)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(ambiguous reference) error = %v", err)
	}
	assertAncestorIdentity(t, got, inReplyTo)

	input = fallbackThreadingInput("RE: misleading subject", "alice@example.test")
	input.References = []string{ambiguousID}
	input.Ancestors = []mail.ThreadAncestor{
		{MessageID: ambiguousID, KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(5)},
		{MessageID: ambiguousID, KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(6)},
	}
	got, err = deriveThreadIdentity(t, input)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(ambiguous only) error = %v", err)
	}
	assertFallbackIdentity(t, got, "misleading subject", []string{"alice@example.test"})
}

func TestDeriveThreadIdentity_ContinuesRightToLeftPastSkippableReference(t *testing.T) {
	older := mail.ThreadAncestor{MessageID: "<older@example.test>", KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(31)}
	inReplyTo := mail.ThreadAncestor{MessageID: "<reply-after-reference@example.test>", KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(32)}
	ambiguousID := "<ambiguous-after-reference@example.test>"
	invalidUTF8 := string([]byte{'<', 'a', '@', 0xff, '>'})
	invalidUTF8Oversize := strings.Repeat("a", 998) + string([]byte{0xff})
	if len(invalidUTF8Oversize) != 999 {
		t.Fatalf("oversize invalid UTF-8 reference fixture bytes = %d, want 999", len(invalidUTF8Oversize))
	}

	for _, assertion := range []struct {
		name      string
		rightmost string
		extra     []mail.ThreadAncestor
	}{
		{"malformed", "<malformed", nil},
		{"999 bytes", strictMessageIDAtBytes(t, 999), nil},
		{"invalid UTF-8", invalidUTF8, nil},
		{"999-byte invalid UTF-8", invalidUTF8Oversize, nil},
		{"unknown", "<unknown-after-reference@example.test>", nil},
		{
			"ambiguous",
			ambiguousID,
			[]mail.ThreadAncestor{
				{MessageID: ambiguousID, KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(33)},
				{MessageID: ambiguousID, KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(34)},
			},
		},
	} {
		t.Run(assertion.name, func(t *testing.T) {
			input := fallbackThreadingInput("RE: must not fallback", "alice@example.test")
			input.References = []string{older.MessageID, assertion.rightmost}
			input.InReplyTo = inReplyTo.MessageID
			input.Ancestors = append([]mail.ThreadAncestor{older, inReplyTo}, assertion.extra...)
			got, err := deriveThreadIdentity(t, input)
			if err != nil {
				t.Fatalf("DeriveThreadIdentity(rightmost=%q) error = %v", assertion.rightmost, err)
			}
			assertAncestorIdentity(t, got, older)
		})
	}
}

func TestDeriveThreadIdentity_RejectsInvalidAncestorRecords(t *testing.T) {
	base := fallbackThreadingInput("subject", "alice@example.test")
	for _, ancestor := range []mail.ThreadAncestor{
		{MessageID: "malformed", KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(1)},
		{MessageID: "<zero@example.test>", KeyVersion: mail.ThreadKeyVersion(1)},
		{MessageID: "<zero-version@example.test>", KeyVersion: mail.ThreadKeyVersion(0), Key: nonzeroThreadKey(2)},
		{MessageID: "<v2@example.test>", KeyVersion: mail.ThreadKeyVersion(2), Key: nonzeroThreadKey(3)},
	} {
		input := cloneThreadingInput(base)
		input.Ancestors = []mail.ThreadAncestor{ancestor}
		requireThreadFailure(t, input, mail.ThreadErrorInvalid, mail.ErrThreadInvalid, mail.ErrThreadTooLarge, "mail thread invalid")
	}
}

func TestDeriveThreadIdentity_ValidatesAllAncestorsBeforeSelectingMatch(t *testing.T) {
	matching := mail.ThreadAncestor{MessageID: "<matching@example.test>", KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(41)}
	for _, assertion := range []struct {
		name         string
		unrelated    mail.ThreadAncestor
		kind         mail.ThreadErrorKind
		sentinel     error
		otherSentinel error
		text         string
	}{
		{
			"malformed Message-ID",
			mail.ThreadAncestor{MessageID: "malformed", KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(42)},
			mail.ThreadErrorInvalid, mail.ErrThreadInvalid, mail.ErrThreadTooLarge, "mail thread invalid",
		},
		{
			"oversize Message-ID",
			mail.ThreadAncestor{MessageID: strictMessageIDAtBytes(t, 999), KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(43)},
			mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds",
		},
		{
			"zero key",
			mail.ThreadAncestor{MessageID: "<zero-key@example.test>", KeyVersion: mail.ThreadKeyVersion(1)},
			mail.ThreadErrorInvalid, mail.ErrThreadInvalid, mail.ErrThreadTooLarge, "mail thread invalid",
		},
		{
			"non-V1 version",
			mail.ThreadAncestor{MessageID: "<v2-key@example.test>", KeyVersion: mail.ThreadKeyVersion(2), Key: nonzeroThreadKey(44)},
			mail.ThreadErrorInvalid, mail.ErrThreadInvalid, mail.ErrThreadTooLarge, "mail thread invalid",
		},
	} {
		t.Run(assertion.name, func(t *testing.T) {
			input := fallbackThreadingInput("RE: should not select", "alice@example.test")
			input.References = []string{matching.MessageID}
			input.Ancestors = []mail.ThreadAncestor{matching, assertion.unrelated}
			requireThreadFailure(t, input, assertion.kind, assertion.sentinel, assertion.otherSentinel, assertion.text)
		})
	}

	unmatched := fallbackThreadingInput("RE: fallback after complete miss", "alice@example.test")
	unmatched.Ancestors = []mail.ThreadAncestor{{
		MessageID:  "<valid-but-unmatched@example.test>",
		KeyVersion: mail.ThreadKeyVersion(1),
		Key:        nonzeroThreadKey(45),
	}}
	got, err := deriveThreadIdentity(t, unmatched)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(valid unmatched ancestor) error = %v", err)
	}
	assertFallbackIdentity(t, got, "fallback after complete miss", []string{"alice@example.test"})
}

func TestDeriveThreadIdentity_ValidatesFallbackInputsBeforeReturningMatchingAncestor(t *testing.T) {
	matching := mail.ThreadAncestor{MessageID: "<prevalidation-match@example.test>", KeyVersion: mail.ThreadKeyVersion(1), Key: nonzeroThreadKey(51)}
	matchingInput := func() mail.ThreadingInput {
		input := fallbackThreadingInput("subject", "from@example.test")
		input.References = []string{matching.MessageID}
		input.Ancestors = []mail.ThreadAncestor{matching}
		return input
	}

	for _, assertion := range []struct {
		name          string
		subject       string
		kind          mail.ThreadErrorKind
		sentinel      error
		otherSentinel error
		text          string
	}{
		{"invalid UTF-8 subject", string([]byte{'a', 0xff}), mail.ThreadErrorInvalid, mail.ErrThreadInvalid, mail.ErrThreadTooLarge, "mail thread invalid"},
		{"control subject", "subject\x00with-control", mail.ThreadErrorInvalid, mail.ErrThreadInvalid, mail.ErrThreadTooLarge, "mail thread invalid"},
		{"999-byte subject", strings.Repeat("a", 999), mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds"},
	} {
		t.Run(assertion.name, func(t *testing.T) {
			input := matchingInput()
			input.Subject = assertion.subject
			requireThreadFailure(t, input, assertion.kind, assertion.sentinel, assertion.otherSentinel, assertion.text)
		})
	}

	incomplete := matchingInput()
	incomplete.AncestorLookupComplete = false
	requireThreadFailure(t, incomplete, mail.ThreadErrorInvalid, mail.ErrThreadInvalid, mail.ErrThreadTooLarge, "mail thread invalid")

	referencesOverLimit := matchingInput()
	referencesOverLimit.References = append(referenceIDs(100), matching.MessageID)
	requireThreadFailure(t, referencesOverLimit, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")

	toOverLimit := matchingInput()
	toOverLimit.To = recipientAddresses("to", 201)
	requireThreadFailure(t, toOverLimit, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")

	ccOverLimit := matchingInput()
	ccOverLimit.CC = recipientAddresses("cc", 201)
	requireThreadFailure(t, ccOverLimit, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")

	ancestorsOverLimit := matchingInput()
	ancestorsOverLimit.Ancestors = append(validAncestors(202), matching)
	requireThreadFailure(t, ancestorsOverLimit, mail.ThreadErrorTooLarge, mail.ErrThreadTooLarge, mail.ErrThreadInvalid, "mail thread exceeds bounds")

	for _, assertion := range []struct {
		name     string
		mutate   func(*mail.ThreadingInput)
		kind     mail.AddressErrorKind
		sentinel error
		text     string
	}{
		{
			"invalid From",
			func(input *mail.ThreadingInput) { input.From.Email = "bad@@example.test" },
			mail.AddressErrorInvalid, mail.ErrAddressInvalid, "mail address invalid",
		},
		{
			"oversize From",
			func(input *mail.ThreadingInput) { input.From.Email = strings.Repeat("a", 321) },
			mail.AddressErrorTooLarge, mail.ErrAddressTooLarge, "mail address exceeds bounds",
		},
		{
			"invalid From name",
			func(input *mail.ThreadingInput) { input.From.Name = string([]byte{0xff}) },
			mail.AddressErrorInvalid, mail.ErrAddressInvalid, "mail address invalid",
		},
		{
			"oversize From name",
			func(input *mail.ThreadingInput) { input.From.Name = strings.Repeat("a", 641) },
			mail.AddressErrorTooLarge, mail.ErrAddressTooLarge, "mail address exceeds bounds",
		},
		{
			"invalid To element",
			func(input *mail.ThreadingInput) { input.To = []mail.EngineAddress{{Email: "bad@@example.test"}} },
			mail.AddressErrorInvalid, mail.ErrAddressInvalid, "mail address invalid",
		},
		{
			"oversize To element",
			func(input *mail.ThreadingInput) { input.To = []mail.EngineAddress{{Email: strings.Repeat("a", 321)}} },
			mail.AddressErrorTooLarge, mail.ErrAddressTooLarge, "mail address exceeds bounds",
		},
		{
			"invalid To name",
			func(input *mail.ThreadingInput) { input.To = []mail.EngineAddress{{Email: "to@example.test", Name: string([]byte{0xff})}} },
			mail.AddressErrorInvalid, mail.ErrAddressInvalid, "mail address invalid",
		},
		{
			"oversize To name",
			func(input *mail.ThreadingInput) { input.To = []mail.EngineAddress{{Email: "to@example.test", Name: strings.Repeat("a", 641)}} },
			mail.AddressErrorTooLarge, mail.ErrAddressTooLarge, "mail address exceeds bounds",
		},
		{
			"invalid CC element",
			func(input *mail.ThreadingInput) { input.CC = []mail.EngineAddress{{Email: "bad@@example.test"}} },
			mail.AddressErrorInvalid, mail.ErrAddressInvalid, "mail address invalid",
		},
		{
			"oversize CC element",
			func(input *mail.ThreadingInput) { input.CC = []mail.EngineAddress{{Email: strings.Repeat("a", 321)}} },
			mail.AddressErrorTooLarge, mail.ErrAddressTooLarge, "mail address exceeds bounds",
		},
		{
			"invalid CC name",
			func(input *mail.ThreadingInput) { input.CC = []mail.EngineAddress{{Email: "cc@example.test", Name: string([]byte{0xff})}} },
			mail.AddressErrorInvalid, mail.ErrAddressInvalid, "mail address invalid",
		},
		{
			"oversize CC name",
			func(input *mail.ThreadingInput) { input.CC = []mail.EngineAddress{{Email: "cc@example.test", Name: strings.Repeat("a", 641)}} },
			mail.AddressErrorTooLarge, mail.ErrAddressTooLarge, "mail address exceeds bounds",
		},
	} {
		t.Run(assertion.name, func(t *testing.T) {
			input := matchingInput()
			assertion.mutate(&input)
			requireAddressFailure(t, input, assertion.kind, assertion.sentinel, assertion.text)
		})
	}
}

func TestDeriveThreadIdentity_AncestorIdentityIsExactConcreteAndDoesNotMutateInput(t *testing.T) {
	ancestor := mail.ThreadAncestor{
		MessageID:  "<parent@example.test>",
		KeyVersion: mail.ThreadKeyVersion(1),
		Key:        threadKeyFromHex(t, "0102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f20"),
	}
	input := fallbackThreadingInput("RE: ignored subject", "alice@example.test")
	input.References = []string{ancestor.MessageID}
	input.Ancestors = []mail.ThreadAncestor{ancestor}
	before := cloneThreadingInput(input)
	got, err := deriveThreadIdentity(t, input)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(ancestor) error = %v", err)
	}
	assertAncestorIdentity(t, got, ancestor)
	if got.CanonicalParticipants == nil {
		t.Fatal("ancestor participants must be a concrete empty slice")
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("DeriveThreadIdentity mutated caller input: got %#v, want %#v", input, before)
	}
}

func TestDeriveThreadIdentity_FallbackUsesOnlyCanonicalFromToCCSeed(t *testing.T) {
	input := fallbackThreadingInput("RE: Stable Subject", "ZED@EXAMPLE.TEST")
	input.From.Name = "Ignored sender display name"
	input.To = []mail.EngineAddress{
		{Email: "Alice@B\u00dcCHER.example", Name: "Ignored first name"},
		{Email: "zed@example.test", Name: "Duplicate sender"},
	}
	input.CC = []mail.EngineAddress{
		{Email: "BOB@EXAMPLE.TEST", Name: "Ignored CC name"},
		{Email: "alice@xn--bcher-kva.example", Name: "Duplicate recipient"},
	}
	got, err := deriveThreadIdentity(t, input)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(canonical participants) error = %v", err)
	}
	wantParticipants := []string{"alice@xn--bcher-kva.example", "bob@example.test", "zed@example.test"}
	assertFallbackIdentity(t, got, "stable subject", wantParticipants)

	permuted := fallbackThreadingInput("re: stable subject", "zed@example.test")
	permuted.To = []mail.EngineAddress{
		{Email: "BOB@example.test"},
		{Email: "ALICE@xn--bcher-kva.example"},
	}
	permuted.CC = []mail.EngineAddress{
		{Email: "ZED@EXAMPLE.TEST"},
		{Email: "alice@B\u00dcCHER.example"},
	}
	again, err := deriveThreadIdentity(t, permuted)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(permuted canonical participants) error = %v", err)
	}
	assertFallbackIdentity(t, again, "stable subject", wantParticipants)
	if got.Key != again.Key {
		t.Fatalf("canonical participant permutations produced different keys: %x != %x", got.Key, again.Key)
	}

	got.CanonicalParticipants[0] = "tampered@example.test"
	if input.From.Email != "ZED@EXAMPLE.TEST" || input.To[0].Email != "Alice@B\u00dcCHER.example" {
		t.Fatal("mutating returned participants must not mutate caller input")
	}
	again, err = deriveThreadIdentity(t, input)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(after output mutation) error = %v", err)
	}
	assertFallbackIdentity(t, again, "stable subject", wantParticipants)

	maximum := fallbackThreadingInput("subject", "from@example.test")
	maximum.To = recipientAddresses("to", 200)
	maximum.CC = recipientAddresses("cc", 200)
	got, err = deriveThreadIdentity(t, maximum)
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(401 participants) error = %v", err)
	}
	if got.CanonicalParticipants == nil || len(got.CanonicalParticipants) != 401 {
		t.Fatalf("401-participant fallback = %#v, want concrete 401-item snapshot", got.CanonicalParticipants)
	}
}

func TestDeriveThreadIdentity_PropagatesAddressPolicyErrorsWithoutPartialIdentity(t *testing.T) {
	for _, assertion := range []struct {
		name     string
		mutate   func(*mail.ThreadingInput)
		kind     mail.AddressErrorKind
		sentinel error
		text     string
	}{
		{
			name: "invalid email",
			mutate: func(input *mail.ThreadingInput) { input.From.Email = "bad@@example.test" },
			kind: mail.AddressErrorInvalid, sentinel: mail.ErrAddressInvalid, text: "mail address invalid",
		},
		{
			name: "oversize email",
			mutate: func(input *mail.ThreadingInput) { input.From.Email = strings.Repeat("a", 321) },
			kind: mail.AddressErrorTooLarge, sentinel: mail.ErrAddressTooLarge, text: "mail address exceeds bounds",
		},
		{
			name: "oversize name",
			mutate: func(input *mail.ThreadingInput) { input.From.Name = strings.Repeat("a", 641) },
			kind: mail.AddressErrorTooLarge, sentinel: mail.ErrAddressTooLarge, text: "mail address exceeds bounds",
		},
	} {
		t.Run(assertion.name, func(t *testing.T) {
			input := fallbackThreadingInput("subject", "alice@example.test")
			assertion.mutate(&input)
			requireAddressFailure(t, input, assertion.kind, assertion.sentinel, assertion.text)
		})
	}
}

func TestDeriveThreadIdentity_PropagatesEveryToAndCCAddressError(t *testing.T) {
	for _, collection := range []string{"To", "CC"} {
		for _, assertion := range []struct {
			name     string
			address  mail.EngineAddress
			kind     mail.AddressErrorKind
			sentinel error
			text     string
		}{
			{"invalid email", mail.EngineAddress{Email: "bad@@example.test"}, mail.AddressErrorInvalid, mail.ErrAddressInvalid, "mail address invalid"},
			{"oversize email", mail.EngineAddress{Email: strings.Repeat("a", 321)}, mail.AddressErrorTooLarge, mail.ErrAddressTooLarge, "mail address exceeds bounds"},
			{"invalid name", mail.EngineAddress{Email: "alice@example.test", Name: string([]byte{0xff})}, mail.AddressErrorInvalid, mail.ErrAddressInvalid, "mail address invalid"},
			{"oversize name", mail.EngineAddress{Email: "alice@example.test", Name: strings.Repeat("a", 641)}, mail.AddressErrorTooLarge, mail.ErrAddressTooLarge, "mail address exceeds bounds"},
		} {
			t.Run(collection+"/"+assertion.name, func(t *testing.T) {
				input := fallbackThreadingInput("subject", "from@example.test")
				if collection == "To" {
					input.To = []mail.EngineAddress{assertion.address}
				} else {
					input.CC = []mail.EngineAddress{assertion.address}
				}
				requireAddressFailure(t, input, assertion.kind, assertion.sentinel, assertion.text)
			})
		}
	}
}

func TestDeriveThreadIdentity_EmptyFallbackParticipantsReturnAddressInvalid(t *testing.T) {
	input := mail.ThreadingInput{AncestorLookupComplete: true}
	requireAddressFailure(t, input, mail.AddressErrorInvalid, mail.ErrAddressInvalid, "mail address invalid")
}

func TestDeriveThreadIdentity_NormalizesSubjectV1Exactly(t *testing.T) {
	for _, assertion := range []struct {
		name  string
		input string
		want  string
	}{
		{"nfc full fold nfc", "\u041e\u0442\u0432\u0435\u0442:  Cafe\u0301", "caf\u00e9"},
		{"unicode spaces collapse and trim", "\u00a0 RE:\u2003Cafe\u2028X \u2009", "cafe x"},
		{"longest пересылка", "\u041f\u0435\u0440\u0435\u0441\u044b\u043b\u043a\u0430: alpha", "alpha"},
		{"пересл", "\u041f\u0435\u0440\u0435\u0441\u043b: beta", "beta"},
		{"ответ", "\u041e\u0442\u0432\u0435\u0442: gamma", "gamma"},
		{"отв", "\u041e\u0442\u0432: delta", "delta"},
		{"fwd", "Fwd: epsilon", "epsilon"},
		{"fw", "Fw: zeta", "zeta"},
		{"re", "Re: eta", "eta"},
		{"one space before ASCII colon", " Re : theta ", "theta"},
		{"empty subject", "", ""},
		{"counter remains", "RE2: alpha", "re2: alpha"},
		{"fullwidth colon remains", "Re\uff1aalpha", "re\uff1aalpha"},
		{"arbitrary token remains", "X: alpha", "x: alpha"},
	} {
		t.Run(assertion.name, func(t *testing.T) {
			got, err := deriveThreadIdentity(t, fallbackThreadingInput(assertion.input, "alice@example.test"))
			if err != nil {
				t.Fatalf("DeriveThreadIdentity(%q) error = %v", assertion.input, err)
			}
			if got.NormalizedSubject != assertion.want {
				t.Fatalf("normalized_subject_v1(%q) = %q, want %q", assertion.input, got.NormalizedSubject, assertion.want)
			}
		})
	}

	sixteen := strings.Repeat("Re: ", 16) + "Alpha"
	got, err := deriveThreadIdentity(t, fallbackThreadingInput(sixteen, "alice@example.test"))
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(16 prefixes) error = %v", err)
	}
	if got.NormalizedSubject != "alpha" {
		t.Fatalf("normalized_subject_v1(16 prefixes) = %q, want %q", got.NormalizedSubject, "alpha")
	}

	seventeen := strings.Repeat("Re: ", 17) + "Alpha"
	got, err = deriveThreadIdentity(t, fallbackThreadingInput(seventeen, "alice@example.test"))
	if err != nil {
		t.Fatalf("DeriveThreadIdentity(17 prefixes) error = %v", err)
	}
	if got.NormalizedSubject != "re: alpha" {
		t.Fatalf("normalized_subject_v1(17 prefixes) = %q, want preserved seventeenth prefix", got.NormalizedSubject)
	}
}

func TestDeriveThreadIdentity_FallbackHashVectorsUseRawVersionedBytes(t *testing.T) {
	for _, assertion := range []struct {
		name         string
		input        mail.ThreadingInput
		preimageHex  string
		wantKeyHex   string
		wantSubject  string
		participants []string
	}{
		{
			name:         "V1 alpha",
			input:        fallbackThreadingInput(" RE: alpha ", "ALICE@example.test"),
			preimageHex:  "6875622e6d61696c2e7468726561642e66616c6c6261636b000100000005616c7068610000000100000012616c696365406578616d706c652e74657374",
			wantKeyHex:   "abd293687e172e7a1df63bde199adf75d910a8e231a1c4374996e69cb582f279",
			wantSubject:  "alpha",
			participants: []string{"alice@example.test"},
		},
		{
			name: "V2 cafe",
			input: mail.ThreadingInput{
				AncestorLookupComplete: true,
				Subject:                "\u041e\u0442\u0432\u0435\u0442:  Cafe\u0301",
				From:                   mail.EngineAddress{Email: "alice@example.test"},
				To:                     []mail.EngineAddress{{Email: "bob@example.test"}},
			},
			preimageHex:  "6875622e6d61696c2e7468726561642e66616c6c6261636b000100000005636166c3a90000000200000012616c696365406578616d706c652e7465737400000010626f62406578616d706c652e74657374",
			wantKeyHex:   "42890e72238d36dec301d71ac6398623d417a56140609b903d7e6527a693e969",
			wantSubject:  "caf\u00e9",
			participants: []string{"alice@example.test", "bob@example.test"},
		},
		{
			name: "V3 empty subject",
			input: mail.ThreadingInput{
				AncestorLookupComplete: true,
				From:                   mail.EngineAddress{Email: "b@x.test"},
				To:                     []mail.EngineAddress{{Email: "a@x.test"}},
			},
			preimageHex:  "6875622e6d61696c2e7468726561642e66616c6c6261636b00010000000000000002000000086140782e74657374000000086240782e74657374",
			wantKeyHex:   "be60db881374e85b6f8129b562b910cb0f95cafc2704eeb855a94ce8051a3999",
			wantSubject:  "",
			participants: []string{"a@x.test", "b@x.test"},
		},
	} {
		t.Run(assertion.name, func(t *testing.T) {
			preimage, err := hex.DecodeString(assertion.preimageHex)
			if err != nil {
				t.Fatalf("documented preimage literal is not hexadecimal: %v", err)
			}
			digest := sha256.Sum256(preimage)
			if got := hex.EncodeToString(digest[:]); got != assertion.wantKeyHex {
				t.Fatalf("documented preimage SHA-256 = %s, want fixed documented digest %s", got, assertion.wantKeyHex)
			}

			got, err := deriveThreadIdentity(t, assertion.input)
			if err != nil {
				t.Fatalf("DeriveThreadIdentity() error = %v", err)
			}
			assertThreadIdentity(t, got,
				threadKeyFromHex(t, assertion.wantKeyHex),
				mail.ThreadKeyVersion(1),
				mail.ThreadIdentitySource("fallback"),
				assertion.wantSubject,
				assertion.participants,
			)
		})
	}
}

type threadField struct {
	name string
	typ  reflect.Type
}

func requireThreadStructFields(t *testing.T, typ reflect.Type, want []threadField) {
	t.Helper()
	if typ.Kind() != reflect.Struct {
		t.Fatalf("%v kind = %v, want struct", typ, typ.Kind())
	}
	if typ.NumField() != len(want) {
		t.Fatalf("%v field count = %d, want %d", typ, typ.NumField(), len(want))
	}
	for index, field := range want {
		got := typ.Field(index)
		if got.Name != field.name || got.Type != field.typ || !got.IsExported() {
			t.Fatalf("%v field %d = %s %v exported=%v, want %s %v exported", typ, index, got.Name, got.Type, got.IsExported(), field.name, field.typ)
		}
	}
}

func fallbackThreadingInput(subject, from string) mail.ThreadingInput {
	return mail.ThreadingInput{
		AncestorLookupComplete: true,
		Subject:                subject,
		From:                   mail.EngineAddress{Email: from},
	}
}

func deriveThreadIdentity(t *testing.T, input mail.ThreadingInput) (mail.ThreadIdentity, error) {
	t.Helper()
	before := cloneThreadingInput(input)
	got, err := mail.DeriveThreadIdentity(input)
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("DeriveThreadIdentity mutated caller input: got %#v, want %#v", input, before)
	}
	return got, err
}

func cloneThreadingInput(input mail.ThreadingInput) mail.ThreadingInput {
	clone := input
	clone.References = append([]string(nil), input.References...)
	clone.To = append([]mail.EngineAddress(nil), input.To...)
	clone.CC = append([]mail.EngineAddress(nil), input.CC...)
	clone.Ancestors = append([]mail.ThreadAncestor(nil), input.Ancestors...)
	return clone
}

func requireThreadFailure(t *testing.T, input mail.ThreadingInput, wantKind mail.ThreadErrorKind, wantSentinel, otherSentinel error, wantText string) {
	t.Helper()
	got, err := deriveThreadIdentity(t, input)
	if err == nil {
		t.Fatalf("DeriveThreadIdentity(%#v) error = nil, want %q", input, wantKind)
	}
	if !reflect.DeepEqual(got, mail.ThreadIdentity{}) {
		t.Fatalf("DeriveThreadIdentity(%#v) result = %#v, want zero identity on error", input, got)
	}
	if !errors.Is(err, wantSentinel) || errors.Is(err, otherSentinel) {
		t.Fatalf("DeriveThreadIdentity(%#v) errors.Is mapping = %v, want only %v", input, err, wantSentinel)
	}
	var typed *mail.ThreadPolicyError
	if !errors.As(err, &typed) || typed.Kind != wantKind {
		t.Fatalf("DeriveThreadIdentity(%#v) typed error = %#v, want ThreadPolicyError kind %q", input, err, wantKind)
	}
	if got := err.Error(); got != wantText {
		t.Fatalf("DeriveThreadIdentity(%#v) error text = %q, want %q", input, got, wantText)
	}
}

func requireAddressFailure(t *testing.T, input mail.ThreadingInput, wantKind mail.AddressErrorKind, wantSentinel error, wantText string) {
	t.Helper()
	got, err := deriveThreadIdentity(t, input)
	if err == nil {
		t.Fatalf("DeriveThreadIdentity(%#v) error = nil, want address error %q", input, wantKind)
	}
	if !reflect.DeepEqual(got, mail.ThreadIdentity{}) {
		t.Fatalf("DeriveThreadIdentity(%#v) result = %#v, want zero identity on address error", input, got)
	}
	if !errors.Is(err, wantSentinel) {
		t.Fatalf("DeriveThreadIdentity(%#v) error = %v, want errors.Is(%v)", input, err, wantSentinel)
	}
	var typed *mail.AddressPolicyError
	if !errors.As(err, &typed) || typed.Kind != wantKind {
		t.Fatalf("DeriveThreadIdentity(%#v) typed address error = %#v, want kind %q", input, err, wantKind)
	}
	if got := err.Error(); got != wantText {
		t.Fatalf("DeriveThreadIdentity(%#v) address error text = %q, want %q", input, got, wantText)
	}
}

func assertThreadIdentity(t *testing.T, got mail.ThreadIdentity, wantKey mail.ThreadKey, wantVersion mail.ThreadKeyVersion, wantSource mail.ThreadIdentitySource, wantSubject string, wantParticipants []string) {
	t.Helper()
	if got.KeyVersion != wantVersion {
		t.Fatalf("ThreadIdentity.KeyVersion = %d, want %d", got.KeyVersion, wantVersion)
	}
	if got.Key != wantKey {
		t.Fatalf("ThreadIdentity.Key = %x, want independently fixed %x", got.Key, wantKey)
	}
	if got.Source != wantSource {
		t.Fatalf("ThreadIdentity.Source = %q, want %q", got.Source, wantSource)
	}
	if got.NormalizedSubject != wantSubject {
		t.Fatalf("ThreadIdentity.NormalizedSubject = %q, want %q", got.NormalizedSubject, wantSubject)
	}
	if got.CanonicalParticipants == nil {
		t.Fatal("ThreadIdentity.CanonicalParticipants must be concrete")
	}
	if !reflect.DeepEqual(got.CanonicalParticipants, wantParticipants) {
		t.Fatalf("ThreadIdentity.CanonicalParticipants = %#v, want %#v", got.CanonicalParticipants, wantParticipants)
	}
}

func assertAncestorIdentity(t *testing.T, got mail.ThreadIdentity, ancestor mail.ThreadAncestor) {
	t.Helper()
	assertThreadIdentity(t, got, ancestor.Key, mail.ThreadKeyVersion(1), mail.ThreadIdentitySource("ancestor"), "", []string{})
}

func assertFallbackIdentity(t *testing.T, got mail.ThreadIdentity, wantSubject string, wantParticipants []string) {
	t.Helper()
	if got.KeyVersion != mail.ThreadKeyVersion(1) {
		t.Fatalf("fallback KeyVersion = %d, want 1", got.KeyVersion)
	}
	if got.Source != mail.ThreadIdentitySource("fallback") {
		t.Fatalf("fallback Source = %q, want %q", got.Source, "fallback")
	}
	if got.NormalizedSubject != wantSubject {
		t.Fatalf("fallback NormalizedSubject = %q, want %q", got.NormalizedSubject, wantSubject)
	}
	if got.CanonicalParticipants == nil || !reflect.DeepEqual(got.CanonicalParticipants, wantParticipants) {
		t.Fatalf("fallback CanonicalParticipants = %#v, want concrete %#v", got.CanonicalParticipants, wantParticipants)
	}
}

func threadKeyFromHex(t *testing.T, encoded string) mail.ThreadKey {
	t.Helper()
	raw, err := hex.DecodeString(encoded)
	if err != nil || len(raw) != 32 {
		t.Fatalf("independent ThreadKey literal %q is invalid: bytes=%d err=%v", encoded, len(raw), err)
	}
	var key mail.ThreadKey
	copy(key[:], raw)
	return key
}

func nonzeroThreadKey(seed byte) mail.ThreadKey {
	var key mail.ThreadKey
	key[0] = seed
	key[31] = seed ^ 0xff
	return key
}

func strictMessageIDAtBytes(t *testing.T, total int) string {
	t.Helper()
	rightBytes := total - len("<a@>")
	if rightBytes < 1 {
		t.Fatalf("strict Message-ID total %d cannot contain both dot-atoms", total)
	}
	value := "<a@" + strings.Repeat("b", rightBytes) + ">"
	if len(value) != total {
		t.Fatalf("strict Message-ID fixture length = %d, want %d", len(value), total)
	}
	return value
}

func referenceIDs(count int) []string {
	values := make([]string, count)
	for index := range values {
		values[index] = "<reference-" + strconv.Itoa(index) + "@example.test>"
	}
	return values
}

func recipientAddresses(prefix string, count int) []mail.EngineAddress {
	values := make([]mail.EngineAddress, count)
	for index := range values {
		values[index] = mail.EngineAddress{Email: prefix + "-" + strconv.Itoa(index) + "@example.test"}
	}
	return values
}

func validAncestors(count int) []mail.ThreadAncestor {
	values := make([]mail.ThreadAncestor, count)
	for index := range values {
		values[index] = mail.ThreadAncestor{
			MessageID:  "<ancestor-" + strconv.Itoa(index) + "@example.test>",
			KeyVersion: mail.ThreadKeyVersion(1),
			Key:        nonzeroThreadKey(byte(index%254 + 1)),
		}
	}
	return values
}
