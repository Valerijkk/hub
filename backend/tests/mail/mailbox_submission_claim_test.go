package mail_test

import (
	"context"
	"crypto/sha256"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/mail"
)

func TestMailboxSubmissionPolicyBinding_CanonicalizesAndDefensivelyCopies(t *testing.T) {
	input := mailboxSubmissionPolicyBindingInput()
	before := mailboxSubmissionPolicyBindingInputCopy(input)

	binding, err := mail.NewMailboxSubmissionPolicyBinding(input)
	if err != nil {
		t.Fatalf("NewMailboxSubmissionPolicyBinding() error = %v", err)
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatal("NewMailboxSubmissionPolicyBinding() mutated its input")
	}
	if got, want := binding.MaxMessageBytes(), input.PolicyMaxMessageBytes; got != want {
		t.Fatalf("MaxMessageBytes() = %d, want %d", got, want)
	}
	if got, want := binding.MaxRecipients(), input.PolicyMaxRecipients; got != want {
		t.Fatalf("MaxRecipients() = %d, want %d", got, want)
	}
	if got, want := binding.ExternalMailEnabled(), input.ExternalMailEnabled; got != want {
		t.Fatalf("ExternalMailEnabled() = %t, want %t", got, want)
	}
	if got, want := binding.InternalDomains(), []string{"alpha.example.test", "zeta.example.test"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("InternalDomains() = %#v, want %#v", got, want)
	}
	if hash := binding.Hash(); hash == [sha256.Size]byte{} {
		t.Fatal("Hash() must be nonzero for a valid policy binding")
	}

	input.VerifiedEnabledDomains[0].Domain = "mutated.example.test"
	input.VerifiedEnabledDomains[1].Version = 99
	if got, want := binding.InternalDomains(), []string{"alpha.example.test", "zeta.example.test"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("source mutation changed InternalDomains() = %#v, want %#v", got, want)
	}
	returned := binding.InternalDomains()
	returned[0] = "also-mutated.example.test"
	if got, want := binding.InternalDomains(), []string{"alpha.example.test", "zeta.example.test"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("returned slice mutation changed InternalDomains() = %#v, want %#v", got, want)
	}

	submission := mailboxSubmissionPolicyTestSubmission()
	request, ok := binding.PreflightRequest(submission)
	if !ok {
		t.Fatal("PreflightRequest() = false for a valid binding and submission")
	}
	if request.Submission != submission {
		t.Fatalf("PreflightRequest().Submission = %#v, want exact %#v", request.Submission, submission)
	}
	if request.PolicyMaxMessageBytes != binding.MaxMessageBytes() || request.MaxRecipients != binding.MaxRecipients() || request.ExternalMailEnabled != binding.ExternalMailEnabled() {
		t.Fatalf("PreflightRequest() policy = %#v, want values from binding", request)
	}
	if want := binding.InternalDomains(); !reflect.DeepEqual(request.InternalDomains, want) {
		t.Fatalf("PreflightRequest().InternalDomains = %#v, want %#v", request.InternalDomains, want)
	}
	request.InternalDomains[0] = "changed-after-request.example.test"
	second, ok := binding.PreflightRequest(mailboxSubmissionPolicyTestSubmission())
	if !ok || !reflect.DeepEqual(second.InternalDomains, binding.InternalDomains()) {
		t.Fatalf("PreflightRequest() did not return a defensive domain copy: %#v", second)
	}
	if invalid, valid := binding.PreflightRequest(mail.EngineSubmission{}); valid || !reflect.DeepEqual(invalid, mail.EngineSubmissionPreflightRequest{}) {
		t.Fatalf("PreflightRequest(zero submission) = %#v, %t; want zero, false", invalid, valid)
	}
}

