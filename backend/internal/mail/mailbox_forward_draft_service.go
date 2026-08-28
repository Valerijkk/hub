package mail

import (
	"context"
	"strings"

	"github.com/google/uuid"
)

// MailboxForwardDraftServiceAPI is the narrow server-authoritative forward-draft bootstrap surface.
type MailboxForwardDraftServiceAPI interface {
	CreateMailboxForwardDraft(context.Context, uuid.UUID, MailboxForwardDraftInput) (MailboxDraftView, error)
}

type mailboxForwardDrafts interface {
	CreateMailboxDraft(context.Context, uuid.UUID, MailboxDraftCreateInput) (MailboxDraftView, error)
	GetMailboxDraft(context.Context, uuid.UUID, uuid.UUID) (MailboxDraftView, error)
	preflightMailboxForwardDraft(context.Context, uuid.UUID, mailboxForwardDraftPreflightInput) (mailboxForwardDraftPreflightResult, error)
}

// MailboxForwardDraftInput deliberately contains no client-controlled recipients, threading headers, or body.
type MailboxForwardDraftInput struct {
	SourceMessageID uuid.UUID
	DraftClientID   uuid.UUID
	FromIdentityID  uuid.UUID
	IdempotencyKey  string
}

// MailboxForwardDraftService derives one forward draft exclusively from a sealed, owner-scoped inbound message.
type MailboxForwardDraftService struct {
	owners   mailboxOwnerResolver
	messages MailboxMessageRepository
	engine   MailboxMessageReadEngine
	drafts   mailboxForwardDrafts
}

