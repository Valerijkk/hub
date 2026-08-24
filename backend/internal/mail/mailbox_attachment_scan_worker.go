package mail

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"
)

const (
	mailboxAttachmentScanWorkerMaximumScanInterval  = 5 * time.Minute
	mailboxAttachmentScanWorkerMaximumLeaseDuration = 10 * time.Minute
	mailboxAttachmentScanWorkerMaximumRetryDelay    = 30 * time.Minute
)

// MailboxAttachmentScanWorkerTimer is the injected wait and timeout boundary.
type MailboxAttachmentScanWorkerTimer interface {
	Wait(context.Context, time.Duration) error
	WithTimeout(context.Context, time.Duration) (context.Context, context.CancelFunc)
}

// MailboxAttachmentScanWorkerConfig bounds one sequential attachment-scan worker.
type MailboxAttachmentScanWorkerConfig struct {
	Enabled       bool
	ScanInterval  time.Duration
	StepTimeout   time.Duration
	LeaseDuration time.Duration
	RetryMin      time.Duration
	RetryMax      time.Duration
	BatchSize     int
	Concurrency   int
}

// MailboxAttachmentScanWorkerErrorKind is the closed worker failure family.
type MailboxAttachmentScanWorkerErrorKind string

const (
	MailboxAttachmentScanWorkerErrorInvalid     MailboxAttachmentScanWorkerErrorKind = "invalid"
	MailboxAttachmentScanWorkerErrorTooLarge    MailboxAttachmentScanWorkerErrorKind = "too_large"
	MailboxAttachmentScanWorkerErrorUnavailable MailboxAttachmentScanWorkerErrorKind = "unavailable"
)

var (
	ErrMailboxAttachmentScanWorkerInvalid     = errors.New("mailbox attachment scan worker invalid")
	ErrMailboxAttachmentScanWorkerTooLarge    = errors.New("mailbox attachment scan worker exceeds bounds")
	ErrMailboxAttachmentScanWorkerUnavailable = errors.New("mailbox attachment scan worker unavailable")
)

// MailboxAttachmentScanWorkerError retains only a closed content-free kind.
type MailboxAttachmentScanWorkerError struct {
	kind MailboxAttachmentScanWorkerErrorKind
}

func (workerError *MailboxAttachmentScanWorkerError) Error() string {
	switch workerError.Kind() {
	case MailboxAttachmentScanWorkerErrorInvalid:
		return ErrMailboxAttachmentScanWorkerInvalid.Error()
	case MailboxAttachmentScanWorkerErrorTooLarge:
		return ErrMailboxAttachmentScanWorkerTooLarge.Error()
	case MailboxAttachmentScanWorkerErrorUnavailable:
		return ErrMailboxAttachmentScanWorkerUnavailable.Error()
	default:
		return "mailbox attachment scan worker error"
	}
}

func (workerError *MailboxAttachmentScanWorkerError) Unwrap() error {
	switch workerError.Kind() {
	case MailboxAttachmentScanWorkerErrorInvalid:
		return ErrMailboxAttachmentScanWorkerInvalid
	case MailboxAttachmentScanWorkerErrorTooLarge:
		return ErrMailboxAttachmentScanWorkerTooLarge
	case MailboxAttachmentScanWorkerErrorUnavailable:
		return ErrMailboxAttachmentScanWorkerUnavailable
	default:
		return nil
	}
}

func (workerError *MailboxAttachmentScanWorkerError) Kind() MailboxAttachmentScanWorkerErrorKind {
	if workerError == nil || !mailboxAttachmentScanWorkerErrorKindValid(workerError.kind) {
		return ""
	}
	return workerError.kind
}

// MailboxAttachmentScanWorker claims, scans, and finalizes attachments sequentially.
type MailboxAttachmentScanWorker struct {
	repository  mailboxAttachmentScanClaimRepository
	objectStore MailboxObjectStore
	scanner     MailboxAttachmentMalwareScanner
	timer       MailboxAttachmentScanWorkerTimer
	config      MailboxAttachmentScanWorkerConfig

	guard   sync.Mutex
	running bool
}

type mailboxAttachmentScanWorkerRuntimeTimer struct{}

type mailboxAttachmentScanWorkerContextState uint8

