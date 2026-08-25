package mail

import (
	"context"
	"errors"
)

// MailboxSubmissionDispatchResolution is the only durable conclusion a
// dispatcher may report after provider submission. Unknown is deliberately
// distinct from a retryable preflight outcome: a sealed submission is never
// dispatched again by this path.
type MailboxSubmissionDispatchResolution string

const (
	MailboxSubmissionDispatchResolutionAccepted MailboxSubmissionDispatchResolution = "accepted"
	MailboxSubmissionDispatchResolutionUnknown  MailboxSubmissionDispatchResolution = "unknown"
)

// MailboxSubmissionDispatchResolutionRequest carries one sealed authorization
// and its provider outcome to the durable repository. Provider identifiers are
// write-only dispatcher input: they are not returned by the dispatcher.
type MailboxSubmissionDispatchResolutionRequest struct {
	Authorization        MailboxSubmissionDispatchAuthorization
	Resolution           MailboxSubmissionDispatchResolution
	ProviderSubmissionID string
	ProviderMessageID    string
}

// MailboxSubmissionDispatchRepository owns the claim, authorization and final
// resolution boundaries. Every method is a durable operation; provider I/O is
// owned exclusively by MailboxSubmissionDispatcher outside transactions.
type MailboxSubmissionDispatchRepository interface {
	MailboxSubmissionClaimRepository
	ResolveMailboxSubmissionDispatch(context.Context, MailboxSubmissionDispatchResolutionRequest) error
}

// MailboxSubmissionDispatchErrorKind is the closed, content-free failure
// family exposed by the dispatcher.
type MailboxSubmissionDispatchErrorKind string

const (
	MailboxSubmissionDispatchErrorInvalid               MailboxSubmissionDispatchErrorKind = "invalid"
	MailboxSubmissionDispatchErrorRepositoryUnavailable MailboxSubmissionDispatchErrorKind = "repository_unavailable"
	MailboxSubmissionDispatchErrorReconciliationRequired MailboxSubmissionDispatchErrorKind = "reconciliation_required"
)

const mailboxSubmissionDispatchErrorText = "mailbox submission dispatch error"

// MailboxSubmissionDispatchError never wraps provider, recipient, or message
// data. Callers can branch only on its bounded kind.
type MailboxSubmissionDispatchError struct {
	kind MailboxSubmissionDispatchErrorKind
}

func (errorValue *MailboxSubmissionDispatchError) Error() string {
	return mailboxSubmissionDispatchErrorText
}

// Kind returns a safe error classification.
func (errorValue *MailboxSubmissionDispatchError) Kind() MailboxSubmissionDispatchErrorKind {
	if errorValue == nil || !validMailboxSubmissionDispatchErrorKind(errorValue.kind) {
		return MailboxSubmissionDispatchErrorInvalid
	}
	return errorValue.kind
}

func newMailboxSubmissionDispatchError(kind MailboxSubmissionDispatchErrorKind) *MailboxSubmissionDispatchError {
	if !validMailboxSubmissionDispatchErrorKind(kind) {
		kind = MailboxSubmissionDispatchErrorInvalid
	}
	return &MailboxSubmissionDispatchError{kind: kind}
}

func validMailboxSubmissionDispatchErrorKind(kind MailboxSubmissionDispatchErrorKind) bool {
	switch kind {
	case MailboxSubmissionDispatchErrorInvalid,
		MailboxSubmissionDispatchErrorRepositoryUnavailable,
		MailboxSubmissionDispatchErrorReconciliationRequired:
		return true
	default:
		return false
	}
}

func validMailboxSubmissionDispatchResolution(resolution MailboxSubmissionDispatchResolution) bool {
	return resolution == MailboxSubmissionDispatchResolutionAccepted || resolution == MailboxSubmissionDispatchResolutionUnknown
}

func mailboxSubmissionDispatchPreflightFailure(claim MailboxSubmissionClaim, cause error) MailboxSubmissionPreflightFailureRequest {
	request := MailboxSubmissionPreflightFailureRequest{Claim: claim}
	switch {
	case errors.Is(cause, ErrEngineNotFound), errors.Is(cause, ErrEngineConflict):
		request.Disposition = MailboxSubmissionPreflightFailureDispositionFailed
		request.FailureCode = MailboxSubmissionErrorCodeRejected
	case errors.Is(cause, ErrEngineRejectedPolicy):
		request.Disposition = MailboxSubmissionPreflightFailureDispositionFailed
		request.FailureCode = MailboxSubmissionErrorCodePolicy
	case errors.Is(cause, ErrEnginePayloadTooLarge):
		request.Disposition = MailboxSubmissionPreflightFailureDispositionFailed
		request.FailureCode = MailboxSubmissionErrorCodeSizeLimit
	case errors.Is(cause, ErrEngineRecipientLimit):
		request.Disposition = MailboxSubmissionPreflightFailureDispositionFailed
		request.FailureCode = MailboxSubmissionErrorCodeRecipientLimit
	case errors.Is(cause, ErrEngineIdentityUnavailable):
		request.Disposition = MailboxSubmissionPreflightFailureDispositionFailed
		request.FailureCode = MailboxSubmissionErrorCodeIdentityUnavailable
	case errors.Is(cause, ErrEngineRateLimited):
		request.Disposition = MailboxSubmissionPreflightFailureDispositionRateLimited
		request.RetryAfterSeconds = 60
	default:
		request.Disposition = MailboxSubmissionPreflightFailureDispositionRetry
		request.RetryAfterSeconds = 60
	}
	return request
}