// NewMailboxForwardDraftService creates the server-side forward draft bootstrap service.
func NewMailboxForwardDraftService(
	owners mailboxOwnerResolver,
	messages MailboxMessageRepository,
	engine MailboxMessageReadEngine,
	drafts mailboxForwardDrafts,
) (*MailboxForwardDraftService, error) {
	if mailboxDraftServiceDependencyNil(owners) ||
		mailboxDraftServiceDependencyNil(messages) ||
		mailboxDraftServiceDependencyNil(engine) ||
		mailboxDraftServiceDependencyNil(drafts) {
		return nil, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return &MailboxForwardDraftService{
		owners: owners, messages: messages, engine: engine, drafts: drafts,
	}, nil
}

// CreateMailboxForwardDraft creates a durable forward draft from trusted source metadata and plain content only.
func (service *MailboxForwardDraftService) CreateMailboxForwardDraft(ctx context.Context, actorID uuid.UUID, input MailboxForwardDraftInput) (MailboxDraftView, error) {
	if service == nil ||
		mailboxDraftServiceDependencyNil(service.owners) ||
		mailboxDraftServiceDependencyNil(service.messages) ||
		mailboxDraftServiceDependencyNil(service.engine) ||
		mailboxDraftServiceDependencyNil(service.drafts) ||
		ctx == nil || actorID == uuid.Nil || input.SourceMessageID == uuid.Nil ||
		input.DraftClientID == uuid.Nil || input.FromIdentityID == uuid.Nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if ctx.Err() != nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}

	idempotencyKey, err := mailboxDraftCreateNormalizeIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return MailboxDraftView{}, err
	}
	preflight, err := service.drafts.preflightMailboxForwardDraft(ctx, actorID, mailboxForwardDraftPreflightInput{
		DraftClientID: input.DraftClientID, FromIdentityID: input.FromIdentityID, SourceMessageID: input.SourceMessageID, IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return MailboxDraftView{}, err
	}
	if !mailboxForwardDraftPreflightResultValid(preflight) {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	switch preflight.Kind {
	case mailboxForwardDraftPreflightAbsent:
	case mailboxForwardDraftPreflightActive:
		view, err := service.drafts.GetMailboxDraft(ctx, actorID, preflight.DraftID)
		if err != nil {
			return MailboxDraftView{}, err
		}
		if !mailboxForwardDraftPreflightActiveViewMatches(view, preflight.DraftID, input) {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorIdempotencyConflict)
		}
		return view, nil
	case mailboxForwardDraftPreflightBusy:
		return MailboxDraftView{}, mailboxDraftServiceBusyError(*preflight.BusyUntil)
	case mailboxForwardDraftPreflightUnknown:
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	default:
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxDraftView{}, mailboxDraftGetOwnerError(err)
	}
	scope, err := mailboxMessageReadScopeFromOwner(actorID, owner)
	if err != nil {
		return MailboxDraftView{}, mailboxReplyDraftMessageError(err)
	}
	snapshot, err := service.messages.ResolveMailboxMessage(ctx, scope, input.SourceMessageID)
	if err != nil {
		return MailboxDraftView{}, mailboxReplyDraftMessageError(err)
	}
	if !mailboxReplyDraftSnapshotMatches(snapshot, scope, input.SourceMessageID) {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if !mailboxReplyDraftSourceEligible(snapshot) {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorNotFound)
	}

	engineMessage, err := service.engine.GetMessage(ctx, EngineMessageReadRequest{
		AccountID: scope.EngineAccountID, MessageID: snapshot.engineMessageID, MaxContentBytes: mailboxMessageEngineReadMaxBytes,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxReplyDraftMessageError(mailboxMessageEngineError(err))
	}
	if !mailboxMessageProjectionMatches(snapshot, engineMessage.Projection) {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
	content, err := mailboxMessageNormalizeEngineBody(engineMessage.Body)
	if err != nil {
		return MailboxDraftView{}, mailboxReplyDraftMessageError(err)
	}
	if err := service.messages.RecheckMailboxMessage(ctx, scope, snapshot); err != nil {
		return MailboxDraftView{}, mailboxReplyDraftMessageError(err)
	}

	subject, err := mailboxForwardDraftSubject(snapshot.summary.Subject)
	if err != nil {
		return MailboxDraftView{}, err
	}
	forwarded, err := mailboxForwardDraftForwardedContent(content)
	if err != nil {
		return MailboxDraftView{}, err
	}

	sourceID := input.SourceMessageID
	return service.drafts.CreateMailboxDraft(ctx, actorID, MailboxDraftCreateInput{
		MailboxDraftWriteInput: MailboxDraftWriteInput{
			DraftClientID:    input.DraftClientID,
			FromIdentityID:   input.FromIdentityID,
			To:               nil,
			CC:               nil,
			BCC:              nil,
			ReplyTo:          nil,
			Subject:          subject,
			InReplyTo:        "",
			References:       nil,
			Priority:         EnginePriorityNormal,
			ReplyToMessageID: nil,
			ForwardMessageID: &sourceID,
			Attachments:      nil,
			Content:          forwarded,
		},
		IdempotencyKey: idempotencyKey,
	})
}

func mailboxForwardDraftPreflightActiveViewMatches(view MailboxDraftView, draftID uuid.UUID, input MailboxForwardDraftInput) bool {
	return draftID != uuid.Nil && view.ID == draftID && view.DraftClientID == input.DraftClientID && view.Status == MailboxDraftStatusActive &&
		view.Effective.FromIdentityID == input.FromIdentityID && view.Effective.ReplyToMessageID == nil &&
		view.Effective.ForwardMessageID != nil && *view.Effective.ForwardMessageID == input.SourceMessageID
}

func mailboxForwardDraftSubject(source string) (string, error) {
	base := strings.TrimSpace(source)
	for len(base) >= len("fwd:") && strings.EqualFold(base[:len("fwd:")], "fwd:") {
		base = strings.TrimSpace(base[len("fwd:"):])
	}
	result := "Fwd:"
	if base != "" {
		result += " " + base
	}
	if !mailboxDraftRequestHashHeaderValid(result) {
		return "", mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	}
	return result, nil
}

func mailboxForwardDraftForwardedContent(content MailboxContentResult) (MailboxContentInput, error) {
	if !mailboxContentResultValid(content) {
		return MailboxContentInput{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	plain := strings.ReplaceAll(strings.ReplaceAll(content.PlainBody, "\r\n", "\n"), "\r", "\n")
	var builder strings.Builder
	if err := mailboxContentAppend(&builder, "---------- Forwarded message ----------"); err != nil {
		return MailboxContentInput{}, mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
	}
	if plain != "" {
		if err := mailboxContentAppend(&builder, "\n"); err != nil {
			return MailboxContentInput{}, mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
		}
		if err := mailboxContentAppend(&builder, plain); err != nil {
			return MailboxContentInput{}, mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
		}
	}
	normalized, err := NormalizeMailboxContent(MailboxContentInput{
		SourceFormat: MailboxContentFormatPlain,
		Body:         builder.String(),
	})
	if err != nil {
		return MailboxContentInput{}, mailboxDraftCreateContentError(err)
	}
	if normalized.SourceFormat != MailboxContentFormatPlain || !mailboxContentResultValid(normalized) {
		return MailboxContentInput{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return MailboxContentInput{SourceFormat: MailboxContentFormatPlain, Body: normalized.PlainBody}, nil
}
