package mail

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const mailboxAttachmentStagingJSONBodyMaxBytes = 16 << 10

// MailboxAttachmentStagingHandler exposes authenticated staging operations for
// attachments that are still outside the durable mailbox projection.
type MailboxAttachmentStagingHandler struct {
	service        MailboxAttachmentStagingServiceAPI
	abandonService MailboxAttachmentAbandonServiceAPI
}

func NewMailboxAttachmentStagingHandler(
	service MailboxAttachmentStagingServiceAPI,
	abandonService MailboxAttachmentAbandonServiceAPI,
) (*MailboxAttachmentStagingHandler, error) {
	if mailboxHTTPDependencyNil(service) || mailboxHTTPDependencyNil(abandonService) {
		return nil, mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	return &MailboxAttachmentStagingHandler{service: service, abandonService: abandonService}, nil
}

func (handler *MailboxAttachmentStagingHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil ||
		mailboxHTTPDependencyNil(handler.service) ||
		mailboxHTTPDependencyNil(handler.abandonService) ||
		mailboxHTTPDependencyNil(api) ||
		mailboxHTTPDependencyNil(authMiddleware) {
		return mailboxAttachmentStagingError(MailboxAttachmentStagingErrorInvalid)
	}

	routes := api.Group("/mail", authMiddleware)
	routes.Post("/attachments/presign", handler.presign)
	routes.Post("/attachments/:attachmentId/confirm", handler.confirm)
	routes.Get("/attachments/:attachmentId/download", handler.download)
	routes.Delete("/attachments/:attachmentId", handler.abandon)
	return nil
}

type mailboxAttachmentPresignHTTPRequest struct {
	DraftClientID string                      `json:"draft_client_id"`
	Filename      string                      `json:"filename"`
	MediaType     string                      `json:"media_type"`
	SizeBytes     int64                       `json:"size_bytes"`
	Disposition   EngineAttachmentDisposition `json:"disposition"`
	ContentID     *string                     `json:"content_id"`
}

type mailboxAttachmentConfirmHTTPRequest struct {
	SHA256    string `json:"sha256"`
	SizeBytes int64  `json:"size_bytes"`
}

type mailboxAttachmentPresignHTTPResponse struct {
	AttachmentID    string            `json:"attachment_id"`
	UploadURL       string            `json:"upload_url"`
	ExpiresAt       time.Time         `json:"expires_at"`
	RequiredHeaders map[string]string `json:"required_headers"`
}

type mailboxAttachmentHTTPResponse struct {
	ID           string                      `json:"id"`
	Filename     string                      `json:"filename"`
	MediaType    string                      `json:"media_type"`
	SizeBytes    int64                       `json:"size_bytes"`
	Disposition  EngineAttachmentDisposition `json:"disposition"`
	ContentID    *string                     `json:"content_id"`
	ScanStatus   AttachmentScanStatus        `json:"scan_status"`
	SHA256       *string                     `json:"sha256"`
}

type mailboxAttachmentDownloadHTTPResponse struct {
	URL       string    `json:"url"`
	ExpiresAt time.Time `json:"expires_at"`
}

func (handler *MailboxAttachmentStagingHandler) presign(ctx fiber.Ctx) error {
	if err := mailboxAttachmentStagingRejectQuery(ctx); err != nil {
		return err
	}

	actor, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	idempotencyKey, err := mailboxAttachmentStagingIdempotencyKey(ctx)
	if err != nil {
		return err
	}

	var request mailboxAttachmentPresignHTTPRequest
	if err := mailboxAttachmentStagingDecodeJSON(ctx, &request); err != nil {
		return err
	}
	draftClientID, err := mailboxAttachmentStagingUUID(request.DraftClientID)
	if err != nil || request.Filename == "" || request.MediaType == "" || request.SizeBytes < 1 {
		return mailboxAttachmentStagingInvalidRequest()
	}

	result, err := handler.service.PresignMailboxAttachment(ctx.Context(), actor, MailboxAttachmentPresignInput{
		IdempotencyKey: idempotencyKey,
		DraftClientID:  draftClientID,
		Filename:       request.Filename,
		MediaType:      request.MediaType,
		ContentID:      request.ContentID,
		Disposition:    request.Disposition,
		SizeBytes:      request.SizeBytes,
	})
	if err != nil {
		return mailboxAttachmentStagingServiceError(ctx, err)
	}
	requiredHeaders, valid := mailboxAttachmentStagingHeaders(result.RequiredHeaders)
	if result.AttachmentID == uuid.Nil ||
		!validMailboxObjectStoreURL(result.UploadURL) ||
		!mailboxAttachmentStagingExpiryValid(result.ExpiresAt) ||
		!valid {
		return mailboxAttachmentStagingInternalError()
	}

	return ctx.Status(http.StatusOK).JSON(mailboxAttachmentPresignHTTPResponse{
		AttachmentID:    result.AttachmentID.String(),
		UploadURL:       result.UploadURL,
		ExpiresAt:       result.ExpiresAt,
		RequiredHeaders: requiredHeaders,
	})
}

func (handler *MailboxAttachmentStagingHandler) confirm(ctx fiber.Ctx) error {
	if err := mailboxAttachmentStagingRejectQuery(ctx); err != nil {
		return err
	}

	actor, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	attachmentID, err := mailboxAttachmentStagingPathUUID(ctx)
	if err != nil {
		return err
	}

	var request mailboxAttachmentConfirmHTTPRequest
	if err := mailboxAttachmentStagingDecodeJSON(ctx, &request); err != nil {
		return err
	}
	digest, err := mailboxAttachmentStagingSHA256(request.SHA256)
	if err != nil || request.SizeBytes < 1 {
		return mailboxAttachmentStagingInvalidRequest()
	}

	result, err := handler.service.ConfirmMailboxAttachment(ctx.Context(), actor, attachmentID, MailboxAttachmentConfirmInput{
		SHA256:    digest,
		SizeBytes: request.SizeBytes,
	})
	if err != nil {
		return mailboxAttachmentStagingServiceError(ctx, err)
	}
	if !mailboxAttachmentConfirmViewValid(result) {
		return mailboxAttachmentStagingInternalError()
	}

	return ctx.Status(http.StatusOK).JSON(mailboxAttachmentStagingViewResponse(result))
}

func (handler *MailboxAttachmentStagingHandler) download(ctx fiber.Ctx) error {
	if err := mailboxAttachmentStagingRejectQuery(ctx); err != nil {
		return err
	}

	actor, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	attachmentID, err := mailboxAttachmentStagingPathUUID(ctx)
	if err != nil {
		return err
	}

	result, err := handler.service.PresignMailboxAttachmentDownload(ctx.Context(), actor, attachmentID)
	if err != nil {
		return mailboxAttachmentStagingServiceError(ctx, err)
	}
	if !validMailboxObjectStoreURL(result.URL) || !mailboxAttachmentStagingExpiryValid(result.ExpiresAt) {
		return mailboxAttachmentStagingInternalError()
	}

	return ctx.Status(http.StatusOK).JSON(mailboxAttachmentDownloadHTTPResponse{
		URL:       result.URL,
		ExpiresAt: result.ExpiresAt,
	})
}

func (handler *MailboxAttachmentStagingHandler) abandon(ctx fiber.Ctx) error {
	if err := mailboxAttachmentStagingRejectQuery(ctx); err != nil {
		return err
	}
	if err := mailboxAttachmentStagingRejectBody(ctx); err != nil {
		return err
	}

	actor, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	attachmentID, err := mailboxAttachmentStagingPathUUID(ctx)
	if err != nil {
		return err
	}

	if err := handler.abandonService.AbandonMailboxAttachment(ctx.Context(), actor, attachmentID); err != nil {
		return mailboxAttachmentStagingAbandonServiceError(ctx, err)
	}
	return ctx.SendStatus(http.StatusNoContent)
}

func mailboxAttachmentStagingRejectQuery(ctx fiber.Ctx) error {
	if len(ctx.Request().URI().QueryString()) != 0 {
		return mailboxAttachmentStagingInvalidRequest()
	}
	return nil
}

func mailboxAttachmentStagingRejectBody(ctx fiber.Ctx) error {
	if len(ctx.BodyRaw()) != 0 || strings.TrimSpace(ctx.Get(fiber.HeaderContentType)) != "" {
		return mailboxAttachmentStagingInvalidRequest()
	}
	return nil
}

func mailboxAttachmentStagingDecodeJSON(ctx fiber.Ctx, target any) error {
	mediaType, _, err := mime.ParseMediaType(strings.TrimSpace(ctx.Get(fiber.HeaderContentType)))
	if err != nil || !strings.EqualFold(mediaType, "application/json") {
		return fiber.NewError(http.StatusUnsupportedMediaType, "unsupported media type")
	}

	rawBody := ctx.BodyRaw()
	if len(rawBody) > mailboxAttachmentStagingJSONBodyMaxBytes {
		return fiber.NewError(http.StatusRequestEntityTooLarge, "request body too large")
	}
	body := bytes.TrimSpace(rawBody)
	if len(body) == 0 || body[0] != '{' {
		return mailboxAttachmentStagingInvalidRequest()
	}

	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return mailboxAttachmentStagingInvalidRequest()
	}
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		return mailboxAttachmentStagingInvalidRequest()
	}
	return nil
}

