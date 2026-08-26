package mail

import (
	"context"
	"crypto/sha256"
	"encoding/base64"

	"github.com/google/uuid"
)

const mailboxReplyAllDraftIdempotencyNamespace = "hub.mail.reply-all-draft.v1\x00"

// MailboxReplyAllDraftServiceAPI is the server-authoritative Reply All bootstrap surface.
type MailboxReplyAllDraftServiceAPI interface {
	CreateMailboxReplyAllDraft(context.Context, uuid.UUID, MailboxReplyAllDraftInput) (MailboxDraftView, error)
}

// MailboxReplyAllDraftInput contains no client-controlled recipients, threading headers, or body.
type MailboxReplyAllDraftInput struct {
	DraftClientID   uuid.UUID
	FromIdentityID  uuid.UUID
	SourceMessageID uuid.UUID
	IdempotencyKey  string
}

// MailboxReplyAllIdentityProofResolver returns the complete bounded set of active owner addresses.
type MailboxReplyAllIdentityProofResolver interface {
	ResolveMailboxReplyAllIdentityProof(context.Context, mailboxReplyAllIdentityProofRequest) (mailboxReplyAllIdentityProof, error)
}

type mailboxReplyAllIdentityProofRequest struct {
	ActorID            uuid.UUID
	OrganizationID     uuid.UUID
	AccountID          uuid.UUID
	SelectedIdentityID uuid.UUID
}

type mailboxReplyAllIdentityProof struct {
	actorID            uuid.UUID
	organizationID     uuid.UUID
	accountID          uuid.UUID
	selectedIdentityID uuid.UUID
	selected           EngineAddress
	owned              []EngineAddress
}

// MailboxReplyAllDraftService derives one editable Reply All draft from trusted source state.
type MailboxReplyAllDraftService struct {
	owners     mailboxOwnerResolver
	messages   MailboxMessageRepository
	engine     MailboxMessageReadEngine
	drafts     mailboxReplyDrafts
	identities MailboxReplyAllIdentityProofResolver
}

var _ MailboxReplyAllDraftServiceAPI = (*MailboxReplyAllDraftService)(nil)

