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
	submissionStaleWorkerIDOne = uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	submissionStaleWorkerIDTwo = uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")
)

func TestMailboxSubmissionStaleWorkerDefaultsAndWiring(t *testing.T) {
	defaults := mail.DefaultMailboxSubmissionStaleWorkerOptions()
	if !defaults.Enabled || defaults.ScanInterval != time.Minute || defaults.RecoveryTimeout != 30*time.Second || defaults.BatchSize != 20 {
		t.Fatalf("DefaultMailboxSubmissionStaleWorkerOptions() = %#v", defaults)
	}

	repository := &submissionStaleWorkerRepositorySpy{
		scan: func(context.Context, mail.MailboxSubmissionRecoveryScanRequest) (mail.MailboxSubmissionRecoveryScanResult, error) {
			return mail.MailboxSubmissionRecoveryScanResult{Submissions: []mail.MailboxSubmissionRecoveryDue{{SubmissionID: submissionStaleWorkerIDOne}}}, nil
		},
	}
	timer := &submissionStaleWorkerTimerSpy{}
	worker, err := mail.NewMailboxSubmissionStaleWorker(repository, timer, mail.MailboxSubmissionStaleWorkerOptions{
		Enabled:         true,
		ScanInterval:    2 * time.Second,
		RecoveryTimeout: 3 * time.Second,
		BatchSize:       7,
	})
	if err != nil {
		t.Fatal(err)
	}

	result, runErr := worker.RunOnce(context.Background())
	if runErr != nil || result != (mail.MailboxSubmissionStaleWorkerRunResult{Due: 1, Recovered: 1}) {
		t.Fatalf("RunOnce() = %#v, %v", result, runErr)
	}
	if requests := repository.scanRequests(); !reflect.DeepEqual(requests, []mail.MailboxSubmissionRecoveryScanRequest{{Limit: 7}}) {
		t.Fatalf("scan requests = %#v", requests)
	}
	if durations := timer.timeoutDurations(); !reflect.DeepEqual(durations, []time.Duration{3 * time.Second}) {
		t.Fatalf("timeout durations = %v", durations)
	}
	if requests := repository.recoveryRequests(); !reflect.DeepEqual(requests, []mail.MailboxSubmissionStaleRecoveryRequest{{SubmissionID: submissionStaleWorkerIDOne}}) {
		t.Fatalf("recovery requests = %#v", requests)
	}
}

func TestMailboxSubmissionStaleWorkerAcceptsEveryDurableOutcome(t *testing.T) {
	for _, outcome := range []mail.MailboxSubmissionStaleRecoveryOutcome{
		mail.MailboxSubmissionStaleRecoveryOutcomeNoop,
		mail.MailboxSubmissionStaleRecoveryOutcomeRequeued,
		mail.MailboxSubmissionStaleRecoveryOutcomeFailed,
		mail.MailboxSubmissionStaleRecoveryOutcomeReconciliationRequired,
		mail.MailboxSubmissionStaleRecoveryOutcomeLeaseReleased,
	} {
		t.Run(string(outcome), func(t *testing.T) {
			repository := &submissionStaleWorkerRepositorySpy{
				scan: func(context.Context, mail.MailboxSubmissionRecoveryScanRequest) (mail.MailboxSubmissionRecoveryScanResult, error) {
					return mail.MailboxSubmissionRecoveryScanResult{Submissions: []mail.MailboxSubmissionRecoveryDue{{SubmissionID: submissionStaleWorkerIDOne}}}, nil
				},
				recover: func(context.Context, mail.MailboxSubmissionStaleRecoveryRequest) (mail.MailboxSubmissionStaleRecoveryResult, error) {
					return submissionStaleWorkerResult(outcome), nil
				},
			}
			worker, err := mail.NewMailboxSubmissionStaleWorker(repository, &submissionStaleWorkerTimerSpy{}, submissionStaleWorkerOptions(nil))
			if err != nil {
				t.Fatal(err)
			}
			if result, runErr := worker.RunOnce(context.Background()); runErr != nil || result != (mail.MailboxSubmissionStaleWorkerRunResult{Due: 1, Recovered: 1}) {
				t.Fatalf("RunOnce() = %#v, %v", result, runErr)
			}
		})
	}
}

