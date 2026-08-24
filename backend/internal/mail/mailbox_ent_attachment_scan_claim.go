package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"math"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/mailobject"
	"github.com/valerii/hub/backend/ent/predicate"
)

var _ mailboxAttachmentScanClaimRepository = (*MailboxEntRepository)(nil)

func (repository *MailboxEntRepository) claimMailboxAttachmentScans(ctx context.Context, request mailboxAttachmentScanClaimRequest) (mailboxAttachmentScanClaimResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxAttachmentScanClaimRequestValid(request) {
		return mailboxAttachmentScanClaimResult{}, mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorInvalid)
	}
	request = mailboxAttachmentScanClaimRequestCopy(request)

	var result mailboxAttachmentScanClaimResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		selectedAt, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
		}
		leaseUntil, ok := mailboxEntCheckedTimeAdd(selectedAt, request.LeaseDuration)
		if !ok {
			return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorInvalid)
		}

		rows, err := transaction.MailAttachment.Query().
			Where(
				mailboxEntAttachmentScanSkipLocked(),
				mailattachment.StorageObjectIDNotNil(),
				mailattachment.ScanStatusEQ(mailattachment.ScanStatus("pending")),
				mailattachment.SHA256NotNil(),
				mailattachment.ScanTerminalAtIsNil(),
				mailattachment.NextScanAtNotNil(),
				mailattachment.NextScanAtLTE(selectedAt),
				mailattachment.MailMessageIDIsNil(),
				mailattachment.MailDraftIDIsNil(),
				mailattachment.PendingMailDraftIDIsNil(),
				mailattachment.EngineBlobIDIsNil(),
				mailboxEntAttachmentScanClaimable(selectedAt),
			).
			Order(ent.Asc(mailattachment.FieldNextScanAt), ent.Asc(mailattachment.FieldID)).
			Limit(request.Limit).
			All(ctx)
		if err != nil {
			return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
		}

		items := make([]mailboxAttachmentScanClaimedItem, 0, len(rows))
		for _, attachment := range rows {
			item, err := mailboxEntAttachmentScanClaim(ctx, transaction, attachment, selectedAt, leaseUntil)
			if err != nil {
				return err
			}
			items = append(items, item)
		}

		claimed := mailboxAttachmentScanClaimResult{SelectedAt: selectedAt, Items: items}
		if !mailboxAttachmentScanClaimResultValid(request, claimed) {
			return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
		}
		result = mailboxAttachmentScanClaimResultCopy(claimed)
		return nil
	})
	if err != nil {
		return mailboxAttachmentScanClaimResult{}, mailboxEntAttachmentScanClaimRepositoryError(err)
	}
	if !mailboxAttachmentScanClaimResultValid(request, result) {
		return mailboxAttachmentScanClaimResult{}, mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
	}
	return mailboxAttachmentScanClaimResultCopy(result), nil
}

