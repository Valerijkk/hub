package mail

import (
	"bytes"
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailmutationreceipt"
)

const (
	mailboxDraftRecoveryReceiptFingerprintVersion int32 = 1
	mailboxDraftRecoveryReceiptReplayLifetime           = 30 * 24 * time.Hour
)

type mailboxDraftRecoveryReceiptOperation string

const (
	mailboxDraftRecoveryReceiptOperationRestart mailboxDraftRecoveryReceiptOperation = "draft_restart"
	mailboxDraftRecoveryReceiptOperationAbandon mailboxDraftRecoveryReceiptOperation = "draft_abandon"
)

// mailboxDraftRecoveryReceiptInput is intentionally sealed to the two recovery operations.
// It carries only idempotency bindings and never draft content, provider values, or child targets.
type mailboxDraftRecoveryReceiptInput struct {
	Operation          mailboxDraftRecoveryReceiptOperation
	DraftID            uuid.UUID
	IdempotencyKeyHash [32]byte
	RequestHash        [32]byte
	FingerprintVersion int32
}

func mailboxEntDraftRecoveryReceiptInputValid(input mailboxDraftRecoveryReceiptInput) bool {
	return (input.Operation == mailboxDraftRecoveryReceiptOperationRestart || input.Operation == mailboxDraftRecoveryReceiptOperationAbandon) && input.DraftID != uuid.Nil && input.FingerprintVersion == mailboxDraftRecoveryReceiptFingerprintVersion && mailboxEntDraftCreateHashPresent(input.IdempotencyKeyHash) && mailboxEntDraftCreateHashPresent(input.RequestHash)
}

func mailboxEntDraftRecoveryReceiptAccountValid(account *ent.MailAccount) bool {
	return account != nil && account.ID != uuid.Nil && account.OrganizationID != uuid.Nil && account.UserID != uuid.Nil
}

// mailboxEntDraftRecoveryReceiptLock reads the sole account/actor/operation/key receipt under
// the canonical receipt lock. A nil receipt with an empty kind means that no replay exists.
func mailboxEntDraftRecoveryReceiptLock(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, input mailboxDraftRecoveryReceiptInput) (*ent.MailMutationReceipt, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil || !mailboxEntDraftRecoveryReceiptAccountValid(account) || !mailboxEntDraftRecoveryReceiptInputValid(input) {
		return nil, MailboxDraftRepositoryErrorInvalid
	}
	receipts, err := transaction.MailMutationReceipt.Query().
		Where(
			mailmutationreceipt.OrganizationIDEQ(account.OrganizationID),
			mailmutationreceipt.ActorUserIDEQ(account.UserID),
			mailmutationreceipt.OperationEQ(mailmutationreceipt.Operation(input.Operation)),
			mailmutationreceipt.IdempotencyKeyHashEQ(input.IdempotencyKeyHash[:]),
		).
		ForUpdate().
		All(ctx)
	if err != nil || len(receipts) > 1 {
		return nil, MailboxDraftRepositoryErrorUnavailable
	}
	if len(receipts) == 0 {
		return nil, ""
	}
	if receipts[0] == nil {
		return nil, MailboxDraftRepositoryErrorIdempotencyConflict
	}
	return receipts[0], ""
}

