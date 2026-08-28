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
	mailboxDomainProvisioningWorkerDefaultProviderTimeout = 30 * time.Second
	mailboxDomainProvisioningWorkerDefaultFinalizeTimeout = 10 * time.Second
	mailboxDomainProvisioningWorkerMaximumScanInterval    = 5 * time.Minute
	mailboxDomainProvisioningWorkerMaximumProviderTimeout = 30 * time.Second
	mailboxDomainProvisioningWorkerMaximumFinalizeTimeout = 10 * time.Second
	mailboxDomainProvisioningWorkerLeaseSafetyMargin      = time.Second
)

// MailboxDomainProvisioningWorkerTimer supplies bounded worker time seams.
type MailboxDomainProvisioningWorkerTimer interface {
	Now() time.Time
	Wait(context.Context, time.Duration) error
	WithDeadline(context.Context, time.Time) (context.Context, context.CancelFunc)
}

// MailboxDomainProvisioningWorkerOptions bounds one sequential worker.
type MailboxDomainProvisioningWorkerOptions struct {
	Enabled         bool
	ScanInterval    time.Duration
	ProviderTimeout time.Duration
	FinalizeTimeout time.Duration
	BatchSize       int
}

// DefaultMailboxDomainProvisioningWorkerOptions returns conservative enabled options.
func DefaultMailboxDomainProvisioningWorkerOptions() MailboxDomainProvisioningWorkerOptions {
	return MailboxDomainProvisioningWorkerOptions{
		Enabled:         true,
		ScanInterval:    MailboxDomainProvisioningDefaultScanInterval,
		ProviderTimeout: mailboxDomainProvisioningWorkerDefaultProviderTimeout,
		FinalizeTimeout: mailboxDomainProvisioningWorkerDefaultFinalizeTimeout,
		BatchSize:       MailboxDomainProvisioningDefaultBatchSize,
	}
}

// MailboxDomainProvisioningWorkerRunResult contains only bounded work counts.
type MailboxDomainProvisioningWorkerRunResult struct {
	Due                 int
	Claimed             int
	Noop                int
	ProviderSucceeded   int
	ProviderFailed      int
	FinalizedDNSRequired int
	FinalizedVerified   int
	FinalizedRetry      int
	FinalizedSuperseded int
	Unfinalized         int
}

// MailboxDomainProvisioningWorkerErrorKind is the closed worker failure family.
type MailboxDomainProvisioningWorkerErrorKind string

const (
	MailboxDomainProvisioningWorkerErrorInvalid        MailboxDomainProvisioningWorkerErrorKind = "invalid"
	MailboxDomainProvisioningWorkerErrorTooLarge       MailboxDomainProvisioningWorkerErrorKind = "too_large"
	MailboxDomainProvisioningWorkerErrorUnavailable    MailboxDomainProvisioningWorkerErrorKind = "unavailable"
	MailboxDomainProvisioningWorkerErrorPartialFailure MailboxDomainProvisioningWorkerErrorKind = "partial_failure"
)

var (
	ErrMailboxDomainProvisioningWorkerInvalid        = errors.New("mailbox domain provisioning worker invalid")
	ErrMailboxDomainProvisioningWorkerTooLarge       = errors.New("mailbox domain provisioning worker exceeds bounds")
	ErrMailboxDomainProvisioningWorkerUnavailable    = errors.New("mailbox domain provisioning worker unavailable")
	ErrMailboxDomainProvisioningWorkerPartialFailure = errors.New("mailbox domain provisioning worker partial failure")
)

// MailboxDomainProvisioningWorkerError retains only a safe error kind.
type MailboxDomainProvisioningWorkerError struct {
	kind MailboxDomainProvisioningWorkerErrorKind
}

func (errorValue *MailboxDomainProvisioningWorkerError) Error() string {
	switch errorValue.Kind() {
	case MailboxDomainProvisioningWorkerErrorInvalid:
		return ErrMailboxDomainProvisioningWorkerInvalid.Error()
	case MailboxDomainProvisioningWorkerErrorTooLarge:
		return ErrMailboxDomainProvisioningWorkerTooLarge.Error()
	case MailboxDomainProvisioningWorkerErrorUnavailable:
		return ErrMailboxDomainProvisioningWorkerUnavailable.Error()
	case MailboxDomainProvisioningWorkerErrorPartialFailure:
		return ErrMailboxDomainProvisioningWorkerPartialFailure.Error()
	default:
		return "mailbox domain provisioning worker error"
	}
}

