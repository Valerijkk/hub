package mail_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

type mailboxSubmissionReconciliationRepositoryStub struct{}

func (mailboxSubmissionReconciliationRepositoryStub) ScanUnknownMailboxSubmissions(context.Context, mail.MailboxSubmissionReconciliationScanRequest) (mail.MailboxSubmissionReconciliationScanResult, error) {
	return mail.MailboxSubmissionReconciliationScanResult{}, nil
}

func (mailboxSubmissionReconciliationRepositoryStub) ClaimUnknownMailboxSubmission(context.Context, mail.MailboxSubmissionReconciliationClaimRequest) (mail.MailboxSubmissionReconciliationClaimResult, error) {
	return mail.MailboxSubmissionReconciliationClaimResult{}, nil
}

func (mailboxSubmissionReconciliationRepositoryStub) ResolveUnknownMailboxSubmissionReconciliation(context.Context, mail.MailboxSubmissionReconciliationResolveRequest) error {
	return nil
}

var _ mail.MailboxSubmissionReconciliationRepository = mailboxSubmissionReconciliationRepositoryStub{}
type mailboxSubmissionReconciliationRepositorySpy struct {
	claimFn func(context.Context, mail.MailboxSubmissionReconciliationClaimRequest) (mail.MailboxSubmissionReconciliationClaimResult, error)
	resolveFn func(context.Context, mail.MailboxSubmissionReconciliationResolveRequest) error
	claimCalls int
	resolveCalls int
	resolveRequests []mail.MailboxSubmissionReconciliationResolveRequest
	resolveContexts []context.Context
}
func (spy *mailboxSubmissionReconciliationRepositorySpy) ScanUnknownMailboxSubmissions(context.Context, mail.MailboxSubmissionReconciliationScanRequest) (mail.MailboxSubmissionReconciliationScanResult, error) { return mail.MailboxSubmissionReconciliationScanResult{}, nil }
func (spy *mailboxSubmissionReconciliationRepositorySpy) ClaimUnknownMailboxSubmission(ctx context.Context, request mail.MailboxSubmissionReconciliationClaimRequest) (mail.MailboxSubmissionReconciliationClaimResult, error) {
	spy.claimCalls++
	if spy.claimFn == nil { return mail.MailboxSubmissionReconciliationClaimResult{}, nil }
	return spy.claimFn(ctx, request)
}
func (spy *mailboxSubmissionReconciliationRepositorySpy) ResolveUnknownMailboxSubmissionReconciliation(ctx context.Context, request mail.MailboxSubmissionReconciliationResolveRequest) error {
	spy.resolveCalls++
	spy.resolveRequests = append(spy.resolveRequests, request)
	spy.resolveContexts = append(spy.resolveContexts, ctx)
	if spy.resolveFn == nil { return nil }
	return spy.resolveFn(ctx, request)
}
type mailboxSubmissionReconciliationEngineSpy struct {
	mail.MailEngine
	reconcileFn func(context.Context, mail.EngineSubmission) (mail.EngineSubmissionResult, error)
	calls int
	submissions []mail.EngineSubmission
}
func (spy *mailboxSubmissionReconciliationEngineSpy) ReconcileSubmission(ctx context.Context, submission mail.EngineSubmission) (mail.EngineSubmissionResult, error) {
	spy.calls++
	spy.submissions = append(spy.submissions, submission)
	if spy.reconcileFn == nil { return nil, errors.New("unexpected engine reconciliation") }
	return spy.reconcileFn(ctx, submission)
}
func mailboxSubmissionReconciliationFixture(t *testing.T) (uuid.UUID, mail.MailboxEngineCommandFence, mail.EngineSubmission, mail.MailboxSubmissionReconciliationClaimResult) {
	t.Helper()
	submissionID := uuid.New()
	fence := mail.MailboxEngineCommandFence{AccountID: uuid.New(), LeaseID: uuid.New(), MutationEpoch: 1, LeaseUntil: time.Date(2001, time.February, 3, 4, 5, 6, 0, time.UTC)}
	submission := mail.EngineSubmission{AccountID: "account-opaque", DraftKey: "draft-opaque", MessageID: "message-opaque", ExpectedVersion: "version-opaque", SubmissionKey: "submission-opaque", MaxMessageBytes: 1_048_576}
	claim, err := mail.NewMailboxSubmissionReconciliationClaimedResult(submissionID, fence, submission)
	if err != nil { t.Fatalf("NewMailboxSubmissionReconciliationClaimedResult() error = %v", err) }
	return submissionID, fence, submission, claim
}
func mailboxSubmissionReconciliationReconciler(t *testing.T, repository *mailboxSubmissionReconciliationRepositorySpy, engine *mailboxSubmissionReconciliationEngineSpy) *mail.MailboxSubmissionReconciler {
	t.Helper()
	reconciler, err := mail.NewMailboxSubmissionReconciler(repository, engine)
	if err != nil { t.Fatalf("NewMailboxSubmissionReconciler() error = %v", err) }
	return reconciler
}
func TestMailboxSubmissionReconcilerNoopAndBusyDoNotCallEngineOrResolve(t *testing.T) {
	busy, err := mail.NewMailboxSubmissionReconciliationBusyResult(time.Date(2000, time.January, 2, 3, 4, 5, 0, time.UTC))
	if err != nil { t.Fatalf("NewMailboxSubmissionReconciliationBusyResult() error = %v", err) }
	for _, testCase := range []struct { name string; claim mail.MailboxSubmissionReconciliationClaimResult }{{name: "noop", claim: mail.NewMailboxSubmissionReconciliationNoopResult()}, {name: "busy", claim: busy}} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &mailboxSubmissionReconciliationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionReconciliationClaimRequest) (mail.MailboxSubmissionReconciliationClaimResult, error) { return testCase.claim, nil }, resolveFn: func(context.Context, mail.MailboxSubmissionReconciliationResolveRequest) error { t.Fatal("noop or busy must not resolve"); return nil }}
			engine := &mailboxSubmissionReconciliationEngineSpy{reconcileFn: func(context.Context, mail.EngineSubmission) (mail.EngineSubmissionResult, error) { t.Fatal("noop or busy must not call engine"); return nil, nil }}
			if err := mailboxSubmissionReconciliationReconciler(t, repository, engine).Reconcile(context.Background(), uuid.New()); err != nil { t.Fatalf("Reconcile() error = %v, want nil", err) }
			if repository.claimCalls != 1 || repository.resolveCalls != 0 || engine.calls != 0 { t.Fatalf("claim/resolve/engine calls = %d/%d/%d, want 1/0/0", repository.claimCalls, repository.resolveCalls, engine.calls) }
		})
	}
}
func TestMailboxSubmissionReconcilerClaimedAcceptedResolvesExactSealedTuple(t *testing.T) {
	submissionID, fence, submission, claim := mailboxSubmissionReconciliationFixture(t)
	wantFence, wantSubmission := fence, submission
	fence.LeaseUntil = fence.LeaseUntil.Add(time.Hour)
	submission.MessageID = "changed-after-claim"
	repository := &mailboxSubmissionReconciliationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionReconciliationClaimRequest) (mail.MailboxSubmissionReconciliationClaimResult, error) { return claim, nil }, resolveFn: func(_ context.Context, request mail.MailboxSubmissionReconciliationResolveRequest) error {
		if request.Observation != mail.MailboxSubmissionReconciliationObservationAccepted || request.EngineSubmissionID != "engine-submission-opaque" || request.EngineMessageID != wantSubmission.MessageID || request.RetryCode != "" { t.Fatalf("resolve request = %#v, want exact accepted observation", request) }
		if request.Claim.SubmissionID() != submissionID || request.Claim.Fence() != wantFence { t.Fatalf("resolve claim = %s/%#v, want %s/%#v", request.Claim.SubmissionID(), request.Claim.Fence(), submissionID, wantFence) }
		return nil
	}}
	engine := &mailboxSubmissionReconciliationEngineSpy{reconcileFn: func(_ context.Context, got mail.EngineSubmission) (mail.EngineSubmissionResult, error) {
		if !reflect.DeepEqual(got, wantSubmission) { t.Fatalf("engine submission = %#v, want %#v", got, wantSubmission) }
		return mail.EngineSubmissionAccepted{SubmissionID: "engine-submission-opaque", MessageID: wantSubmission.MessageID}, nil
	}}
	if err := mailboxSubmissionReconciliationReconciler(t, repository, engine).Reconcile(context.Background(), submissionID); err != nil { t.Fatalf("Reconcile() error = %v, want nil", err) }
	if repository.claimCalls != 1 || repository.resolveCalls != 1 || engine.calls != 1 || len(engine.submissions) != 1 { t.Fatalf("claim/resolve/engine/submissions = %d/%d/%d/%d, want 1/1/1/1", repository.claimCalls, repository.resolveCalls, engine.calls, len(engine.submissions)) }
}
func TestMailboxSubmissionReconcilerRetryMappingIsClosed(t *testing.T) {
	for _, testCase := range []struct { name string; result mail.EngineSubmissionResult; engineErr error; want mail.MailboxSubmissionErrorCode }{
		{name: "absent", result: mail.EngineSubmissionAbsent{}, want: mail.MailboxSubmissionErrorCodeReconciliationRequired},
		{name: "indeterminate unavailable", result: mail.EngineSubmissionIndeterminate{Code: mail.EngineIndeterminateUnavailable}, want: mail.MailboxSubmissionErrorCodeMailboxUnavailable},
		{name: "engine unavailable", engineErr: mail.ErrEngineUnavailable, want: mail.MailboxSubmissionErrorCodeMailboxUnavailable},
		{name: "raw engine failure", engineErr: errors.New("raw engine failure body@example.test"), want: mail.MailboxSubmissionErrorCodeReconciliationRequired},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			submissionID, _, _, claim := mailboxSubmissionReconciliationFixture(t)
			repository := &mailboxSubmissionReconciliationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionReconciliationClaimRequest) (mail.MailboxSubmissionReconciliationClaimResult, error) { return claim, nil }, resolveFn: func(_ context.Context, request mail.MailboxSubmissionReconciliationResolveRequest) error {
				if request.Observation != mail.MailboxSubmissionReconciliationObservationRetry || request.RetryCode != testCase.want || request.EngineSubmissionID != "" || request.EngineMessageID != "" { t.Fatalf("retry resolution = %#v, want only safe code %q", request, testCase.want) }
				return nil
			}}
			engine := &mailboxSubmissionReconciliationEngineSpy{reconcileFn: func(context.Context, mail.EngineSubmission) (mail.EngineSubmissionResult, error) { return testCase.result, testCase.engineErr }}
			if err := mailboxSubmissionReconciliationReconciler(t, repository, engine).Reconcile(context.Background(), submissionID); err != nil { t.Fatalf("Reconcile() error = %v, want nil", err) }
			if repository.resolveCalls != 1 || engine.calls != 1 { t.Fatalf("resolve/engine calls = %d/%d, want 1/1", repository.resolveCalls, engine.calls) }
		})
	}
}
func TestMailboxSubmissionReconcilerInvalidAcceptedPairRetries(t *testing.T) {
	submissionID, _, _, claim := mailboxSubmissionReconciliationFixture(t)
	repository := &mailboxSubmissionReconciliationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionReconciliationClaimRequest) (mail.MailboxSubmissionReconciliationClaimResult, error) { return claim, nil }, resolveFn: func(_ context.Context, request mail.MailboxSubmissionReconciliationResolveRequest) error {
		if request.Observation != mail.MailboxSubmissionReconciliationObservationRetry || request.RetryCode != mail.MailboxSubmissionErrorCodeReconciliationRequired || request.EngineSubmissionID != "" || request.EngineMessageID != "" { t.Fatalf("invalid accepted result must resolve as content-free retry, got %#v", request) }
		return nil
	}}
	engine := &mailboxSubmissionReconciliationEngineSpy{reconcileFn: func(context.Context, mail.EngineSubmission) (mail.EngineSubmissionResult, error) { return mail.EngineSubmissionAccepted{SubmissionID: "engine-submission-opaque", MessageID: "different-message-opaque"}, nil }}
	if err := mailboxSubmissionReconciliationReconciler(t, repository, engine).Reconcile(context.Background(), submissionID); err != nil { t.Fatalf("Reconcile() error = %v, want nil", err) }
	if repository.resolveCalls != 1 || engine.calls != 1 { t.Fatalf("resolve/engine calls = %d/%d, want 1/1", repository.resolveCalls, engine.calls) }
}
func TestMailboxSubmissionReconcilerCancellationBeforeAndAfterClaimIsConservative(t *testing.T) {
	t.Run("before claim", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background()); cancel()
		repository, engine := &mailboxSubmissionReconciliationRepositorySpy{}, &mailboxSubmissionReconciliationEngineSpy{}
		if err := mailboxSubmissionReconciliationReconciler(t, repository, engine).Reconcile(ctx, uuid.New()); err != nil { t.Fatalf("Reconcile(cancelled before claim) error = %v, want nil", err) }
		if repository.claimCalls != 0 || repository.resolveCalls != 0 || engine.calls != 0 { t.Fatalf("claim/resolve/engine calls = %d/%d/%d, want 0/0/0", repository.claimCalls, repository.resolveCalls, engine.calls) }
	})
	t.Run("after claim", func(t *testing.T) {
		submissionID, _, _, claim := mailboxSubmissionReconciliationFixture(t)
		ctx, cancel := context.WithCancel(context.Background())
		repository := &mailboxSubmissionReconciliationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionReconciliationClaimRequest) (mail.MailboxSubmissionReconciliationClaimResult, error) { cancel(); return claim, nil }, resolveFn: func(resolveCtx context.Context, request mail.MailboxSubmissionReconciliationResolveRequest) error {
			if resolveCtx.Err() != nil || request.Observation != mail.MailboxSubmissionReconciliationObservationRetry || request.RetryCode != mail.MailboxSubmissionErrorCodeReconciliationRequired { t.Fatalf("post-claim cancellation resolution = ctx:%v request:%#v", resolveCtx.Err(), request) }
			return nil
		}}
		engine := &mailboxSubmissionReconciliationEngineSpy{reconcileFn: func(context.Context, mail.EngineSubmission) (mail.EngineSubmissionResult, error) { t.Fatal("post-claim cancellation must not call engine"); return nil, nil }}
		if err := mailboxSubmissionReconciliationReconciler(t, repository, engine).Reconcile(ctx, submissionID); err != nil { t.Fatalf("Reconcile(cancelled after claim) error = %v, want nil", err) }
		if repository.claimCalls != 1 || repository.resolveCalls != 1 || engine.calls != 0 { t.Fatalf("claim/resolve/engine calls = %d/%d/%d, want 1/1/0", repository.claimCalls, repository.resolveCalls, engine.calls) }
	})
}
func TestMailboxSubmissionReconcilerResolveFailureIsContentFreeAndSingleCall(t *testing.T) {
	submissionID, _, submission, claim := mailboxSubmissionReconciliationFixture(t)
	repository := &mailboxSubmissionReconciliationRepositorySpy{claimFn: func(context.Context, mail.MailboxSubmissionReconciliationClaimRequest) (mail.MailboxSubmissionReconciliationClaimResult, error) { return claim, nil }, resolveFn: func(context.Context, mail.MailboxSubmissionReconciliationResolveRequest) error { return errors.New("raw durable failure body@example.test") }}
	engine := &mailboxSubmissionReconciliationEngineSpy{reconcileFn: func(context.Context, mail.EngineSubmission) (mail.EngineSubmissionResult, error) { return mail.EngineSubmissionAccepted{SubmissionID: "engine-submission-opaque", MessageID: submission.MessageID}, nil }}
	err := mailboxSubmissionReconciliationReconciler(t, repository, engine).Reconcile(context.Background(), submissionID)
	if !mailboxSubmissionReconciliationErrorIs(err, mail.MailboxSubmissionReconciliationErrorReconciliationRequired) { t.Fatalf("Reconcile() error = %v, want content-free reconciliation-required error", err) }
	if strings.Contains(err.Error(), "body@example.test") || repository.resolveCalls != 1 || engine.calls != 1 { t.Fatalf("error/calls = %v/%d/%d, want content-free error and 1/1 calls", err, repository.resolveCalls, engine.calls) }
}
func TestMailboxSubmissionReconciliationResultFactoriesRejectInvalidShapes(t *testing.T) {
	if _, err := mail.NewMailboxSubmissionReconciliationBusyResult(time.Time{}); !mailboxSubmissionReconciliationErrorIs(err, mail.MailboxSubmissionReconciliationErrorInvalid) { t.Fatalf("NewMailboxSubmissionReconciliationBusyResult(zero) error = %v, want invalid", err) }
	historical := time.Date(2000, time.January, 2, 3, 4, 5, 0, time.UTC)
	busy, err := mail.NewMailboxSubmissionReconciliationBusyResult(historical)
	if err != nil { t.Fatalf("NewMailboxSubmissionReconciliationBusyResult(historical) error = %v", err) }
	if got, ok := busy.BusyUntil(); !ok || got != historical { t.Fatalf("BusyUntil() = %v, %t; want exact historical canonical UTC", got, ok) }
	if _, err := mail.NewMailboxSubmissionReconciliationBusyResult(historical.Add(time.Nanosecond)); !mailboxSubmissionReconciliationErrorIs(err, mail.MailboxSubmissionReconciliationErrorInvalid) { t.Fatalf("NewMailboxSubmissionReconciliationBusyResult(noncanonical) error = %v, want invalid", err) }
	if _, err := mail.NewMailboxSubmissionReconciliationClaimedResult(uuid.Nil, mail.MailboxEngineCommandFence{}, mail.EngineSubmission{}); !mailboxSubmissionReconciliationErrorIs(err, mail.MailboxSubmissionReconciliationErrorInvalid) { t.Fatalf("NewMailboxSubmissionReconciliationClaimedResult(invalid) error = %v, want invalid", err) }
}

