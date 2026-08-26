package mail

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type mailboxReplyAllDraftCreator interface {
	CreateMailboxReplyAllDraft(context.Context, uuid.UUID, MailboxReplyAllDraftInput) (MailboxDraftView, error)
}

// MailboxReplyAllDraftHandler exposes the authenticated server-authoritative Reply All bootstrap transport.
type MailboxReplyAllDraftHandler struct {
	service mailboxReplyAllDraftCreator
}

// NewMailboxReplyAllDraftHandler creates the authenticated Reply All draft handler.
func NewMailboxReplyAllDraftHandler(service mailboxReplyAllDraftCreator) (*MailboxReplyAllDraftHandler, error) {
	if mailboxHTTPDependencyNil(service) {
		return nil, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return &MailboxReplyAllDraftHandler{service: service}, nil
}

// RegisterRoutes registers only POST /api/mail/messages/:messageId/reply-all-draft on the supplied /api router.
func (handler *MailboxReplyAllDraftHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) || mailboxHTTPDependencyNil(api) || mailboxHTTPDependencyNil(authMiddleware) {
		return mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	api.Group("/mail", authMiddleware).Post("/messages/:messageId/reply-all-draft", handler.create)
	return nil
}

func (handler *MailboxReplyAllDraftHandler) create(ctx fiber.Ctx) error {
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
	input, err := mailboxReplyAllDraftHTTPInput(sourceMessageID, request, ctx.Get("Idempotency-Key"))
	if err != nil {
		return err
	}
	view, err := handler.service.CreateMailboxReplyAllDraft(ctx.Context(), actorID, input)
	if err != nil {
		return mailboxDraftCreateHTTPServiceError(ctx, err)
	}
	if !mailboxReplyAllDraftHTTPViewMatches(view, sourceMessageID, input) {
		return mailboxDraftCreateHTTPUnavailable()
	}
	response, ok := mailboxDraftCreateHTTPResponseFromView(view)
	if !ok {
		return mailboxDraftCreateHTTPUnavailable()
	}
	return ctx.Status(http.StatusCreated).JSON(response)
}

func mailboxReplyAllDraftHTTPInput(
	sourceMessageID uuid.UUID,
	request mailboxReplyDraftHTTPRequest,
	idempotencyKey string,
) (MailboxReplyAllDraftInput, error) {
	if sourceMessageID == uuid.Nil {
		return MailboxReplyAllDraftInput{}, mailboxDraftCreateHTTPInvalidRequest()
	}
	draftClientID, err := mailboxDraftCreateHTTPCanonicalUUID(request.DraftClientID)
	if err != nil {
		return MailboxReplyAllDraftInput{}, err
	}
	fromIdentityID, err := mailboxDraftCreateHTTPCanonicalUUID(request.FromIdentityID)
	if err != nil {
		return MailboxReplyAllDraftInput{}, err
	}
	return MailboxReplyAllDraftInput{
		SourceMessageID: sourceMessageID,
		DraftClientID:   draftClientID,
		FromIdentityID:  fromIdentityID,
		IdempotencyKey:  idempotencyKey,
	}, nil
}

func mailboxReplyAllDraftHTTPViewMatches(view MailboxDraftView, sourceMessageID uuid.UUID, input MailboxReplyAllDraftInput) bool {
	return mailboxReplyDraftHTTPViewMatches(view, sourceMessageID, MailboxReplyDraftInput{
		SourceMessageID: input.SourceMessageID,
		DraftClientID:   input.DraftClientID,
		FromIdentityID:  input.FromIdentityID,
		IdempotencyKey:  input.IdempotencyKey,
	})
}
