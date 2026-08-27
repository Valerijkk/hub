package mail_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

func TestMailboxAccountProvisioningWorkerConstructorAdversarialMatrix(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	repository := &provisioningWorkerRepositoryFake{}
	provisioner := &provisioningWorkerProvisionerFake{}
	timer := &provisioningWorkerTimerFake{now: now}
	var typedRepository *provisioningWorkerRepositoryFake
	var typedProvisioner *provisioningWorkerProvisionerFake
	var typedTimer *provisioningWorkerTimerFake

	for _, testCase := range []struct {
		name       string
		repository mail.MailboxAccountProvisioningRepository
		provisioner mail.MailboxAccountProvisioner
		timer      mail.MailboxAccountProvisioningWorkerTimer
		options    mail.MailboxAccountProvisioningWorkerOptions
		wantErr    error
	}{
		{name: "enabled typed nil repository", repository: typedRepository, provisioner: provisioner, timer: timer, options: mail.MailboxAccountProvisioningWorkerOptions{Enabled: true}, wantErr: mail.ErrMailboxAccountProvisioningWorkerInvalid},
		{name: "enabled typed nil provisioner", repository: repository, provisioner: typedProvisioner, timer: timer, options: mail.MailboxAccountProvisioningWorkerOptions{Enabled: true}, wantErr: mail.ErrMailboxAccountProvisioningWorkerInvalid},
		{name: "enabled typed nil timer", repository: repository, provisioner: provisioner, timer: typedTimer, options: mail.MailboxAccountProvisioningWorkerOptions{Enabled: true}, wantErr: mail.ErrMailboxAccountProvisioningWorkerInvalid},
		{name: "disabled repository present", repository: repository, options: mail.MailboxAccountProvisioningWorkerOptions{}, wantErr: mail.ErrMailboxAccountProvisioningWorkerInvalid},
		{name: "disabled provisioner present", provisioner: provisioner, options: mail.MailboxAccountProvisioningWorkerOptions{}, wantErr: mail.ErrMailboxAccountProvisioningWorkerInvalid},
		{name: "disabled timer present", timer: timer, options: mail.MailboxAccountProvisioningWorkerOptions{}, wantErr: mail.ErrMailboxAccountProvisioningWorkerInvalid},
		{name: "disabled numeric option", options: mail.MailboxAccountProvisioningWorkerOptions{ScanInterval: time.Second}, wantErr: mail.ErrMailboxAccountProvisioningWorkerInvalid},
		{name: "subsecond option", repository: repository, provisioner: provisioner, timer: timer, options: mail.MailboxAccountProvisioningWorkerOptions{Enabled: true, ScanInterval: 500 * time.Millisecond}, wantErr: mail.ErrMailboxAccountProvisioningWorkerInvalid},
		{name: "scan bound", repository: repository, provisioner: provisioner, timer: timer, options: mail.MailboxAccountProvisioningWorkerOptions{Enabled: true, ScanInterval: 5*time.Minute + time.Second}, wantErr: mail.ErrMailboxAccountProvisioningWorkerTooLarge},
		{name: "provider bound", repository: repository, provisioner: provisioner, timer: timer, options: mail.MailboxAccountProvisioningWorkerOptions{Enabled: true, ProviderTimeout: 31 * time.Second}, wantErr: mail.ErrMailboxAccountProvisioningWorkerTooLarge},
		{name: "finalize bound", repository: repository, provisioner: provisioner, timer: timer, options: mail.MailboxAccountProvisioningWorkerOptions{Enabled: true, FinalizeTimeout: 11 * time.Second}, wantErr: mail.ErrMailboxAccountProvisioningWorkerTooLarge},
		{name: "batch bound", repository: repository, provisioner: provisioner, timer: timer, options: mail.MailboxAccountProvisioningWorkerOptions{Enabled: true, BatchSize: 51}, wantErr: mail.ErrMailboxAccountProvisioningWorkerTooLarge},
		{name: "maximum valid budget", repository: repository, provisioner: provisioner, timer: timer, options: mail.MailboxAccountProvisioningWorkerOptions{Enabled: true, ScanInterval: 5 * time.Minute, ProviderTimeout: 30 * time.Second, FinalizeTimeout: 10 * time.Second, BatchSize: 50}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			worker, err := mail.NewMailboxAccountProvisioningWorker(testCase.repository, testCase.provisioner, testCase.timer, testCase.options)
			if testCase.wantErr == nil {
				if worker == nil || err != nil {
					t.Fatalf("constructor = %#v, %v", worker, err)
				}
				return
			}
			if worker != nil || !errors.Is(err, testCase.wantErr) {
				t.Fatalf("constructor = %#v, %v; want %v", worker, err, testCase.wantErr)
			}
		})
	}

	disabled, err := mail.NewMailboxAccountProvisioningWorker(typedRepository, typedProvisioner, typedTimer, mail.MailboxAccountProvisioningWorkerOptions{})
	if err != nil || disabled == nil {
		t.Fatalf("typed-nil disabled constructor = %#v, %v", disabled, err)
	}
	if result, runErr := disabled.RunOnce(context.Background()); result != (mail.MailboxAccountProvisioningWorkerRunResult{}) || runErr != nil {
		t.Fatalf("typed-nil disabled RunOnce = %#v, %v", result, runErr)
	}
}

