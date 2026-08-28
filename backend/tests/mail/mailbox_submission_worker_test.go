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
	submissionWorkerIDOne = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	submissionWorkerIDTwo = uuid.MustParse("22222222-2222-4222-8222-222222222222")
)

func TestMailboxSubmissionWorkerDefaultsAndValidation(t *testing.T) {
	defaults := mail.DefaultMailboxSubmissionWorkerOptions()
	if !defaults.Enabled || defaults.ScanInterval != time.Minute || defaults.DispatchTimeout != 30*time.Second || defaults.BatchSize != 20 {
		t.Fatalf("DefaultMailboxSubmissionWorkerOptions() = %#v", defaults)
	}

	disabled, err := mail.NewMailboxSubmissionWorker(nil, nil, nil, mail.MailboxSubmissionWorkerOptions{})
	if err != nil || disabled == nil {
		t.Fatalf("disabled worker = %#v, %v", disabled, err)
	}
	if result, runErr := disabled.RunOnce(context.Background()); result != (mail.MailboxSubmissionWorkerRunResult{}) || runErr != nil {
		t.Fatalf("disabled RunOnce() = %#v, %v", result, runErr)
	}
	if runErr := disabled.Run(context.Background()); runErr != nil {
		t.Fatalf("disabled Run() = %v", runErr)
	}

	scanner := &submissionWorkerScannerFake{}
	dispatcher := &submissionWorkerDispatcherFake{}
	timer := &submissionWorkerTimerFake{}
	var typedNilScanner *submissionWorkerScannerFake
	var typedNilDispatcher *submissionWorkerDispatcherFake
	var typedNilTimer *submissionWorkerTimerFake
	for _, testCase := range []struct {
		name       string
		scanner    mail.MailboxSubmissionWorkerDueScanner
		dispatcher mail.MailboxSubmissionWorkerDispatcher
		timer      mail.MailboxSubmissionWorkerTimer
		options    mail.MailboxSubmissionWorkerOptions
		want       error
	}{
		{"disabled dependency", scanner, nil, nil, mail.MailboxSubmissionWorkerOptions{}, mail.ErrMailboxSubmissionWorkerInvalid},
		{"disabled option", nil, nil, nil, mail.MailboxSubmissionWorkerOptions{BatchSize: 1}, mail.ErrMailboxSubmissionWorkerInvalid},
		{"nil scanner", nil, dispatcher, timer, defaults, mail.ErrMailboxSubmissionWorkerInvalid},
		{"typed nil scanner", typedNilScanner, dispatcher, timer, defaults, mail.ErrMailboxSubmissionWorkerInvalid},
		{"typed nil dispatcher", scanner, typedNilDispatcher, timer, defaults, mail.ErrMailboxSubmissionWorkerInvalid},
		{"typed nil timer", scanner, dispatcher, typedNilTimer, defaults, mail.ErrMailboxSubmissionWorkerInvalid},
		{"negative scan interval", scanner, dispatcher, timer, submissionWorkerOptions(func(options *mail.MailboxSubmissionWorkerOptions) { options.ScanInterval = -time.Second }), mail.ErrMailboxSubmissionWorkerInvalid},
		{"short dispatch timeout", scanner, dispatcher, timer, submissionWorkerOptions(func(options *mail.MailboxSubmissionWorkerOptions) { options.DispatchTimeout = time.Second - time.Nanosecond }), mail.ErrMailboxSubmissionWorkerInvalid},
		{"batch cap", scanner, dispatcher, timer, submissionWorkerOptions(func(options *mail.MailboxSubmissionWorkerOptions) { options.BatchSize = 51 }), mail.ErrMailboxSubmissionWorkerTooLarge},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			worker, constructorErr := mail.NewMailboxSubmissionWorker(testCase.scanner, testCase.dispatcher, testCase.timer, testCase.options)
			if worker != nil || !errors.Is(constructorErr, testCase.want) {
				t.Fatalf("NewMailboxSubmissionWorker() = %#v, %v; want %v", worker, constructorErr, testCase.want)
			}
		})
	}
}

