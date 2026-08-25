package mail_test

import (
	"context"
	"errors"
	"os"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

var (
	submissionFinalizationWorkerIDOne = uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	submissionFinalizationWorkerIDTwo = uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")
)

func TestMailboxSubmissionFinalizationWorkerAdaptsAcceptedScanAndResult(t *testing.T) {
	repository := &submissionFinalizationWorkerRepositorySpy{
		scan: func(context.Context, mail.MailboxSubmissionFinalizationScanRequest) (mail.MailboxSubmissionFinalizationScanResult, error) {
			return mail.MailboxSubmissionFinalizationScanResult{SubmissionIDs: []uuid.UUID{submissionFinalizationWorkerIDOne, submissionFinalizationWorkerIDTwo}}, nil
		},
	}
	executor := &submissionFinalizationWorkerExecutorSpy{
		finalize: func(_ context.Context, submissionID uuid.UUID) error {
			if submissionID == submissionFinalizationWorkerIDTwo {
				return errors.New("finalization failed")
			}
			return nil
		},
	}
	timer := &submissionFinalizationWorkerTimerSpy{}
	worker, err := mail.NewMailboxSubmissionFinalizationWorker(repository, executor, timer, mail.MailboxSubmissionFinalizationWorkerOptions{
		Enabled:         true,
		ScanInterval:    2 * time.Second,
		FinalizeTimeout: 3 * time.Second,
		BatchSize:       7,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, runErr := worker.RunOnce(context.Background())
	if result != (mail.MailboxSubmissionFinalizationWorkerRunResult{Due: 2, Finalized: 1, Failed: 1}) || !errors.Is(runErr, mail.ErrMailboxSubmissionWorkerPartialFailure) {
		t.Fatalf("RunOnce() = %#v, %v", result, runErr)
	}
	if requests := repository.scanRequests(); !reflect.DeepEqual(requests, []mail.MailboxSubmissionFinalizationScanRequest{{Limit: 7}}) {
		t.Fatalf("scan requests = %#v", requests)
	}
	if durations := timer.timeoutDurations(); !reflect.DeepEqual(durations, []time.Duration{3 * time.Second, 3 * time.Second}) {
		t.Fatalf("timeout durations = %v", durations)
	}
	if ids := executor.submissionIDs(); !reflect.DeepEqual(ids, []uuid.UUID{submissionFinalizationWorkerIDOne, submissionFinalizationWorkerIDTwo}) {
		t.Fatalf("finalized IDs = %v", ids)
	}
}

func TestMailboxSubmissionFinalizationWorkerDefaultsAndDisabledZero(t *testing.T) {
	defaults := mail.DefaultMailboxSubmissionFinalizationWorkerOptions()
	if !defaults.Enabled || defaults.ScanInterval != time.Minute || defaults.FinalizeTimeout != 30*time.Second || defaults.BatchSize != 20 {
		t.Fatalf("DefaultMailboxSubmissionFinalizationWorkerOptions() = %#v", defaults)
	}

	worker, err := mail.NewMailboxSubmissionFinalizationWorker(nil, nil, nil, mail.MailboxSubmissionFinalizationWorkerOptions{})
	if err != nil || worker == nil {
		t.Fatalf("disabled worker = %#v, %v", worker, err)
	}
	if result, runErr := worker.RunOnce(context.Background()); result != (mail.MailboxSubmissionFinalizationWorkerRunResult{}) || runErr != nil {
		t.Fatalf("disabled RunOnce() = %#v, %v", result, runErr)
	}
	if runErr := worker.Run(context.Background()); runErr != nil {
		t.Fatalf("disabled Run() = %v", runErr)
	}
}

func TestMailboxSubmissionFinalizationWorkerRejectsTypedNilDependencies(t *testing.T) {
	repository := &submissionFinalizationWorkerRepositorySpy{}
	executor := &submissionFinalizationWorkerExecutorSpy{}
	timer := &submissionFinalizationWorkerTimerSpy{}
	var typedNilRepository *submissionFinalizationWorkerRepositorySpy
	var typedNilExecutor *submissionFinalizationWorkerExecutorSpy
	var typedNilTimer *submissionFinalizationWorkerTimerSpy
	for _, testCase := range []struct {
		name       string
		repository mail.MailboxSubmissionFinalizationRepository
		executor   mail.MailboxSubmissionFinalizationExecutor
		timer      mail.MailboxSubmissionWorkerTimer
	}{
		{name: "repository", repository: typedNilRepository, executor: executor, timer: timer},
		{name: "executor", repository: repository, executor: typedNilExecutor, timer: timer},
		{name: "timer", repository: repository, executor: executor, timer: typedNilTimer},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got, constructorErr := mail.NewMailboxSubmissionFinalizationWorker(testCase.repository, testCase.executor, testCase.timer, mail.DefaultMailboxSubmissionFinalizationWorkerOptions())
			if got != nil || !errors.Is(constructorErr, mail.ErrMailboxSubmissionWorkerInvalid) {
				t.Fatalf("NewMailboxSubmissionFinalizationWorker() = %#v, %v", got, constructorErr)
			}
		})
	}
}

func TestMailboxSubmissionFinalizationWorkerSourceContract(t *testing.T) {
	source, err := os.ReadFile("../../internal/mail/mailbox_submission_finalization_worker.go")
	if err != nil {
		t.Fatalf("read finalization worker source: %v", err)
	}
	text := string(source)
	for _, required := range []string{
		"type MailboxSubmissionFinalizationExecutor interface",
		"func NewMailboxSubmissionFinalizationWorker",
		"func (worker *MailboxSubmissionFinalizationWorker) RunOnce",
		"ScanAcceptedMailboxSubmissions",
		"MailboxSubmissionWorkerOptions{",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("finalization worker source is missing %q", required)
		}
	}
	for _, forbidden := range []string{"MailEngine", "Provider", "Transport", "NATS", "Submit(", "Reconcile(", "Preflight", "GetMessage", "GetBlob", "http.", "nats.", "Claim", "Complete", "Release", "Storage"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("finalization worker source contains forbidden capability %q", forbidden)
		}
	}
	if calls := strings.Count(text, ".Finalize("); calls != 1 {
		t.Fatalf("Finalize call sites = %d, want one private dispatcher adapter call", calls)
	}
}

