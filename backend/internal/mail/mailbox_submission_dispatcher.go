package mail

import (
	"context"
	"reflect"

	"github.com/google/uuid"
)

type mailboxSubmissionEngine interface {
	MailEngine
	submitMailboxSubmission(context.Context, EngineSubmission) (EngineSubmissionResult, error)
}

// MailboxSubmissionDispatcher is the sole production path that can invoke a
// provider submission. It performs provider I/O only after the repository has
// durably sealed MailboxSubmissionDispatchAuthorization.
type MailboxSubmissionDispatcher struct {
	repository MailboxSubmissionDispatchRepository
	engine     mailboxSubmissionEngine
}

// NewMailboxSubmissionDispatcher requires a MailEngine implementation that
// also exposes the package-private mutation method. External consumers can use
// the read/preflight port but cannot construct a provider mutation call.
func NewMailboxSubmissionDispatcher(repository MailboxSubmissionDispatchRepository, engine MailEngine) (*MailboxSubmissionDispatcher, error) {
	if mailboxSubmissionTypedNil(repository) || mailboxSubmissionTypedNil(engine) {
		return nil, newMailboxSubmissionDispatchError(MailboxSubmissionDispatchErrorInvalid)
	}
	submissionEngine, ok := engine.(mailboxSubmissionEngine)
	if !ok || mailboxSubmissionTypedNil(submissionEngine) {
		return nil, newMailboxSubmissionDispatchError(MailboxSubmissionDispatchErrorInvalid)
	}
	return &MailboxSubmissionDispatcher{repository: repository, engine: submissionEngine}, nil
}

func mailboxSubmissionTypedNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

// Dispatch claims one queued receipt, preflights it outside a transaction,
// authorizes it durably, and makes at most one provider mutation. A busy or
// stale submitting receipt is a no-op: it never enters the mutation path.
func (dispatcher *MailboxSubmissionDispatcher) Dispatch(ctx context.Context, submissionID uuid.UUID) error {
	if dispatcher == nil || dispatcher.repository == nil || dispatcher.engine == nil || ctx == nil || submissionID == uuid.Nil {
		return newMailboxSubmissionDispatchError(MailboxSubmissionDispatchErrorInvalid)
	}

	claimResult, err := dispatcher.repository.ClaimMailboxSubmission(ctx, MailboxSubmissionClaimRequest{SubmissionID: submissionID})
	if err != nil {
		return newMailboxSubmissionDispatchError(MailboxSubmissionDispatchErrorRepositoryUnavailable)
	}
	claim, claimed := claimResult.Claim()
	if !claimed {
		if _, busy := claimResult.BusyUntil(); busy {
			return nil
		}
		return newMailboxSubmissionDispatchError(MailboxSubmissionDispatchErrorInvalid)
	}

	preflightRequest, ok := claim.Policy().PreflightRequest(claim.submission)
	if !ok {
		return dispatcher.resolvePreflightFailure(ctx, mailboxSubmissionDispatchPreflightFailure(claim, nil))
	}
	preflight, err := dispatcher.engine.PreflightSubmission(ctx, preflightRequest)
	if err != nil || !mailboxSubmissionPreflightValid(preflight) {
		return dispatcher.resolvePreflightFailure(ctx, mailboxSubmissionDispatchPreflightFailure(claim, err))
	}

	authorizeResult, err := dispatcher.repository.AuthorizeMailboxSubmission(ctx, MailboxSubmissionAuthorizeRequest{Claim: claim, Preflight: preflight})
	if err != nil {
		return newMailboxSubmissionDispatchError(MailboxSubmissionDispatchErrorRepositoryUnavailable)
	}
	switch authorizeResult.Outcome() {
	case MailboxSubmissionAuthorizeOutcomeRateLimited, MailboxSubmissionAuthorizeOutcomeRequeued, MailboxSubmissionAuthorizeOutcomeFailed:
		return nil
	case MailboxSubmissionAuthorizeOutcomeAuthorized:
		authorization, authorized := authorizeResult.Authorization()
		if !authorized {
			return newMailboxSubmissionDispatchError(MailboxSubmissionDispatchErrorInvalid)
		}
		return dispatcher.dispatchAuthorized(ctx, claim, authorization)
	default:
		return newMailboxSubmissionDispatchError(MailboxSubmissionDispatchErrorInvalid)
	}
}

func (dispatcher *MailboxSubmissionDispatcher) resolvePreflightFailure(ctx context.Context, request MailboxSubmissionPreflightFailureRequest) error {
	result, err := dispatcher.repository.ResolveMailboxSubmissionPreflightFailure(context.WithoutCancel(ctx), request)
	if err != nil {
		return newMailboxSubmissionDispatchError(MailboxSubmissionDispatchErrorRepositoryUnavailable)
	}
	switch result.Outcome() {
	case MailboxSubmissionAuthorizeOutcomeRateLimited, MailboxSubmissionAuthorizeOutcomeRequeued, MailboxSubmissionAuthorizeOutcomeFailed:
		return nil
	default:
		return newMailboxSubmissionDispatchError(MailboxSubmissionDispatchErrorInvalid)
	}
}

func (dispatcher *MailboxSubmissionDispatcher) dispatchAuthorized(ctx context.Context, claim MailboxSubmissionClaim, authorization MailboxSubmissionDispatchAuthorization) error {
	if authorization.SubmissionID() != claim.SubmissionID() || authorization.Fence() != claim.Fence() || authorization.AuthorizedAt().IsZero() || ctx.Err() != nil {
		return dispatcher.resolve(ctx, authorization, MailboxSubmissionDispatchResolutionUnknown, "", "")
	}

	result, err := dispatcher.engine.submitMailboxSubmission(ctx, claim.submission)
	if err != nil {
		return dispatcher.resolve(ctx, authorization, MailboxSubmissionDispatchResolutionUnknown, "", "")
	}
	accepted, ok := result.(EngineSubmissionAccepted)
	if !ok || !validOpaqueID(accepted.SubmissionID) || !validOpaqueID(accepted.MessageID) {
		return dispatcher.resolve(ctx, authorization, MailboxSubmissionDispatchResolutionUnknown, "", "")
	}
	return dispatcher.resolve(ctx, authorization, MailboxSubmissionDispatchResolutionAccepted, accepted.SubmissionID, accepted.MessageID)
}

func (dispatcher *MailboxSubmissionDispatcher) resolve(ctx context.Context, authorization MailboxSubmissionDispatchAuthorization, resolution MailboxSubmissionDispatchResolution, providerSubmissionID, providerMessageID string) error {
	if !validMailboxSubmissionDispatchResolution(resolution) {
		return newMailboxSubmissionDispatchError(MailboxSubmissionDispatchErrorInvalid)
	}
	if err := dispatcher.repository.ResolveMailboxSubmissionDispatch(context.WithoutCancel(ctx), MailboxSubmissionDispatchResolutionRequest{
		Authorization:        authorization,
		Resolution:           resolution,
		ProviderSubmissionID: providerSubmissionID,
		ProviderMessageID:    providerMessageID,
	}); err != nil {
		return newMailboxSubmissionDispatchError(MailboxSubmissionDispatchErrorReconciliationRequired)
	}
	return nil
}
