package mail_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/mail"
	`github.com/valerii/hub/backend/internal/mail/stalwart`
)

func TestMailboxPasswordEnrollmentService_EnrollsAndClearsOwnedPassword(t *testing.T) {
	claim := mailboxPasswordEnrollmentClaim(t, time.Now().UTC().Add(30*time.Second).Truncate(time.Second))
	repository := &mailboxPasswordEnrollmentRepositorySpy{claimResult: claim}
	passwords := &mailboxPasswordEnrollmentPasswordSpy{}
	service := mailboxPasswordEnrollmentService(t, repository, passwords)
	callerPassword := []byte("correct horse battery staple")
	before := append([]byte(nil), callerPassword...)

	result, err := service.EnrollPassword(context.Background(), uuid.New(), callerPassword)
	if err != nil || result != mail.MailboxPasswordEnrollmentResultEnrolled {
		t.Fatalf("EnrollPassword() = %q, %v; want enrolled, nil", result, err)
	}
	if repository.claimCalls != 1 || repository.completeCalls != 1 || repository.releaseCalls != 0 || passwords.calls != 1 {
		t.Fatalf("claim/complete/release/provider = %d/%d/%d/%d, want 1/1/0/1", repository.claimCalls, repository.completeCalls, repository.releaseCalls, passwords.calls)
	}
	if !bytes.Equal(callerPassword, before) {
		t.Fatalf("EnrollPassword() changed caller password buffer")
	}
	if !bytes.Equal(passwords.passwordAtCall, before) || !allZero(passwords.retainedPassword) {
		t.Fatalf("provider password at-call/after-return = %q/%v; want preserved/cleared", passwords.passwordAtCall, passwords.retainedPassword)
	}
	if passwords.accountID != "registry-account-1" {
		t.Fatalf("provider account = %q, want sealed account", passwords.accountID)
	}
}

func TestMailboxPasswordEnrollmentService_NoopAndBusyDoNotCallProvider(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		claim  mail.MailboxPasswordEnrollmentClaimResult
		result mail.MailboxPasswordEnrollmentResult
	}{
		{name: "already enrolled", claim: mail.NewMailboxPasswordEnrollmentNoopResult(), result: mail.MailboxPasswordEnrollmentResultAlreadyEnrolled},
		{name: "busy", claim: mail.NewMailboxPasswordEnrollmentBusyResult(), result: mail.MailboxPasswordEnrollmentResultBusy},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &mailboxPasswordEnrollmentRepositorySpy{claimResult: testCase.claim}
			passwords := &mailboxPasswordEnrollmentPasswordSpy{}
			service := mailboxPasswordEnrollmentService(t, repository, passwords)

			result, err := service.EnrollPassword(context.Background(), uuid.New(), []byte("correct horse battery staple"))
			if err != nil || result != testCase.result {
				t.Fatalf("EnrollPassword() = %q, %v; want %q, nil", result, err, testCase.result)
			}
			if repository.completeCalls != 0 || repository.releaseCalls != 0 || passwords.calls != 0 {
				t.Fatalf("complete/release/provider = %d/%d/%d, want 0/0/0", repository.completeCalls, repository.releaseCalls, passwords.calls)
			}
		})
	}
}

func TestMailboxPasswordEnrollmentService_RejectsInvalidInputBeforeClaim(t *testing.T) {
	repository := &mailboxPasswordEnrollmentRepositorySpy{}
	passwords := &mailboxPasswordEnrollmentPasswordSpy{}
	service := mailboxPasswordEnrollmentService(t, repository, passwords)

	for _, testCase := range []struct {
		name     string
		actorID  uuid.UUID
		password []byte
	}{
		{name: "zero actor", password: []byte("correct horse battery staple")},
		{name: "short password", actorID: uuid.New(), password: []byte("too-short")},
		{name: "control password", actorID: uuid.New(), password: []byte("twelve\x00password")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			result, err := service.EnrollPassword(context.Background(), testCase.actorID, testCase.password)
			if result != "" || !mailboxPasswordEnrollmentErrorIs(err, mail.MailboxPasswordEnrollmentErrorInvalid) {
				t.Fatalf("EnrollPassword() = %q, %v; want invalid error", result, err)
			}
		})
	}
	if repository.claimCalls != 0 || passwords.calls != 0 {
		t.Fatalf("claim/provider = %d/%d, want 0/0", repository.claimCalls, passwords.calls)
	}
}