func TestMailboxSubmissionStaleWorkerRejectsInvalidNilErrorRecoveryResults(t *testing.T) {
	retryAt := time.Date(2026, time.August, 25, 10, 0, 0, 0, time.UTC)
	retryExhausted := mail.MailboxSubmissionErrorCodeRetryExhausted
	reconciliationRequired := mail.MailboxSubmissionErrorCodeReconciliationRequired
	for _, testCase := range []struct {
		name   string
		result mail.MailboxSubmissionStaleRecoveryResult
	}{
		{name: "zero result"},
		{name: "unknown outcome", result: mail.MailboxSubmissionStaleRecoveryResult{Outcome: "unknown"}},
		{name: "noop retry time", result: mail.MailboxSubmissionStaleRecoveryResult{Outcome: mail.MailboxSubmissionStaleRecoveryOutcomeNoop, RetryAt: &retryAt}},
		{name: "requeued without retry time", result: mail.MailboxSubmissionStaleRecoveryResult{Outcome: mail.MailboxSubmissionStaleRecoveryOutcomeRequeued}},
		{name: "failed reconciliation code", result: mail.MailboxSubmissionStaleRecoveryResult{Outcome: mail.MailboxSubmissionStaleRecoveryOutcomeFailed, FailureCode: &reconciliationRequired}},
		{name: "reconciliation retry exhausted code", result: mail.MailboxSubmissionStaleRecoveryResult{Outcome: mail.MailboxSubmissionStaleRecoveryOutcomeReconciliationRequired, FailureCode: &retryExhausted}},
		{name: "lease released failure code", result: mail.MailboxSubmissionStaleRecoveryResult{Outcome: mail.MailboxSubmissionStaleRecoveryOutcomeLeaseReleased, FailureCode: &retryExhausted}},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &submissionStaleWorkerRepositorySpy{
				scan: func(context.Context, mail.MailboxSubmissionRecoveryScanRequest) (mail.MailboxSubmissionRecoveryScanResult, error) {
					return mail.MailboxSubmissionRecoveryScanResult{Submissions: []mail.MailboxSubmissionRecoveryDue{{SubmissionID: submissionStaleWorkerIDOne}}}, nil
				},
				recover: func(context.Context, mail.MailboxSubmissionStaleRecoveryRequest) (mail.MailboxSubmissionStaleRecoveryResult, error) {
					return testCase.result, nil
				},
			}
			worker, err := mail.NewMailboxSubmissionStaleWorker(repository, &submissionStaleWorkerTimerSpy{}, submissionStaleWorkerOptions(nil))
			if err != nil {
				t.Fatal(err)
			}
			result, runErr := worker.RunOnce(context.Background())
			if result != (mail.MailboxSubmissionStaleWorkerRunResult{Due: 1, Failed: 1}) || !errors.Is(runErr, mail.ErrMailboxSubmissionWorkerPartialFailure) {
				t.Fatalf("RunOnce() = %#v, %v", result, runErr)
			}
			if requests := repository.recoveryRequests(); !reflect.DeepEqual(requests, []mail.MailboxSubmissionStaleRecoveryRequest{{SubmissionID: submissionStaleWorkerIDOne}}) {
				t.Fatalf("recovery requests = %#v", requests)
			}
		})
	}
}

