package mail

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const mailboxSubmissionReconciliationMaximum = 50

// MailboxSubmissionReconciliationScanRequest bounds one advisory selection of
// unknown submissions. The result contains durable identities only.
type MailboxSubmissionReconciliationScanRequest struct {
	Limit int
}

// MailboxSubmissionReconciliationScanResult is a bounded, copied selection
// of submission identities. It grants no claim or engine capability.
type MailboxSubmissionReconciliationScanResult struct {
	SubmissionIDs []uuid.UUID
}

// MailboxSubmissionReconciliationClaimRequest asks the repository to take one
// short-lived reconciliation claim for a durable unknown submission.
type MailboxSubmissionReconciliationClaimRequest struct {
	SubmissionID uuid.UUID
}

// MailboxSubmissionReconciliationClaim is a sealed capability minted only by
// the repository. The immutable engine tuple stays private to this package.
type MailboxSubmissionReconciliationClaim struct {
	submissionID uuid.UUID
	fence        MailboxEngineCommandFence
	submission   EngineSubmission
}

// SubmissionID returns the durable identity held by this sealed claim.
func (claim MailboxSubmissionReconciliationClaim) SubmissionID() uuid.UUID {
	if !mailboxSubmissionReconciliationClaimValid(claim) {
		return uuid.Nil
	}
	return claim.submissionID
}

// Fence returns a defensive exact copy of the shared command fence.
func (claim MailboxSubmissionReconciliationClaim) Fence() MailboxEngineCommandFence {
	if !mailboxSubmissionReconciliationClaimValid(claim) {
		return MailboxEngineCommandFence{}
	}
	return mailboxSubmissionClaimFenceCopy(claim.fence)
}

// MailboxSubmissionReconciliationClaimOutcome is the closed result of the
// repository's short claim transaction.
type MailboxSubmissionReconciliationClaimOutcome string

const (
	MailboxSubmissionReconciliationClaimOutcomeNoop    MailboxSubmissionReconciliationClaimOutcome = "noop"
	MailboxSubmissionReconciliationClaimOutcomeBusy    MailboxSubmissionReconciliationClaimOutcome = "busy"
	MailboxSubmissionReconciliationClaimOutcomeClaimed MailboxSubmissionReconciliationClaimOutcome = "claimed"
)

// MailboxSubmissionReconciliationClaimResult carries exactly one of noop,
// busy, or a sealed reconciliation claim.
type MailboxSubmissionReconciliationClaimResult struct {
	outcome   MailboxSubmissionReconciliationClaimOutcome
	claim     MailboxSubmissionReconciliationClaim
	busyUntil *time.Time
}
// NewMailboxSubmissionReconciliationNoopResult returns the canonical no-op
// outcome without a sealed capability or lease observation.
func NewMailboxSubmissionReconciliationNoopResult() MailboxSubmissionReconciliationClaimResult {
	return MailboxSubmissionReconciliationClaimResult{outcome: MailboxSubmissionReconciliationClaimOutcomeNoop}
}
// NewMailboxSubmissionReconciliationBusyResult validates and copies a
// canonical UTC busy observation before it crosses the repository boundary.
func NewMailboxSubmissionReconciliationBusyResult(busyUntil time.Time) (MailboxSubmissionReconciliationClaimResult, error) {
	if !mailboxSubmissionClaimExactTimeValid(busyUntil) {
		return MailboxSubmissionReconciliationClaimResult{}, newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
	}
	copyUntil := busyUntil.Round(0).UTC()
	return MailboxSubmissionReconciliationClaimResult{outcome: MailboxSubmissionReconciliationClaimOutcomeBusy, busyUntil: &copyUntil}, nil
}
// NewMailboxSubmissionReconciliationClaimedResult validates and seals the
// exact durable claim tuple without exposing the immutable engine tuple.
func NewMailboxSubmissionReconciliationClaimedResult(submissionID uuid.UUID, fence MailboxEngineCommandFence, submission EngineSubmission) (MailboxSubmissionReconciliationClaimResult, error) {
	claim := MailboxSubmissionReconciliationClaim{submissionID: submissionID, fence: mailboxSubmissionClaimFenceCopy(fence), submission: submission}
	if !mailboxSubmissionReconciliationClaimValid(claim) {
		return MailboxSubmissionReconciliationClaimResult{}, newMailboxSubmissionReconciliationError(MailboxSubmissionReconciliationErrorInvalid)
	}
	return MailboxSubmissionReconciliationClaimResult{outcome: MailboxSubmissionReconciliationClaimOutcomeClaimed, claim: mailboxSubmissionReconciliationClaimCopy(claim)}, nil
}