func (repository *MailboxEntRepository) finalizeMailboxAttachmentScan(ctx context.Context, request mailboxAttachmentScanFinalizeRequest) (mailboxAttachmentScanFinalizeResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxAttachmentScanFinalizeRequestValid(request) {
		return mailboxAttachmentScanFinalizeResult{}, mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorInvalid)
	}
	request = mailboxAttachmentScanFinalizeRequestCopy(request)

	var result mailboxAttachmentScanFinalizeResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		recordedAt, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
		}

		attachment, err := transaction.MailAttachment.Query().
			Where(
				mailattachment.IDEQ(request.Item.AttachmentID),
				mailattachment.DraftClientIDEQ(request.Item.DraftClientID),
				mailattachment.StorageObjectIDEQ(request.Item.ObjectID),
				mailattachment.MailMessageIDIsNil(),
				mailattachment.MailDraftIDIsNil(),
				mailattachment.PendingMailDraftIDIsNil(),
				mailattachment.EngineBlobIDIsNil(),
				mailattachment.SizeBytesEQ(request.Item.ExpectedSizeBytes),
				mailattachment.ScanStatusEQ(mailattachment.ScanStatus("pending")),
				mailattachment.SHA256EQ(request.Item.ExpectedSHA256[:]),
				mailattachment.ScanTerminalAtIsNil(),
				mailattachment.NextScanAtEQ(request.Item.DueAt),
				mailattachment.ScanLeaseTokenEQ(request.Item.LeaseToken),
				mailattachment.ScanLeaseGenerationEQ(request.Item.LeaseGeneration),
				mailattachment.ScanLeaseUntilEQ(request.Item.LeaseUntil),
				mailattachment.ScanAttemptsEQ(request.Item.Attempt),
				mailattachment.UpdatedAtEQ(request.Item.AttachmentUpdatedAt),
				mailattachment.ScanLeaseUntilGT(recordedAt),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
		}

		object, err := transaction.MailObject.Query().
			Where(mailobject.IDEQ(request.Item.ObjectID)).
			ForUpdate().
			Only(ctx)
		if err != nil {
			return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
		}
		if !mailboxEntAttachmentScanFinalizeRowsValid(attachment, object, request.Item, recordedAt) {
			return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
		}

		switch request.Kind {
		case mailboxAttachmentScanFinalizeClean, mailboxAttachmentScanFinalizeRejected:
			status := mailattachment.ScanStatus(request.Kind)
			if err := attachment.Update().
				SetScanStatus(status).
				ClearScanLeaseToken().
				ClearScanLeaseUntil().
				ClearNextScanAt().
				SetScanTerminalAt(recordedAt).
				SetUpdatedAt(recordedAt).
				Exec(ctx); err != nil {
				return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
			}
			result = mailboxAttachmentScanFinalizeResult{Applied: true, RecordedAt: recordedAt}
		case mailboxAttachmentScanFinalizeRetry:
			nextScanAt, ok := mailboxEntCheckedTimeAdd(recordedAt, request.RetryDelay)
			if !ok {
				return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorInvalid)
			}
			if err := attachment.Update().
				ClearScanLeaseToken().
				ClearScanLeaseUntil().
				SetNextScanAt(nextScanAt).
				SetUpdatedAt(recordedAt).
				Exec(ctx); err != nil {
				return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
			}
			result = mailboxAttachmentScanFinalizeResult{Applied: true, RecordedAt: recordedAt, NextScanAt: &nextScanAt}
		case mailboxAttachmentScanFinalizeTerminal:
			if err := attachment.Update().
				SetScanStatus(mailattachment.ScanStatus("error")).
				ClearScanLeaseToken().
				ClearScanLeaseUntil().
				ClearNextScanAt().
				SetScanTerminalAt(recordedAt).
				SetUpdatedAt(recordedAt).
				Exec(ctx); err != nil {
				return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
			}
			result = mailboxAttachmentScanFinalizeResult{Applied: true, RecordedAt: recordedAt}
		default:
			return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorInvalid)
		}

		if !mailboxAttachmentScanFinalizeResultValid(request, result) {
			return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
		}
		result = mailboxAttachmentScanFinalizeResultCopy(result)
		return nil
	})
	if err != nil {
		return mailboxAttachmentScanFinalizeResult{}, mailboxEntAttachmentScanClaimRepositoryError(err)
	}
	if !mailboxAttachmentScanFinalizeResultValid(request, result) {
		return mailboxAttachmentScanFinalizeResult{}, mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
	}
	return mailboxAttachmentScanFinalizeResultCopy(result), nil
}

