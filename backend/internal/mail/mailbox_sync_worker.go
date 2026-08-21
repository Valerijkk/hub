package mail

import (
	"bytes"
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"
)

type MailboxSyncWorkerDueAccount struct {
	AccountID     uuid.UUID
	DueAt         time.Time
	UpdatedAt     time.Time
	SyncEpoch     int64
	MutationEpoch int64
	LeaseEpoch    int64
	SyncAttempts  int32
}

type MailboxSyncWorkerScanRequest struct {
	Limit int
}

type MailboxSyncWorkerScanResult struct {
	SelectedAt time.Time
	Accounts   []MailboxSyncWorkerDueAccount
}

type MailboxSyncWorkerFaultKind string

const (
	MailboxSyncWorkerFaultTimeout       MailboxSyncWorkerFaultKind = "timeout"
	MailboxSyncWorkerFaultPanic         MailboxSyncWorkerFaultKind = "panic"
	MailboxSyncWorkerFaultUnavailable   MailboxSyncWorkerFaultKind = "unavailable"
	MailboxSyncWorkerFaultInvalidResult MailboxSyncWorkerFaultKind = "invalid_result"
)

type MailboxSyncWorkerFaultRequest struct {
	Account    MailboxSyncWorkerDueAccount
	SelectedAt time.Time
	Failure    MailboxSyncWorkerFaultKind
	RetryDelay time.Duration
}

type MailboxSyncWorkerFaultResult struct {
	Rescheduled   bool
	RecordedAt    time.Time
	NextAttemptAt time.Time
	SyncAttempts  int32
}

type MailboxSyncWorkerRepository interface {
	ScanDueMailboxSyncAccounts(context.Context, MailboxSyncWorkerScanRequest) (MailboxSyncWorkerScanResult, error)
	RescheduleMailboxSyncWorkerFault(context.Context, MailboxSyncWorkerFaultRequest) (MailboxSyncWorkerFaultResult, error)
}

type MailboxSyncWorkerStepper interface {
	Step(context.Context, uuid.UUID) (MailboxSyncStepResult, error)
}

type MailboxSyncWorkerTimer interface {
	Wait(context.Context, time.Duration) error
	WithTimeout(context.Context, time.Duration) (context.Context, context.CancelFunc)
}

type MailboxSyncWorkerConfig struct {
	Enabled      bool
	ScanInterval time.Duration
	StepTimeout  time.Duration
	RetryMin     time.Duration
	RetryMax     time.Duration
	BatchSize    int
	Concurrency  int
}

type MailboxSyncWorkerErrorKind string

const (
	MailboxSyncWorkerErrorInvalid     MailboxSyncWorkerErrorKind = "invalid"
	MailboxSyncWorkerErrorTooLarge    MailboxSyncWorkerErrorKind = "too_large"
	MailboxSyncWorkerErrorUnavailable MailboxSyncWorkerErrorKind = "unavailable"
)

var (
	ErrMailboxSyncWorkerInvalid     = errors.New("mailbox sync worker invalid")
	ErrMailboxSyncWorkerTooLarge    = errors.New("mailbox sync worker exceeds bounds")
	ErrMailboxSyncWorkerUnavailable = errors.New("mailbox sync worker unavailable")
)

type MailboxSyncWorkerError struct {
	Kind MailboxSyncWorkerErrorKind
}

func (errorValue *MailboxSyncWorkerError) Error() string {
	if errorValue == nil {
		return "mailbox sync worker error"
	}
	switch errorValue.Kind {
	case MailboxSyncWorkerErrorInvalid:
		return ErrMailboxSyncWorkerInvalid.Error()
	case MailboxSyncWorkerErrorTooLarge:
		return ErrMailboxSyncWorkerTooLarge.Error()
	case MailboxSyncWorkerErrorUnavailable:
		return ErrMailboxSyncWorkerUnavailable.Error()
	default:
		return "mailbox sync worker error"
	}
}

func (errorValue *MailboxSyncWorkerError) Unwrap() error {
	if errorValue == nil {
		return nil
	}
	switch errorValue.Kind {
	case MailboxSyncWorkerErrorInvalid:
		return ErrMailboxSyncWorkerInvalid
	case MailboxSyncWorkerErrorTooLarge:
		return ErrMailboxSyncWorkerTooLarge
	case MailboxSyncWorkerErrorUnavailable:
		return ErrMailboxSyncWorkerUnavailable
	default:
		return nil
	}
}

