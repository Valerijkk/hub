package mail_test

import (
	"context"
	"errors"
	"math"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestMailboxAccountProvisioningScanRequest_NormalizationPreventsUnboundedDueReads(t *testing.T) {
	for _, testCase := range []struct {
		name  string
		input mail.MailboxAccountProvisioningScanRequest
		ok    bool
		want  mail.MailboxAccountProvisioningScanRequest
	}{
		{
			name:  "defaults an omitted batch and interval before scanning due accounts",
			input: mail.MailboxAccountProvisioningScanRequest{},
			ok:    true,
			want: mail.MailboxAccountProvisioningScanRequest{
				BatchSize:    20,
				ScanInterval: time.Minute,
			},
		},
		{
			name: "preserves an explicitly bounded batch and interval",
			input: mail.MailboxAccountProvisioningScanRequest{
				BatchSize:    50,
				ScanInterval: 15 * time.Second,
			},
			ok: true,
			want: mail.MailboxAccountProvisioningScanRequest{
				BatchSize:    50,
				ScanInterval: 15 * time.Second,
			},
		},
		{
			name: "rejects a batch above the hard cap before a repository can overread",
			input: mail.MailboxAccountProvisioningScanRequest{BatchSize: 51},
		},
		{
			name: "rejects a negative batch before a repository can reinterpret it",
			input: mail.MailboxAccountProvisioningScanRequest{BatchSize: -1},
		},
		{
			name: "rejects a negative scan interval instead of scheduling a hot loop",
			input: mail.MailboxAccountProvisioningScanRequest{ScanInterval: -time.Second},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got, ok := mail.NormalizeMailboxAccountProvisioningScanRequest(testCase.input)
			if ok != testCase.ok {
				t.Fatalf("NormalizeMailboxAccountProvisioningScanRequest(%#v) ok = %t, want %t", testCase.input, ok, testCase.ok)
			}
			if ok && got != testCase.want {
				t.Fatalf("NormalizeMailboxAccountProvisioningScanRequest(%#v) = %#v, want %#v", testCase.input, got, testCase.want)
			}
		})
	}

	request := mail.MailboxAccountProvisioningScanRequest{BatchSize: 2, ScanInterval: time.Minute}
	if !((mail.MailboxAccountProvisioningScanResult{AccountIDs: []uuid.UUID{provisioningAccountID, provisioningOrganizationID}}).ValidFor(request)) {
		t.Fatal("two distinct nonzero account UUIDs must remain a valid bounded due result")
	}
	if (mail.MailboxAccountProvisioningScanResult{AccountIDs: []uuid.UUID{provisioningAccountID, provisioningAccountID}}).ValidFor(request) {
		t.Fatal("duplicate due account UUIDs must not cause duplicate provider work")
	}
	if (mail.MailboxAccountProvisioningScanResult{AccountIDs: []uuid.UUID{uuid.Nil}}).ValidFor(request) {
		t.Fatal("a zero due account UUID must not become claimable work")
	}
	if (mail.MailboxAccountProvisioningScanResult{AccountIDs: []uuid.UUID{provisioningAccountID, provisioningOrganizationID, provisioningDomainID}}).ValidFor(request) {
		t.Fatal("a result above the requested batch must not bypass the scan bound")
	}
}

