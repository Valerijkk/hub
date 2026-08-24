package mail

import (
	"context"
	"net/http"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

type mailboxForwardDraftCreator interface {
	CreateMailboxForwardDraft(context.Context, uuid.UUID, MailboxForwardDraftInput) (MailboxDraftView, error)
}

// MailboxForwardDraftHandler exposes the authenticated server-authoritative forward-draft bootstrap transport.
type MailboxForwardDraftHandler struct {
	service mailboxForwardDraftCreator
}

// NewMailboxForwardDraftHandler creates the authenticated forward-draft handler.
func NewMailboxForwardDraftHandler(service mailboxForwardDraftCreator) (*MailboxForwardDraftHandler, error) {
	if mailboxHTTPDependencyNil(service) {
		return nil, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return &MailboxForwardDraftHandler{service: service}, nil
}

// RegisterRoutes registers only POST /api/mail/messages/:messageId/forward-draft on the supplied /api router.
func (handler *MailboxForwardDraftHandler) RegisterRoutes(api fiber.Router, authMiddleware fiber.Handler) error {
	if handler == nil || mailboxHTTPDependencyNil(handler.service) || mailboxHTTPDependencyNil(api) || mailboxHTTPDependencyNil(authMiddleware) {
		return mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	api.Group("/mail", authMiddleware).Post("/messages/:messageId/forward-draft", handler.create)
	return nil
}

type mailboxForwardDraftHTTPRequest struct {
	DraftClientID  string `json:"draft_client_id"`
	FromIdentityID string `json:"from_identity_id"`
}

func (handler *MailboxForwardDraftHandler) create(ctx fiber.Ctx) error {
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
	var request mailboxForwardDraftHTTPRequest
	if err := mailboxDraftCreateHTTPDecodeJSON(ctx, &request); err != nil {
		return err
	}
	input, err := mailboxForwardDraftHTTPInput(sourceMessageID, request, ctx.Get("Idempotency-Key"))
	if err != nil {
		return err
	}
	view, err := handler.service.CreateMailboxForwardDraft(ctx.Context(), actorID, input)
	if err != nil {
		return mailboxDraftCreateHTTPServiceError(ctx, err)
	}
	if !mailboxForwardDraftHTTPViewMatches(view, sourceMessageID, input) {
		return mailboxDraftCreateHTTPUnavailable()
	}
	response, ok := mailboxDraftCreateHTTPResponseFromView(view)
	if !ok {
		return mailboxDraftCreateHTTPUnavailable()
	}
	return ctx.Status(http.StatusCreated).JSON(response)
}

func mailboxForwardDraftHTTPInput(sourceMessageID uuid.UUID, request mailboxForwardDraftHTTPRequest, idempotencyKey string) (MailboxForwardDraftInput, error) {
	if sourceMessageID == uuid.Nil {
		return MailboxForwardDraftInput{}, mailboxDraftCreateHTTPInvalidRequest()
	}
	draftClientID, err := mailboxDraftCreateHTTPCanonicalUUID(request.DraftClientID)
	if err != nil {
		return MailboxForwardDraftInput{}, err
	}
	fromIdentityID, err := mailboxDraftCreateHTTPCanonicalUUID(request.FromIdentityID)
	if err != nil {
		return MailboxForwardDraftInput{}, err
	}
	return MailboxForwardDraftInput{
		SourceMessageID: sourceMessageID,
		DraftClientID:   draftClientID,
		FromIdentityID:  fromIdentityID,
		IdempotencyKey:  idempotencyKey,
	}, nil
}

func mailboxForwardDraftHTTPViewMatches(view MailboxDraftView, sourceMessageID uuid.UUID, input MailboxForwardDraftInput) bool {
	return sourceMessageID != uuid.Nil && input.DraftClientID != uuid.Nil && input.FromIdentityID != uuid.Nil &&
		view.ID != uuid.Nil && view.DraftClientID == input.DraftClientID && view.Status == MailboxDraftStatusActive &&
		view.Effective.FromIdentityID == input.FromIdentityID && view.Effective.ReplyToMessageID == nil &&
		view.Effective.ForwardMessageID != nil && *view.Effective.ForwardMessageID == sourceMessageID
}