func (errorValue *MailboxDomainProvisioningWorkerError) Unwrap() error {
	switch errorValue.Kind() {
	case MailboxDomainProvisioningWorkerErrorInvalid:
		return ErrMailboxDomainProvisioningWorkerInvalid
	case MailboxDomainProvisioningWorkerErrorTooLarge:
		return ErrMailboxDomainProvisioningWorkerTooLarge
	case MailboxDomainProvisioningWorkerErrorUnavailable:
		return ErrMailboxDomainProvisioningWorkerUnavailable
	case MailboxDomainProvisioningWorkerErrorPartialFailure:
		return ErrMailboxDomainProvisioningWorkerPartialFailure
	default:
		return nil
	}
}

// Kind returns the safe error classification.
func (errorValue *MailboxDomainProvisioningWorkerError) Kind() MailboxDomainProvisioningWorkerErrorKind {
	if errorValue == nil || !mailboxDomainProvisioningWorkerErrorKindValid(errorValue.kind) {
		return ""
	}
	return errorValue.kind
}

// MailboxDomainProvisioningWorker performs bounded sequential provider work.
type MailboxDomainProvisioningWorker struct {
	repository  MailboxDomainProvisioningRepository
	provisioner MailboxDomainProvisioner
	timer       MailboxDomainProvisioningWorkerTimer
	options     MailboxDomainProvisioningWorkerOptions

	guard   sync.Mutex
	running bool
}

type mailboxDomainProvisioningWorkerRuntimeTimer struct{}

type mailboxDomainProvisioningWorkerContextState uint8

const (
	mailboxDomainProvisioningWorkerContextActive mailboxDomainProvisioningWorkerContextState = iota + 1
	mailboxDomainProvisioningWorkerContextCancelled
	mailboxDomainProvisioningWorkerContextUnavailable
)

var _ MailboxDomainProvisioningWorkerTimer = mailboxDomainProvisioningWorkerRuntimeTimer{}

// NewMailboxDomainProvisioningWorker constructs a disabled zero worker or a
// bounded enabled worker without repository or provider I/O.
func NewMailboxDomainProvisioningWorker(repository MailboxDomainProvisioningRepository, provisioner MailboxDomainProvisioner, timer MailboxDomainProvisioningWorkerTimer, options MailboxDomainProvisioningWorkerOptions) (*MailboxDomainProvisioningWorker, error) {
	if !options.Enabled {
		if options != (MailboxDomainProvisioningWorkerOptions{}) ||
			!mailboxDomainProvisioningWorkerNilDependency(repository) ||
			!mailboxDomainProvisioningWorkerNilDependency(provisioner) ||
			!mailboxDomainProvisioningWorkerNilDependency(timer) {
			return nil, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorInvalid)
		}
		return &MailboxDomainProvisioningWorker{}, nil
	}
	if mailboxDomainProvisioningWorkerNilDependency(repository) ||
		mailboxDomainProvisioningWorkerNilDependency(provisioner) ||
		mailboxDomainProvisioningWorkerNilDependency(timer) {
		return nil, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorInvalid)
	}
	configured, kind := mailboxDomainProvisioningWorkerOptionsConfigured(options)
	if kind != "" {
		return nil, mailboxDomainProvisioningWorkerError(kind)
	}
	return &MailboxDomainProvisioningWorker{
		repository:  repository,
		provisioner: provisioner,
		timer:       timer,
		options:     configured,
	}, nil
}

// NewMailboxDomainProvisioningWorkerTimer supplies the production timer seam.
func NewMailboxDomainProvisioningWorkerTimer() MailboxDomainProvisioningWorkerTimer {
	return mailboxDomainProvisioningWorkerRuntimeTimer{}
}

func (mailboxDomainProvisioningWorkerRuntimeTimer) Now() time.Time {
	return time.Now().UTC()
}