func TestMailboxAccountProvisioningWorkerNoopAndLocalGateNoIO(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	for _, testCase := range []struct {
		name  string
		claim mail.MailboxAccountProvisioningClaimResult
		want  mail.MailboxAccountProvisioningWorkerRunResult
	}{
		{name: "noop", claim: mail.MailboxAccountProvisioningClaimResult{Outcome: mail.MailboxAccountProvisioningNoop}, want: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Noop: 1}},
		{name: "local gate", claim: mail.MailboxAccountProvisioningClaimResult{Outcome: mail.MailboxAccountProvisioningLocalGate, Code: mail.MailboxAccountProvisioningCodeDomainRequired}, want: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, LocalGated: 1}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &provisioningWorkerRepositoryFake{scan: oneProvisioningWorkerID, claim: func(context.Context, mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error) { return testCase.claim, nil }}
			provisioner := &provisioningWorkerProvisionerFake{ensure: func(context.Context, mail.AccountSpec) (mail.EngineAccount, error) { t.Fatal("provider must not be called"); return mail.EngineAccount{}, nil }}
			worker := provisioningWorkerNew(t, repository, provisioner, &provisioningWorkerTimerFake{now: now})
			result, err := worker.RunOnce(context.Background())
			adversarialWorkerResult(t, result, testCase.want, err, nil)
			if provisioner.calls() != 0 || repository.finalizeCalls() != 0 {
				t.Fatalf("unexpected external calls: provider=%d finalizer=%d", provisioner.calls(), repository.finalizeCalls())
			}
		})
	}
}

