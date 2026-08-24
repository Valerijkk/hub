package mail

import (
	"context"
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

var _ mailboxObjectDeletionRepository = (*MailboxEntRepository)(nil)

func (repository *MailboxEntRepository) claimMailboxObjectDeletions(ctx context.Context, request mailboxObjectDeletionClaimRequest) (mailboxObjectDeletionClaimResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxObjectDeletionClaimRequestValid(request) {
		return mailboxObjectDeletionClaimResult{}, mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorInvalid)
	}

	var result mailboxObjectDeletionClaimResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		selectedAt, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
		}
		leaseUntil, ok := mailboxEntCheckedTimeAdd(selectedAt, request.LeaseDuration)
		if !ok {
			return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorInvalid)
		}

		rows, err := transaction.MailObject.Query().
			Where(
				mailboxEntObjectDeletionSkipLocked(),
				mailobject.UploadStateIn(mailobject.UploadState("orphaned"), mailobject.UploadState("deleted")),
				mailobject.DeleteTerminalAtIsNil(),
				mailobject.NextDeleteAtNotNil(),
				mailobject.NextDeleteAtLTE(selectedAt),
				mailobject.Not(mailobject.HasAttachment()),
				mailboxEntObjectDeletionClaimable(selectedAt),
			).
			Order(ent.Asc(mailobject.FieldNextDeleteAt), ent.Asc(mailobject.FieldID)).
			Limit(request.Limit).
			All(ctx)
		if err != nil {
			return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
		}

		items := make([]mailboxObjectDeletionClaimedItem, 0, len(rows))
		for _, object := range rows {
			item, err := mailboxEntObjectDeletionClaim(ctx, object, selectedAt, leaseUntil)
			if err != nil {
				return err
			}
			items = append(items, item)
		}

		claimed := mailboxObjectDeletionClaimResult{SelectedAt: selectedAt, Items: items}
		if !mailboxObjectDeletionClaimResultValid(request, claimed) {
			return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
		}
		result = mailboxEntObjectDeletionClaimResultCopy(claimed)
		return nil
	})
	if err != nil {
		return mailboxObjectDeletionClaimResult{}, mailboxEntObjectDeletionRepositoryError(err)
	}
	if !mailboxObjectDeletionClaimResultValid(request, result) {
		return mailboxObjectDeletionClaimResult{}, mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
	}
	return mailboxEntObjectDeletionClaimResultCopy(result), nil
}

func (repository *MailboxEntRepository) finalizeMailboxObjectDeletion(ctx context.Context, request mailboxObjectDeletionFinalizeRequest) (mailboxObjectDeletionFinalizeResult, error) {
	if repository == nil || repository.client == nil || ctx == nil || !mailboxObjectDeletionFinalizeRequestValid(request) {
		return mailboxObjectDeletionFinalizeResult{}, mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorInvalid)
	}

	var result mailboxObjectDeletionFinalizeResult
	err := repository.withMailboxEntTransaction(ctx, func(transaction *ent.Tx) error {
		recordedAt, err := mailboxEntDatabaseWallClock(ctx, transaction)
		if err != nil {
			return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
		}

		object, err := transaction.MailObject.Query().
			Where(
				mailobject.IDEQ(request.Item.ObjectID),
				mailobject.DeletionIDEQ(request.Item.DeletionID),
				mailobject.BucketEQ(request.Item.Bucket),
				mailobject.ObjectKeyEQ(request.Item.ObjectKey),
				mailobject.NextDeleteAtEQ(request.Item.DueAt),
				mailobject.DeleteLeaseTokenEQ(request.Item.LeaseToken),
				mailobject.DeleteLeaseGenerationEQ(request.Item.LeaseGeneration),
				mailobject.DeleteLeaseUntilEQ(request.Item.LeaseUntil),
				mailobject.DeleteAttemptsEQ(request.Item.Attempt),
				mailobject.UpdatedAtEQ(request.Item.UpdatedAt),
				mailobject.UploadStateIn(mailobject.UploadState("orphaned"), mailobject.UploadState("deleted")),
				mailobject.DeleteTerminalAtIsNil(),
				mailobject.DeleteLeaseUntilGT(recordedAt),
			).
			ForUpdate().
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return nil
			}
			return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
		}
		if !mailboxEntObjectDeletionFinalizeRowValid(object, request.Item, recordedAt) {
			return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
		}

		references, err := transaction.MailAttachment.Query().
			Where(mailattachment.StorageObjectIDEQ(object.ID)).
			Count(ctx)
		if err != nil || references != 0 {
			return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
		}

		switch request.Kind {
		case mailboxObjectDeletionFinalizeDeleted, mailboxObjectDeletionFinalizeAuthoritativelyAbsent:
			if err := transaction.MailObject.DeleteOne(object).Exec(ctx); err != nil {
				return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
			}
			result = mailboxObjectDeletionFinalizeResult{Applied: true, RecordedAt: recordedAt}
		case mailboxObjectDeletionFinalizeRetry:
			nextDeleteAt, ok := mailboxEntCheckedTimeAdd(recordedAt, request.RetryDelay)
			if !ok {
				return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorInvalid)
			}
			if err := object.Update().
				ClearDeleteLeaseToken().
				ClearDeleteLeaseUntil().
				SetNextDeleteAt(nextDeleteAt).
				SetUpdatedAt(recordedAt).
				Exec(ctx); err != nil {
				return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
			}
			result = mailboxObjectDeletionFinalizeResult{Applied: true, RecordedAt: recordedAt, NextDeleteAt: &nextDeleteAt}
		case mailboxObjectDeletionFinalizeTerminal:
			if err := object.Update().
				ClearDeleteLeaseToken().
				ClearDeleteLeaseUntil().
				ClearNextDeleteAt().
				SetDeleteTerminalAt(recordedAt).
				SetUpdatedAt(recordedAt).
				Exec(ctx); err != nil {
				return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
			}
			result = mailboxObjectDeletionFinalizeResult{Applied: true, RecordedAt: recordedAt}
		default:
			return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorInvalid)
		}

		if !mailboxObjectDeletionFinalizeResultValid(request, result) {
			return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
		}
		result = mailboxObjectDeletionFinalizeResultCopy(result)
		return nil
	})
	if err != nil {
		return mailboxObjectDeletionFinalizeResult{}, mailboxEntObjectDeletionRepositoryError(err)
	}
	if !mailboxObjectDeletionFinalizeResultValid(request, result) {
		return mailboxObjectDeletionFinalizeResult{}, mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
	}
	return mailboxObjectDeletionFinalizeResultCopy(result), nil
}