func (mailboxDomainProvisioningWorkerRuntimeTimer) Wait(ctx context.Context, duration time.Duration) error {
	if mailboxDomainProvisioningWorkerNilDependency(ctx) {
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

func (mailboxDomainProvisioningWorkerRuntimeTimer) WithDeadline(ctx context.Context, deadline time.Time) (context.Context, context.CancelFunc) {
	if mailboxDomainProvisioningWorkerNilDependency(ctx) || deadline.IsZero() {
		return nil, nil
	}
	return context.WithDeadline(ctx, deadline)
}

// RunOnce processes at most one normalized due batch. It never exposes a
// domain identity or dependency error to callers.
func (worker *MailboxDomainProvisioningWorker) RunOnce(ctx context.Context) (result MailboxDomainProvisioningWorkerRunResult, resultErr error) {
	if worker == nil || mailboxDomainProvisioningWorkerNilDependency(ctx) {
		return MailboxDomainProvisioningWorkerRunResult{}, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorInvalid)
	}
	if !worker.enter() {
		return MailboxDomainProvisioningWorkerRunResult{}, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorInvalid)
	}
	defer func() {
		worker.leave()
		if recover() != nil {
			result = MailboxDomainProvisioningWorkerRunResult{}
			resultErr = mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorUnavailable)
		}
	}()
	return worker.runOnce(ctx)
}

// Run repeats completed retryable cycles and stops normally on parent cancellation.
func (worker *MailboxDomainProvisioningWorker) Run(ctx context.Context) (resultErr error) {
	if worker == nil || mailboxDomainProvisioningWorkerNilDependency(ctx) {
		return mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorInvalid)
	}
	if !worker.enter() {
		return mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorInvalid)
	}
	defer func() {
		worker.leave()
		if recover() != nil {
			resultErr = mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorUnavailable)
		}
	}()
	if !worker.options.Enabled {
		return nil
	}
	for {
		switch mailboxDomainProvisioningWorkerParentState(ctx) {
		case mailboxDomainProvisioningWorkerContextCancelled:
			return nil
		case mailboxDomainProvisioningWorkerContextUnavailable:
			return mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorUnavailable)
		}
		_, runErr := worker.runOnce(ctx)
		switch mailboxDomainProvisioningWorkerParentState(ctx) {
		case mailboxDomainProvisioningWorkerContextCancelled:
			return nil
		case mailboxDomainProvisioningWorkerContextUnavailable:
			return mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorUnavailable)
		}
		kind := mailboxDomainProvisioningWorkerErrorKindOf(runErr)
		if kind == MailboxDomainProvisioningWorkerErrorInvalid || kind == MailboxDomainProvisioningWorkerErrorTooLarge {
			return runErr
		}
		if err := worker.wait(ctx, worker.options.ScanInterval); err != nil {
			return err
		}
	}
}

func (worker *MailboxDomainProvisioningWorker) runOnce(ctx context.Context) (MailboxDomainProvisioningWorkerRunResult, error) {
	switch mailboxDomainProvisioningWorkerParentState(ctx) {
	case mailboxDomainProvisioningWorkerContextCancelled:
		return MailboxDomainProvisioningWorkerRunResult{}, nil
	case mailboxDomainProvisioningWorkerContextUnavailable:
		return MailboxDomainProvisioningWorkerRunResult{}, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorUnavailable)
	}
	if !worker.options.Enabled {
		return MailboxDomainProvisioningWorkerRunResult{}, nil
	}

	scan, scanErr, scanPanicked := worker.safeScan(ctx)
	if scanPanicked || scanErr != nil {
		return MailboxDomainProvisioningWorkerRunResult{}, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorUnavailable)
	}
	ids, idsErr := mailboxDomainProvisioningWorkerDueIDs(scan, worker.options.BatchSize)
	if idsErr != nil {
		return MailboxDomainProvisioningWorkerRunResult{}, idsErr
	}

	result := MailboxDomainProvisioningWorkerRunResult{Due: len(ids)}
	for _, domainID := range ids {
		if mailboxDomainProvisioningWorkerParentState(ctx) != mailboxDomainProvisioningWorkerContextActive {
			return result, worker.resultError(result)
		}
		claimResult, claimErr, claimPanicked := worker.safeClaim(ctx, MailboxDomainProvisioningClaimRequest{DomainID: domainID})
		if claimPanicked || claimErr != nil || !claimResult.Valid() {
			return result, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorUnavailable)
		}
		switch claimResult.Outcome {
		case MailboxDomainProvisioningNoop:
			result.Noop++
			continue
		case MailboxDomainProvisioningClaimed:
			if !mailboxDomainProvisioningWorkerUsableClaim(claimResult.Claim, domainID) {
				return result, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorUnavailable)
			}
		default:
			return result, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorUnavailable)
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

