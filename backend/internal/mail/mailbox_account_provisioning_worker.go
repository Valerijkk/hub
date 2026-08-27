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
	mailboxAccountProvisioningWorkerDefaultProviderTimeout = 30 * time.Second
	mailboxAccountProvisioningWorkerDefaultFinalizeTimeout = 10 * time.Second
	mailboxAccountProvisioningWorkerMaximumScanInterval    = 5 * time.Minute
	mailboxAccountProvisioningWorkerMaximumProviderTimeout = 30 * time.Second
	mailboxAccountProvisioningWorkerMaximumFinalizeTimeout = 10 * time.Second
	mailboxAccountProvisioningWorkerLeaseSafetyMargin      = time.Second
)

// MailboxAccountProvisioner is the provider-neutral account ensure boundary.
type MailboxAccountProvisioner interface {
	EnsureAccount(context.Context, AccountSpec) (EngineAccount, error)
}

// MailboxAccountProvisioningWorkerTimer supplies bounded worker time seams.
type MailboxAccountProvisioningWorkerTimer interface {
	Now() time.Time
	Wait(context.Context, time.Duration) error
	WithDeadline(context.Context, time.Time) (context.Context, context.CancelFunc)
}

// MailboxAccountProvisioningWorkerOptions bounds one sequential worker.
type MailboxAccountProvisioningWorkerOptions struct {
	Enabled         bool
	ScanInterval    time.Duration
	ProviderTimeout time.Duration
	FinalizeTimeout time.Duration
	BatchSize       int
}

// DefaultMailboxAccountProvisioningWorkerOptions returns conservative enabled options.
func DefaultMailboxAccountProvisioningWorkerOptions() MailboxAccountProvisioningWorkerOptions {
	return MailboxAccountProvisioningWorkerOptions{
		Enabled:         true,
		ScanInterval:    MailboxAccountProvisioningDefaultScanInterval,
		ProviderTimeout: mailboxAccountProvisioningWorkerDefaultProviderTimeout,
		FinalizeTimeout: mailboxAccountProvisioningWorkerDefaultFinalizeTimeout,
		BatchSize:       MailboxAccountProvisioningDefaultBatchSize,
	}
}

// MailboxAccountProvisioningWorkerRunResult contains only bounded work counts.
type MailboxAccountProvisioningWorkerRunResult struct {
	Due               int
	Claimed           int
	Noop              int
	LocalGated        int
	ProviderSucceeded int
	ProviderFailed    int
	FinalizedSuccess  int
	FinalizedRetry    int
	Unfinalized       int
}

// MailboxAccountProvisioningWorkerErrorKind is the closed worker failure family.
type MailboxAccountProvisioningWorkerErrorKind string

const (
	MailboxAccountProvisioningWorkerErrorInvalid        MailboxAccountProvisioningWorkerErrorKind = "invalid"
	MailboxAccountProvisioningWorkerErrorTooLarge       MailboxAccountProvisioningWorkerErrorKind = "too_large"
	MailboxAccountProvisioningWorkerErrorUnavailable    MailboxAccountProvisioningWorkerErrorKind = "unavailable"
	MailboxAccountProvisioningWorkerErrorPartialFailure MailboxAccountProvisioningWorkerErrorKind = "partial_failure"
)

var (
	ErrMailboxAccountProvisioningWorkerInvalid        = errors.New("mailbox account provisioning worker invalid")
	ErrMailboxAccountProvisioningWorkerTooLarge       = errors.New("mailbox account provisioning worker exceeds bounds")
	ErrMailboxAccountProvisioningWorkerUnavailable    = errors.New("mailbox account provisioning worker unavailable")
	ErrMailboxAccountProvisioningWorkerPartialFailure = errors.New("mailbox account provisioning worker partial failure")
)

// MailboxAccountProvisioningWorkerError retains only a safe error kind.
type MailboxAccountProvisioningWorkerError struct {
	kind MailboxAccountProvisioningWorkerErrorKind
}

