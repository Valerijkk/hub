package mail_test

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/mail"
)

var (
	workerAccountA = uuid.MustParse("11111111-1111-4111-8111-111111111111")
	workerAccountB = uuid.MustParse("22222222-2222-4222-8222-222222222222")
	workerSelected = time.Date(2032, time.January, 2, 3, 4, 5, 987654321, time.UTC)
)

func TestMailboxSyncWorker_FrozenABIAndContentFreeErrors(t *testing.T) {
	requireWorkerFields(t, reflect.TypeOf(mail.MailboxSyncWorkerDueAccount{}), []string{"AccountID", "DueAt", "UpdatedAt", "SyncEpoch", "MutationEpoch", "LeaseEpoch", "SyncAttempts"}, []reflect.Type{reflect.TypeOf(uuid.UUID{}), reflect.TypeOf(time.Time{}), reflect.TypeOf(time.Time{}), reflect.TypeOf(int64(0)), reflect.TypeOf(int64(0)), reflect.TypeOf(int64(0)), reflect.TypeOf(int32(0))})
	requireWorkerFields(t, reflect.TypeOf(mail.MailboxSyncWorkerScanRequest{}), []string{"Limit"}, []reflect.Type{reflect.TypeOf(int(0))})
	requireWorkerFields(t, reflect.TypeOf(mail.MailboxSyncWorkerScanResult{}), []string{"SelectedAt", "Accounts"}, []reflect.Type{reflect.TypeOf(time.Time{}), reflect.TypeOf([]mail.MailboxSyncWorkerDueAccount{})})
	requireWorkerFields(t, reflect.TypeOf(mail.MailboxSyncWorkerFaultRequest{}), []string{"Account", "SelectedAt", "Failure", "RetryDelay"}, []reflect.Type{reflect.TypeOf(mail.MailboxSyncWorkerDueAccount{}), reflect.TypeOf(time.Time{}), reflect.TypeOf(mail.MailboxSyncWorkerFaultKind("")), reflect.TypeOf(time.Duration(0))})
	requireWorkerFields(t, reflect.TypeOf(mail.MailboxSyncWorkerFaultResult{}), []string{"Rescheduled", "RecordedAt", "NextAttemptAt", "SyncAttempts"}, []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(time.Time{}), reflect.TypeOf(time.Time{}), reflect.TypeOf(int32(0))})
	requireWorkerFields(t, reflect.TypeOf(mail.MailboxSyncWorkerConfig{}), []string{"Enabled", "ScanInterval", "StepTimeout", "RetryMin", "RetryMax", "BatchSize", "Concurrency"}, []reflect.Type{reflect.TypeOf(false), reflect.TypeOf(time.Duration(0)), reflect.TypeOf(time.Duration(0)), reflect.TypeOf(time.Duration(0)), reflect.TypeOf(time.Duration(0)), reflect.TypeOf(int(0)), reflect.TypeOf(int(0))})
	requireWorkerFields(t, reflect.TypeOf(mail.MailboxSyncWorkerError{}), []string{"Kind"}, []reflect.Type{reflect.TypeOf(mail.MailboxSyncWorkerErrorKind(""))})

	requireWorkerMethod(t, reflect.TypeOf((*mail.MailboxSyncWorkerRepository)(nil)).Elem(), "ScanDueMailboxSyncAccounts", reflect.TypeOf((func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error))(nil)))
	requireWorkerMethod(t, reflect.TypeOf((*mail.MailboxSyncWorkerRepository)(nil)).Elem(), "RescheduleMailboxSyncWorkerFault", reflect.TypeOf((func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error))(nil)))
	requireWorkerMethod(t, reflect.TypeOf((*mail.MailboxSyncWorkerStepper)(nil)).Elem(), "Step", reflect.TypeOf((func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error))(nil)))
	requireWorkerMethod(t, reflect.TypeOf((*mail.MailboxSyncWorkerTimer)(nil)).Elem(), "Wait", reflect.TypeOf((func(context.Context, time.Duration) error)(nil)))
	requireWorkerMethod(t, reflect.TypeOf((*mail.MailboxSyncWorkerTimer)(nil)).Elem(), "WithTimeout", reflect.TypeOf((func(context.Context, time.Duration) (context.Context, context.CancelFunc))(nil)))
	requireWorkerMethod(t, reflect.TypeOf((*mail.MailboxSyncWorker)(nil)), "Run", reflect.TypeOf((func(*mail.MailboxSyncWorker, context.Context) error)(nil)))
	requireWorkerMethod(t, reflect.TypeOf((*mail.MailboxSyncWorkerError)(nil)), "Error", reflect.TypeOf((func(*mail.MailboxSyncWorkerError) string)(nil)))
	requireWorkerMethod(t, reflect.TypeOf((*mail.MailboxSyncWorkerError)(nil)), "Unwrap", reflect.TypeOf((func(*mail.MailboxSyncWorkerError) error)(nil)))

	constructor := reflect.TypeOf(mail.NewMailboxSyncWorker)
	if constructor.NumIn() != 4 || constructor.NumOut() != 2 ||
		constructor.In(0) != reflect.TypeOf((*mail.MailboxSyncWorkerRepository)(nil)).Elem() ||
		constructor.In(1) != reflect.TypeOf((*mail.MailboxSyncWorkerStepper)(nil)).Elem() ||
		constructor.In(2) != reflect.TypeOf((*mail.MailboxSyncWorkerTimer)(nil)).Elem() ||
		constructor.In(3) != reflect.TypeOf(mail.MailboxSyncWorkerConfig{}) ||
		constructor.Out(0) != reflect.TypeOf((*mail.MailboxSyncWorker)(nil)) ||
		constructor.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		t.Fatalf("NewMailboxSyncWorker signature changed: %v", constructor)
	}
	if mail.MailboxSyncWorkerFaultTimeout != "timeout" || mail.MailboxSyncWorkerFaultPanic != "panic" || mail.MailboxSyncWorkerFaultUnavailable != "unavailable" || mail.MailboxSyncWorkerFaultInvalidResult != "invalid_result" ||
		mail.MailboxSyncWorkerErrorInvalid != "invalid" || mail.MailboxSyncWorkerErrorTooLarge != "too_large" || mail.MailboxSyncWorkerErrorUnavailable != "unavailable" {
		t.Fatal("frozen worker constants changed")
	}

	for _, testCase := range []struct {
		kind mail.MailboxSyncWorkerErrorKind
		want error
		text string
	}{
		{mail.MailboxSyncWorkerErrorInvalid, mail.ErrMailboxSyncWorkerInvalid, "mailbox sync worker invalid"},
		{mail.MailboxSyncWorkerErrorTooLarge, mail.ErrMailboxSyncWorkerTooLarge, "mailbox sync worker exceeds bounds"},
		{mail.MailboxSyncWorkerErrorUnavailable, mail.ErrMailboxSyncWorkerUnavailable, "mailbox sync worker unavailable"},
	} {
		err := &mail.MailboxSyncWorkerError{Kind: testCase.kind}
		if err.Error() != testCase.text || err.Unwrap() != testCase.want || !errors.Is(err, testCase.want) {
			t.Fatalf("known safe error %#v changed", err)
		}
	}
	var nilError *mail.MailboxSyncWorkerError
	if nilError.Error() != "mailbox sync worker error" || nilError.Unwrap() != nil {
		t.Fatal("nil worker error leaked a detail")
	}
	unknown := &mail.MailboxSyncWorkerError{Kind: "unknown"}
	if unknown.Error() != "mailbox sync worker error" || unknown.Unwrap() != nil || errors.Is(unknown, mail.ErrMailboxSyncWorkerInvalid) || errors.Is(unknown, mail.ErrMailboxSyncWorkerTooLarge) || errors.Is(unknown, mail.ErrMailboxSyncWorkerUnavailable) {
		t.Fatalf("unknown error leaked a sentinel: %#v", unknown)
	}
}