func TestMailboxSubmissionPolicyBinding_RejectsInvalidInput(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		input mail.MailboxSubmissionPolicyBindingInput
	}{
		{name: "nil organization", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.OrganizationID = uuid.Nil; return value }()},
		{name: "nil policy", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.PolicyID = uuid.Nil; return value }()},
		{name: "zero policy version", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.PolicyVersion = 0; return value }()},
		{name: "message cap below mailbox policy lower bound", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.PolicyMaxMessageBytes = 1_048_575; return value }()},
		{name: "message cap above engine bound", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.PolicyMaxMessageBytes = mail.MaxEngineMessageBytes + 1; return value }()},
		{name: "zero recipient cap", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.PolicyMaxRecipients = 0; return value }()},
		{name: "recipient cap above engine bound", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.PolicyMaxRecipients = mail.MaxEngineDraftTotalRecipientItems + 1; return value }()},
		{name: "no verified domains", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.VerifiedEnabledDomains = nil; return value }()},
		{name: "too many domains", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.VerifiedEnabledDomains = make([]mail.MailboxSubmissionPolicyDomain, 101); return value }()},
		{name: "nil domain id", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.VerifiedEnabledDomains[0].ID = uuid.Nil; return value }()},
		{name: "zero domain version", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.VerifiedEnabledDomains[0].Version = 0; return value }()},
		{name: "uppercase domain", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.VerifiedEnabledDomains[0].Domain = "Alpha.example.test"; return value }()},
		{name: "invalid domain shape", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.VerifiedEnabledDomains[0].Domain = "invalid domain"; return value }()},
		{name: "duplicate domain", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.VerifiedEnabledDomains[1].Domain = value.VerifiedEnabledDomains[0].Domain; return value }()},
		{name: "duplicate domain id", input: func() mail.MailboxSubmissionPolicyBindingInput { value := mailboxSubmissionPolicyBindingInput(); value.VerifiedEnabledDomains[1].ID = value.VerifiedEnabledDomains[0].ID; return value }()},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			binding, err := mail.NewMailboxSubmissionPolicyBinding(testCase.input)
			if !errors.Is(err, mail.ErrMailboxSubmissionPolicyBindingInvalid) {
				t.Fatalf("NewMailboxSubmissionPolicyBinding() error = %v, want invalid policy binding", err)
			}
			if !reflect.DeepEqual(binding, mail.MailboxSubmissionPolicyBinding{}) {
				t.Fatalf("binding = %#v, want zero value", binding)
			}
		})
	}
}

func TestMailboxSubmissionPolicyBinding_HashIsDeterministicAndSensitive(t *testing.T) {
	base := mailboxSubmissionPolicyBindingInput()
	first, err := mail.NewMailboxSubmissionPolicyBinding(base)
	if err != nil {
		t.Fatalf("first NewMailboxSubmissionPolicyBinding() error = %v", err)
	}
	reordered := mailboxSubmissionPolicyBindingInputCopy(base)
	reordered.VerifiedEnabledDomains[0], reordered.VerifiedEnabledDomains[1] = reordered.VerifiedEnabledDomains[1], reordered.VerifiedEnabledDomains[0]
	second, err := mail.NewMailboxSubmissionPolicyBinding(reordered)
	if err != nil {
		t.Fatalf("reordered NewMailboxSubmissionPolicyBinding() error = %v", err)
	}
	if first.Hash() != second.Hash() {
		t.Fatalf("canonical reordering changed hash: first=%x second=%x", first.Hash(), second.Hash())
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailboxSubmissionPolicyBindingInput)
	}{
		{name: "organization id", mutate: func(value *mail.MailboxSubmissionPolicyBindingInput) { value.OrganizationID = uuid.MustParse("00000000-0000-0000-0000-000000000021") }},
		{name: "policy id", mutate: func(value *mail.MailboxSubmissionPolicyBindingInput) { value.PolicyID = uuid.MustParse("00000000-0000-0000-0000-000000000022") }},
		{name: "policy version", mutate: func(value *mail.MailboxSubmissionPolicyBindingInput) { value.PolicyVersion++ }},
		{name: "message cap", mutate: func(value *mail.MailboxSubmissionPolicyBindingInput) { value.PolicyMaxMessageBytes++ }},
		{name: "recipient cap", mutate: func(value *mail.MailboxSubmissionPolicyBindingInput) { value.PolicyMaxRecipients-- }},
		{name: "external mail toggle", mutate: func(value *mail.MailboxSubmissionPolicyBindingInput) { value.ExternalMailEnabled = !value.ExternalMailEnabled }},
		{name: "domain id", mutate: func(value *mail.MailboxSubmissionPolicyBindingInput) { value.VerifiedEnabledDomains[0].ID = uuid.MustParse("00000000-0000-0000-0000-000000000023") }},
		{name: "domain version", mutate: func(value *mail.MailboxSubmissionPolicyBindingInput) { value.VerifiedEnabledDomains[0].Version++ }},
		{name: "domain value", mutate: func(value *mail.MailboxSubmissionPolicyBindingInput) { value.VerifiedEnabledDomains[0].Domain = "beta.example.test" }},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			candidate := mailboxSubmissionPolicyBindingInputCopy(base)
			testCase.mutate(&candidate)
			binding, err := mail.NewMailboxSubmissionPolicyBinding(candidate)
			if err != nil {
				t.Fatalf("NewMailboxSubmissionPolicyBinding() error = %v", err)
			}
			if binding.Hash() == first.Hash() {
				t.Fatalf("%s mutation did not change policy hash", testCase.name)
			}
		})
	}
}