func (errorValue *MailboxAccountProvisioningWorkerError) Error() string {
	switch errorValue.Kind() {
	case MailboxAccountProvisioningWorkerErrorInvalid:
		return ErrMailboxAccountProvisioningWorkerInvalid.Error()
	case MailboxAccountProvisioningWorkerErrorTooLarge:
		return ErrMailboxAccountProvisioningWorkerTooLarge.Error()
	case MailboxAccountProvisioningWorkerErrorUnavailable:
		return ErrMailboxAccountProvisioningWorkerUnavailable.Error()
	case MailboxAccountProvisioningWorkerErrorPartialFailure:
		return ErrMailboxAccountProvisioningWorkerPartialFailure.Error()
	default:
		return "mailbox account provisioning worker error"
	}
}

func (errorValue *MailboxAccountProvisioningWorkerError) Unwrap() error {
	switch errorValue.Kind() {
	case MailboxAccountProvisioningWorkerErrorInvalid:
		return ErrMailboxAccountProvisioningWorkerInvalid
	case MailboxAccountProvisioningWorkerErrorTooLarge:
		return ErrMailboxAccountProvisioningWorkerTooLarge
	case MailboxAccountProvisioningWorkerErrorUnavailable:
		return ErrMailboxAccountProvisioningWorkerUnavailable
	case MailboxAccountProvisioningWorkerErrorPartialFailure:
		return ErrMailboxAccountProvisioningWorkerPartialFailure
	default:
		return nil
	}
}

// Kind returns the safe error classification.
func (errorValue *MailboxAccountProvisioningWorkerError) Kind() MailboxAccountProvisioningWorkerErrorKind {
	if errorValue == nil || !mailboxAccountProvisioningWorkerErrorKindValid(errorValue.kind) {
		return ""
	}
	return errorValue.kind
}

// MailboxAccountProvisioningWorker performs bounded sequential provider work.
type MailboxAccountProvisioningWorker struct {
	repository  MailboxAccountProvisioningRepository
	provisioner MailboxAccountProvisioner
	timer       MailboxAccountProvisioningWorkerTimer
	options     MailboxAccountProvisioningWorkerOptions

	guard   sync.Mutex
	running bool
}

type mailboxAccountProvisioningWorkerRuntimeTimer struct{}

type mailboxAccountProvisioningWorkerContextState uint8

const (
	mailboxAccountProvisioningWorkerContextActive mailboxAccountProvisioningWorkerContextState = iota + 1
	mailboxAccountProvisioningWorkerContextCancelled
	mailboxAccountProvisioningWorkerContextUnavailable
)

var _ MailboxAccountProvisioningWorkerTimer = mailboxAccountProvisioningWorkerRuntimeTimer{}

// NewMailboxAccountProvisioningWorker constructs a disabled zero worker or a
// bounded enabled worker without repository or provider I/O.
func NewMailboxAccountProvisioningWorker(repository MailboxAccountProvisioningRepository, provisioner MailboxAccountProvisioner, timer MailboxAccountProvisioningWorkerTimer, options MailboxAccountProvisioningWorkerOptions) (*MailboxAccountProvisioningWorker, error) {
	if !options.Enabled {
		if options != (MailboxAccountProvisioningWorkerOptions{}) || !mailboxAccountProvisioningWorkerNilDependency(repository) || !mailboxAccountProvisioningWorkerNilDependency(provisioner) || !mailboxAccountProvisioningWorkerNilDependency(timer) {
			return nil, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorInvalid)
		}
		return &MailboxAccountProvisioningWorker{}, nil
	}
	if mailboxAccountProvisioningWorkerNilDependency(repository) || mailboxAccountProvisioningWorkerNilDependency(provisioner) || mailboxAccountProvisioningWorkerNilDependency(timer) {
		return nil, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorInvalid)
	}
	configured, kind := mailboxAccountProvisioningWorkerOptionsConfigured(options)
	if kind != "" {
		return nil, mailboxAccountProvisioningWorkerError(kind)
	}
	return &MailboxAccountProvisioningWorker{repository: repository, provisioner: provisioner, timer: timer, options: configured}, nil
}

// NewMailboxAccountProvisioningWorkerTimer supplies the production timer seam.
func NewMailboxAccountProvisioningWorkerTimer() MailboxAccountProvisioningWorkerTimer {
	return mailboxAccountProvisioningWorkerRuntimeTimer{}
}

