package mail

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type mailboxReplyDraftCreator interface {
	CreateMailboxReplyDraft(context.Context, uuid.UUID, MailboxReplyDraftInput) (MailboxDraftView, error)
}

// MailboxReplyDraftHandler exposes the authenticated server-authoritative reply-draft bootstrap transport.
type MailboxReplyDraftHandler struct {
	service mailboxReplyDraftCreator
}

// NewMailboxReplyDraftHandler creates the authenticated reply-draft handler.
func NewMailboxReplyDraftHandler(service mailboxReplyDraftCreator) (*MailboxReplyDraftHandler, error) {
	if mailboxHTTPDependencyNil(service) {
		return nil, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return &MailboxReplyDraftHandler{service: service}, nil
}

// RegisterRoutes registers only POST /api/mail/messages/:messageId/reply-draft on the supplied /api router.
func (handler *MailboxReplyDraftHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) || mailboxHTTPDependencyNil(api) || mailboxHTTPDependencyNil(authMiddleware) {
		return mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	api.Group("/mail", authMiddleware).Post("/messages/:messageId/reply-draft", handler.create)
	return nil
}

type mailboxReplyDraftHTTPRequest struct {
	DraftClientID  string `json:"draft_client_id"`
	FromIdentityID string `json:"from_identity_id"`
}

func (handler *MailboxReplyDraftHandler) create(ctx fiber.Ctx) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) {
		return mailboxDraftCreateHTTPUnavailable()
	}
	if err := mailboxDraftCreateHTTPRejectQuery(ctx); err != nil {
		return err
	}
	actorID, err := mailboxActorID(ctx)
	if err != nil {
		return err
	}
	sourceMessageID, err := mailboxDraftCreateHTTPCanonicalUUID(ctx.Params("messageId"))
	if err != nil {
		return err
	}
	var request mailboxReplyDraftHTTPRequest
	if err := mailboxDraftCreateHTTPDecodeJSON(ctx, &request); err != nil {
		return err
	}
	input, err := mailboxReplyDraftHTTPInput(sourceMessageID, request, ctx.Get("Idempotency-Key"))
	if err != nil {
		return err
	}
	view, err := handler.service.CreateMailboxReplyDraft(ctx.Context(), actorID, input)
	if err != nil {
		return mailboxDraftCreateHTTPServiceError(ctx, err)
	}
	if !mailboxReplyDraftHTTPViewMatches(view, sourceMessageID, input) {
		return mailboxDraftCreateHTTPUnavailable()
	}
	response, ok := mailboxDraftCreateHTTPResponseFromView(view)
	if !ok {
		return mailboxDraftCreateHTTPUnavailable()
	}
	return ctx.Status(http.StatusCreated).JSON(response)
}

func mailboxReplyDraftHTTPInput(sourceMessageID uuid.UUID, request mailboxReplyDraftHTTPRequest, idempotencyKey string) (MailboxReplyDraftInput, error) {
	if sourceMessageID == uuid.Nil {
		return MailboxReplyDraftInput{}, mailboxDraftCreateHTTPInvalidRequest()
	}
	draftClientID, err := mailboxDraftCreateHTTPCanonicalUUID(request.DraftClientID)
	if err != nil {
		return MailboxReplyDraftInput{}, err
	}
	fromIdentityID, err := mailboxDraftCreateHTTPCanonicalUUID(request.FromIdentityID)
	if err != nil {
		return MailboxReplyDraftInput{}, err
	}
	return MailboxReplyDraftInput{
		SourceMessageID: sourceMessageID,
		DraftClientID:   draftClientID,
		FromIdentityID:  fromIdentityID,
		IdempotencyKey:  idempotencyKey,
	}, nil
}

func mailboxReplyDraftHTTPViewMatches(view MailboxDraftView, sourceMessageID uuid.UUID, input MailboxReplyDraftInput) bool {
	return sourceMessageID != uuid.Nil && input.DraftClientID != uuid.Nil && input.FromIdentityID != uuid.Nil &&
		view.ID != uuid.Nil && view.DraftClientID == input.DraftClientID && view.Status == MailboxDraftStatusActive &&
		view.Effective.FromIdentityID == input.FromIdentityID && view.Effective.ReplyToMessageID != nil &&
		*view.Effective.ReplyToMessageID == sourceMessageID && view.Effective.ForwardMessageID == nil
}
