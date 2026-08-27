package mail_test

import (
	"context"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestMailboxAccountProvisioningWorkerConstructorDefaultsAndDisabledZero(t *testing.T) {
	defaults := mail.DefaultMailboxAccountProvisioningWorkerOptions()
	if defaults != (mail.MailboxAccountProvisioningWorkerOptions{
		Enabled: true, ScanInterval: time.Minute, ProviderTimeout: 30 * time.Second, FinalizeTimeout: 10 * time.Second, BatchSize: 20,
	}) {
		t.Fatalf("DefaultMailboxAccountProvisioningWorkerOptions() = %#v", defaults)
	}

	disabled, err := mail.NewMailboxAccountProvisioningWorker(nil, nil, nil, mail.MailboxAccountProvisioningWorkerOptions{})
	if err != nil || disabled == nil {
		t.Fatalf("disabled worker = %#v, %v", disabled, err)
	}
	if result, runErr := disabled.RunOnce(context.Background()); result != (mail.MailboxAccountProvisioningWorkerRunResult{}) || runErr != nil {
		t.Fatalf("disabled RunOnce() = %#v, %v", result, runErr)
	}

	worker, constructorErr := mail.NewMailboxAccountProvisioningWorker(nil, nil, nil, mail.MailboxAccountProvisioningWorkerOptions{Enabled: true})
	if worker != nil || !errors.Is(constructorErr, mail.ErrMailboxAccountProvisioningWorkerInvalid) {
		t.Fatalf("enabled nil dependencies = %#v, %v", worker, constructorErr)
	}
}

func TestMailboxAccountProvisioningWorkerSealsProviderSuccessAndRetry(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	first := provisioningWorkerClaim(t, provisioningWorkerIDOne, now.Add(time.Minute), strings.Repeat("a", 64))
	second := provisioningWorkerClaim(t, provisioningWorkerIDTwo, now.Add(time.Minute), strings.Repeat("b", 64))
	repository := &provisioningWorkerRepositoryFake{
		scan: func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) {
			return mail.MailboxAccountProvisioningScanResult{AccountIDs: []uuid.UUID{provisioningWorkerIDOne, provisioningWorkerIDTwo}}, nil
		},
		claim: func(_ context.Context, request mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error) {
			claim := first
			if request.AccountID == provisioningWorkerIDTwo { claim = second }
			return mail.MailboxAccountProvisioningClaimResult{Outcome: mail.MailboxAccountProvisioningClaimed, Claim: claim}, nil
		},
		finalize: func(_ context.Context, request mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error) {
			return mail.MailboxAccountProvisioningFinalizeResult{Outcome: request.Outcome, Code: request.Code}, nil
		},
	}
	provisioner := &provisioningWorkerProvisionerFake{ensure: func(_ context.Context, spec mail.AccountSpec) (mail.EngineAccount, error) {
		if spec.DesiredHash == strings.Repeat("a", 64) { return provisioningWorkerSuccess(spec), nil }
		return mail.EngineAccount{}, errors.New("provider text")
	}}
	worker, err := mail.NewMailboxAccountProvisioningWorker(repository, provisioner, &provisioningWorkerTimerFake{now: now}, mail.DefaultMailboxAccountProvisioningWorkerOptions())
	if err != nil { t.Fatal(err) }
	result, runErr := worker.RunOnce(context.Background())
	want := mail.MailboxAccountProvisioningWorkerRunResult{Due: 2, Claimed: 2, ProviderSucceeded: 1, ProviderFailed: 1, FinalizedSuccess: 1, FinalizedRetry: 1}
	if result != want || runErr != nil { t.Fatalf("RunOnce() = %#v, %v; want %#v, nil", result, runErr, want) }
	requests := repository.finalizeRequests()
	if len(requests) != 2 || !requests[0].Valid() || !requests[1].Valid() || requests[0].Outcome != mail.MailboxAccountProvisioningFinalizeSuccess || requests[1].Outcome != mail.MailboxAccountProvisioningFinalizeRetry || requests[1].Code != mail.MailboxAccountProvisioningCodeProvisioningRetry { t.Fatalf("finalize requests = %#v", requests) }
}

