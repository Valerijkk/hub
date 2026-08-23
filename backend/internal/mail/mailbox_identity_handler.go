package mail

import (
	"encoding/json"
	"errors"
	"net/url"
	"strconv"

	"github.com/gofiber/fiber/v3"
)

// MailboxIdentityHandler exposes the authenticated identity-list operation.
// Route registration remains deliberately separate from this narrow packet.
type MailboxIdentityHandler struct {
	service mailboxIdentityReader
}

// NewMailboxIdentityHandler creates the identity-list handler without wiring routes.
func NewMailboxIdentityHandler(service mailboxIdentityReader) (*MailboxIdentityHandler, error) {
	if mailboxHTTPDependencyNil(service) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return &MailboxIdentityHandler{service: service}, nil
}

// List returns the bounded From-identity page for the authenticated mailbox owner.
func (handler *MailboxIdentityHandler) List(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) || ctx == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox identity unavailable")
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	query, err := mailboxIdentityListQueryFromRequest(ctx)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid mailbox identity query")
	}
	page, err := handler.service.ListMailboxIdentities(ctx.Context(), actorID, query)
	if err != nil {
		return mailboxIdentityHTTPError(ctx, err)
	}
	encoded, err := json.Marshal(page)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox identity unavailable")
	}
	if int64(len(encoded)) > maxMailboxIdentityResponseBytes {
		ctx.Set("X-Hub-Error-Code", "MAIL_IDENTITY_RESPONSE_TOO_LARGE")
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, "mailbox identity response exceeds safe limit")
	}
	ctx.Set(fiber.HeaderContentType, "application/json; charset=utf-8")
	return ctx.Send(encoded)
}

func mailboxIdentityListQueryFromRequest(ctx fiber.Ctx) (MailboxIdentityListQuery, error) {
	if ctx == nil {
		return MailboxIdentityListQuery{}, errMailboxIdentityRequestInvalid
	}
	values, err := url.ParseQuery(string(ctx.Request().URI().QueryString()))
	if err != nil {
		return MailboxIdentityListQuery{}, errMailboxIdentityRequestInvalid
	}
	for name, valuesForName := range values {
		switch name {
		case "cursor", "limit":
			if len(valuesForName) != 1 {
				return MailboxIdentityListQuery{}, errMailboxIdentityRequestInvalid
			}
		default:
			return MailboxIdentityListQuery{}, errMailboxIdentityRequestInvalid
		}
	}
	query := MailboxIdentityListQuery{Limit: defaultMailboxIdentityLimit}
	if values.Has("limit") {
		limit, err := strconv.Atoi(values.Get("limit"))
		if err != nil || limit < 1 || limit > maxMailboxIdentityLimit {
			return MailboxIdentityListQuery{}, errMailboxIdentityRequestInvalid
		}
		query.Limit = limit
	}
	if values.Has("cursor") {
		cursor, err := mailboxIdentityDecodeCursor(values.Get("cursor"))
		if err != nil {
			return MailboxIdentityListQuery{}, errMailboxIdentityRequestInvalid
		}
		query.Cursor = cursor
	}
	if !mailboxIdentityListQueryValid(query) {
		return MailboxIdentityListQuery{}, errMailboxIdentityRequestInvalid
	}
	return query, nil
}

func mailboxIdentityHTTPError(ctx fiber.Ctx, err error) error {
	if errors.Is(err, errMailboxIdentityRequestInvalid) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid mailbox identity query")
	}
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox identity unavailable")
	}
	switch serviceError.Kind {
	case MailboxServiceErrorNotFound:
		return fiber.NewError(fiber.StatusNotFound, "mailbox identity not found")
	case MailboxServiceErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(fiber.StatusConflict, "mailbox account deleting")
	case MailboxServiceErrorUnavailable:
		return fiber.NewError(fiber.StatusServiceUnavailable, "mailbox identity unavailable")
	case MailboxServiceErrorInvalid:
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox identity unavailable")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox identity unavailable")
	}
}

var _ mailboxIdentityReader = (*MailboxIdentityService)(nil)