func TestMailboxAccountProvisioningWorkerScanAndClaimFailureMatrix(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	wrongID := provisioningWorkerClaim(t, provisioningWorkerIDTwo, now.Add(time.Minute), strings.Repeat("a", 64))
	malformedHash := provisioningWorkerClaim(t, provisioningWorkerIDOne, now.Add(time.Minute), "not-a-lowercase-hex-hash")
	nonCanonicalHash := provisioningWorkerClaim(t, provisioningWorkerIDOne, now.Add(time.Minute), strings.Repeat("A", 64))
	for _, testCase := range []struct {
		name  string
		scan  func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error)
		claim func(context.Context, mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error)
		want  error
	}{
		{name: "scan error", scan: func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) { return mail.MailboxAccountProvisioningScanResult{}, errors.New("scan private detail") }, want: mail.ErrMailboxAccountProvisioningWorkerUnavailable},
		{name: "scan panic", scan: func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) { panic("scan") }, want: mail.ErrMailboxAccountProvisioningWorkerUnavailable},
		{name: "zero scan id", scan: func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) { return mail.MailboxAccountProvisioningScanResult{AccountIDs: []uuid.UUID{uuid.Nil}}, nil }, want: mail.ErrMailboxAccountProvisioningWorkerInvalid},
		{name: "duplicate scan id", scan: func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) { return mail.MailboxAccountProvisioningScanResult{AccountIDs: []uuid.UUID{provisioningWorkerIDOne, provisioningWorkerIDOne}}, nil }, want: mail.ErrMailboxAccountProvisioningWorkerInvalid},
		{name: "oversize scan", scan: func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) { return mail.MailboxAccountProvisioningScanResult{AccountIDs: make([]uuid.UUID, 21)}, nil }, want: mail.ErrMailboxAccountProvisioningWorkerTooLarge},
		{name: "claim error", scan: oneProvisioningWorkerID, claim: func(context.Context, mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error) { return mail.MailboxAccountProvisioningClaimResult{}, errors.New("claim private detail") }, want: mail.ErrMailboxAccountProvisioningWorkerUnavailable},
		{name: "claim panic", scan: oneProvisioningWorkerID, claim: func(context.Context, mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error) { panic("claim") }, want: mail.ErrMailboxAccountProvisioningWorkerUnavailable},
		{name: "invalid claim", scan: oneProvisioningWorkerID, claim: func(context.Context, mail.MailboxAccountProvisioningClaimRequest) (mail.MailboxAccountProvisioningClaimResult, error) { return mail.MailboxAccountProvisioningClaimResult{}, nil }, want: mail.ErrMailboxAccountProvisioningWorkerUnavailable},
		{name: "mismatched fence", scan: oneProvisioningWorkerID, claim: claimedProvisioningWorkerResult(wrongID), want: mail.ErrMailboxAccountProvisioningWorkerUnavailable},
		{name: "malformed hash", scan: oneProvisioningWorkerID, claim: claimedProvisioningWorkerResult(malformedHash), want: mail.ErrMailboxAccountProvisioningWorkerUnavailable},
		{name: "noncanonical hash", scan: oneProvisioningWorkerID, claim: claimedProvisioningWorkerResult(nonCanonicalHash), want: mail.ErrMailboxAccountProvisioningWorkerUnavailable},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &provisioningWorkerRepositoryFake{scan: testCase.scan, claim: testCase.claim}
			provisioner := &provisioningWorkerProvisionerFake{ensure: func(context.Context, mail.AccountSpec) (mail.EngineAccount, error) { t.Fatal("provider must not be called"); return mail.EngineAccount{}, nil }}
			worker := provisioningWorkerNew(t, repository, provisioner, &provisioningWorkerTimerFake{now: now})
			_, err := worker.RunOnce(context.Background())
			if !errors.Is(err, testCase.want) || provisioner.calls() != 0 || repository.finalizeCalls() != 0 {
				t.Fatalf("closed failure = err:%v provider:%d finalizer:%d", err, provisioner.calls(), repository.finalizeCalls())
			}
		})
	}
}