type MailboxSyncWorker struct {
	repository MailboxSyncWorkerRepository
	stepper    MailboxSyncWorkerStepper
	timer      MailboxSyncWorkerTimer
	config     MailboxSyncWorkerConfig

	guard   sync.Mutex
	running bool
}

type mailboxSyncWorkerContextState uint8

const (
	mailboxSyncWorkerContextActive mailboxSyncWorkerContextState = iota + 1
	mailboxSyncWorkerContextCancelled
	mailboxSyncWorkerContextUnavailable
)

type mailboxSyncWorkerBatchResult struct {
	tooLarge    bool
	unavailable bool
}

type mailboxSyncWorkerAccountResult struct {
	tooLarge    bool
	unavailable bool
}

func NewMailboxSyncWorker(repository MailboxSyncWorkerRepository, stepper MailboxSyncWorkerStepper, timer MailboxSyncWorkerTimer, config MailboxSyncWorkerConfig) (*MailboxSyncWorker, error) {
	if !config.Enabled {
		if repository != nil || stepper != nil || timer != nil || config != (MailboxSyncWorkerConfig{}) {
			return nil, mailboxSyncWorkerError(MailboxSyncWorkerErrorInvalid)
		}
		return &MailboxSyncWorker{config: config}, nil
	}
	if mailboxSyncNilDependency(repository) || mailboxSyncNilDependency(stepper) || mailboxSyncNilDependency(timer) {
		return nil, mailboxSyncWorkerError(MailboxSyncWorkerErrorInvalid)
	}
	if kind := mailboxSyncWorkerConfigKind(config); kind != "" {
		return nil, mailboxSyncWorkerError(kind)
	}
	return &MailboxSyncWorker{repository: repository, stepper: stepper, timer: timer, config: config}, nil
}

func (worker *MailboxSyncWorker) Run(ctx context.Context) error {
	if worker == nil || mailboxSyncNilDependency(ctx) {
		return mailboxSyncWorkerError(MailboxSyncWorkerErrorInvalid)
	}
	worker.guard.Lock()
	if worker.running {
		worker.guard.Unlock()
		return mailboxSyncWorkerError(MailboxSyncWorkerErrorInvalid)
	}
	worker.running = true
	worker.guard.Unlock()
	defer func() {
		worker.guard.Lock()
		worker.running = false
		worker.guard.Unlock()
	}()
	switch mailboxSyncWorkerParentState(ctx) {
	case mailboxSyncWorkerContextCancelled:
		return nil
	case mailboxSyncWorkerContextUnavailable:
		return mailboxSyncWorkerError(MailboxSyncWorkerErrorUnavailable)
	}
	if !worker.config.Enabled {
		return nil
	}
	return worker.run(ctx)
}

func (worker *MailboxSyncWorker) run(ctx context.Context) error {
	for {
		switch mailboxSyncWorkerParentState(ctx) {
		case mailboxSyncWorkerContextCancelled:
			return nil
		case mailboxSyncWorkerContextUnavailable:
			return mailboxSyncWorkerError(MailboxSyncWorkerErrorUnavailable)
		}
		result, scanErr, panicked := worker.safeScan(ctx)
		switch mailboxSyncWorkerParentState(ctx) {
		case mailboxSyncWorkerContextCancelled:
			return nil
		case mailboxSyncWorkerContextUnavailable:
			return mailboxSyncWorkerError(MailboxSyncWorkerErrorUnavailable)
		}
		if panicked || scanErr != nil {
			if err := worker.wait(ctx, worker.config.RetryMin); err != nil {
				return err
			}
			continue
		}
		accounts, err := worker.validateScan(result)
		if err != nil {
			return err
		}
		if len(accounts) != 0 {
			batch := worker.runBatch(ctx, accounts, result.SelectedAt)
			switch mailboxSyncWorkerParentState(ctx) {
			case mailboxSyncWorkerContextCancelled:
				return nil
			case mailboxSyncWorkerContextUnavailable:
				return mailboxSyncWorkerError(MailboxSyncWorkerErrorUnavailable)
			}
			if batch.unavailable {
				return mailboxSyncWorkerError(MailboxSyncWorkerErrorUnavailable)
			}
			if batch.tooLarge {
				return mailboxSyncWorkerError(MailboxSyncWorkerErrorTooLarge)
			}
		}
		if err := worker.wait(ctx, worker.config.ScanInterval); err != nil {
			return err
		}
	}
}

