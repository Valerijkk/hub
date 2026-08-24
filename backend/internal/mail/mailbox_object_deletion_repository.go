package mail

import (
	"bytes"
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

const mailboxObjectDeletionClaimMaximum = 50

// mailboxObjectDeletionRepository owns durable, content-free object-deletion leases.
type mailboxObjectDeletionRepository interface {
	claimMailboxObjectDeletions(context.Context, mailboxObjectDeletionClaimRequest) (mailboxObjectDeletionClaimResult, error)
	finalizeMailboxObjectDeletion(context.Context, mailboxObjectDeletionFinalizeRequest) (mailboxObjectDeletionFinalizeResult, error)
}

// mailboxObjectDeletionClaimRequest bounds one database claim transaction.
type mailboxObjectDeletionClaimRequest struct {
	Limit         int
	LeaseDuration time.Duration
}

// mailboxObjectDeletionClaimedItem is the exact copied compare-and-swap fence for one object.
type mailboxObjectDeletionClaimedItem struct {
	ObjectID        uuid.UUID
	DeletionID      uuid.UUID
	Bucket          string
	ObjectKey       string
	DueAt           time.Time
	LeaseToken      uuid.UUID
	LeaseGeneration int64
	LeaseUntil      time.Time
	Attempt         int32
	UpdatedAt       time.Time
}

// mailboxObjectDeletionClaimResult contains one DB-time selection and a nonnil canonical item list.
type mailboxObjectDeletionClaimResult struct {
	SelectedAt time.Time
	Items      []mailboxObjectDeletionClaimedItem
}

// mailboxObjectDeletionFinalizeKind is the closed durable outcome after provider I/O.
type mailboxObjectDeletionFinalizeKind string

const (
	mailboxObjectDeletionFinalizeDeleted               mailboxObjectDeletionFinalizeKind = "deleted"
	mailboxObjectDeletionFinalizeAuthoritativelyAbsent mailboxObjectDeletionFinalizeKind = "authoritatively_absent"
	mailboxObjectDeletionFinalizeRetry                 mailboxObjectDeletionFinalizeKind = "retry"
	mailboxObjectDeletionFinalizeTerminal              mailboxObjectDeletionFinalizeKind = "terminal"
)

// mailboxObjectDeletionFinalizeRequest carries the exact claimed fence. Deleted and
// authoritatively-absent outcomes may hard-delete only on that exact CAS and zero attachment
// references. Retry clears the lease and reschedules; terminal clears it without rescheduling.
// Authoritative absence requires a separate proof: provider NotFound alone is never enough.
type mailboxObjectDeletionFinalizeRequest struct {
	Item       mailboxObjectDeletionClaimedItem
	Kind       mailboxObjectDeletionFinalizeKind
	RetryDelay time.Duration
}

// mailboxObjectDeletionFinalizeResult reports only whether the exact lease CAS applied.
type mailboxObjectDeletionFinalizeResult struct {
	Applied       bool
	RecordedAt    time.Time
	NextDeleteAt  *time.Time
}

// mailboxObjectDeletionRepositoryErrorKind is the closed, content-free repository failure family.
type mailboxObjectDeletionRepositoryErrorKind string

const (
	mailboxObjectDeletionRepositoryErrorInvalid     mailboxObjectDeletionRepositoryErrorKind = "invalid"
	mailboxObjectDeletionRepositoryErrorUnavailable mailboxObjectDeletionRepositoryErrorKind = "unavailable"
)

const mailboxObjectDeletionRepositoryErrorText = "mailbox object deletion repository error"

// mailboxObjectDeletionRepositoryError carries no provider, object, or attachment diagnostics.
type mailboxObjectDeletionRepositoryError struct {
	kind mailboxObjectDeletionRepositoryErrorKind
}

func (err *mailboxObjectDeletionRepositoryError) Error() string {
	return mailboxObjectDeletionRepositoryErrorText
}

func (err *mailboxObjectDeletionRepositoryError) Kind() mailboxObjectDeletionRepositoryErrorKind {
	if err == nil || !validMailboxObjectDeletionRepositoryErrorKind(err.kind) {
		return mailboxObjectDeletionRepositoryErrorInvalid
	}
	return err.kind
}

func mailboxObjectDeletionRepositoryErrorKindOf(err error) (mailboxObjectDeletionRepositoryErrorKind, bool) {
	var target *mailboxObjectDeletionRepositoryError
	if !errors.As(err, &target) || target == nil || !validMailboxObjectDeletionRepositoryErrorKind(target.kind) || target.kind == mailboxObjectDeletionRepositoryErrorInvalid {
		return mailboxObjectDeletionRepositoryErrorInvalid, false
	}
	return target.kind, true
}

func mailboxObjectDeletionRepositoryFailure(kind mailboxObjectDeletionRepositoryErrorKind) error {
	if !validMailboxObjectDeletionRepositoryErrorKind(kind) {
		kind = mailboxObjectDeletionRepositoryErrorInvalid
	}
	return &mailboxObjectDeletionRepositoryError{kind: kind}
}

func validMailboxObjectDeletionRepositoryErrorKind(kind mailboxObjectDeletionRepositoryErrorKind) bool {
	return kind == mailboxObjectDeletionRepositoryErrorInvalid || kind == mailboxObjectDeletionRepositoryErrorUnavailable
}

func mailboxObjectDeletionClaimRequestValid(request mailboxObjectDeletionClaimRequest) bool {
	return request.Limit >= 1 && request.Limit <= mailboxObjectDeletionClaimMaximum && request.LeaseDuration > 0 && request.LeaseDuration%time.Second == 0
}

func mailboxObjectDeletionClaimResultValid(request mailboxObjectDeletionClaimRequest, result mailboxObjectDeletionClaimResult) bool {
	if !mailboxObjectDeletionClaimRequestValid(request) || !mailboxEntValidTime(result.SelectedAt) || result.Items == nil || len(result.Items) > request.Limit {
		return false
	}

	leaseUntil, ok := mailboxEntCheckedTimeAdd(result.SelectedAt, request.LeaseDuration)
	if !ok {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(result.Items))
	var previous mailboxObjectDeletionClaimedItem
	for index := range result.Items {
		item := result.Items[index]
		if !mailboxObjectDeletionClaimedItemValid(item) || item.DueAt.After(result.SelectedAt) || !item.UpdatedAt.Equal(result.SelectedAt) || !item.LeaseUntil.Equal(leaseUntil) {
			return false
		}
		if _, duplicate := seen[item.ObjectID]; duplicate {
			return false
		}
		seen[item.ObjectID] = struct{}{}
		if index != 0 && (item.DueAt.Before(previous.DueAt) || (item.DueAt.Equal(previous.DueAt) && bytes.Compare(item.ObjectID[:], previous.ObjectID[:]) <= 0)) {
			return false
		}
		previous = item
	}
	return true
}

func mailboxObjectDeletionClaimedItemValid(item mailboxObjectDeletionClaimedItem) bool {
	return item.ObjectID != uuid.Nil && item.DeletionID != uuid.Nil && item.LeaseToken != uuid.Nil && validMailObjectDeletionIdentity(item.Bucket, item.ObjectKey) && mailboxEntValidTime(item.DueAt) && mailboxEntValidTime(item.LeaseUntil) && mailboxEntValidTime(item.UpdatedAt) && item.LeaseGeneration >= 1 && item.Attempt >= 1 && item.Attempt <= 10 && item.LeaseUntil.After(item.DueAt)
}

func mailboxObjectDeletionClaimResultCopy(result mailboxObjectDeletionClaimResult) mailboxObjectDeletionClaimResult {
	copied := result
	if result.Items == nil {
		return copied
	}
	copied.Items = make([]mailboxObjectDeletionClaimedItem, len(result.Items))
	copy(copied.Items, result.Items)
	return copied
}

func validMailboxObjectDeletionFinalizeKind(kind mailboxObjectDeletionFinalizeKind) bool {
	switch kind {
	case mailboxObjectDeletionFinalizeDeleted,
		mailboxObjectDeletionFinalizeAuthoritativelyAbsent,
		mailboxObjectDeletionFinalizeRetry,
		mailboxObjectDeletionFinalizeTerminal:
		return true
	default:
		return false
	}
}

func mailboxObjectDeletionFinalizeRequestValid(request mailboxObjectDeletionFinalizeRequest) bool {
	if !mailboxObjectDeletionClaimedItemValid(request.Item) || !validMailboxObjectDeletionFinalizeKind(request.Kind) {
		return false
	}
	switch request.Kind {
	case mailboxObjectDeletionFinalizeRetry:
		return request.Item.Attempt < 10 && request.RetryDelay > 0 && request.RetryDelay%time.Second == 0
	case mailboxObjectDeletionFinalizeTerminal:
		return request.Item.Attempt == 10 && request.RetryDelay == 0
	default:
		return request.RetryDelay == 0
	}
}

func mailboxObjectDeletionFinalizeResultValid(request mailboxObjectDeletionFinalizeRequest, result mailboxObjectDeletionFinalizeResult) bool {
	if !mailboxObjectDeletionFinalizeRequestValid(request) {
		return false
	}
	if !result.Applied {
		return result.RecordedAt.IsZero() && result.NextDeleteAt == nil
	}
	if !mailboxEntValidTime(result.RecordedAt) {
		return false
	}
	if request.Kind != mailboxObjectDeletionFinalizeRetry {
		return result.NextDeleteAt == nil
	}
	if result.NextDeleteAt == nil || !mailboxEntValidTime(*result.NextDeleteAt) {
		return false
	}
	expected, ok := mailboxEntCheckedTimeAdd(result.RecordedAt, request.RetryDelay)
	return ok && result.NextDeleteAt.Equal(expected)
}

func mailboxObjectDeletionFinalizeResultCopy(result mailboxObjectDeletionFinalizeResult) mailboxObjectDeletionFinalizeResult {
	copied := result
	if result.NextDeleteAt != nil {
		nextDeleteAt := *result.NextDeleteAt
		copied.NextDeleteAt = &nextDeleteAt
	}
	return copied
}
