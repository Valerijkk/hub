package mail

import (
	"context"
	"crypto/sha256"
	"errors"
	"io"
	"time"

	"github.com/google/uuid"
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

// mailboxAttachmentScanPrepareRequest carries the exact resolved caller scope into a short repository transaction.
type mailboxAttachmentScanPrepareRequest struct {
	ActorID        uuid.UUID
	OrganizationID uuid.UUID
	AccountID      uuid.UUID
	AttachmentID   uuid.UUID
}

// mailboxAttachmentScanPrepared is an opaque, copied read authority for one object stream.
type mailboxAttachmentScanPrepared struct {
	Object        MailboxObjectLocator
	ContentLength int64
	Fence         *mailboxAttachmentScanFence
}

// mailboxAttachmentScanCompleteRequest is one terminal clean or rejected compare-and-swap.
type mailboxAttachmentScanCompleteRequest struct {
	ActorID        uuid.UUID
	OrganizationID uuid.UUID
	AccountID      uuid.UUID
	AttachmentID   uuid.UUID
	ScanStatus     AttachmentScanStatus
	Fence          *mailboxAttachmentScanFence
}

// MailboxAttachmentScanResult contains no attachment content or scanner diagnostics.
type MailboxAttachmentScanResult struct {
	AttachmentID uuid.UUID
	ScanStatus   AttachmentScanStatus
}

type mailboxAttachmentScanCompletion struct {
	attachmentID uuid.UUID
	scanStatus   AttachmentScanStatus
}

// mailboxAttachmentScanRepository keeps lock-owning persistence separate from streaming I/O.
type mailboxAttachmentScanRepository interface {
	prepareMailboxAttachmentScan(context.Context, mailboxAttachmentScanPrepareRequest) (mailboxAttachmentScanPrepared, error)
	completeMailboxAttachmentScan(context.Context, mailboxAttachmentScanCompleteRequest) (mailboxAttachmentScanCompletion, error)
}

// MailboxAttachmentScanService verifies an uploaded object and makes its scanner decision durable.
type MailboxAttachmentScanService struct {
	owners      mailboxOwnerResolver
	repository  mailboxAttachmentScanRepository
	objectStore MailboxObjectStore
	scanner     MailboxAttachmentMalwareScanner
}

func NewMailboxAttachmentScanService(
	owners mailboxOwnerResolver,
	repository mailboxAttachmentScanRepository,
	objectStore MailboxObjectStore,
	scanner MailboxAttachmentMalwareScanner,
) (*MailboxAttachmentScanService, error) {
	if mailboxAttachmentStagingDependencyNil(owners) || mailboxAttachmentStagingDependencyNil(repository) || mailboxAttachmentStagingDependencyNil(objectStore) || mailboxAttachmentStagingDependencyNil(scanner) {
		return nil, mailboxAttachmentScanError(MailboxAttachmentScanErrorInvalid)
	}
	return &MailboxAttachmentScanService{owners: owners, repository: repository, objectStore: objectStore, scanner: scanner}, nil
}

// ScanMailboxAttachment streams a drain-complete, confirmed object through the malware scanner before one terminal CAS.
func (service *MailboxAttachmentScanService) ScanMailboxAttachment(ctx context.Context, actorID, attachmentID uuid.UUID) (MailboxAttachmentScanResult, error) {
	if !mailboxAttachmentScanServiceValid(service) || ctx == nil || actorID == uuid.Nil || attachmentID == uuid.Nil {
		return MailboxAttachmentScanResult{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorInvalid)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxAttachmentScanResult{}, mailboxAttachmentScanOwnerError(err)
	}
	if owner.State() == MailboxOwnerDeleting {
		return MailboxAttachmentScanResult{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorDeleting)
	}
	accountID, hasAccount := owner.AccountID()
	if owner.State() != MailboxOwnerActive || owner.OrganizationID() == uuid.Nil || !hasAccount || accountID == uuid.Nil {
		return MailboxAttachmentScanResult{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}

	prepared, err := service.repository.prepareMailboxAttachmentScan(ctx, mailboxAttachmentScanPrepareRequest{
		ActorID:        actorID,
		OrganizationID: owner.OrganizationID(),
		AccountID:      accountID,
		AttachmentID:   attachmentID,
	})
	if err != nil {
		return MailboxAttachmentScanResult{}, mailboxAttachmentScanRepositoryError(err)
	}
	if !mailboxAttachmentScanPreparedValid(prepared, actorID, owner.OrganizationID(), accountID, attachmentID) {
		return MailboxAttachmentScanResult{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}

	integrityClean, verdict, err := mailboxAttachmentScanStream(service.objectStore, service.scanner, ctx, prepared.Object, prepared.ContentLength, prepared.Fence.ExpectedSHA256)
	if err != nil {
		return MailboxAttachmentScanResult{}, err
	}
	status := AttachmentScanStatusRejected
	if integrityClean && verdict == MailboxAttachmentMalwareVerdictClean {
		status = AttachmentScanStatusClean
	}

	completed, err := service.repository.completeMailboxAttachmentScan(ctx, mailboxAttachmentScanCompleteRequest{
		ActorID:        actorID,
		OrganizationID: owner.OrganizationID(),
		AccountID:      accountID,
		AttachmentID:   attachmentID,
		ScanStatus:     status,
		Fence:          mailboxAttachmentScanFenceCopy(prepared.Fence),
	})
	if err != nil {
		return MailboxAttachmentScanResult{}, mailboxAttachmentScanRepositoryError(err)
	}
	if completed.attachmentID != attachmentID || completed.scanStatus != status {
		return MailboxAttachmentScanResult{}, mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	return MailboxAttachmentScanResult{AttachmentID: completed.attachmentID, ScanStatus: completed.scanStatus}, nil
}

func mailboxAttachmentScanServiceValid(service *MailboxAttachmentScanService) bool {
	return service != nil && !mailboxAttachmentStagingDependencyNil(service.owners) && !mailboxAttachmentStagingDependencyNil(service.repository) && !mailboxAttachmentStagingDependencyNil(service.objectStore) && !mailboxAttachmentStagingDependencyNil(service.scanner)
}

func mailboxAttachmentScanPreparedValid(prepared mailboxAttachmentScanPrepared, actorID, organizationID, accountID, attachmentID uuid.UUID) bool {
	if !validMailboxObjectLocator(prepared.Object) || prepared.ContentLength < 1 || prepared.ContentLength > MaxEngineAttachmentBytes || !mailboxAttachmentScanFenceValid(prepared.Fence) {
		return false
	}
	fence := prepared.Fence
	return fence.UploaderID == actorID && fence.OrganizationID == organizationID && fence.AccountID == accountID && fence.AttachmentID == attachmentID && prepared.ContentLength == fence.ExpectedSizeBytes && prepared.Object == (MailboxObjectLocator{Bucket: fence.Bucket, Key: fence.ObjectKey})
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

func mailboxAttachmentScanFenceCopy(fence *mailboxAttachmentScanFence) *mailboxAttachmentScanFence {
	if fence == nil {
		return nil
	}
	copied := *fence
	return &copied
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

func mailboxAttachmentScanOwnerError(err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
	switch serviceError.Kind {
	case MailboxServiceErrorInvalid:
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorInvalid)
	case MailboxServiceErrorNotFound:
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorNotFound)
	case MailboxServiceErrorDeleting:
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorDeleting)
	default:
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
	}
}

func mailboxAttachmentScanRepositoryError(err error) error {
	if kind, ok := MailboxAttachmentScanErrorKindOf(err); ok {
		return mailboxAttachmentScanError(kind)
	}
	return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
}

func mailboxAttachmentScanObjectStoreError(err error) error {
	if kind, ok := MailboxObjectStoreErrorKindOf(err); ok && kind == MailboxObjectStoreErrorNotFound {
		return mailboxAttachmentScanError(MailboxAttachmentScanErrorNotFound)
	}
	return mailboxAttachmentScanError(MailboxAttachmentScanErrorUnavailable)
}