func TestMailboxSyncWorker_DisabledTypedNilAndConfigBounds(t *testing.T) {
	disabled, err := mail.NewMailboxSyncWorker(nil, nil, nil, mail.MailboxSyncWorkerConfig{})
	if err != nil || disabled == nil {
		t.Fatalf("disabled constructor = %#v, %v", disabled, err)
	}
	if err := disabled.Run(context.Background()); err != nil {
		t.Fatalf("disabled Run() = %v", err)
	}

	validRepository := &workerRepositoryFake{}
	validStepper := &workerStepperFake{}
	validTimer := &workerTimerFake{}
	var nilRepository *workerRepositoryFake
	var nilStepper *workerStepperFake
	var nilTimer *workerTimerFake
	for _, testCase := range []struct {
		name string
		repository mail.MailboxSyncWorkerRepository
		stepper mail.MailboxSyncWorkerStepper
		timer mail.MailboxSyncWorkerTimer
		config mail.MailboxSyncWorkerConfig
		want error
	}{
		{"disabled dependency", validRepository, nil, nil, mail.MailboxSyncWorkerConfig{}, mail.ErrMailboxSyncWorkerInvalid},
		{"disabled nonzero config", nil, nil, nil, mail.MailboxSyncWorkerConfig{BatchSize: 1}, mail.ErrMailboxSyncWorkerInvalid},
		{"nil repository", nil, validStepper, validTimer, workerConfig(), mail.ErrMailboxSyncWorkerInvalid},
		{"typed nil repository", nilRepository, validStepper, validTimer, workerConfig(), mail.ErrMailboxSyncWorkerInvalid},
		{"typed nil stepper", validRepository, nilStepper, validTimer, workerConfig(), mail.ErrMailboxSyncWorkerInvalid},
		{"typed nil timer", validRepository, validStepper, nilTimer, workerConfig(), mail.ErrMailboxSyncWorkerInvalid},
		{"zero batch", validRepository, validStepper, validTimer, mutateWorkerConfig(func(value *mail.MailboxSyncWorkerConfig) { value.BatchSize = 0 }), mail.ErrMailboxSyncWorkerInvalid},
		{"batch cap", validRepository, validStepper, validTimer, mutateWorkerConfig(func(value *mail.MailboxSyncWorkerConfig) { value.BatchSize = 51 }), mail.ErrMailboxSyncWorkerTooLarge},
		{"zero concurrency", validRepository, validStepper, validTimer, mutateWorkerConfig(func(value *mail.MailboxSyncWorkerConfig) { value.Concurrency = 0 }), mail.ErrMailboxSyncWorkerInvalid},
		{"concurrency cap", validRepository, validStepper, validTimer, mutateWorkerConfig(func(value *mail.MailboxSyncWorkerConfig) { value.Concurrency = 5 }), mail.ErrMailboxSyncWorkerTooLarge},
		{"scan low", validRepository, validStepper, validTimer, mutateWorkerConfig(func(value *mail.MailboxSyncWorkerConfig) { value.ScanInterval = time.Second - time.Nanosecond }), mail.ErrMailboxSyncWorkerInvalid},
		{"scan cap", validRepository, validStepper, validTimer, mutateWorkerConfig(func(value *mail.MailboxSyncWorkerConfig) { value.ScanInterval = 5*time.Minute + time.Nanosecond }), mail.ErrMailboxSyncWorkerTooLarge},
		{"step low", validRepository, validStepper, validTimer, mutateWorkerConfig(func(value *mail.MailboxSyncWorkerConfig) { value.StepTimeout = time.Second - time.Nanosecond }), mail.ErrMailboxSyncWorkerInvalid},
		{"step cap", validRepository, validStepper, validTimer, mutateWorkerConfig(func(value *mail.MailboxSyncWorkerConfig) { value.StepTimeout = 30*time.Second + time.Nanosecond }), mail.ErrMailboxSyncWorkerTooLarge},
		{"retry ordering", validRepository, validStepper, validTimer, mutateWorkerConfig(func(value *mail.MailboxSyncWorkerConfig) { value.RetryMin, value.RetryMax = 2 * time.Second, time.Second }), mail.ErrMailboxSyncWorkerInvalid},
		{"retry cap", validRepository, validStepper, validTimer, mutateWorkerConfig(func(value *mail.MailboxSyncWorkerConfig) { value.RetryMax = 5*time.Minute + time.Nanosecond }), mail.ErrMailboxSyncWorkerTooLarge},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			worker, constructorErr := mail.NewMailboxSyncWorker(testCase.repository, testCase.stepper, testCase.timer, testCase.config)
			if worker != nil || !errors.Is(constructorErr, testCase.want) {
				t.Fatalf("NewMailboxSyncWorker() = %#v, %v; want %v", worker, constructorErr, testCase.want)
			}
		})
	}
}

func TestMailboxSyncWorker_ScanAndDrainContract(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	accountA := workerDueAccount(workerAccountA, workerSelected.Add(-2*time.Minute))
	accountB := workerDueAccount(workerAccountB, workerSelected.Add(-time.Minute))
	repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) {
		return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{accountA, accountB}}, nil
	}}
	stepper := &workerStepperFake{step: func(_ context.Context, id uuid.UUID) (mail.MailboxSyncStepResult, error) {
		if id == workerAccountA { accountA.DueAt = time.Time{} }
		return mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepIdle, NextAttemptAt: workerSelected.Add(time.Minute)}, nil
	}}
	timer := &workerTimerFake{wait: func(_ context.Context, duration time.Duration) error {
		if duration != workerConfig().ScanInterval { t.Fatalf("wait duration = %v", duration) }
		cancel()
		return context.Canceled
	}}
	worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, workerConfig())
	if err != nil { t.Fatalf("NewMailboxSyncWorker() error = %v", err) }
	if err := worker.Run(ctx); err != nil { t.Fatalf("Run() error = %v", err) }
	if got := repository.scanRequests(); len(got) != 1 || got[0].Limit != workerConfig().BatchSize || stepper.calls() != 2 || repository.faultCount() != 0 || timer.waitCount() != 1 {
		t.Fatalf("scan/step/fault/wait = %#v/%d/%d/%d", got, stepper.calls(), repository.faultCount(), timer.waitCount())
	}
}

