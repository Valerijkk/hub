package mail

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"
)

const (
	mailboxSubmissionWorkerDefaultScanInterval    = time.Minute
	mailboxSubmissionWorkerDefaultDispatchTimeout = 30 * time.Second
	mailboxSubmissionWorkerDefaultBatchSize       = 20
	mailboxSubmissionWorkerMaximumScanInterval    = 5 * time.Minute
	mailboxSubmissionWorkerMaximumDispatchTimeout = 30 * time.Second
	mailboxSubmissionWorkerMaximumBatchSize       = 50
)

// MailboxSubmissionWorkerDueScanRequest limits one durable queued-due scan.
// Implementations select only `queued` submissions due through the durable
// `mail_submissions_queued_due_idx (next_attempt_at,id)` authority.
type MailboxSubmissionWorkerDueScanRequest struct {
	Limit int
}

// MailboxSubmissionWorkerDueScanResult returns durable submission identities
// only. It contains no draft, provider, recipient, or message data.
type MailboxSubmissionWorkerDueScanResult struct {
	SubmissionIDs []uuid.UUID
}

// MailboxSubmissionWorkerDueScanner is the durable queued-authority boundary.
// It does not claim, authorize, resolve, or mutate a submission.
type MailboxSubmissionWorkerDueScanner interface {
	ScanDueMailboxSubmissionIDs(context.Context, MailboxSubmissionWorkerDueScanRequest) (MailboxSubmissionWorkerDueScanResult, error)
}

// MailboxSubmissionWorkerDispatcher invokes the sealed dispatcher for one
// durable submission identity. The worker never obtains provider capabilities.
type MailboxSubmissionWorkerDispatcher interface {
	Dispatch(context.Context, uuid.UUID) error
}

// MailboxSubmissionWorkerTimer supplies bounded waits and dispatch deadlines.
type MailboxSubmissionWorkerTimer interface {
	Wait(context.Context, time.Duration) error
	WithTimeout(context.Context, time.Duration) (context.Context, context.CancelFunc)
}

// MailboxSubmissionWorkerOptions keeps a single sequential worker bounded.
// Zero duration and batch values on an enabled worker use conservative defaults.
type MailboxSubmissionWorkerOptions struct {
	Enabled         bool
	ScanInterval    time.Duration
	DispatchTimeout time.Duration
	BatchSize       int
}

// DefaultMailboxSubmissionWorkerOptions returns conservative enabled options.
func DefaultMailboxSubmissionWorkerOptions() MailboxSubmissionWorkerOptions {
	return MailboxSubmissionWorkerOptions{
		Enabled:         true,
		ScanInterval:    mailboxSubmissionWorkerDefaultScanInterval,
		DispatchTimeout: mailboxSubmissionWorkerDefaultDispatchTimeout,
		BatchSize:       mailboxSubmissionWorkerDefaultBatchSize,
	}
}

// MailboxSubmissionWorkerRunResult is a bounded, content-free RunOnce summary.
type MailboxSubmissionWorkerRunResult struct {
	Due                    int
	Dispatched             int
	Failed                 int
	ReconciliationRequired int
}

// MailboxSubmissionWorkerErrorKind is the closed worker failure family.
type MailboxSubmissionWorkerErrorKind string

const (
	MailboxSubmissionWorkerErrorInvalid        MailboxSubmissionWorkerErrorKind = "invalid"
	MailboxSubmissionWorkerErrorTooLarge       MailboxSubmissionWorkerErrorKind = "too_large"
	MailboxSubmissionWorkerErrorUnavailable    MailboxSubmissionWorkerErrorKind = "unavailable"
	MailboxSubmissionWorkerErrorPartialFailure MailboxSubmissionWorkerErrorKind = "partial_failure"
)

var (
	ErrMailboxSubmissionWorkerInvalid        = errors.New("mailbox submission worker invalid")
	ErrMailboxSubmissionWorkerTooLarge       = errors.New("mailbox submission worker exceeds bounds")
	ErrMailboxSubmissionWorkerUnavailable    = errors.New("mailbox submission worker unavailable")
	ErrMailboxSubmissionWorkerPartialFailure = errors.New("mailbox submission worker partial failure")
)