func mailboxEntObjectDeletionClaim(ctx context.Context, object *ent.MailObject, selectedAt, leaseUntil time.Time) (mailboxObjectDeletionClaimedItem, error) {
	if ctx == nil || !mailboxEntObjectDeletionClaimRowValid(object, selectedAt) || !mailboxEntValidTime(leaseUntil) || !leaseUntil.After(selectedAt) || object.DeleteLeaseGeneration == math.MaxInt64 {
		return mailboxObjectDeletionClaimedItem{}, mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
	}

	leaseToken := uuid.New()
	if leaseToken == uuid.Nil {
		return mailboxObjectDeletionClaimedItem{}, mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
	}
	leaseGeneration := object.DeleteLeaseGeneration + 1
	attempt := object.DeleteAttempts
	if attempt < 10 {
		attempt++
	}
	dueAt := *object.NextDeleteAt
	if err := object.Update().
		SetDeleteLeaseToken(leaseToken).
		SetDeleteLeaseGeneration(leaseGeneration).
		SetDeleteLeaseUntil(leaseUntil).
		SetDeleteAttempts(attempt).
		SetUpdatedAt(selectedAt).
		Exec(ctx); err != nil {
		return mailboxObjectDeletionClaimedItem{}, mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
	}

	item := mailboxObjectDeletionClaimedItem{
		ObjectID:        object.ID,
		DeletionID:      object.DeletionID,
		Bucket:          object.Bucket,
		ObjectKey:       object.ObjectKey,
		DueAt:           dueAt,
		LeaseToken:      leaseToken,
		LeaseGeneration: leaseGeneration,
		LeaseUntil:      leaseUntil,
		Attempt:         attempt,
		UpdatedAt:       selectedAt,
	}
	if !mailboxObjectDeletionClaimedItemValid(item) {
		return mailboxObjectDeletionClaimedItem{}, mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
	}
	return item, nil
}

func mailboxEntObjectDeletionClaimRowValid(object *ent.MailObject, selectedAt time.Time) bool {
	if object == nil || object.ID == uuid.Nil || object.DeletionID == uuid.Nil || !validMailObjectDeletionIdentity(object.Bucket, object.ObjectKey) || !mailboxEntValidTime(selectedAt) || !mailboxEntValidTime(object.CreatedAt) || !mailboxEntValidTime(object.UpdatedAt) || !mailboxEntValidTime(object.ExpiresAt) || !mailboxEntValidTime(object.UploadDrainUntil) || object.NextDeleteAt == nil || !mailboxEntValidTime(*object.NextDeleteAt) || object.DeleteTerminalAt != nil || object.CreatedAt.After(object.UpdatedAt) || object.UpdatedAt.After(selectedAt) || object.NextDeleteAt.After(selectedAt) || object.NextDeleteAt.Before(object.UploadDrainUntil) || object.DeleteLeaseGeneration < 0 || object.DeleteLeaseGeneration == math.MaxInt64 {
		return false
	}
	drainUntil, ok := mailboxEntCheckedTimeAdd(object.ExpiresAt, mailboxAttachmentPresignObjectDrainLifetime)
	if !ok || !object.UploadDrainUntil.Equal(drainUntil) {
		return false
	}
	if object.UploadState != mailobject.UploadState("orphaned") && object.UploadState != mailobject.UploadState("deleted") {
		return false
	}
	if (object.DeleteLeaseToken == nil) != (object.DeleteLeaseUntil == nil) {
		return false
	}
	if object.DeleteAttempts >= 0 && object.DeleteAttempts <= 9 {
		if object.DeleteLeaseToken == nil {
			return true
		}
		return *object.DeleteLeaseToken != uuid.Nil && object.DeleteLeaseGeneration >= 1 && mailboxEntValidTime(*object.DeleteLeaseUntil) && !object.DeleteLeaseUntil.After(selectedAt)
	}
	return object.DeleteAttempts == 10 && object.DeleteLeaseToken != nil && *object.DeleteLeaseToken != uuid.Nil && object.DeleteLeaseUntil != nil && object.DeleteLeaseGeneration >= 1 && mailboxEntValidTime(*object.DeleteLeaseUntil) && !object.DeleteLeaseUntil.After(selectedAt)
}

