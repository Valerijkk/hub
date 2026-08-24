package mail

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"time"
)

const (
	mailboxObjectDeletionWorkerMaximumScanInterval = 5 * time.Minute
	mailboxObjectDeletionWorkerMaximumLeaseDuration = 10 * time.Minute
	mailboxObjectDeletionWorkerMaximumRetryDelay    = 30 * time.Minute
)

// MailboxObjectDeletionWorkerTimer is the injected wait and timeout boundary.
type MailboxObjectDeletionWorkerTimer interface {
	Wait(context.Context, time.Duration) error
	WithTimeout(context.Context, time.Duration) (context.Context, context.CancelFunc)
}

// MailboxObjectDeletionWorkerConfig bounds one sequential deletion worker.
type MailboxObjectDeletionWorkerConfig struct {
	Enabled       bool
	ScanInterval  time.Duration
	StepTimeout   time.Duration
	LeaseDuration time.Duration
	RetryMin      time.Duration
	RetryMax      time.Duration
	BatchSize     int
}

// MailboxObjectDeletionWorkerErrorKind is the closed worker failure family.
type MailboxObjectDeletionWorkerErrorKind string

const (
	MailboxObjectDeletionWorkerErrorInvalid     MailboxObjectDeletionWorkerErrorKind = "invalid"
	MailboxObjectDeletionWorkerErrorTooLarge    MailboxObjectDeletionWorkerErrorKind = "too_large"
	MailboxObjectDeletionWorkerErrorUnavailable MailboxObjectDeletionWorkerErrorKind = "unavailable"
)

var (
	ErrMailboxObjectDeletionWorkerInvalid     = errors.New("mailbox object deletion worker invalid")
	ErrMailboxObjectDeletionWorkerTooLarge    = errors.New("mailbox object deletion worker exceeds bounds")
	ErrMailboxObjectDeletionWorkerUnavailable = errors.New("mailbox object deletion worker unavailable")
)

// MailboxObjectDeletionWorkerError retains only a closed content-free kind.
type MailboxObjectDeletionWorkerError struct {
	kind MailboxObjectDeletionWorkerErrorKind
}

func (workerError *MailboxObjectDeletionWorkerError) Error() string {
	switch workerError.Kind() {
	case MailboxObjectDeletionWorkerErrorInvalid:
		return ErrMailboxObjectDeletionWorkerInvalid.Error()
	case MailboxObjectDeletionWorkerErrorTooLarge:
		return ErrMailboxObjectDeletionWorkerTooLarge.Error()
	case MailboxObjectDeletionWorkerErrorUnavailable:
		return ErrMailboxObjectDeletionWorkerUnavailable.Error()
	default:
		return "mailbox object deletion worker error"
	}
}

func (workerError *MailboxObjectDeletionWorkerError) Unwrap() error {
	switch workerError.Kind() {
	case MailboxObjectDeletionWorkerErrorInvalid:
		return ErrMailboxObjectDeletionWorkerInvalid
	case MailboxObjectDeletionWorkerErrorTooLarge:
		return ErrMailboxObjectDeletionWorkerTooLarge
	case MailboxObjectDeletionWorkerErrorUnavailable:
		return ErrMailboxObjectDeletionWorkerUnavailable
	default:
		return nil
	}
}

func (workerError *MailboxObjectDeletionWorkerError) Kind() MailboxObjectDeletionWorkerErrorKind {
	if workerError == nil || !mailboxObjectDeletionWorkerErrorKindValid(workerError.kind) {
		return ""
	}
	return workerError.kind
}

// MailboxObjectDeletionWorker claims, deletes, and finalizes objects sequentially.
type MailboxObjectDeletionWorker struct {
	repository mailboxObjectDeletionRepository
	deletion   MailObjectDeletionPort
	timer      MailboxObjectDeletionWorkerTimer
	config     MailboxObjectDeletionWorkerConfig

	guard   sync.Mutex
	running bool
}

type mailboxObjectDeletionWorkerRuntimeTimer struct{}

type mailboxObjectDeletionWorkerContextState uint8

const (
	mailboxObjectDeletionWorkerContextActive mailboxObjectDeletionWorkerContextState = iota + 1
	mailboxObjectDeletionWorkerContextCancelled
	mailboxObjectDeletionWorkerContextUnavailable
)