func mailboxEntAttachmentScanClaim(ctx context.Context, transaction *ent.Tx, attachment *ent.MailAttachment, selectedAt, leaseUntil time.Time) (mailboxAttachmentScanClaimedItem, error) {
	if ctx == nil || transaction == nil || attachment == nil || attachment.StorageObjectID == nil || *attachment.StorageObjectID == uuid.Nil || !mailboxEntValidTime(selectedAt) || !mailboxEntValidTime(leaseUntil) || !leaseUntil.After(selectedAt) {
		return mailboxAttachmentScanClaimedItem{}, mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
	}

	object, err := transaction.MailObject.Query().
		Where(mailobject.IDEQ(*attachment.StorageObjectID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return mailboxAttachmentScanClaimedItem{}, mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
	}
	if !mailboxEntAttachmentScanClaimRowsValid(attachment, object, selectedAt) || attachment.ScanLeaseGeneration == math.MaxInt64 {
		return mailboxAttachmentScanClaimedItem{}, mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
	}

	leaseToken := uuid.New()
	if leaseToken == uuid.Nil {
		return mailboxAttachmentScanClaimedItem{}, mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
	}
	leaseGeneration := attachment.ScanLeaseGeneration + 1
	attempt := attachment.ScanAttempts
	if attempt < 10 {
		attempt++
	}
	dueAt := *attachment.NextScanAt
	if err := attachment.Update().
		SetScanLeaseToken(leaseToken).
		SetScanLeaseGeneration(leaseGeneration).
		SetScanLeaseUntil(leaseUntil).
		SetScanAttempts(attempt).
		SetUpdatedAt(selectedAt).
		Exec(ctx); err != nil {
		return mailboxAttachmentScanClaimedItem{}, mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
	}

	var digest [sha256.Size]byte
	copy(digest[:], attachment.SHA256)
	item := mailboxAttachmentScanClaimedItem{
		OrganizationID:      object.OrganizationID,
		AccountID:           object.MailAccountID,
		UploaderID:          object.UploaderUserID,
		DraftClientID:       object.DraftClientID,
		AttachmentID:        attachment.ID,
		ObjectID:            object.ID,
		Bucket:              object.Bucket,
		ObjectKey:           object.ObjectKey,
		ExpectedSizeBytes:   object.ExpectedSizeBytes,
		ExpectedSHA256:      digest,
		ObjectExpiresAt:     object.ExpiresAt,
		UploadDrainUntil:    object.UploadDrainUntil,
		DueAt:               dueAt,
		LeaseToken:          leaseToken,
		LeaseGeneration:     leaseGeneration,
		LeaseUntil:          leaseUntil,
		Attempt:             attempt,
		AttachmentUpdatedAt: selectedAt,
		ObjectUpdatedAt:     object.UpdatedAt,
	}
	if !mailboxAttachmentScanClaimedItemValid(item) {
		return mailboxAttachmentScanClaimedItem{}, mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
	}
	return mailboxAttachmentScanClaimedItemCopy(item), nil
}

func mailboxEntAttachmentScanClaimRowsValid(attachment *ent.MailAttachment, object *ent.MailObject, selectedAt time.Time) bool {
	if !mailboxAttachmentScanConfirmedRowsValid(attachment, object) || !mailboxEntAttachmentScanObjectAvailable(object) || !mailboxEntValidTime(selectedAt) || attachment.NextScanAt == nil || !mailboxEntValidTime(*attachment.NextScanAt) || attachment.NextScanAt.After(selectedAt) || attachment.NextScanAt.Before(object.UploadDrainUntil) || attachment.ScanTerminalAt != nil || attachment.CreatedAt.After(attachment.UpdatedAt) || attachment.UpdatedAt.After(selectedAt) || object.CreatedAt.After(object.UpdatedAt) || object.UpdatedAt.After(selectedAt) || object.ID == attachment.ID {
		return false
	}
	return mailboxEntAttachmentScanLeaseClaimableRow(attachment, selectedAt)
}

func mailboxEntAttachmentScanLeaseClaimableRow(attachment *ent.MailAttachment, selectedAt time.Time) bool {
	if attachment == nil || !mailboxEntValidTime(selectedAt) || attachment.ScanLeaseGeneration < 0 || attachment.ScanLeaseGeneration == math.MaxInt64 || attachment.ScanAttempts < 0 || attachment.ScanAttempts > 10 || (attachment.ScanLeaseToken == nil) != (attachment.ScanLeaseUntil == nil) {
		return false
	}
	if attachment.ScanAttempts <= 9 {
		if attachment.ScanLeaseToken == nil {
			return true
		}
		return *attachment.ScanLeaseToken != uuid.Nil && attachment.ScanLeaseGeneration >= 1 && mailboxEntValidTime(*attachment.ScanLeaseUntil) && !attachment.ScanLeaseUntil.After(selectedAt)
	}
	return attachment.ScanAttempts == 10 && attachment.ScanLeaseToken != nil && *attachment.ScanLeaseToken != uuid.Nil && attachment.ScanLeaseUntil != nil && attachment.ScanLeaseGeneration >= 1 && mailboxEntValidTime(*attachment.ScanLeaseUntil) && !attachment.ScanLeaseUntil.After(selectedAt)
}

func mailboxEntAttachmentScanObjectAvailable(object *ent.MailObject) bool {
	if object == nil || object.ID == uuid.Nil || object.OrganizationID == uuid.Nil || object.MailAccountID == uuid.Nil || object.UploaderUserID == uuid.Nil || object.DraftClientID == uuid.Nil || object.DeleteTerminalAt != nil || object.DeleteLeaseGeneration != 0 || object.DeleteAttempts != 0 {
		return false
	}
	return object.NextDeleteAt == nil && object.DeleteLeaseToken == nil && object.DeleteLeaseUntil == nil
}

func mailboxEntAttachmentScanFinalizeRowsValid(attachment *ent.MailAttachment, object *ent.MailObject, item mailboxAttachmentScanClaimedItem, recordedAt time.Time) bool {
	if !mailboxAttachmentScanClaimedItemValid(item) || !mailboxEntValidTime(recordedAt) || recordedAt.Before(item.AttachmentUpdatedAt) || attachment == nil || attachment.ID != item.AttachmentID || attachment.DraftClientID == nil || *attachment.DraftClientID != item.DraftClientID || attachment.StorageObjectID == nil || *attachment.StorageObjectID != item.ObjectID || attachment.MailMessageID != nil || attachment.MailDraftID != nil || attachment.PendingMailDraftID != nil || attachment.EngineBlobID != nil || attachment.SizeBytes != item.ExpectedSizeBytes || attachment.ScanStatus != mailattachment.ScanStatus("pending") || !bytes.Equal(attachment.SHA256, item.ExpectedSHA256[:]) || attachment.ScanTerminalAt != nil || attachment.NextScanAt == nil || !attachment.NextScanAt.Equal(item.DueAt) || attachment.ScanLeaseToken == nil || *attachment.ScanLeaseToken != item.LeaseToken || attachment.ScanLeaseGeneration != item.LeaseGeneration || attachment.ScanLeaseUntil == nil || !attachment.ScanLeaseUntil.Equal(item.LeaseUntil) || attachment.ScanAttempts != item.Attempt || !attachment.UpdatedAt.Equal(item.AttachmentUpdatedAt) || !attachment.ScanLeaseUntil.After(recordedAt) {
		return false
	}
	if !mailboxAttachmentScanConfirmedRowsValid(attachment, object) || !mailboxEntAttachmentScanObjectAvailable(object) || object.ID != item.ObjectID || object.OrganizationID != item.OrganizationID || object.MailAccountID != item.AccountID || object.UploaderUserID != item.UploaderID || object.DraftClientID != item.DraftClientID || object.Bucket != item.Bucket || object.ObjectKey != item.ObjectKey || object.ExpectedSizeBytes != item.ExpectedSizeBytes || object.ActualSizeBytes == nil || *object.ActualSizeBytes != item.ExpectedSizeBytes || !bytes.Equal(object.SHA256, item.ExpectedSHA256[:]) || !object.ExpiresAt.Equal(item.ObjectExpiresAt) || !object.UploadDrainUntil.Equal(item.UploadDrainUntil) || !object.UpdatedAt.Equal(item.ObjectUpdatedAt) {
		return false
	}
	return attachment.CreatedAt.Before(attachment.UpdatedAt) || attachment.CreatedAt.Equal(attachment.UpdatedAt)
}

func mailboxEntAttachmentScanSkipLocked() predicate.MailAttachment {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.SkipLocked))
	}
}