func TestMailboxSyncWorker_ScanPreflightAndFractionalTimestamps(t *testing.T) {
	valid := workerDueAccount(workerAccountA, workerSelected.Add(-time.Minute))
	for _, testCase := range []struct {
		name string
		result mail.MailboxSyncWorkerScanResult
		want error
	}{
		{"nil accounts", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected}, mail.ErrMailboxSyncWorkerInvalid},
		{"zero selected", mail.MailboxSyncWorkerScanResult{Accounts: []mail.MailboxSyncWorkerDueAccount{}}, mail.ErrMailboxSyncWorkerInvalid},
		{"non utc selected", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected.In(time.FixedZone("other", 3600)), Accounts: []mail.MailboxSyncWorkerDueAccount{}}, mail.ErrMailboxSyncWorkerInvalid},
		{"future due", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{workerDueAccount(workerAccountA, workerSelected.Add(time.Nanosecond))}}, mail.ErrMailboxSyncWorkerInvalid},
		{"zero account", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{workerDueAccount(uuid.Nil, workerSelected.Add(-time.Second))}}, mail.ErrMailboxSyncWorkerInvalid},
		{"non utc due", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{workerDueAccount(workerAccountA, workerSelected.Add(-time.Second).In(time.FixedZone("other", 3600)))}}, mail.ErrMailboxSyncWorkerInvalid},
		{"out of json range selected", mail.MailboxSyncWorkerScanResult{SelectedAt: time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC), Accounts: []mail.MailboxSyncWorkerDueAccount{}}, mail.ErrMailboxSyncWorkerInvalid},
		{"out of json range due", mail.MailboxSyncWorkerScanResult{SelectedAt: time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC), Accounts: []mail.MailboxSyncWorkerDueAccount{workerDueAccount(workerAccountA, time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC))}}, mail.ErrMailboxSyncWorkerInvalid},
		{"future updated", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{func() mail.MailboxSyncWorkerDueAccount { value := valid; value.UpdatedAt = workerSelected.Add(time.Nanosecond); return value }()}}, mail.ErrMailboxSyncWorkerInvalid},
		{"negative sync epoch", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{func() mail.MailboxSyncWorkerDueAccount { value := valid; value.SyncEpoch = -1; return value }()}}, mail.ErrMailboxSyncWorkerInvalid},
		{"negative mutation epoch", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{func() mail.MailboxSyncWorkerDueAccount { value := valid; value.MutationEpoch = -1; return value }()}}, mail.ErrMailboxSyncWorkerInvalid},
		{"negative attempts", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{func() mail.MailboxSyncWorkerDueAccount { value := valid; value.SyncAttempts = -1; return value }()}}, mail.ErrMailboxSyncWorkerInvalid},
		{"duplicate", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{valid, valid}}, mail.ErrMailboxSyncWorkerInvalid},
		{"duplicate id distinct due", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{workerDueAccount(workerAccountA, workerSelected.Add(-2 * time.Second)), workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))}}, mail.ErrMailboxSyncWorkerInvalid},
		{"due disorder", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{workerDueAccount(workerAccountA, workerSelected.Add(-time.Second)), workerDueAccount(workerAccountB, workerSelected.Add(-2 * time.Second))}}, mail.ErrMailboxSyncWorkerInvalid},
		{"uuid disorder at equal due", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{workerDueAccount(workerAccountB, workerSelected.Add(-time.Second)), workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))}}, mail.ErrMailboxSyncWorkerInvalid},
		{"negative lease epoch", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{func() mail.MailboxSyncWorkerDueAccount { value := valid; value.LeaseEpoch = 0; return value }()}}, mail.ErrMailboxSyncWorkerInvalid},
		{"oversize", mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: workerAccounts(workerConfig().BatchSize + 1)}, mail.ErrMailboxSyncWorkerTooLarge},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) { return testCase.result, nil }}
			stepper := &workerStepperFake{}
			worker, err := mail.NewMailboxSyncWorker(repository, stepper, &workerTimerFake{}, workerConfig())
			if err != nil { t.Fatal(err) }
			if got := worker.Run(context.Background()); !errors.Is(got, testCase.want) || repository.faultCount() != 0 || stepper.calls() != 0 {
				t.Fatalf("Run() = %v, faults/steps=%d/%d; want %v and zero effects", got, repository.faultCount(), stepper.calls(), testCase.want)
			}
		})
	}
}

func TestMailboxSyncWorker_OutcomeFaultAndMaxAttemptContract(t *testing.T) {
	for _, testCase := range []struct {
		name string
		result mail.MailboxSyncStepResult
		stepErr error
		attempts int32
		faults int
		want error
	}{
		{"idle", mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepIdle, NextAttemptAt: workerSelected.Add(time.Second)}, nil, 0, 0, nil},
		{"applied", mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepApplied, NextAttemptAt: workerSelected.Add(time.Second)}, nil, 0, 0, nil},
		{"rebuild", mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepRebuildScheduled, NextAttemptAt: workerSelected.Add(time.Second)}, nil, 0, 0, nil},
		{"retry", mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepRetryScheduled, NextAttemptAt: workerSelected.Add(time.Second)}, nil, 0, 0, nil},
		{"lease lost", mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepLeaseLost}, nil, 0, 0, nil},
		{"invalid result", mail.MailboxSyncStepResult{}, nil, 0, 1, nil},
		{"unavailable", mail.MailboxSyncStepResult{}, errors.New("private engine failure"), 0, 1, nil},
		{"maximum retry", mail.MailboxSyncStepResult{}, errors.New("private engine failure"), int32(^uint32(0) >> 1), 0, mail.ErrMailboxSyncWorkerTooLarge},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			account := workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))
			account.SyncAttempts = testCase.attempts
			repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) { return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{account}}, nil }, fault: func(_ context.Context, request mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) {
				return mail.MailboxSyncWorkerFaultResult{Rescheduled: true, RecordedAt: workerSelected, NextAttemptAt: workerSelected.Add(request.RetryDelay), SyncAttempts: request.Account.SyncAttempts + 1}, nil
			}}
			stepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) { return testCase.result, testCase.stepErr }}
			timer := &workerTimerFake{wait: func(context.Context, time.Duration) error { cancel(); return context.Canceled }}
			worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, workerConfig())
			if err != nil { t.Fatal(err) }
			got := worker.Run(ctx)
			if testCase.want == nil {
				if got != nil { t.Fatalf("Run() = %v", got) }
			} else if !errors.Is(got, testCase.want) {
				t.Fatalf("Run() = %v; want %v", got, testCase.want)
			}
			if repository.faultCount() != testCase.faults { t.Fatalf("fault calls = %d; want %d", repository.faultCount(), testCase.faults) }
		})
	}
}

func TestMailboxSyncWorker_FaultPrecisionBackoffAndTimeoutTotality(t *testing.T) {
	for _, testCase := range []struct {
		name string
		attempts int32
		configureConfig func(*mail.MailboxSyncWorkerConfig)
		configureTimer func(*workerTimerFake)
		wantFailure mail.MailboxSyncWorkerFaultKind
		wantDelay time.Duration
		wantSteps int
	}{
		{"unavailable attempt zero", 0, nil, nil, mail.MailboxSyncWorkerFaultUnavailable, time.Second, 1},
		{"unavailable attempt one", 1, nil, nil, mail.MailboxSyncWorkerFaultUnavailable, 2 * time.Second, 1},
		{"invalid attempt two", 2, nil, nil, mail.MailboxSyncWorkerFaultInvalidResult, 4 * time.Second, 1},
		{"configured clamp", 2, func(config *mail.MailboxSyncWorkerConfig) { config.RetryMax = 3 * time.Second }, nil, mail.MailboxSyncWorkerFaultInvalidResult, 3 * time.Second, 1},
		{"saturating attempt", 62, nil, nil, mail.MailboxSyncWorkerFaultUnavailable, time.Minute, 1},
		{"with timeout panic", 0, nil, func(timer *workerTimerFake) { timer.withTimeout = func(context.Context, time.Duration) (context.Context, context.CancelFunc) { panic("private timer value") } }, mail.MailboxSyncWorkerFaultPanic, time.Second, 0},
		{"nil child", 0, nil, func(timer *workerTimerFake) { timer.withTimeout = func(context.Context, time.Duration) (context.Context, context.CancelFunc) { return nil, func() {} } }, mail.MailboxSyncWorkerFaultInvalidResult, time.Second, 0},
		{"nil cancel", 0, nil, func(timer *workerTimerFake) { timer.withTimeout = func(ctx context.Context, _ time.Duration) (context.Context, context.CancelFunc) { return ctx, nil } }, mail.MailboxSyncWorkerFaultInvalidResult, time.Second, 0},
		{"deadline", 0, nil, func(timer *workerTimerFake) { timer.withTimeout = func(context.Context, time.Duration) (context.Context, context.CancelFunc) { return workerDeadlineContext{}, func() {} } }, mail.MailboxSyncWorkerFaultTimeout, time.Second, 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			account := workerDueAccount(workerAccountA, workerSelected.Add(-time.Minute))
			account.SyncAttempts = testCase.attempts
			repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) {
				return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{account}}, nil
			}, fault: func(_ context.Context, request mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) {
				if request.Account != account || !request.SelectedAt.Equal(workerSelected) || request.SelectedAt.Nanosecond() != workerSelected.Nanosecond() || request.Failure != testCase.wantFailure || request.RetryDelay != testCase.wantDelay {
					t.Fatalf("fault request = %#v", request)
				}
				return mail.MailboxSyncWorkerFaultResult{Rescheduled: true, RecordedAt: workerSelected, NextAttemptAt: workerSelected.Add(request.RetryDelay), SyncAttempts: request.Account.SyncAttempts + 1}, nil
			}}
			stepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) {
				if testCase.wantFailure == mail.MailboxSyncWorkerFaultInvalidResult { return mail.MailboxSyncStepResult{}, nil }
				return mail.MailboxSyncStepResult{}, errors.New("provider detail")
			}}
			timer := &workerTimerFake{wait: func(context.Context, time.Duration) error { cancel(); return context.Canceled }}
			if testCase.configureTimer != nil { testCase.configureTimer(timer) }
			config := workerConfig()
			if testCase.configureConfig != nil { testCase.configureConfig(&config) }
			worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, config)
			if err != nil { t.Fatal(err) }
			if err := worker.Run(ctx); err != nil { t.Fatalf("Run() leaked a contained fault: %v", err) }
			if repository.faultCount() != 1 || stepper.calls() != testCase.wantSteps {
				t.Fatalf("fault/step = %d/%d; want 1/%d", repository.faultCount(), stepper.calls(), testCase.wantSteps)
			}
		})
	}
}

