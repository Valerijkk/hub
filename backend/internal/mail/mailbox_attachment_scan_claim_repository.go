package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"time"

	"github.com/google/uuid"
)

const mailboxAttachmentScanClaimMaximum = 50

// mailboxAttachmentScanClaimRepository owns durable, content-free attachment-scan leases.
type mailboxAttachmentScanClaimRepository interface {
	claimMailboxAttachmentScans(context.Context, mailboxAttachmentScanClaimRequest) (mailboxAttachmentScanClaimResult, error)
	finalizeMailboxAttachmentScan(context.Context, mailboxAttachmentScanFinalizeRequest) (mailboxAttachmentScanFinalizeResult, error)
}

// mailboxAttachmentScanClaimRequest bounds one database claim transaction.
type mailboxAttachmentScanClaimRequest struct {
	Limit         int
	LeaseDuration time.Duration
}

// mailboxAttachmentScanClaimedItem is the exact copied compare-and-swap fence for one attachment scan.
type mailboxAttachmentScanClaimedItem struct {
	OrganizationID      uuid.UUID
	AccountID           uuid.UUID
	UploaderID          uuid.UUID
	DraftClientID       uuid.UUID
	AttachmentID        uuid.UUID
	ObjectID            uuid.UUID
	Bucket              string
	ObjectKey           string
	ExpectedSizeBytes   int64
	ExpectedSHA256      [sha256.Size]byte
	ObjectExpiresAt     time.Time
	UploadDrainUntil    time.Time
	DueAt               time.Time
	LeaseToken          uuid.UUID
	LeaseGeneration     int64
	LeaseUntil          time.Time
	Attempt             int32
	AttachmentUpdatedAt time.Time
	ObjectUpdatedAt     time.Time
}

// mailboxAttachmentScanClaimResult contains one DB-time selection and a nonnil canonical item list.
type mailboxAttachmentScanClaimResult struct {
	SelectedAt time.Time
	Items      []mailboxAttachmentScanClaimedItem
}

// mailboxAttachmentScanFinalizeKind is the closed durable outcome after scanner I/O.
type mailboxAttachmentScanFinalizeKind string

const (
	mailboxAttachmentScanFinalizeClean    mailboxAttachmentScanFinalizeKind = "clean"
	mailboxAttachmentScanFinalizeRejected mailboxAttachmentScanFinalizeKind = "rejected"
	mailboxAttachmentScanFinalizeRetry    mailboxAttachmentScanFinalizeKind = "retry"
	mailboxAttachmentScanFinalizeTerminal mailboxAttachmentScanFinalizeKind = "terminal"
)

// mailboxAttachmentScanFinalizeRequest carries the exact claimed fence.
type mailboxAttachmentScanFinalizeRequest struct {
	Item       mailboxAttachmentScanClaimedItem
	Kind       mailboxAttachmentScanFinalizeKind
	RetryDelay time.Duration
}

// mailboxAttachmentScanFinalizeResult reports only whether the exact lease CAS applied.
type mailboxAttachmentScanFinalizeResult struct {
	Applied     bool
	RecordedAt  time.Time
	NextScanAt *time.Time
}

// mailboxAttachmentScanClaimRepositoryErrorKind is the closed, content-free repository failure family.
type mailboxAttachmentScanClaimRepositoryErrorKind string

const (
	mailboxAttachmentScanClaimRepositoryErrorInvalid     mailboxAttachmentScanClaimRepositoryErrorKind = "invalid"
	mailboxAttachmentScanClaimRepositoryErrorUnavailable mailboxAttachmentScanClaimRepositoryErrorKind = "unavailable"
)

const mailboxAttachmentScanClaimRepositoryErrorText = "mailbox attachment scan claim repository error"

// mailboxAttachmentScanClaimRepositoryError carries no attachment, object, or scanner diagnostics.
type mailboxAttachmentScanClaimRepositoryError struct {
	kind mailboxAttachmentScanClaimRepositoryErrorKind
}

