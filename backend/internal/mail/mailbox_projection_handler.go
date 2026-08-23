package mail

import (
	"context"
	"errors"
	"net/url"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type mailboxFolderProjectionReader interface {
	ListMailboxFolders(context.Context, uuid.UUID) (MailboxFolderProjection, error)
}

type mailboxThreadProjectionReader interface {
	ListMailboxThreads(context.Context, uuid.UUID, MailboxThreadProjectionQuery) (MailboxThreadProjection, error)
}

type mailboxProjectionReader interface {
	mailboxFolderProjectionReader
	mailboxThreadProjectionReader
}

// MailboxFolderProjectionHandler exposes the authenticated durable folder projection.
type MailboxFolderProjectionHandler struct {
	service mailboxProjectionReader
}

type mailboxFolderProjectionResponse struct {
	Items           []MailboxFolderProjectionItem `json:"items"`
	ProjectionState MailboxProjectionState        `json:"projection_state"`
	ProjectionAsOf  *time.Time                    `json:"projection_as_of"`
}

type mailboxThreadProjectionResponse struct {
	Items           []MailboxThreadProjectionItem `json:"items"`
	NextCursor      *string                       `json:"next_cursor"`
	ProjectionState MailboxProjectionState        `json:"projection_state"`
	ProjectionAsOf  *time.Time                    `json:"projection_as_of"`
}

// NewMailboxFolderProjectionHandler creates the authenticated folder projection handler.
func NewMailboxFolderProjectionHandler(service mailboxProjectionReader) (*MailboxFolderProjectionHandler, error) {
	if mailboxFolderProjectionDependencyNil(service) {
		return nil, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	return &MailboxFolderProjectionHandler{service: service}, nil
}

// RegisterRoutes registers the authenticated folder and thread projection reads on the supplied /api router.
func (handler *MailboxFolderProjectionHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxFolderProjectionDependencyNil(handler.service) || mailboxFolderProjectionDependencyNil(api) || authMiddleware == nil {
		return mailboxServiceError(MailboxServiceErrorInvalid)
	}

	mail := api.Group("/mail", authMiddleware)
	mail.Get("/folders", handler.listFolders)
	mail.Get("/threads", handler.listThreads)
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

func (handler *MailboxFolderProjectionHandler) listThreads(ctx fiber.Ctx) error {
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	query, err := mailboxThreadProjectionQueryFromRequest(ctx)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid mailbox thread query")
	}
	projection, err := handler.service.ListMailboxThreads(ctx.Context(), actorID, query)
	if err != nil {
		return mailboxThreadProjectionHTTPError(ctx, err)
	}
	response, err := mailboxThreadProjectionResponseFromProjection(projection, query)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox thread projection unavailable")
	}
	return ctx.JSON(response)
}

func mailboxThreadProjectionQueryFromRequest(ctx fiber.Ctx) (MailboxThreadProjectionQuery, error) {
	if ctx == nil {
		return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
	}
	values, err := url.ParseQuery(string(ctx.Request().URI().QueryString()))
	if err != nil {
		return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
	}
	for name, valuesForName := range values {
		switch name {
		case "cursor", "limit", "folder_id", "label_id", "unread", "has_attachments", "starred", "search":
			if len(valuesForName) != 1 {
				return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
			}
		default:
			return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
		}
	}
	query := MailboxThreadProjectionQuery{Limit: defaultMailboxThreadProjectionLimit}
	if values.Has("limit") {
		limit, err := strconv.Atoi(values.Get("limit"))
		if err != nil || limit < 1 || limit > maxMailboxThreadProjectionLimit {
			return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
		}
		query.Limit = limit
	}
	if values.Has("cursor") {
		cursor, err := mailboxThreadProjectionDecodeCursor(values.Get("cursor"))
		if err != nil {
			return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
		}
		query.Cursor = cursor
	}
	for _, value := range []struct {
		name   string
		target **uuid.UUID
	}{
		{name: "folder_id", target: &query.FolderID},
		{name: "label_id", target: &query.LabelID},
	} {
		if !values.Has(value.name) {
			continue
		}
		id, err := uuid.Parse(values.Get(value.name))
		if err != nil || id == uuid.Nil {
			return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
		}
		*value.target = &id
	}
	var errBool error
	query.Unread, errBool = mailboxThreadProjectionOptionalBool(values, "unread")
	if errBool != nil {
		return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
	}
	query.HasAttachments, errBool = mailboxThreadProjectionOptionalBool(values, "has_attachments")
	if errBool != nil {
		return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
	}
	query.Starred, errBool = mailboxThreadProjectionOptionalBool(values, "starred")
	if errBool != nil {
		return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
	}
	if values.Has("search") {
		search, err := mailboxThreadProjectionNormalizeSearch(values.Get("search"))
		if err != nil {
			return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
		}
		query.Search = &search
	}
	if !mailboxThreadProjectionQueryValid(query) {
		return MailboxThreadProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
	}
	return query, nil
}

func mailboxThreadProjectionOptionalBool(values url.Values, name string) (*bool, error) {
	if !values.Has(name) {
		return nil, nil
	}
	switch values.Get(name) {
	case "true":
		value := true
		return &value, nil
	case "false":
		value := false
		return &value, nil
	default:
		return nil, errMailboxThreadProjectionRequestInvalid
	}
}

func mailboxThreadProjectionResponseFromProjection(projection MailboxThreadProjection, query MailboxThreadProjectionQuery) (mailboxThreadProjectionResponse, error) {
	if !mailboxThreadProjectionQueryValid(query) || len(projection.Items) > query.Limit || !validMailboxProjectionState(projection.ProjectionState) {
		return mailboxThreadProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if projection.ProjectionState == MailboxProjectionReady && projection.ProjectionAsOf == nil {
		return mailboxThreadProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if projection.ProjectionAsOf != nil && !mailboxEntValidTime(*projection.ProjectionAsOf) {
		return mailboxThreadProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	seen := make(map[uuid.UUID]struct{}, len(projection.Items))
	for index, item := range projection.Items {
		if !validMailboxThreadProjectionItem(item) || index > 0 && !mailboxThreadProjectionItemBefore(projection.Items[index-1], item) {
			return mailboxThreadProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := seen[item.ID]; duplicate {
			return mailboxThreadProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seen[item.ID] = struct{}{}
	}
	if projection.NextCursor != nil {
		if len(projection.Items) != query.Limit {
			return mailboxThreadProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		cursor, err := mailboxThreadProjectionDecodeCursor(*projection.NextCursor)
		if err != nil {
			return mailboxThreadProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		last := projection.Items[len(projection.Items)-1]
		if cursor.isPinned != last.IsPinned || !cursor.lastMessageAt.Equal(last.LastMessageAt) || cursor.id != last.ID || cursor.filterSignature != mailboxThreadProjectionQuerySignature(query) {
			return mailboxThreadProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}
	copy := mailboxThreadProjectionCopy(projection)
	return mailboxThreadProjectionResponse{
		Items:           copy.Items,
		NextCursor:      mailboxFolderProjectionStringPointer(copy.NextCursor),
		ProjectionState: copy.ProjectionState,
		ProjectionAsOf:  mailboxFolderProjectionTimePointer(copy.ProjectionAsOf),
	}, nil
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

func mailboxThreadProjectionHTTPError(ctx fiber.Ctx, err error) error {
	if errors.Is(err, errMailboxThreadProjectionRequestInvalid) {
		return fiber.NewError(fiber.StatusBadRequest, "invalid mailbox thread query")
	}
	var serviceError *MailboxServiceError
	if !errors.As(err, &serviceError) || serviceError == nil {
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox thread projection unavailable")
	}
	switch serviceError.Kind {
	case MailboxServiceErrorNotFound:
		return fiber.NewError(fiber.StatusNotFound, "mailbox thread projection not found")
	case MailboxServiceErrorDeleting:
		ctx.Set("X-Hub-Error-Code", "MAIL_ACCOUNT_DELETING")
		return fiber.NewError(fiber.StatusConflict, "mailbox account deleting")
	case MailboxServiceErrorUnavailable:
		return fiber.NewError(fiber.StatusServiceUnavailable, "mailbox thread projection unavailable")
	case MailboxServiceErrorInvalid:
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox thread projection unavailable")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "mailbox thread projection unavailable")
	}
}