func TestMailboxAccountProvisioningClaim_SealsACompleteProviderCallFence(t *testing.T) {
	for _, local := range []string{"ada+ops", "_ada"} {
		candidate := validMailboxAccountProvisioningFence()
		candidate.LocalPart = local
		if _, ok := mail.NewMailboxAccountProvisioningClaim(candidate); ok {
			t.Fatal("bad local")
		}
	}
	candidate := validMailboxAccountProvisioningFence()
	candidate.Domain = "localhost"
	if _, ok := mail.NewMailboxAccountProvisioningClaim(candidate); ok {
		t.Fatal("bad domain")
	}
	fence := validMailboxAccountProvisioningFence()
	claim, ok := mail.NewMailboxAccountProvisioningClaim(fence)
	if !ok {
		t.Fatal("a complete canonical fence must be claimable")
	}
	if got := claim.Fence(); got != fence {
		t.Fatalf("Claim.Fence() = %#v, want exact sealed fence %#v", got, fence)
	}
	if got := claim.AccountSpec(); got != (mail.AccountSpec{
		ProvisioningKey:   "mailbox-account-41",
		LocalPart:         "ada",
		Domain:            "example.test",
		DesiredGeneration: 41,
		DesiredHash:       "sha256:account-control-41",
		Enabled:           true,
		Primary: mail.AccountIdentitySpec{
			LocalPart: "ada",
			Domain:    "example.test",
			Enabled:   true,
		},
	}) {
		t.Fatalf("Claim.AccountSpec() = %#v, want the one exact provider-neutral ensure input", got)
	}

	claimed := mail.MailboxAccountProvisioningClaimResult{
		Outcome: mail.MailboxAccountProvisioningClaimed,
		Claim:   claim,
	}
	if !claimed.Valid() {
		t.Fatal("a claimed outcome must carry the sealed provider input")
	}
	if !(mail.MailboxAccountProvisioningClaimResult{Outcome: mail.MailboxAccountProvisioningNoop}).Valid() {
		t.Fatal("a durable no-op must be distinguishable without a provider call")
	}
	for _, code := range []mail.MailboxAccountProvisioningCode{
		mail.MailboxAccountProvisioningCodeDomainRequired,
		mail.MailboxAccountProvisioningCodeEngineDisabled,
	} {
		if !(mail.MailboxAccountProvisioningClaimResult{
			Outcome: mail.MailboxAccountProvisioningLocalGate,
			Code:    code,
		}).Valid() {
			t.Fatalf("local gate %q must remain a closed claim outcome", code)
		}
	}
	if (mail.MailboxAccountProvisioningClaimResult{
		Outcome: mail.MailboxAccountProvisioningLocalGate,
		Code:    mail.MailboxAccountProvisioningCodeProvisioningRetry,
	}).Valid() {
		t.Fatal("a retry code must not be misreported as a local no-provider gate")
	}
}

func TestMailboxAccountProvisioningClaim_RejectsMalformedFencesBeforeProviderCall(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*mail.MailboxAccountProvisioningFence)
	}{
		{
			name: "rejects a zero account UUID that would unbind the provider call",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.AccountID = uuid.Nil
			},
		},
		{
			name: "rejects a zero organization UUID that would cross tenant control",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.OrganizationID = uuid.Nil
			},
		},
		{
			name: "rejects a zero domain UUID that would detach domain authorization",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.DomainID = uuid.Nil
			},
		},
		{
			name: "rejects a zero user UUID that would detach the mailbox owner",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.UserID = uuid.Nil
			},
		},
		{
			name: "rejects a zero lease token that would allow an unfenced finalization",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.LeaseToken = uuid.Nil
			},
		},
		{
			name: "rejects a zero lease deadline that would make the lease timeless",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.LeaseDeadline = time.Time{}
			},
		},
		{
			name: "rejects a noncanonical lease deadline with fractional seconds",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.LeaseDeadline = time.Date(2030, time.January, 2, 3, 4, 5, 1, time.UTC)
			},
		},
		{
			name: "rejects a non-UTC lease deadline that would change the persisted canonical fence",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.LeaseDeadline = time.Date(2030, time.January, 2, 3, 4, 5, 0, time.FixedZone("non-utc", 3600))
			},
		},
		{
			name: "rejects a lease duration different from the shared engine-command lease",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.LeaseDuration = 59 * time.Second
			},
		},
		{
			name: "rejects a zero control generation that would lose the current desired state",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.ControlGeneration = 0
			},
		},
		{
			name: "rejects a zero control hash version that would weaken the content fence",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.ControlHashVersion = 0
			},
		},
		{
			name: "rejects an empty control hash that would accept any desired configuration",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.ControlHash = ""
			},
		},
		{
			name: "rejects an unsafe control hash that could cross a logging boundary",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.ControlHash = "sha256:account\ncontrol"
			},
		},
		{
			name: "rejects an empty provisioning key that would destroy idempotency",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.ProvisioningKey = ""
			},
		},
		{
			name: "rejects a noncanonical local part before it reaches a Registry ensure",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.LocalPart = "Ada"
			},
		},
		{
			name: "rejects a noncanonical domain before it reaches a Registry ensure",
			mutate: func(fence *mail.MailboxAccountProvisioningFence) {
				fence.Domain = "Example.test"
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			fence := validMailboxAccountProvisioningFence()
			testCase.mutate(&fence)
			if claim, ok := mail.NewMailboxAccountProvisioningClaim(fence); ok || claim.Valid() {
				t.Fatalf("NewMailboxAccountProvisioningClaim(%#v) = %#v, %t; want an unclaimable fence", fence, claim, ok)
			}
		})
	}
}