func TestMailboxAccountProvisioningWorkerProviderMatrixAndExactDeadlines(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	for _, testCase := range []struct {
		name       string
		deadline   time.Time
		ensure     func(context.Context, mail.AccountSpec) (mail.EngineAccount, error)
		wantCalls  int
		wantCode   mail.MailboxAccountProvisioningCode
		wantResult mail.MailboxAccountProvisioningWorkerRunResult
	}{
		{name: "exact success", deadline: now.Add(time.Minute), ensure: func(ctx context.Context, spec mail.AccountSpec) (mail.EngineAccount, error) {
			deadline, ok := ctx.Deadline()
			if !ok || !deadline.Equal(now.Add(30*time.Second)) || spec.ProvisioningKey != "mailbox-account-7" || spec.LocalPart != "ada" || spec.Domain != "example.test" || spec.DesiredGeneration != 7 || spec.DesiredHash != strings.Repeat("a", 64) || !spec.Enabled || spec.Primary.LocalPart != "ada" || spec.Primary.Domain != "example.test" || !spec.Primary.Enabled || len(spec.Aliases) != 0 {
				t.Fatalf("provider context/spec = %v, %v, %#v", deadline, ok, spec)
			}
			return provisioningWorkerSuccess(spec), nil
		}, wantCalls: 1, wantResult: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderSucceeded: 1, FinalizedSuccess: 1}},
		{name: "normal provider error", deadline: now.Add(time.Minute), ensure: func(context.Context, mail.AccountSpec) (mail.EngineAccount, error) { return mail.EngineAccount{}, errors.New("provider private detail") }, wantCalls: 1, wantCode: mail.MailboxAccountProvisioningCodeProvisioningRetry, wantResult: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderFailed: 1, FinalizedRetry: 1}},
		{name: "provider timeout", deadline: now.Add(time.Minute), ensure: func(context.Context, mail.AccountSpec) (mail.EngineAccount, error) { return mail.EngineAccount{}, context.DeadlineExceeded }, wantCalls: 1, wantCode: mail.MailboxAccountProvisioningCodeUnavailable, wantResult: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderFailed: 1, FinalizedRetry: 1}},
		{name: "provider panic", deadline: now.Add(time.Minute), ensure: func(context.Context, mail.AccountSpec) (mail.EngineAccount, error) { panic("provider") }, wantCalls: 1, wantCode: mail.MailboxAccountProvisioningCodeUnavailable, wantResult: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderFailed: 1, FinalizedRetry: 1}},
		{name: "invalid provider success", deadline: now.Add(time.Minute), ensure: func(context.Context, mail.AccountSpec) (mail.EngineAccount, error) { return mail.EngineAccount{}, nil }, wantCalls: 1, wantCode: mail.MailboxAccountProvisioningCodeUnavailable, wantResult: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderFailed: 1, FinalizedRetry: 1}},
		{name: "no provider window", deadline: now.Add(11 * time.Second), ensure: func(context.Context, mail.AccountSpec) (mail.EngineAccount, error) { t.Fatal("provider must not be called"); return mail.EngineAccount{}, nil }, wantCalls: 0, wantCode: mail.MailboxAccountProvisioningCodeUnavailable, wantResult: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderFailed: 1, FinalizedRetry: 1}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			claim := provisioningWorkerClaim(t, provisioningWorkerIDOne, testCase.deadline, strings.Repeat("a", 64))
			repository := &provisioningWorkerRepositoryFake{scan: oneProvisioningWorkerID, claim: claimedProvisioningWorkerResult(claim), finalize: func(ctx context.Context, request mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error) {
				deadline, ok := ctx.Deadline()
				if !ok || !deadline.Equal(now.Add(10*time.Second)) {
					t.Fatalf("finalize deadline = %v, %v", deadline, ok)
				}
				return mail.MailboxAccountProvisioningFinalizeResult{Outcome: request.Outcome, Code: request.Code}, nil
			}}
			provisioner := &provisioningWorkerProvisionerFake{ensure: testCase.ensure}
			worker := provisioningWorkerNew(t, repository, provisioner, &provisioningWorkerTimerFake{now: now})
			result, err := worker.RunOnce(context.Background())
			adversarialWorkerResult(t, result, testCase.wantResult, err, nil)
			if provisioner.calls() != testCase.wantCalls || repository.finalizeCalls() != 1 || len(repository.finalizeRequests()) != 1 || repository.finalizeRequests()[0].Code != testCase.wantCode {
				t.Fatalf("finalize request = %#v", repository.finalizeRequests())
			}
		})
	}
}