func (mailboxAccountProvisioningWorkerRuntimeTimer) Now() time.Time {
	return time.Now().UTC()
}

func (mailboxAccountProvisioningWorkerRuntimeTimer) Wait(ctx context.Context, duration time.Duration) error {
	if mailboxAccountProvisioningWorkerNilDependency(ctx) {
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

func (mailboxAccountProvisioningWorkerRuntimeTimer) WithDeadline(ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	if mailboxAccountProvisioningWorkerNilDependency(ctx) || deadline.IsZero() {
		return nil, nil
	}
	return context.WithDeadline(ctx, deadline)
}

// RunOnce processes at most one normalized due batch. It never exposes an
// account identity or dependency error to callers.
func (worker *MailboxAccountProvisioningWorker) RunOnce(ctx context.Context) (result MailboxAccountProvisioningWorkerRunResult, resultErr error) {
	if worker == nil || mailboxAccountProvisioningWorkerNilDependency(ctx) {
		return MailboxAccountProvisioningWorkerRunResult{}, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorInvalid)
	}
	if !worker.enter() {
		return MailboxAccountProvisioningWorkerRunResult{}, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorInvalid)
	}
	defer func() {
		worker.leave()
		if recover() != nil {
			result = MailboxAccountProvisioningWorkerRunResult{}
			resultErr = mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorUnavailable)
		}
	}()
	return worker.runOnce(ctx)
}

// Run repeats completed retryable cycles and stops normally on parent cancellation.
func (worker *MailboxAccountProvisioningWorker) Run(ctx context.Context) (resultErr error) {
	if worker == nil || mailboxAccountProvisioningWorkerNilDependency(ctx) {
		return mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorInvalid)
	}
	if !worker.enter() {
		return mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorInvalid)
	}
	defer func() {
		worker.leave()
		if recover() != nil {
			resultErr = mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorUnavailable)
		}
	}()
	if !worker.options.Enabled {
		return nil
	}
	for {
		switch mailboxAccountProvisioningWorkerParentState(ctx) {
		case mailboxAccountProvisioningWorkerContextCancelled:
			return nil
		case mailboxAccountProvisioningWorkerContextUnavailable:
			return mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorUnavailable)
		}
		_, runErr := worker.runOnce(ctx)
		switch mailboxAccountProvisioningWorkerParentState(ctx) {
		case mailboxAccountProvisioningWorkerContextCancelled:
			return nil
		case mailboxAccountProvisioningWorkerContextUnavailable:
			return mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorUnavailable)
		}
		kind := mailboxAccountProvisioningWorkerErrorKindOf(runErr)
		if kind == MailboxAccountProvisioningWorkerErrorInvalid || kind == MailboxAccountProvisioningWorkerErrorTooLarge {
			return runErr
		}
		if err := worker.wait(ctx, worker.options.ScanInterval); err != nil {
			return err
		}
	}
}