func TestMailboxAccountProvisioningFinalize_RejectsRegistrySuccessMismatches(t *testing.T) {
	claim, ok := mail.NewMailboxAccountProvisioningClaim(validMailboxAccountProvisioningFence())
	if !ok {
		t.Fatal("test fixture fence must be claimable")
	}
	validAccount := mail.EngineAccount{
		AccountID:         "engine-account-41",
		State:             mail.EngineAccountProvisioning,
		AppliedGeneration: 41,
		AppliedHash:       "sha256:account-control-41",
		AliasIdentityIDs:  []string{},
	}
	request, ok := mail.NewMailboxAccountProvisioningFinalizeSuccess(claim, validAccount)
	if !ok || !request.Valid() {
		t.Fatalf("NewMailboxAccountProvisioningFinalizeSuccess() = %#v, %t; want valid Registry success", request, ok)
	}
	if request.Outcome != mail.MailboxAccountProvisioningFinalizeSuccess || request.EngineAccountID != "engine-account-41" || request.Code != "" || request.Claim.Fence() != claim.Fence() {
		t.Fatalf("success finalize request = %#v, want the exact sealed fence and safe account ID", request)
	}
	if (mail.MailboxAccountProvisioningFinalizeRequest{Claim: claim, Outcome: mail.MailboxAccountProvisioningFinalizeSuccess, EngineAccountID: "engine-account-41"}).Valid() {
		t.Fatal("a manually assembled success must not bypass Registry-shape validation")
	}
	if !(mail.MailboxAccountProvisioningFinalizeResult{Outcome: mail.MailboxAccountProvisioningFinalizeSuccess}).Valid() {
		t.Fatal("a durable provider-success finalization must remain distinct from retry")
	}

	for _, testCase := range []struct {
		name   string
		mutate func(*mail.EngineAccount)
	}{
		{
			name: "rejects an empty engine account ID instead of persisting an unusable provider binding",
			mutate: func(account *mail.EngineAccount) {
				account.AccountID = ""
			},
		},
		{
			name: "rejects an unsafe engine account ID instead of passing provider detail through the boundary",
			mutate: func(account *mail.EngineAccount) {
				account.AccountID = "engine-account\nsecret"
			},
		},
		{
			name: "rejects an active account before provisioning finalization is complete",
			mutate: func(account *mail.EngineAccount) {
				account.State = mail.EngineAccountActive
			},
		},
		{
			name: "rejects an applied generation that does not echo the sealed control generation",
			mutate: func(account *mail.EngineAccount) {
				account.AppliedGeneration = 42
			},
		},
		{
			name: "rejects an applied hash that does not echo the sealed control hash",
			mutate: func(account *mail.EngineAccount) {
				account.AppliedHash = "sha256:other"
			},
		},
		{
			name: "rejects a primary identity ID because Registry account success has none",
			mutate: func(account *mail.EngineAccount) {
				account.PrimaryIdentityID = "identity-41"
			},
		},
		{
			name: "rejects a nil alias list because Registry success is explicitly empty",
			mutate: func(account *mail.EngineAccount) {
				account.AliasIdentityIDs = nil
			},
		},
		{
			name: "rejects an alias identity because Registry account success has no aliases",
			mutate: func(account *mail.EngineAccount) {
				account.AliasIdentityIDs = []string{"alias-41"}
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			account := validAccount
			testCase.mutate(&account)
			if request, ok := mail.NewMailboxAccountProvisioningFinalizeSuccess(claim, account); ok || request.Valid() {
				t.Fatalf("NewMailboxAccountProvisioningFinalizeSuccess(%#v) = %#v, %t; want rejected provider success", account, request, ok)
			}
		})
	}

	for _, code := range []mail.MailboxAccountProvisioningCode{
		mail.MailboxAccountProvisioningCodeProvisioningRetry,
		mail.MailboxAccountProvisioningCodeUnavailable,
	} {
		retry, ok := mail.NewMailboxAccountProvisioningFinalizeRetry(claim, code)
		if !ok || !retry.Valid() || retry.Outcome != mail.MailboxAccountProvisioningFinalizeRetry || retry.Code != code || retry.Claim.Fence() != claim.Fence() {
			t.Fatalf("NewMailboxAccountProvisioningFinalizeRetry(%q) = %#v, %t; want valid retry with the exact sealed fence", code, retry, ok)
		}
		if !(mail.MailboxAccountProvisioningFinalizeResult{Outcome: mail.MailboxAccountProvisioningFinalizeRetry, Code: code}).Valid() {
			t.Fatalf("retry result %q must remain a valid closed finalization outcome", code)
		}
	}
	if (mail.MailboxAccountProvisioningFinalizeResult{
		Outcome: mail.MailboxAccountProvisioningFinalizeRetry,
		Code:    mail.MailboxAccountProvisioningCodeDomainRequired,
	}).Valid() {
		t.Fatal("a local claim gate must not become a finalization retry")
	}
}