// mailboxEntDraftRecoveryReceiptReplayValid admits only the terminal, un-compacted,
// exact-account receipt shape shared by restart and abandon. Callers must additionally
// validate the operation-specific current draft post-state before treating it as a replay.
func mailboxEntDraftRecoveryReceiptReplayValid(receipt *ent.MailMutationReceipt, account *ent.MailAccount, input mailboxDraftRecoveryReceiptInput, databaseNow time.Time) bool {
	if receipt == nil || !mailboxEntDraftRecoveryReceiptAccountValid(account) || !mailboxEntDraftRecoveryReceiptInputValid(input) || !mailboxEntValidTime(databaseNow) {
		return false
	}
	if receipt.ID == uuid.Nil || receipt.OrganizationID != account.OrganizationID || receipt.ActorUserID != account.UserID || receipt.MailAccountID == nil || *receipt.MailAccountID != account.ID || receipt.Operation != mailmutationreceipt.Operation(input.Operation) || receipt.KeySource != mailmutationreceipt.KeySource("client") || !bytes.Equal(receipt.IdempotencyKeyHash, input.IdempotencyKeyHash[:]) || !bytes.Equal(receipt.RequestHash, input.RequestHash[:]) || receipt.FingerprintVersion != input.FingerprintVersion {
		return false
	}
	if receipt.State != mailmutationreceipt.State("completed") || receipt.TargetKind == nil || *receipt.TargetKind != mailmutationreceipt.TargetKind("draft") || receipt.TargetID == nil || *receipt.TargetID != input.DraftID || receipt.BulkAction != nil || receipt.ArgumentKind != nil || receipt.ArgumentID != nil || receipt.SafeCode != nil || receipt.OriginalHTTPStatus == nil || *receipt.OriginalHTTPStatus != 200 || receipt.ReconcileAttempts != 0 || receipt.NextReconcileAt != nil || receipt.ReconcileLeaseToken != nil || receipt.ReconcileLeaseUntil != nil || receipt.TerminalAt == nil || !mailboxEntValidTime(*receipt.TerminalAt) || receipt.ReplayUntil == nil || !mailboxEntValidTime(*receipt.ReplayUntil) || !receipt.ReplayUntil.Equal(receipt.TerminalAt.Add(mailboxDraftRecoveryReceiptReplayLifetime)) || !databaseNow.Before(*receipt.ReplayUntil) || receipt.CompactedAt != nil {
		return false
	}
	return true
}

// mailboxEntDraftRecoveryReceiptComplete records one terminal recovery mutation.
// It deliberately creates no MailMutationReceiptTarget rows: the parent receipt's draft
// target is the complete idempotency proof for these singleton commands.
func mailboxEntDraftRecoveryReceiptComplete(ctx context.Context, transaction *ent.Tx, account *ent.MailAccount, input mailboxDraftRecoveryReceiptInput, databaseNow time.Time) (*ent.MailMutationReceipt, MailboxDraftRepositoryErrorKind) {
	if ctx == nil || transaction == nil || !mailboxEntDraftRecoveryReceiptAccountValid(account) || !mailboxEntDraftRecoveryReceiptInputValid(input) || !mailboxEntValidTime(databaseNow) {
		return nil, MailboxDraftRepositoryErrorInvalid
	}
	replayUntil, ok := mailboxEntCheckedTimeAdd(databaseNow, mailboxDraftRecoveryReceiptReplayLifetime)
	if !ok {
		return nil, MailboxDraftRepositoryErrorUnavailable
	}
	receipt, err := transaction.MailMutationReceipt.Create().
		SetOrganizationID(account.OrganizationID).
		SetActorUserID(account.UserID).
		SetMailAccountID(account.ID).
		SetOperation(mailmutationreceipt.Operation(input.Operation)).
		SetKeySource(mailmutationreceipt.KeySource("client")).
		SetState(mailmutationreceipt.State("completed")).
		SetIdempotencyKeyHash(append([]byte(nil), input.IdempotencyKeyHash[:]...)).
		SetRequestHash(append([]byte(nil), input.RequestHash[:]...)).
		SetFingerprintVersion(input.FingerprintVersion).
		SetTargetKind(mailmutationreceipt.TargetKind("draft")).
		SetTargetID(input.DraftID).
		SetOriginalHTTPStatus(200).
		SetTerminalAt(databaseNow).
		SetReplayUntil(replayUntil).
		SetCreatedAt(databaseNow).
		SetUpdatedAt(databaseNow).
		Save(ctx)
	if err != nil || !mailboxEntDraftRecoveryReceiptReplayValid(receipt, account, input, databaseNow) {
		return nil, MailboxDraftRepositoryErrorUnavailable
	}
	return receipt, ""
}