func TestMailboxAccountProvisioningWorkerExpiredAndMalformedContexts(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	claim := provisioningWorkerClaim(t, provisioningWorkerIDOne, now.Add(time.Minute), strings.Repeat("a", 64))
	for _, testCase := range []struct {
		name         string
		withDeadline func(context.Context, time.Time) (context.Context, context.CancelFunc)
		want         mail.MailboxAccountProvisioningWorkerRunResult
		wantErr      error
		providerIO   int
		finalizerIO  int
	}{
		{name: "expired provider but not late deadline", withDeadline: func(ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) { if deadline.Equal(now.Add(30*time.Second)) { return context.WithDeadline(ctx, now.Add(-time.Second)) }; return context.WithDeadline(ctx, deadline) }, want: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderFailed: 1, FinalizedRetry: 1}, providerIO: 0, finalizerIO: 1},
		{name: "expired finalize", withDeadline: func(ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) { if deadline.Equal(now.Add(10*time.Second)) { return context.WithDeadline(ctx, now.Add(-time.Second)) }; return context.WithDeadline(ctx, deadline) }, want: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderSucceeded: 1, Unfinalized: 1}, wantErr: mail.ErrMailboxAccountProvisioningWorkerPartialFailure, providerIO: 1, finalizerIO: 0},
		{name: "provider no deadline", withDeadline: adversarialDeadlineCall(func(call int, ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) { if call == 1 { return context.Background(), func() {} }; return context.WithDeadline(ctx, deadline) }), want: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderFailed: 1, FinalizedRetry: 1}, providerIO: 0, finalizerIO: 1},
		{name: "provider late deadline", withDeadline: adversarialDeadlineCall(func(call int, ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) { if call == 1 { return context.WithDeadline(ctx, deadline.Add(time.Second)) }; return context.WithDeadline(ctx, deadline) }), want: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderFailed: 1, FinalizedRetry: 1}, providerIO: 0, finalizerIO: 1},
		{name: "finalize no deadline", withDeadline: adversarialDeadlineCall(func(call int, ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) { if call == 2 { return context.Background(), func() {} }; return context.WithDeadline(ctx, deadline) }), want: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderSucceeded: 1, Unfinalized: 1}, wantErr: mail.ErrMailboxAccountProvisioningWorkerPartialFailure, providerIO: 1, finalizerIO: 0},
		{name: "finalize late deadline", withDeadline: adversarialDeadlineCall(func(call int, ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) { if call == 2 { return context.WithDeadline(ctx, deadline.Add(time.Second)) }; return context.WithDeadline(ctx, deadline) }), want: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderSucceeded: 1, Unfinalized: 1}, wantErr: mail.ErrMailboxAccountProvisioningWorkerPartialFailure, providerIO: 1, finalizerIO: 0},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &provisioningWorkerRepositoryFake{scan: oneProvisioningWorkerID, claim: claimedProvisioningWorkerResult(claim)}
			provisioner := &provisioningWorkerProvisionerFake{}
			worker := provisioningWorkerNew(t, repository, provisioner, &provisioningWorkerTimerFake{now: now, withDeadline: testCase.withDeadline})
			result, err := worker.RunOnce(context.Background())
			adversarialWorkerResult(t, result, testCase.want, err, testCase.wantErr)
			if provisioner.calls() != testCase.providerIO || repository.finalizeCalls() != testCase.finalizerIO {
				t.Fatalf("external calls = provider:%d finalizer:%d", provisioner.calls(), repository.finalizeCalls())
			}
		})
	}
}