func mailboxEntObjectDeletionFinalizeRowValid(object *ent.MailObject, item mailboxObjectDeletionClaimedItem, recordedAt time.Time) bool {
	if object == nil || !mailboxObjectDeletionClaimedItemValid(item) || !mailboxEntValidTime(recordedAt) || recordedAt.Before(item.UpdatedAt) || object.ID != item.ObjectID || object.DeletionID != item.DeletionID || object.Bucket != item.Bucket || object.ObjectKey != item.ObjectKey || object.NextDeleteAt == nil || !object.NextDeleteAt.Equal(item.DueAt) || object.DeleteLeaseToken == nil || *object.DeleteLeaseToken != item.LeaseToken || object.DeleteLeaseGeneration != item.LeaseGeneration || object.DeleteLeaseUntil == nil || !object.DeleteLeaseUntil.Equal(item.LeaseUntil) || object.DeleteAttempts != item.Attempt || !object.UpdatedAt.Equal(item.UpdatedAt) || !object.DeleteLeaseUntil.After(recordedAt) || object.DeleteTerminalAt != nil {
		return false
	}
	if object.UploadState != mailobject.UploadState("orphaned") && object.UploadState != mailobject.UploadState("deleted") {
		return false
	}
	if !mailboxEntValidTime(object.CreatedAt) || !mailboxEntValidTime(object.UpdatedAt) || !mailboxEntValidTime(object.ExpiresAt) || !mailboxEntValidTime(object.UploadDrainUntil) || !mailboxEntValidTime(*object.NextDeleteAt) || !mailboxEntValidTime(*object.DeleteLeaseUntil) || object.CreatedAt.After(object.UpdatedAt) || object.NextDeleteAt.Before(object.UploadDrainUntil) || object.DeleteLeaseGeneration < 1 || object.DeleteAttempts < 1 || object.DeleteAttempts > 10 {
		return false
	}
	drainUntil, ok := mailboxEntCheckedTimeAdd(object.ExpiresAt, mailboxAttachmentPresignObjectDrainLifetime)
	return ok && object.UploadDrainUntil.Equal(drainUntil)
}

func mailboxEntObjectDeletionSkipLocked() predicate.MailObject {
	return func(selector *sql.Selector) {
		selector.ForUpdate(sql.WithLockAction(sql.SkipLocked))
	}
}

func mailboxEntObjectDeletionClaimable(selectedAt time.Time) predicate.MailObject {
	return mailobject.Or(
		mailobject.And(
			mailobject.DeleteAttemptsGTE(0),
			mailobject.DeleteAttemptsLTE(9),
			mailobject.Or(
				mailobject.And(mailobject.DeleteLeaseTokenIsNil(), mailobject.DeleteLeaseUntilIsNil()),
				mailobject.And(
					mailobject.DeleteLeaseTokenNotNil(),
					mailobject.DeleteLeaseUntilNotNil(),
					mailobject.DeleteLeaseGenerationGTE(1),
					mailobject.DeleteLeaseUntilLTE(selectedAt),
				),
			),
		),
		mailobject.And(
			mailobject.DeleteAttemptsEQ(10),
			mailobject.DeleteLeaseTokenNotNil(),
			mailobject.DeleteLeaseUntilNotNil(),
			mailobject.DeleteLeaseGenerationGTE(1),
			mailobject.DeleteLeaseUntilLTE(selectedAt),
		),
	)
}

func mailboxEntObjectDeletionClaimResultCopy(result mailboxObjectDeletionClaimResult) mailboxObjectDeletionClaimResult {
	copied := result
	copied.Items = make([]mailboxObjectDeletionClaimedItem, len(result.Items))
	copy(copied.Items, result.Items)
	return copied
}

func mailboxEntObjectDeletionRepositoryError(err error) error {
	var repositoryError *mailboxObjectDeletionRepositoryError
	if errors.As(err, &repositoryError) && repositoryError != nil && validMailboxObjectDeletionRepositoryErrorKind(repositoryError.kind) {
		return mailboxObjectDeletionRepositoryFailure(repositoryError.kind)
	}
	return mailboxObjectDeletionRepositoryFailure(mailboxObjectDeletionRepositoryErrorUnavailable)
}