func (worker *MailboxAccountProvisioningWorker) runOnce(ctx context.Context) (MailboxAccountProvisioningWorkerRunResult, error) {
	switch mailboxAccountProvisioningWorkerParentState(ctx) {
	case mailboxAccountProvisioningWorkerContextCancelled:
		return MailboxAccountProvisioningWorkerRunResult{}, nil
	case mailboxAccountProvisioningWorkerContextUnavailable:
		return MailboxAccountProvisioningWorkerRunResult{}, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorUnavailable)
	}
	if !worker.options.Enabled {
		return MailboxAccountProvisioningWorkerRunResult{}, nil
	}

	scan, scanErr, scanPanicked := worker.safeScan(ctx)
	if scanPanicked || scanErr != nil {
		return MailboxAccountProvisioningWorkerRunResult{}, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorUnavailable)
	}
	ids, idsErr := mailboxAccountProvisioningWorkerDueIDs(scan, worker.options.BatchSize)
	if idsErr != nil {
		return MailboxAccountProvisioningWorkerRunResult{}, idsErr
	}

	result := MailboxAccountProvisioningWorkerRunResult{Due: len(ids)}
	for _, accountID := range ids {
		if mailboxAccountProvisioningWorkerParentState(ctx) != mailboxAccountProvisioningWorkerContextActive {
			return result, worker.resultError(result)
		}
		claimResult, claimErr, claimPanicked := worker.safeClaim(ctx, MailboxAccountProvisioningClaimRequest{AccountID: accountID})
		if claimPanicked || claimErr != nil || !claimResult.Valid() {
			return result, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorUnavailable)
		}
		switch claimResult.Outcome {
		case MailboxAccountProvisioningNoop:
			result.Noop++
			continue
		case MailboxAccountProvisioningLocalGate:
			result.LocalGated++
			continue
		case MailboxAccountProvisioningClaimed:
			if !mailboxAccountProvisioningWorkerUsableClaim(claimResult.Claim, accountID) {
				return result, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorUnavailable)
			}
		default:
			return result, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorUnavailable)
		}

		result.Claimed++
		request, providerSucceeded := worker.providerFinalizeRequest(ctx, claimResult.Claim)
		if providerSucceeded {
			result.ProviderSucceeded++
		} else {
			result.ProviderFailed++
		}
		worker.finalizeClaim(ctx, claimResult.Claim, request, &result)
	}
	return result, worker.resultError(result)
}

func (worker *MailboxAccountProvisioningWorker) providerFinalizeRequest(parent context.Context, claim MailboxAccountProvisioningClaim) (MailboxAccountProvisioningFinalizeRequest, bool) {
	retry := func(code MailboxAccountProvisioningCode) MailboxAccountProvisioningFinalizeRequest {
		request, ok := NewMailboxAccountProvisioningFinalizeRetry(claim, code)
		if !ok {
			return MailboxAccountProvisioningFinalizeRequest{}
		}
		return request
	}
	now, nowOK := worker.safeNow()
	if !nowOK {
		return retry(MailboxAccountProvisioningCodeUnavailable), false
	}
	fence := claim.Fence()
	providerDeadline := mailboxAccountProvisioningWorkerEarlier(now.Add(worker.options.ProviderTimeout), fence.LeaseDeadline.Add(-worker.options.FinalizeTimeout-mailboxAccountProvisioningWorkerLeaseSafetyMargin))
	if !providerDeadline.After(now) {
		return retry(MailboxAccountProvisioningCodeUnavailable), false
	}
	providerContext, cancel, contextPanicked := worker.withDeadline(parent, providerDeadline)
	if contextPanicked || !mailboxAccountProvisioningWorkerValidDeadlineContext(providerContext, providerDeadline) || cancel == nil {
		if cancel != nil {
			_ = mailboxAccountProvisioningWorkerCancel(cancel)
		}
		return retry(MailboxAccountProvisioningCodeUnavailable), false
	}
	if mailboxAccountProvisioningWorkerParentState(providerContext) != mailboxAccountProvisioningWorkerContextActive {
		_ = mailboxAccountProvisioningWorkerCancel(cancel)
		return retry(MailboxAccountProvisioningCodeUnavailable), false
	}
	account, providerErr, providerPanicked := worker.safeEnsure(providerContext, claim.AccountSpec())
	contextState := mailboxAccountProvisioningWorkerParentState(providerContext)
	cancelPanicked := mailboxAccountProvisioningWorkerCancel(cancel)
	if providerPanicked || cancelPanicked || contextState != mailboxAccountProvisioningWorkerContextActive || providerErr != nil {
		if providerPanicked || cancelPanicked || contextState == mailboxAccountProvisioningWorkerContextUnavailable {
			return retry(MailboxAccountProvisioningCodeUnavailable), false
		}
		if contextState == mailboxAccountProvisioningWorkerContextCancelled && providerErr == nil {
			providerErr = context.DeadlineExceeded
		}
		return retry(MailboxAccountProvisioningProviderErrorCode(providerErr)), false
	}
	request, ok := NewMailboxAccountProvisioningFinalizeSuccess(claim, account)
	if !ok {
		return retry(MailboxAccountProvisioningCodeUnavailable), false
	}
	return request, true
}

