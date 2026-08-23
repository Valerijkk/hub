package mail

import (
	"context"
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type mailboxFolderProjectionReader interface {
	ListMailboxFolders(context.Context, uuid.UUID) (MailboxFolderProjection, error)
}

// MailboxFolderProjectionHandler exposes the authenticated durable folder projection.
type MailboxFolderProjectionHandler struct {
	service mailboxFolderProjectionReader
}

type mailboxFolderProjectionResponse struct {
	Items           []MailboxFolderProjectionItem `json:"items"`
	ProjectionState MailboxProjectionState        `json:"projection_state"`
	ProjectionAsOf  *time.Time                    `json:"projection_as_of"`
}

// NewMailboxFolderProjectionHandler creates the authenticated folder projection handler.
func NewMailboxFolderProjectionHandler(service mailboxFolderProjectionReader) (*MailboxFolderProjectionHandler, error) {
	if mailboxFolderProjectionDependencyNil(service) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return &MailboxFolderProjectionHandler{service: service}, nil
}

// RegisterRoutes registers only GET /api/mail/folders on the supplied /api router.
func (handler *MailboxFolderProjectionHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxFolderProjectionDependencyNil(handler.service) || mailboxFolderProjectionDependencyNil(api) || authMiddleware == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}

	api.Group("/mail", authMiddleware).Get("/folders", handler.listFolders)
	return nil
}

func (handler *MailboxFolderProjectionHandler) listFolders(ctx fiber.Ctx) error {
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}

	projection, err := handler.service.ListMailboxFolders(ctx.Context(), actorID)
	if err != nil {
		return mailboxFolderProjectionHTTPError(ctx, err)
	}

	response, err := mailboxFolderProjectionResponseFromProjection(projection)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox folder projection unavailable")
	}
	return ctx.JSON(response)
}

func mailboxFolderProjectionResponseFromProjection(projection MailboxFolderProjection) (mailboxFolderProjectionResponse, error) {
	if err := validateMailboxFolderProjection(projection); err != nil {
		return mailboxFolderProjectionResponse{}, err
	}

	response := mailboxFolderProjectionResponse{
		Items:           mailboxFolderProjectionCopy(projection).Items,
		ProjectionState: projection.ProjectionState,
		ProjectionAsOf:  mailboxFolderProjectionTimePointer(projection.ProjectionAsOf),
	}
	return response, nil
}

func mailboxFolderProjectionHTTPError(ctx fiber.Ctx, err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox folder projection unavailable")
	}

	switch serviceError.Kind {
	case MailboxServiceErrorNotFound:
		return fiber.NewError(fiber.StatusNotFound, "mailbox folder projection not found")
	case MailboxServiceErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(fiber.StatusConflict, "mailbox account deleting")
	case MailboxServiceErrorUnavailable:
		return fiber.NewError(fiber.StatusServiceUnavailable, "mailbox folder projection unavailable")
	case MailboxServiceErrorInvalid:
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox folder projection unavailable")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox folder projection unavailable")
	}
}