func TestMailboxSyncWorker_FaultResultWaitAndCancelContainment(t *testing.T) {
	for _, testCase := range []struct {
		name string
		fault func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error)
		wait func(context.Context, time.Duration) error
		want error
	}{
		{"false zero", func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) { return mail.MailboxSyncWorkerFaultResult{}, nil }, func(context.Context, time.Duration) error { return context.Canceled }, nil},
		{"fault port error contained", func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) { return mail.MailboxSyncWorkerFaultResult{}, errors.New("private database error") }, func(context.Context, time.Duration) error { return context.Canceled }, nil},
		{"malformed true contained", func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) { return mail.MailboxSyncWorkerFaultResult{Rescheduled: true}, nil }, func(context.Context, time.Duration) error { return context.Canceled }, nil},
		{"fault port panic continues", func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) { panic("private database value") }, func(context.Context, time.Duration) error { return context.Canceled }, nil},
		{"wait error safe", func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) { return mail.MailboxSyncWorkerFaultResult{}, nil }, func(context.Context, time.Duration) error { return errors.New("private timer error") }, mail.ErrMailboxSyncWorkerUnavailable},
		{"wait panic safe", func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) { return mail.MailboxSyncWorkerFaultResult{}, nil }, func(context.Context, time.Duration) error { panic("private timer panic") }, mail.ErrMailboxSyncWorkerUnavailable},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			account := workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))
			repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) { return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{account}}, nil }, fault: testCase.fault}
			stepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) { return mail.MailboxSyncStepResult{}, errors.New("unavailable") }}
			timer := &workerTimerFake{wait: func(waitCtx context.Context, duration time.Duration) error {
				if duration != workerConfig().ScanInterval { t.Fatalf("wait = %v", duration) }
				if testCase.want == nil {
					cancel()
					return context.Canceled
				}
				if testCase.wait != nil { return testCase.wait(waitCtx, duration) }
				return nil
			}}
			worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, workerConfig())
			if err != nil { t.Fatal(err) }
			got := worker.Run(ctx)
			if testCase.want == nil {
				if got != nil { t.Fatalf("Run() = %v", got) }
			} else if !errors.Is(got, testCase.want) {
				t.Fatalf("Run() = %v; want %v", got, testCase.want)
			}
		})
	}
}

func TestMailboxSyncWorker_ScanPanicAndRawErrorRetryWithoutFault(t *testing.T) {
	for _, panics := range []bool{false, true} {
		t.Run(map[bool]string{false: "raw error", true: "panic"}[panics], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) {
				if panics { panic("private scanner panic") }
				return mail.MailboxSyncWorkerScanResult{}, errors.New("private scanner error")
			}}
			timer := &workerTimerFake{wait: func(_ context.Context, duration time.Duration) error {
				if duration != workerConfig().RetryMin { t.Fatalf("retry wait = %v", duration) }
				cancel()
				return context.Canceled
			}}
			worker, err := mail.NewMailboxSyncWorker(repository, &workerStepperFake{}, timer, workerConfig())
			if err != nil { t.Fatal(err) }
			if err := worker.Run(ctx); err != nil || repository.faultCount() != 0 || timer.waitCount() != 1 {
				t.Fatalf("Run()/faults/waits = %v/%d/%d", err, repository.faultCount(), timer.waitCount())
			}
		})
	}
}

func TestMailboxSyncWorker_BoundedFIFOFixedPoolAndDrainBeforeWait(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config := workerConfig()
	config.BatchSize = 50
	config.Concurrency = 4
	accounts := workerAccounts(50)
	var trace struct {
		sync.Mutex
		active int
		maximum int
		seen map[uuid.UUID]int
	}
	trace.seen = make(map[uuid.UUID]int)
	firstWave := make(chan uuid.UUID, config.Concurrency)
	release := make(chan struct{})
	repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) {
		return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: accounts}, nil
	}}
	stepper := &workerStepperFake{step: func(_ context.Context, accountID uuid.UUID) (mail.MailboxSyncStepResult, error) {
		trace.Lock()
		trace.active++
		if trace.active > trace.maximum { trace.maximum = trace.active }
		trace.seen[accountID]++
		admitted := len(trace.seen)
		trace.Unlock()
		if admitted <= config.Concurrency {
			firstWave <- accountID
			<-release
		}
		trace.Lock()
		trace.active--
		trace.Unlock()
		return mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepIdle, NextAttemptAt: workerSelected.Add(time.Second)}, nil
	}}
	timer := &workerTimerFake{wait: func(_ context.Context, duration time.Duration) error {
		if duration != config.ScanInterval || stepper.calls() != len(accounts) { t.Fatalf("wait before complete drain: duration=%v calls=%d", duration, stepper.calls()) }
		trace.Lock()
		defer trace.Unlock()
		if trace.active != 0 || len(trace.seen) != len(accounts) || trace.maximum != config.Concurrency {
			t.Fatalf("pool trace active/seen/max = %d/%d/%d", trace.active, len(trace.seen), trace.maximum)
		}
		for _, account := range accounts { if trace.seen[account.AccountID] != 1 { t.Fatalf("account %s steps=%d", account.AccountID, trace.seen[account.AccountID]) } }
		cancel()
		return context.Canceled
	}}
	worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, config)
	if err != nil { t.Fatal(err) }
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()
	first := make(map[uuid.UUID]struct{}, config.Concurrency)
	for index := 0; index < config.Concurrency; index++ { first[<-firstWave] = struct{}{} }
	if len(first) != config.Concurrency { t.Fatalf("first admitted wave had duplicate IDs: %#v", first) }
	for index := 0; index < config.Concurrency; index++ {
		if _, ok := first[accounts[index].AccountID]; !ok { t.Fatalf("first FIFO wave omitted %s: %#v", accounts[index].AccountID, first) }
	}
	close(release)
	if err := <-done; err != nil { t.Fatalf("Run() = %v", err) }
	if len(repository.scanRequests()) != 1 || repository.faultCount() != 0 {
		t.Fatalf("scan/fault counts = %d/%d", len(repository.scanRequests()), repository.faultCount())
	}
}

func TestMailboxSyncWorker_ParentCancelAndConcurrentRunAreSafe(t *testing.T) {
	t.Run("parent cancellation has fault priority", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		account := workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))
		repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) { return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{account}}, nil }}
		stepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) { cancel(); return mail.MailboxSyncStepResult{}, errors.New("provider value") }}
		worker, err := mail.NewMailboxSyncWorker(repository, stepper, &workerTimerFake{}, workerConfig())
		if err != nil { t.Fatal(err) }
		if err := worker.Run(ctx); err != nil || repository.faultCount() != 0 {
			t.Fatalf("Run()/faults = %v/%d", err, repository.faultCount())
		}
	})

	t.Run("one concurrent run only", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		entered := make(chan struct{})
		release := make(chan struct{})
		repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) {
			close(entered)
			<-release
			return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{}}, nil
		}}
		timer := &workerTimerFake{wait: func(context.Context, time.Duration) error { cancel(); return context.Canceled }}
		worker, err := mail.NewMailboxSyncWorker(repository, &workerStepperFake{}, timer, workerConfig())
		if err != nil { t.Fatal(err) }
		first := make(chan error, 1)
		go func() { first <- worker.Run(ctx) }()
		<-entered
		if err := worker.Run(context.Background()); !errors.Is(err, mail.ErrMailboxSyncWorkerInvalid) { t.Fatalf("concurrent Run() = %v", err) }
		close(release)
		if err := <-first; err != nil { t.Fatalf("first Run() = %v", err) }
	})
}

