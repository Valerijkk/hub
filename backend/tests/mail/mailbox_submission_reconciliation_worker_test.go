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
	submissionReconciliationWorkerIDOne = uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	submissionReconciliationWorkerIDTwo = uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
)

func TestMailboxSubmissionReconciliationWorkerDefaultsAndWiring(t *testing.T) {
	defaults := mail.DefaultMailboxSubmissionReconciliationWorkerOptions()
	if !defaults.Enabled || defaults.ScanInterval != time.Minute || defaults.ReconcileTimeout != 30*time.Second || defaults.BatchSize != 20 {
		t.Fatalf("DefaultMailboxSubmissionReconciliationWorkerOptions() = %#v", defaults)
	}

	repository := &submissionReconciliationWorkerRepositorySpy{
		scan: func(context.Context, mail.MailboxSubmissionReconciliationScanRequest) (mail.MailboxSubmissionReconciliationScanResult, error) {
			return mail.MailboxSubmissionReconciliationScanResult{SubmissionIDs: []uuid.UUID{submissionReconciliationWorkerIDOne}}, nil
		},
	}
	executor := &submissionReconciliationWorkerExecutorSpy{}
	timer := &submissionReconciliationWorkerTimerSpy{}
	worker, err := mail.NewMailboxSubmissionReconciliationWorker(repository, executor, timer, mail.MailboxSubmissionReconciliationWorkerOptions{
		Enabled:          true,
		ScanInterval:     2 * time.Second,
		ReconcileTimeout: 3 * time.Second,
		BatchSize:        7,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, runErr := worker.RunOnce(context.Background())
	if runErr != nil || result != (mail.MailboxSubmissionReconciliationWorkerRunResult{Due: 1, Reconciled: 1}) {
		t.Fatalf("RunOnce() = %#v, %v", result, runErr)
	}
	if requests := repository.scanRequests(); !reflect.DeepEqual(requests, []mail.MailboxSubmissionReconciliationScanRequest{{Limit: 7}}) {
		t.Fatalf("scan requests = %#v", requests)
	}
	if durations := timer.timeoutDurations(); !reflect.DeepEqual(durations, []time.Duration{3 * time.Second}) {
		t.Fatalf("timeout durations = %v", durations)
	}
	if ids := executor.submissionIDs(); !reflect.DeepEqual(ids, []uuid.UUID{submissionReconciliationWorkerIDOne}) {
		t.Fatalf("reconciled IDs = %v", ids)
	}
}

func TestMailboxSubmissionReconciliationWorkerDisabledIsNoop(t *testing.T) {
	worker, err := mail.NewMailboxSubmissionReconciliationWorker(nil, nil, nil, mail.MailboxSubmissionReconciliationWorkerOptions{})
	if err != nil || worker == nil {
		t.Fatalf("disabled worker = %#v, %v", worker, err)
	}
	if result, runErr := worker.RunOnce(context.Background()); result != (mail.MailboxSubmissionReconciliationWorkerRunResult{}) || runErr != nil {
		t.Fatalf("disabled RunOnce() = %#v, %v", result, runErr)
	}
	if runErr := worker.Run(context.Background()); runErr != nil {
		t.Fatalf("disabled Run() = %v", runErr)
	}
}

func TestMailboxSubmissionReconciliationWorkerRunOnceDelegatesOrderedUniqueIDs(t *testing.T) {
	repository := &submissionReconciliationWorkerRepositorySpy{
		scan: func(context.Context, mail.MailboxSubmissionReconciliationScanRequest) (mail.MailboxSubmissionReconciliationScanResult, error) {
			return mail.MailboxSubmissionReconciliationScanResult{SubmissionIDs: []uuid.UUID{uuid.Nil, submissionReconciliationWorkerIDOne, submissionReconciliationWorkerIDOne, submissionReconciliationWorkerIDTwo}}, nil
		},
	}
	executor := &submissionReconciliationWorkerExecutorSpy{}
	worker, err := mail.NewMailboxSubmissionReconciliationWorker(repository, executor, &submissionReconciliationWorkerTimerSpy{}, submissionReconciliationWorkerOptions(nil))
	if err != nil {
		t.Fatal(err)
	}

	result, runErr := worker.RunOnce(context.Background())
	if runErr != nil || result != (mail.MailboxSubmissionReconciliationWorkerRunResult{Due: 2, Reconciled: 2}) {
		t.Fatalf("RunOnce() = %#v, %v", result, runErr)
	}
	if requests := repository.scanRequests(); !reflect.DeepEqual(requests, []mail.MailboxSubmissionReconciliationScanRequest{{Limit: 20}}) {
		t.Fatalf("scan requests = %#v", requests)
	}
	if ids := executor.submissionIDs(); !reflect.DeepEqual(ids, []uuid.UUID{submissionReconciliationWorkerIDOne, submissionReconciliationWorkerIDTwo}) {
		t.Fatalf("reconciled IDs = %v", ids)
	}
}

func TestMailboxSubmissionReconciliationWorkerPartialFailureIsClosed(t *testing.T) {
	repository := &submissionReconciliationWorkerRepositorySpy{
		scan: func(context.Context, mail.MailboxSubmissionReconciliationScanRequest) (mail.MailboxSubmissionReconciliationScanResult, error) {
			return mail.MailboxSubmissionReconciliationScanResult{SubmissionIDs: []uuid.UUID{submissionReconciliationWorkerIDOne, submissionReconciliationWorkerIDTwo}}, nil
		},
	}
	executor := &submissionReconciliationWorkerExecutorSpy{
		reconcile: func(_ context.Context, submissionID uuid.UUID) error {
			if submissionID == submissionReconciliationWorkerIDTwo {
				return errors.New("private reconciliation failure " + submissionID.String())
			}
			return nil
		},
	}
	worker, err := mail.NewMailboxSubmissionReconciliationWorker(repository, executor, &submissionReconciliationWorkerTimerSpy{}, submissionReconciliationWorkerOptions(nil))
	if err != nil {
		t.Fatal(err)
	}

	result, runErr := worker.RunOnce(context.Background())
	if result != (mail.MailboxSubmissionReconciliationWorkerRunResult{Due: 2, Reconciled: 1, Failed: 1}) || !errors.Is(runErr, mail.ErrMailboxSubmissionWorkerPartialFailure) {
		t.Fatalf("RunOnce() = %#v, %v", result, runErr)
	}
	if strings.Contains(runErr.Error(), submissionReconciliationWorkerIDTwo.String()) {
		t.Fatalf("partial failure leaked submission ID: %v", runErr)
	}
}

func TestMailboxSubmissionReconciliationWorkerCancellationAndTypedNil(t *testing.T) {
	repository := &submissionReconciliationWorkerRepositorySpy{}
	executor := &submissionReconciliationWorkerExecutorSpy{}
	timer := &submissionReconciliationWorkerTimerSpy{}
	worker, err := mail.NewMailboxSubmissionReconciliationWorker(repository, executor, timer, submissionReconciliationWorkerOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if result, runErr := worker.RunOnce(cancelled); result != (mail.MailboxSubmissionReconciliationWorkerRunResult{}) || runErr != nil {
		t.Fatalf("cancelled RunOnce() = %#v, %v", result, runErr)
	}
	if requests := repository.scanRequests(); len(requests) != 0 {
		t.Fatalf("cancelled RunOnce scanned unknown submissions: %#v", requests)
	}

	var typedNilRepository *submissionReconciliationWorkerRepositorySpy
	var typedNilExecutor *submissionReconciliationWorkerExecutorSpy
	var typedNilTimer *submissionReconciliationWorkerTimerSpy
	for _, testCase := range []struct {
		name       string
		repository mail.MailboxSubmissionReconciliationRepository
		executor   mail.MailboxSubmissionReconciliationExecutor
		timer      mail.MailboxSubmissionWorkerTimer
	}{
		{name: "repository", repository: typedNilRepository, executor: executor, timer: timer},
		{name: "executor", repository: repository, executor: typedNilExecutor, timer: timer},
		{name: "timer", repository: repository, executor: executor, timer: typedNilTimer},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got, constructorErr := mail.NewMailboxSubmissionReconciliationWorker(testCase.repository, testCase.executor, testCase.timer, submissionReconciliationWorkerOptions(nil))
			if got != nil || !errors.Is(constructorErr, mail.ErrMailboxSubmissionWorkerInvalid) {
				t.Fatalf("NewMailboxSubmissionReconciliationWorker() = %#v, %v", got, constructorErr)
			}
		})
	}
}

func TestMailboxSubmissionReconciliationWorkerSourceContract(t *testing.T) {
	source, err := os.ReadFile("../../internal/mail/mailbox_submission_reconciliation_worker.go")
	if err != nil {
		t.Fatalf("read reconciliation worker source: %v", err)
	}
	text := string(source)
	for _, required := range []string{
		"type MailboxSubmissionReconciliationExecutor interface",
		"func NewMailboxSubmissionReconciliationWorker",
		"func (worker *MailboxSubmissionReconciliationWorker) RunOnce",
		"ScanUnknownMailboxSubmissions",
		"MailboxSubmissionWorkerOptions{",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("reconciliation worker source is missing %q", required)
		}
	}
	for _, forbidden := range []string{"MailEngine", "Provider", "Transport", "NATS", "Submit(", "Preflight", "GetMessage", "http.", "nats."} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("reconciliation worker source contains forbidden capability %q", forbidden)
		}
	}
	if calls := strings.Count(text, ".Reconcile("); calls != 1 {
		t.Fatalf("Reconcile call sites = %d, want one private dispatcher adapter call", calls)
	}
}