func TestMailboxPasswordEnrollmentService_CancellationBeforeProviderReleasesDetached(t *testing.T) {
	claim := mailboxPasswordEnrollmentClaim(t, time.Now().UTC().Add(30*time.Second).Truncate(time.Second))
	ctx, cancel := context.WithCancel(context.Background())
	repository := &mailboxPasswordEnrollmentRepositorySpy{
		claimResult: claim,
		claimFn: func(context.Context, uuid.UUID, uuid.UUID) (mail.MailboxPasswordEnrollmentClaimResult, error) {
			cancel()
			return claim, nil
		},
		releaseFn: func(releaseCtx context.Context, _ mail.MailboxPasswordEnrollmentClaim, code mail.MailboxPasswordEnrollmentReleaseCode) error {
			if releaseCtx.Err() != nil || code != mail.MailboxPasswordEnrollmentReleaseCodeTransient {
				t.Fatalf("release context/code = %v/%q, want live/transient", releaseCtx.Err(), code)
			}
			if _, ok := releaseCtx.Deadline(); !ok {
				t.Fatal("release context must be bounded")
			}
			return nil
		},
	}
	passwords := &mailboxPasswordEnrollmentPasswordSpy{}
	service := mailboxPasswordEnrollmentService(t, repository, passwords)

	result, err := service.EnrollPassword(ctx, uuid.New(), []byte("correct horse battery staple"))
	if result != "" || !mailboxPasswordEnrollmentErrorIs(err, mail.MailboxPasswordEnrollmentErrorRetryable) {
		t.Fatalf("EnrollPassword() = %q, %v; want retryable error", result, err)
	}
	if passwords.calls != 0 || repository.releaseCalls != 1 {
		t.Fatalf("provider/release = %d/%d, want 0/1", passwords.calls, repository.releaseCalls)
	}
}

func TestMailboxPasswordEnrollmentService_RespectsLeaseDeadlineWithoutProviderIO(t *testing.T) {
	claim := mailboxPasswordEnrollmentClaim(t, time.Now().UTC().Add(500*time.Millisecond).Truncate(time.Second))
	repository := &mailboxPasswordEnrollmentRepositorySpy{claimResult: claim}
	passwords := &mailboxPasswordEnrollmentPasswordSpy{}
	service := mailboxPasswordEnrollmentService(t, repository, passwords)

	result, err := service.EnrollPassword(context.Background(), uuid.New(), []byte("correct horse battery staple"))
	if result != "" || !mailboxPasswordEnrollmentErrorIs(err, mail.MailboxPasswordEnrollmentErrorRetryable) {
		t.Fatalf("EnrollPassword() = %q, %v; want retryable error", result, err)
	}
	if passwords.calls != 0 || repository.releaseCalls != 1 || repository.releaseCode != mail.MailboxPasswordEnrollmentReleaseCodeTransient {
		t.Fatalf("provider/release/code = %d/%d/%q, want 0/1/transient", passwords.calls, repository.releaseCalls, repository.releaseCode)
	}
}

func TestMailboxPasswordEnrollmentService_MapsProviderFailuresToClosedReleaseCodes(t *testing.T) {
	for _, testCase := range []struct {
		name string
		err  error
		code mail.MailboxPasswordEnrollmentReleaseCode
	}{
		{name: "rejected", err: mail.ErrEngineRejectedPolicy, code: mail.MailboxPasswordEnrollmentReleaseCodeRejected},
		{name: "transient", err: mail.ErrEngineUnavailable, code: mail.MailboxPasswordEnrollmentReleaseCodeTransient},
		{name: "indeterminate", err: mail.ErrEngineIndeterminate, code: mail.MailboxPasswordEnrollmentReleaseCodeIndeterminate},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			claim := mailboxPasswordEnrollmentClaim(t, time.Now().UTC().Add(30*time.Second).Truncate(time.Second))
			repository := &mailboxPasswordEnrollmentRepositorySpy{claimResult: claim}
			passwords := &mailboxPasswordEnrollmentPasswordSpy{err: testCase.err}
			service := mailboxPasswordEnrollmentService(t, repository, passwords)

			result, err := service.EnrollPassword(context.Background(), uuid.New(), []byte("correct horse battery staple"))
			if result != "" || !mailboxPasswordEnrollmentErrorIs(err, mail.MailboxPasswordEnrollmentErrorRetryable) {
				t.Fatalf("EnrollPassword() = %q, %v; want retryable error", result, err)
			}
			if repository.completeCalls != 0 || repository.releaseCalls != 1 || repository.releaseCode != testCase.code {
				t.Fatalf("complete/release/code = %d/%d/%q, want 0/1/%q", repository.completeCalls, repository.releaseCalls, repository.releaseCode, testCase.code)
			}
			if strings.Contains(err.Error(), "provider") || strings.Contains(err.Error(), "registry-account-1") {
				t.Fatalf("EnrollPassword() leaked provider/account detail: %q", err)
			}
		})
	}
}