func TestMailboxAccountProvisioningWorkerFinalizeCancelAndRunMatrix(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	claim := provisioningWorkerClaim(t, provisioningWorkerIDOne, now.Add(time.Minute), strings.Repeat("a", 64))
	for _, testCase := range []struct {
		name     string
		ensure   func(context.Context, mail.AccountSpec) (mail.EngineAccount, error)
		finalize func(context.Context, mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error)
		want     mail.MailboxAccountProvisioningWorkerRunResult
	}{
		{name: "outcome mismatch", ensure: func(_ context.Context, spec mail.AccountSpec) (mail.EngineAccount, error) { return provisioningWorkerSuccess(spec), nil }, finalize: func(context.Context, mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error) { return mail.MailboxAccountProvisioningFinalizeResult{Outcome: mail.MailboxAccountProvisioningFinalizeRetry, Code: mail.MailboxAccountProvisioningCodeUnavailable}, nil }, want: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderSucceeded: 1, Unfinalized: 1}},
		{name: "other valid retry code", ensure: func(context.Context, mail.AccountSpec) (mail.EngineAccount, error) { return mail.EngineAccount{}, errors.New("provider") }, finalize: func(context.Context, mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error) { return mail.MailboxAccountProvisioningFinalizeResult{Outcome: mail.MailboxAccountProvisioningFinalizeRetry, Code: mail.MailboxAccountProvisioningCodeUnavailable}, nil }, want: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderFailed: 1, Unfinalized: 1}},
		{name: "finalize error", ensure: func(_ context.Context, spec mail.AccountSpec) (mail.EngineAccount, error) { return provisioningWorkerSuccess(spec), nil }, finalize: func(context.Context, mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error) { return mail.MailboxAccountProvisioningFinalizeResult{}, errors.New("finalize private detail") }, want: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderSucceeded: 1, Unfinalized: 1}},
		{name: "finalize panic", ensure: func(_ context.Context, spec mail.AccountSpec) (mail.EngineAccount, error) { return provisioningWorkerSuccess(spec), nil }, finalize: func(context.Context, mail.MailboxAccountProvisioningFinalizeRequest) (mail.MailboxAccountProvisioningFinalizeResult, error) { panic("finalize") }, want: mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderSucceeded: 1, Unfinalized: 1}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &provisioningWorkerRepositoryFake{scan: oneProvisioningWorkerID, claim: claimedProvisioningWorkerResult(claim), finalize: testCase.finalize}
			worker := provisioningWorkerNew(t, repository, &provisioningWorkerProvisionerFake{ensure: testCase.ensure}, &provisioningWorkerTimerFake{now: now})
			result, err := worker.RunOnce(context.Background())
			adversarialWorkerResult(t, result, testCase.want, err, mail.ErrMailboxAccountProvisioningWorkerPartialFailure)
		})
	}

	t.Run("provider cancel panic still finalizes", func(t *testing.T) {
		timer := &provisioningWorkerTimerFake{now: now, withDeadline: adversarialDeadlineCall(func(call int, ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) { child, cancel := context.WithDeadline(ctx, deadline); if call == 1 { return child, func() { cancel(); panic("cancel") } }; return child, cancel })}
		repository := &provisioningWorkerRepositoryFake{scan: oneProvisioningWorkerID, claim: claimedProvisioningWorkerResult(claim)}
		worker := provisioningWorkerNew(t, repository, &provisioningWorkerProvisionerFake{}, timer)
		result, err := worker.RunOnce(context.Background())
		adversarialWorkerResult(t, result, mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderFailed: 1, FinalizedRetry: 1}, err, nil)
	})

	t.Run("finalize cancel panic is unfinalized", func(t *testing.T) {
		timer := &provisioningWorkerTimerFake{now: now, withDeadline: adversarialDeadlineCall(func(call int, ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) { child, cancel := context.WithDeadline(ctx, deadline); if call == 2 { return child, func() { cancel(); panic("cancel") } }; return child, cancel })}
		repository := &provisioningWorkerRepositoryFake{scan: oneProvisioningWorkerID, claim: claimedProvisioningWorkerResult(claim)}
		worker := provisioningWorkerNew(t, repository, &provisioningWorkerProvisionerFake{}, timer)
		result, err := worker.RunOnce(context.Background())
		adversarialWorkerResult(t, result, mail.MailboxAccountProvisioningWorkerRunResult{Due: 1, Claimed: 1, ProviderSucceeded: 1, Unfinalized: 1}, err, mail.ErrMailboxAccountProvisioningWorkerPartialFailure)
	})

	t.Run("wait error and cancellation", func(t *testing.T) {
		worker := provisioningWorkerNew(t, &provisioningWorkerRepositoryFake{}, &provisioningWorkerProvisionerFake{}, &provisioningWorkerTimerFake{now: now, wait: func(context.Context, time.Duration) error { return errors.New("wait private detail") }})
		if err := worker.Run(context.Background()); !errors.Is(err, mail.ErrMailboxAccountProvisioningWorkerUnavailable) { t.Fatalf("wait error = %v", err) }
		parent, cancel := context.WithCancel(context.Background())
		worker = provisioningWorkerNew(t, &provisioningWorkerRepositoryFake{}, &provisioningWorkerProvisionerFake{}, &provisioningWorkerTimerFake{now: now, wait: func(context.Context, time.Duration) error { cancel(); return context.Canceled }})
		if err := worker.Run(parent); err != nil { t.Fatalf("wait cancellation = %v", err) }
	})

	t.Run("no item after parent cancellation", func(t *testing.T) {
		parent, cancel := context.WithCancel(context.Background())
		defer cancel()
		repository := &provisioningWorkerRepositoryFake{scan: func(context.Context, mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) { return mail.MailboxAccountProvisioningScanResult{AccountIDs: []uuid.UUID{provisioningWorkerIDOne, provisioningWorkerIDTwo}}, nil }, claim: claimedProvisioningWorkerResult(claim)}
		provisioner := &provisioningWorkerProvisionerFake{ensure: func(context.Context, mail.AccountSpec) (mail.EngineAccount, error) { cancel(); return mail.EngineAccount{}, context.Canceled }}
		worker := provisioningWorkerNew(t, repository, provisioner, &provisioningWorkerTimerFake{now: now})
		result, err := worker.RunOnce(parent)
		adversarialWorkerResult(t, result, mail.MailboxAccountProvisioningWorkerRunResult{Due: 2, Claimed: 1, ProviderFailed: 1, FinalizedRetry: 1}, err, nil)
		if repository.claimCalls() != 1 { t.Fatalf("claims after cancellation = %d", repository.claimCalls()) }
	})
}