func (worker *MailboxAccountProvisioningWorker) finalizeClaim(parent context.Context, claim MailboxAccountProvisioningClaim, request MailboxAccountProvisioningFinalizeRequest, result *MailboxAccountProvisioningWorkerRunResult) {
	if result == nil || !request.Valid() {
		if result != nil {
			result.Unfinalized++
		}
		return
	}
	now, nowOK := worker.safeNow()
	if !nowOK {
		result.Unfinalized++
		return
	}
	finalizeDeadline := mailboxAccountProvisioningWorkerEarlier(now.Add(worker.options.FinalizeTimeout), claim.Fence().LeaseDeadline.Add(-mailboxAccountProvisioningWorkerLeaseSafetyMargin))
	if !finalizeDeadline.After(now) {
		result.Unfinalized++
		return
	}
	detached := context.WithoutCancel(parent)
	finalizeContext, cancel, contextPanicked := worker.withDeadline(detached, finalizeDeadline)
	if contextPanicked || !mailboxAccountProvisioningWorkerValidDeadlineContext(finalizeContext, finalizeDeadline) || cancel == nil {
		if cancel != nil {
			_ = mailboxAccountProvisioningWorkerCancel(cancel)
		}
		result.Unfinalized++
		return
	}
	if mailboxAccountProvisioningWorkerParentState(finalizeContext) != mailboxAccountProvisioningWorkerContextActive {
		_ = mailboxAccountProvisioningWorkerCancel(cancel)
		result.Unfinalized++
		return
	}
	finalized, finalizeErr, finalizePanicked := worker.safeFinalize(finalizeContext, request)
	cancelPanicked := mailboxAccountProvisioningWorkerCancel(cancel)
	if finalizePanicked || cancelPanicked || finalizeErr != nil || !mailboxAccountProvisioningWorkerFinalizeMatches(request, finalized) {
		result.Unfinalized++
		return
	}
	switch request.Outcome {
	case MailboxAccountProvisioningFinalizeSuccess:
		result.FinalizedSuccess++
	case MailboxAccountProvisioningFinalizeRetry:
		result.FinalizedRetry++
	default:
		result.Unfinalized++
	}
}

func (worker *MailboxAccountProvisioningWorker) safeScan(ctx context.Context) (result MailboxAccountProvisioningScanResult, resultErr error, panicked bool) {
	defer func() { if recover() != nil { result, resultErr, panicked = MailboxAccountProvisioningScanResult{}, nil, true } }()
	result, resultErr = worker.repository.Scan(ctx, MailboxAccountProvisioningScanRequest{BatchSize: worker.options.BatchSize, ScanInterval: worker.options.ScanInterval})
	return result, resultErr, false
}

func (worker *MailboxAccountProvisioningWorker) safeClaim(ctx context.Context, request MailboxAccountProvisioningClaimRequest) (result MailboxAccountProvisioningClaimResult, resultErr error, panicked bool) {
	defer func() { if recover() != nil { result, resultErr, panicked = MailboxAccountProvisioningClaimResult{}, nil, true } }()
	result, resultErr = worker.repository.ClaimMailboxAccountProvisioning(ctx, request)
	return result, resultErr, false
}

func (worker *MailboxAccountProvisioningWorker) safeEnsure(ctx context.Context, spec AccountSpec) (result EngineAccount, resultErr error, panicked bool) {
	defer func() { if recover() != nil { result, resultErr, panicked = EngineAccount{}, nil, true } }()
	result, resultErr = worker.provisioner.EnsureAccount(ctx, spec)
	return result, resultErr, false
}

func (worker *MailboxAccountProvisioningWorker) safeFinalize(ctx context.Context, request MailboxAccountProvisioningFinalizeRequest) (result MailboxAccountProvisioningFinalizeResult, resultErr error, panicked bool) {
	defer func() { if recover() != nil { result, resultErr, panicked = MailboxAccountProvisioningFinalizeResult{}, nil, true } }()
	result, resultErr = worker.repository.Finalize(ctx, request)
	return result, resultErr, false
}

func (worker *MailboxAccountProvisioningWorker) safeNow() (value time.Time, ok bool) {
	defer func() { if recover() != nil { value, ok = time.Time{}, false } }()
	value = worker.timer.Now().UTC()
	return value, !value.IsZero()
}