func TestMailboxSyncWorker_ExactFaultFenceAndFalseResult(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	account := workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))
	account.SyncEpoch = 7
	account.MutationEpoch = 11
	account.LeaseEpoch = 13
	account.SyncAttempts = 3
	repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) { return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{account}}, nil }, fault: func(_ context.Context, request mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) {
		if request.Account != account || !request.SelectedAt.Equal(workerSelected) || request.RetryDelay != 8*time.Second || request.Failure != mail.MailboxSyncWorkerFaultUnavailable {
			t.Fatalf("fault CAS request = %#v", request)
		}
		return mail.MailboxSyncWorkerFaultResult{}, nil
	}}
	stepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) { return mail.MailboxSyncStepResult{}, errors.New("private error") }}
	timer := &workerTimerFake{wait: func(context.Context, time.Duration) error { cancel(); return context.Canceled }}
	worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, workerConfig())
	if err != nil { t.Fatal(err) }
	if err := worker.Run(ctx); err != nil || repository.faultCount() != 1 {
		t.Fatalf("Run()/faults = %v/%d", err, repository.faultCount())
	}
}

func TestMailboxSyncWorker_HostileParentContextsAndReuse(t *testing.T) {
	var nilWorker *mail.MailboxSyncWorker
	if err := nilWorker.Run(context.Background()); !errors.Is(err, mail.ErrMailboxSyncWorkerInvalid) { t.Fatalf("nil receiver = %v", err) }
	repository := &workerRepositoryFake{}
	stepper := &workerStepperFake{}
	timer := &workerTimerFake{}
	worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, workerConfig())
	if err != nil { t.Fatal(err) }
	var nilContext *workerHostileContext
	for _, testCase := range []struct { name string; ctx context.Context; want error }{
		{"nil", nil, mail.ErrMailboxSyncWorkerInvalid},
		{"typed nil", nilContext, mail.ErrMailboxSyncWorkerInvalid},
		{"panic", &workerHostileContext{panicErr: true}, mail.ErrMailboxSyncWorkerUnavailable},
		{"unknown", &workerHostileContext{err: errors.New("private parent state")}, mail.ErrMailboxSyncWorkerUnavailable},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			panicked := false
			var got error
			func() { defer func() { panicked = recover() != nil }(); got = worker.Run(testCase.ctx) }()
			if panicked || !errors.Is(got, testCase.want) { t.Fatalf("Run() = %v panic=%v; want %v", got, panicked, testCase.want) }
		})
	}
	ctx, cancel := context.WithCancel(context.Background())
	timer.wait = func(context.Context, time.Duration) error { cancel(); return context.Canceled }
	if err := worker.Run(ctx); err != nil { t.Fatalf("first reusable Run() = %v", err) }
	ctx, cancel = context.WithCancel(context.Background())
	defer cancel()
	if err := worker.Run(ctx); err != nil { t.Fatalf("second reusable Run() = %v", err) }
}

func TestMailboxSyncWorker_ExactConfigBoundsConcreteEmptyAndConcurrentGuard(t *testing.T) {
	for _, config := range []mail.MailboxSyncWorkerConfig{
		{Enabled: true, ScanInterval: time.Second, StepTimeout: time.Second, RetryMin: time.Second, RetryMax: time.Second, BatchSize: 1, Concurrency: 1},
		{Enabled: true, ScanInterval: 5 * time.Minute, StepTimeout: 30 * time.Second, RetryMin: time.Second, RetryMax: 5 * time.Minute, BatchSize: 50, Concurrency: 4},
	} {
		worker, err := mail.NewMailboxSyncWorker(&workerRepositoryFake{}, &workerStepperFake{}, &workerTimerFake{}, config)
		if err != nil || worker == nil { t.Fatalf("valid bounds worker=%#v err=%v", worker, err) }
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	emptyRepository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) { return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{}}, nil }}
	emptyTimer := &workerTimerFake{wait: func(_ context.Context, duration time.Duration) error { if duration != workerConfig().ScanInterval { t.Fatalf("empty wait duration = %v", duration) }; cancel(); return context.Canceled }}
	emptyWorker, err := mail.NewMailboxSyncWorker(emptyRepository, &workerStepperFake{}, emptyTimer, workerConfig())
	if err != nil { t.Fatal(err) }
	if err := emptyWorker.Run(ctx); err != nil || emptyTimer.waitCount() != 1 || len(emptyRepository.scanRequests()) != 1 || emptyRepository.scanRequests()[0].Limit != workerConfig().BatchSize {
		t.Fatalf("empty Run()/wait/scans = %v/%d/%#v", err, emptyTimer.waitCount(), emptyRepository.scanRequests())
	}

	activeContext, activeCancel := context.WithCancel(context.Background())
	defer activeCancel()
	entered, release := make(chan struct{}), make(chan struct{})
	repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) {
		close(entered)
		<-release
		return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{}}, nil
	}}
	timer := &workerTimerFake{wait: func(context.Context, time.Duration) error { activeCancel(); return context.Canceled }}
	worker, err := mail.NewMailboxSyncWorker(repository, &workerStepperFake{}, timer, workerConfig())
	if err != nil { t.Fatal(err) }
	done := make(chan error, 1)
	go func() { done <- worker.Run(activeContext) }()
	<-entered
	canceledContext, cancelSecond := context.WithCancel(context.Background())
	cancelSecond()
	for _, second := range []context.Context{context.Background(), canceledContext, workerDeadlineContext{}, &workerHostileContext{err: errors.New("private")}, &workerHostileContext{panicErr: true}} {
		if got := worker.Run(second); !errors.Is(got, mail.ErrMailboxSyncWorkerInvalid) { t.Fatalf("concurrent Run(%T) = %v", second, got) }
	}
	close(release)
	if err := <-done; err != nil { t.Fatalf("active Run() = %v", err) }
}

func TestMailboxSyncWorker_StepResultRejectionMatrix(t *testing.T) {
	nonUTC := workerSelected.In(time.FixedZone("other", 3600))
	outOfRange := time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC)
	for _, testCase := range []mail.MailboxSyncStepResult{
		{Outcome: mail.MailboxSyncStepIdle},
		{Outcome: mail.MailboxSyncStepApplied, NextAttemptAt: nonUTC},
		{Outcome: mail.MailboxSyncStepRebuildScheduled, NextAttemptAt: outOfRange},
		{Outcome: mail.MailboxSyncStepRetryScheduled},
		{Outcome: mail.MailboxSyncStepLeaseLost, NextAttemptAt: workerSelected},
		{},
	} {
		t.Run("invalid successful result", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			account := workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))
			repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) { return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{account}}, nil }, fault: func(_ context.Context, request mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) {
				if request.Failure != mail.MailboxSyncWorkerFaultInvalidResult { t.Fatalf("fault failure = %q", request.Failure) }
				return mail.MailboxSyncWorkerFaultResult{}, nil
			}}
			stepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) { return testCase, nil }}
			timer := &workerTimerFake{wait: func(context.Context, time.Duration) error { cancel(); return context.Canceled }}
			worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, workerConfig())
			if err != nil { t.Fatal(err) }
			if err := worker.Run(ctx); err != nil || repository.faultCount() != 1 { t.Fatalf("Run()/faults = %v/%d", err, repository.faultCount()) }
		})
	}
}

