package mail

import (
	"bytes"
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/mailattachment"
	"github.com/valerii/hub/backend/ent/mailobject"
)

const mailboxAttachmentScanBufferBytes = 32 * 1024

// MailboxAttachmentScanErrorKind classifies one content-free attachment-scan failure.
type MailboxAttachmentScanErrorKind string

const (
	MailboxAttachmentScanErrorInvalid       MailboxAttachmentScanErrorKind = "invalid"
	MailboxAttachmentScanErrorNotFound      MailboxAttachmentScanErrorKind = "not_found"
	MailboxAttachmentScanErrorDeleting      MailboxAttachmentScanErrorKind = "deleting"
	MailboxAttachmentScanErrorNotReady      MailboxAttachmentScanErrorKind = "not_ready"
	MailboxAttachmentScanErrorUnavailable   MailboxAttachmentScanErrorKind = "unavailable"
	MailboxAttachmentScanErrorStateConflict MailboxAttachmentScanErrorKind = "state_conflict"
)

const mailboxAttachmentScanErrorText = "mailbox attachment scan error"

// MailboxAttachmentScanError deliberately contains no object, attachment, or scanner content.
type MailboxAttachmentScanError struct {
	kind MailboxAttachmentScanErrorKind
}

func NewMailboxAttachmentScanError(kind MailboxAttachmentScanErrorKind) *MailboxAttachmentScanError {
	if !validMailboxAttachmentScanErrorKind(kind) {
		kind = MailboxAttachmentScanErrorInvalid
	}
	return &MailboxAttachmentScanError{kind: kind}
}

func (*MailboxAttachmentScanError) Error() string {
	return mailboxAttachmentScanErrorText
}

func (err *MailboxAttachmentScanError) Kind() MailboxAttachmentScanErrorKind {
	if err == nil || !validMailboxAttachmentScanErrorKind(err.kind) {
		return MailboxAttachmentScanErrorInvalid
	}
	return err.kind
}

func MailboxAttachmentScanErrorKindOf(err error) (MailboxAttachmentScanErrorKind, bool) {
	var target *MailboxAttachmentScanError
	if !errors.As(err, &target) || target == nil || !validMailboxAttachmentScanErrorKind(target.kind) || target.kind == MailboxAttachmentScanErrorInvalid {
		return MailboxAttachmentScanErrorInvalid, false
	}
	return target.kind, true
}

func mailboxAttachmentScanError(kind MailboxAttachmentScanErrorKind) error {
	return NewMailboxAttachmentScanError(kind)
}

func validMailboxAttachmentScanErrorKind(kind MailboxAttachmentScanErrorKind) bool {
	switch kind {
	case MailboxAttachmentScanErrorInvalid,
		MailboxAttachmentScanErrorNotFound,
		MailboxAttachmentScanErrorDeleting,
		MailboxAttachmentScanErrorNotReady,
		MailboxAttachmentScanErrorUnavailable,
		MailboxAttachmentScanErrorStateConflict:
		return true
	default:
		return false
	}
}

// MailboxAttachmentMalwareVerdict is the closed terminal output of one scanner session.
type MailboxAttachmentMalwareVerdict string

const (
	MailboxAttachmentMalwareVerdictClean    MailboxAttachmentMalwareVerdict = "clean"
	MailboxAttachmentMalwareVerdictRejected MailboxAttachmentMalwareVerdict = "rejected"
)

// MailboxAttachmentMalwareScanSession receives one object exactly once and yields one terminal verdict after Close.
type MailboxAttachmentMalwareScanSession interface {
	io.Writer
	Close() error
	Verdict(context.Context) (MailboxAttachmentMalwareVerdict, error)
}

// MailboxAttachmentMalwareScanner creates one streaming scanner session without exposing attachment content to persistence.
type MailboxAttachmentMalwareScanner interface {
	Begin(context.Context) (MailboxAttachmentMalwareScanSession, error)
}

type mailboxAttachmentScanFence struct {
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
	AttachmentUpdatedAt time.Time
	ObjectUpdatedAt     time.Time
}

func mailboxAttachmentScanFenceValid(fence *mailboxAttachmentScanFence) bool {
	if fence == nil || fence.OrganizationID == uuid.Nil || fence.AccountID == uuid.Nil || fence.UploaderID == uuid.Nil || fence.DraftClientID == uuid.Nil || fence.AttachmentID == uuid.Nil || fence.ObjectID == uuid.Nil || fence.AttachmentID == fence.ObjectID || fence.ExpectedSizeBytes < 1 || fence.ExpectedSizeBytes > MaxEngineAttachmentBytes || !mailboxEntValidTime(fence.ObjectExpiresAt) || !mailboxEntValidTime(fence.UploadDrainUntil) || !mailboxEntValidTime(fence.AttachmentUpdatedAt) || !mailboxEntValidTime(fence.ObjectUpdatedAt) {
		return false
	}
	if !fence.ObjectExpiresAt.Before(fence.UploadDrainUntil) || !fence.UploadDrainUntil.Equal(fence.ObjectExpiresAt.Add(mailboxAttachmentPresignObjectDrainLifetime)) || fence.ObjectKey != mailboxAttachmentPresignObjectKey(fence.OrganizationID, fence.AccountID, fence.ObjectID) {
		return false
	}
	return validMailboxObjectLocator(MailboxObjectLocator{Bucket: fence.Bucket, Key: fence.ObjectKey})
}

