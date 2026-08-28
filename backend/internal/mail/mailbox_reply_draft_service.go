package mail

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"github.com/google/uuid"
)

// MailboxReplyDraftServiceAPI is the narrow server-authoritative reply-draft bootstrap surface.
type MailboxReplyDraftServiceAPI interface {
	CreateMailboxReplyDraft(context.Context, uuid.UUID, MailboxReplyDraftInput) (MailboxDraftView, error)
}

type mailboxReplyDrafts interface {
	CreateMailboxDraft(context.Context, uuid.UUID, MailboxDraftCreateInput) (MailboxDraftView, error)
	GetMailboxDraft(context.Context, uuid.UUID, uuid.UUID) (MailboxDraftView, error)
	preflightMailboxReplyDraft(context.Context, uuid.UUID, mailboxReplyDraftPreflightInput) (mailboxReplyDraftPreflightResult, error)
}

// MailboxReplyDraftInput deliberately contains no client-controlled recipients, threading headers, or body.
type MailboxReplyDraftInput struct {
	SourceMessageID uuid.UUID
	DraftClientID   uuid.UUID
	FromIdentityID  uuid.UUID
	IdempotencyKey  string
}

// MailboxReplyDraftService derives one reply draft exclusively from a sealed, owner-scoped inbound message.
type MailboxReplyDraftService struct {
	owners     mailboxOwnerResolver
	messages   MailboxMessageRepository
	engine     MailboxMessageReadEngine
	drafts     mailboxReplyDrafts
}