type submissionFinalizationWorkerRepositorySpy struct {
	mu       sync.Mutex
	scan     func(context.Context, mail.MailboxSubmissionFinalizationScanRequest) (mail.MailboxSubmissionFinalizationScanResult, error)
	requests []mail.MailboxSubmissionFinalizationScanRequest
}

var _ mail.MailboxSubmissionFinalizationRepository = (*submissionFinalizationWorkerRepositorySpy)(nil)

func (spy *submissionFinalizationWorkerRepositorySpy) ScanAcceptedMailboxSubmissions(ctx context.Context, request mail.MailboxSubmissionFinalizationScanRequest) (mail.MailboxSubmissionFinalizationScanResult, error) {
	spy.mu.Lock()
	spy.requests = append(spy.requests, request)
	callback := spy.scan
	spy.mu.Unlock()
	if callback == nil {
		return mail.MailboxSubmissionFinalizationScanResult{SubmissionIDs: []uuid.UUID{}}, nil
	}
	return callback(ctx, request)
}

func (*submissionFinalizationWorkerRepositorySpy) ClaimAcceptedMailboxSubmission(context.Context, mail.MailboxSubmissionFinalizationClaimRequest) (mail.MailboxSubmissionFinalizationClaimResult, error) {
	return mail.NewMailboxSubmissionFinalizationNoopResult(), nil
}

func (*submissionFinalizationWorkerRepositorySpy) CompleteAcceptedMailboxSubmission(context.Context, mail.MailboxSubmissionFinalizationCompletion) error {
	return nil
}

func (*submissionFinalizationWorkerRepositorySpy) ReleaseAcceptedMailboxSubmission(context.Context, mail.MailboxSubmissionFinalizationClaim) error {
	return nil
}

func (spy *submissionFinalizationWorkerRepositorySpy) scanRequests() []mail.MailboxSubmissionFinalizationScanRequest {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	return append([]mail.MailboxSubmissionFinalizationScanRequest(nil), spy.requests...)
}

type submissionFinalizationWorkerExecutorSpy struct {
	mu       sync.Mutex
	finalize func(context.Context, uuid.UUID) error
	ids      []uuid.UUID
}

var _ mail.MailboxSubmissionFinalizationExecutor = (*submissionFinalizationWorkerExecutorSpy)(nil)

func (spy *submissionFinalizationWorkerExecutorSpy) Finalize(ctx context.Context, submissionID uuid.UUID) error {
	spy.mu.Lock()
	spy.ids = append(spy.ids, submissionID)
	callback := spy.finalize
	spy.mu.Unlock()
	if callback == nil {
		return nil
	}
	return callback(ctx, submissionID)
}

func (spy *submissionFinalizationWorkerExecutorSpy) submissionIDs() []uuid.UUID {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	return append([]uuid.UUID(nil), spy.ids...)
}

type submissionFinalizationWorkerTimerSpy struct {
	mu       sync.Mutex
	timeouts []time.Duration
}

var _ mail.MailboxSubmissionWorkerTimer = (*submissionFinalizationWorkerTimerSpy)(nil)

func (*submissionFinalizationWorkerTimerSpy) Wait(context.Context, time.Duration) error {
	return nil
}

func (spy *submissionFinalizationWorkerTimerSpy) WithTimeout(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	spy.mu.Lock()
	spy.timeouts = append(spy.timeouts, duration)
	spy.mu.Unlock()
	return context.WithCancel(ctx)
}

func (spy *submissionFinalizationWorkerTimerSpy) timeoutDurations() []time.Duration {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	return append([]time.Duration(nil), spy.timeouts...)
}