type mailboxObjectDeletionWorkerItemStatus uint8

const (
	mailboxObjectDeletionWorkerItemComplete mailboxObjectDeletionWorkerItemStatus = iota + 1
	mailboxObjectDeletionWorkerItemCancelled
	mailboxObjectDeletionWorkerItemUnavailable
)

var _ MailboxObjectDeletionWorkerTimer = mailboxObjectDeletionWorkerRuntimeTimer{}

// NewMailboxObjectDeletionWorker constructs a disabled zero worker or one fully
// bounded enabled worker. Construction performs no repository or provider I/O.
func NewMailboxObjectDeletionWorker(repository mailboxObjectDeletionRepository, deletion MailObjectDeletionPort, timer MailboxObjectDeletionWorkerTimer, config MailboxObjectDeletionWorkerConfig) (*MailboxObjectDeletionWorker, error) {
	if !config.Enabled {
		if config != (MailboxObjectDeletionWorkerConfig{}) || !mailboxObjectDeletionWorkerNilDependency(repository) || !mailboxObjectDeletionWorkerNilDependency(deletion) || !mailboxObjectDeletionWorkerNilDependency(timer) {
			return nil, mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorInvalid)
		}
		return &MailboxObjectDeletionWorker{}, nil
	}
	if mailboxObjectDeletionWorkerNilDependency(repository) || mailboxObjectDeletionWorkerNilDependency(deletion) || mailboxObjectDeletionWorkerNilDependency(timer) {
		return nil, mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorInvalid)
	}
	if kind := mailboxObjectDeletionWorkerConfigKind(config); kind != "" {
		return nil, mailboxObjectDeletionWorkerError(kind)
	}
	return &MailboxObjectDeletionWorker{repository: repository, deletion: deletion, timer: timer, config: config}, nil
}

// NewMailboxObjectDeletionWorkerTimer supplies the stateless production timer.
func NewMailboxObjectDeletionWorkerTimer() MailboxObjectDeletionWorkerTimer {
	return mailboxObjectDeletionWorkerRuntimeTimer{}
}

func (mailboxObjectDeletionWorkerRuntimeTimer) Wait(ctx context.Context, duration time.Duration) error {
	if mailboxObjectDeletionWorkerNilDependency(ctx) {
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

func (mailboxObjectDeletionWorkerRuntimeTimer) WithTimeout(ctx context.Context, duration time.Duration) (context.Context, context.CancelFunc) {
	if mailboxObjectDeletionWorkerNilDependency(ctx) {
		return nil, nil
	}
	return context.WithTimeout(ctx, duration)
}

// Run holds a single-run guard for the lifetime of the worker loop.
func (worker *MailboxObjectDeletionWorker) Run(ctx context.Context) (resultErr error) {
	if worker == nil || mailboxObjectDeletionWorkerNilDependency(ctx) {
		return mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorInvalid)
	}
	worker.guard.Lock()
	if worker.running {
		worker.guard.Unlock()
		return mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorInvalid)
	}
	worker.running = true
	worker.guard.Unlock()
	defer func() {
		panicked := recover() != nil
		worker.guard.Lock()
		worker.running = false
		worker.guard.Unlock()
		if panicked {
			resultErr = mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorUnavailable)
		}
	}()

	switch mailboxObjectDeletionWorkerParentState(ctx) {
	case mailboxObjectDeletionWorkerContextCancelled:
		return nil
	case mailboxObjectDeletionWorkerContextUnavailable:
		return mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorUnavailable)
	}
	if !worker.config.Enabled {
		return nil
	}
	return worker.run(ctx)
}