func (worker *MailboxDomainProvisioningWorker) providerFinalizeRequest(parent context.Context, claim MailboxDomainProvisioningClaim) (MailboxDomainProvisioningFinalizeRequest, bool) {
	retry := func(code MailboxDomainProvisioningCode) MailboxDomainProvisioningFinalizeRequest {
		request, ok := NewMailboxDomainProvisioningFinalizeRetry(claim, code)
		if !ok {
			return MailboxDomainProvisioningFinalizeRequest{}
		}
		return request
	}
	now, nowOK := worker.safeNow()
	if !nowOK {
		return retry(MailboxDomainProvisioningCodeOutcomeUnknown), false
	}
	fence := claim.Fence()
	providerDeadline := mailboxDomainProvisioningWorkerEarlier(
		now.Add(worker.options.ProviderTimeout),
		fence.LeaseDeadline.Add(-worker.options.FinalizeTimeout-mailboxDomainProvisioningWorkerLeaseSafetyMargin),
	)
	if !providerDeadline.After(now) {
		return retry(MailboxDomainProvisioningCodeOutcomeUnknown), false
	}
	providerContext, cancel, contextPanicked := worker.withDeadline(parent, providerDeadline)
	if contextPanicked || !mailboxDomainProvisioningWorkerValidDeadlineContext(providerContext, providerDeadline) || cancel == nil {
		if cancel != nil {
			_ = mailboxDomainProvisioningWorkerCancel(cancel)
		}
		return retry(MailboxDomainProvisioningCodeOutcomeUnknown), false
	}
	if mailboxDomainProvisioningWorkerParentState(providerContext) != mailboxDomainProvisioningWorkerContextActive {
		_ = mailboxDomainProvisioningWorkerCancel(cancel)
		return retry(MailboxDomainProvisioningCodeEngineUnavailable), false
	}
	request, executed, executePanicked := worker.safeExecute(providerContext, claim)
	contextState := mailboxDomainProvisioningWorkerParentState(providerContext)
	cancelPanicked := mailboxDomainProvisioningWorkerCancel(cancel)
	if executePanicked || cancelPanicked || contextState == mailboxDomainProvisioningWorkerContextUnavailable {
		return retry(MailboxDomainProvisioningCodeOutcomeUnknown), false
	}
	if !executed || !request.Valid() {
		return retry(MailboxDomainProvisioningCodeOutcomeUnknown), false
	}
	switch request.Outcome {
	case MailboxDomainProvisioningFinalizeDNSRequired, MailboxDomainProvisioningFinalizeVerified:
		return request, true
	case MailboxDomainProvisioningFinalizeRetry:
		return request, false
	default:
		return retry(MailboxDomainProvisioningCodeOutcomeUnknown), false
	}
}

