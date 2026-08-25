package mail

import (
	"context"
	"encoding/json"
	"mime"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type mailboxMessageReader interface {
	GetMailboxMessage(context.Context, uuid.UUID, uuid.UUID) (MailboxMessageDetail, error)
	DownloadMailboxMessageAttachment(context.Context, uuid.UUID, uuid.UUID, uuid.UUID) (MailboxMessageAttachmentDownload, error)
}

// MailboxMessageHandler exposes only the authenticated safe message-detail read.
type MailboxMessageHandler struct {
	service mailboxMessageReader
}

// NewMailboxMessageHandler creates the isolated public message-detail handler.
func NewMailboxMessageHandler(service mailboxMessageReader) (*MailboxMessageHandler, error) {
	if mailboxFolderProjectionDependencyNil(service) {
		return nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	return &MailboxMessageHandler{service: service}, nil
}

// RegisterRoutes registers the authenticated message and attachment download readers on the supplied /api router.
func (handler *MailboxMessageHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxFolderProjectionDependencyNil(handler.service) || mailboxFolderProjectionDependencyNil(api) || authMiddleware == nil {
		return mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	router := api.Group("/mail", authMiddleware)
	router.Get("/messages/:messageId", handler.getMessage)
	router.Get("/messages/:messageId/attachments/:attachmentId/download", handler.downloadAttachment)
	return nil
}

func (handler *MailboxMessageHandler) getMessage(ctx fiber.Ctx) error {
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	if len(ctx.Request().URI().QueryString()) != 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid mail message query")
	}
	messageID, err := mailboxMessageIDFromRequest(ctx)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid mail message id")
	}
	detail, err := handler.service.GetMailboxMessage(ctx.Context(), actorID, messageID)
	if err != nil {
		return mailboxMessageHTTPError(ctx, err)
	}
	if !mailboxMessageDetailValid(detail) {
		return fiber.NewError(fiber.StatusInternalServerError, "mail message unavailable")
	}
	encoded, err := json.Marshal(detail)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "mail message unavailable")
	}
	if int64(len(encoded)) > maxMailboxMessageDetailResponseBytes {
		ctx.Set("X-Hub-Error-Code", "MAIL_MESSAGE_RESPONSE_TOO_LARGE")
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "mail message exceeds safe response limit")
	}
	ctx.Set(fiber.HeaderContentType, "application/json; charset=utf-8")
	return ctx.Send(encoded)
}

func mailboxMessageIDFromRequest(ctx fiber.Ctx) (uuid.UUID, error) {
	if ctx == nil {
		return uuid.Nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	value := ctx.Params("messageId")
	messageID, err := uuid.Parse(value)
	if err != nil || messageID == uuid.Nil || value != messageID.String() {
		return uuid.Nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	return messageID, nil
}

func mailboxMessageAttachmentIDFromRequest(ctx fiber.Ctx) (uuid.UUID, error) {
	if ctx == nil {
		return uuid.Nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	value := ctx.Params("attachmentId")
	attachmentID, err := uuid.Parse(value)
	if err != nil || attachmentID == uuid.Nil || value != attachmentID.String() {
		return uuid.Nil, mailboxMessageReadError(MailboxMessageReadErrorInvalid)
	}
	return attachmentID, nil
}

func mailboxMessageHTTPError(ctx fiber.Ctx, err error) error {
	kind, ok := mailboxMessageReadErrorKindOf(err)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "mail message unavailable")
	}
	switch kind {
	case MailboxMessageReadErrorNotFound:
		return fiber.NewError(fiber.StatusNotFound, "mail message not found")
	case MailboxMessageReadErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(fiber.StatusConflict, "mailbox account deleting")
	case MailboxMessageReadErrorPayloadTooLarge:
		ctx.Set("X-Hub-Error-Code", "MAIL_MESSAGE_CONTENT_TOO_LARGE")
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "mail message exceeds safe content limit")
	case MailboxMessageReadErrorUnprocessable:
		ctx.Set("X-Hub-Error-Code", "MAIL_MESSAGE_CONTENT_UNPROCESSABLE")
		return fiber.NewError(fiber.StatusUnprocessableEntity, "mail message content unavailable")
	case MailboxMessageReadErrorUnavailable:
		return fiber.NewError(fiber.StatusServiceUnavailable, "mail message unavailable")
	case MailboxMessageReadErrorInvalid:
		return fiber.NewError(fiber.StatusInternalServerError, "mail message unavailable")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "mail message unavailable")
	}
}

func (handler *MailboxMessageHandler) downloadAttachment(ctx fiber.Ctx) error {
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	if len(ctx.Request().URI().QueryString()) != 0 {
		return fiber.NewError(fiber.StatusBadRequest, "invalid mail attachment query")
	}
	messageID, err := mailboxMessageIDFromRequest(ctx)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid mail message id")
	}
	attachmentID, err := mailboxMessageAttachmentIDFromRequest(ctx)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid mail attachment id")
	}
	attachment, err := handler.service.DownloadMailboxMessageAttachment(ctx.Context(), actorID, messageID, attachmentID)
	if err != nil {
		return mailboxMessageAttachmentHTTPError(ctx, err)
	}
	if !mailboxMessageAttachmentDownloadValid(attachment) {
		return fiber.NewError(fiber.StatusInternalServerError, "mail attachment unavailable")
	}
	contentDisposition := mime.FormatMediaType("attachment", map[string]string{"filename": attachment.Filename})
	if contentDisposition == "" {
		return fiber.NewError(fiber.StatusInternalServerError, "mail attachment unavailable")
	}
	ctx.Set("Content-Disposition", contentDisposition)
	ctx.Set(fiber.HeaderContentType, "application/octet-stream")
	ctx.Set("X-Content-Type-Options", "nosniff")
	ctx.Set("Cache-Control", "no-store")
	ctx.Set("Referrer-Policy", "no-referrer")
	return ctx.Send(attachment.Bytes)
}

func mailboxMessageAttachmentHTTPError(ctx fiber.Ctx, err error) error {
	kind, ok := mailboxMessageReadErrorKindOf(err)
	if !ok {
		return fiber.NewError(fiber.StatusInternalServerError, "mail attachment unavailable")
	}
	switch kind {
	case MailboxMessageReadErrorNotFound:
		return fiber.NewError(fiber.StatusNotFound, "mail attachment not found")
	case MailboxMessageReadErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(fiber.StatusConflict, "mailbox account deleting")
	case MailboxMessageReadErrorPayloadTooLarge:
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "mail attachment exceeds safe limit")
	case MailboxMessageReadErrorUnprocessable:
		return fiber.NewError(fiber.StatusUnprocessableEntity, "mail attachment unavailable")
	case MailboxMessageReadErrorUnavailable:
		return fiber.NewError(fiber.StatusServiceUnavailable, "mail attachment unavailable")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "mail attachment unavailable")
	}
}