// NewMailboxReplyDraftService creates the server-side reply draft bootstrap service.
func NewMailboxReplyDraftService(
	owners mailboxOwnerResolver,
	messages MailboxMessageRepository,
	engine MailboxMessageReadEngine,
	drafts mailboxReplyDrafts,
) (*MailboxReplyDraftService, error) {
	if mailboxDraftServiceDependencyNil(owners) ||
		mailboxDraftServiceDependencyNil(messages) ||
		mailboxDraftServiceDependencyNil(engine) ||
		mailboxDraftServiceDependencyNil(drafts) {
		return nil, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return &MailboxReplyDraftService{
		owners: owners, messages: messages, engine: engine, drafts: drafts,
	}, nil
}

// CreateMailboxReplyDraft creates a durable reply draft from trusted source metadata and content only.
func (service *MailboxReplyDraftService) CreateMailboxReplyDraft(ctx context.Context, actorID uuid.UUID, input MailboxReplyDraftInput) (MailboxDraftView, error) {
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
	preflight, err := service.drafts.preflightMailboxReplyDraft(ctx, actorID, mailboxReplyDraftPreflightInput{
		DraftClientID: input.DraftClientID, FromIdentityID: input.FromIdentityID, SourceMessageID: input.SourceMessageID, IdempotencyKey: idempotencyKey,
	})
	if err != nil {
		return MailboxDraftView{}, err
	}
	if !mailboxReplyDraftPreflightResultValid(preflight) {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	switch preflight.Kind {
	case mailboxReplyDraftPreflightAbsent:
	case mailboxReplyDraftPreflightActive:
		view, err := service.drafts.GetMailboxDraft(ctx, actorID, preflight.DraftID)
		if err != nil {
			return MailboxDraftView{}, err
		}
		if !mailboxReplyDraftPreflightActiveViewMatches(view, preflight.DraftID, input) {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorIdempotencyConflict)
		}
		return view, nil
	case mailboxReplyDraftPreflightBusy:
		return MailboxDraftView{}, mailboxDraftServiceBusyError(*preflight.BusyUntil)
	case mailboxReplyDraftPreflightUnknown:
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

	to, err := mailboxReplyDraftRecipients(snapshot)
	if err != nil {
		return MailboxDraftView{}, err
	}
	subject, err := mailboxReplyDraftSubject(snapshot.summary.Subject)
	if err != nil {
		return MailboxDraftView{}, err
	}
	references, err := mailboxReplyDraftReferences(snapshot)
	if err != nil {
		return MailboxDraftView{}, err
	}
	quoted, err := mailboxReplyDraftQuotedContent(content)
	if err != nil {
		return MailboxDraftView{}, err
	}

	sourceID := input.SourceMessageID
	return service.drafts.CreateMailboxDraft(ctx, actorID, MailboxDraftCreateInput{
		MailboxDraftWriteInput: MailboxDraftWriteInput{
			DraftClientID:    input.DraftClientID,
			FromIdentityID:   input.FromIdentityID,
			To:               to,
			CC:               nil,
			BCC:              nil,
			ReplyTo:          nil,
			Subject:          subject,
			InReplyTo:        snapshot.internetMessageID,
			References:       references,
			Priority:         EnginePriorityNormal,
			ReplyToMessageID: &sourceID,
			ForwardMessageID: nil,
			Attachments:      nil,
			Content:          quoted,
		},
		IdempotencyKey: idempotencyKey,
	})
}

func mailboxReplyDraftSnapshotMatches(snapshot MailboxMessageSnapshot, scope MailboxMessageReadScope, sourceID uuid.UUID) bool {
	return sourceID != uuid.Nil && mailboxMessageReadScopeValid(scope) && mailboxMessageSnapshotValid(snapshot) &&
		snapshot.summary.ID == sourceID && snapshot.engineAccountID == scope.EngineAccountID
}

func mailboxReplyDraftPreflightActiveViewMatches(view MailboxDraftView, draftID uuid.UUID, input MailboxReplyDraftInput) bool {
	return draftID != uuid.Nil && view.ID == draftID && view.DraftClientID == input.DraftClientID && view.Status == MailboxDraftStatusActive &&
		view.Effective.FromIdentityID == input.FromIdentityID && view.Effective.ReplyToMessageID != nil &&
		*view.Effective.ReplyToMessageID == input.SourceMessageID && view.Effective.ForwardMessageID == nil
}

func mailboxReplyDraftSourceEligible(snapshot MailboxMessageSnapshot) bool {
	return !snapshot.summary.IsDraft && snapshot.projectionOrigin == EngineProjectionOriginInboundOrUnknown &&
		snapshot.summary.CalendarMethod == nil && len(snapshot.calendarPartLocatorHash) == 0
}

func mailboxReplyDraftRecipients(snapshot MailboxMessageSnapshot) ([]EngineAddress, error) {
	values := snapshot.replyTo
	if len(values) == 0 {
		values = []EngineAddress{snapshot.from}
	}
	normalized, err := NormalizeAddressSet(values, MaxEngineDraftRecipientItems)
	if err != nil || len(normalized) == 0 {
		return nil, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	}
	return normalized, nil
}

func mailboxReplyDraftSubject(source string) (string, error) {
	base := strings.TrimSpace(source)
	for len(base) >= len("re:") && strings.EqualFold(base[:len("re:")], "re:") {
		base = strings.TrimSpace(base[len("re:"):])
	}
	result := "Re:"
	if base != "" {
		result += " " + base
	}
	if !mailboxDraftRequestHashHeaderValid(result) {
		return "", mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	}
	return result, nil
}

func mailboxReplyDraftReferences(snapshot MailboxMessageSnapshot) ([]string, error) {
	if !validThreadMessageID(snapshot.internetMessageID) {
		return nil, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	}
	seen := make(map[string]struct{}, len(snapshot.references)+1)
	values := make([]string, 0, len(snapshot.references)+1)
	for _, value := range snapshot.references {
		if !validThreadMessageID(value) {
			return nil, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
		}
		if _, exists := seen[value]; exists {
			return nil, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
		}
		seen[value] = struct{}{}
		if value != snapshot.internetMessageID {
			values = append(values, value)
		}
	}
	if len(values) >= MaxEngineReferenceItems {
		values = append([]string(nil), values[len(values)-(MaxEngineReferenceItems-1):]...)
	}
	values = append(values, snapshot.internetMessageID)
	return values, nil
}

func mailboxReplyDraftQuotedContent(content MailboxContentResult) (MailboxContentInput, error) {
	if !mailboxContentResultValid(content) {
		return MailboxContentInput{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	plain := strings.ReplaceAll(strings.ReplaceAll(content.PlainBody, "\r\n", "\n"), "\r", "\n")
	quoted, err := mailboxReplyDraftQuotePlain(plain)
	if err != nil {
		return MailboxContentInput{}, err
	}
	normalized, err := NormalizeMailboxContent(MailboxContentInput{SourceFormat: MailboxContentFormatPlain, Body: quoted})
	if err != nil {
		switch {
		case errors.Is(err, ErrMailboxContentTooLarge):
			return MailboxContentInput{}, mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
		case errors.Is(err, ErrMailboxContentInvalid):
			return MailboxContentInput{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
		default:
			return MailboxContentInput{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
		}
	}
	if normalized.SourceFormat != MailboxContentFormatPlain || !mailboxContentResultValid(normalized) {
		return MailboxContentInput{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return MailboxContentInput{SourceFormat: MailboxContentFormatPlain, Body: normalized.PlainBody}, nil
}

func mailboxReplyDraftQuotePlain(plain string) (string, error) {
	if plain == "" {
		return "", nil
	}
	var builder strings.Builder
	for first := true; ; first = false {
		line, remainder, found := strings.Cut(plain, "\n")
		if !first {
			if err := mailboxContentAppend(&builder, "\n"); err != nil {
				return "", mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
			}
		}
		if err := mailboxContentAppend(&builder, "> "); err != nil {
			return "", mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
		}
		if err := mailboxContentAppend(&builder, line); err != nil {
			return "", mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
		}
		if !found {
			break
		}
		plain = remainder
	}
	quoted := builder.String()
	if len(quoted) > MailboxContentMaxPlainBytes || utf8.RuneCountInString(quoted) > MailboxContentMaxRunes {
		return "", mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
	}
	return quoted, nil
}

func mailboxReplyDraftMessageError(err error) error {
	kind, ok := mailboxMessageReadErrorKindOf(err)
	if !ok {
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
	switch kind {
	case MailboxMessageReadErrorNotFound:
		return mailboxDraftServiceError(MailboxDraftServiceErrorNotFound)
	case MailboxMessageReadErrorDeleting:
		return mailboxDraftServiceError(MailboxDraftServiceErrorDeleting)
	case MailboxMessageReadErrorUnavailable:
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	case MailboxMessageReadErrorPayloadTooLarge:
		return mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
	case MailboxMessageReadErrorUnprocessable:
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	default:
		return mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
}