func submissionReconciliationWorkerOptions(mutate func(*mail.MailboxSubmissionReconciliationWorkerOptions)) mail.MailboxSubmissionReconciliationWorkerOptions {
	options := mail.DefaultMailboxSubmissionReconciliationWorkerOptions()
	if mutate != nil {
		mutate(&options)
	}
	return options
}

type submissionReconciliationWorkerRepositorySpy struct {
	mu       sync.Mutex
	scan     func(context.Context, mail.MailboxSubmissionReconciliationScanRequest) (mail.MailboxSubmissionReconciliationScanResult, error)
	requests []mail.MailboxSubmissionReconciliationScanRequest
}

var _ mail.MailboxSubmissionReconciliationRepository = (*submissionReconciliationWorkerRepositorySpy)(nil)

func (spy *submissionReconciliationWorkerRepositorySpy) ScanUnknownMailboxSubmissions(ctx context.Context, request mail.MailboxSubmissionReconciliationScanRequest) (mail.MailboxSubmissionReconciliationScanResult, error) {
	spy.mu.Lock()
	spy.requests = append(spy.requests, request)
	callback := spy.scan
	spy.mu.Unlock()
	if callback == nil {
		return mail.MailboxSubmissionReconciliationScanResult{SubmissionIDs: []uuid.UUID{}}, nil
	}
	return callback(ctx, request)
}