func mailboxAttachmentStagingIdempotencyKey(ctx fiber.Ctx) (uuid.UUID, error) {
	key, err := mailboxAttachmentStagingUUID(ctx.Get("Idempotency-Key"))
	if err != nil || key == uuid.Nil || (key.Version() != 4 && key.Version() != 7) {
		return uuid.Nil, mailboxAttachmentStagingInvalidRequest()
	}
	return key, nil
}

func mailboxAttachmentStagingPathUUID(ctx fiber.Ctx) (uuid.UUID, error) {
	return mailboxAttachmentStagingUUID(ctx.Params("attachmentId"))
}

func mailboxAttachmentStagingUUID(raw string) (uuid.UUID, error) {
	value, err := uuid.Parse(raw)
	if err != nil || value == uuid.Nil || raw != value.String() {
		return uuid.Nil, mailboxAttachmentStagingInvalidRequest()
	}
	return value, nil
}

func mailboxAttachmentStagingSHA256(raw string) ([sha256.Size]byte, error) {
	var digest [sha256.Size]byte
	if len(raw) != hex.EncodedLen(sha256.Size) {
		return digest, mailboxAttachmentStagingInvalidRequest()
	}
	decoded, err := hex.DecodeString(raw)
	if err != nil || len(decoded) != sha256.Size || raw != hex.EncodeToString(decoded) {
		return digest, mailboxAttachmentStagingInvalidRequest()
	}
	copy(digest[:], decoded)
	return digest, nil
}