func TestMailboxSyncWorker_PortReentryNeverHoldsRunGuard(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	account := workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))
	var worker *mail.MailboxSyncWorker
	reentries := 0
	reenter := func() {
		reentries++
		if err := worker.Run(context.Background()); !errors.Is(err, mail.ErrMailboxSyncWorkerInvalid) { t.Fatalf("reentry Run() = %v", err) }
	}
	repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) {
		reenter()
		return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{account}}, nil
	}, fault: func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) {
		reenter()
		return mail.MailboxSyncWorkerFaultResult{}, nil
	}}
	stepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) { reenter(); return mail.MailboxSyncStepResult{}, errors.New("private") }}
	timer := &workerTimerFake{withTimeout: func(parent context.Context, _ time.Duration) (context.Context, context.CancelFunc) { reenter(); return parent, func() {} }, wait: func(context.Context, time.Duration) error { reenter(); cancel(); return context.Canceled }}
	var err error
	worker, err = mail.NewMailboxSyncWorker(repository, stepper, timer, workerConfig())
	if err != nil { t.Fatal(err) }
	if err := worker.Run(ctx); err != nil || reentries != 5 { t.Fatalf("Run()/reentries = %v/%d", err, reentries) }
}

func TestMailboxSyncWorker_FaultPortFailureContinuesRemainingFIFO(t *testing.T) {
	for _, panics := range []bool{false, true} {
		t.Run(map[bool]string{false: "error", true: "panic"}[panics], func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			first := workerDueAccount(workerAccountA, workerSelected.Add(-2*time.Second))
			second := workerDueAccount(workerAccountB, workerSelected.Add(-time.Second))
			repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) { return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{first, second}}, nil }, fault: func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) {
				if panics { panic("private fault") }
				return mail.MailboxSyncWorkerFaultResult{}, errors.New("private fault")
			}}
			stepper := &workerStepperFake{step: func(_ context.Context, accountID uuid.UUID) (mail.MailboxSyncStepResult, error) {
				if accountID == workerAccountA { return mail.MailboxSyncStepResult{}, errors.New("private step") }
				return mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepIdle, NextAttemptAt: workerSelected}, nil
			}}
			timer := &workerTimerFake{wait: func(context.Context, time.Duration) error { cancel(); return context.Canceled }}
			worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, workerConfig())
			if err != nil { t.Fatal(err) }
			if err := worker.Run(ctx); err != nil || stepper.calls() != 2 || repository.faultCount() != 1 { t.Fatalf("Run()/steps/faults = %v/%d/%d", err, stepper.calls(), repository.faultCount()) }
		})
	}
}

func TestMailboxSyncWorker_HostileScanSliceIsDetachedBeforeStepAndFault(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	accounts := []mail.MailboxSyncWorkerDueAccount{workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))}
	expected := accounts[0]
	repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) {
		return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: accounts}, nil
	}, fault: func(_ context.Context, request mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) {
		if request.Account != expected || !request.SelectedAt.Equal(workerSelected) || request.SelectedAt.Nanosecond() != workerSelected.Nanosecond() {
			t.Fatalf("fault observed caller mutation: %#v", request)
		}
		return mail.MailboxSyncWorkerFaultResult{}, nil
	}}
	stepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) {
		accounts[0].AccountID = workerAccountB
		accounts[0].DueAt = workerSelected
		accounts[0].UpdatedAt = workerSelected
		accounts[0].SyncEpoch = 71
		accounts[0].MutationEpoch = 72
		accounts[0].LeaseEpoch = 73
		accounts[0].SyncAttempts = 74
		return mail.MailboxSyncStepResult{}, errors.New("private step mutation")
	}}
	timer := &workerTimerFake{wait: func(context.Context, time.Duration) error { cancel(); return context.Canceled }}
	worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, workerConfig())
	if err != nil { t.Fatal(err) }
	if err := worker.Run(ctx); err != nil || repository.faultCount() != 1 { t.Fatalf("Run()/faults = %v/%d", err, repository.faultCount()) }
}

func TestMailboxSyncWorker_LeaseEpochRaceNeverOverwritesAmbiguousPriorWork(t *testing.T) {
	state := &workerLeaseRaceState{leaseEpoch: 13}
	accountA := workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))
	accountA.LeaseEpoch = 13
	accountB := accountA
	accountB.LeaseEpoch = 14
	aRepository := &workerLeaseRaceRepository{state: state, account: accountA}
	bRepository := &workerLeaseRaceRepository{state: state, account: accountB}
	aCtx, aCancel := context.WithCancel(context.Background())
	defer aCancel()
	bCtx, bCancel := context.WithCancel(context.Background())
	defer bCancel()
	ambiguous := make(chan struct{}, 1)
	releaseA := make(chan struct{})
	aStepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) {
		state.mu.Lock()
		state.ambiguousApplied = true
		state.mu.Unlock()
		ambiguous <- struct{}{}
		<-releaseA
		return mail.MailboxSyncStepResult{}, errors.New("ambiguous raw result")
	}}
	bStepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) {
		state.mu.Lock()
		state.leaseEpoch = 14
		state.mu.Unlock()
		return mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepApplied, NextAttemptAt: workerSelected}, nil
	}}
	aTimer := &workerTimerFake{wait: func(context.Context, time.Duration) error { aCancel(); return context.Canceled }}
	bTimer := &workerTimerFake{wait: func(context.Context, time.Duration) error { bCancel(); return context.Canceled }}
	workerA, err := mail.NewMailboxSyncWorker(aRepository, aStepper, aTimer, workerConfig())
	if err != nil { t.Fatal(err) }
	workerB, err := mail.NewMailboxSyncWorker(bRepository, bStepper, bTimer, workerConfig())
	if err != nil { t.Fatal(err) }
	doneA := make(chan error, 1)
	go func() { doneA <- workerA.Run(aCtx) }()
	<-ambiguous
	if err := workerB.Run(bCtx); err != nil { t.Fatalf("worker B = %v", err) }
	close(releaseA)
	if err := <-doneA; err != nil { t.Fatalf("worker A = %v", err) }
	state.mu.Lock()
	defer state.mu.Unlock()
	if !state.ambiguousApplied || state.leaseEpoch != 14 || state.successfulReschedules != 0 || len(state.faultRequests) != 1 || state.faultRequests[0].Account != accountA {
		t.Fatalf("race state = %#v", state)
	}
}

func TestMailboxSyncWorker_FullBatchCancellationJoinsOnlyAdmittedFour(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	config := workerConfig()
	config.BatchSize, config.Concurrency = 50, 4
	accounts := workerAccounts(50)
	reached := make(chan struct{}, config.Concurrency)
	release := make(chan struct{})
	var trace struct {
		sync.Mutex
		active, timeoutCalls, stepCalls, cancelCalls, faults, waits int
	}
	repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) {
		return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: accounts}, nil
	}, fault: func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) {
		trace.Lock(); trace.faults++; trace.Unlock()
		return mail.MailboxSyncWorkerFaultResult{}, nil
	}}
	stepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) {
		trace.Lock(); trace.active++; trace.stepCalls++; trace.Unlock()
		reached <- struct{}{}
		<-release
		trace.Lock(); trace.active--; trace.Unlock()
		return mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepIdle, NextAttemptAt: workerSelected}, nil
	}}
	timer := &workerTimerFake{withTimeout: func(parent context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
		if duration != config.StepTimeout { t.Fatalf("timeout duration = %v", duration) }
		trace.Lock(); trace.timeoutCalls++; trace.Unlock()
		child, childCancel := context.WithCancel(parent)
		return child, func() { trace.Lock(); trace.cancelCalls++; trace.Unlock(); childCancel() }
	}, wait: func(context.Context, time.Duration) error {
		trace.Lock(); trace.waits++; trace.Unlock()
		return nil
	}}
	worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, config)
	if err != nil { t.Fatal(err) }
	done := make(chan error, 1)
	go func() { done <- worker.Run(ctx) }()
	for index := 0; index < config.Concurrency; index++ { <-reached }
	cancel()
	close(release)
	if err := <-done; err != nil { t.Fatalf("Run() = %v", err) }
	trace.Lock()
	defer trace.Unlock()
	if trace.active != 0 || trace.timeoutCalls != 4 || trace.stepCalls != 4 || trace.cancelCalls != 4 || trace.faults != 0 || trace.waits != 0 || repository.faultCount() != 0 {
		t.Fatalf("active/timeouts/steps/cancels/faults/waits/repoFaults = %d/%d/%d/%d/%d/%d/%d", trace.active, trace.timeoutCalls, trace.stepCalls, trace.cancelCalls, trace.faults, trace.waits, repository.faultCount())
	}
}