// mailboxAttachmentScanConfirmedRowsValid protects the durable worker's claim
// and finalization paths from scanning any object other than its confirmed
// attachment object.
func mailboxAttachmentScanConfirmedRowsValid(attachment *ent.MailAttachment, object *ent.MailObject) bool {
	if attachment == nil || object == nil || attachment.ID == uuid.Nil || object.ID == uuid.Nil || attachment.MailMessageID != nil || attachment.MailDraftID != nil || attachment.PendingMailDraftID != nil || attachment.DraftClientID == nil || *attachment.DraftClientID == uuid.Nil || attachment.StorageObjectID == nil || *attachment.StorageObjectID != object.ID || attachment.EngineBlobID != nil || attachment.SizeBytes < 1 || attachment.SizeBytes > MaxEngineAttachmentBytes || attachment.ScanStatus != mailattachment.ScanStatus("pending") || len(attachment.SHA256) != sha256.Size || !mailboxEntValidTime(attachment.CreatedAt) || !mailboxEntValidTime(attachment.UpdatedAt) {
		return false
	}
	if !mailboxAttachmentConfirmPendingMetadataValid(attachment) {
		return false
	}
	if object.OrganizationID == uuid.Nil || object.MailAccountID == uuid.Nil || object.UploaderUserID == uuid.Nil || object.DraftClientID != *attachment.DraftClientID || object.Bucket == "" || object.ObjectKey != mailboxAttachmentPresignObjectKey(object.OrganizationID, object.MailAccountID, object.ID) || !validMailboxObjectLocator(MailboxObjectLocator{Bucket: object.Bucket, Key: object.ObjectKey}) || object.ExpectedSizeBytes != attachment.SizeBytes || object.ExpectedSizeBytes < 1 || object.ExpectedSizeBytes > MaxEngineAttachmentBytes || object.ActualSizeBytes == nil || *object.ActualSizeBytes != object.ExpectedSizeBytes || object.Filename != attachment.Filename || object.MediaType != attachment.MediaType || len(object.SHA256) != sha256.Size || !bytes.Equal(object.SHA256, attachment.SHA256) || object.UploadState != mailobject.UploadState("confirmed") || object.NextDeleteAt != nil || object.DeleteLeaseToken != nil || object.DeleteLeaseUntil != nil || !mailboxEntValidTime(object.ExpiresAt) || !mailboxEntValidTime(object.UploadDrainUntil) || !object.UploadDrainUntil.Equal(object.ExpiresAt.Add(mailboxAttachmentPresignObjectDrainLifetime)) || !mailboxEntValidTime(object.CreatedAt) || !mailboxEntValidTime(object.UpdatedAt) {
		return false
	}
	return true
}

type mailboxAttachmentScanIntegrity uint8

const (
	mailboxAttachmentScanIntegrityUnknown mailboxAttachmentScanIntegrity = iota
	mailboxAttachmentScanIntegrityExact
	mailboxAttachmentScanIntegrityRejected
)

