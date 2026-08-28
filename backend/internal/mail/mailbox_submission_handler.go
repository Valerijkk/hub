package mail

import (
	"bytes"
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type mailboxSubmissionCreator interface {
	CreateMailboxSubmission(context.Context, uuid.UUID, uuid.UUID, MailboxSubmissionCreateInput) (MailboxSubmissionView, error)
}

type mailboxSubmissionReader interface {
	GetMailboxSubmission(context.Context, uuid.UUID, uuid.UUID) (MailboxSubmissionView, error)
}

// MailboxSubmissionHandler exposes the authenticated durable send-intent transport.
type MailboxSubmissionHandler struct {
	service mailboxSubmissionCreator
	reader  mailboxSubmissionReader
}

// NewMailboxSubmissionHandler creates the authenticated submission handler.
func NewMailboxSubmissionHandler(service mailboxSubmissionCreator) (*MailboxSubmissionHandler, error) {
	if mailboxHTTPDependencyNil(service) {
		return nil, mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
	handler := &MailboxSubmissionHandler{service: service}
	if reader, ok := service.(mailboxSubmissionReader); ok && !mailboxHTTPDependencyNil(reader) {
		handler.reader = reader
	}
	return handler, nil
}

// RegisterRoutes registers authenticated submission creation and receipt reads on the supplied /api router.
func (handler *MailboxSubmissionHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) || mailboxHTTPDependencyNil(api) || mailboxHTTPDependencyNil(authMiddleware) {
		return mailboxSubmissionServiceError(MailboxSubmissionServiceErrorInvalid)
	}
	mail := api.Group("/mail", authMiddleware)
	mail.Post("/drafts/:draftId/send", handler.send)
	mail.Get("/submissions/:submissionId", handler.get)
	return nil
}

func (handler *MailboxSubmissionHandler) send(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) {
		return mailboxSubmissionHTTPUnavailable()
	}
	if err := mailboxSubmissionHTTPRejectTransport(ctx); err != nil {
		return err
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	draftID, err := mailboxDraftCreateHTTPCanonicalUUID(ctx.Params("draftId"))
	if err != nil {
		return err
	}
	rawIdempotencyKey, oneIdempotencyKey := mailboxSubmissionHTTPExactlyOneHeaderValue(ctx, "Idempotency-Key")
	rawIfMatch, oneIfMatch := mailboxSubmissionHTTPExactlyOneHeaderValue(ctx, "If-Match")
	if !oneIdempotencyKey || !oneIfMatch {
		return mailboxSubmissionHTTPInvalidRequest()
	}
	idempotencyKey, err := mailboxSubmissionNormalizeIdempotencyKey(string(rawIdempotencyKey))
	if err != nil {
		return mailboxSubmissionHTTPInvalidRequest()
	}
	expectedRevision, err := mailboxDraftSaveHTTPExpectedRevision(string(rawIfMatch))
	if err != nil {
		return err
	}
	view, err := handler.service.CreateMailboxSubmission(ctx.Context(), actorID, draftID, MailboxSubmissionCreateInput{
		ExpectedRevision: expectedRevision,
		IdempotencyKey:   idempotencyKey,
	})
	if err != nil {
		return mailboxSubmissionHTTPServiceError(ctx, err)
	}
	response, ok := mailboxSubmissionHTTPResponseFromView(view)
	if !ok {
		return mailboxSubmissionHTTPUnavailable()
	}
	return ctx.Status(http.StatusAccepted).JSON(response)
}

func (handler *MailboxSubmissionHandler) get(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.reader) {
		return mailboxSubmissionHTTPUnavailable()
	}
	if err := mailboxSubmissionHTTPRejectTransport(ctx); err != nil {
		return err
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	submissionID, err := mailboxSubmissionHTTPCanonicalUUID(ctx.Params("submissionId"))
	if err != nil {
		return err
	}
	view, err := handler.reader.GetMailboxSubmission(ctx.Context(), actorID, submissionID)
	if err != nil {
		return mailboxSubmissionHTTPReadServiceError(ctx, err)
	}
	response, ok := mailboxSubmissionHTTPResponseFromView(view)
	if !ok {
		return mailboxSubmissionHTTPUnavailable()
	}
	ctx.Set("Cache-Control", "no-store")
	return ctx.Status(http.StatusOK).JSON(response)
}

func mailboxSubmissionHTTPRejectTransport(ctx fiber.Ctx) error {
	if ctx == nil || len(ctx.Request().URI().QueryString()) != 0 || len(ctx.BodyRaw()) != 0 || ctx.Request().Header.Peek(fiber.HeaderContentType) != nil {
		return mailboxSubmissionHTTPInvalidRequest()
	}
	return nil
}

func mailboxSubmissionHTTPExactlyOneHeaderValue(ctx fiber.Ctx, name string) ([]byte, bool) {
	if ctx == nil {
		return nil, false
	}
	var (
		value []byte
		count int
	)
	ctx.Request().Header.VisitAllInOrder(func(key, headerValue []byte) {
		if !bytes.EqualFold(key, []byte(name)) {
			return
		}
		count++
		if count == 1 {
			value = append([]byte(nil), headerValue...)
		}
	})
	return value, count == 1
}

func mailboxSubmissionHTTPCanonicalUUID(raw string) (uuid.UUID, error) {
	value, err := uuid.Parse(raw)
	if err != nil || value == uuid.Nil || raw != value.String() {
		return uuid.Nil, mailboxSubmissionHTTPInvalidRequest()
	}
	return value, nil
}

type mailboxSubmissionHTTPResponse struct {
	ID        string    `json:"id"`
	Status    string    `json:"status"`
	MessageID *string   `json:"message_id"`
	ErrorCode *string   `json:"error_code"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func mailboxSubmissionHTTPResponseFromView(view MailboxSubmissionView) (mailboxSubmissionHTTPResponse, bool) {
	if view.ID == uuid.Nil || !mailboxEntValidTime(view.CreatedAt) || !mailboxEntValidTime(view.UpdatedAt) || view.UpdatedAt.Before(view.CreatedAt) {
		return mailboxSubmissionHTTPResponse{}, false
	}
	messageID, ok := mailboxSubmissionHTTPMessageID(view.MessageID)
	if !ok {
		return mailboxSubmissionHTTPResponse{}, false
	}
	errorCode, ok := mailboxSubmissionHTTPErrorCode(view.ErrorCode)
	if !ok || !mailboxSubmissionHTTPPublicViewValid(view.Status, view.MessageID, view.ErrorCode) {
		return mailboxSubmissionHTTPResponse{}, false
	}
	return mailboxSubmissionHTTPResponse{
		ID:        view.ID.String(),
		Status:    string(view.Status),
		MessageID: messageID,
		ErrorCode: errorCode,
		CreatedAt: view.CreatedAt.UTC(),
		UpdatedAt: view.UpdatedAt.UTC(),
	}, true
}

func mailboxSubmissionHTTPMessageID(value *uuid.UUID) (*string, bool) {
	if value == nil {
		return nil, true
	}
	if *value == uuid.Nil {
		return nil, false
	}
	result := value.String()
	return &result, true
}

func mailboxSubmissionHTTPErrorCode(value *MailboxSubmissionErrorCode) (*string, bool) {
	if value == nil {
		return nil, true
	}
	if !validMailboxSubmissionErrorCode(*value) {
		return nil, false
	}
	result := string(*value)
	return &result, true
}

func mailboxSubmissionHTTPPublicViewValid(status MailboxSubmissionStatus, messageID *uuid.UUID, errorCode *MailboxSubmissionErrorCode) bool {
	switch status {
	case MailboxSubmissionStatusQueued, MailboxSubmissionStatusAuthorizing, MailboxSubmissionStatusSubmitting, MailboxSubmissionStatusAccepted:
		return messageID == nil && errorCode == nil
	case MailboxSubmissionStatusSent:
		return messageID != nil && *messageID != uuid.Nil && errorCode == nil
	case MailboxSubmissionStatusUnknown:
		if messageID != nil || errorCode == nil {
			return false
		}
		switch *errorCode {
		case MailboxSubmissionErrorCodeOutcomeUnknown, MailboxSubmissionErrorCodeReconciliationRequired, MailboxSubmissionErrorCodeMailboxUnavailable:
			return true
		case MailboxSubmissionErrorCodeReconciliationExhausted:
			return true
		default:
			return false
		}
	case MailboxSubmissionStatusFailed:
		if messageID != nil || errorCode == nil {
			return false
		}
		switch *errorCode {
		case MailboxSubmissionErrorCodeRejected,
			MailboxSubmissionErrorCodePolicy,
			MailboxSubmissionErrorCodeAttachmentNotClean,
			MailboxSubmissionErrorCodeIdentityUnavailable,
			MailboxSubmissionErrorCodeRecipientLimit,
			MailboxSubmissionErrorCodeSizeLimit,
			MailboxSubmissionErrorCodeRetryExhausted:
			return true
		default:
			return false
		}
	default:
		return false
	}
}

func mailboxSubmissionHTTPServiceError(ctx fiber.Ctx, err error) error {
	var serviceError *MailboxSubmissionServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return mailboxSubmissionHTTPUnavailable()
	}
	switch serviceError.Kind() {
	case MailboxSubmissionServiceErrorInvalid:
		return mailboxSubmissionHTTPInvalidRequest()
	case MailboxSubmissionServiceErrorNotFound:
		return fiber.NewError(http.StatusNotFound, "mail draft unavailable")
	case MailboxSubmissionServiceErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(http.StatusConflict, "mail account is deleting")
	case MailboxSubmissionServiceErrorVersionConflict:
		ctx.Set("X-Hub-Error-Code", "MAIL_DRAFT_VERSION_CONFLICT")
		return fiber.NewError(http.StatusConflict, "mail draft version conflict")
	case MailboxSubmissionServiceErrorSubmissionInProgress:
		ctx.Set("X-Hub-Error-Code", "MAIL_DRAFT_SUBMISSION_IN_PROGRESS")
		return fiber.NewError(http.StatusConflict, "mail draft submission in progress")
	case MailboxSubmissionServiceErrorDraftNotSendable:
		ctx.Set("X-Hub-Error-Code", "MAIL_DRAFT_NOT_SENDABLE")
		return fiber.NewError(http.StatusConflict, "mail draft is not sendable")
	case MailboxSubmissionServiceErrorIdempotencyConflict:
		ctx.Set("X-Hub-Error-Code", "IDEMPOTENCY_KEY_REUSED")
		return fiber.NewError(http.StatusConflict, "mail submission idempotency conflict")
	default:
		return mailboxSubmissionHTTPUnavailable()
	}
}

func mailboxSubmissionHTTPReadServiceError(ctx fiber.Ctx, err error) error {
	var serviceError *MailboxSubmissionServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return mailboxSubmissionHTTPUnavailable()
	}
	switch serviceError.Kind() {
	case MailboxSubmissionServiceErrorInvalid:
		return mailboxSubmissionHTTPInvalidRequest()
	case MailboxSubmissionServiceErrorNotFound:
		return fiber.NewError(http.StatusNotFound, "mail submission unavailable")
	case MailboxSubmissionServiceErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(http.StatusConflict, "mail account is deleting")
	default:
		return mailboxSubmissionHTTPUnavailable()
	}
}

func mailboxSubmissionHTTPInvalidRequest() error {
	return fiber.NewError(http.StatusBadRequest, "invalid mail submission request")
}

func mailboxSubmissionHTTPUnavailable() error {
	return fiber.NewError(http.StatusServiceUnavailable, "mail submission unavailable")
}