func TestMailboxSubmissionWorkerRunOnceDeduplicatesAndContinues(t *testing.T) {
	scanner := &submissionWorkerScannerFake{scan: func(context.Context, mail.MailboxSubmissionWorkerDueScanRequest) (mail.MailboxSubmissionWorkerDueScanResult, error) {
		return mail.MailboxSubmissionWorkerDueScanResult{SubmissionIDs: []uuid.UUID{uuid.Nil, submissionWorkerIDOne, submissionWorkerIDOne, submissionWorkerIDTwo}}, nil
	}}
	dispatcher := &submissionWorkerDispatcherFake{dispatch: func(_ context.Context, id uuid.UUID) error {
		if id == submissionWorkerIDTwo {
			return errors.New("private dispatcher failure")
		}
		return nil
	}}
	worker, err := mail.NewMailboxSubmissionWorker(scanner, dispatcher, &submissionWorkerTimerFake{}, submissionWorkerOptions(nil))
	if err != nil {
		t.Fatal(err)
	}

	result, runErr := worker.RunOnce(context.Background())
	if !errors.Is(runErr, mail.ErrMailboxSubmissionWorkerPartialFailure) {
		t.Fatalf("RunOnce() error = %v, want content-free partial failure", runErr)
	}
	if result.Due != 2 || result.Dispatched != 1 || result.Failed != 1 || result.ReconciliationRequired != 0 {
		t.Fatalf("RunOnce() result = %#v", result)
	}
	if got := dispatcher.submissionIDs(); !reflect.DeepEqual(got, []uuid.UUID{submissionWorkerIDOne, submissionWorkerIDTwo}) {
		t.Fatalf("dispatcher IDs = %v", got)
	}
	if got := scanner.requests(); len(got) != 1 || got[0].Limit != 20 {
		t.Fatalf("scan requests = %#v", got)
	}
}

func TestMailboxSubmissionWorkerRunOnceBoundsAndCancellation(t *testing.T) {
	scanner := &submissionWorkerScannerFake{scan: func(context.Context, mail.MailboxSubmissionWorkerDueScanRequest) (mail.MailboxSubmissionWorkerDueScanResult, error) {
		return mail.MailboxSubmissionWorkerDueScanResult{SubmissionIDs: submissionWorkerIDs(21)}, nil
	}}
	dispatcher := &submissionWorkerDispatcherFake{}
	worker, err := mail.NewMailboxSubmissionWorker(scanner, dispatcher, &submissionWorkerTimerFake{}, submissionWorkerOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if result, runErr := worker.RunOnce(context.Background()); result != (mail.MailboxSubmissionWorkerRunResult{}) || !errors.Is(runErr, mail.ErrMailboxSubmissionWorkerTooLarge) {
		t.Fatalf("oversize RunOnce() = %#v, %v", result, runErr)
	}
	if len(dispatcher.submissionIDs()) != 0 {
		t.Fatal("oversize scan dispatched an item")
	}

	cancelled, cancel := context.WithCancel(context.Background())
	cancel()
	if result, runErr := worker.RunOnce(cancelled); result != (mail.MailboxSubmissionWorkerRunResult{}) || runErr != nil {
		t.Fatalf("cancelled RunOnce() = %#v, %v", result, runErr)
	}
	if len(scanner.requests()) != 1 {
		t.Fatal("cancelled RunOnce scanned due state")
	}
}

func TestMailboxSubmissionWorkerRunUsesTimerAndCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	scanner := &submissionWorkerScannerFake{scan: func(context.Context, mail.MailboxSubmissionWorkerDueScanRequest) (mail.MailboxSubmissionWorkerDueScanResult, error) {
		return mail.MailboxSubmissionWorkerDueScanResult{SubmissionIDs: []uuid.UUID{}}, nil
	}}
	timer := &submissionWorkerTimerFake{wait: func(_ context.Context, duration time.Duration) error {
		if duration != time.Minute {
			t.Fatalf("Wait duration = %v", duration)
		}
		cancel()
		return context.Canceled
	}}
	worker, err := mail.NewMailboxSubmissionWorker(scanner, &submissionWorkerDispatcherFake{}, timer, submissionWorkerOptions(nil))
	if err != nil {
		t.Fatal(err)
	}
	if err := worker.Run(ctx); err != nil {
		t.Fatalf("Run() = %v", err)
	}
	if timer.waitCount() != 1 || len(scanner.requests()) != 1 {
		t.Fatalf("waits/scans = %d/%d", timer.waitCount(), len(scanner.requests()))
	}
}