func (worker *MailboxSyncWorker) safeScan(ctx context.Context) (result MailboxSyncWorkerScanResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result = MailboxSyncWorkerScanResult{}
			resultErr = nil
			panicked = true
		}
	}()
	result, resultErr = worker.repository.ScanDueMailboxSyncAccounts(ctx, MailboxSyncWorkerScanRequest{Limit: worker.config.BatchSize})
	return result, resultErr, false
}

func (worker *MailboxSyncWorker) wait(ctx context.Context, duration time.Duration) (resultErr error) {
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		resultErr = worker.timer.Wait(ctx, duration)
	}()
	switch mailboxSyncWorkerParentState(ctx) {
	case mailboxSyncWorkerContextCancelled:
		return nil
	case mailboxSyncWorkerContextUnavailable:
		return mailboxSyncWorkerError(MailboxSyncWorkerErrorUnavailable)
	}
	if panicked || resultErr != nil {
		return mailboxSyncWorkerError(MailboxSyncWorkerErrorUnavailable)
	}
	return nil
}

func (worker *MailboxSyncWorker) validateScan(result MailboxSyncWorkerScanResult) ([]MailboxSyncWorkerDueAccount, error) {
	if !syncValidTime(result.SelectedAt) || result.Accounts == nil {
		return nil, mailboxSyncWorkerError(MailboxSyncWorkerErrorInvalid)
	}
	if len(result.Accounts) > worker.config.BatchSize {
		return nil, mailboxSyncWorkerError(MailboxSyncWorkerErrorTooLarge)
	}
	seen := make(map[uuid.UUID]struct{}, len(result.Accounts))
	var previous MailboxSyncWorkerDueAccount
	for index := range result.Accounts {
		account := result.Accounts[index]
		if account.AccountID == uuid.Nil || !syncValidTime(account.DueAt) || !syncValidTime(account.UpdatedAt) || account.DueAt.After(result.SelectedAt) || account.UpdatedAt.After(result.SelectedAt) || account.SyncEpoch < 0 || account.MutationEpoch < 0 || account.LeaseEpoch <= 0 || account.SyncAttempts < 0 {
			return nil, mailboxSyncWorkerError(MailboxSyncWorkerErrorInvalid)
		}
		if _, duplicate := seen[account.AccountID]; duplicate {
			return nil, mailboxSyncWorkerError(MailboxSyncWorkerErrorInvalid)
		}
		seen[account.AccountID] = struct{}{}
		if index != 0 && (account.DueAt.Before(previous.DueAt) || (account.DueAt.Equal(previous.DueAt) && bytes.Compare(account.AccountID[:], previous.AccountID[:]) <= 0)) {
			return nil, mailboxSyncWorkerError(MailboxSyncWorkerErrorInvalid)
		}
		previous = account
	}
	return append([]MailboxSyncWorkerDueAccount(nil), result.Accounts...), nil
}

func (worker *MailboxSyncWorker) runBatch(ctx context.Context, accounts []MailboxSyncWorkerDueAccount, selectedAt time.Time) mailboxSyncWorkerBatchResult {
	jobs := make(chan MailboxSyncWorkerDueAccount, len(accounts))
	for _, account := range accounts {
		jobs <- account
	}
	close(jobs)

	var batch struct {
		sync.Mutex
		result mailboxSyncWorkerBatchResult
	}
	var group sync.WaitGroup
	for index := 0; index < worker.config.Concurrency; index++ {
		group.Add(1)
		go func() {
			defer func() {
				if recover() != nil {
					batch.Lock()
					batch.result.unavailable = true
					batch.Unlock()
				}
				group.Done()
			}()
			for account := range jobs {
				switch mailboxSyncWorkerParentState(ctx) {
				case mailboxSyncWorkerContextCancelled:
					continue
				case mailboxSyncWorkerContextUnavailable:
					batch.Lock()
					batch.result.unavailable = true
					batch.Unlock()
					continue
				}
				result := worker.processAccount(ctx, account, selectedAt)
				if result.tooLarge || result.unavailable {
					batch.Lock()
					batch.result.tooLarge = batch.result.tooLarge || result.tooLarge
					batch.result.unavailable = batch.result.unavailable || result.unavailable
					batch.Unlock()
				}
			}
		}()
	}
	group.Wait()
	batch.Lock()
	defer batch.Unlock()
	return batch.result
}