func TestMailboxSubmissionStaleWorkerDisabledAndOrderedUniqueIDs(t *testing.T) {
	disabled, err := mail.NewMailboxSubmissionStaleWorker(nil, nil, mail.MailboxSubmissionStaleWorkerOptions{})
	if err != nil || disabled == nil {
		t.Fatalf("disabled worker = %#v, %v", disabled, err)
	}
	if result, runErr := disabled.RunOnce(context.Background()); result != (mail.MailboxSubmissionStaleWorkerRunResult{}) || runErr != nil {
		t.Fatalf("disabled RunOnce() = %#v, %v", result, runErr)
	}

	repository := &submissionStaleWorkerRepositorySpy{
		scan: func(context.Context, mail.MailboxSubmissionRecoveryScanRequest) (mail.MailboxSubmissionRecoveryScanResult, error) {
			return mail.MailboxSubmissionRecoveryScanResult{Submissions: []mail.MailboxSubmissionRecoveryDue{
				{SubmissionID: uuid.Nil},
				{SubmissionID: submissionStaleWorkerIDOne},
				{SubmissionID: submissionStaleWorkerIDOne},
				{SubmissionID: submissionStaleWorkerIDTwo},
			}}, nil
		},
	}
	worker, err := mail.NewMailboxSubmissionStaleWorker(repository, &submissionStaleWorkerTimerSpy{}, submissionStaleWorkerOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	result, runErr := worker.RunOnce(context.Background())
	if runErr != nil || result != (mail.MailboxSubmissionStaleWorkerRunResult{Due: 2, Recovered: 2}) {
		t.Fatalf("RunOnce() = %#v, %v", result, runErr)
	}
	if requests := repository.recoveryRequests(); !reflect.DeepEqual(requests, []mail.MailboxSubmissionStaleRecoveryRequest{{SubmissionID: submissionStaleWorkerIDOne}, {SubmissionID: submissionStaleWorkerIDTwo}}) {
		t.Fatalf("recovery requests = %#v", requests)
	}
}

func TestMailboxSubmissionStaleWorkerFailureCancellationAndTypedNil(t *testing.T) {
	repository := &submissionStaleWorkerRepositorySpy{
		scan: func(context.Context, mail.MailboxSubmissionRecoveryScanRequest) (mail.MailboxSubmissionRecoveryScanResult, error) {
			return mail.MailboxSubmissionRecoveryScanResult{Submissions: []mail.MailboxSubmissionRecoveryDue{{SubmissionID: submissionStaleWorkerIDOne}, {SubmissionID: submissionStaleWorkerIDTwo}}}, nil
		},
		recover: func(_ context.Context, request mail.MailboxSubmissionStaleRecoveryRequest) (mail.MailboxSubmissionStaleRecoveryResult, error) {
			if request.SubmissionID == submissionStaleWorkerIDTwo {
				return mail.MailboxSubmissionStaleRecoveryResult{}, errors.New("private stale recovery failure " + request.SubmissionID.String())
			}
			return mail.MailboxSubmissionStaleRecoveryResult{Outcome: mail.MailboxSubmissionStaleRecoveryOutcomeNoop}, nil
		},
	}
	worker, err := mail.NewMailboxSubmissionStaleWorker(repository, &submissionStaleWorkerTimerSpy{}, submissionStaleWorkerOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	result, runErr := worker.RunOnce(context.Background())
	if result != (mail.MailboxSubmissionStaleWorkerRunResult{Due: 2, Recovered: 1, Failed: 1}) || !errors.Is(runErr, mail.ErrMailboxSubmissionWorkerPartialFailure) {
		t.Fatalf("RunOnce() = %#v, %v", result, runErr)
	}
	if strings.Contains(runErr.Error(), submissionStaleWorkerIDTwo.String()) {
		t.Fatalf("partial failure leaked submission ID: %v", runErr)
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if result, runErr := worker.RunOnce(cancelled); result != (mail.MailboxSubmissionStaleWorkerRunResult{}) || runErr != nil {
		t.Fatalf("cancelled RunOnce() = %#v, %v", result, runErr)
	}

	var typedNilRepository *submissionStaleWorkerRepositorySpy
	var typedNilTimer *submissionStaleWorkerTimerSpy
	for _, testCase := range []struct {
		name       string
		repository mail.MailboxSubmissionStaleRecoveryRepository
		timer      mail.MailboxSubmissionWorkerTimer
	}{
		{name: "repository", repository: typedNilRepository, timer: &submissionStaleWorkerTimerSpy{}},
		{name: "timer", repository: &submissionStaleWorkerRepositorySpy{}, timer: typedNilTimer},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			got, constructorErr := mail.NewMailboxSubmissionStaleWorker(testCase.repository, testCase.timer, submissionStaleWorkerOptions(nil))
			if got != nil || !errors.Is(constructorErr, mail.ErrMailboxSubmissionWorkerInvalid) {
				t.Fatalf("NewMailboxSubmissionStaleWorker() = %#v, %v", got, constructorErr)
			}
		})
	}
}

func TestMailboxSubmissionStaleWorkerSourceContract(t *testing.T) {
	source, err := os.ReadFile("../../internal/mail/mailbox_submission_stale_worker.go")
	if err != nil {
		t.Fatalf("read stale worker source: %v", err)
	}
	text := string(source)
	for _, required := range []string{
		"type MailboxSubmissionStaleWorkerOptions struct",
		"func NewMailboxSubmissionStaleWorker",
		"func (worker *MailboxSubmissionStaleWorker) RunOnce",
		"ScanDueMailboxSubmissionRecovery",
		"mailboxSubmissionStaleRecoveryResultValid(result)",
		"MailboxSubmissionWorkerOptions{",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("stale worker source is missing %q", required)
		}
	}
	for _, forbidden := range []string{"MailEngine", "Provider", "Transport", "NATS", "Submit(", "Preflight", "GetMessage", "GetBlob", "Reconcile(", "Finalize(", "http.", "nats."} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("stale worker source contains forbidden capability %q", forbidden)
		}
	}
	if calls := strings.Count(text, ".RecoverStaleMailboxSubmission("); calls != 1 {
		t.Fatalf("RecoverStaleMailboxSubmission call sites = %d, want one private dispatcher adapter call", calls)
	}
}