func TestMailboxSubmissionClaimABI_ZeroValuesFailClosedAndKeepCapabilitiesPrivate(t *testing.T) {
	// Positive sealed capabilities are intentionally not minted outside mail;
	// Ent claim/CAS path tests wait for the repository implementation.
	var zeroBinding mail.MailboxSubmissionPolicyBinding
	if zeroBinding.MaxMessageBytes() != 0 || zeroBinding.MaxRecipients() != 0 || zeroBinding.ExternalMailEnabled() || zeroBinding.InternalDomains() != nil || zeroBinding.Hash() != [sha256.Size]byte{} {
		t.Fatalf("zero policy binding exposed usable state: %#v", zeroBinding)
	}
	if request, ok := zeroBinding.PreflightRequest(mailboxSubmissionPolicyTestSubmission()); ok || !reflect.DeepEqual(request, mail.EngineSubmissionPreflightRequest{}) {
		t.Fatalf("zero binding PreflightRequest() = %#v, %t; want zero, false", request, ok)
	}

	var zeroClaim mail.MailboxSubmissionClaim
	if zeroClaim.SubmissionID() != uuid.Nil || zeroClaim.Fence() != (mail.MailboxEngineCommandFence{}) || !reflect.DeepEqual(zeroClaim.Policy(), mail.MailboxSubmissionPolicyBinding{}) {
		t.Fatalf("zero claim exposed usable state: %#v", zeroClaim)
	}
	var zeroAuthorization mail.MailboxSubmissionDispatchAuthorization
	if zeroAuthorization.SubmissionID() != uuid.Nil || zeroAuthorization.Fence() != (mail.MailboxEngineCommandFence{}) || !zeroAuthorization.AuthorizedAt().IsZero() {
		t.Fatalf("zero authorization exposed usable state: %#v", zeroAuthorization)
	}
	var zeroClaimResult mail.MailboxSubmissionClaimResult
	if zeroClaimResult.Claimed() {
		t.Fatal("zero claim result must not be claimed")
	}
	if claim, ok := zeroClaimResult.Claim(); ok || !reflect.DeepEqual(claim, mail.MailboxSubmissionClaim{}) {
		t.Fatalf("zero claim result Claim() = %#v, %t", claim, ok)
	}
	if busyUntil, ok := zeroClaimResult.BusyUntil(); ok || !busyUntil.IsZero() {
		t.Fatalf("zero claim result BusyUntil() = %v, %t", busyUntil, ok)
	}
	var zeroAuthorizeResult mail.MailboxSubmissionAuthorizeResult
	if zeroAuthorizeResult.Outcome() != "" {
		t.Fatalf("zero authorization result Outcome() = %q", zeroAuthorizeResult.Outcome())
	}
	if authorization, ok := zeroAuthorizeResult.Authorization(); ok || authorization != (mail.MailboxSubmissionDispatchAuthorization{}) {
		t.Fatalf("zero authorization result Authorization() = %#v, %t", authorization, ok)
	}
	if retryAt, ok := zeroAuthorizeResult.RetryAt(); ok || !retryAt.IsZero() {
		t.Fatalf("zero authorization result RetryAt() = %v, %t", retryAt, ok)
	}

	for _, candidate := range []reflect.Type{
		reflect.TypeOf(mail.MailboxSubmissionClaim{}),
		reflect.TypeOf(mail.MailboxSubmissionClaimResult{}),
		reflect.TypeOf(mail.MailboxSubmissionDispatchAuthorization{}),
		reflect.TypeOf(mail.MailboxSubmissionAuthorizeResult{}),
	} {
		for index := 0; index < candidate.NumField(); index++ {
			field := candidate.Field(index)
			if field.PkgPath == "" {
				t.Fatalf("%s exposes mutable field %q", candidate.Name(), field.Name)
			}
			if candidate == reflect.TypeOf(mail.MailboxSubmissionDispatchAuthorization{}) && field.Name == "authorizationHash" {
				continue
			}
			lower := strings.ToLower(field.Name)
			for _, forbidden := range []string{"provider", "grant", "hash", "address", "draftkey", "messageid", "expectedversion", "submissionkey"} {
				if strings.Contains(lower, forbidden) {
					t.Fatalf("%s leaks forbidden field %q", candidate.Name(), field.Name)
				}
			}
		}
	}
}