func TestMailboxSubmissionWorkerSourceContract(t *testing.T) {
	source, err := os.ReadFile("../../internal/mail/mailbox_submission_worker.go")
	if err != nil {
		t.Fatalf("read worker source: %v", err)
	}
	text := string(source)
	for _, required := range []string{
		"type MailboxSubmissionWorkerDueScanner interface",
		"ScanDueMailboxSubmissionIDs",
		"type MailboxSubmissionWorkerDispatcher interface",
		"func (worker *MailboxSubmissionWorker) RunOnce",
		"func (worker *MailboxSubmissionWorker) Run",
		"MailboxSubmissionDispatchErrorReconciliationRequired",
		"mailboxSubmissionWorkerDueIDs",
	} {
		if !strings.Contains(text, required) {
			t.Fatalf("worker source is missing %q", required)
		}
	}
	for _, forbidden := range []string{"Submit(", "nats.", "ent.", "NewMailboxEntRepository", "ResolveMailboxSubmissionDispatch", "ClaimMailboxSubmission", "AuthorizeMailboxSubmission"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("worker source contains forbidden direct authority %q", forbidden)
		}
	}
	start := strings.Index(text, "func (worker *MailboxSubmissionWorker) runOnce")
	end := strings.Index(text[start:], "func (worker *MailboxSubmissionWorker) safeScan")
	if start < 0 || end < 0 {
		t.Fatal("worker runOnce source boundary is missing")
	}
	runOnce := text[start : start+end]
	if calls := strings.Count(runOnce, "worker.dispatcher.Dispatch("); calls != 1 {
		t.Fatalf("RunOnce dispatch call sites = %d, want exactly one", calls)
	}
	if strings.Contains(strings.ToLower(runOnce), "retry") {
		t.Fatal("RunOnce must not retry due dispatches; durable due state is authoritative")
	}
}

func submissionWorkerOptions(mutate func(*mail.MailboxSubmissionWorkerOptions)) mail.MailboxSubmissionWorkerOptions {
	options := mail.DefaultMailboxSubmissionWorkerOptions()
	if mutate != nil {
		mutate(&options)
	}
	return options
}

func submissionWorkerIDs(count int) []uuid.UUID {
	values := make([]uuid.UUID, count)
	for index := range values {
		values[index] = uuid.NewSHA1(uuid.NameSpaceURL, []byte{byte(index), byte(index >> 8)})
	}
	return values
}

type submissionWorkerScannerFake struct {
	mu               sync.Mutex
	scan             func(context.Context, mail.MailboxSubmissionWorkerDueScanRequest) (mail.MailboxSubmissionWorkerDueScanResult, error)
	recordedRequests []mail.MailboxSubmissionWorkerDueScanRequest
}

func (fake *submissionWorkerScannerFake) ScanDueMailboxSubmissionIDs(ctx context.Context, request mail.MailboxSubmissionWorkerDueScanRequest) (mail.MailboxSubmissionWorkerDueScanResult, error) {
	fake.mu.Lock()
	fake.recordedRequests = append(fake.recordedRequests, request)
	callback := fake.scan
	fake.mu.Unlock()
	if callback == nil {
		return mail.MailboxSubmissionWorkerDueScanResult{SubmissionIDs: []uuid.UUID{}}, nil
	}
	return callback(ctx, request)
}

func (fake *submissionWorkerScannerFake) requests() []mail.MailboxSubmissionWorkerDueScanRequest {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	return append([]mail.MailboxSubmissionWorkerDueScanRequest(nil), fake.recordedRequests...)
}

type submissionWorkerDispatcherFake struct {
	mu     sync.Mutex
	dispatch func(context.Context, uuid.UUID) error
	ids    []uuid.UUID
}

func (fake *submissionWorkerDispatcherFake) Dispatch(ctx context.Context, id uuid.UUID) error {
	fake.mu.Lock()
	fake.ids = append(fake.ids, id)
	callback := fake.dispatch
	fake.mu.Unlock()
	if callback == nil {
		return nil
	}
	return callback(ctx, id)
}

func (fake *submissionWorkerDispatcherFake) submissionIDs() []uuid.UUID {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	return append([]uuid.UUID(nil), fake.ids...)
}

type submissionWorkerTimerFake struct {
	mu          sync.Mutex
	wait        func(context.Context, time.Duration) error
	withTimeout func(context.Context, time.Duration) (context.Context, context.CancelFunc)
	waits       int
}

func (fake *submissionWorkerTimerFake) Wait(ctx context.Context, duration time.Duration) error {
	fake.mu.Lock()
	fake.waits++
	callback := fake.wait
	fake.mu.Unlock()
	if callback == nil {
		return context.Canceled
	}
	return callback(ctx, duration)
}

func (fake *submissionWorkerTimerFake) WithTimeout(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	fake.mu.Lock()
	callback := fake.withTimeout
	fake.mu.Unlock()
	if callback == nil {
		return context.WithCancel(ctx)
	}
	return callback(ctx, duration)
}

func (fake *submissionWorkerTimerFake) waitCount() int {
	fake.mu.Lock()
	defer fake.mu.Unlock()
	return fake.waits
}
