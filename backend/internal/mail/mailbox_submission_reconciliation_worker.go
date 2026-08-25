package mail

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MailboxSubmissionReconciliationExecutor performs one reconciliation for a
// durable submission identity.
type MailboxSubmissionReconciliationExecutor interface {
	Reconcile(context.Context, uuid.UUID) error
}

var _ MailboxSubmissionReconciliationExecutor = (*MailboxSubmissionReconciler)(nil)

// MailboxSubmissionReconciliationWorkerOptions bounds one sequential scan and
// reconciliation loop. Zero enabled values use the submission worker defaults.
type MailboxSubmissionReconciliationWorkerOptions struct {
	Enabled          bool
	ScanInterval     time.Duration
	ReconcileTimeout time.Duration
	BatchSize        int
}

// DefaultMailboxSubmissionReconciliationWorkerOptions returns conservative
// enabled options matching the submission worker defaults.
func DefaultMailboxSubmissionReconciliationWorkerOptions() MailboxSubmissionReconciliationWorkerOptions {
	defaults := DefaultMailboxSubmissionWorkerOptions()
	return MailboxSubmissionReconciliationWorkerOptions{
		Enabled:          defaults.Enabled,
		ScanInterval:     defaults.ScanInterval,
		ReconcileTimeout: defaults.DispatchTimeout,
		BatchSize:        defaults.BatchSize,
	}
}

// MailboxSubmissionReconciliationWorkerRunResult summarizes one bounded scan
// without exposing durable identities.
type MailboxSubmissionReconciliationWorkerRunResult struct {
	Due        int
	Reconciled int
	Failed     int
}

// MailboxSubmissionReconciliationWorker adapts unknown-submission scans to
// the reusable sequential submission worker.
type MailboxSubmissionReconciliationWorker struct {
	worker *MailboxSubmissionWorker
}

// NewMailboxSubmissionReconciliationWorker constructs a disabled zero worker
// or delegates enabled validation and execution boundaries to the submission
// worker.
func NewMailboxSubmissionReconciliationWorker(repository MailboxSubmissionReconciliationRepository, executor MailboxSubmissionReconciliationExecutor, timer MailboxSubmissionWorkerTimer, options MailboxSubmissionReconciliationWorkerOptions) (*MailboxSubmissionReconciliationWorker, error) {
	worker, err := NewMailboxSubmissionWorker(
		mailboxSubmissionReconciliationWorkerScannerFor(repository),
		mailboxSubmissionReconciliationWorkerDispatcherFor(executor),
		timer,
		mailboxSubmissionReconciliationWorkerOptions(options),
	)
	if err != nil {
		return nil, err
	}
	return &MailboxSubmissionReconciliationWorker{worker: worker}, nil
}

// Run delegates the guarded sequential loop to the submission worker.
func (worker *MailboxSubmissionReconciliationWorker) Run(ctx context.Context) error {
	if worker == nil {
		return (*MailboxSubmissionWorker)(nil).Run(ctx)
	}
	return worker.worker.Run(ctx)
}

// RunOnce delegates one bounded scan and maps its content-free summary.
func (worker *MailboxSubmissionReconciliationWorker) RunOnce(ctx context.Context) (MailboxSubmissionReconciliationWorkerRunResult, error) {
	if worker == nil {
		result, err := (*MailboxSubmissionWorker)(nil).RunOnce(ctx)
		return mailboxSubmissionReconciliationWorkerRunResult(result), err
	}
	result, err := worker.worker.RunOnce(ctx)
	return mailboxSubmissionReconciliationWorkerRunResult(result), err
}

type mailboxSubmissionReconciliationWorkerScanner struct {
	repository MailboxSubmissionReconciliationRepository
}

func mailboxSubmissionReconciliationWorkerScannerFor(repository MailboxSubmissionReconciliationRepository) MailboxSubmissionWorkerDueScanner {
	if mailboxSubmissionWorkerNilDependency(repository) {
		return nil
	}
	return mailboxSubmissionReconciliationWorkerScanner{repository: repository}
}

func (scanner mailboxSubmissionReconciliationWorkerScanner) ScanDueMailboxSubmissionIDs(ctx context.Context, request MailboxSubmissionWorkerDueScanRequest) (MailboxSubmissionWorkerDueScanResult, error) {
	result, err := scanner.repository.ScanUnknownMailboxSubmissions(ctx, MailboxSubmissionReconciliationScanRequest{Limit: request.Limit})
	return MailboxSubmissionWorkerDueScanResult{SubmissionIDs: result.SubmissionIDs}, err
}

type mailboxSubmissionReconciliationWorkerDispatcher struct {
	executor MailboxSubmissionReconciliationExecutor
}

func mailboxSubmissionReconciliationWorkerDispatcherFor(executor MailboxSubmissionReconciliationExecutor) MailboxSubmissionWorkerDispatcher {
	if mailboxSubmissionWorkerNilDependency(executor) {
		return nil
	}
	return mailboxSubmissionReconciliationWorkerDispatcher{executor: executor}
}

func (dispatcher mailboxSubmissionReconciliationWorkerDispatcher) Dispatch(ctx context.Context, submissionID uuid.UUID) error {
	return dispatcher.executor.Reconcile(ctx, submissionID)
}

func mailboxSubmissionReconciliationWorkerOptions(options MailboxSubmissionReconciliationWorkerOptions) MailboxSubmissionWorkerOptions {
	return MailboxSubmissionWorkerOptions{
		Enabled:         options.Enabled,
		ScanInterval:    options.ScanInterval,
		DispatchTimeout: options.ReconcileTimeout,
		BatchSize:       options.BatchSize,
	}
}

func mailboxSubmissionReconciliationWorkerRunResult(result MailboxSubmissionWorkerRunResult) MailboxSubmissionReconciliationWorkerRunResult {
	return MailboxSubmissionReconciliationWorkerRunResult{
		Due:        result.Due,
		Reconciled: result.Dispatched,
		Failed:     result.Failed,
	}
}