func (worker *MailboxObjectDeletionWorker) run(ctx context.Context) error {
	request := mailboxObjectDeletionClaimRequest{Limit: worker.config.BatchSize, LeaseDuration: worker.config.LeaseDuration}
	for {
		switch mailboxObjectDeletionWorkerParentState(ctx) {
		case mailboxObjectDeletionWorkerContextCancelled:
			return nil
		case mailboxObjectDeletionWorkerContextUnavailable:
			return mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorUnavailable)
		}

		claimed, claimErr, claimPanicked := worker.claim(ctx, request)
		switch mailboxObjectDeletionWorkerParentState(ctx) {
		case mailboxObjectDeletionWorkerContextCancelled:
			return nil
		case mailboxObjectDeletionWorkerContextUnavailable:
			return mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorUnavailable)
		}
		if claimPanicked {
			return mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorUnavailable)
		}
		if claimErr != nil {
			kind, ok := mailboxObjectDeletionRepositoryErrorKindOf(claimErr)
			if !ok || kind != mailboxObjectDeletionRepositoryErrorUnavailable {
				return mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorUnavailable)
			}
			if err := worker.wait(ctx, worker.config.RetryMin); err != nil {
				return err
			}
			continue
		}
		if len(claimed.Items) > worker.config.BatchSize {
			return mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorTooLarge)
		}
		if !mailboxObjectDeletionClaimResultValid(request, claimed) {
			return mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorInvalid)
		}
		claimed = mailboxObjectDeletionClaimResultCopy(claimed)

		batchUnavailable := false
		for index := range claimed.Items {
			status := worker.processItem(ctx, claimed.Items[index])
			switch status {
			case mailboxObjectDeletionWorkerItemCancelled:
				return nil
			case mailboxObjectDeletionWorkerItemUnavailable:
				batchUnavailable = true
			}
		}
		if batchUnavailable {
			return mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorUnavailable)
		}
		if err := worker.wait(ctx, worker.config.ScanInterval); err != nil {
			return err
		}
	}
}

func (worker *MailboxObjectDeletionWorker) claim(ctx context.Context, request mailboxObjectDeletionClaimRequest) (result mailboxObjectDeletionClaimResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result = mailboxObjectDeletionClaimResult{}
			resultErr = nil
			panicked = true
		}
	}()
	result, resultErr = worker.repository.claimMailboxObjectDeletions(ctx, request)
	return result, resultErr, false
}

func (worker *MailboxObjectDeletionWorker) processItem(ctx context.Context, item mailboxObjectDeletionClaimedItem) mailboxObjectDeletionWorkerItemStatus {
	switch mailboxObjectDeletionWorkerParentState(ctx) {
	case mailboxObjectDeletionWorkerContextCancelled:
		return mailboxObjectDeletionWorkerItemCancelled
	case mailboxObjectDeletionWorkerContextUnavailable:
		return mailboxObjectDeletionWorkerItemUnavailable
	}

	child, cancel, timeoutPanicked := worker.withTimeout(ctx)
	if timeoutPanicked || mailboxObjectDeletionWorkerNilDependency(child) || cancel == nil {
		cancelPanicked := false
		if cancel != nil {
			cancelPanicked = mailboxObjectDeletionWorkerCancel(cancel)
		}
		switch mailboxObjectDeletionWorkerParentState(ctx) {
		case mailboxObjectDeletionWorkerContextCancelled:
			return mailboxObjectDeletionWorkerItemCancelled
		case mailboxObjectDeletionWorkerContextUnavailable:
			return mailboxObjectDeletionWorkerItemUnavailable
		}
		_ = cancelPanicked
		return worker.finalizeFailure(ctx, item)
	}

	finalizeKind, failed := worker.deleteOutcome(child, item)
	cancelPanicked := mailboxObjectDeletionWorkerCancel(cancel)
	switch mailboxObjectDeletionWorkerParentState(ctx) {
	case mailboxObjectDeletionWorkerContextCancelled:
		return mailboxObjectDeletionWorkerItemCancelled
	case mailboxObjectDeletionWorkerContextUnavailable:
		return mailboxObjectDeletionWorkerItemUnavailable
	}
	if cancelPanicked || failed {
		return worker.finalizeFailure(ctx, item)
	}
	return worker.finalize(ctx, mailboxObjectDeletionFinalizeRequest{Item: item, Kind: finalizeKind})
}