func TestMailboxAccountProvisioningWorkerFailsClosedForInvalidScanAndClaim(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	wrongID := provisioningWorkerClaim(t, provisioningWorkerIDTwo, now.Add(time.Minute), strings.Repeat("a", 64))
	badHash := provisioningWorkerClaim(t, provisioningWorkerIDOne, now.Add(time.Minute), "not-a-lowercase-hex-hash")
	for _, testCase := range []struct {
		name string
		scan func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error)
		claim func(context.Context, mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error)
		want error
	}{
		{name: "scan panic", scan: func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) { panic("scan") }, want: mail.ErrMailboxAccountProvisioningWorkerUnavailable},
		{name: "zero scan id", scan: func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) { return mail.MailboxAccountProvisioningScanResult{AccountIDs: []uuid.UUID{uuid.Nil}}, nil }, want: mail.ErrMailboxAccountProvisioningWorkerInvalid},
		{name: "duplicate scan id", scan: func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) { return mail.MailboxAccountProvisioningScanResult{AccountIDs: []uuid.UUID{provisioningWorkerIDOne, provisioningWorkerIDOne}}, nil }, want: mail.ErrMailboxAccountProvisioningWorkerInvalid},
		{name: "oversize scan", scan: func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) { return mail.MailboxAccountProvisioningScanResult{AccountIDs: make([]uuid.UUID, 21)}, nil }, want: mail.ErrMailboxAccountProvisioningWorkerTooLarge},
		{name: "claim error", scan: oneProvisioningWorkerID, claim: func(context.Context, mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error) { return mail.MailboxAccountProvisioningClaimResult{}, errors.New("claim") }, want: mail.ErrMailboxAccountProvisioningWorkerUnavailable},
		{name: "fence differs from scanned id", scan: oneProvisioningWorkerID, claim: claimedProvisioningWorkerResult(wrongID), want: mail.ErrMailboxAccountProvisioningWorkerUnavailable},
		{name: "control hash malformed", scan: oneProvisioningWorkerID, claim: claimedProvisioningWorkerResult(badHash), want: mail.ErrMailboxAccountProvisioningWorkerUnavailable},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &provisioningWorkerRepositoryFake{scan: testCase.scan, claim: testCase.claim}
			provisioner := &provisioningWorkerProvisionerFake{}
			worker, err := mail.NewMailboxAccountProvisioningWorker(repository, provisioner, &provisioningWorkerTimerFake{now: now}, mail.DefaultMailboxAccountProvisioningWorkerOptions())
			if err != nil { t.Fatal(err) }
			_, runErr := worker.RunOnce(context.Background())
			if !errors.Is(runErr, testCase.want) { t.Fatalf("RunOnce() error = %v, want %v", runErr, testCase.want) }
			if provisioner.calls() != 0 || len(repository.finalizeRequests()) != 0 { t.Fatalf("closed failure called provider/finalizer") }
		})
	}
}

func TestMailboxAccountProvisioningWorkerFinalizesAfterParentCancellationAndClipsLease(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	claim := provisioningWorkerClaim(t, provisioningWorkerIDOne, now.Add(time.Minute), strings.Repeat("a", 64))
	parent, cancelParent := context.WithCancel(context.Background())
	defer cancelParent()
	repository := &provisioningWorkerRepositoryFake{scan: oneProvisioningWorkerID, claim: claimedProvisioningWorkerResult(claim), finalize: func(ctx context.Context, request mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error) {
		if ctx.Err() != nil || request.Outcome != mail.MailboxAccountProvisioningFinalizeRetry || request.Code != mail.MailboxAccountProvisioningCodeUnavailable { t.Fatalf("detached finalization = ctx=%v request=%#v", ctx.Err(), request) }
		return mail.MailboxAccountProvisioningFinalizeResult{Outcome: request.Outcome, Code: request.Code}, nil
	}}
	provisioner := &provisioningWorkerProvisionerFake{ensure: func(context.Context, mail.AccountSpec) (mail.EngineAccount, error) { cancelParent(); return mail.EngineAccount{}, context.Canceled }}
	worker, err := mail.NewMailboxAccountProvisioningWorker(repository, provisioner, &provisioningWorkerTimerFake{now: now}, mail.DefaultMailboxAccountProvisioningWorkerOptions())
	if err != nil { t.Fatal(err) }
	result, runErr := worker.RunOnce(parent)
	want := mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderFailed: 1, FinalizedRetry: 1}
	if result != want || runErr != nil { t.Fatalf("RunOnce() = %#v, %v; want %#v, nil", result, runErr, want) }
}

var (
	provisioningWorkerIDOne = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	provisioningWorkerIDTwo = uuid.MustParse("22222222-2222-4222-8222-222222222222")
)

func provisioningWorkerClaim(t *testing.T, accountID uuid.UUID, deadline time.Time, hash string) mail.MailboxAccountProvisioningClaim {
	t.Helper()
	claim, ok := mail.NewMailboxAccountProvisioningClaim(mail.MailboxAccountProvisioningFence{AccountID: accountID, OrganizationID: uuid.MustParse("33333333-3333-4333-8333-333333333333"), DomainID: uuid.MustParse("44444444-4444-4444-8444-444444444444"), UserID: uuid.MustParse("55555555-5555-4555-8555-555555555555"), LeaseToken: uuid.MustParse("66666666-6666-4666-8666-666666666666"), LeaseDeadline: deadline, LeaseDuration: mail.MailboxEngineCommandLeaseDuration, ControlGeneration: 7, ControlHashVersion: 1, ControlHash: hash, ProvisioningKey: "mailbox-account-7", LocalPart: "ada", Domain: "example.test"})
	if !ok { t.Fatal("test claim must be valid") }
	return claim
}

