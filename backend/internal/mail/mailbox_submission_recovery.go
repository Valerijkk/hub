package mail

import (
	"bytes"
	"context"
	"time"

	"github.com/google/uuid"
)

const mailboxSubmissionRecoveryMaximum = 50

// MailboxSubmissionRecoveryScanRequest bounds one advisory stale-lease scan.
// The scan never transfers a recovery capability: callers must recover by ID.
type MailboxSubmissionRecoveryScanRequest struct {
	Limit int
}

// MailboxSubmissionRecoveryDue is one content-free stale submission selected
// by the database clock. Its lease is advisory only and is re-read under lock.
type MailboxSubmissionRecoveryDue struct {
	SubmissionID uuid.UUID
	Status       MailboxSubmissionStatus
	LeaseUntil   time.Time
}

// MailboxSubmissionRecoveryScanResult is one bounded database-time selection.
type MailboxSubmissionRecoveryScanResult struct {
	SelectedAt  time.Time
	Submissions []MailboxSubmissionRecoveryDue
}

// MailboxSubmissionStaleRecoveryRequest addresses one advisory selection. It
// deliberately carries no lease, epoch, account, or draft capability.
type MailboxSubmissionStaleRecoveryRequest struct {
	SubmissionID uuid.UUID
}

// MailboxSubmissionStaleRecoveryOutcome is the closed durable outcome of one
// stale submission lease recovery.
type MailboxSubmissionStaleRecoveryOutcome string

const (
	MailboxSubmissionStaleRecoveryOutcomeNoop                   MailboxSubmissionStaleRecoveryOutcome = "noop"
	MailboxSubmissionStaleRecoveryOutcomeRequeued               MailboxSubmissionStaleRecoveryOutcome = "requeued"
	MailboxSubmissionStaleRecoveryOutcomeFailed                 MailboxSubmissionStaleRecoveryOutcome = "failed"
	MailboxSubmissionStaleRecoveryOutcomeReconciliationRequired MailboxSubmissionStaleRecoveryOutcome = "reconciliation_required"
	MailboxSubmissionStaleRecoveryOutcomeLeaseReleased          MailboxSubmissionStaleRecoveryOutcome = "lease_released"
)

// MailboxSubmissionStaleRecoveryResult contains no provider data. RetryAt is
// set only for a requeue; FailureCode is set only for terminal or
// reconciliation-required outcomes.
type MailboxSubmissionStaleRecoveryResult struct {
	Outcome     MailboxSubmissionStaleRecoveryOutcome
	RetryAt     *time.Time
	FailureCode *MailboxSubmissionErrorCode
}

// MailboxSubmissionStaleRecoveryRepository owns advisory scanning and the
// short durable recovery transaction. It performs no provider or transport I/O.
type MailboxSubmissionStaleRecoveryRepository interface {
	ScanDueMailboxSubmissionRecovery(context.Context, MailboxSubmissionRecoveryScanRequest) (MailboxSubmissionRecoveryScanResult, error)
	RecoverStaleMailboxSubmission(context.Context, MailboxSubmissionStaleRecoveryRequest) (MailboxSubmissionStaleRecoveryResult, error)
}

func mailboxSubmissionRecoveryScanRequestValid(request MailboxSubmissionRecoveryScanRequest) bool {
	return request.Limit >= 1 && request.Limit <= mailboxSubmissionRecoveryMaximum
}

func mailboxSubmissionRecoveryDueValid(due MailboxSubmissionRecoveryDue) bool {
	return due.SubmissionID != uuid.Nil &&
		(due.Status == MailboxSubmissionStatusAuthorizing || due.Status == MailboxSubmissionStatusSubmitting ||
			due.Status == MailboxSubmissionStatusUnknown || due.Status == MailboxSubmissionStatusAccepted) &&
		mailboxSubmissionClaimExactTimeValid(due.LeaseUntil)
}