func TestMailboxSyncWorker_ChildCancellationAndHostileStepShapes(t *testing.T) {
	for _, testCase := range []struct {
		name string
		withTimeout func(context.Context, time.Duration) (context.Context, context.CancelFunc)
		result mail.MailboxSyncStepResult
		stepErr error
		wantFailure mail.MailboxSyncWorkerFaultKind
		wantSteps int
		wantCancels int
	}{
		{"normal cancellation once", func(ctx context.Context, _ time.Duration) (context.Context, context.CancelFunc) { return ctx, func() {} }, mail.MailboxSyncStepResult{}, errors.New("raw"), mail.MailboxSyncWorkerFaultUnavailable, 1, 1},
		{"nil child cancellation once", func(context.Context, time.Duration) (context.Context, context.CancelFunc) { return nil, func() {} }, mail.MailboxSyncStepResult{}, nil, mail.MailboxSyncWorkerFaultInvalidResult, 0, 1},
		{"nil child cancel panic", func(context.Context, time.Duration) (context.Context, context.CancelFunc) { return nil, func() { panic("private cancel") } }, mail.MailboxSyncStepResult{}, nil, mail.MailboxSyncWorkerFaultPanic, 0, 1},
		{"child err panic", func(context.Context, time.Duration) (context.Context, context.CancelFunc) { return &workerHostileContext{panicErr: true}, func() {} }, mail.MailboxSyncStepResult{}, nil, mail.MailboxSyncWorkerFaultPanic, 1, 1},
		{"lease lost has next", func(ctx context.Context, _ time.Duration) (context.Context, context.CancelFunc) { return ctx, func() {} }, mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepLeaseLost, NextAttemptAt: workerSelected}, nil, mail.MailboxSyncWorkerFaultInvalidResult, 1, 1},
		{"error plus result discarded unavailable", func(ctx context.Context, _ time.Duration) (context.Context, context.CancelFunc) { return ctx, func() {} }, mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepApplied, NextAttemptAt: workerSelected}, errors.New("raw"), mail.MailboxSyncWorkerFaultUnavailable, 1, 1},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			account := workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))
			cancels := 0
			repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) { return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{account}}, nil }, fault: func(_ context.Context, request mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) {
				if request.Failure != testCase.wantFailure { t.Fatalf("fault = %#v", request) }
				return mail.MailboxSyncWorkerFaultResult{}, nil
			}}
			stepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) { return testCase.result, testCase.stepErr }}
			timer := &workerTimerFake{withTimeout: func(parent context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
				if duration != workerConfig().StepTimeout { t.Fatalf("timeout duration = %v", duration) }
				child, childCancel := testCase.withTimeout(parent, duration)
				return child, func() { cancels++; childCancel() }
			}, wait: func(context.Context, time.Duration) error { cancel(); return context.Canceled }}
			worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, workerConfig())
			if err != nil { t.Fatal(err) }
			if err := worker.Run(ctx); err != nil { t.Fatal(err) }
			if stepper.calls() != testCase.wantSteps || cancels != testCase.wantCancels || repository.faultCount() != 1 {
				t.Fatalf("steps/cancels/faults = %d/%d/%d", stepper.calls(), cancels, repository.faultCount())
			}
		})
	}
}

func TestMailboxSyncWorker_ParentContextFaultsAreContainedAtEveryBoundary(t *testing.T) {
	for _, testCase := range []struct {
		name string
		stage string
		parentErr error
		panics bool
		want error
	}{
		{"scan unknown", "scan", errors.New("private parent"), false, mail.ErrMailboxSyncWorkerUnavailable},
		{"batch unknown", "batch", errors.New("private parent"), false, mail.ErrMailboxSyncWorkerUnavailable},
		{"fault unknown", "fault", errors.New("private parent"), false, mail.ErrMailboxSyncWorkerUnavailable},
		{"wait unknown", "wait", errors.New("private parent"), false, mail.ErrMailboxSyncWorkerUnavailable},
		{"scan panic", "scan", nil, true, mail.ErrMailboxSyncWorkerUnavailable},
		{"batch panic", "batch", nil, true, mail.ErrMailboxSyncWorkerUnavailable},
		{"fault panic", "fault", nil, true, mail.ErrMailboxSyncWorkerUnavailable},
		{"wait panic", "wait", nil, true, mail.ErrMailboxSyncWorkerUnavailable},
		{"scan cancel priority", "scan", context.Canceled, false, nil},
		{"batch deadline priority", "batch", context.DeadlineExceeded, false, nil},
		{"fault cancel priority", "fault", context.Canceled, false, nil},
		{"wait deadline priority", "wait", context.DeadlineExceeded, false, nil},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			parent := &workerMutableContext{}
			account := workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))
			repository := &workerRepositoryFake{scan: func(ctx context.Context, _ mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) {
				if testCase.stage == "scan" { parent.set(testCase.parentErr, testCase.panics) }
				return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{account}}, nil
			}, fault: func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) {
				if testCase.stage == "fault" { parent.set(testCase.parentErr, testCase.panics) }
				return mail.MailboxSyncWorkerFaultResult{}, nil
			}}
			stepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) {
				if testCase.stage == "batch" { parent.set(testCase.parentErr, testCase.panics) }
				return mail.MailboxSyncStepResult{}, errors.New("private step")
			}}
			timer := &workerTimerFake{wait: func(context.Context, time.Duration) error {
				if testCase.stage == "wait" { parent.set(testCase.parentErr, testCase.panics) }
				return nil
			}}
			worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, workerConfig())
			if err != nil { t.Fatal(err) }
			got := worker.Run(parent)
			if testCase.want == nil {
				if got != nil { t.Fatalf("Run() = %v", got) }
			} else if !errors.Is(got, testCase.want) {
				t.Fatalf("Run() = %v; want %v", got, testCase.want)
			}
		})
	}
}

func TestMailboxSyncWorker_MalformedFaultResultsAndLaterLeaseEpochAreContained(t *testing.T) {
	for _, result := range []mail.MailboxSyncWorkerFaultResult{
		{RecordedAt: workerSelected},
		{NextAttemptAt: workerSelected},
		{SyncAttempts: 1},
		{Rescheduled: true},
		{Rescheduled: true, RecordedAt: workerSelected.Add(-time.Nanosecond), NextAttemptAt: workerSelected, SyncAttempts: 1},
		{Rescheduled: true, RecordedAt: workerSelected, NextAttemptAt: workerSelected.Add(2 * time.Second), SyncAttempts: 1},
		{Rescheduled: true, RecordedAt: time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC), NextAttemptAt: workerSelected, SyncAttempts: 1},
		{Rescheduled: true, RecordedAt: time.Date(9999, time.December, 31, 23, 59, 59, 0, time.UTC), NextAttemptAt: time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC), SyncAttempts: 1},
		{Rescheduled: true, RecordedAt: workerSelected, NextAttemptAt: workerSelected.Add(time.Second), SyncAttempts: 2},
	} {
		t.Run("hostile fault result", func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			account := workerDueAccount(workerAccountA, workerSelected.Add(-time.Second))
			repository := &workerRepositoryFake{scan: func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) { return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{account}}, nil }, fault: func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) { return result, nil }}
			stepper := &workerStepperFake{step: func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error) { return mail.MailboxSyncStepResult{}, errors.New("private") }}
			timer := &workerTimerFake{wait: func(context.Context, time.Duration) error { cancel(); return context.Canceled }}
			worker, err := mail.NewMailboxSyncWorker(repository, stepper, timer, workerConfig())
			if err != nil { t.Fatal(err) }
			if err := worker.Run(ctx); err != nil || repository.faultCount() != 1 { t.Fatalf("Run()/faults = %v/%d", err, repository.faultCount()) }
		})
	}
}