func (worker *MailboxObjectDeletionWorker) deleteOutcome(ctx context.Context, item mailboxObjectDeletionClaimedItem) (mailboxObjectDeletionFinalizeKind, bool) {
	request := MailObjectDeleteRequest{Bucket: item.Bucket, ObjectKey: item.ObjectKey, DeletionID: item.DeletionID}
	deleted, deleteErr, deletePanicked := worker.deleteObject(ctx, request)
	if deletePanicked || !mailboxObjectDeletionWorkerChildUsable(ctx) {
		return "", true
	}
	if deleteErr == nil {
		switch deleted.State {
		case MailObjectDeletionStateDeleted:
			return mailboxObjectDeletionFinalizeDeleted, false
		case MailObjectDeletionStateAuthoritativelyAbsent:
			return mailboxObjectDeletionFinalizeAuthoritativelyAbsent, false
		default:
			return "", true
		}
	}
	if deleted != (MailObjectDeleteResult{}) {
		return "", true
	}
	kind, ok := MailObjectDeletionErrorKindOf(deleteErr)
	if !ok || kind != MailObjectDeletionErrorIndeterminate || !mailboxObjectDeletionWorkerChildUsable(ctx) {
		return "", true
	}

	absence, absenceErr, absencePanicked := worker.reconcileObjectAbsence(ctx, MailObjectAbsenceRequest(request))
	if absencePanicked || !mailboxObjectDeletionWorkerChildUsable(ctx) || absenceErr != nil || !validMailObjectAbsenceState(absence.State) {
		return "", true
	}
	if absence.State == MailObjectAbsenceStateAuthoritativelyAbsent {
		return mailboxObjectDeletionFinalizeAuthoritativelyAbsent, false
	}
	return "", true
}

func (worker *MailboxObjectDeletionWorker) deleteObject(ctx context.Context, request MailObjectDeleteRequest) (result MailObjectDeleteResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result = MailObjectDeleteResult{}
			resultErr = nil
			panicked = true
		}
	}()
	result, resultErr = worker.deletion.DeleteObject(ctx, request)
	return result, resultErr, false
}

func (worker *MailboxObjectDeletionWorker) reconcileObjectAbsence(ctx context.Context, request MailObjectAbsenceRequest) (result MailObjectAbsenceResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result = MailObjectAbsenceResult{}
			resultErr = nil
			panicked = true
		}
	}()
	result, resultErr = worker.deletion.ReconcileObjectAbsence(ctx, request)
	return result, resultErr, false
}

func (worker *MailboxObjectDeletionWorker) finalizeFailure(ctx context.Context, item mailboxObjectDeletionClaimedItem) mailboxObjectDeletionWorkerItemStatus {
	request := mailboxObjectDeletionFinalizeRequest{Item: item}
	if item.Attempt == 10 {
		request.Kind = mailboxObjectDeletionFinalizeTerminal
	} else {
		request.Kind = mailboxObjectDeletionFinalizeRetry
		request.RetryDelay = worker.retryDelay(item.Attempt)
	}
	return worker.finalize(ctx, request)
}

func (worker *MailboxObjectDeletionWorker) finalize(ctx context.Context, request mailboxObjectDeletionFinalizeRequest) mailboxObjectDeletionWorkerItemStatus {
	switch mailboxObjectDeletionWorkerParentState(ctx) {
	case mailboxObjectDeletionWorkerContextCancelled:
		return mailboxObjectDeletionWorkerItemCancelled
	case mailboxObjectDeletionWorkerContextUnavailable:
		return mailboxObjectDeletionWorkerItemUnavailable
	}
	result, resultErr, panicked := worker.finalizeObjectDeletion(ctx, request)
	switch mailboxObjectDeletionWorkerParentState(ctx) {
	case mailboxObjectDeletionWorkerContextCancelled:
		return mailboxObjectDeletionWorkerItemCancelled
	case mailboxObjectDeletionWorkerContextUnavailable:
		return mailboxObjectDeletionWorkerItemUnavailable
	}
	result = mailboxObjectDeletionFinalizeResultCopy(result)
	if panicked || resultErr != nil || !mailboxObjectDeletionFinalizeResultValid(request, result) {
		return mailboxObjectDeletionWorkerItemUnavailable
	}
	return mailboxObjectDeletionWorkerItemComplete
}