const (
	mailboxAttachmentScanWorkerContextActive mailboxAttachmentScanWorkerContextState = iota + 1
	mailboxAttachmentScanWorkerContextCancelled
	mailboxAttachmentScanWorkerContextUnavailable
)

type mailboxAttachmentScanWorkerItemStatus uint8

const (
	mailboxAttachmentScanWorkerItemComplete mailboxAttachmentScanWorkerItemStatus = iota + 1
	mailboxAttachmentScanWorkerItemCancelled
	mailboxAttachmentScanWorkerItemUnavailable
)

var _ MailboxAttachmentScanWorkerTimer = mailboxAttachmentScanWorkerRuntimeTimer{}

// NewMailboxAttachmentScanWorker constructs a disabled zero worker or one fully
// bounded enabled worker. Construction performs no repository or provider I/O.
func NewMailboxAttachmentScanWorker(
	repository mailboxAttachmentScanClaimRepository,
	objectStore MailboxObjectStore,
	scanner MailboxAttachmentMalwareScanner,
	timer MailboxAttachmentScanWorkerTimer,
	config MailboxAttachmentScanWorkerConfig,
) (*MailboxAttachmentScanWorker, error) {
	if !config.Enabled {
		if config != (MailboxAttachmentScanWorkerConfig{}) || !mailboxAttachmentScanWorkerNilDependency(repository) || !mailboxAttachmentScanWorkerNilDependency(objectStore) || !mailboxAttachmentScanWorkerNilDependency(scanner) || !mailboxAttachmentScanWorkerNilDependency(timer) {
			return nil, mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorInvalid)
		}
		return &MailboxAttachmentScanWorker{}, nil
	}
	if mailboxAttachmentScanWorkerNilDependency(repository) || mailboxAttachmentScanWorkerNilDependency(objectStore) || mailboxAttachmentScanWorkerNilDependency(scanner) || mailboxAttachmentScanWorkerNilDependency(timer) {
		return nil, mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorInvalid)
	}
	if kind := mailboxAttachmentScanWorkerConfigKind(config); kind != "" {
		return nil, mailboxAttachmentScanWorkerError(kind)
	}
	return &MailboxAttachmentScanWorker{
		repository:  repository,
		objectStore: objectStore,
		scanner:     scanner,
		timer:       timer,
		config:      config,
	}, nil
}

// NewMailboxAttachmentScanWorkerTimer supplies the stateless production timer.
func NewMailboxAttachmentScanWorkerTimer() MailboxAttachmentScanWorkerTimer {
	return mailboxAttachmentScanWorkerRuntimeTimer{}
}

func (mailboxAttachmentScanWorkerRuntimeTimer) Wait(ctx context.Context, duration time.Duration) error {
	if mailboxAttachmentScanWorkerNilDependency(ctx) {
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

func (mailboxAttachmentScanWorkerRuntimeTimer) WithTimeout(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	if mailboxAttachmentScanWorkerNilDependency(ctx) {
		return nil, nil
	}
	return context.WithTimeout(ctx, duration)
}

// Run holds a single-run guard for the lifetime of the worker loop.
func (worker *MailboxAttachmentScanWorker) Run(ctx context.Context) (resultErr error) {
	if worker == nil || mailboxAttachmentScanWorkerNilDependency(ctx) {
		return mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorInvalid)
	}
	worker.guard.Lock()
	if worker.running {
		worker.guard.Unlock()
		return mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorInvalid)
	}
	worker.running = true
	worker.guard.Unlock()
	defer func() {
		panicked := recover() != nil
		worker.guard.Lock()
		worker.running = false
		worker.guard.Unlock()
		if panicked {
			resultErr = mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorUnavailable)
		}
	}()

	switch mailboxAttachmentScanWorkerParentState(ctx) {
	case mailboxAttachmentScanWorkerContextCancelled:
		return nil
	case mailboxAttachmentScanWorkerContextUnavailable:
		return mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorUnavailable)
	}
	if !worker.config.Enabled {
		return nil
	}
	return worker.run(ctx)
}