func TestMailboxSubmissionPreflightFailureABI_PublicShapeAndZeroContract(t *testing.T) {
	for _, testCase := range []struct {
		name string
		got  string
		want string
	}{
		{name: "authorized", got: string(mail.MailboxSubmissionAuthorizeOutcomeAuthorized), want: "authorized"},
		{name: "rate limited", got: string(mail.MailboxSubmissionAuthorizeOutcomeRateLimited), want: "rate_limited"},
		{name: "requeued", got: string(mail.MailboxSubmissionAuthorizeOutcomeRequeued), want: "requeued"},
		{name: "failed", got: string(mail.MailboxSubmissionAuthorizeOutcomeFailed), want: "failed"},
		{name: "preflight failed", got: string(mail.MailboxSubmissionPreflightFailureDispositionFailed), want: "failed"},
		{name: "preflight retry", got: string(mail.MailboxSubmissionPreflightFailureDispositionRetry), want: "retry"},
		{name: "preflight rate limited", got: string(mail.MailboxSubmissionPreflightFailureDispositionRateLimited), want: "rate_limited"},
	} {
		if testCase.got != testCase.want {
			t.Fatalf("%s value = %q, want %q", testCase.name, testCase.got, testCase.want)
		}
	}

	requestType := reflect.TypeOf(mail.MailboxSubmissionPreflightFailureRequest{})
	if got, want := requestType.NumField(), 4; got != want {
		t.Fatalf("MailboxSubmissionPreflightFailureRequest field count = %d, want %d", got, want)
	}
	for index, want := range []struct {
		name string
		type_ reflect.Type
	}{
		{name: "Claim", type_: reflect.TypeOf(mail.MailboxSubmissionClaim{})},
		{name: "Disposition", type_: reflect.TypeOf(mail.MailboxSubmissionPreflightFailureDisposition(""))},
		{name: "FailureCode", type_: reflect.TypeOf(mail.MailboxSubmissionErrorCode(""))},
		{name: "RetryAfterSeconds", type_: reflect.TypeOf(int32(0))},
	} {
		field := requestType.Field(index)
		if field.Name != want.name || field.Type != want.type_ || field.PkgPath != "" {
			t.Fatalf("MailboxSubmissionPreflightFailureRequest field %d = %s %s exported=%t, want %s %s exported", index, field.Name, field.Type, field.PkgPath == "", want.name, want.type_)
		}
	}

	repositoryType := reflect.TypeOf((*mail.MailboxSubmissionClaimRepository)(nil)).Elem()
	if got, want := repositoryType.NumMethod(), 3; got != want {
		t.Fatalf("MailboxSubmissionClaimRepository method count = %d, want %d", got, want)
	}
	method, ok := repositoryType.MethodByName("ResolveMailboxSubmissionPreflightFailure")
	if !ok {
		t.Fatal("MailboxSubmissionClaimRepository lacks ResolveMailboxSubmissionPreflightFailure")
	}
	if got, want := method.Type.NumIn(), 2; got != want {
		t.Fatalf("ResolveMailboxSubmissionPreflightFailure input count = %d, want %d", got, want)
	}
	if got, want := method.Type.In(0), reflect.TypeOf((*context.Context)(nil)).Elem(); got != want {
		t.Fatalf("ResolveMailboxSubmissionPreflightFailure context type = %s, want %s", got, want)
	}
	if got, want := method.Type.In(1), reflect.TypeOf(mail.MailboxSubmissionPreflightFailureRequest{}); got != want {
		t.Fatalf("ResolveMailboxSubmissionPreflightFailure request type = %s, want %s", got, want)
	}
	if got, want := method.Type.NumOut(), 2; got != want {
		t.Fatalf("ResolveMailboxSubmissionPreflightFailure output count = %d, want %d", got, want)
	}
	if got, want := method.Type.Out(0), reflect.TypeOf(mail.MailboxSubmissionAuthorizeResult{}); got != want {
		t.Fatalf("ResolveMailboxSubmissionPreflightFailure result type = %s, want %s", got, want)
	}
	if got, want := method.Type.Out(1), reflect.TypeOf((*error)(nil)).Elem(); got != want {
		t.Fatalf("ResolveMailboxSubmissionPreflightFailure error type = %s, want %s", got, want)
	}

	var zeroRequest mail.MailboxSubmissionPreflightFailureRequest
	if zeroRequest.Claim.SubmissionID() != uuid.Nil || zeroRequest.Disposition != "" || zeroRequest.FailureCode != "" || zeroRequest.RetryAfterSeconds != 0 {
		t.Fatalf("zero preflight failure request minted capability or state: %#v", zeroRequest)
	}
	// Positive claims/results are intentionally unconstructable in mail_test;
	// future Ent repository tests cover each package-private validator branch.
	var zeroResult mail.MailboxSubmissionAuthorizeResult
	if retryAt, ok := zeroResult.RetryAt(); ok || !retryAt.IsZero() {
		t.Fatalf("zero preflight failure result RetryAt() = %v, %t", retryAt, ok)
	}
	if code, ok := zeroResult.FailureCode(); ok || code != "" {
		t.Fatalf("zero preflight failure result FailureCode() = %q, %t", code, ok)
	}
}