func (worker *MailboxObjectDeletionWorker) finalizeObjectDeletion(ctx context.Context, request mailboxObjectDeletionFinalizeRequest) (result mailboxObjectDeletionFinalizeResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result = mailboxObjectDeletionFinalizeResult{}
			resultErr = nil
			panicked = true
		}
	}()
	result, resultErr = worker.repository.finalizeMailboxObjectDeletion(ctx, request)
	return result, resultErr, false
}

func (worker *MailboxObjectDeletionWorker) withTimeout(ctx context.Context) (child context.Context, cancel context.CancelFunc, panicked bool) {
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

func (worker *MailboxObjectDeletionWorker) wait(ctx context.Context, duration time.Duration) (resultErr error) {
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		resultErr = worker.timer.Wait(ctx, duration)
	}()
	switch mailboxObjectDeletionWorkerParentState(ctx) {
	case mailboxObjectDeletionWorkerContextCancelled:
		return nil
	case mailboxObjectDeletionWorkerContextUnavailable:
		return mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorUnavailable)
	}
	if panicked || resultErr != nil {
		return mailboxObjectDeletionWorkerError(MailboxObjectDeletionWorkerErrorUnavailable)
	}
	return nil
}

func (worker *MailboxObjectDeletionWorker) retryDelay(attempt int32) time.Duration {
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

func mailboxObjectDeletionWorkerConfigKind(config MailboxObjectDeletionWorkerConfig) MailboxObjectDeletionWorkerErrorKind {
	if config.BatchSize > mailboxObjectDeletionClaimMaximum || config.ScanInterval > mailboxObjectDeletionWorkerMaximumScanInterval || config.LeaseDuration > mailboxObjectDeletionWorkerMaximumLeaseDuration || config.RetryMax > mailboxObjectDeletionWorkerMaximumRetryDelay {
		return MailboxObjectDeletionWorkerErrorTooLarge
	}
	if config.BatchSize < 1 || !mailboxObjectDeletionWorkerWholeSecond(config.ScanInterval) || !mailboxObjectDeletionWorkerWholeSecond(config.StepTimeout) || !mailboxObjectDeletionWorkerWholeSecond(config.LeaseDuration) || !mailboxObjectDeletionWorkerWholeSecond(config.RetryMin) || !mailboxObjectDeletionWorkerWholeSecond(config.RetryMax) || config.StepTimeout >= config.LeaseDuration || config.RetryMin > config.RetryMax {
		return MailboxObjectDeletionWorkerErrorInvalid
	}
	return ""
}

func mailboxObjectDeletionWorkerWholeSecond(value time.Duration) bool {
	return value > 0 && value%time.Second == 0
}

func mailboxObjectDeletionWorkerError(kind MailboxObjectDeletionWorkerErrorKind) error {
	if !mailboxObjectDeletionWorkerErrorKindValid(kind) {
		kind = MailboxObjectDeletionWorkerErrorInvalid
	}
	return &MailboxObjectDeletionWorkerError{kind: kind}
}

func mailboxObjectDeletionWorkerErrorKindValid(kind MailboxObjectDeletionWorkerErrorKind) bool {
	return kind == MailboxObjectDeletionWorkerErrorInvalid || kind == MailboxObjectDeletionWorkerErrorTooLarge || kind == MailboxObjectDeletionWorkerErrorUnavailable
}

func mailboxObjectDeletionWorkerNilDependency(dependency any) bool {
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

func mailboxObjectDeletionWorkerParentState(ctx context.Context) (state mailboxObjectDeletionWorkerContextState) {
	defer func() {
		if recover() != nil {
			state = mailboxObjectDeletionWorkerContextUnavailable
		}
	}()
	if mailboxObjectDeletionWorkerNilDependency(ctx) {
		return mailboxObjectDeletionWorkerContextUnavailable
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return mailboxObjectDeletionWorkerContextCancelled
		}
		return mailboxObjectDeletionWorkerContextUnavailable
	}
	return mailboxObjectDeletionWorkerContextActive
}

func mailboxObjectDeletionWorkerChildUsable(ctx context.Context) (usable bool) {
	defer func() {
		if recover() != nil {
			usable = false
		}
	}()
	return !mailboxObjectDeletionWorkerNilDependency(ctx) && ctx.Err() == nil
}

func mailboxObjectDeletionWorkerCancel(cancel context.CancelFunc) (panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	cancel()
	return false
}