func (worker *MailboxAttachmentScanWorker) run(ctx context.Context) error {
	request := mailboxAttachmentScanClaimRequest{Limit: worker.config.BatchSize, LeaseDuration: worker.config.LeaseDuration}
	for {
		switch mailboxAttachmentScanWorkerParentState(ctx) {
		case mailboxAttachmentScanWorkerContextCancelled:
			return nil
		case mailboxAttachmentScanWorkerContextUnavailable:
			return mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorUnavailable)
		}

		claimed, claimErr, claimPanicked := worker.claim(ctx, request)
		switch mailboxAttachmentScanWorkerParentState(ctx) {
		case mailboxAttachmentScanWorkerContextCancelled:
			return nil
		case mailboxAttachmentScanWorkerContextUnavailable:
			return mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorUnavailable)
		}
		if claimPanicked {
			return mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorUnavailable)
		}
		if claimErr != nil {
			kind, ok := mailboxAttachmentScanClaimRepositoryErrorKindOf(claimErr)
			if !ok || kind != mailboxAttachmentScanClaimRepositoryErrorUnavailable {
				return mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorUnavailable)
			}
			if err := worker.wait(ctx, worker.config.RetryMin); err != nil {
				return err
			}
			continue
		}
		if len(claimed.Items) > worker.config.BatchSize {
			return mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorTooLarge)
		}
		if !mailboxAttachmentScanClaimResultValid(request, claimed) {
			return mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorInvalid)
		}
		claimed = mailboxAttachmentScanClaimResultCopy(claimed)

		batchUnavailable := false
		for index := range claimed.Items {
			status := worker.processItem(ctx, claimed.Items[index])
			switch status {
			case mailboxAttachmentScanWorkerItemCancelled:
				return nil
			case mailboxAttachmentScanWorkerItemUnavailable:
				batchUnavailable = true
			}
		}
		if batchUnavailable {
			return mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorUnavailable)
		}
		if err := worker.wait(ctx, worker.config.ScanInterval); err != nil {
			return err
		}
	}
}

func (worker *MailboxAttachmentScanWorker) claim(ctx context.Context, request mailboxAttachmentScanClaimRequest) (result mailboxAttachmentScanClaimResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result = mailboxAttachmentScanClaimResult{}
			resultErr = nil
			panicked = true
		}
	}()
	result, resultErr = worker.repository.claimMailboxAttachmentScans(ctx, request)
	return result, resultErr, false
}

func (worker *MailboxAttachmentScanWorker) processItem(ctx context.Context, item mailboxAttachmentScanClaimedItem) (status mailboxAttachmentScanWorkerItemStatus) {
	switch mailboxAttachmentScanWorkerParentState(ctx) {
	case mailboxAttachmentScanWorkerContextCancelled:
		return mailboxAttachmentScanWorkerItemCancelled
	case mailboxAttachmentScanWorkerContextUnavailable:
		return mailboxAttachmentScanWorkerItemUnavailable
	}

	child, cancel, timeoutPanicked := worker.withTimeout(ctx)
	if timeoutPanicked || mailboxAttachmentScanWorkerNilDependency(child) || cancel == nil {
		if cancel != nil {
			_ = mailboxAttachmentScanWorkerCancel(cancel)
		}
		switch mailboxAttachmentScanWorkerParentState(ctx) {
		case mailboxAttachmentScanWorkerContextCancelled:
			return mailboxAttachmentScanWorkerItemCancelled
		case mailboxAttachmentScanWorkerContextUnavailable:
			return mailboxAttachmentScanWorkerItemUnavailable
		default:
			return mailboxAttachmentScanWorkerItemUnavailable
		}
	}
	defer func() {
		if mailboxAttachmentScanWorkerCancel(cancel) && status != mailboxAttachmentScanWorkerItemCancelled {
			status = mailboxAttachmentScanWorkerItemUnavailable
		}
	}()

	finalizeKind, failed := worker.scanOutcome(child, item)
	switch mailboxAttachmentScanWorkerParentState(ctx) {
	case mailboxAttachmentScanWorkerContextCancelled:
		return mailboxAttachmentScanWorkerItemCancelled
	case mailboxAttachmentScanWorkerContextUnavailable:
		return mailboxAttachmentScanWorkerItemUnavailable
	}
	if !mailboxAttachmentScanWorkerChildUsable(child) {
		return mailboxAttachmentScanWorkerItemUnavailable
	}
	if failed {
		return worker.finalizeFailure(ctx, child, item)
	}
	return worker.finalize(ctx, child, mailboxAttachmentScanFinalizeRequest{Item: item, Kind: finalizeKind})
}