func submissionStaleWorkerOptions(mutate func(*mail.MailboxSubmissionStaleWorkerOptions)) mail.MailboxSubmissionStaleWorkerOptions {
	options := mail.DefaultMailboxSubmissionStaleWorkerOptions()
	if mutate != nil {
		mutate(&options)
	}
	return options
}

func submissionStaleWorkerResult(outcome mail.MailboxSubmissionStaleRecoveryOutcome) mail.MailboxSubmissionStaleRecoveryResult {
	switch outcome {
	case mail.MailboxSubmissionStaleRecoveryOutcomeRequeued:
		retryAt := time.Date(2026, time.August, 25, 10, 0, 0, 0, time.UTC)
		return mail.MailboxSubmissionStaleRecoveryResult{Outcome: outcome, RetryAt: &retryAt}
	case mail.MailboxSubmissionStaleRecoveryOutcomeFailed:
		code := mail.MailboxSubmissionErrorCodeRetryExhausted
		return mail.MailboxSubmissionStaleRecoveryResult{Outcome: outcome, FailureCode: &code}
	case mail.MailboxSubmissionStaleRecoveryOutcomeReconciliationRequired:
		code := mail.MailboxSubmissionErrorCodeReconciliationRequired
		return mail.MailboxSubmissionStaleRecoveryResult{Outcome: outcome, FailureCode: &code}
	default:
		return mail.MailboxSubmissionStaleRecoveryResult{Outcome: outcome}
	}
}

type submissionStaleWorkerRepositorySpy struct {
	mu       sync.Mutex
	scan     func(context.Context, mail.MailboxSubmissionRecoveryScanRequest) (mail.MailboxSubmissionRecoveryScanResult, error)
	recover  func(context.Context, mail.MailboxSubmissionStaleRecoveryRequest) (mail.MailboxSubmissionStaleRecoveryResult, error)
	scans    []mail.MailboxSubmissionRecoveryScanRequest
	recovery []mail.MailboxSubmissionStaleRecoveryRequest
}

var _ mail.MailboxSubmissionStaleRecoveryRepository = (*submissionStaleWorkerRepositorySpy)(nil)

func (spy *submissionStaleWorkerRepositorySpy) ScanDueMailboxSubmissionRecovery(ctx context.Context, request mail.MailboxSubmissionRecoveryScanRequest) (mail.MailboxSubmissionRecoveryScanResult, error) {
	spy.mu.Lock()
	spy.scans = append(spy.scans, request)
	callback := spy.scan
	spy.mu.Unlock()
	if callback == nil {
		return mail.MailboxSubmissionRecoveryScanResult{Submissions: []mail.MailboxSubmissionRecoveryDue{}}, nil
	}
	return callback(ctx, request)
}

func (spy *submissionStaleWorkerRepositorySpy) RecoverStaleMailboxSubmission(ctx context.Context, request mail.MailboxSubmissionStaleRecoveryRequest) (mail.MailboxSubmissionStaleRecoveryResult, error) {
	spy.mu.Lock()
	spy.recovery = append(spy.recovery, request)
	callback := spy.recover
	spy.mu.Unlock()
	if callback == nil {
		return mail.MailboxSubmissionStaleRecoveryResult{Outcome: mail.MailboxSubmissionStaleRecoveryOutcomeNoop}, nil
	}
	return callback(ctx, request)
}

func (spy *submissionStaleWorkerRepositorySpy) scanRequests() []mail.MailboxSubmissionRecoveryScanRequest {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	return append([]mail.MailboxSubmissionRecoveryScanRequest(nil), spy.scans...)
}

func (spy *submissionStaleWorkerRepositorySpy) recoveryRequests() []mail.MailboxSubmissionStaleRecoveryRequest {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	return append([]mail.MailboxSubmissionStaleRecoveryRequest(nil), spy.recovery...)
}

type submissionStaleWorkerTimerSpy struct {
	mu       sync.Mutex
	timeouts []time.Duration
}

var _ mail.MailboxSubmissionWorkerTimer = (*submissionStaleWorkerTimerSpy)(nil)

func (*submissionStaleWorkerTimerSpy) Wait(context.Context, time.Duration) error {
	return nil
}

func (spy *submissionStaleWorkerTimerSpy) WithTimeout(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	spy.mu.Lock()
	spy.timeouts = append(spy.timeouts, duration)
	spy.mu.Unlock()
	return context.WithCancel(ctx)
}

func (spy *submissionStaleWorkerTimerSpy) timeoutDurations() []time.Duration {
	spy.mu.Lock()
	defer spy.mu.Unlock()
	return append([]time.Duration(nil), spy.timeouts...)
}