func TestMailboxPasswordEnrollmentService_MapsStalwartProviderFailuresToClosedReleaseCodes(t *testing.T) {
	for _, testCase := range []struct {
		kind stalwart.ErrorKind
		code mail.MailboxPasswordEnrollmentReleaseCode
	}{
		{stalwart.ErrorInvalidConfiguration, mail.MailboxPasswordEnrollmentReleaseCodeRejected},
		{stalwart.ErrorInvalidRequest, mail.MailboxPasswordEnrollmentReleaseCodeRejected},
		{stalwart.ErrorUnauthorized, mail.MailboxPasswordEnrollmentReleaseCodeRejected},
		{stalwart.ErrorNotFound, mail.MailboxPasswordEnrollmentReleaseCodeRejected},
		{stalwart.ErrorPayloadTooLarge, mail.MailboxPasswordEnrollmentReleaseCodeRejected},
		{stalwart.ErrorUnsupported, mail.MailboxPasswordEnrollmentReleaseCodeRejected},
		{stalwart.ErrorRestartRequired, mail.MailboxPasswordEnrollmentReleaseCodeTransient},
		{stalwart.ErrorRateLimited, mail.MailboxPasswordEnrollmentReleaseCodeTransient},
		{stalwart.ErrorStateMismatch, mail.MailboxPasswordEnrollmentReleaseCodeIndeterminate},
		{stalwart.ErrorProtocol, mail.MailboxPasswordEnrollmentReleaseCodeIndeterminate},
		{stalwart.ErrorUnavailable, mail.MailboxPasswordEnrollmentReleaseCodeIndeterminate},
		{stalwart.ErrorIndeterminate, mail.MailboxPasswordEnrollmentReleaseCodeIndeterminate},
		{stalwart.ErrorCancelled, mail.MailboxPasswordEnrollmentReleaseCodeIndeterminate},
	} {
		t.Run(string(testCase.kind), func(t *testing.T) {
			claim := mailboxPasswordEnrollmentClaim(t, time.Now().UTC().Add(30*time.Second).Truncate(time.Second))
			repository := &mailboxPasswordEnrollmentRepositorySpy{claimResult: claim}
			passwords := &mailboxPasswordEnrollmentPasswordSpy{err: &stalwart.Error{Kind: testCase.kind}}
			service := mailboxPasswordEnrollmentService(t, repository, passwords)
			result, err := service.EnrollPassword(context.Background(), uuid.New(), []byte(`correct horse battery staple`))
			if result != `` || !mailboxPasswordEnrollmentErrorIs(err, mail.MailboxPasswordEnrollmentErrorRetryable) || repository.releaseCalls != 1 || repository.releaseCode != testCase.code {
				t.Fatalf(`EnrollPassword() = %q, %v, release/code = %d/%q`, result, err, repository.releaseCalls, repository.releaseCode)
			}
			if strings.Contains(err.Error(), string(testCase.kind)) || strings.Contains(err.Error(), `registry-account-1`) || strings.Contains(err.Error(), `correct horse battery staple`) {
				t.Fatalf(`EnrollPassword() leaked provider, account, or password detail: %q`, err)
			}
		})
	}
	claim := mailboxPasswordEnrollmentClaim(t, time.Now().UTC().Add(30*time.Second).Truncate(time.Second))
	repository := &mailboxPasswordEnrollmentRepositorySpy{claimResult: claim}
	passwords := &mailboxPasswordEnrollmentPasswordSpy{err: errors.New(`future provider detail registry-account-1`)}
	service := mailboxPasswordEnrollmentService(t, repository, passwords)
	_, err := service.EnrollPassword(context.Background(), uuid.New(), []byte(`correct horse battery staple`))
	if !mailboxPasswordEnrollmentErrorIs(err, mail.MailboxPasswordEnrollmentErrorRetryable) || repository.releaseCode != mail.MailboxPasswordEnrollmentReleaseCodeIndeterminate || strings.Contains(err.Error(), `future provider detail`) {
		t.Fatalf(`unknown provider error = %v, release=%q`, err, repository.releaseCode)
	}
}