func (worker *MailboxAttachmentScanWorker) scanOutcome(ctx context.Context, item mailboxAttachmentScanClaimedItem) (mailboxAttachmentScanFinalizeKind, bool) {
	integrityClean, verdict, scanErr, panicked := worker.scan(ctx, item)
	if panicked || !mailboxAttachmentScanWorkerChildUsable(ctx) || scanErr != nil {
		return "", true
	}
	if verdict == MailboxAttachmentMalwareVerdictRejected {
		return mailboxAttachmentScanFinalizeRejected, false
	}
	if integrityClean && verdict == MailboxAttachmentMalwareVerdictClean {
		return mailboxAttachmentScanFinalizeClean, false
	}
	return "", true
}

func (worker *MailboxAttachmentScanWorker) scan(ctx context.Context, item mailboxAttachmentScanClaimedItem) (integrityClean bool, verdict MailboxAttachmentMalwareVerdict, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			integrityClean = false
			verdict = ""
			resultErr = nil
			panicked = true
		}
	}()
	integrityClean, verdict, resultErr = mailboxAttachmentScanStream(
		worker.objectStore,
		worker.scanner,
		ctx,
		MailboxObjectLocator{Bucket: item.Bucket, Key: item.ObjectKey},
		item.ExpectedSizeBytes,
		item.ExpectedSHA256,
	)
	return integrityClean, verdict, resultErr, false
}

func (worker *MailboxAttachmentScanWorker) finalizeFailure(ctx, child context.Context, item mailboxAttachmentScanClaimedItem) mailboxAttachmentScanWorkerItemStatus {
	request := mailboxAttachmentScanFinalizeRequest{Item: item}
	if item.Attempt == 10 {
		request.Kind = mailboxAttachmentScanFinalizeTerminal
	} else {
		request.Kind = mailboxAttachmentScanFinalizeRetry
		request.RetryDelay = worker.retryDelay(item.Attempt)
	}
	return worker.finalize(ctx, child, request)
}

func (worker *MailboxAttachmentScanWorker) finalize(parent, child context.Context, request mailboxAttachmentScanFinalizeRequest) mailboxAttachmentScanWorkerItemStatus {
	switch mailboxAttachmentScanWorkerParentState(parent) {
	case mailboxAttachmentScanWorkerContextCancelled:
		return mailboxAttachmentScanWorkerItemCancelled
	case mailboxAttachmentScanWorkerContextUnavailable:
		return mailboxAttachmentScanWorkerItemUnavailable
	}
	if !mailboxAttachmentScanWorkerChildUsable(child) {
		return mailboxAttachmentScanWorkerItemUnavailable
	}
	result, resultErr, panicked := worker.finalizeAttachmentScan(child, request)
	switch mailboxAttachmentScanWorkerParentState(parent) {
	case mailboxAttachmentScanWorkerContextCancelled:
		return mailboxAttachmentScanWorkerItemCancelled
	case mailboxAttachmentScanWorkerContextUnavailable:
		return mailboxAttachmentScanWorkerItemUnavailable
	}
	if panicked || resultErr != nil || !mailboxAttachmentScanWorkerChildUsable(child) {
		return mailboxAttachmentScanWorkerItemUnavailable
	}
	result = mailboxAttachmentScanFinalizeResultCopy(result)
	if !mailboxAttachmentScanFinalizeResultValid(request, result) {
		return mailboxAttachmentScanWorkerItemUnavailable
	}
	return mailboxAttachmentScanWorkerItemComplete
}

func (worker *MailboxAttachmentScanWorker) finalizeAttachmentScan(ctx context.Context, request mailboxAttachmentScanFinalizeRequest) (result mailboxAttachmentScanFinalizeResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result = mailboxAttachmentScanFinalizeResult{}
			resultErr = nil
			panicked = true
		}
	}()
	result, resultErr = worker.repository.finalizeMailboxAttachmentScan(ctx, request)
	return result, resultErr, false
}

