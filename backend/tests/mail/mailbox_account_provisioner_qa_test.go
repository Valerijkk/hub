//go:build (qa || hubtest) && !release

package mail_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

const (
	qaMailboxAccountProvisionerNamespace = "hub.qa.mail.account.provisioner.v1"
	qaMailboxAccountProvisionerIDPrefix  = "qa-mail-account-v1:"
)

func TestQAMailboxAccountProvisionerProjectsDeterministicAcceptedAccount(t *testing.T) {
	provisioner := mail.NewQAMailboxAccountProvisioner()
	if provisioner == nil {
		t.Fatal("NewQAMailboxAccountProvisioner() = nil")
	}

	spec := validQAMailboxAccountProvisionerSpec()
	first, err := provisioner.EnsureAccount(context.Background(), spec)
	if err != nil {
		t.Fatalf("first EnsureAccount() error = %v", err)
	}
	second, err := provisioner.EnsureAccount(context.Background(), spec)
	if err != nil {
		t.Fatalf("second EnsureAccount() error = %v", err)
	}
	if !reflect.DeepEqual(first, second) {
		t.Fatalf("repeated EnsureAccount() = %#v and %#v, want deterministic identical accounts", first, second)
	}

	want := mail.EngineAccount{
		AccountID:         qaMailboxAccountProvisionerID(spec.ProvisioningKey),
		State:             mail.EngineAccountProvisioning,
		AppliedGeneration: spec.DesiredGeneration,
		AppliedHash:       spec.DesiredHash,
		AliasIdentityIDs:  []string{},
	}
	if !reflect.DeepEqual(first, want) {
		t.Fatalf("EnsureAccount() = %#v, want exact accepted projection %#v", first, want)
	}
	if strings.Contains(first.AccountID, spec.ProvisioningKey) {
		t.Fatalf("AccountID %q must not disclose provisioning key %q", first.AccountID, spec.ProvisioningKey)
	}

	claim := qaMailboxAccountProvisionerClaim(t, spec)
	if request, ok := mail.NewMailboxAccountProvisioningFinalizeSuccess(claim, first); !ok || !request.Valid() {
		t.Fatalf("NewMailboxAccountProvisioningFinalizeSuccess(%#v) = %#v, %t; want exact accepted Registry success shape", first, request, ok)
	}
}

func TestQAMailboxAccountProvisionerSeparatesProvisioningKeys(t *testing.T) {
	provisioner := mail.NewQAMailboxAccountProvisioner()
	firstSpec := validQAMailboxAccountProvisionerSpec()
	secondSpec := validQAMailboxAccountProvisionerSpec()
	secondSpec.ProvisioningKey = "qa-mailbox-account-key-42"

	first, err := provisioner.EnsureAccount(context.Background(), firstSpec)
	if err != nil {
		t.Fatalf("first EnsureAccount() error = %v", err)
	}
	second, err := provisioner.EnsureAccount(context.Background(), secondSpec)
	if err != nil {
		t.Fatalf("second EnsureAccount() error = %v", err)
	}
	if first.AccountID == second.AccountID {
		t.Fatalf("different provisioning keys produced the same AccountID %q", first.AccountID)
	}
}

func TestQAMailboxAccountProvisionerRejectsMalformedInputBeforeCancellation(t *testing.T) {
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()

	for _, testCase := range []struct {
		name        string
		nilReceiver bool
		context     context.Context
		mutate      func(*mail.AccountSpec)
	}{
		{
			name:        "nil receiver",
			nilReceiver: true,
			context:     context.Background(),
		},
		{
			name:    "nil context",
			context: nil,
		},
		{
			name:    "disabled account",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.Enabled = false
			},
		},
		{
			name:    "zero desired generation",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.DesiredGeneration = 0
			},
		},
		{
			name:    "empty provisioning key",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.ProvisioningKey = ""
			},
		},
		{
			name:    "unsafe provisioning key",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.ProvisioningKey = "qa-mailbox\naccount-key"
			},
		},
		{
			name:    "noncanonical local part",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.LocalPart = "Ada"
			},
		},
		{
			name:    "noncanonical domain",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.Domain = "Example.test"
			},
		},
		{
			name:    "disabled primary",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.Primary.Enabled = false
			},
		},
		{
			name:    "primary local part mismatch",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.Primary.LocalPart = "grace"
			},
		},
		{
			name:    "primary domain mismatch",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.Primary.Domain = "other.test"
			},
		},
		{
			name:    "nonempty aliases",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.Aliases = []mail.AccountIdentitySpec{{LocalPart: "alias", Domain: "example.test", Enabled: true}}
			},
		},
		{
			name:    "empty desired hash",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.DesiredHash = ""
			},
		},
		{
			name:    "63-character desired hash",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.DesiredHash = strings.Repeat("a", 63)
			},
		},
		{
			name:    "65-character desired hash",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.DesiredHash = strings.Repeat("a", 65)
			},
		},
		{
			name:    "uppercase desired hash",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.DesiredHash = strings.ToUpper(spec.DesiredHash)
			},
		},
		{
			name:    "nonhex desired hash",
			context: context.Background(),
			mutate: func(spec *mail.AccountSpec) {
				spec.DesiredHash = strings.Repeat("g", 64)
			},
		},
		{
			name:    "invalid input wins over already cancelled context",
			context: cancelled,
			mutate: func(spec *mail.AccountSpec) {
				spec.Enabled = false
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			spec := validQAMailboxAccountProvisionerSpec()
			if testCase.mutate != nil {
				testCase.mutate(&spec)
			}
			provisioner := mail.NewQAMailboxAccountProvisioner()
			if testCase.nilReceiver {
				provisioner = nil
			}

			_, err := provisioner.EnsureAccount(testCase.context, spec)
			requireQAMailboxAccountProvisionerError(t, err, mail.ErrEngineRejectedPolicy)
		})
	}
}