func TestMailboxPasswordEnrollmentService_ContainsCompletionReleaseAndPortPanics(t *testing.T) {
	t.Run("completion failure is not released", func(t *testing.T) {
		claim := mailboxPasswordEnrollmentClaim(t, time.Now().UTC().Add(30*time.Second).Truncate(time.Second))
		repository := &mailboxPasswordEnrollmentRepositorySpy{claimResult: claim, completeErr: errors.New("complete registry-account-1")}
		service := mailboxPasswordEnrollmentService(t, repository, &mailboxPasswordEnrollmentPasswordSpy{})

		result, err := service.EnrollPassword(context.Background(), uuid.New(), []byte("correct horse battery staple"))
		if result != "" || !mailboxPasswordEnrollmentErrorIs(err, mail.MailboxPasswordEnrollmentErrorRetryable) || repository.releaseCalls != 0 {
			t.Fatalf("EnrollPassword() = %q, %v, release=%d; want retryable/no release", result, err, repository.releaseCalls)
		}
	})

	t.Run("release failure stays content free", func(t *testing.T) {
		claim := mailboxPasswordEnrollmentClaim(t, time.Now().UTC().Add(30*time.Second).Truncate(time.Second))
		repository := &mailboxPasswordEnrollmentRepositorySpy{claimResult: claim, releaseErr: errors.New("release registry-account-1")}
		service := mailboxPasswordEnrollmentService(t, repository, &mailboxPasswordEnrollmentPasswordSpy{err: mail.ErrEngineUnavailable})

		result, err := service.EnrollPassword(context.Background(), uuid.New(), []byte("correct horse battery staple"))
		if result != "" || !mailboxPasswordEnrollmentErrorIs(err, mail.MailboxPasswordEnrollmentErrorRepositoryUnavailable) || strings.Contains(err.Error(), "registry-account-1") {
			t.Fatalf("EnrollPassword() = %q, %v; want content-free repository error", result, err)
		}
	})

	t.Run("provider panic is released", func(t *testing.T) {
		claim := mailboxPasswordEnrollmentClaim(t, time.Now().UTC().Add(30*time.Second).Truncate(time.Second))
		repository := &mailboxPasswordEnrollmentRepositorySpy{claimResult: claim}
		passwords := &mailboxPasswordEnrollmentPasswordSpy{panicValue: "registry-account-1"}
		service := mailboxPasswordEnrollmentService(t, repository, passwords)

		result, err := service.EnrollPassword(context.Background(), uuid.New(), []byte("correct horse battery staple"))
		if result != "" || !mailboxPasswordEnrollmentErrorIs(err, mail.MailboxPasswordEnrollmentErrorRetryable) || repository.releaseCode != mail.MailboxPasswordEnrollmentReleaseCodeIndeterminate {
			t.Fatalf("EnrollPassword() = %q, %v, code=%q; want retryable/indeterminate", result, err, repository.releaseCode)
		}
	})
}

func TestMailboxPasswordEnrollmentService_ClaimFactoryAndErrorsKeepIdentifiersSealed(t *testing.T) {
	if _, err := mail.NewMailboxPasswordEnrollmentClaimedResult(mail.MailboxPasswordEnrollmentClaimedInput{}); !mailboxPasswordEnrollmentErrorIs(err, mail.MailboxPasswordEnrollmentErrorInvalid) {
		t.Fatalf("zero claimed input error = %v; want invalid", err)
	}
	claim := mailboxPasswordEnrollmentClaim(t, time.Now().UTC().Add(30*time.Second).Truncate(time.Second))
	if _, ok := claim.Claim(); !ok {
		t.Fatal("claimed result must yield sealed claim")
	}
	for _, value := range []string{"registry-account-1", uuid.NewString(), "correct horse battery staple"} {
		err := mailboxPasswordEnrollmentError(mail.MailboxPasswordEnrollmentErrorRetryable)
		if strings.Contains(err.Error(), value) {
			t.Fatalf("closed error leaked %q", value)
		}
	}
}

func mailboxPasswordEnrollmentService(t *testing.T, repository mail.MailboxPasswordEnrollmentRepository, passwords mail.RegistryPasswordPort) *mail.MailboxPasswordEnrollmentService {
	t.Helper()
	service, err := mail.NewMailboxPasswordEnrollmentService(repository, passwords)
	if err != nil {
		t.Fatalf("NewMailboxPasswordEnrollmentService() error = %v", err)
	}
	return service
}