func (err *mailboxAttachmentScanClaimRepositoryError) Error() string {
	return mailboxAttachmentScanClaimRepositoryErrorText
}

func (err *mailboxAttachmentScanClaimRepositoryError) Kind() mailboxAttachmentScanClaimRepositoryErrorKind {
	if err == nil || !validMailboxAttachmentScanClaimRepositoryErrorKind(err.kind) {
		return mailboxAttachmentScanClaimRepositoryErrorInvalid
	}
	return err.kind
}

func mailboxAttachmentScanClaimRepositoryErrorKindOf(err error) (mailboxAttachmentScanClaimRepositoryErrorKind, bool) {
	var target *mailboxAttachmentScanClaimRepositoryError
	if !errors.As(err, &target) || target == nil || !validMailboxAttachmentScanClaimRepositoryErrorKind(target.kind) || target.kind == mailboxAttachmentScanClaimRepositoryErrorInvalid {
		return mailboxAttachmentScanClaimRepositoryErrorInvalid, false
	}
	return target.kind, true
}

func mailboxAttachmentScanClaimRepositoryFailure(kind mailboxAttachmentScanClaimRepositoryErrorKind) error {
	if !validMailboxAttachmentScanClaimRepositoryErrorKind(kind) {
		kind = mailboxAttachmentScanClaimRepositoryErrorInvalid
	}
	return &mailboxAttachmentScanClaimRepositoryError{kind: kind}
}

func validMailboxAttachmentScanClaimRepositoryErrorKind(kind mailboxAttachmentScanClaimRepositoryErrorKind) bool {
	return kind == mailboxAttachmentScanClaimRepositoryErrorInvalid || kind == mailboxAttachmentScanClaimRepositoryErrorUnavailable
}

func mailboxAttachmentScanClaimRequestValid(request mailboxAttachmentScanClaimRequest) bool {
	return request.Limit >= 1 && request.Limit <= mailboxAttachmentScanClaimMaximum && request.LeaseDuration > 0 && request.LeaseDuration%time.Second == 0
}

func mailboxAttachmentScanClaimRequestCopy(request mailboxAttachmentScanClaimRequest) mailboxAttachmentScanClaimRequest {
	return request
}

func mailboxAttachmentScanClaimResultValid(request mailboxAttachmentScanClaimRequest, result mailboxAttachmentScanClaimResult) bool {
	if !mailboxAttachmentScanClaimRequestValid(request) || !mailboxEntValidTime(result.SelectedAt) || result.Items == nil || len(result.Items) > request.Limit {
		return false
	}

	leaseUntil, ok := mailboxEntCheckedTimeAdd(result.SelectedAt, request.LeaseDuration)
	if !ok {
		return false
	}
	seen := make(map[uuid.UUID]struct{}, len(result.Items))
	var previous mailboxAttachmentScanClaimedItem
	for index := range result.Items {
		item := result.Items[index]
		if !mailboxAttachmentScanClaimedItemValid(item) || item.DueAt.After(result.SelectedAt) || !item.AttachmentUpdatedAt.Equal(result.SelectedAt) || !item.LeaseUntil.Equal(leaseUntil) {
			return false
		}
		if _, duplicate := seen[item.AttachmentID]; duplicate {
			return false
		}
		seen[item.AttachmentID] = struct{}{}
		if index != 0 && (item.DueAt.Before(previous.DueAt) || (item.DueAt.Equal(previous.DueAt) && bytes.Compare(item.AttachmentID[:], previous.AttachmentID[:]) <= 0)) {
			return false
		}
		previous = item
	}
	return true
}

