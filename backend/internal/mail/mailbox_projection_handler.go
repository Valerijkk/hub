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

type mailboxThreadDetailProjectionReader interface {
	GetMailboxThreadDetail(context.Context, uuid.UUID, MailboxThreadDetailProjectionQuery) (MailboxThreadDetailProjection, error)
}

type mailboxProjectionReader interface {
	mailboxFolderProjectionReader
	mailboxThreadProjectionReader
	mailboxThreadDetailProjectionReader
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

type mailboxThreadDetailProjectionResponse struct {
	ID             uuid.UUID                            `json:"id"`
	Subject        string                               `json:"subject"`
	Participants   []MailboxThreadProjectionParticipant `json:"participants"`
	MessageCount   int32                                `json:"message_count"`
	UnreadCount    int32                                `json:"unread_count"`
	LastMessageAt  time.Time                            `json:"last_message_at"`
	Snippet        string                               `json:"snippet"`
	IsPinned       bool                                 `json:"is_pinned"`
	HasAttachments bool                                 `json:"has_attachments"`
	Labels         []MailboxThreadProjectionLabel       `json:"labels"`
	Messages       []MailboxThreadDetailProjectionMessage `json:"messages"`
	NextCursor     *string                              `json:"next_cursor"`
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
	mail.Get("/threads/:threadId", handler.getThreadDetail)
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

func (handler *MailboxFolderProjectionHandler) getThreadDetail(ctx fiber.Ctx) error {
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	threadID, err := mailboxThreadDetailProjectionThreadIDFromRequest(ctx)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid mailbox thread id")
	}
	query, err := mailboxThreadDetailProjectionQueryFromRequest(ctx, threadID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid mailbox thread query")
	}
	projection, err := handler.service.GetMailboxThreadDetail(ctx.Context(), actorID, query)
	if err != nil {
		return mailboxThreadProjectionHTTPError(ctx, err)
	}
	response, err := mailboxThreadDetailProjectionResponseFromProjection(projection, query)
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

func mailboxThreadDetailProjectionThreadIDFromRequest(ctx fiber.Ctx) (uuid.UUID, error) {
	if ctx == nil {
		return uuid.Nil, errMailboxThreadProjectionRequestInvalid
	}
	value := ctx.Params("threadId")
	threadID, err := uuid.Parse(value)
	if err != nil || threadID == uuid.Nil || value != threadID.String() {
		return uuid.Nil, errMailboxThreadProjectionRequestInvalid
	}
	return threadID, nil
}

func mailboxThreadDetailProjectionQueryFromRequest(ctx fiber.Ctx, threadID uuid.UUID) (MailboxThreadDetailProjectionQuery, error) {
	if ctx == nil || threadID == uuid.Nil {
		return MailboxThreadDetailProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
	}
	values, err := url.ParseQuery(string(ctx.Request().URI().QueryString()))
	if err != nil {
		return MailboxThreadDetailProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
	}
	for name, valuesForName := range values {
		switch name {
		case "cursor", "limit":
			if len(valuesForName) != 1 {
				return MailboxThreadDetailProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
			}
		default:
			return MailboxThreadDetailProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
		}
	}
	query := MailboxThreadDetailProjectionQuery{ThreadID: threadID, Limit: defaultMailboxThreadProjectionLimit}
	if values.Has("limit") {
		limit, err := strconv.Atoi(values.Get("limit"))
		if err != nil || limit < 1 || limit > maxMailboxThreadProjectionLimit {
			return MailboxThreadDetailProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
		}
		query.Limit = limit
	}
	if values.Has("cursor") {
		cursor, err := mailboxThreadDetailProjectionDecodeCursor(values.Get("cursor"))
		if err != nil {
			return MailboxThreadDetailProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
		}
		query.Cursor = cursor
	}
	if !mailboxThreadDetailProjectionQueryValid(query) {
		return MailboxThreadDetailProjectionQuery{}, errMailboxThreadProjectionRequestInvalid
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

func mailboxThreadDetailProjectionResponseFromProjection(projection MailboxThreadDetailProjection, query MailboxThreadDetailProjectionQuery) (mailboxThreadDetailProjectionResponse, error) {
	if !mailboxThreadDetailProjectionQueryValid(query) || projection.Thread.ID != query.ThreadID || len(projection.Messages) > query.Limit || !validMailboxProjectionState(projection.ProjectionState) || !validMailboxThreadProjectionItem(projection.Thread) {
		return mailboxThreadDetailProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if projection.ProjectionState == MailboxProjectionReady && projection.ProjectionAsOf == nil {
		return mailboxThreadDetailProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	if projection.ProjectionAsOf != nil && !mailboxEntValidTime(*projection.ProjectionAsOf) {
		return mailboxThreadDetailProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
	}
	seen := make(map[uuid.UUID]struct{}, len(projection.Messages))
	for index, message := range projection.Messages {
		if !validMailboxThreadDetailProjectionMessage(message, query.ThreadID) || index > 0 && !mailboxThreadDetailProjectionMessageBefore(projection.Messages[index-1], message) {
			return mailboxThreadDetailProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		if _, duplicate := seen[message.ID]; duplicate {
			return mailboxThreadDetailProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		seen[message.ID] = struct{}{}
	}
	if projection.NextCursor != nil {
		if len(projection.Messages) != query.Limit {
			return mailboxThreadDetailProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		cursor, err := mailboxThreadDetailProjectionDecodeCursor(*projection.NextCursor)
		if err != nil || cursor.threadID != query.ThreadID || cursor.filterSignature != mailboxThreadDetailProjectionQuerySignature() {
			return mailboxThreadDetailProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
		last := projection.Messages[len(projection.Messages)-1]
		if !cursor.createdAt.Equal(last.createdAt) || cursor.id != last.ID {
			return mailboxThreadDetailProjectionResponse{}, mailboxServiceError(MailboxServiceErrorInvalid)
		}
	}
	copy := mailboxThreadDetailProjectionCopy(projection)
	return mailboxThreadDetailProjectionResponse{
		ID:              copy.Thread.ID,
		Subject:         copy.Thread.Subject,
		Participants:    copy.Thread.Participants,
		MessageCount:    copy.Thread.MessageCount,
		UnreadCount:     copy.Thread.UnreadCount,
		LastMessageAt:   copy.Thread.LastMessageAt,
		Snippet:         copy.Thread.Snippet,
		IsPinned:        copy.Thread.IsPinned,
		HasAttachments:  copy.Thread.HasAttachments,
		Labels:          copy.Thread.Labels,
		Messages:        copy.Messages,
		NextCursor:      mailboxFolderProjectionStringPointer(copy.NextCursor),
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