// Outcome returns the closed claim outcome, or the zero value for malformed
// repository output.
func (result MailboxSubmissionReconciliationClaimResult) Outcome() MailboxSubmissionReconciliationClaimOutcome {
	if !mailboxSubmissionReconciliationClaimResultValid(result) {
		return ""
	}
	return result.outcome
}

// Claim returns a defensive sealed capability only for a valid claimed result.
func (result MailboxSubmissionReconciliationClaimResult) Claim() (MailboxSubmissionReconciliationClaim, bool) {
	if !mailboxSubmissionReconciliationClaimResultValid(result) || result.outcome != MailboxSubmissionReconciliationClaimOutcomeClaimed {
		return MailboxSubmissionReconciliationClaim{}, false
	}
	return mailboxSubmissionReconciliationClaimCopy(result.claim), true
}

// BusyUntil returns a defensive UTC lease observation only for a valid busy
// result. It is advisory and cannot be used as a future claim capability.
func (result MailboxSubmissionReconciliationClaimResult) BusyUntil() (time.Time, bool) {
	if !mailboxSubmissionReconciliationClaimResultValid(result) || result.outcome != MailboxSubmissionReconciliationClaimOutcomeBusy {
		return time.Time{}, false
	}
	return result.busyUntil.Round(0).UTC(), true
}

// MailboxSubmissionReconciliationObservation is the only durable conclusion
// after one engine reconciliation read.
type MailboxSubmissionReconciliationObservation string

const (
	MailboxSubmissionReconciliationObservationAccepted MailboxSubmissionReconciliationObservation = "accepted"
	MailboxSubmissionReconciliationObservationRetry    MailboxSubmissionReconciliationObservation = "retry"
)

// MailboxSubmissionReconciliationResolveRequest gives the repository either
// an exact accepted engine pair or one safe retry code. The repository owns
// attempt consumption, backoff, and exhaustion.
type MailboxSubmissionReconciliationResolveRequest struct {
	Claim              MailboxSubmissionReconciliationClaim
	Observation        MailboxSubmissionReconciliationObservation
	EngineSubmissionID string
	EngineMessageID    string
	RetryCode          MailboxSubmissionErrorCode
}

// MailboxSubmissionReconciliationRepository owns advisory scans and short
// durable claim and resolution boundaries. It performs no engine I/O.
type MailboxSubmissionReconciliationRepository interface {
	ScanUnknownMailboxSubmissions(context.Context, MailboxSubmissionReconciliationScanRequest) (MailboxSubmissionReconciliationScanResult, error)
	ClaimUnknownMailboxSubmission(context.Context, MailboxSubmissionReconciliationClaimRequest) (MailboxSubmissionReconciliationClaimResult, error)
	ResolveUnknownMailboxSubmissionReconciliation(context.Context, MailboxSubmissionReconciliationResolveRequest) error
}

func mailboxSubmissionReconciliationScanRequestValid(request MailboxSubmissionReconciliationScanRequest) bool {
	return request.Limit >= 1 && request.Limit <= mailboxSubmissionReconciliationMaximum
}

func mailboxSubmissionReconciliationScanResultValid(request MailboxSubmissionReconciliationScanRequest, result MailboxSubmissionReconciliationScanResult) bool {
	if !mailboxSubmissionReconciliationScanRequestValid(request) || result.SubmissionIDs == nil || len(result.SubmissionIDs) > request.Limit {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(result.SubmissionIDs))
	for _, submissionID := range result.SubmissionIDs {
		if submissionID == uuid.Nil {
			return false
		}
		if _, duplicate := seen[submissionID]; duplicate {
			return false
		}
		seen[submissionID] = struct{}{}
	}
	return true
}

func mailboxSubmissionReconciliationScanResultCopy(result MailboxSubmissionReconciliationScanResult) MailboxSubmissionReconciliationScanResult {
	result.SubmissionIDs = append([]uuid.UUID(nil), result.SubmissionIDs...)
	return result
}

func mailboxSubmissionReconciliationClaimRequestValid(request MailboxSubmissionReconciliationClaimRequest) bool {
	return request.SubmissionID != uuid.Nil
}

func mailboxSubmissionReconciliationClaimValid(claim MailboxSubmissionReconciliationClaim) bool {
	return claim.submissionID != uuid.Nil && mailboxSubmissionClaimFenceValid(claim.fence) && validateSubmission(claim.submission) == nil
}

func mailboxSubmissionReconciliationClaimCopy(claim MailboxSubmissionReconciliationClaim) MailboxSubmissionReconciliationClaim {
	claim.fence = mailboxSubmissionClaimFenceCopy(claim.fence)
	return claim
}