// MailboxSubmissionWorkerError retains only a closed, content-free kind.
type MailboxSubmissionWorkerError struct {
	kind MailboxSubmissionWorkerErrorKind
}

func (errorValue *MailboxSubmissionWorkerError) Error() string {
	switch errorValue.Kind() {
	case MailboxSubmissionWorkerErrorInvalid:
		return ErrMailboxSubmissionWorkerInvalid.Error()
	case MailboxSubmissionWorkerErrorTooLarge:
		return ErrMailboxSubmissionWorkerTooLarge.Error()
	case MailboxSubmissionWorkerErrorUnavailable:
		return ErrMailboxSubmissionWorkerUnavailable.Error()
	case MailboxSubmissionWorkerErrorPartialFailure:
		return ErrMailboxSubmissionWorkerPartialFailure.Error()
	default:
		return "mailbox submission worker error"
	}
}

func (errorValue *MailboxSubmissionWorkerError) Unwrap() error {
	switch errorValue.Kind() {
	case MailboxSubmissionWorkerErrorInvalid:
		return ErrMailboxSubmissionWorkerInvalid
	case MailboxSubmissionWorkerErrorTooLarge:
		return ErrMailboxSubmissionWorkerTooLarge
	case MailboxSubmissionWorkerErrorUnavailable:
		return ErrMailboxSubmissionWorkerUnavailable
	case MailboxSubmissionWorkerErrorPartialFailure:
		return ErrMailboxSubmissionWorkerPartialFailure
	default:
		return nil
	}
}

// Kind returns a safe error classification.
func (errorValue *MailboxSubmissionWorkerError) Kind() MailboxSubmissionWorkerErrorKind {
	if errorValue == nil || !mailboxSubmissionWorkerErrorKindValid(errorValue.kind) {
		return ""
	}
	return errorValue.kind
}

// MailboxSubmissionWorker scans due durable IDs and sequentially delegates
// each one to the sealed dispatcher. It owns neither persistent state nor
// provider, engine, transport, or NATS capabilities.
type MailboxSubmissionWorker struct {
	scanner    MailboxSubmissionWorkerDueScanner
	dispatcher MailboxSubmissionWorkerDispatcher
	timer      MailboxSubmissionWorkerTimer
	options    MailboxSubmissionWorkerOptions

	guard   sync.Mutex
	running bool
}

type mailboxSubmissionWorkerRuntimeTimer struct{}

type mailboxSubmissionWorkerContextState uint8

const (
	mailboxSubmissionWorkerContextActive mailboxSubmissionWorkerContextState = iota + 1
	mailboxSubmissionWorkerContextCancelled
	mailboxSubmissionWorkerContextUnavailable
)

var _ MailboxSubmissionWorkerTimer = mailboxSubmissionWorkerRuntimeTimer{}

// NewMailboxSubmissionWorker constructs a disabled zero worker or a bounded
// enabled worker. Construction performs no scanner, dispatcher, or provider I/O.
func NewMailboxSubmissionWorker(scanner MailboxSubmissionWorkerDueScanner, dispatcher MailboxSubmissionWorkerDispatcher, timer MailboxSubmissionWorkerTimer, options MailboxSubmissionWorkerOptions) (*MailboxSubmissionWorker, error) {
	if !options.Enabled {
		if options != (MailboxSubmissionWorkerOptions{}) || !mailboxSubmissionWorkerNilDependency(scanner) || !mailboxSubmissionWorkerNilDependency(dispatcher) || !mailboxSubmissionWorkerNilDependency(timer) {
			return nil, mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorInvalid)
		}
		return &MailboxSubmissionWorker{}, nil
	}
	if mailboxSubmissionWorkerNilDependency(scanner) || mailboxSubmissionWorkerNilDependency(dispatcher) || mailboxSubmissionWorkerNilDependency(timer) {
		return nil, mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorInvalid)
	}
	configured, kind := mailboxSubmissionWorkerOptionsConfigured(options)
	if kind != "" {
		return nil, mailboxSubmissionWorkerError(kind)
	}
	return &MailboxSubmissionWorker{scanner: scanner, dispatcher: dispatcher, timer: timer, options: configured}, nil
}