func mailboxSubmissionPolicyBindingInput() mail.MailboxSubmissionPolicyBindingInput {
	return mail.MailboxSubmissionPolicyBindingInput{
		OrganizationID:        uuid.MustParse("00000000-0000-0000-0000-000000000001"),
		PolicyID:              uuid.MustParse("00000000-0000-0000-0000-000000000002"),
		PolicyVersion:         7,
		PolicyMaxMessageBytes: 8 * 1_048_576,
		PolicyMaxRecipients:   42,
		ExternalMailEnabled:   false,
		VerifiedEnabledDomains: []mail.MailboxSubmissionPolicyDomain{
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000004"), Version: 3, Domain: "zeta.example.test"},
			{ID: uuid.MustParse("00000000-0000-0000-0000-000000000003"), Version: 2, Domain: "alpha.example.test"},
		},
	}
}

func mailboxSubmissionPolicyBindingInputCopy(input mail.MailboxSubmissionPolicyBindingInput) mail.MailboxSubmissionPolicyBindingInput {
	input.VerifiedEnabledDomains = append([]mail.MailboxSubmissionPolicyDomain(nil), input.VerifiedEnabledDomains...)
	return input
}

func mailboxSubmissionPolicyTestSubmission() mail.EngineSubmission {
	return mail.EngineSubmission{
		AccountID:       "account-a",
		DraftKey:        "draft-a",
		MessageID:       "message-a",
		ExpectedVersion: "version-a",
		SubmissionKey:   "submission-a",
		MaxMessageBytes: 4 * 1_048_576,
	}
}
