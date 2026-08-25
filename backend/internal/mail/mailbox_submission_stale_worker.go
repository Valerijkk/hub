package mail

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// MailboxSubmissionStaleWorkerOptions bounds one sequential stale-recovery
// loop. Zero enabled values use the submission worker defaults.
type MailboxSubmissionStaleWorkerOptions struct {
	Enabled         bool
	ScanInterval    time.Duration
	RecoveryTimeout time.Duration
	BatchSize       int
}

// DefaultMailboxSubmissionStaleWorkerOptions returns conservative enabled
// options matching the submission worker defaults.
func DefaultMailboxSubmissionStaleWorkerOptions() MailboxSubmissionStaleWorkerOptions {
	defaults := DefaultMailboxSubmissionWorkerOptions()
	return MailboxSubmissionStaleWorkerOptions{
		Enabled:         defaults.Enabled,
		ScanInterval:    defaults.ScanInterval,
		RecoveryTimeout: defaults.DispatchTimeout,
		BatchSize:       defaults.BatchSize,
	}
}

// MailboxSubmissionStaleWorkerRunResult summarizes one bounded recovery scan
// without exposing durable identities or recovery details.
type MailboxSubmissionStaleWorkerRunResult struct {
	Due       int
	Recovered int
	Failed    int
}

// MailboxSubmissionStaleWorker adapts advisory stale-lease scans and durable
// recovery to the reusable sequential submission worker.
type MailboxSubmissionStaleWorker struct {
	worker *MailboxSubmissionWorker
}

// NewMailboxSubmissionStaleWorker constructs a disabled zero worker or
// delegates enabled validation and execution boundaries to the submission
// worker. The repository remains the sole durable recovery authority.
func NewMailboxSubmissionStaleWorker(repository MailboxSubmissionStaleRecoveryRepository, timer MailboxSubmissionWorkerTimer, options MailboxSubmissionStaleWorkerOptions) (*MailboxSubmissionStaleWorker, error) {
	worker, err := NewMailboxSubmissionWorker(
		mailboxSubmissionStaleWorkerScannerFor(repository),
		mailboxSubmissionStaleWorkerDispatcherFor(repository),
		timer,
		mailboxSubmissionStaleWorkerOptions(options),
	)
	if err != nil {
		return nil, err
	}
	return &MailboxSubmissionStaleWorker{worker: worker}, nil
}

// Run delegates the guarded sequential loop to the submission worker.
func (worker *MailboxSubmissionStaleWorker) Run(ctx context.Context) error {
	if worker == nil {
		return (*MailboxSubmissionWorker)(nil).Run(ctx)
	}
	return worker.worker.Run(ctx)
}

// RunOnce delegates one bounded scan and maps its content-free summary.
func (worker *MailboxSubmissionStaleWorker) RunOnce(ctx context.Context) (MailboxSubmissionStaleWorkerRunResult, error) {
	if worker == nil {
		result, err := (*MailboxSubmissionWorker)(nil).RunOnce(ctx)
		return mailboxSubmissionStaleWorkerRunResult(result), err
	}
	result, err := worker.worker.RunOnce(ctx)
	return mailboxSubmissionStaleWorkerRunResult(result), err
}

type mailboxSubmissionStaleWorkerScanner struct {
	repository MailboxSubmissionStaleRecoveryRepository
}

func mailboxSubmissionStaleWorkerScannerFor(repository MailboxSubmissionStaleRecoveryRepository) MailboxSubmissionWorkerDueScanner {
	if mailboxSubmissionWorkerNilDependency(repository) {
		return nil
	}
	return mailboxSubmissionStaleWorkerScanner{repository: repository}
}

func (scanner mailboxSubmissionStaleWorkerScanner) ScanDueMailboxSubmissionIDs(ctx context.Context, request MailboxSubmissionWorkerDueScanRequest) (MailboxSubmissionWorkerDueScanResult, error) {
	result, err := scanner.repository.ScanDueMailboxSubmissionRecovery(ctx, MailboxSubmissionRecoveryScanRequest{Limit: request.Limit})
	ids := make([]uuid.UUID, len(result.Submissions))
	for index, due := range result.Submissions {
		ids[index] = due.SubmissionID
	}
	return MailboxSubmissionWorkerDueScanResult{SubmissionIDs: ids}, err
}

type mailboxSubmissionStaleWorkerDispatcher struct {
	repository MailboxSubmissionStaleRecoveryRepository
}

func mailboxSubmissionStaleWorkerDispatcherFor(repository MailboxSubmissionStaleRecoveryRepository) MailboxSubmissionWorkerDispatcher {
	if mailboxSubmissionWorkerNilDependency(repository) {
		return nil
	}
	return mailboxSubmissionStaleWorkerDispatcher{repository: repository}
}

func (dispatcher mailboxSubmissionStaleWorkerDispatcher) Dispatch(ctx context.Context, submissionID uuid.UUID) error {
	result, err := dispatcher.repository.RecoverStaleMailboxSubmission(ctx, MailboxSubmissionStaleRecoveryRequest{SubmissionID: submissionID})
	if err != nil {
		return err
	}
	if !mailboxSubmissionStaleRecoveryResultValid(result) {
		return mailboxSubmissionWorkerError(MailboxSubmissionWorkerErrorUnavailable)
	}
	return nil
}

func mailboxSubmissionStaleWorkerOptions(options MailboxSubmissionStaleWorkerOptions) MailboxSubmissionWorkerOptions {
	return MailboxSubmissionWorkerOptions{
		Enabled:         options.Enabled,
		ScanInterval:    options.ScanInterval,
		DispatchTimeout: options.RecoveryTimeout,
		BatchSize:       options.BatchSize,
	}
}

func mailboxSubmissionStaleWorkerRunResult(result MailboxSubmissionWorkerRunResult) MailboxSubmissionStaleWorkerRunResult {
	return MailboxSubmissionStaleWorkerRunResult{
		Due:       result.Due,
		Recovered: result.Dispatched,
		Failed:    result.Failed,
	}
}