func mailboxSubmissionRecoveryScanResultValid(request MailboxSubmissionRecoveryScanRequest, result MailboxSubmissionRecoveryScanResult) bool {
	if !mailboxSubmissionRecoveryScanRequestValid(request) || !mailboxSubmissionClaimExactTimeValid(result.SelectedAt) ||
		result.Submissions == nil || len(result.Submissions) > request.Limit {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(result.Submissions))
	var previous MailboxSubmissionRecoveryDue
	for index, due := range result.Submissions {
		if !mailboxSubmissionRecoveryDueValid(due) || due.LeaseUntil.After(result.SelectedAt) {
			return false
		}
		if _, duplicate := seen[due.SubmissionID]; duplicate {
			return false
		}
		seen[due.SubmissionID] = struct{}{}
		if index != 0 && (due.LeaseUntil.Before(previous.LeaseUntil) ||
			(due.LeaseUntil.Equal(previous.LeaseUntil) && bytes.Compare(due.SubmissionID[:], previous.SubmissionID[:]) <= 0)) {
			return false
		}
		previous = due
	}
	return true
}

func mailboxSubmissionRecoveryScanResultCopy(result MailboxSubmissionRecoveryScanResult) MailboxSubmissionRecoveryScanResult {
	result.SelectedAt = result.SelectedAt.Round(0).UTC()
	submissions := result.Submissions
	result.Submissions = make([]MailboxSubmissionRecoveryDue, len(submissions))
	copy(result.Submissions, submissions)
	for index := range result.Submissions {
		result.Submissions[index].LeaseUntil = result.Submissions[index].LeaseUntil.Round(0).UTC()
	}
	return result
}

func mailboxSubmissionStaleRecoveryRequestValid(request MailboxSubmissionStaleRecoveryRequest) bool {
	return request.SubmissionID != uuid.Nil
}

func mailboxSubmissionStaleRecoveryResultValid(result MailboxSubmissionStaleRecoveryResult) bool {
	switch result.Outcome {
	case MailboxSubmissionStaleRecoveryOutcomeNoop:
		return result.RetryAt == nil && result.FailureCode == nil
	case MailboxSubmissionStaleRecoveryOutcomeRequeued:
		return result.RetryAt != nil && mailboxSubmissionClaimExactTimeValid(*result.RetryAt) && result.FailureCode == nil
	case MailboxSubmissionStaleRecoveryOutcomeFailed:
		return result.RetryAt == nil && result.FailureCode != nil && *result.FailureCode == MailboxSubmissionErrorCodeRetryExhausted
	case MailboxSubmissionStaleRecoveryOutcomeReconciliationRequired:
		return result.RetryAt == nil && result.FailureCode != nil && *result.FailureCode == MailboxSubmissionErrorCodeReconciliationRequired
	case MailboxSubmissionStaleRecoveryOutcomeLeaseReleased:
		return result.RetryAt == nil && result.FailureCode == nil
	default:
		return false
	}
}

func mailboxSubmissionStaleRecoveryResultCopy(result MailboxSubmissionStaleRecoveryResult) MailboxSubmissionStaleRecoveryResult {
	if result.RetryAt != nil {
		value := result.RetryAt.Round(0).UTC()
		result.RetryAt = &value
	}
	if result.FailureCode != nil {
		value := *result.FailureCode
		result.FailureCode = &value
	}
	return result
}

func mailboxSubmissionStaleRecoveryNoop() MailboxSubmissionStaleRecoveryResult {
	return MailboxSubmissionStaleRecoveryResult{Outcome: MailboxSubmissionStaleRecoveryOutcomeNoop}
}

func mailboxSubmissionStaleRecoveryRequeued(retryAt time.Time) MailboxSubmissionStaleRecoveryResult {
	value := retryAt.Round(0).UTC()
	return MailboxSubmissionStaleRecoveryResult{Outcome: MailboxSubmissionStaleRecoveryOutcomeRequeued, RetryAt: &value}
}

func mailboxSubmissionStaleRecoveryFailed() MailboxSubmissionStaleRecoveryResult {
	code := MailboxSubmissionErrorCodeRetryExhausted
	return MailboxSubmissionStaleRecoveryResult{Outcome: MailboxSubmissionStaleRecoveryOutcomeFailed, FailureCode: &code}
}

func mailboxSubmissionStaleRecoveryReconciliationRequired() MailboxSubmissionStaleRecoveryResult {
	code := MailboxSubmissionErrorCodeReconciliationRequired
	return MailboxSubmissionStaleRecoveryResult{Outcome: MailboxSubmissionStaleRecoveryOutcomeReconciliationRequired, FailureCode: &code}
}

func mailboxSubmissionStaleRecoveryLeaseReleased() MailboxSubmissionStaleRecoveryResult {
	return MailboxSubmissionStaleRecoveryResult{Outcome: MailboxSubmissionStaleRecoveryOutcomeLeaseReleased}
}