func provisioningWorkerSuccess(spec mail.AccountSpec) mail.EngineAccount {
	return mail.EngineAccount{AccountID: "engine-account-7", State: mail.EngineAccountProvisioning, AppliedGeneration: spec.DesiredGeneration, AppliedHash: spec.DesiredHash, AliasIdentityIDs: []string{}}
}

func oneProvisioningWorkerID(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) { return mail.MailboxAccountProvisioningScanResult{AccountIDs: []uuid.UUID{provisioningWorkerIDOne}}, nil }

func claimedProvisioningWorkerResult(claim mail.MailboxAccountProvisioningClaim) func(context.Context, mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error) {
	return func(context.Context, mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error) { return mail.MailboxAccountProvisioningClaimResult{Outcome: mail.MailboxAccountProvisioningClaimed, Claim: claim}, nil }
}

type provisioningWorkerRepositoryFake struct {
	mu sync.Mutex
	scan func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error)
	claim func(context.Context, mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error)
	finalize func(context.Context, mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error)
	finalizes []mail.MailboxAccountProvisioningFinalizeRequest
	scanCount int
	claimCount int
	finalizeCount int
}

func (fake *provisioningWorkerRepositoryFake) Scan(ctx context.Context, request mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) { fake.mu.Lock(); fake.scanCount++; callback := fake.scan; fake.mu.Unlock(); if callback == nil { return mail.MailboxAccountProvisioningScanResult{}, nil }; return callback(ctx, request) }
func (fake *provisioningWorkerRepositoryFake) ClaimMailboxAccountProvisioning(ctx context.Context, request mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error) { fake.mu.Lock(); fake.claimCount++; callback := fake.claim; fake.mu.Unlock(); if callback == nil { return mail.MailboxAccountProvisioningClaimResult{Outcome: mail.MailboxAccountProvisioningNoop}, nil }; return callback(ctx, request) }
func (fake *provisioningWorkerRepositoryFake) Finalize(ctx context.Context, request mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error) { fake.mu.Lock(); fake.finalizeCount++; fake.finalizes = append(fake.finalizes, request); callback := fake.finalize; fake.mu.Unlock(); if callback == nil { return mail.MailboxAccountProvisioningFinalizeResult{Outcome: request.Outcome, Code: request.Code}, nil }; return callback(ctx, request) }
func (fake *provisioningWorkerRepositoryFake) finalizeRequests() []mail.MailboxAccountProvisioningFinalizeRequest { fake.mu.Lock(); defer fake.mu.Unlock(); return append([]mail.MailboxAccountProvisioningFinalizeRequest(nil), fake.finalizes...) }
func (fake *provisioningWorkerRepositoryFake) finalizeCalls() int { fake.mu.Lock(); defer fake.mu.Unlock(); return fake.finalizeCount }
func (fake *provisioningWorkerRepositoryFake) claimCalls() int { fake.mu.Lock(); defer fake.mu.Unlock(); return fake.claimCount }

type provisioningWorkerProvisionerFake struct { mu sync.Mutex; ensure func(context.Context, mail.AccountSpec) (mail.EngineAccount, error); count int }
func (fake *provisioningWorkerProvisionerFake) EnsureAccount(ctx context.Context, spec mail.AccountSpec) (mail.EngineAccount, error) { fake.mu.Lock(); fake.count++; callback := fake.ensure; fake.mu.Unlock(); if callback == nil { return provisioningWorkerSuccess(spec), nil }; return callback(ctx, spec) }
func (fake *provisioningWorkerProvisionerFake) calls() int { fake.mu.Lock(); defer fake.mu.Unlock(); return fake.count }

type provisioningWorkerTimerFake struct {
	now time.Time
	nowFn func() time.Time
	wait func(context.Context, time.Duration) error
	withDeadline func(context.Context, time.Time) (context.Context, context.CancelFunc)
}

func (fake *provisioningWorkerTimerFake) Now() time.Time { if fake.nowFn != nil { return fake.nowFn() }; return fake.now }
func (fake *provisioningWorkerTimerFake) Wait(ctx context.Context, duration time.Duration) error { if fake.wait == nil { return nil }; return fake.wait(ctx, duration) }
func (fake *provisioningWorkerTimerFake) WithDeadline(ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) { if fake.withDeadline != nil { return fake.withDeadline(ctx, deadline) }; return context.WithDeadline(ctx, deadline) }

func provisioningWorkerNew(t *testing.T, repository mail.MailboxAccountProvisioningRepository, provisioner mail.MailboxAccountProvisioner, timer mail.MailboxAccountProvisioningWorkerTimer) *mail.MailboxAccountProvisioningWorker {
	t.Helper()
	worker, err := mail.NewMailboxAccountProvisioningWorker(repository, provisioner, timer, mail.DefaultMailboxAccountProvisioningWorkerOptions())
	if err != nil { t.Fatal(err) }
	return worker
}
