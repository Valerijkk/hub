package mail

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"errors"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
)

const (
	mailboxAttachmentPresignOperation                  = "attachment_presign"
	mailboxAttachmentPresignFingerprintVersion   int32 = 1
	mailboxAttachmentPresignMaximumURLLifetime          = 10 * time.Minute
	mailboxAttachmentPresignObjectLifetime              = 11 * time.Minute
	mailboxAttachmentPresignObjectDrainLifetime         = 11 * time.Minute
	mailboxAttachmentPresignReplayLifetime              = 30 * 24 * time.Hour
	mailboxAttachmentPresignMaximumUnclaimedAttachments = MaxEngineAttachmentItems
)

const mailboxAttachmentPresignRequestHashDomain = "hub.mail.attachment-presign.request.v1"

// PresignMailboxAttachment reserves one pending mail-owned object and returns its bounded upload contract.
func (service *MailboxAttachmentStagingService) PresignMailboxAttachment(ctx context.Context, actorID uuid.UUID, input MailboxAttachmentPresignInput) (MailboxAttachmentPresignResult, error) {
	if !mailboxAttachmentPresignServiceValid(service) || ctx == nil || actorID == uuid.Nil || !mailboxAttachmentPresignIdempotencyKeyValid(input.IdempotencyKey) || input.DraftClientID == uuid.Nil {
		return MailboxAttachmentPresignResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	normalized, contentID, err := mailboxAttachmentPresignNormalizedInput(input)
	if err != nil {
		return MailboxAttachmentPresignResult{}, mailboxAttachmentPresignPolicyError(err)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxAttachmentPresignResult{}, mailboxAttachmentPresignOwnerError(err)
	}
	if owner.State() == MailboxOwnerDeleting {
		return MailboxAttachmentPresignResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorDeleting)
	}
	accountID, hasAccount := owner.AccountID()
	if owner.State() != MailboxOwnerActive || !hasAccount || accountID == uuid.Nil || owner.OrganizationID() == uuid.Nil {
		return MailboxAttachmentPresignResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	attachmentID := uuid.New()
	objectID := uuid.New()
	objectKey := mailboxAttachmentPresignObjectKey(owner.OrganizationID(), accountID, objectID)
	request := MailboxAttachmentPresignRepositoryRequest{
		ActorID:            actorID,
		OrganizationID:     owner.OrganizationID(),
		AccountID:          accountID,
		IdempotencyKeyHash: sha256.Sum256(input.IdempotencyKey[:]),
		RequestHash: mailboxAttachmentPresignRequestHash(
			input.IdempotencyKey,
			actorID,
			owner.OrganizationID(),
			accountID,
			input.DraftClientID,
			normalized,
		),
		DraftClientID:      input.DraftClientID,
		AttachmentID:       attachmentID,
		ObjectID:           objectID,
		Bucket:             service.config.Bucket,
		ObjectKey:          objectKey,
		Filename:           normalized.Filename,
		MediaType:          normalized.MediaType,
		ContentID:          contentID,
		Disposition:        normalized.Disposition,
		SizeBytes:          normalized.SizeBytes,
		FingerprintVersion: mailboxAttachmentPresignFingerprintVersion,
	}
	reservation, err := service.presigns.ReserveMailboxAttachmentPresign(ctx, request)
	if err != nil {
		return MailboxAttachmentPresignResult{}, mailboxAttachmentPresignStagingError(err)
	}
	if reservation.AttachmentID == uuid.Nil || !validMailboxObjectLocator(reservation.Object) {
		return MailboxAttachmentPresignResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}

	expiresAt, ok := mailboxAttachmentPresignUploadExpiry(reservation.IssuedAt, reservation.ObjectExpiresAt)
	if !ok {
		return MailboxAttachmentPresignResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	presignRequest := MailboxObjectPresignPutRequest{
		Object:        reservation.Object,
		ContentLength: normalized.SizeBytes,
		ContentType:   normalized.MediaType,
		IssuedAt:      reservation.IssuedAt,
		ExpiresAt:     expiresAt,
	}
	if !validMailboxObjectPresignPutRequest(presignRequest) {
		return MailboxAttachmentPresignResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	presignResult, err := service.objectStore.PresignPut(ctx, presignRequest)
	if err != nil || !validMailboxObjectPresignPutResult(presignRequest, presignResult) || !mailboxAttachmentPresignResultOriginValid(service.config.PresignOrigin, presignResult.URL) {
		return MailboxAttachmentPresignResult{}, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	presignResult = cloneMailboxObjectPresignPutResult(presignResult)

	return MailboxAttachmentPresignResult{
		AttachmentID:    reservation.AttachmentID,
		UploadURL:       presignResult.URL,
		ExpiresAt:       presignResult.ExpiresAt,
		RequiredHeaders: presignResult.Headers,
	}, nil
}

func mailboxAttachmentPresignServiceValid(service *MailboxAttachmentStagingService) bool {
	if service == nil || mailboxAttachmentStagingDependencyNil(service.owners) || mailboxAttachmentStagingDependencyNil(service.presigns) || mailboxAttachmentStagingDependencyNil(service.objectStore) || service.config.Bucket == "" {
		return false
	}
	_, ok := mailboxAttachmentPresignFixedOrigin(service.config.PresignOrigin)
	return ok
}

func mailboxAttachmentPresignIdempotencyKeyValid(value uuid.UUID) bool {
	if value == uuid.Nil || value.Variant() != uuid.RFC4122 {
		return false
	}
	version := value.Version()
	return version == 4 || version == 7
}

func mailboxAttachmentPresignNormalizedInput(input MailboxAttachmentPresignInput) (AttachmentPolicyAttachment, *string, error) {
	contentID := ""
	if input.ContentID != nil {
		contentID = *input.ContentID
	}
	decision, err := EvaluateAttachmentPolicy(AttachmentPolicyInput{
		Operation:          AttachmentPolicyOperationValidate,
		MaxAttachmentBytes: MaxEngineAttachmentBytes,
		Attachments: []AttachmentPolicyAttachment{{
			Origin:      AttachmentByteOriginUpload,
			Filename:    input.Filename,
			MediaType:   input.MediaType,
			ContentID:   contentID,
			Disposition: input.Disposition,
			ScanStatus:  AttachmentScanStatusPending,
			SizeBytes:   input.SizeBytes,
		}},
	})
	if err != nil || len(decision.Attachments) != 1 {
		return AttachmentPolicyAttachment{}, nil, err
	}
	normalized := decision.Attachments[0]
	return normalized, mailboxAttachmentPresignContentIDPointer(normalized.ContentID), nil
}

func mailboxAttachmentPresignContentIDPointer(value string) *string {
	if value == "" {
		return nil
	}
	value = strings.Clone(value)
	return &value
}

func mailboxAttachmentPresignObjectKey(organizationID, accountID, objectID uuid.UUID) string {
	return "mail/" + organizationID.String() + "/" + accountID.String() + "/" + objectID.String()
}

func mailboxAttachmentPresignRequestHash(idempotencyKey, actorID, organizationID, accountID, draftClientID uuid.UUID, attachment AttachmentPolicyAttachment) [sha256.Size]byte {
	mac := hmac.New(sha256.New, idempotencyKey[:])
	_, _ = mac.Write(canonicalRecord(
		canonicalTLV(1, []byte(mailboxAttachmentPresignRequestHashDomain)),
		canonicalTLV(2, canonicalUUID(actorID)),
		canonicalTLV(3, canonicalUUID(organizationID)),
		canonicalTLV(4, canonicalUUID(accountID)),
		canonicalTLV(5, canonicalUUID(draftClientID)),
		canonicalTLV(6, []byte(attachment.Filename)),
		canonicalTLV(7, []byte(attachment.MediaType)),
		canonicalTLV(8, mailboxAttachmentPresignOptionalString(attachment.ContentID)),
		canonicalTLV(9, []byte(attachment.Disposition)),
		canonicalTLV(10, canonicalI64(attachment.SizeBytes)),
	))
	var result [sha256.Size]byte
	copy(result[:], mac.Sum(nil))
	return result
}

func mailboxAttachmentPresignOptionalString(value string) []byte {
	if value == "" {
		return []byte{0}
	}
	return append([]byte{1}, []byte(value)...)
}

func mailboxAttachmentPresignUploadExpiry(issuedAt, objectExpiresAt time.Time) (time.Time, bool) {
	if !validMailboxObjectStoreTime(issuedAt) || !validMailboxObjectStoreTime(objectExpiresAt) || !objectExpiresAt.After(issuedAt) {
		return time.Time{}, false
	}
	expiresAt := issuedAt.Add(mailboxAttachmentPresignMaximumURLLifetime)
	if !expiresAt.After(issuedAt) {
		return time.Time{}, false
	}
	if objectExpiresAt.Before(expiresAt) {
		expiresAt = objectExpiresAt
	}
	return expiresAt, validMailboxObjectStoreLifetime(issuedAt, expiresAt, mailboxAttachmentPresignMaximumURLLifetime)
}

func mailboxAttachmentPresignFixedOrigin(value string) (*url.URL, bool) {
	parsed, err := url.ParseRequestURI(value)
	if err != nil || parsed == nil || !parsed.IsAbs() || !strings.EqualFold(parsed.Scheme, "https") || parsed.Host == "" || parsed.Hostname() == "" || parsed.User != nil || parsed.Fragment != "" || parsed.Opaque != "" || parsed.RawPath != "" || parsed.RawQuery != "" || parsed.ForceQuery || parsed.Path != "" && parsed.Path != "/" {
		return nil, false
	}
	return &url.URL{Scheme: strings.ToLower(parsed.Scheme), Host: strings.ToLower(parsed.Host)}, true
}

func mailboxAttachmentPresignResultOriginValid(originValue, resultURL string) bool {
	origin, ok := mailboxAttachmentPresignFixedOrigin(originValue)
	if !ok || !validMailboxObjectStoreURL(resultURL) {
		return false
	}
	result, err := url.ParseRequestURI(resultURL)
	return err == nil && result != nil && strings.EqualFold(result.Scheme, origin.Scheme) && strings.EqualFold(result.Host, origin.Host)
}

func mailboxAttachmentPresignPolicyError(err error) error {
	var policyError *AttachmentPolicyError
	if !errors.As(err, &policyError) || policyError == nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}
	switch policyError.Kind {
	case AttachmentErrorTooLarge:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorLimit)
	case AttachmentErrorNotClean:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorNotClean)
	default:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}
}

func mailboxAttachmentPresignOwnerError(err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	switch serviceError.Kind {
	case MailboxServiceErrorInvalid:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	case MailboxServiceErrorNotFound:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorNotFound)
	case MailboxServiceErrorDeleting:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorDeleting)
	default:
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
}

func mailboxAttachmentPresignStagingError(err error) error {
	var stagingError *MailboxAttachmentStagingError
	if !errors.As(err, &stagingError) || stagingError == nil {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorUnavailable)
	}
	return mailboxAttachmentStagingError(stagingError.Kind())
}
