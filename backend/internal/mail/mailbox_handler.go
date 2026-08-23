package mail

import (
	"context"
	"errors"
	"reflect"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

const mailboxAccountSuspendedCode MailboxOwnerRecoveryCode = "MAILBOX_SUSPENDED"

type mailboxOwnerResolver interface {
	ResolveOwner(context.Context, uuid.UUID) (MailboxOwner, error)
}

// MailboxHandler exposes the authenticated mailbox account status surface.
type MailboxHandler struct {
	service mailboxOwnerResolver
}

type mailboxAccountCapabilities struct {
	RichContent   bool `json:"rich_content"`
	Rules         bool `json:"rules"`
	SpamTraining  bool `json:"spam_training"`
	RemoteImages  bool `json:"remote_images"`
}

type mailboxAccountStatusResponse struct {
	Status            MailboxOwnerState          `json:"status"`
	PrimaryAddress    *string                    `json:"primary_address"`
	RetryAfterSeconds *int                       `json:"retry_after_seconds"`
	ErrorCode         *MailboxOwnerRecoveryCode  `json:"error_code"`
	Capabilities      mailboxAccountCapabilities `json:"capabilities"`
}

// NewMailboxHandler creates the authenticated mailbox account handler.
func NewMailboxHandler(service mailboxOwnerResolver) (*MailboxHandler, error) {
	if mailboxHTTPDependencyNil(service) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return &MailboxHandler{service: service}, nil
}

// RegisterRoutes registers only GET /api/mail/account on the supplied /api router.
func (handler *MailboxHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) || mailboxHTTPDependencyNil(api) || authMiddleware == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}

	api.Group("/mail", authMiddleware).Get("/account", handler.getAccountStatus)
	return nil
}

func (handler *MailboxHandler) getAccountStatus(ctx fiber.Ctx) error {
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}

	owner, err := handler.service.ResolveOwner(ctx.Context(), actorID)
	if err != nil {
		return mailboxAccountServiceHTTPError(err)
	}

	response, err := mailboxAccountStatusFromOwner(owner)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox account unavailable")
	}
	return ctx.JSON(response)
}

func mailboxActorID(ctx fiber.Ctx) (uuid.UUID, error) {
	actorIDValue, ok := ctx.Locals("userID").(string)
	if !ok || actorIDValue == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}

	actorID, err := uuid.Parse(actorIDValue)
	if err != nil || actorID == uuid.Nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
	}
	return actorID, nil
}

func mailboxAccountStatusFromOwner(owner MailboxOwner) (mailboxAccountStatusResponse, error) {
	if owner.OrganizationID() == uuid.Nil {
		return mailboxAccountStatusResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	response := mailboxAccountStatusResponse{Status: owner.State()}
	address, hasAddress := owner.PrimaryAddress()
	if hasAddress {
		if !validMailboxOwnerPrimaryAddress(address) {
			return mailboxAccountStatusResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		addressCopy := address
		response.PrimaryAddress = &addressCopy
	}

	recoveryCode, hasRecoveryCode := owner.RecoveryCode()
	if hasRecoveryCode && !validMailboxOwnerRecoveryCode(recoveryCode) {
		return mailboxAccountStatusResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	accountID, hasAccount := owner.AccountID()
	projectionState, hasProjectionState := owner.ProjectionState()
	switch owner.State() {
	case MailboxOwnerNotConfigured:
		if hasAccount || accountID != uuid.Nil || hasAddress || hasRecoveryCode || hasProjectionState {
			return mailboxAccountStatusResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	case MailboxOwnerProvisioning:
		if !hasAccount || accountID == uuid.Nil || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return mailboxAccountStatusResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if hasRecoveryCode {
			recoveryCodeCopy := recoveryCode
			response.ErrorCode = &recoveryCodeCopy
		}
	case MailboxOwnerActive:
		if !hasAccount || accountID == uuid.Nil || !hasAddress || hasRecoveryCode || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return mailboxAccountStatusResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if projectionState == MailboxProjectionReady {
			response.Capabilities = mailboxAccountCapabilities{
				RichContent:  true,
				Rules:        true,
				SpamTraining: true,
				RemoteImages: true,
			}
		}
	case MailboxOwnerSuspended:
		if !hasAccount || accountID == uuid.Nil || hasRecoveryCode || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return mailboxAccountStatusResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		code := mailboxAccountSuspendedCode
		response.ErrorCode = &code
	case MailboxOwnerError:
		if !hasAccount || accountID == uuid.Nil || !hasRecoveryCode || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return mailboxAccountStatusResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		recoveryCodeCopy := recoveryCode
		response.ErrorCode = &recoveryCodeCopy
	case MailboxOwnerDeleting:
		if !hasAccount || accountID == uuid.Nil || hasRecoveryCode || !hasProjectionState || !validMailboxProjectionState(projectionState) {
			return mailboxAccountStatusResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	default:
		return mailboxAccountStatusResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}

	return response, nil
}

func mailboxAccountServiceHTTPError(err error) error {
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox account unavailable")
	}

	switch serviceError.Kind {
	case MailboxServiceErrorNotFound:
		return fiber.NewError(fiber.StatusNotFound, "mailbox account not found")
	case MailboxServiceErrorUnavailable:
		return fiber.NewError(fiber.StatusServiceUnavailable, "mailbox service unavailable")
	case MailboxServiceErrorInvalid, MailboxServiceErrorDeleting:
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox account unavailable")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox account unavailable")
	}
}

func mailboxHTTPDependencyNil(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