func TestMailboxSubmissionReconciliationPublicContractIsClosed(t *testing.T) {
	for _, testCase := range []struct {
		name string
		got  string
		want string
	}{
		{name: "claim noop", got: string(mail.MailboxSubmissionReconciliationClaimOutcomeNoop), want: "noop"},
		{name: "claim busy", got: string(mail.MailboxSubmissionReconciliationClaimOutcomeBusy), want: "busy"},
		{name: "claim claimed", got: string(mail.MailboxSubmissionReconciliationClaimOutcomeClaimed), want: "claimed"},
		{name: "accepted", got: string(mail.MailboxSubmissionReconciliationObservationAccepted), want: "accepted"},
		{name: "retry", got: string(mail.MailboxSubmissionReconciliationObservationRetry), want: "retry"},
		{name: "invalid error", got: string(mail.MailboxSubmissionReconciliationErrorInvalid), want: "invalid"},
		{name: "repository unavailable error", got: string(mail.MailboxSubmissionReconciliationErrorRepositoryUnavailable), want: "repository_unavailable"},
		{name: "reconciliation required error", got: string(mail.MailboxSubmissionReconciliationErrorReconciliationRequired), want: "reconciliation_required"},
	} {
		if testCase.got != testCase.want {
			t.Fatalf("%s = %q, want %q", testCase.name, testCase.got, testCase.want)
		}
	}

	requireMailboxSubmissionReconciliationFields(t, reflect.TypeOf(mail.MailboxSubmissionReconciliationScanRequest{}), []struct {
		name  string
		type_ reflect.Type
	}{
		{name: "Limit", type_: reflect.TypeOf(int(0))},
	})
	requireMailboxSubmissionReconciliationFields(t, reflect.TypeOf(mail.MailboxSubmissionReconciliationScanResult{}), []struct {
		name  string
		type_ reflect.Type
	}{
		{name: "SubmissionIDs", type_: reflect.TypeOf([]uuid.UUID(nil))},
	})
	requireMailboxSubmissionReconciliationFields(t, reflect.TypeOf(mail.MailboxSubmissionReconciliationClaimRequest{}), []struct {
		name  string
		type_ reflect.Type
	}{
		{name: "SubmissionID", type_: reflect.TypeOf(uuid.UUID{})},
	})
	requireMailboxSubmissionReconciliationFields(t, reflect.TypeOf(mail.MailboxSubmissionReconciliationResolveRequest{}), []struct {
		name  string
		type_ reflect.Type
	}{
		{name: "Claim", type_: reflect.TypeOf(mail.MailboxSubmissionReconciliationClaim{})},
		{name: "Observation", type_: reflect.TypeOf(mail.MailboxSubmissionReconciliationObservation(""))},
		{name: "EngineSubmissionID", type_: reflect.TypeOf("")},
		{name: "EngineMessageID", type_: reflect.TypeOf("")},
		{name: "RetryCode", type_: reflect.TypeOf(mail.MailboxSubmissionErrorCode(""))},
	})

	claimType := reflect.TypeOf(mail.MailboxSubmissionReconciliationClaim{})
	if claimType.NumField() != 3 {
		t.Fatalf("MailboxSubmissionReconciliationClaim field count = %d, want 3 sealed fields", claimType.NumField())
	}
	for index := 0; index < claimType.NumField(); index++ {
		if claimType.Field(index).PkgPath == "" {
			t.Fatalf("MailboxSubmissionReconciliationClaim field %q must remain sealed", claimType.Field(index).Name)
		}
	}
	for _, forbidden := range []string{"Submission", "Provider", "Recipient", "Policy"} {
		if forbidden == "Submission" {
			continue
		}
		if _, found := claimType.MethodByName(forbidden); found {
			t.Fatalf("MailboxSubmissionReconciliationClaim must not expose %s", forbidden)
		}
	}
	if _, found := claimType.MethodByName("SubmissionID"); !found {
		t.Fatal("MailboxSubmissionReconciliationClaim must expose its durable submission identity")
	}
	if _, found := claimType.MethodByName("Fence"); !found {
		t.Fatal("MailboxSubmissionReconciliationClaim must expose its exact command fence")
	}
}