func TestMailboxAccountProvisioningRetryDelay_UsesCappedBackoffWithoutTerminalAttempt(t *testing.T) {
	for _, testCase := range []struct {
		attempt int
		want    time.Duration
	}{
		{attempt: 1, want: 30 * time.Second},
		{attempt: 2, want: time.Minute},
		{attempt: 3, want: 2 * time.Minute},
		{attempt: 4, want: 4 * time.Minute},
		{attempt: 5, want: 5 * time.Minute},
		{attempt: 6, want: 5 * time.Minute},
		{attempt: math.MaxInt32 - 1, want: 5 * time.Minute},
	} {
		got, ok := mail.MailboxAccountProvisioningRetryDelay(testCase.attempt)
		if !ok || got != testCase.want {
			t.Fatalf("MailboxAccountProvisioningRetryDelay(%d) = %s, %t; want %s, true", testCase.attempt, got, ok, testCase.want)
		}
	}
	for _, attempt := range []int{0, -1, math.MaxInt32, math.MaxInt32 + 1} {
		if delay, ok := mail.MailboxAccountProvisioningRetryDelay(attempt); ok || delay != 0 {
			t.Fatalf("MailboxAccountProvisioningRetryDelay(%d) = %s, %t; want overflow-safe rejection", attempt, delay, ok)
		}
	}
}

func TestMailboxAccountProvisioningProviderErrorCode_StripsProviderDetailsIntoClosedRetryCodes(t *testing.T) {
	for _, testCase := range []struct {
		name string
		err  error
		want mail.MailboxAccountProvisioningCode
	}{
		{
			name: "maps cancellation to unavailable instead of exposing context detail",
			err:  context.Canceled,
			want: mail.MailboxAccountProvisioningCodeUnavailable,
		},
		{
			name: "maps deadline exhaustion to unavailable instead of exposing context detail",
			err:  context.DeadlineExceeded,
			want: mail.MailboxAccountProvisioningCodeUnavailable,
		},
		{
			name: "maps engine unavailability to unavailable",
			err:  &mail.EngineError{Kind: mail.EngineErrorUnavailable},
			want: mail.MailboxAccountProvisioningCodeUnavailable,
		},
		{
			name: "maps indeterminate engine results to unavailable",
			err:  &mail.EngineError{Kind: mail.EngineErrorIndeterminate},
			want: mail.MailboxAccountProvisioningCodeUnavailable,
		},
		{
			name: "maps restart-required engine results to unavailable",
			err:  &mail.EngineError{Kind: mail.EngineErrorRestartRequired},
			want: mail.MailboxAccountProvisioningCodeUnavailable,
		},
		{
			name: "maps rejected policy to retry without leaking the provider reason",
			err:  &mail.EngineError{Kind: mail.EngineErrorRejectedPolicy},
			want: mail.MailboxAccountProvisioningCodeProvisioningRetry,
		},
		{
			name: "maps an unknown provider error to retry without preserving its text",
			err:  errors.New("provider-secret@example.test"),
			want: mail.MailboxAccountProvisioningCodeProvisioningRetry,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			if got := mail.MailboxAccountProvisioningProviderErrorCode(testCase.err); got != testCase.want {
				t.Fatalf("MailboxAccountProvisioningProviderErrorCode(%v) = %q, want %q", testCase.err, got, testCase.want)
			}
		})
	}
}

var (
	provisioningAccountID      = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	provisioningOrganizationID = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	provisioningDomainID       = uuid.MustParse("33333333-3333-4333-8333-333333333333")
	provisioningUserID         = uuid.MustParse("44444444-4444-4444-8444-444444444444")
	provisioningLeaseToken     = uuid.MustParse("55555555-5555-4555-8555-555555555555")
)

func validMailboxAccountProvisioningFence() mail.MailboxAccountProvisioningFence {
	return mail.MailboxAccountProvisioningFence{
		AccountID:          provisioningAccountID,
		OrganizationID:     provisioningOrganizationID,
		DomainID:           provisioningDomainID,
		UserID:             provisioningUserID,
		LeaseToken:         provisioningLeaseToken,
		LeaseDeadline:      time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC),
		LeaseDuration:      60 * time.Second,
		ControlGeneration:  41,
		ControlHashVersion: int32(1),
		ControlHash:        "sha256:account-control-41",
		ProvisioningKey:    "mailbox-account-41",
		LocalPart:          "ada",
		Domain:             "example.test",
	}
}