func mailboxPasswordEnrollmentClaim(t *testing.T, leaseUntil time.Time) mail.MailboxPasswordEnrollmentClaimResult {
	t.Helper()
	result, err := mail.NewMailboxPasswordEnrollmentClaimedResult(mail.MailboxPasswordEnrollmentClaimedInput{
		AccountID:           uuid.New(),
		OrganizationID:      uuid.New(),
		EngineAccountID:     "registry-account-1",
		DeployedGeneration:  1,
		DeployedHashVersion: 1,
		DeployedHash:        [sha256.Size]byte{1},
		Generation:          1,
		LeaseToken:          uuid.New(),
		LeaseUntil:          leaseUntil,
	})
	if err != nil {
		t.Fatalf("NewMailboxPasswordEnrollmentClaimedResult() error = %v", err)
	}
	return result
}

type mailboxPasswordEnrollmentRepositorySpy struct {
	claimResult  mail.MailboxPasswordEnrollmentClaimResult
	claimErr     error
	claimFn      func(context.Context, uuid.UUID, uuid.UUID) (mail.MailboxPasswordEnrollmentClaimResult, error)
	completeErr  error
	releaseErr   error
	releaseFn    func(context.Context, mail.MailboxPasswordEnrollmentClaim, mail.MailboxPasswordEnrollmentReleaseCode) error
	claimCalls   int
	completeCalls int
	releaseCalls int
	releaseCode  mail.MailboxPasswordEnrollmentReleaseCode
}

func (spy *mailboxPasswordEnrollmentRepositorySpy) Claim(ctx context.Context, actorID, leaseToken uuid.UUID) (mail.MailboxPasswordEnrollmentClaimResult, error) {
	spy.claimCalls++
	if actorID == uuid.Nil || leaseToken == uuid.Nil {
		return mail.MailboxPasswordEnrollmentClaimResult{}, errors.New("unexpected zero claim identity")
	}
	if spy.claimFn != nil {
		return spy.claimFn(ctx, actorID, leaseToken)
	}
	return spy.claimResult, spy.claimErr
}

func (spy *mailboxPasswordEnrollmentRepositorySpy) Complete(ctx context.Context, claim mail.MailboxPasswordEnrollmentClaim, observedAt time.Time) error {
	spy.completeCalls++
	if ctx.Err() != nil || observedAt.IsZero() || observedAt.Location() != time.UTC || observedAt.Nanosecond() != 0 {
		return errors.New("unexpected complete context or observation")
	}
	return spy.completeErr
}

func (spy *mailboxPasswordEnrollmentRepositorySpy) Release(ctx context.Context, claim mail.MailboxPasswordEnrollmentClaim, code mail.MailboxPasswordEnrollmentReleaseCode) error {
	spy.releaseCalls++
	spy.releaseCode = code
	if spy.releaseFn != nil {
		return spy.releaseFn(ctx, claim, code)
	}
	return spy.releaseErr
}

type mailboxPasswordEnrollmentPasswordSpy struct {
	mu               sync.Mutex
	calls            int
	accountID        string
	passwordAtCall   []byte
	retainedPassword []byte
	err              error
	panicValue       any
}

func (spy *mailboxPasswordEnrollmentPasswordSpy) SetUserPassword(_ context.Context, accountID string, password []byte) error {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	spy.calls++
	spy.accountID = accountID
	spy.passwordAtCall = append([]byte(nil), password...)
	spy.retainedPassword = password
	if spy.panicValue != nil {
		panic(spy.panicValue)
	}
	return spy.err
}

func mailboxPasswordEnrollmentError(kind mail.MailboxPasswordEnrollmentErrorKind) error {
	if kind == mail.MailboxPasswordEnrollmentErrorRetryable {
		return errors.New("mailbox password enrollment error")
	}
	return errors.New("mailbox password enrollment error")
}

func mailboxPasswordEnrollmentErrorIs(err error, want mail.MailboxPasswordEnrollmentErrorKind) bool {
	var target *mail.MailboxPasswordEnrollmentError
	return errors.As(err, &target) && target != nil && target.Kind() == want && target.Error() == "mailbox password enrollment error"
}

func allZero(value []byte) bool {
	for _, item := range value {
		if item != 0 {
			return false
		}
	}
	return true
}
