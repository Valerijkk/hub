//go:build !qa && !release

package main

import (
	"bytes"
	"context"
	"crypto/sha256"

	"github.com/google/uuid"

	appent "github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailaccount"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/mailmessage"
)

const (
	devMailDemoBrandKitEngineMessageID = "devmail-message-brand-kit-v1"
	devMailDemoBrandKitEngineVersion   = "devmail-version-brand-kit-v1"
	devMailDemoBrandKitEngineBlobID    = "devmail-blob-brand-brief-v1"
	devMailDemoBrandKitFilename        = "mail-demo-brief.txt"
	devMailDemoBrandKitMediaType       = "text/plain"
	devMailDemoBrandKitSizeBytes int64 = 85
)

// recoverDevMailDemoBrandKitAttachment is deliberately a dev-only exception
// for the one immutable FakeMailEngine fixture. Provider-backed attachments
// have no storage scan lifecycle, so the normal scanner cannot move this row
// from pending to clean. Any drift fails closed.
func (adapter *devMailDemoAdapter) recoverDevMailDemoBrandKitAttachment(ctx context.Context, prepared devMailDemoPrepared) error {
	if adapter == nil || adapter.client == nil || !validDevMailControlContext(ctx) ||
		!prepared.matches(prepared.organizationID, prepared.pmUserID, prepared.dev1UserID) ||
		!validDevMailControlResult(prepared.control) || !validDevMailSourceResult(prepared.source) {
		return errDevMailDemoInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	dev1Control, exists := prepared.control.accounts[devMailSourceDev1Email]
	if !exists || !devMailDemoBrandKitSourceMatches(prepared.source, dev1Control) {
		return errDevMailDemoUnavailable
	}

	transaction, err := adapter.client.Tx(ctx)
	if err != nil || transaction == nil {
		return errDevMailDemoUnavailable
	}
	committed := false
	defer func() {
		if !committed {
			_ = transaction.Rollback()
		}
	}()

	account, err := transaction.MailAccount.Query().
		Where(
			mailaccount.IDEQ(dev1Control.accountID),
			mailaccount.OrganizationIDEQ(prepared.organizationID),
			mailaccount.UserIDEQ(prepared.dev1UserID),
			mailaccount.EngineAccountIDEQ(dev1Control.engineAccountID),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || account == nil ||
		account.ID != dev1Control.accountID ||
		account.OrganizationID != prepared.organizationID ||
		account.UserID != prepared.dev1UserID ||
		account.EngineAccountID == nil ||
		*account.EngineAccountID != dev1Control.engineAccountID ||
		account.Status != mailaccount.Status("active") ||
		!account.DesiredEnabled {
		return errDevMailDemoUnavailable
	}

	message, err := transaction.MailMessage.Query().
		Where(
			mailmessage.MailAccountIDEQ(account.ID),
			mailmessage.EngineMessageIDEQ(devMailDemoBrandKitEngineMessageID),
			mailmessage.EngineVersionEQ(devMailDemoBrandKitEngineVersion),
			mailmessage.DeletedAtIsNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil || message == nil ||
		message.ID == uuid.Nil ||
		message.MailAccountID != account.ID ||
		message.EngineMessageID != devMailDemoBrandKitEngineMessageID ||
		message.EngineVersion != devMailDemoBrandKitEngineVersion ||
		message.DeletedAt != nil ||
		!message.HasAttachments ||
		message.ProjectionOrigin != mailmessage.ProjectionOrigin("inbound_or_unknown") {
		return errDevMailDemoUnavailable
	}

	rows, err := transaction.MailAttachment.Query().
		Where(mailattachment.MailMessageIDEQ(message.ID)).
		Limit(2).
		ForUpdate().
		All(ctx)
	if err != nil || len(rows) != 1 ||
		!devMailDemoBrandKitAttachmentMatches(rows[0], message.ID) {
		return errDevMailDemoUnavailable
	}
	attachment := rows[0]

	requiresUpdatedAtRepair := false
	requiresCleanTransition := false
	switch attachment.ScanStatus {
	case mailattachment.ScanStatus("clean"):
		// Exact, already-recovered fixture: idempotent no-op.
		requiresUpdatedAtRepair = !validDevMailControlPersistedTime(attachment.UpdatedAt)
	case mailattachment.ScanStatus("pending"):
		requiresUpdatedAtRepair = true
		requiresCleanTransition = true
	default:
		return errDevMailDemoUnavailable
	}

	if requiresUpdatedAtRepair {
		if adapter.now == nil {
			return errDevMailDemoInvalid
		}
		recordedAt := adapter.now()
		if !validDevMailControlNow(recordedAt) {
			return errDevMailDemoInvalid
		}
		update := attachment.Update().SetUpdatedAt(recordedAt)
		if requiresCleanTransition {
			update.SetScanStatus(mailattachment.ScanStatus("clean"))
		}
		updated, err := update.Save(ctx)
		if err != nil || updated == nil ||
			updated.ScanStatus != mailattachment.ScanStatus("clean") ||
			!validDevMailControlPersistedTime(updated.UpdatedAt) ||
			!devMailDemoBrandKitAttachmentMatches(updated, message.ID) {
			return errDevMailDemoUnavailable
		}
	}

	if err := transaction.Commit(); err != nil {
		return errDevMailDemoUnavailable
	}
	committed = true
	return nil
}

func devMailDemoBrandKitSourceMatches(source devMailSourceResult, control devMailControlPlaneAccount) bool {
	sourceAccount, accountOK := source.accounts[devMailSourceDev1Email]
	message, messageOK := source.messages[devMailSourceBrandKitKey]
	return accountOK && messageOK &&
		sourceAccount.engineAccountID == control.engineAccountID &&
		message.accountEmail == devMailSourceDev1Email &&
		message.messageID == devMailDemoBrandKitEngineMessageID &&
		message.version == devMailDemoBrandKitEngineVersion &&
		message.blobID == devMailDemoBrandKitEngineBlobID
}

func devMailDemoBrandKitAttachmentMatches(row *appent.MailAttachment, messageID uuid.UUID) bool {
	expectedSHA256 := sha256.Sum256([]byte("Use the shared release palette and include the Calendar handoff in the final review.\n"))
	return row != nil &&
		row.ID != uuid.Nil &&
		row.MailMessageID != nil && *row.MailMessageID == messageID &&
		row.MailDraftID == nil &&
		row.PendingMailDraftID == nil &&
		row.DraftClientID == nil &&
		row.StorageObjectID == nil &&
		row.EngineBlobID != nil && *row.EngineBlobID == devMailDemoBrandKitEngineBlobID &&
		row.ContentID == nil &&
		row.Filename == devMailDemoBrandKitFilename &&
		row.MediaType == devMailDemoBrandKitMediaType &&
		row.SizeBytes == devMailDemoBrandKitSizeBytes &&
		row.Disposition == mailattachment.Disposition("attachment") &&
		row.Sha256 != nil && len(*row.Sha256) == sha256.Size &&
		bytes.Equal(*row.Sha256, expectedSHA256[:]) &&
		row.ScanLeaseToken == nil &&
		row.ScanLeaseGeneration == 0 &&
		row.ScanLeaseUntil == nil &&
		row.ScanAttempts == 0 &&
		row.NextScanAt == nil &&
		row.ScanTerminalAt == nil &&
		!row.CreatedAt.IsZero() &&
		!row.UpdatedAt.IsZero() &&
		!row.CreatedAt.After(row.UpdatedAt)
}