func (worker *MailboxDomainProvisioningWorker) finalizeClaim(parent context.Context, claim MailboxDomainProvisioningClaim, request MailboxDomainProvisioningFinalizeRequest, result *MailboxDomainProvisioningWorkerRunResult) {
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
	finalizeDeadline := mailboxDomainProvisioningWorkerEarlier(
		now.Add(worker.options.FinalizeTimeout),
		claim.Fence().LeaseDeadline.Add(-mailboxDomainProvisioningWorkerLeaseSafetyMargin),
	)
	if !finalizeDeadline.After(now) {
		result.Unfinalized++
		return
	}
	detached := context.WithoutCancel(parent)
	finalizeContext, cancel, contextPanicked := worker.withDeadline(detached, finalizeDeadline)
	if contextPanicked || !mailboxDomainProvisioningWorkerValidDeadlineContext(finalizeContext, finalizeDeadline) || cancel == nil {
		if cancel != nil {
			_ = mailboxDomainProvisioningWorkerCancel(cancel)
		}
		result.Unfinalized++
		return
	}
	if mailboxDomainProvisioningWorkerParentState(finalizeContext) != mailboxDomainProvisioningWorkerContextActive {
		_ = mailboxDomainProvisioningWorkerCancel(cancel)
		result.Unfinalized++
		return
	}
	finalized, finalizeErr, finalizePanicked := worker.safeFinalize(finalizeContext, request)
	cancelPanicked := mailboxDomainProvisioningWorkerCancel(cancel)
	if finalizePanicked || cancelPanicked || finalizeErr != nil || !mailboxDomainProvisioningWorkerFinalizeMatches(request, finalized) {
		result.Unfinalized++
		return
	}
	switch finalized.Outcome {
	case MailboxDomainProvisioningFinalizeDNSRequired:
		result.FinalizedDNSRequired++
	case MailboxDomainProvisioningFinalizeVerified:
		result.FinalizedVerified++
	case MailboxDomainProvisioningFinalizeRetry:
		result.FinalizedRetry++
	case MailboxDomainProvisioningFinalizeSuperseded:
		result.FinalizedSuperseded++
	default:
		result.Unfinalized++
	}
}

func (worker *MailboxDomainProvisioningWorker) safeScan(ctx context.Context) (result MailboxDomainProvisioningScanResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result, resultErr, panicked = MailboxDomainProvisioningScanResult{}, nil, true
		}
	}()
	result, resultErr = worker.repository.ScanMailboxDomainProvisioning(ctx, MailboxDomainProvisioningScanRequest{
		BatchSize:    worker.options.BatchSize,
		ScanInterval: worker.options.ScanInterval,
	})
	return result, resultErr, false
}

func (worker *MailboxDomainProvisioningWorker) safeClaim(ctx context.Context, request MailboxDomainProvisioningClaimRequest) (result MailboxDomainProvisioningClaimResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result, resultErr, panicked = MailboxDomainProvisioningClaimResult{}, nil, true
		}
	}()
	result, resultErr = worker.repository.ClaimMailboxDomainProvisioning(ctx, request)
	return result, resultErr, false
}

func (worker *MailboxDomainProvisioningWorker) safeExecute(ctx context.Context, claim MailboxDomainProvisioningClaim) (request MailboxDomainProvisioningFinalizeRequest, ok bool, panicked bool) {
	defer func() {
		if recover() != nil {
			request, ok, panicked = MailboxDomainProvisioningFinalizeRequest{}, false, true
		}
	}()
	request, ok = ExecuteMailboxDomainProvisioning(ctx, worker.provisioner, claim)
	return request, ok, false
}

func (worker *MailboxDomainProvisioningWorker) safeFinalize(ctx context.Context, request MailboxDomainProvisioningFinalizeRequest) (result MailboxDomainProvisioningFinalizeResult, resultErr error, panicked bool) {
	defer func() {
		if recover() != nil {
			result, resultErr, panicked = MailboxDomainProvisioningFinalizeResult{}, nil, true
		}
	}()
	result, resultErr = worker.repository.FinalizeMailboxDomainProvisioning(ctx, request)
	return result, resultErr, false
}

func (worker *MailboxDomainProvisioningWorker) safeNow() (value time.Time, ok bool) {
	defer func() {
		if recover() != nil {
			value, ok = time.Time{}, false
		}
	}()
	value = worker.timer.Now().UTC()
	return value, !value.IsZero()
}

func (worker *MailboxDomainProvisioningWorker) withDeadline(ctx context.Context, deadline time.Time) (child context.Context, cancel context.CancelFunc, panicked bool) {
	defer func() {
		if recover() != nil {
			child, cancel, panicked = nil, nil, true
		}
	}()
	child, cancel = worker.timer.WithDeadline(ctx, deadline)
	return child, cancel, false
}