func TestMailboxAccountProvisioningWorkerGuardAndClosedError(t *testing.T) {
	now := time.Date(2030, time.January, 2, 3, 4, 5, 0, time.UTC)
	var worker *mail.MailboxAccountProvisioningWorker
	repository := &provisioningWorkerRepositoryFake{scan: func(ctx context.Context, request mail.MailboxAccountProvisioningScanRequest) (mail.MailboxAccountProvisioningScanResult, error) {
		if _, err := worker.RunOnce(ctx); !errors.Is(err, mail.ErrMailboxAccountProvisioningWorkerInvalid) { t.Fatalf("guard error = %v", err) }
		return mail.MailboxAccountProvisioningScanResult{}, errors.New("account=ada@example.test token=private")
	}}
	worker = provisioningWorkerNew(t, repository, &provisioningWorkerProvisionerFake{}, &provisioningWorkerTimerFake{now: now})
	result, err := worker.RunOnce(context.Background())
	if result != (mail.MailboxAccountProvisioningWorkerRunResult{}) || !errors.Is(err, mail.ErrMailboxAccountProvisioningWorkerUnavailable) || err.Error() != mail.ErrMailboxAccountProvisioningWorkerUnavailable.Error() || strings.Contains(err.Error(), "ada") || strings.Contains(err.Error(), "private") { t.Fatalf("closed result/error = %#v, %v", result, err) }
}

func adversarialDeadlineCall(callback func(int, context.Context, time.Time) (context.Context, context.CancelFunc)) func(context.Context, time.Time) (context.Context, context.CancelFunc) {
	calls := 0
	return func(ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) { calls++; return callback(calls, ctx, deadline) }
}

func adversarialWorkerResult(t *testing.T, got, want mail.MailboxAccountProvisioningWorkerRunResult, gotErr, wantErr error) {
	t.Helper()
	if got != want || (wantErr == nil && gotErr != nil) || (wantErr != nil && !errors.Is(gotErr, wantErr)) {
		t.Fatalf("result = %#v, %v; want %#v, %v", got, gotErr, want, wantErr)
	}
	if got.Noop+got.LocalGated+got.Claimed > got.Due || got.ProviderSucceeded+got.ProviderFailed != got.Claimed || got.FinalizedSuccess+got.FinalizedRetry+got.Unfinalized != got.Claimed {
		t.Fatalf("count invariants violated: %#v", got)
	}
}