func (*submissionReconciliationWorkerRepositorySpy) ClaimUnknownMailboxSubmission(context.Context, mail.MailboxSubmissionReconciliationClaimRequest) (mail.MailboxSubmissionReconciliationClaimResult, error) {
	return mail.MailboxSubmissionReconciliationClaimResult{}, nil
}

func (*submissionReconciliationWorkerRepositorySpy) ResolveUnknownMailboxSubmissionReconciliation(context.Context, mail.MailboxSubmissionReconciliationResolveRequest) error {
	return nil
}

func (spy *submissionReconciliationWorkerRepositorySpy) scanRequests() []mail.MailboxSubmissionReconciliationScanRequest {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	return append([]mail.MailboxSubmissionReconciliationScanRequest(nil), spy.requests...)
}

type submissionReconciliationWorkerExecutorSpy struct {
	mu        sync.Mutex
	reconcile func(context.Context, uuid.UUID) error
	ids       []uuid.UUID
}

var _ mail.MailboxSubmissionReconciliationExecutor = (*submissionReconciliationWorkerExecutorSpy)(nil)

func (spy *submissionReconciliationWorkerExecutorSpy) Reconcile(ctx context.Context, submissionID uuid.UUID) error {
	spy.mu.Lock()
	spy.ids = append(spy.ids, submissionID)
	callback := spy.reconcile
	spy.mu.Unlock()
	if callback == nil {
		return nil
	}
	return callback(ctx, submissionID)
}

func (spy *submissionReconciliationWorkerExecutorSpy) submissionIDs() []uuid.UUID {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	return append([]uuid.UUID(nil), spy.ids...)
}

type submissionReconciliationWorkerTimerSpy struct {
	mu       sync.Mutex
	timeouts []time.Duration
}

var _ mail.MailboxSubmissionWorkerTimer = (*submissionReconciliationWorkerTimerSpy)(nil)

func (*submissionReconciliationWorkerTimerSpy) Wait(context.Context, time.Duration) error {
	return nil
}

func (spy *submissionReconciliationWorkerTimerSpy) WithTimeout(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	spy.mu.Lock()
	spy.timeouts = append(spy.timeouts, duration)
	spy.mu.Unlock()
	return context.WithCancel(ctx)
}

func (spy *submissionReconciliationWorkerTimerSpy) timeoutDurations() []time.Duration {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	return append([]time.Duration(nil), spy.timeouts...)
}