func (worker *MailboxDomainProvisioningWorker) wait(ctx context.Context, duration time.Duration) (resultErr error) {
	panicked := false
	func() {
		defer func() {
			if recover() != nil {
				panicked = true
			}
		}()
		resultErr = worker.timer.Wait(ctx, duration)
	}()
	if mailboxDomainProvisioningWorkerParentState(ctx) == mailboxDomainProvisioningWorkerContextCancelled {
		return nil
	}
	if panicked || resultErr != nil || mailboxDomainProvisioningWorkerParentState(ctx) != mailboxDomainProvisioningWorkerContextActive {
		return mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorUnavailable)
	}
	return nil
}

func (worker *MailboxDomainProvisioningWorker) enter() bool {
	worker.guard.Lock()
	defer worker.guard.Unlock()
	if worker.running {
		return false
	}
	worker.running = true
	return true
}

func (worker *MailboxDomainProvisioningWorker) leave() {
	worker.guard.Lock()
	worker.running = false
	worker.guard.Unlock()
}

func (worker *MailboxDomainProvisioningWorker) resultError(result MailboxDomainProvisioningWorkerRunResult) error {
	if result.Unfinalized != 0 || result.ProviderFailed != 0 || result.FinalizedRetry != 0 {
		return mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorPartialFailure)
	}
	return nil
}

func mailboxDomainProvisioningWorkerDueIDs(scan MailboxDomainProvisioningScanResult, batchSize int) ([]uuid.UUID, error) {
	if batchSize < 1 {
		return nil, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorInvalid)
	}
	if len(scan.DomainIDs) > batchSize {
		return nil, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorTooLarge)
	}
	seen := make(map[uuid.UUID]struct{}, len(scan.DomainIDs))
	for _, domainID := range scan.DomainIDs {
		if domainID == uuid.Nil {
			return nil, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorInvalid)
		}
		if _, exists := seen[domainID]; exists {
			return nil, mailboxDomainProvisioningWorkerError(MailboxDomainProvisioningWorkerErrorInvalid)
		}
		seen[domainID] = struct{}{}
	}
	return append([]uuid.UUID(nil), scan.DomainIDs...), nil
}

func mailboxDomainProvisioningWorkerUsableClaim(claim MailboxDomainProvisioningClaim, scannedID uuid.UUID) bool {
	fence := claim.Fence()
	if !claim.Valid() ||
		fence.DomainID != scannedID ||
		!mailboxDomainProvisioningWorkerDesiredHashValid(fence.DesiredHash) ||
		!fence.DesiredPresent ||
		!fence.DesiredEnabled {
		return false
	}
	spec := claim.DomainSpec()
	return spec.ProvisioningKey == fence.ProvisioningKey &&
		spec.Domain == fence.Domain &&
		spec.DesiredGeneration == fence.DesiredGeneration &&
		spec.DesiredHash == fence.DesiredHash &&
		spec.Present &&
		spec.Enabled
}