func (worker *MailboxAccountProvisioningWorker) withDeadline(ctx context.Context, deadline time.Time) (child context.Context, cancel context.CancelFunc, panicked bool) {
	defer func() { if recover() != nil { child, cancel, panicked = nil, nil, true } }()
	child, cancel = worker.timer.WithDeadline(ctx, deadline)
	return child, cancel, false
}

func (worker *MailboxAccountProvisioningWorker) wait(ctx context.Context, duration time.Duration) (resultErr error) {
	panicked := false
	func() {
		defer func() { if recover() != nil { panicked = true } }()
		resultErr = worker.timer.Wait(ctx, duration)
	}()
	if mailboxAccountProvisioningWorkerParentState(ctx) == mailboxAccountProvisioningWorkerContextCancelled {
		return nil
	}
	if panicked || resultErr != nil || mailboxAccountProvisioningWorkerParentState(ctx) != mailboxAccountProvisioningWorkerContextActive {
		return mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorUnavailable)
	}
	return nil
}

func (worker *MailboxAccountProvisioningWorker) enter() bool {
	worker.guard.Lock()
	defer worker.guard.Unlock()
	if worker.running {
		return false
	}
	worker.running = true
	return true
}

func (worker *MailboxAccountProvisioningWorker) leave() {
	worker.guard.Lock()
	worker.running = false
	worker.guard.Unlock()
}

func (worker *MailboxAccountProvisioningWorker) resultError(result MailboxAccountProvisioningWorkerRunResult) error {
	if result.Unfinalized != 0 {
		return mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorPartialFailure)
	}
	return nil
}

func mailboxAccountProvisioningWorkerDueIDs(scan MailboxAccountProvisioningScanResult, batchSize int) ([]uuid.UUID, error) {
	if batchSize < 1 {
		return nil, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorInvalid)
	}
	if len(scan.AccountIDs) > batchSize {
		return nil, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorTooLarge)
	}
	seen := make(map[uuid.UUID]struct{}, len(scan.AccountIDs))
	for _, accountID := range scan.AccountIDs {
		if accountID == uuid.Nil {
			return nil, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorInvalid)
		}
		if _, exists := seen[accountID]; exists {
			return nil, mailboxAccountProvisioningWorkerError(MailboxAccountProvisioningWorkerErrorInvalid)
		}
		seen[accountID] = struct{}{}
	}
	return append([]uuid.UUID(nil), scan.AccountIDs...), nil
}

func mailboxAccountProvisioningWorkerUsableClaim(claim MailboxAccountProvisioningClaim, scannedID uuid.UUID) bool {
	fence := claim.Fence()
	if !claim.Valid() || fence.AccountID != scannedID || !mailboxAccountProvisioningWorkerControlHashValid(fence.ControlHash) {
		return false
	}
	spec := claim.AccountSpec()
	return spec.ProvisioningKey == fence.ProvisioningKey && spec.LocalPart == fence.LocalPart && spec.Domain == fence.Domain && spec.DesiredGeneration == fence.ControlGeneration && spec.DesiredHash == fence.ControlHash && spec.Enabled && spec.Primary == (AccountIdentitySpec{LocalPart: fence.LocalPart, Domain: fence.Domain, Enabled: true}) && len(spec.Aliases) == 0
}

func mailboxAccountProvisioningWorkerControlHashValid(value string) bool {
	if len(value) != 64 {
		return false
	}
	for index := 0; index < len(value); index++ {
		if !(value[index] >= '0' && value[index] <= '9' || value[index] >= 'a' && value[index] <= 'f') {
			return false
		}
	}
	return true
}

func mailboxAccountProvisioningWorkerFinalizeMatches(request MailboxAccountProvisioningFinalizeRequest, result MailboxAccountProvisioningFinalizeResult) bool {
	if !request.Valid() || !result.Valid() || result.Outcome != request.Outcome {
		return false
	}
	return result.Code == request.Code
}