func mailboxAttachmentScanClaimedItemValid(item mailboxAttachmentScanClaimedItem) bool {
	fence := mailboxAttachmentScanFence{
		OrganizationID:      item.OrganizationID,
		AccountID:           item.AccountID,
		UploaderID:          item.UploaderID,
		DraftClientID:       item.DraftClientID,
		AttachmentID:        item.AttachmentID,
		ObjectID:            item.ObjectID,
		Bucket:              item.Bucket,
		ObjectKey:           item.ObjectKey,
		ExpectedSizeBytes:   item.ExpectedSizeBytes,
		ExpectedSHA256:      item.ExpectedSHA256,
		ObjectExpiresAt:     item.ObjectExpiresAt,
		UploadDrainUntil:    item.UploadDrainUntil,
		AttachmentUpdatedAt: item.AttachmentUpdatedAt,
		ObjectUpdatedAt:     item.ObjectUpdatedAt,
	}
	return mailboxAttachmentScanFenceValid(&fence) && mailboxEntValidTime(item.DueAt) && mailboxEntValidTime(item.LeaseUntil) && !item.DueAt.Before(item.UploadDrainUntil) && item.LeaseToken != uuid.Nil && item.LeaseGeneration >= 1 && item.Attempt >= 1 && item.Attempt <= 10 && item.LeaseUntil.After(item.DueAt)
}

func mailboxAttachmentScanClaimedItemCopy(item mailboxAttachmentScanClaimedItem) mailboxAttachmentScanClaimedItem {
	return item
}

func mailboxAttachmentScanClaimResultCopy(result mailboxAttachmentScanClaimResult) mailboxAttachmentScanClaimResult {
	copied := result
	if result.Items == nil {
		return copied
	}
	copied.Items = make([]mailboxAttachmentScanClaimedItem, len(result.Items))
	for index := range result.Items {
		copied.Items[index] = mailboxAttachmentScanClaimedItemCopy(result.Items[index])
	}
	return copied
}

func validMailboxAttachmentScanFinalizeKind(kind mailboxAttachmentScanFinalizeKind) bool {
	switch kind {
	case mailboxAttachmentScanFinalizeClean,
		mailboxAttachmentScanFinalizeRejected,
		mailboxAttachmentScanFinalizeRetry,
		mailboxAttachmentScanFinalizeTerminal:
		return true
	default:
		return false
	}
}

func mailboxAttachmentScanFinalizeRequestValid(request mailboxAttachmentScanFinalizeRequest) bool {
	if !mailboxAttachmentScanClaimedItemValid(request.Item) || !validMailboxAttachmentScanFinalizeKind(request.Kind) {
		return false
	}
	switch request.Kind {
	case mailboxAttachmentScanFinalizeRetry:
		return request.Item.Attempt < 10 && request.RetryDelay > 0 && request.RetryDelay%time.Second == 0
	case mailboxAttachmentScanFinalizeTerminal:
		return request.Item.Attempt == 10 && request.RetryDelay == 0
	default:
		return request.RetryDelay == 0
	}
}

func mailboxAttachmentScanFinalizeRequestCopy(request mailboxAttachmentScanFinalizeRequest) mailboxAttachmentScanFinalizeRequest {
	copied := request
	copied.Item = mailboxAttachmentScanClaimedItemCopy(request.Item)
	return copied
}

func mailboxAttachmentScanFinalizeResultValid(request mailboxAttachmentScanFinalizeRequest, result mailboxAttachmentScanFinalizeResult) bool {
	if !mailboxAttachmentScanFinalizeRequestValid(request) {
		return false
	}
	if !result.Applied {
		return result.RecordedAt.IsZero() && result.NextScanAt == nil
	}
	if !mailboxEntValidTime(result.RecordedAt) {
		return false
	}
	if request.Kind != mailboxAttachmentScanFinalizeRetry {
		return result.NextScanAt == nil
	}
	if result.NextScanAt == nil || !mailboxEntValidTime(*result.NextScanAt) {
		return false
	}
	expected, ok := mailboxEntCheckedTimeAdd(result.RecordedAt, request.RetryDelay)
	return ok && result.NextScanAt.Equal(expected)
}

func mailboxAttachmentScanFinalizeResultCopy(result mailboxAttachmentScanFinalizeResult) mailboxAttachmentScanFinalizeResult {
	copied := result
	if result.NextScanAt != nil {
		nextScanAt := *result.NextScanAt
		copied.NextScanAt = &nextScanAt
	}
	return copied
}