func mailboxAttachmentStagingHeaders(headers map[string]string) (map[string]string, bool) {
	if !validMailboxObjectStoreHeaders(headers) {
		return nil, false
	}
	return cloneMailboxObjectStoreHeaders(headers), true
}

func mailboxAttachmentStagingViewResponse(view MailboxAttachmentView) mailboxAttachmentHTTPResponse {
	var digest *string
	if view.SHA256 != nil {
		encoded := hex.EncodeToString(view.SHA256[:])
		digest = &encoded
	}

	return mailboxAttachmentHTTPResponse{
		ID:           view.ID.String(),
		Filename:     view.Filename,
		MediaType:    view.MediaType,
		SizeBytes:    view.SizeBytes,
		Disposition:  view.Disposition,
		ContentID:    view.ContentID,
		ScanStatus:   view.ScanStatus,
		SHA256:       digest,
	}
}

func mailboxAttachmentStagingInvalidRequest() error {
	return fiber.NewError(http.StatusBadRequest, "invalid attachment request")
}

func mailboxAttachmentStagingInternalError() error {
	return fiber.NewError(http.StatusInternalServerError, "internal server error")
}

func mailboxAttachmentStagingExpiryValid(expiresAt time.Time) bool {
	return !expiresAt.IsZero() && expiresAt.After(time.Now())
}

func mailboxAttachmentStagingServiceError(ctx fiber.Ctx, err error) error {
	var stagingError *MailboxAttachmentStagingError
	if !errors.As(err, &stagingError) || stagingError == nil {
		return mailboxAttachmentStagingInternalError()
	}

	switch stagingError.Kind() {
	case MailboxAttachmentStagingErrorInvalid:
		return mailboxAttachmentStagingInvalidRequest()
	case MailboxAttachmentStagingErrorNotFound:
		return fiber.NewError(http.StatusNotFound, "attachment not found")
	case MailboxAttachmentStagingErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(http.StatusConflict, "mail account is deleting")
	case MailboxAttachmentStagingErrorUnavailable, MailboxAttachmentStagingErrorIdempotencyUnavailable:
		return fiber.NewError(http.StatusServiceUnavailable, "attachment staging unavailable")
	case MailboxAttachmentStagingErrorLimit:
		return fiber.NewError(http.StatusRequestEntityTooLarge, "attachment limit exceeded")
	case MailboxAttachmentStagingErrorIdempotencyReused,
		MailboxAttachmentStagingErrorIdempotencyExpired,
		MailboxAttachmentStagingErrorStateConflict:
		return fiber.NewError(http.StatusConflict, "attachment state conflict")
	case MailboxAttachmentStagingErrorExpired:
		return fiber.NewError(http.StatusGone, "attachment has expired")
	case MailboxAttachmentStagingErrorConfirmMismatch, MailboxAttachmentStagingErrorNotClean:
		return fiber.NewError(http.StatusUnprocessableEntity, "attachment cannot be confirmed")
	default:
		return mailboxAttachmentStagingInternalError()
	}
}

func mailboxAttachmentStagingAbandonServiceError(ctx fiber.Ctx, err error) error {
	var stagingError *MailboxAttachmentStagingError
	if !errors.As(err, &stagingError) || stagingError == nil {
		return mailboxAttachmentStagingInternalError()
	}

	switch stagingError.Kind() {
	case MailboxAttachmentStagingErrorInvalid:
		return mailboxAttachmentStagingInvalidRequest()
	case MailboxAttachmentStagingErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(http.StatusConflict, "mail account is deleting")
	case MailboxAttachmentStagingErrorUnavailable:
		return fiber.NewError(http.StatusServiceUnavailable, "attachment staging unavailable")
	case MailboxAttachmentStagingErrorStateConflict:
		return fiber.NewError(http.StatusConflict, "attachment state conflict")
	default:
		return mailboxAttachmentStagingInternalError()
	}
}