func mailboxAttachmentScanStream(objectStore MailboxObjectStore, scanner MailboxAttachmentMalwareScanner, ctx context.Context, object MailboxObjectLocator, expectedSizeBytes int64, expectedSHA256 [sha256.Size]byte) (bool, MailboxAttachmentMalwareVerdict, error) {
	if mailboxAttachmentStagingDependencyNil(objectStore) || mailboxAttachmentStagingDependencyNil(scanner) || ctx == nil || !validMailboxObjectLocator(object) || expectedSizeBytes < 1 || expectedSizeBytes > MaxEngineAttachmentBytes {
		return false, "", mailboxAttachmentScanError(MailboxAttachmentScanErrorInvalid)
	}

	session, err := scanner.Begin(ctx)
	if err != nil {
		if !mailboxAttachmentStagingDependencyNil(session) {
			_ = session.Close()
		}
		return false, "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	if mailboxAttachmentStagingDependencyNil(session) {
		return false, "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}

	stream, err := objectStore.Open(ctx, MailboxObjectReadRequest{Object: object})
	if err != nil {
		if !mailboxAttachmentStagingDependencyNil(stream.Body) {
			_ = stream.Body.Close()
		}
		if !mailboxAttachmentStagingDependencyNil(session) && session.Close() != nil {
			return false, "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
		}
		return false, "", mailboxAttachmentScanObjectStoreError(err)
	}
	if mailboxAttachmentStagingDependencyNil(stream.Body) {
		if !mailboxAttachmentStagingDependencyNil(session) && session.Close() != nil {
			return false, "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
		}
		return false, "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}

	integrity, copyErr := mailboxAttachmentScanCopyAndHash(stream.Body, session, stream.ContentLength, expectedSizeBytes, expectedSHA256)
	var streamCloseErr error
	if mailboxAttachmentStagingDependencyNil(stream.Body) {
		streamCloseErr = errors.New("mailbox attachment scan stream unavailable")
	} else {
		streamCloseErr = stream.Body.Close()
	}
	var scannerCloseErr error
	if mailboxAttachmentStagingDependencyNil(session) {
		scannerCloseErr = errors.New("mailbox attachment scan session unavailable")
	} else {
		scannerCloseErr = session.Close()
	}
	var verdict MailboxAttachmentMalwareVerdict
	var verdictErr error
	if mailboxAttachmentStagingDependencyNil(session) {
		verdictErr = errors.New("mailbox attachment scan session unavailable")
	} else {
		verdict, verdictErr = session.Verdict(ctx)
	}

	if integrity == mailboxAttachmentScanIntegrityRejected {
		return false, MailboxAttachmentMalwareVerdictRejected, nil
	}
	if copyErr != nil || streamCloseErr != nil || scannerCloseErr != nil || verdictErr != nil || integrity != mailboxAttachmentScanIntegrityExact || !validMailboxAttachmentMalwareVerdict(verdict) {
		return false, "", mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	if verdict == MailboxAttachmentMalwareVerdictRejected {
		return false, MailboxAttachmentMalwareVerdictRejected, nil
	}
	return true, verdict, nil
}

func mailboxAttachmentScanCopyAndHash(body io.Reader, scanner io.Writer, contentLength, expectedSizeBytes int64, expectedSHA256 [sha256.Size]byte) (mailboxAttachmentScanIntegrity, error) {
	if mailboxAttachmentStagingDependencyNil(body) || mailboxAttachmentStagingDependencyNil(scanner) || contentLength < 0 || expectedSizeBytes < 1 || expectedSizeBytes > MaxEngineAttachmentBytes {
		return mailboxAttachmentScanIntegrityUnknown, errors.New("mailbox attachment scan stream unavailable")
	}

	integrity := mailboxAttachmentScanIntegrityExact
	if contentLength != expectedSizeBytes {
		integrity = mailboxAttachmentScanIntegrityRejected
	}
	hash := sha256.New()
	var buffer [mailboxAttachmentScanBufferBytes]byte
	var sizeBytes int64
	for {
		remaining := expectedSizeBytes - sizeBytes
		limit := len(buffer)
		if remaining < int64(limit) {
			limit = int(remaining + 1)
		}
		if limit < 1 || mailboxAttachmentStagingDependencyNil(body) || mailboxAttachmentStagingDependencyNil(scanner) {
			return integrity, errors.New("mailbox attachment scan stream unavailable")
		}

		read, readErr := body.Read(buffer[:limit])
		if read < 0 || read > limit {
			return integrity, errors.New("mailbox attachment scan stream unavailable")
		}
		if read > 0 {
			if _, err := hash.Write(buffer[:read]); err != nil {
				return integrity, errors.New("mailbox attachment scan stream unavailable")
			}
			written, writeErr := scanner.Write(buffer[:read])
			if writeErr != nil || written != read {
				return integrity, errors.New("mailbox attachment scan stream unavailable")
			}
			sizeBytes += int64(read)
		}
		if sizeBytes > expectedSizeBytes {
			return mailboxAttachmentScanIntegrityRejected, nil
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil || read == 0 {
			return integrity, errors.New("mailbox attachment scan stream unavailable")
		}
	}
	if sizeBytes != expectedSizeBytes {
		integrity = mailboxAttachmentScanIntegrityRejected
	}
	var actualSHA256 [sha256.Size]byte
	copy(actualSHA256[:], hash.Sum(nil))
	if actualSHA256 != expectedSHA256 {
		integrity = mailboxAttachmentScanIntegrityRejected
	}
	return integrity, nil
}

func validMailboxAttachmentMalwareVerdict(verdict MailboxAttachmentMalwareVerdict) bool {
	return verdict == MailboxAttachmentMalwareVerdictClean || verdict == MailboxAttachmentMalwareVerdictRejected
}

func mailboxAttachmentScanObjectStoreError(err error) error {
	if kind, ok := MailboxObjectStoreErrorKindOf(err); ok && kind == MailboxObjectStoreErrorNotFound {
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorNotFound)
	}
	return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
}