// NewMailboxSubmissionWorkerTimer supplies the production wait/deadline timer.
func NewMailboxSubmissionWorkerTimer() MailboxSubmissionWorkerTimer {
	return mailboxSubmissionWorkerRuntimeTimer{}
}

func (mailboxSubmissionWorkerRuntimeTimer) Wait(ctx context.Context, duration time.Duration) error {
	if mailboxSubmissionWorkerNilDependency(ctx) {
		return context.Canceled
	}
	timer := time.NewTimer(duration)
	defer func() {
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
	}()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (mailboxSubmissionWorkerRuntimeTimer) WithTimeout(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	if mailboxSubmissionWorkerNilDependency(ctx) {
		return nil, nil
	}
	return context.WithTimeout(ctx, duration)
}

// RunOnce scans one bounded queued-due batch and attempts each distinct,
// nonzero identity once. Per-item failures are aggregated without exposing IDs
// or provider data; the durable due state decides any later dispatch.
func (worker *MailboxSubmissionWorker) RunOnce(ctx context.Context) (MailboxSubmissionWorkerRunResult, error) {
	if worker == nil || mailboxSubmissionWorkerNilDependency(ctx) {
		return MailboxSubmissionWorkerRunResult{}, mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorInvalid)
	}
	if !worker.enter() {
		return MailboxSubmissionWorkerRunResult{}, mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorInvalid)
	}
	defer worker.leave()
	return worker.runOnce(ctx)
}

// Run keeps one guarded sequential loop alive until the parent is cancelled.
func (worker *MailboxSubmissionWorker) Run(ctx context.Context) (resultErr error) {
	if worker == nil || mailboxSubmissionWorkerNilDependency(ctx) {
		return mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorInvalid)
	}
	if !worker.enter() {
		return mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorInvalid)
	}
	defer func() {
		worker.leave()
		if recover() != nil {
			resultErr = mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorUnavailable)
		}
	}()
	if !worker.options.Enabled {
		return nil
	}
	for {
		switch mailboxSubmissionWorkerParentState(ctx) {
		case mailboxSubmissionWorkerContextCancelled:
			return nil
		case mailboxSubmissionWorkerContextUnavailable:
			return mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorUnavailable)
		}
		_, runErr := worker.runOnce(ctx)
		switch mailboxSubmissionWorkerParentState(ctx) {
		case mailboxSubmissionWorkerContextCancelled:
			return nil
		case mailboxSubmissionWorkerContextUnavailable:
			return mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorUnavailable)
		}
		if kind := mailboxSubmissionWorkerErrorKindOf(runErr); kind == MailboxSubmissionWorkerErrorInvalid || kind == MailboxSubmissionWorkerErrorTooLarge {
			return runErr
		}
		if err := worker.wait(ctx, worker.options.ScanInterval); err != nil {
			return err
		}
	}
}