func requireWorkerFields(t *testing.T, actual reflect.Type, names []string, types []reflect.Type) {
	t.Helper()
	if actual.NumField() != len(names) || len(types) != len(names) { t.Fatalf("fields for %v malformed", actual) }
	for index, name := range names {
		field := actual.Field(index)
		if field.Name != name || field.Type != types[index] { t.Fatalf("field %d = %s %v; want %s %v", index, field.Name, field.Type, name, types[index]) }
	}
}

func requireWorkerMethod(t *testing.T, actual reflect.Type, name string, expected reflect.Type) {
	t.Helper()
	method, ok := actual.MethodByName(name)
	if !ok || method.Type != expected { t.Fatalf("method %v = %v; want %v", name, method.Type, expected) }
}

func workerConfig() mail.MailboxSyncWorkerConfig {
	return mail.MailboxSyncWorkerConfig{Enabled: true, ScanInterval: time.Minute, StepTimeout: time.Second, RetryMin: time.Second, RetryMax: time.Minute, BatchSize: 2, Concurrency: 1}
}

func mutateWorkerConfig(mutate func(*mail.MailboxSyncWorkerConfig)) mail.MailboxSyncWorkerConfig {
	value := workerConfig()
	mutate(&value)
	return value
}

func workerDueAccount(id uuid.UUID, dueAt time.Time) mail.MailboxSyncWorkerDueAccount {
	return mail.MailboxSyncWorkerDueAccount{AccountID: id, DueAt: dueAt, UpdatedAt: dueAt, SyncEpoch: 1, MutationEpoch: 2, LeaseEpoch: 3}
}

func workerAccounts(count int) []mail.MailboxSyncWorkerDueAccount {
	values := make([]mail.MailboxSyncWorkerDueAccount, count)
	for index := range values {
		values[index] = workerDueAccount(uuid.NewSHA1(uuid.NameSpaceURL, []byte{byte(index), byte(index >> 8)}), workerSelected.Add(time.Duration(index-count) * time.Second))
	}
	return values
}

type workerRepositoryFake struct {
	mu sync.Mutex
	scan func(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error)
	fault func(context.Context, mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error)
	requests []mail.MailboxSyncWorkerScanRequest
	faultRequests []mail.MailboxSyncWorkerFaultRequest
}

func (fake *workerRepositoryFake) ScanDueMailboxSyncAccounts(ctx context.Context, request mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) {
	fake.mu.Lock()
	fake.requests = append(fake.requests, request)
	callback := fake.scan
	fake.mu.Unlock()
	if callback == nil { return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{}}, nil }
	return callback(ctx, request)
}

func (fake *workerRepositoryFake) RescheduleMailboxSyncWorkerFault(ctx context.Context, request mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) {
	fake.mu.Lock()
	fake.faultRequests = append(fake.faultRequests, request)
	callback := fake.fault
	fake.mu.Unlock()
	if callback == nil { return mail.MailboxSyncWorkerFaultResult{}, nil }
	return callback(ctx, request)
}

func (fake *workerRepositoryFake) scanRequests() []mail.MailboxSyncWorkerScanRequest { fake.mu.Lock(); defer fake.mu.Unlock(); return append([]mail.MailboxSyncWorkerScanRequest(nil), fake.requests...) }
func (fake *workerRepositoryFake) faultCount() int { fake.mu.Lock(); defer fake.mu.Unlock(); return len(fake.faultRequests) }

type workerLeaseRaceState struct {
	mu sync.Mutex
	leaseEpoch int64
	ambiguousApplied bool
	successfulReschedules int
	faultRequests []mail.MailboxSyncWorkerFaultRequest
}

type workerLeaseRaceRepository struct { state *workerLeaseRaceState; account mail.MailboxSyncWorkerDueAccount }
func (repository *workerLeaseRaceRepository) ScanDueMailboxSyncAccounts(context.Context, mail.MailboxSyncWorkerScanRequest) (mail.MailboxSyncWorkerScanResult, error) {
	return mail.MailboxSyncWorkerScanResult{SelectedAt: workerSelected, Accounts: []mail.MailboxSyncWorkerDueAccount{repository.account}}, nil
}
func (repository *workerLeaseRaceRepository) RescheduleMailboxSyncWorkerFault(_ context.Context, request mail.MailboxSyncWorkerFaultRequest) (mail.MailboxSyncWorkerFaultResult, error) {
	repository.state.mu.Lock()
	defer repository.state.mu.Unlock()
	repository.state.faultRequests = append(repository.state.faultRequests, request)
	if request.Account != repository.account || request.Account.LeaseEpoch != repository.state.leaseEpoch { return mail.MailboxSyncWorkerFaultResult{}, nil }
	repository.state.successfulReschedules++
	return mail.MailboxSyncWorkerFaultResult{Rescheduled: true, RecordedAt: workerSelected, NextAttemptAt: workerSelected.Add(request.RetryDelay), SyncAttempts: request.Account.SyncAttempts + 1}, nil
}

type workerDeadlineContext struct{}
func (workerDeadlineContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (workerDeadlineContext) Done() <-chan struct{} { return nil }
func (workerDeadlineContext) Err() error { return context.DeadlineExceeded }
func (workerDeadlineContext) Value(any) any { return nil }

type workerHostileContext struct { err error; panicErr bool }
func (*workerHostileContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (*workerHostileContext) Done() <-chan struct{} { return nil }
func (value *workerHostileContext) Err() error { if value.panicErr { panic("private context value") }; return value.err }
func (*workerHostileContext) Value(any) any { return nil }

type workerMutableContext struct { mu sync.Mutex; err error; panicErr bool }
func (*workerMutableContext) Deadline() (time.Time, bool) { return time.Time{}, false }
func (*workerMutableContext) Done() <-chan struct{} { return nil }
func (value *workerMutableContext) Err() error { value.mu.Lock(); defer value.mu.Unlock(); if value.panicErr { panic("private mutable context") }; return value.err }
func (*workerMutableContext) Value(any) any { return nil }
func (value *workerMutableContext) set(err error, panicErr bool) { value.mu.Lock(); defer value.mu.Unlock(); value.err, value.panicErr = err, panicErr }

type workerStepperFake struct { mu sync.Mutex; step func(context.Context, uuid.UUID) (mail.MailboxSyncStepResult, error); ids []uuid.UUID }
func (fake *workerStepperFake) Step(ctx context.Context, id uuid.UUID) (mail.MailboxSyncStepResult, error) { fake.mu.Lock(); fake.ids = append(fake.ids, id); callback := fake.step; fake.mu.Unlock(); if callback == nil { return mail.MailboxSyncStepResult{Outcome: mail.MailboxSyncStepIdle, NextAttemptAt: workerSelected.Add(time.Second)}, nil }; return callback(ctx, id) }
func (fake *workerStepperFake) calls() int { fake.mu.Lock(); defer fake.mu.Unlock(); return len(fake.ids) }

type workerTimerFake struct { mu sync.Mutex; wait func(context.Context, time.Duration) error; withTimeout func(context.Context, time.Duration) (context.Context, context.CancelFunc); waits int }
func (fake *workerTimerFake) Wait(ctx context.Context, duration time.Duration) error { fake.mu.Lock(); fake.waits++; callback := fake.wait; fake.mu.Unlock(); if callback == nil { return context.Canceled }; return callback(ctx, duration) }
func (fake *workerTimerFake) WithTimeout(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) { fake.mu.Lock(); callback := fake.withTimeout; fake.mu.Unlock(); if callback == nil { return context.WithCancel(ctx) }; return callback(ctx, duration) }
func (fake *workerTimerFake) waitCount() int { fake.mu.Lock(); defer fake.mu.Unlock(); return fake.waits }
