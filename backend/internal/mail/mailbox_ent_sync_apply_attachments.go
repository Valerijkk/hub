package mail

import (
	"bytes"
	"context"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailattachment"
)

const (
	mailboxEntAttachmentDefaultFilename  = "attachment"
	mailboxEntAttachmentDefaultMediaType = "application/octet-stream"
)

type mailboxEntApplyAttachmentIntent struct {
	blobID      string
	contentID   *string
	filename    string
	mediaType   string
	sizeBytes   int64
	disposition mailattachment.Disposition
	sha256      []byte
}

func mailboxEntApplyMessageAttachments(ctx context.Context, tx *ent.Tx, accountID, messageID uuid.UUID, desired []EngineProjectionAttachment, transactionTime time.Time) (semanticChanged bool, hasAttachments bool, err error) {
	if ctx == nil || tx == nil || accountID == uuid.Nil || messageID == uuid.Nil || desired == nil || len(desired) > MaxEngineAttachmentItems || !mailboxEntValidTime(transactionTime) {
		return false, false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if err := ctx.Err(); err != nil {
		return false, false, mailboxEntSafeError(err)
	}

	desiredIntents, err := mailboxEntApplyAttachmentIntents(desired)
	if err != nil {
		return false, false, err
	}
	if err := mailboxEntApplyLockLiveMessage(ctx, tx, accountID, messageID); err != nil {
		return false, false, err
	}

	existingRows, err := tx.MailAttachment.Query().
		Where(mailattachment.MailMessageIDEQ(messageID)).
		Order(ent.Asc(mailattachment.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return false, false, mailboxEntSafeError(err)
	}
	if len(existingRows) > MaxEngineAttachmentItems {
		return false, false, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	existingProviderByBlobID := make(map[string]*ent.MailAttachment, len(existingRows))
	hasStorageBacked := false
	for _, row := range existingRows {
		providerBacked, err := mailboxEntApplyAttachmentRowValid(row, messageID)
		if err != nil {
			return false, false, err
		}
		if !providerBacked {
			hasStorageBacked = true
			continue
		}
		if _, duplicate := existingProviderByBlobID[*row.EngineBlobID]; duplicate {
			return false, false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		existingProviderByBlobID[*row.EngineBlobID] = row
	}
	for _, attachment := range desired {
		intent := desiredIntents[attachment.BlobID]
		if row, exists := existingProviderByBlobID[intent.blobID]; exists && !mailboxEntApplyAttachmentImmutableMatch(row, messageID, intent) {
			return false, false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}

	obsoleteProviderIDs := make([]uuid.UUID, 0, len(existingProviderByBlobID))
	for _, row := range existingRows {
		if row.EngineBlobID == nil {
			continue
		}
		if _, present := desiredIntents[*row.EngineBlobID]; !present {
			obsoleteProviderIDs = append(obsoleteProviderIDs, row.ID)
		}
	}
	if len(obsoleteProviderIDs) != 0 {
		deleted, err := tx.MailAttachment.Delete().
			Where(mailattachment.IDIn(obsoleteProviderIDs...)).
			Exec(ctx)
		if err != nil {
			return false, false, mailboxEntSafeError(err)
		}
		if deleted != len(obsoleteProviderIDs) {
			return false, false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		semanticChanged = true
	}

	for _, attachment := range desired {
		intent := desiredIntents[attachment.BlobID]
		_, exists := existingProviderByBlobID[intent.blobID]
		if exists {
			continue
		}

		create := tx.MailAttachment.Create().
			SetMailMessageID(messageID).
			SetEngineBlobID(intent.blobID).
			SetFilename(intent.filename).
			SetMediaType(intent.mediaType).
			SetSizeBytes(intent.sizeBytes).
			SetDisposition(intent.disposition).
			SetScanStatus(mailattachment.ScanStatus("pending")).
			SetCreatedAt(transactionTime).
			SetUpdatedAt(transactionTime)
		if intent.contentID != nil {
			create.SetContentID(*intent.contentID)
		}
		if len(intent.sha256) != 0 {
			create.SetSha256(append([]byte(nil), intent.sha256...))
		}
		if _, err := create.Save(ctx); err != nil {
			return false, false, mailboxEntSafeError(err)
		}
		semanticChanged = true
	}

	return semanticChanged, len(desired) != 0 || hasStorageBacked, nil
}

func mailboxEntApplyAttachmentIntents(desired []EngineProjectionAttachment) (map[string]mailboxEntApplyAttachmentIntent, error) {
	intents := make(map[string]mailboxEntApplyAttachmentIntent, len(desired))
	for index, attachment := range desired {
		if !syncValidOpaque(attachment.BlobID, MaxEngineOpaqueIDBytes) || (index > 0 && desired[index-1].BlobID >= attachment.BlobID) {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := intents[attachment.BlobID]; duplicate {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}

		contentID, err := mailboxEntApplyAttachmentOptionalText(attachment.ContentID)
		if err != nil {
			return nil, err
		}
		filename := attachment.Filename
		if filename == "" {
			filename = mailboxEntAttachmentDefaultFilename
		}
		mediaType := attachment.MediaType
		if mediaType == "" {
			mediaType = mailboxEntAttachmentDefaultMediaType
		}
		if !mailboxEntApplyAttachmentTextValid(filename, false) || !mailboxEntApplyAttachmentTextValid(mediaType, false) || attachment.SizeBytes < 0 || attachment.SizeBytes > MaxEngineAttachmentBytes || (len(attachment.SHA256) != 0 && len(attachment.SHA256) != AttachmentSHA256Bytes) || (attachment.Disposition != EngineAttachmentDispositionAttachment && attachment.Disposition != EngineAttachmentDispositionInline) || (attachment.Disposition == EngineAttachmentDispositionInline && contentID == nil) {
			return nil, mailboxServiceError(MailboxServiceErrorInvalid)
		}

		intents[attachment.BlobID] = mailboxEntApplyAttachmentIntent{
			blobID:      attachment.BlobID,
			contentID:   contentID,
			filename:    filename,
			mediaType:   mediaType,
			sizeBytes:   attachment.SizeBytes,
			disposition: mailattachment.Disposition(attachment.Disposition),
			sha256:      append([]byte(nil), attachment.SHA256...),
		}
	}
	return intents, nil
}

func mailboxEntApplyAttachmentRowValid(row *ent.MailAttachment, messageID uuid.UUID) (bool, error) {
	if row == nil || row.ID == uuid.Nil || row.MailMessageID == nil || *row.MailMessageID != messageID || row.MailDraftID != nil || row.CreatedAt.IsZero() || row.UpdatedAt.IsZero() || !mailboxEntApplyAttachmentOptionalTextValid(row.ContentID) || !mailboxEntApplyAttachmentTextValid(row.Filename, false) || !mailboxEntApplyAttachmentTextValid(row.MediaType, false) || row.SizeBytes < 0 || row.SizeBytes > MaxEngineAttachmentBytes || !mailboxEntApplyAttachmentDispositionValid(row.Disposition) || !mailboxEntApplyAttachmentScanStatusValid(row.ScanStatus) || (len(row.Sha256) != 0 && len(row.Sha256) != AttachmentSHA256Bytes) || ((row.ScanStatus == mailattachment.ScanStatus("clean") || row.ScanStatus == mailattachment.ScanStatus("rejected")) && len(row.Sha256) != AttachmentSHA256Bytes) {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if row.Disposition == mailattachment.Disposition(EngineAttachmentDispositionInline) && row.ContentID == nil {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	providerBacked := row.EngineBlobID != nil
	storageBacked := row.StorageObjectID != nil
	if providerBacked == storageBacked {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if providerBacked {
		if !syncValidOpaque(*row.EngineBlobID, MaxEngineOpaqueIDBytes) || row.DraftClientID != nil {
			return false, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		return true, nil
	}
	if *row.StorageObjectID == uuid.Nil || row.DraftClientID == nil || *row.DraftClientID == uuid.Nil || row.SizeBytes < 1 {
		return false, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return false, nil
}

func mailboxEntApplyAttachmentImmutableMatch(row *ent.MailAttachment, messageID uuid.UUID, intent mailboxEntApplyAttachmentIntent) bool {
	return row != nil &&
		row.MailMessageID != nil && *row.MailMessageID == messageID &&
		row.MailDraftID == nil &&
		row.DraftClientID == nil &&
		row.StorageObjectID == nil &&
		row.EngineBlobID != nil && *row.EngineBlobID == intent.blobID &&
		mailboxEntApplyAttachmentOptionalTextEqual(row.ContentID, intent.contentID) &&
		row.Filename == intent.filename &&
		row.MediaType == intent.mediaType &&
		row.SizeBytes == intent.sizeBytes &&
		row.Disposition == intent.disposition &&
		bytes.Equal(row.Sha256, intent.sha256)
}

func mailboxEntApplyAttachmentOptionalText(value string) (*string, error) {
	if value == "" {
		return nil, nil
	}
	if !mailboxEntApplyAttachmentTextValid(value, false) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	result := value
	return &result, nil
}

func mailboxEntApplyAttachmentOptionalTextValid(value *string) bool {
	return value == nil || mailboxEntApplyAttachmentTextValid(*value, false)
}

func mailboxEntApplyAttachmentOptionalTextEqual(left, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func mailboxEntApplyAttachmentTextValid(value string, allowEmpty bool) bool {
	if (!allowEmpty && value == "") || len(value) > MaxAttachmentTextBytes || !utf8.ValidString(value) || utf8.RuneCountInString(value) > MaxAttachmentTextRunes {
		return false
	}
	for _, character := range value {
		if unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func mailboxEntApplyAttachmentDispositionValid(value mailattachment.Disposition) bool {
	return value == mailattachment.Disposition(EngineAttachmentDispositionAttachment) || value == mailattachment.Disposition(EngineAttachmentDispositionInline)
}

func mailboxEntApplyAttachmentScanStatusValid(value mailattachment.ScanStatus) bool {
	return value == mailattachment.ScanStatus("pending") || value == mailattachment.ScanStatus("clean") || value == mailattachment.ScanStatus("rejected") || value == mailattachment.ScanStatus("error")
}
