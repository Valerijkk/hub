package mail

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const mailboxSetupStatusHTTPMaximumQueryBytes = 2048

// MailboxSetupStage is the closed public setup state returned to the mailbox owner.
type MailboxSetupStage string

const (
	MailboxSetupStagePreparing        MailboxSetupStage = "preparing"
	MailboxSetupStagePasswordRequired MailboxSetupStage = "password_required"
	MailboxSetupStageOAuthRequired    MailboxSetupStage = "oauth_required"
	MailboxSetupStageSyncing          MailboxSetupStage = "syncing"
	MailboxSetupStageReady            MailboxSetupStage = "ready"
	MailboxSetupStageActionRequired   MailboxSetupStage = "action_required"
)

// MailboxSetupStatus is the copy-safe public mailbox setup state.
type MailboxSetupStatus struct {
	Stage          MailboxSetupStage
	PrimaryAddress *string
}

// MailboxSetupStatusSource resolves the authenticated actor's safe setup state.
type MailboxSetupStatusSource interface {
	ResolveMailboxSetupStatus(context.Context, uuid.UUID) (MailboxSetupStatus, error)
}

// MailboxSetupStatusHandler exposes the authenticated setup-status boundary.
type MailboxSetupStatusHandler struct {
	source MailboxSetupStatusSource
}

// NewMailboxSetupStatusHandler creates the setup-status HTTP boundary.
func NewMailboxSetupStatusHandler(source MailboxSetupStatusSource) (*MailboxSetupStatusHandler, error) {
	if mailboxHTTPDependencyNil(source) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return &MailboxSetupStatusHandler{source: source}, nil
}

// RegisterRoutes registers the authenticated setup-status route on the supplied /api router.
func (handler *MailboxSetupStatusHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.source) || mailboxHTTPDependencyNil(api) || mailboxHTTPDependencyNil(authMiddleware) {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}
	setup := api.Group("/mail/setup", authMiddleware)
	setup.Get("/status", handler.status)
	return nil
}

func (handler *MailboxSetupStatusHandler) status(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.source) {
		return mailboxSetupStatusHTTPUnavailable(ctx)
	}
	if err := mailboxSetupStatusHTTPRejectQuery(ctx); err != nil {
		return err
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusUnauthorized, "MAIL_SETUP_UNAUTHORIZED", "unauthorized")
	}
	status, err := handler.source.ResolveMailboxSetupStatus(ctx.Context(), actorID)
	if err != nil {
		return mailboxSetupStatusHTTPSourceError(ctx, err)
	}
	response, valid := mailboxSetupStatusHTTPResponseFromStatus(status)
	if !valid {
		return mailboxSetupStatusHTTPUnavailable(ctx)
	}
	return ctx.JSON(response)
}

type mailboxSetupStatusHTTPResponse struct {
	Stage          MailboxSetupStage `json:"stage"`
	PrimaryAddress *string           `json:"primary_address"`
}

func mailboxSetupStatusHTTPResponseFromStatus(status MailboxSetupStatus) (mailboxSetupStatusHTTPResponse, bool) {
	if !validMailboxSetupStatus(status) {
		return mailboxSetupStatusHTTPResponse{}, false
	}

	response := mailboxSetupStatusHTTPResponse{Stage: status.Stage}
	if status.PrimaryAddress != nil {
		primaryAddress := *status.PrimaryAddress
		response.PrimaryAddress = &primaryAddress
	}
	return response, true
}

func validMailboxSetupStatus(status MailboxSetupStatus) bool {
	if status.PrimaryAddress != nil && !validMailboxOwnerPrimaryAddress(*status.PrimaryAddress) {
		return false
	}
	switch status.Stage {
	case MailboxSetupStagePreparing, MailboxSetupStageActionRequired:
		return true
	case MailboxSetupStagePasswordRequired, MailboxSetupStageOAuthRequired, MailboxSetupStageSyncing, MailboxSetupStageReady:
		return status.PrimaryAddress != nil
	default:
		return false
	}
}

func mailboxSetupStatusHTTPRejectQuery(ctx fiber.Ctx) error {
	if ctx == nil {
		return mailboxSetupStatusHTTPUnavailable(ctx)
	}
	query := ctx.Request().URI().QueryString()
	if len(query) > mailboxSetupStatusHTTPMaximumQueryBytes {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusRequestURITooLong, "MAIL_SETUP_QUERY_TOO_LARGE", "invalid mailbox setup query")
	}
	if len(query) != 0 {
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusBadRequest, "MAIL_SETUP_INVALID_QUERY", "invalid mailbox setup query")
	}
	return nil
}

func mailboxSetupStatusHTTPSourceError(ctx fiber.Ctx, err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return mailboxSetupStatusHTTPUnavailable(ctx)
	}
	switch serviceError.Kind {
	case MailboxServiceErrorNotFound, MailboxServiceErrorDeleting, MailboxServiceErrorInvalid:
		return mailboxSetupHTTPFixedError(ctx, fiber.StatusConflict, "MAIL_SETUP_NOT_READY", "mailbox setup not ready")
	case MailboxServiceErrorUnavailable:
		return mailboxSetupStatusHTTPUnavailable(ctx)
	default:
		return mailboxSetupStatusHTTPUnavailable(ctx)
	}
}

func mailboxSetupStatusHTTPUnavailable(ctx fiber.Ctx) error {
	if ctx == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "mailbox setup unavailable")
	}
	return mailboxSetupHTTPFixedError(ctx, fiber.StatusServiceUnavailable, "MAIL_SETUP_UNAVAILABLE", "mailbox setup unavailable")
}