func mailboxSubmissionReconciliationClaimZero(claim MailboxSubmissionReconciliationClaim) bool {
	return claim.submissionID == uuid.Nil && claim.fence == (MailboxEngineCommandFence{}) && claim.submission == (EngineSubmission{})
}

func mailboxSubmissionReconciliationClaimResultValid(result MailboxSubmissionReconciliationClaimResult) bool {
	switch result.outcome {
	case MailboxSubmissionReconciliationClaimOutcomeNoop:
		return mailboxSubmissionReconciliationClaimZero(result.claim) && result.busyUntil == nil
	case MailboxSubmissionReconciliationClaimOutcomeBusy:
		return mailboxSubmissionReconciliationClaimZero(result.claim) && mailboxSubmissionClaimBusyUntilValid(result.busyUntil)
	case MailboxSubmissionReconciliationClaimOutcomeClaimed:
		return mailboxSubmissionReconciliationClaimValid(result.claim) && result.busyUntil == nil
	default:
		return false
	}
}

func mailboxSubmissionReconciliationResolveRequestValid(request MailboxSubmissionReconciliationResolveRequest) bool {
	if !mailboxSubmissionReconciliationClaimValid(request.Claim) {
		return false
	}
	accepted := mailboxSubmissionReconciliationAccepted{
		EngineSubmissionID: request.EngineSubmissionID,
		EngineMessageID:    request.EngineMessageID,
	}
	switch request.Observation {
	case MailboxSubmissionReconciliationObservationAccepted:
		if accepted.EngineMessageID != request.Claim.submission.MessageID {
			return false
		}
		return validOpaqueID(accepted.EngineSubmissionID) && validOpaqueID(accepted.EngineMessageID) && request.RetryCode == ""
	case MailboxSubmissionReconciliationObservationRetry:
		return accepted.EngineSubmissionID == "" && accepted.EngineMessageID == "" && mailboxSubmissionReconciliationRetryCodeValid(request.RetryCode)
	default:
		return false
	}
}

type mailboxSubmissionReconciliationAccepted struct {
	EngineSubmissionID string
	EngineMessageID    string
}

func mailboxSubmissionReconciliationRetryCodeValid(code MailboxSubmissionErrorCode) bool {
	return code == MailboxSubmissionErrorCodeMailboxUnavailable || code == MailboxSubmissionErrorCodeReconciliationRequired
}

// MailboxSubmissionReconciliationErrorKind is the closed content-free error
// family exposed by the reconciliation flow.
type MailboxSubmissionReconciliationErrorKind string

const (
	MailboxSubmissionReconciliationErrorInvalid                MailboxSubmissionReconciliationErrorKind = "invalid"
	MailboxSubmissionReconciliationErrorRepositoryUnavailable  MailboxSubmissionReconciliationErrorKind = "repository_unavailable"
	MailboxSubmissionReconciliationErrorReconciliationRequired MailboxSubmissionReconciliationErrorKind = "reconciliation_required"
)

const mailboxSubmissionReconciliationErrorText = "mailbox submission reconciliation error"

// MailboxSubmissionReconciliationError carries no raw engine or durable
// failure detail.
type MailboxSubmissionReconciliationError struct {
	kind MailboxSubmissionReconciliationErrorKind
}

func (errorValue *MailboxSubmissionReconciliationError) Error() string {
	return mailboxSubmissionReconciliationErrorText
}

// Kind returns a safe reconciliation error classification.
func (errorValue *MailboxSubmissionReconciliationError) Kind() MailboxSubmissionReconciliationErrorKind {
	if errorValue == nil || !validMailboxSubmissionReconciliationErrorKind(errorValue.kind) {
		return MailboxSubmissionReconciliationErrorInvalid
	}
	return errorValue.kind
}

func newMailboxSubmissionReconciliationError(kind MailboxSubmissionReconciliationErrorKind) *MailboxSubmissionReconciliationError {
	if !validMailboxSubmissionReconciliationErrorKind(kind) {
		kind = MailboxSubmissionReconciliationErrorInvalid
	}
	return &MailboxSubmissionReconciliationError{kind: kind}
}

func validMailboxSubmissionReconciliationErrorKind(kind MailboxSubmissionReconciliationErrorKind) bool {
	switch kind {
	case MailboxSubmissionReconciliationErrorInvalid,
		MailboxSubmissionReconciliationErrorRepositoryUnavailable,
		MailboxSubmissionReconciliationErrorReconciliationRequired:
		return true
	default:
		return false
	}
}