func (worker *MailboxSubmissionWorker) runOnce(ctx context.Context) (MailboxSubmissionWorkerRunResult, error) {
	switch mailboxSubmissionWorkerParentState(ctx) {
	case mailboxSubmissionWorkerContextCancelled:
		return MailboxSubmissionWorkerRunResult{}, nil
	case mailboxSubmissionWorkerContextUnavailable:
		return MailboxSubmissionWorkerRunResult{}, mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorUnavailable)
	}
	if !worker.options.Enabled {
		return MailboxSubmissionWorkerRunResult{}, nil
	}

	scan, scanErr, scanPanicked := worker.safeScan(ctx)
	switch mailboxSubmissionWorkerParentState(ctx) {
	case mailboxSubmissionWorkerContextCancelled:
		return MailboxSubmissionWorkerRunResult{}, nil
	case mailboxSubmissionWorkerContextUnavailable:
		return MailboxSubmissionWorkerRunResult{}, mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorUnavailable)
	}
	if scanPanicked || scanErr != nil {
		return MailboxSubmissionWorkerRunResult{}, mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorUnavailable)
	}
	ids, idsErr := mailboxSubmissionWorkerDueIDs(scan, worker.options.BatchSize)
	if idsErr != nil {
		return MailboxSubmissionWorkerRunResult{}, idsErr
	}

	result := MailboxSubmissionWorkerRunResult{Due: len(ids)}
	for _, submissionID := range ids {
		switch mailboxSubmissionWorkerParentState(ctx) {
		case mailboxSubmissionWorkerContextCancelled:
			return result, nil
		case mailboxSubmissionWorkerContextUnavailable:
			return result, mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorUnavailable)
		}
		child, cancel, timeoutPanicked := worker.withTimeout(ctx)
		if timeoutPanicked || mailboxSubmissionWorkerNilDependency(child) || cancel == nil {
			if cancel != nil {
				_ = mailboxSubmissionWorkerCancel(cancel)
			}
			result.Failed++
			continue
		}
		dispatchErr, dispatchPanicked := worker.safeDispatch(child, submissionID)
		childState := mailboxSubmissionWorkerParentState(child)
		cancelPanicked := mailboxSubmissionWorkerCancel(cancel)
		switch mailboxSubmissionWorkerParentState(ctx) {
		case mailboxSubmissionWorkerContextCancelled:
			return result, nil
		case mailboxSubmissionWorkerContextUnavailable:
			return result, mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorUnavailable)
		}
		if timeoutPanicked || dispatchPanicked || cancelPanicked || childState != mailboxSubmissionWorkerContextActive || dispatchErr != nil {
			result.Failed++
			if mailboxSubmissionWorkerReconciliationRequired(dispatchErr) {
				result.ReconciliationRequired++
			}
			continue
		}
		result.Dispatched++
	}
	if result.Failed != 0 {
		return result, mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorPartialFailure)
	}
	return result, nil
}

func (worker *MailboxSubmissionWorker) safeScan(ctx context.Context) (result MailboxSubmissionWorkerDueScanResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result = MailboxSubmissionWorkerDueScanResult{}
			resultErr = nil
			panicked = true
		}
	}()
	result, resultErr = worker.scanner.ScanDueMailboxSubmissionIDs(ctx, MailboxSubmissionWorkerDueScanRequest{Limit: worker.options.BatchSize})
	return result, resultErr, false
}

func (worker *MailboxSubmissionWorker) safeDispatch(ctx context.Context, submissionID uuid.UUID) (resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			resultErr = nil
			panicked = true
		}
	}()
	resultErr = worker.dispatcher.Dispatch(ctx, submissionID)
	return resultErr, false
}

func (worker *MailboxSubmissionWorker) withTimeout(ctx context.Context) (child context.Context, cancel context.CancelFunc, panicked bool) {
	defer func() {
		if recover() != nil {
			child = nil
			cancel = nil
			panicked = true
		}
	}()
	child, cancel = worker.timer.WithTimeout(ctx, worker.options.DispatchTimeout)
	return child, cancel, false
}

func (worker *MailboxSubmissionWorker) wait(ctx context.Context, duration time.Duration) (resultErr error) {
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		resultErr = worker.timer.Wait(ctx, duration)
	}()
	switch mailboxSubmissionWorkerParentState(ctx) {
	case mailboxSubmissionWorkerContextCancelled:
		return nil
	case mailboxSubmissionWorkerContextUnavailable:
		return mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorUnavailable)
	}
	if panicked || resultErr != nil {
		return mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorUnavailable)
	}
	return nil
}

func (worker *MailboxSubmissionWorker) enter() bool {
	worker.guard.Lock()
	defer worker.guard.Unlock()
	if worker.running {
		return false
	}
	worker.running = true
	return true
}

func (worker *MailboxSubmissionWorker) leave() {
	worker.guard.Lock()
	worker.running = false
	worker.guard.Unlock()
}

