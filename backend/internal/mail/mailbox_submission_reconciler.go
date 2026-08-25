package mail

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

// MailboxSubmissionReconciler performs one claimed, read-only submission
// reconciliation and durably records only its closed observation.
type MailboxSubmissionReconciler struct {
	repository MailboxSubmissionReconciliationRepository
	engine     MailEngine
}

// NewMailboxSubmissionReconciler requires non-nil durable and engine ports.
func NewMailboxSubmissionReconciler(repository MailboxSubmissionReconciliationRepository, engine MailEngine) (*MailboxSubmissionReconciler, error) {
	if mailboxSubmissionTypedNil(repository) || mailboxSubmissionTypedNil(engine) {
		return nil, newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
	}
	return &MailboxSubmissionReconciler{repository: repository, engine: engine}, nil
}

// Reconcile claims one unknown submission and makes at most one engine read.
// Cancellation before the claim is a no-op; cancellation after a claim is
// durably recorded as a safe retry without entering the engine call.
func (reconciler *MailboxSubmissionReconciler) Reconcile(ctx context.Context, submissionID uuid.UUID) error {
	if reconciler == nil || mailboxSubmissionTypedNil(reconciler.repository) || mailboxSubmissionTypedNil(reconciler.engine) || ctx == nil || submissionID == uuid.Nil {
		return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
	}
	if ctx.Err() != nil {
		return nil
	}

	claimResult, err := reconciler.repository.ClaimUnknownMailboxSubmission(ctx, MailboxSubmissionReconciliationClaimRequest{SubmissionID: submissionID})
	if err != nil {
		return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorRepositoryUnavailable)
	}
	switch claimResult.Outcome() {
	case MailboxSubmissionReconciliationClaimOutcomeNoop, MailboxSubmissionReconciliationClaimOutcomeBusy:
		return nil
	case MailboxSubmissionReconciliationClaimOutcomeClaimed:
		// The sealed claim is extracted below.
	default:
		return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
	}
	claim, claimed := claimResult.Claim()
	if !claimed {
		return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
	}
	if ctx.Err() != nil {
		return reconciler.resolveRetry(ctx, claim, MailboxSubmissionErrorCodeReconciliationRequired)
	}

	result, err := reconciler.engine.ReconcileSubmission(ctx, claim.submission)
	if err != nil {
		return reconciler.resolveRetry(ctx, claim, mailboxSubmissionReconciliationRetryCode(err, nil))
	}
	accepted, ok := result.(EngineSubmissionAccepted)
	if !ok {
		return reconciler.resolveRetry(ctx, claim, mailboxSubmissionReconciliationRetryCode(nil, result))
	}
	request := MailboxSubmissionReconciliationResolveRequest{
		Claim:              claim,
		Observation:        MailboxSubmissionReconciliationObservationAccepted,
		EngineSubmissionID: accepted.SubmissionID,
		EngineMessageID:    accepted.MessageID,
	}
	if !mailboxSubmissionReconciliationResolveRequestValid(request) {
		return reconciler.resolveRetry(ctx, claim, MailboxSubmissionErrorCodeReconciliationRequired)
	}
	return reconciler.resolve(ctx, request)
}

func (reconciler *MailboxSubmissionReconciler) resolveRetry(ctx context.Context, claim MailboxSubmissionReconciliationClaim, code MailboxSubmissionErrorCode) error {
	return reconciler.resolve(ctx, MailboxSubmissionReconciliationResolveRequest{
		Claim:       claim,
		Observation: MailboxSubmissionReconciliationObservationRetry,
		RetryCode:   code,
	})
}

func (reconciler *MailboxSubmissionReconciler) resolve(ctx context.Context, request MailboxSubmissionReconciliationResolveRequest) error {
	if !mailboxSubmissionReconciliationResolveRequestValid(request) {
		return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
	}
	if err := reconciler.repository.ResolveUnknownMailboxSubmissionReconciliation(context.WithoutCancel(ctx), request); err != nil {
		return newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorReconciliationRequired)
	}
	return nil
}

func mailboxSubmissionReconciliationRetryCode(cause error, result EngineSubmissionResult) MailboxSubmissionErrorCode {
	if errors.Is(cause, ErrEngineUnavailable) || errors.Is(cause, ErrEngineRateLimited) {
		return MailboxSubmissionErrorCodeMailboxUnavailable
	}
	if indeterminate, ok := result.(EngineSubmissionIndeterminate); ok && indeterminate.Code == EngineIndeterminateUnavailable {
		return MailboxSubmissionErrorCodeMailboxUnavailable
	}
	return MailboxSubmissionErrorCodeReconciliationRequired
}