func mailboxAccountProvisioningWorkerOptionsConfigured(options MailboxAccountProvisioningWorkerOptions) (MailboxAccountProvisioningWorkerOptions, MailboxAccountProvisioningWorkerErrorKind) {
	defaults := DefaultMailboxAccountProvisioningWorkerOptions()
	if options.ScanInterval == 0 { options.ScanInterval = defaults.ScanInterval }
	if options.ProviderTimeout == 0 { options.ProviderTimeout = defaults.ProviderTimeout }
	if options.FinalizeTimeout == 0 { options.FinalizeTimeout = defaults.FinalizeTimeout }
	if options.BatchSize == 0 { options.BatchSize = defaults.BatchSize }
	if options.ScanInterval > mailboxAccountProvisioningWorkerMaximumScanInterval || options.ProviderTimeout > mailboxAccountProvisioningWorkerMaximumProviderTimeout || options.FinalizeTimeout > mailboxAccountProvisioningWorkerMaximumFinalizeTimeout || options.BatchSize > MailboxAccountProvisioningMaximumBatchSize || options.ProviderTimeout+options.FinalizeTimeout+mailboxAccountProvisioningWorkerLeaseSafetyMargin > MailboxEngineCommandLeaseDuration {
		return MailboxAccountProvisioningWorkerOptions{}, MailboxAccountProvisioningWorkerErrorTooLarge
	}
	if options.ScanInterval < time.Second || options.ProviderTimeout < time.Second || options.FinalizeTimeout < time.Second || options.BatchSize < 1 || options.ScanInterval%time.Second != 0 || options.ProviderTimeout%time.Second != 0 || options.FinalizeTimeout%time.Second != 0 {
		return MailboxAccountProvisioningWorkerOptions{}, MailboxAccountProvisioningWorkerErrorInvalid
	}
	return options, ""
}

func mailboxAccountProvisioningWorkerError(kind MailboxAccountProvisioningWorkerErrorKind) error {
	if !mailboxAccountProvisioningWorkerErrorKindValid(kind) { kind = MailboxAccountProvisioningWorkerErrorInvalid }
	return &MailboxAccountProvisioningWorkerError{kind: kind}
}

func mailboxAccountProvisioningWorkerErrorKindValid(kind MailboxAccountProvisioningWorkerErrorKind) bool {
	switch kind {
	case MailboxAccountProvisioningWorkerErrorInvalid, MailboxAccountProvisioningWorkerErrorTooLarge, MailboxAccountProvisioningWorkerErrorUnavailable, MailboxAccountProvisioningWorkerErrorPartialFailure:
		return true
	default:
		return false
	}
}

func mailboxAccountProvisioningWorkerErrorKindOf(err error) MailboxAccountProvisioningWorkerErrorKind {
	var workerError *MailboxAccountProvisioningWorkerError
	if errors.As(err, &workerError) { return workerError.Kind() }
	return ""
}

func mailboxAccountProvisioningWorkerNilDependency(dependency any) bool {
	if dependency == nil { return true }
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

func mailboxAccountProvisioningWorkerParentState(ctx context.Context) (state mailboxAccountProvisioningWorkerContextState) {
	defer func() { if recover() != nil { state = mailboxAccountProvisioningWorkerContextUnavailable } }()
	if mailboxAccountProvisioningWorkerNilDependency(ctx) { return mailboxAccountProvisioningWorkerContextUnavailable }
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) { return mailboxAccountProvisioningWorkerContextCancelled }
		return mailboxAccountProvisioningWorkerContextUnavailable
	}
	return mailboxAccountProvisioningWorkerContextActive
}

func mailboxAccountProvisioningWorkerValidDeadlineContext(ctx context.Context, requested time.Time) (valid bool) {
	defer func() { if recover() != nil { valid = false } }()
	if mailboxAccountProvisioningWorkerNilDependency(ctx) { return false }
	deadline, ok := ctx.Deadline()
	return ok && !deadline.After(requested)
}

func mailboxAccountProvisioningWorkerCancel(cancel context.CancelFunc) (panicked bool) {
	defer func() { if recover() != nil { panicked = true } }()
	cancel()
	return false
}

func mailboxAccountProvisioningWorkerEarlier(left, right time.Time) time.Time {
	if right.Before(left) { return right }
	return left
}