func mailboxEntAttachmentScanClaimable(selectedAt time.Time) predicate.MailAttachment {
	return mailattachment.Or(
		mailattachment.And(
			mailattachment.ScanAttemptsGTE(0),
			mailattachment.ScanAttemptsLTE(9),
			mailattachment.Or(
				mailattachment.And(mailattachment.ScanLeaseTokenIsNil(), mailattachment.ScanLeaseUntilIsNil()),
				mailattachment.And(
					mailattachment.ScanLeaseTokenNotNil(),
					mailattachment.ScanLeaseUntilNotNil(),
					mailattachment.ScanLeaseGenerationGTE(1),
					mailattachment.ScanLeaseUntilLTE(selectedAt),
				),
			),
		),
		mailattachment.And(
			mailattachment.ScanAttemptsEQ(10),
			mailattachment.ScanLeaseTokenNotNil(),
			mailattachment.ScanLeaseUntilNotNil(),
			mailattachment.ScanLeaseGenerationGTE(1),
			mailattachment.ScanLeaseUntilLTE(selectedAt),
		),
	)
}

func mailboxEntAttachmentScanClaimRepositoryError(err error) error {
	var repositoryError *mailboxAttachmentScanClaimRepositoryError
	if errors.As(err, &repositoryError) && repositoryError != nil && validMailboxAttachmentScanClaimRepositoryErrorKind(repositoryError.kind) {
		return mailboxAttachmentScanClaimRepositoryFailure(repositoryError.kind)
	}
	return mailboxAttachmentScanClaimRepositoryFailure(mailboxAttachmentScanClaimRepositoryErrorUnavailable)
}