func mailboxDomainProvisioningWorkerDesiredHashValid(value string) bool {
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

func mailboxDomainProvisioningWorkerFinalizeMatches(request MailboxDomainProvisioningFinalizeRequest, result MailboxDomainProvisioningFinalizeResult) bool {
	if !request.Valid() || !result.Valid() {
		return false
	}
	switch request.Outcome {
	case MailboxDomainProvisioningFinalizeDNSRequired, MailboxDomainProvisioningFinalizeVerified:
		return result.Outcome == request.Outcome || result.Outcome == MailboxDomainProvisioningFinalizeSuperseded
	case MailboxDomainProvisioningFinalizeRetry:
		return (result.Outcome == MailboxDomainProvisioningFinalizeRetry && result.Code == request.Code) ||
			result.Outcome == MailboxDomainProvisioningFinalizeSuperseded
	default:
		return false
	}
}

func mailboxDomainProvisioningWorkerOptionsConfigured(options MailboxDomainProvisioningWorkerOptions) (MailboxDomainProvisioningWorkerOptions, MailboxDomainProvisioningWorkerErrorKind) {
	defaults := DefaultMailboxDomainProvisioningWorkerOptions()
	if options.ScanInterval == 0 {
		options.ScanInterval = defaults.ScanInterval
	}
	if options.ProviderTimeout == 0 {
		options.ProviderTimeout = defaults.ProviderTimeout
	}
	if options.FinalizeTimeout == 0 {
		options.FinalizeTimeout = defaults.FinalizeTimeout
	}
	if options.BatchSize == 0 {
		options.BatchSize = defaults.BatchSize
	}
	if options.ScanInterval > mailboxDomainProvisioningWorkerMaximumScanInterval ||
		options.ProviderTimeout > mailboxDomainProvisioningWorkerMaximumProviderTimeout ||
		options.FinalizeTimeout > mailboxDomainProvisioningWorkerMaximumFinalizeTimeout ||
		options.BatchSize > MailboxDomainProvisioningMaximumBatchSize ||
		options.ProviderTimeout+options.FinalizeTimeout+mailboxDomainProvisioningWorkerLeaseSafetyMargin > MailboxEngineCommandLeaseDuration {
		return MailboxDomainProvisioningWorkerOptions{}, MailboxDomainProvisioningWorkerErrorTooLarge
	}
	if options.ScanInterval < time.Second ||
		options.ProviderTimeout < time.Second ||
		options.FinalizeTimeout < time.Second ||
		options.BatchSize < 1 ||
		options.ScanInterval%time.Second != 0 ||
		options.ProviderTimeout%time.Second != 0 ||
		options.FinalizeTimeout%time.Second != 0 {
		return MailboxDomainProvisioningWorkerOptions{}, MailboxDomainProvisioningWorkerErrorInvalid
	}
	return options, ""
}

func mailboxDomainProvisioningWorkerError(kind MailboxDomainProvisioningWorkerErrorKind) error {
	if !mailboxDomainProvisioningWorkerErrorKindValid(kind) {
		kind = MailboxDomainProvisioningWorkerErrorInvalid
	}
	return &MailboxDomainProvisioningWorkerError{kind: kind}
}

func mailboxDomainProvisioningWorkerErrorKindValid(kind MailboxDomainProvisioningWorkerErrorKind) bool {
	switch kind {
	case MailboxDomainProvisioningWorkerErrorInvalid,
		MailboxDomainProvisioningWorkerErrorTooLarge,
		MailboxDomainProvisioningWorkerErrorUnavailable,
		MailboxDomainProvisioningWorkerErrorPartialFailure:
		return true
	default:
		return false
	}
}

func mailboxDomainProvisioningWorkerErrorKindOf(err error) MailboxDomainProvisioningWorkerErrorKind {
	var workerError *MailboxDomainProvisioningWorkerError
	if errors.As(err, &workerError) {
		return workerError.Kind()
	}
	return ""
}

func mailboxDomainProvisioningWorkerNilDependency(dependency any) bool {
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

func mailboxDomainProvisioningWorkerParentState(ctx context.Context) (state mailboxDomainProvisioningWorkerContextState) {
	defer func() {
		if recover() != nil {
			state = mailboxDomainProvisioningWorkerContextUnavailable
		}
	}()
	if mailboxDomainProvisioningWorkerNilDependency(ctx) {
		return mailboxDomainProvisioningWorkerContextUnavailable
	}
	if err := ctx.Err(); err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return mailboxDomainProvisioningWorkerContextCancelled
		}
		return mailboxDomainProvisioningWorkerContextUnavailable
	}
	return mailboxDomainProvisioningWorkerContextActive
}

func mailboxDomainProvisioningWorkerValidDeadlineContext(ctx context.Context, requested time.Time) (valid bool) {
	defer func() {
		if recover() != nil {
			valid = false
		}
	}()
	if mailboxDomainProvisioningWorkerNilDependency(ctx) {
		return false
	}
	deadline, ok := ctx.Deadline()
	return ok && !deadline.After(requested)
}

func mailboxDomainProvisioningWorkerCancel(cancel context.CancelFunc) (panicked bool) {
	defer func() {
		if recover() != nil {
			panicked = true
		}
	}()
	cancel()
	return false
}

func mailboxDomainProvisioningWorkerEarlier(left, right time.Time) time.Time {
	if right.Before(left) {
		return right
	}
	return left
}