func TestQAMailboxAccountProvisionerReturnsUnavailableForCancelledContexts(t *testing.T) {
	for _, testCase := range []struct {
		name        string
		newContext  func() (context.Context, context.CancelFunc)
	}{
		{
			name: "cancelled context",
			newContext: func() (context.Context, context.CancelFunc) {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx, func() {}
			},
		},
		{
			name: "expired context",
			newContext: func() (context.Context, context.CancelFunc) {
				return context.WithDeadline(context.Background(), time.Now().Add(-time.Second))
			},
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := testCase.newContext()
			defer cancel()

			_, err := mail.NewQAMailboxAccountProvisioner().EnsureAccount(ctx, validQAMailboxAccountProvisionerSpec())
			requireQAMailboxAccountProvisionerError(t, err, mail.ErrEngineUnavailable)
		})
	}
}

func TestQAMailboxAccountProvisionerRechecksContextBeforeSuccess(t *testing.T) {
	ctx := &qaMailboxAccountProvisionerBoundaryCancellationContext{}
	_, err := mail.NewQAMailboxAccountProvisioner().EnsureAccount(ctx, validQAMailboxAccountProvisionerSpec())
	requireQAMailboxAccountProvisionerError(t, err, mail.ErrEngineUnavailable)
	if ctx.errCalls < 2 {
		t.Fatalf("EnsureAccount() checked context %d times, want a recheck immediately before success", ctx.errCalls)
	}
}

func validQAMailboxAccountProvisionerSpec() mail.AccountSpec {
	return mail.AccountSpec{
		ProvisioningKey:   "qa-mailbox-account-key-41",
		LocalPart:         "ada",
		Domain:            "example.test",
		DesiredGeneration: 41,
		DesiredHash:       "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef",
		Enabled:           true,
		Primary: mail.AccountIdentitySpec{
			LocalPart: "ada",
			Domain:    "example.test",
			Enabled:   true,
		},
	}
}

func qaMailboxAccountProvisionerID(provisioningKey string) string {
	digest := sha256.Sum256([]byte(qaMailboxAccountProvisionerNamespace + "\x00" + provisioningKey))
	return qaMailboxAccountProvisionerIDPrefix + hex.EncodeToString(digest[:])
}

func qaMailboxAccountProvisionerClaim(t *testing.T, spec mail.AccountSpec) mail.MailboxAccountProvisioningClaim {
	t.Helper()
	claim, ok := mail.NewMailboxAccountProvisioningClaim(mail.MailboxAccountProvisioningFence{
		AccountID:          uuid.MustParse("11111111-1111-4111-8111-111111111111"),
		OrganizationID:     uuid.MustParse("22222222-2222-4222-8222-222222222222"),
		DomainID:           uuid.MustParse("33333333-3333-4333-8333-333333333333"),
		UserID:             uuid.MustParse("44444444-4444-4444-8444-444444444444"),
		LeaseToken:         uuid.MustParse("55555555-5555-4555-8555-555555555555"),
		LeaseDeadline:      time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC),
		LeaseDuration:      60 * time.Second,
		ControlGeneration:  spec.DesiredGeneration,
		ControlHashVersion: 1,
		ControlHash:        spec.DesiredHash,
		ProvisioningKey:    spec.ProvisioningKey,
		LocalPart:          spec.LocalPart,
		Domain:             spec.Domain,
	})
	if !ok {
		t.Fatalf("NewMailboxAccountProvisioningClaim() rejected test fixture")
	}
	return claim
}

func requireQAMailboxAccountProvisionerError(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) || err == nil || err.Error() != want.Error() {
		t.Fatalf("EnsureAccount() error = %v, want content-free %v", err, want)
	}
	if _, ok := err.(*mail.EngineError); !ok {
		t.Fatalf("EnsureAccount() error type = %T, want *mail.EngineError", err)
	}
}

type qaMailboxAccountProvisionerBoundaryCancellationContext struct {
	errCalls int
}

func (ctx *qaMailboxAccountProvisionerBoundaryCancellationContext) Deadline() (time.Time, bool) {
	return time.Time{}, false
}

func (ctx *qaMailboxAccountProvisionerBoundaryCancellationContext) Done() <-chan struct{} {
	return nil
}

func (ctx *qaMailboxAccountProvisionerBoundaryCancellationContext) Err() error {
	ctx.errCalls++
	if ctx.errCalls >= 2 {
		return context.Canceled
	}
	return nil
}

func (ctx *qaMailboxAccountProvisionerBoundaryCancellationContext) Value(any) any {
	return nil
}