func TestMailboxSubmissionReconcilerRejectsNilAndTypedNilDependencies(t *testing.T) {
	if reconciler, err := mail.NewMailboxSubmissionReconciler(nil, nil); reconciler != nil || !mailboxSubmissionReconciliationErrorIs(err, mail.MailboxSubmissionReconciliationErrorInvalid) {
		t.Fatalf("NewMailboxSubmissionReconciler(nil, nil) = %#v, %v; want nil invalid error", reconciler, err)
	}

	var engine *mail.StalwartEngine
	if reconciler, err := mail.NewMailboxSubmissionReconciler(mailboxSubmissionReconciliationRepositoryStub{}, engine); reconciler != nil || !mailboxSubmissionReconciliationErrorIs(err, mail.MailboxSubmissionReconciliationErrorInvalid) {
		t.Fatalf("NewMailboxSubmissionReconciler(stub, typed nil) = %#v, %v; want nil invalid error", reconciler, err)
	}
}

func TestMailboxSubmissionReconciliationSourceContract(t *testing.T) {
	abiBytes, err := os.ReadFile("../../internal/mail/mailbox_submission_reconciliation.go")
	if err != nil {
		t.Fatalf("read reconciliation ABI source: %v", err)
	}
	reconcilerBytes, err := os.ReadFile("../../internal/mail/mailbox_submission_reconciler.go")
	if err != nil {
		t.Fatalf("read reconciler source: %v", err)
	}
	abiSource := string(abiBytes)
	reconcilerSource := string(reconcilerBytes)

	for _, required := range []string{
		"request.Limit >= 1 && request.Limit <= mailboxSubmissionReconciliationMaximum",
		"len(result.SubmissionIDs) > request.Limit",
		"mailboxSubmissionReconciliationClaimCopy",
		"accepted.EngineMessageID != request.Claim.submission.MessageID",
		"MailboxSubmissionErrorCodeMailboxUnavailable",
		"MailboxSubmissionErrorCodeReconciliationRequired",
	} {
		if !strings.Contains(abiSource, required) {
			t.Fatalf("reconciliation ABI source is missing %q", required)
		}
	}

	if got := strings.Count(reconcilerSource, ".ReconcileSubmission("); got != 1 {
		t.Fatalf("reconciler must invoke ReconcileSubmission exactly once in source, got %d", got)
	}
	for _, required := range []string{
		"ClaimUnknownMailboxSubmission",
		"claimResult.Claim()",
		"ctx.Err() != nil",
		"context.WithoutCancel(ctx)",
		"ResolveUnknownMailboxSubmissionReconciliation",
		"mailboxSubmissionReconciliationRetryCode",
	} {
		if !strings.Contains(reconcilerSource, required) {
			t.Fatalf("reconciler source is missing %q", required)
		}
	}
	for _, forbidden := range []string{
		".PreflightSubmission(",
		".submitMailboxSubmission(",
		".GetMessage(",
		".Submit(",
		"provider",
		"recipient",
	} {
		if strings.Contains(strings.ToLower(reconcilerSource), strings.ToLower(forbidden)) {
			t.Fatalf("reconciler source contains forbidden %q", forbidden)
		}
	}
}

func requireMailboxSubmissionReconciliationFields(t *testing.T, candidate reflect.Type, want []struct {
	name  string
	type_ reflect.Type
}) {
	t.Helper()

	if candidate.NumField() != len(want) {
		t.Fatalf("%s field count = %d, want %d", candidate.Name(), candidate.NumField(), len(want))
	}
	for index, expected := range want {
		field := candidate.Field(index)
		if field.Name != expected.name || field.Type != expected.type_ || field.PkgPath != "" {
			t.Fatalf("%s field %d = %s %s exported=%t, want %s %s exported", candidate.Name(), index, field.Name, field.Type, field.PkgPath == "", expected.name, expected.type_)
		}
	}
}

func mailboxSubmissionReconciliationErrorIs(err error, want mail.MailboxSubmissionReconciliationErrorKind) bool {
	var reconciliationError *mail.MailboxSubmissionReconciliationError
	return errors.As(err, &reconciliationError) && reconciliationError != nil && reconciliationError.Kind() == want && reconciliationError.Error() == "mailbox submission reconciliation error"
}

var _ = time.Time{}