func (worker *MailboxSyncWorker) processAccount(ctx context.Context, account MailboxSyncWorkerDueAccount, selectedAt time.Time) mailboxSyncWorkerAccountResult {
	child, cancel, timeoutPanicked := worker.withTimeout(ctx)
	if timeoutPanicked {
		return worker.recordFault(ctx, account, selectedAt, MailboxSyncWorkerFaultPanic)
	}
	if mailboxSyncNilDependency(child) || cancel == nil {
		cancelPanicked := false
		if cancel != nil {
			cancelPanicked = mailboxSyncWorkerCancel(cancel)
		}
		switch mailboxSyncWorkerParentState(ctx) {
		case mailboxSyncWorkerContextCancelled:
			return mailboxSyncWorkerAccountResult{}
		case mailboxSyncWorkerContextUnavailable:
			return mailboxSyncWorkerAccountResult{unavailable: true}
		}
		if cancelPanicked {
			return worker.recordFault(ctx, account, selectedAt, MailboxSyncWorkerFaultPanic)
		}
		return worker.recordFault(ctx, account, selectedAt, MailboxSyncWorkerFaultInvalidResult)
	}
	result, stepErr, stepPanicked := worker.step(child, account.AccountID)
	childErr, childErrPanicked := mailboxSyncWorkerChildErr(child)
	cancelPanicked := mailboxSyncWorkerCancel(cancel)
	switch mailboxSyncWorkerParentState(ctx) {
	case mailboxSyncWorkerContextCancelled:
		return mailboxSyncWorkerAccountResult{}
	case mailboxSyncWorkerContextUnavailable:
		return mailboxSyncWorkerAccountResult{unavailable: true}
	}
	if stepPanicked || childErrPanicked || cancelPanicked {
		return worker.recordFault(ctx, account, selectedAt, MailboxSyncWorkerFaultPanic)
	}
	if childErr == context.DeadlineExceeded {
		return worker.recordFault(ctx, account, selectedAt, MailboxSyncWorkerFaultTimeout)
	}
	if childErr != nil {
		return worker.recordFault(ctx, account, selectedAt, MailboxSyncWorkerFaultInvalidResult)
	}
	if stepErr != nil {
		return worker.recordFault(ctx, account, selectedAt, MailboxSyncWorkerFaultUnavailable)
	}
	if !mailboxSyncWorkerValidStep(result) {
		return worker.recordFault(ctx, account, selectedAt, MailboxSyncWorkerFaultInvalidResult)
	}
	return mailboxSyncWorkerAccountResult{}
}

func (worker *MailboxSyncWorker) withTimeout(ctx context.Context) (child context.Context, cancel context.CancelFunc, panicked bool) {
	defer func() {
		if recover() != nil {
			child = nil
			cancel = nil
			panicked = true
		}
	}()
	child, cancel = worker.timer.WithTimeout(ctx, worker.config.StepTimeout)
	return child, cancel, false
}

func (worker *MailboxSyncWorker) step(ctx context.Context, accountID uuid.UUID) (result MailboxSyncStepResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result = MailboxSyncStepResult{}
			resultErr = nil
			panicked = true
		}
	}()
	result, resultErr = worker.stepper.Step(ctx, accountID)
	return result, resultErr, false
}

func mailboxSyncWorkerCancel(cancel context.CancelFunc) (panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	cancel()
	return false
}

func mailboxSyncWorkerChildErr(ctx context.Context) (resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			resultErr = nil
			panicked = true
		}
	}()
	return ctx.Err(), false
}

func (worker *MailboxSyncWorker) recordFault(ctx context.Context, account MailboxSyncWorkerDueAccount, selectedAt time.Time, failure MailboxSyncWorkerFaultKind) mailboxSyncWorkerAccountResult {
	switch mailboxSyncWorkerParentState(ctx) {
	case mailboxSyncWorkerContextCancelled:
		return mailboxSyncWorkerAccountResult{}
	case mailboxSyncWorkerContextUnavailable:
		return mailboxSyncWorkerAccountResult{unavailable: true}
	}
	if account.SyncAttempts == mailboxSyncWorkerMaxAttempts {
		return mailboxSyncWorkerAccountResult{tooLarge: true}
	}
	delay := worker.retryDelay(account.SyncAttempts)
	request := MailboxSyncWorkerFaultRequest{Account: account, SelectedAt: selectedAt, Failure: failure, RetryDelay: delay}
	result, resultErr, panicked := worker.fault(ctx, request)
	switch mailboxSyncWorkerParentState(ctx) {
	case mailboxSyncWorkerContextCancelled:
		return mailboxSyncWorkerAccountResult{}
	case mailboxSyncWorkerContextUnavailable:
		return mailboxSyncWorkerAccountResult{unavailable: true}
	}
	if panicked || resultErr != nil {
		return mailboxSyncWorkerAccountResult{}
	}
	if !mailboxSyncWorkerValidFaultResult(result, request) {
		return mailboxSyncWorkerAccountResult{}
	}
	return mailboxSyncWorkerAccountResult{}
}