func mailboxSubmissionWorkerDueIDs(scan MailboxSubmissionWorkerDueScanResult, limit int) ([]uuid.UUID, error) {
	if limit < 1 {
		return nil, mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorInvalid)
	}
	if len(scan.SubmissionIDs) > limit {
		return nil, mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorTooLarge)
	}
	ids := make([]uuid.UUID, 0, len(scan.SubmissionIDs))
	seen := make(map[uuid.UUID]struct{}, len(scan.SubmissionIDs))
	for _, submissionID := range scan.SubmissionIDs {
		if submissionID == uuid.Nil {
			continue
		}
		if _, duplicate := seen[submissionID]; duplicate {
			continue
		}
		seen[submissionID] = struct{}{}
		ids = append(ids, submissionID)
	}
	return ids, nil
}

func mailboxSubmissionWorkerOptionsConfigured(options MailboxSubmissionWorkerOptions) (MailboxSubmissionWorkerOptions, MailboxSubmissionWorkerErrorKind) {
	defaults := DefaultMailboxSubmissionWorkerOptions()
	if options.ScanInterval == 0 {
		options.ScanInterval = defaults.ScanInterval
	}
	if options.DispatchTimeout == 0 {
		options.DispatchTimeout = defaults.DispatchTimeout
	}
	if options.BatchSize == 0 {
		options.BatchSize = defaults.BatchSize
	}
	if options.ScanInterval > mailboxSubmissionWorkerMaximumScanInterval || options.DispatchTimeout > mailboxSubmissionWorkerMaximumDispatchTimeout || options.BatchSize > mailboxSubmissionWorkerMaximumBatchSize {
		return MailboxSubmissionWorkerOptions{}, MailboxSubmissionWorkerErrorTooLarge
	}
	if options.ScanInterval < time.Second || options.DispatchTimeout < time.Second || options.BatchSize < 1 {
		return MailboxSubmissionWorkerOptions{}, MailboxSubmissionWorkerErrorInvalid
	}
	return options, ""
}

func mailboxSubmissionWorkerError(kind MailboxSubmissionWorkerErrorKind) error {
	if !mailboxSubmissionWorkerErrorKindValid(kind) {
		kind = MailboxSubmissionWorkerErrorInvalid
	}
	return &MailboxSubmissionWorkerError{kind: kind}
}

func mailboxSubmissionWorkerErrorKindValid(kind MailboxSubmissionWorkerErrorKind) bool {
	switch kind {
	case MailboxSubmissionWorkerErrorInvalid,
		MailboxSubmissionWorkerErrorTooLarge,
		MailboxSubmissionWorkerErrorUnavailable,
		MailboxSubmissionWorkerErrorPartialFailure:
		return true
	default:
		return false
	}
}

func mailboxSubmissionWorkerErrorKindOf(err error) MailboxSubmissionWorkerErrorKind {
	var workerError *MailboxSubmissionWorkerError
	if errors.As(err, &workerError) {
		return workerError.Kind()
	}
	return ""
}

func mailboxSubmissionWorkerReconciliationRequired(err error) bool {
	var dispatchError *MailboxSubmissionDispatchError
	return errors.As(err, &dispatchError) && dispatchError != nil && dispatchError.Kind() == MailboxSubmissionDispatchErrorReconciliationRequired
}

func mailboxSubmissionWorkerNilDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func mailboxSubmissionWorkerParentState(ctx context.Context) (state mailboxSubmissionWorkerContextState) {
	defer func() {
		if recover() != nil {
			state = mailboxSubmissionWorkerContextUnavailable
		}
	}()
	if mailboxSubmissionWorkerNilDependency(ctx) {
		return mailboxSubmissionWorkerContextUnavailable
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return mailboxSubmissionWorkerContextCancelled
		}
		return mailboxSubmissionWorkerContextUnavailable
	}
	return mailboxSubmissionWorkerContextActive
}

func mailboxSubmissionWorkerCancel(cancel context.CancelFunc) (panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	cancel()
	return false
}