// NewMailboxReplyAllDraftService creates the server-side Reply All bootstrap service.
func NewMailboxReplyAllDraftService(
	owners mailboxOwnerResolver,
	messages MailboxMessageRepository,
	engine MailboxMessageReadEngine,
	drafts mailboxReplyDrafts,
	identities MailboxReplyAllIdentityProofResolver,
) (*MailboxReplyAllDraftService, error) {
	if mailboxDraftServiceDependencyNil(owners) ||
		mailboxDraftServiceDependencyNil(messages) ||
		mailboxDraftServiceDependencyNil(engine) ||
		mailboxDraftServiceDependencyNil(drafts) ||
		mailboxDraftServiceDependencyNil(identities) {
		return nil, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return &MailboxReplyAllDraftService{
		owners:     owners,
		messages:   messages,
		engine:     engine,
		drafts:     drafts,
		identities: identities,
	}, nil
}

// CreateMailboxReplyAllDraft creates or exactly replays one server-derived Reply All draft.
func (service *MailboxReplyAllDraftService) CreateMailboxReplyAllDraft(
	ctx context.Context,
	actorID uuid.UUID,
	input MailboxReplyAllDraftInput,
) (MailboxDraftView, error) {
	if service == nil ||
		mailboxDraftServiceDependencyNil(service.owners) ||
		mailboxDraftServiceDependencyNil(service.messages) ||
		mailboxDraftServiceDependencyNil(service.engine) ||
		mailboxDraftServiceDependencyNil(service.drafts) ||
		mailboxDraftServiceDependencyNil(service.identities) ||
		ctx == nil || actorID == uuid.Nil || input.DraftClientID == uuid.Nil ||
		input.FromIdentityID == uuid.Nil || input.SourceMessageID == uuid.Nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if ctx.Err() != nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}

	namespacedKey, err := mailboxReplyAllDraftNamespacedKey(input.IdempotencyKey)
	if err != nil {
		return MailboxDraftView{}, err
	}
	preflight, err := service.drafts.preflightMailboxReplyDraft(ctx, actorID, mailboxReplyDraftPreflightInput{
		DraftClientID: input.DraftClientID, FromIdentityID: input.FromIdentityID,
		SourceMessageID: input.SourceMessageID, IdempotencyKey: namespacedKey,
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
		if !mailboxReplyDraftPreflightActiveViewMatches(view, preflight.DraftID, mailboxReplyAllDraftReplyInput(input, namespacedKey)) {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
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
		AccountID: scope.EngineAccountID, MessageID: snapshot.engineMessageID,
		MaxContentBytes: mailboxMessageEngineReadMaxBytes,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxReplyDraftMessageError(err)
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

	proof, err := service.identities.ResolveMailboxReplyAllIdentityProof(ctx, mailboxReplyAllIdentityProofRequest{
		ActorID: scope.ActorID, OrganizationID: scope.OrganizationID, AccountID: scope.AccountID,
		SelectedIdentityID: input.FromIdentityID,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftCreateRepositoryError(err)
	}
	ownedEmails, ok := mailboxReplyAllDraftOwnedEmails(proof, scope, input.FromIdentityID)
	if !ok {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	to, cc, err := mailboxReplyAllDraftRecipients(snapshot, ownedEmails)
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
	draft, err := service.drafts.CreateMailboxDraft(ctx, actorID, MailboxDraftCreateInput{
		MailboxDraftWriteInput: MailboxDraftWriteInput{
			DraftClientID: input.DraftClientID, FromIdentityID: input.FromIdentityID,
			To: to, CC: cc, BCC: nil, ReplyTo: nil, Subject: subject,
			InReplyTo: snapshot.internetMessageID, References: references, Priority: EnginePriorityNormal,
			ReplyToMessageID: &sourceID, ForwardMessageID: nil, Attachments: nil, Content: quoted,
		},
		IdempotencyKey: namespacedKey,
	})
	if err != nil {
		return MailboxDraftView{}, err
	}
	if !mailboxReplyAllDraftCreatedViewMatches(draft, input, namespacedKey, to, cc) {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	return draft, nil
}

func mailboxReplyAllDraftNamespacedKey(value string) (string, error) {
	normalized, err := mailboxDraftCreateNormalizeIdempotencyKey(value)
	if err != nil {
		return "", err
	}
	digest := sha256.Sum256([]byte(mailboxReplyAllDraftIdempotencyNamespace + normalized))
	return base64.RawURLEncoding.EncodeToString(digest[:]), nil
}

func mailboxReplyAllDraftReplyInput(input MailboxReplyAllDraftInput, key string) MailboxReplyDraftInput {
	return MailboxReplyDraftInput{
		DraftClientID: input.DraftClientID, FromIdentityID: input.FromIdentityID,
		SourceMessageID: input.SourceMessageID, IdempotencyKey: key,
	}
}

func mailboxReplyAllDraftOwnedEmails(
	proof mailboxReplyAllIdentityProof,
	scope MailboxMessageReadScope,
	selectedIdentityID uuid.UUID,
) (map[string]struct{}, bool) {
	if proof.actorID != scope.ActorID || proof.organizationID != scope.OrganizationID ||
		proof.accountID != scope.AccountID || proof.selectedIdentityID != selectedIdentityID ||
		len(proof.owned) < 1 || len(proof.owned) > maxMailboxIdentityLimit {
		return nil, false
	}
	selected, err := NormalizeAddress(proof.selected)
	if err != nil || selected != proof.selected {
		return nil, false
	}
	emails := make(map[string]struct{}, len(proof.owned))
	for _, address := range append([]EngineAddress(nil), proof.owned...) {
		normalized, err := NormalizeAddress(address)
		if err != nil || normalized != address {
			return nil, false
		}
		if _, duplicate := emails[normalized.Email]; duplicate {
			return nil, false
		}
		emails[normalized.Email] = struct{}{}
	}
	if _, present := emails[selected.Email]; !present {
		return nil, false
	}
	return emails, true
}

func mailboxReplyAllDraftRecipients(snapshot MailboxMessageSnapshot, ownedEmails map[string]struct{}) ([]EngineAddress, []EngineAddress, error) {
	if len(ownedEmails) == 0 {
		return nil, nil, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	replyTargets := append([]EngineAddress(nil), snapshot.replyTo...)
	if len(replyTargets) == 0 {
		replyTargets = []EngineAddress{snapshot.from}
	}
	toCandidates := make([]EngineAddress, 0, len(replyTargets)+len(snapshot.to))
	toCandidates = append(toCandidates, replyTargets...)
	toCandidates = append(toCandidates, snapshot.to...)
	ccCandidates := append([]EngineAddress(nil), snapshot.cc...)

	seen := make(map[string]struct{}, len(toCandidates)+len(ccCandidates))
	to := make([]EngineAddress, 0, len(toCandidates))
	cc := make([]EngineAddress, 0, len(ccCandidates))
	appendCandidate := func(destination *[]EngineAddress, candidate EngineAddress) error {
		normalized, err := NormalizeAddress(candidate)
		if err != nil {
			return mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
		}
		if _, owned := ownedEmails[normalized.Email]; owned {
			return nil
		}
		if _, duplicate := seen[normalized.Email]; duplicate {
			return nil
		}
		if len(*destination) >= MaxEngineDraftRecipientItems {
			return mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
		}
		seen[normalized.Email] = struct{}{}
		*destination = append(*destination, normalized)
		return nil
	}
	for _, candidate := range toCandidates {
		if err := appendCandidate(&to, candidate); err != nil {
			return nil, nil, err
		}
	}
	for _, candidate := range ccCandidates {
		if err := appendCandidate(&cc, candidate); err != nil {
			return nil, nil, err
		}
	}
	if len(to) == 0 || len(to)+len(cc) > MaxEngineDraftTotalRecipientItems {
		return nil, nil, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	}
	return to, cc, nil
}

func mailboxReplyAllDraftCreatedViewMatches(
	view MailboxDraftView,
	input MailboxReplyAllDraftInput,
	namespacedKey string,
	to []EngineAddress,
	cc []EngineAddress,
) bool {
	if !mailboxReplyDraftPreflightActiveViewMatches(view, view.ID, mailboxReplyAllDraftReplyInput(input, namespacedKey)) ||
		!mailboxReplyAllDraftAddressesEqual(view.Content.To, to) ||
		!mailboxReplyAllDraftAddressesEqual(view.Content.CC, cc) ||
		len(view.Content.BCC) != 0 || len(view.Content.ReplyTo) != 0 {
		return false
	}
	return true
}

func mailboxReplyAllDraftAddressesEqual(left, right []EngineAddress) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