func (worker *MailboxAttachmentScanWorker) withTimeout(ctx context.Context) (child context.Context, cancel context.CancelFunc, panicked bool) {
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

func (worker *MailboxAttachmentScanWorker) wait(ctx context.Context, duration time.Duration) (resultErr error) {
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		resultErr = worker.timer.Wait(ctx, duration)
	}()
	switch mailboxAttachmentScanWorkerParentState(ctx) {
	case mailboxAttachmentScanWorkerContextCancelled:
		return nil
	case mailboxAttachmentScanWorkerContextUnavailable:
		return mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorUnavailable)
	}
	if panicked || resultErr != nil {
		return mailboxAttachmentScanWorkerError(MailboxAttachmentScanWorkerErrorUnavailable)
	}
	return nil
}

func (worker *MailboxAttachmentScanWorker) retryDelay(attempt int32) time.Duration {
	if attempt < 1 || attempt >= 10 {
		return worker.config.RetryMax
	}
	shift := uint(attempt - 1)
	maximum := time.Duration(1<<63 - 1)
	if worker.config.RetryMin > maximum>>shift {
		return worker.config.RetryMax
	}
	delay := worker.config.RetryMin << shift
	if delay > worker.config.RetryMax {
		return worker.config.RetryMax
	}
	return delay
}

func mailboxAttachmentScanWorkerConfigKind(config MailboxAttachmentScanWorkerConfig) MailboxAttachmentScanWorkerErrorKind {
	if config.BatchSize > mailboxAttachmentScanClaimMaximum || config.ScanInterval > mailboxAttachmentScanWorkerMaximumScanInterval || config.LeaseDuration > mailboxAttachmentScanWorkerMaximumLeaseDuration || config.RetryMax > mailboxAttachmentScanWorkerMaximumRetryDelay {
		return MailboxAttachmentScanWorkerErrorTooLarge
	}
	if config.BatchSize < 1 || config.Concurrency != 1 || !mailboxAttachmentScanWorkerWholeSecond(config.ScanInterval) || !mailboxAttachmentScanWorkerWholeSecond(config.StepTimeout) || !mailboxAttachmentScanWorkerWholeSecond(config.LeaseDuration) || !mailboxAttachmentScanWorkerWholeSecond(config.RetryMin) || !mailboxAttachmentScanWorkerWholeSecond(config.RetryMax) || config.StepTimeout >= config.LeaseDuration || config.RetryMin > config.RetryMax {
		return MailboxAttachmentScanWorkerErrorInvalid
	}
	return ""
}

func mailboxAttachmentScanWorkerWholeSecond(value time.Duration) bool {
	return value > 0 && value%time.Second == 0
}

func mailboxAttachmentScanWorkerError(kind MailboxAttachmentScanWorkerErrorKind) error {
	if !mailboxAttachmentScanWorkerErrorKindValid(kind) {
		kind = MailboxAttachmentScanWorkerErrorInvalid
	}
	return &MailboxAttachmentScanWorkerError{kind: kind}
}

func mailboxAttachmentScanWorkerErrorKindValid(kind MailboxAttachmentScanWorkerErrorKind) bool {
	return kind == MailboxAttachmentScanWorkerErrorInvalid || kind == MailboxAttachmentScanWorkerErrorTooLarge || kind == MailboxAttachmentScanWorkerErrorUnavailable
}

func mailboxAttachmentScanWorkerNilDependency(dependency any) bool {
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

func mailboxAttachmentScanWorkerParentState(ctx context.Context) (state mailboxAttachmentScanWorkerContextState) {
	defer func() {
		if recover() != nil {
			state = mailboxAttachmentScanWorkerContextUnavailable
		}
	}()
	if mailboxAttachmentScanWorkerNilDependency(ctx) {
		return mailboxAttachmentScanWorkerContextUnavailable
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return mailboxAttachmentScanWorkerContextCancelled
		}
		return mailboxAttachmentScanWorkerContextUnavailable
	}
	return mailboxAttachmentScanWorkerContextActive
}

func mailboxAttachmentScanWorkerChildUsable(ctx context.Context) (usable bool) {
	defer func() {
		if recover() != nil {
			usable = false
		}
	}()
	return !mailboxAttachmentScanWorkerNilDependency(ctx) && ctx.Err() == nil
}

func mailboxAttachmentScanWorkerCancel(cancel context.CancelFunc) (panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	cancel()
	return false
}
