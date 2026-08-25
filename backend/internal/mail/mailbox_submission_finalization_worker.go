package mail

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MailboxSubmissionFinalizationExecutor performs one finalization for a
// durable submission identity.
type MailboxSubmissionFinalizationExecutor interface {
	Finalize(context.Context, uuid.UUID) error
}

var _ MailboxSubmissionFinalizationExecutor = (*MailboxSubmissionFinalizer)(nil)

// MailboxSubmissionFinalizationWorkerOptions bounds one sequential scan and
// finalization loop. Zero enabled values use the submission worker defaults.
type MailboxSubmissionFinalizationWorkerOptions struct {
	Enabled         bool
	ScanInterval    time.Duration
	FinalizeTimeout time.Duration
	BatchSize       int
}

// DefaultMailboxSubmissionFinalizationWorkerOptions returns conservative
// enabled options matching the submission worker defaults.
func DefaultMailboxSubmissionFinalizationWorkerOptions() MailboxSubmissionFinalizationWorkerOptions {
	defaults := DefaultMailboxSubmissionWorkerOptions()
	return MailboxSubmissionFinalizationWorkerOptions{
		Enabled:         defaults.Enabled,
		ScanInterval:    defaults.ScanInterval,
		FinalizeTimeout: defaults.DispatchTimeout,
		BatchSize:       defaults.BatchSize,
	}
}

// MailboxSubmissionFinalizationWorkerRunResult summarizes one bounded scan
// without exposing durable identities.
type MailboxSubmissionFinalizationWorkerRunResult struct {
	Due       int
	Finalized int
	Failed    int
}

// MailboxSubmissionFinalizationWorker adapts accepted-submission scans to the
// reusable sequential submission worker.
type MailboxSubmissionFinalizationWorker struct {
	worker *MailboxSubmissionWorker
}

// NewMailboxSubmissionFinalizationWorker constructs a disabled zero worker or
// delegates enabled validation and execution boundaries to the submission
// worker.
func NewMailboxSubmissionFinalizationWorker(repository MailboxSubmissionFinalizationRepository, executor MailboxSubmissionFinalizationExecutor, timer MailboxSubmissionWorkerTimer, options MailboxSubmissionFinalizationWorkerOptions) (*MailboxSubmissionFinalizationWorker, error) {
	worker, err := NewMailboxSubmissionWorker(
		mailboxSubmissionFinalizationWorkerScannerFor(repository),
		mailboxSubmissionFinalizationWorkerDispatcherFor(executor),
		timer,
		mailboxSubmissionFinalizationWorkerOptions(options),
	)
	if err != nil {
		return nil, err
	}
	return &MailboxSubmissionFinalizationWorker{worker: worker}, nil
}

// Run delegates the guarded sequential loop to the submission worker.
func (worker *MailboxSubmissionFinalizationWorker) Run(ctx context.Context) error {
	if worker == nil {
		return (*MailboxSubmissionWorker)(nil).Run(ctx)
	}
	return worker.worker.Run(ctx)
}

// RunOnce delegates one bounded scan and maps its content-free summary.
func (worker *MailboxSubmissionFinalizationWorker) RunOnce(ctx context.Context) (MailboxSubmissionFinalizationWorkerRunResult, error) {
	if worker == nil {
		result, err := (*MailboxSubmissionWorker)(nil).RunOnce(ctx)
		return mailboxSubmissionFinalizationWorkerRunResult(result), err
	}
	result, err := worker.worker.RunOnce(ctx)
	return mailboxSubmissionFinalizationWorkerRunResult(result), err
}

type mailboxSubmissionFinalizationWorkerScanner struct {
	repository MailboxSubmissionFinalizationRepository
}

func mailboxSubmissionFinalizationWorkerScannerFor(repository MailboxSubmissionFinalizationRepository) MailboxSubmissionWorkerDueScanner {
	if mailboxSubmissionWorkerNilDependency(repository) {
		return nil
	}
	return mailboxSubmissionFinalizationWorkerScanner{repository: repository}
}

func (scanner mailboxSubmissionFinalizationWorkerScanner) ScanDueMailboxSubmissionIDs(ctx context.Context, request MailboxSubmissionWorkerDueScanRequest) (MailboxSubmissionWorkerDueScanResult, error) {
	result, err := scanner.repository.ScanAcceptedMailboxSubmissions(ctx, MailboxSubmissionFinalizationScanRequest{Limit: request.Limit})
	return MailboxSubmissionWorkerDueScanResult{SubmissionIDs: result.SubmissionIDs}, err
}

type mailboxSubmissionFinalizationWorkerDispatcher struct {
	executor MailboxSubmissionFinalizationExecutor
}

func mailboxSubmissionFinalizationWorkerDispatcherFor(executor MailboxSubmissionFinalizationExecutor) MailboxSubmissionWorkerDispatcher {
	if mailboxSubmissionWorkerNilDependency(executor) {
		return nil
	}
	return mailboxSubmissionFinalizationWorkerDispatcher{executor: executor}
}

func (dispatcher mailboxSubmissionFinalizationWorkerDispatcher) Dispatch(ctx context.Context, submissionID uuid.UUID) error {
	return dispatcher.executor.Finalize(ctx, submissionID)
}

func mailboxSubmissionFinalizationWorkerOptions(options MailboxSubmissionFinalizationWorkerOptions) MailboxSubmissionWorkerOptions {
	return MailboxSubmissionWorkerOptions{
		Enabled:         options.Enabled,
		ScanInterval:    options.ScanInterval,
		DispatchTimeout: options.FinalizeTimeout,
		BatchSize:       options.BatchSize,
	}
}

func mailboxSubmissionFinalizationWorkerRunResult(result MailboxSubmissionWorkerRunResult) MailboxSubmissionFinalizationWorkerRunResult {
	return MailboxSubmissionFinalizationWorkerRunResult{
		Due:       result.Due,
		Finalized: result.Dispatched,
		Failed:    result.Failed,
	}
}