func (worker *MailboxSyncWorker) fault(ctx context.Context, request MailboxSyncWorkerFaultRequest) (result MailboxSyncWorkerFaultResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result = MailboxSyncWorkerFaultResult{}
			resultErr = nil
			panicked = true
		}
	}()
	result, resultErr = worker.repository.RescheduleMailboxSyncWorkerFault(ctx, request)
	return result, resultErr, false
}

func (worker *MailboxSyncWorker) retryDelay(attempts int32) time.Duration {
	if attempts <= 0 || attempts >= 63 {
		if attempts <= 0 {
			return worker.config.RetryMin
		}
		return worker.config.RetryMax
	}
	maximum := time.Duration(1<<63 - 1)
	shift := uint(attempts)
	if worker.config.RetryMin > maximum>>shift {
		return worker.config.RetryMax
	}
	delay := worker.config.RetryMin << shift
	if delay > worker.config.RetryMax {
		return worker.config.RetryMax
	}
	return delay
}

func mailboxSyncWorkerConfigKind(config MailboxSyncWorkerConfig) MailboxSyncWorkerErrorKind {
	if config.BatchSize > 50 || config.Concurrency > 4 || config.ScanInterval > 5*time.Minute || config.StepTimeout > 30*time.Second || config.RetryMin > 5*time.Minute || config.RetryMax > 5*time.Minute {
		return MailboxSyncWorkerErrorTooLarge
	}
	if config.BatchSize < 1 || config.Concurrency < 1 || config.ScanInterval < time.Second || config.StepTimeout < time.Second || config.RetryMin < time.Second || config.RetryMax < time.Second || config.RetryMin > config.RetryMax {
		return MailboxSyncWorkerErrorInvalid
	}
	return ""
}

func mailboxSyncWorkerValidStep(result MailboxSyncStepResult) bool {
	switch result.Outcome {
	case MailboxSyncStepIdle, MailboxSyncStepApplied, MailboxSyncStepRebuildScheduled, MailboxSyncStepRetryScheduled:
		return syncValidTime(result.NextAttemptAt)
	case MailboxSyncStepLeaseLost:
		return result.NextAttemptAt.IsZero()
	default:
		return false
	}
}

func mailboxSyncWorkerValidFaultResult(result MailboxSyncWorkerFaultResult, request MailboxSyncWorkerFaultRequest) bool {
	if !result.Rescheduled {
		return result == (MailboxSyncWorkerFaultResult{})
	}
	if !syncValidTime(request.SelectedAt) || !syncValidTime(result.RecordedAt) || !syncValidTime(result.NextAttemptAt) || result.RecordedAt.Before(request.SelectedAt) || result.SyncAttempts != request.Account.SyncAttempts+1 {
		return false
	}
	expected, ok := mailboxSyncCheckedTimeAdd(result.RecordedAt, request.RetryDelay)
	return ok && result.NextAttemptAt.Equal(expected)
}

func mailboxSyncWorkerParentState(ctx context.Context) (state mailboxSyncWorkerContextState) {
	defer func() {
		if recover() != nil {
			state = mailboxSyncWorkerContextUnavailable
		}
	}()
	switch ctx.Err() {
	case nil:
		return mailboxSyncWorkerContextActive
	case context.Canceled, context.DeadlineExceeded:
		return mailboxSyncWorkerContextCancelled
	default:
		return mailboxSyncWorkerContextUnavailable
	}
}

func mailboxSyncWorkerError(kind MailboxSyncWorkerErrorKind) error {
	return &MailboxSyncWorkerError{Kind: kind}
}

const mailboxSyncWorkerMaxAttempts int32 = 1<<31 - 1
