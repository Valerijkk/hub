package mail

import (
	"context"
	"crypto/sha256"
	"errors"
	"sort"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

const mailboxDraftCreateMaxIdempotencyKeyBytes = 128

type mailboxDraftCreateScope struct {
	actorID         uuid.UUID
	organizationID  uuid.UUID
	accountID       uuid.UUID
	engineAccountID string
}

type mailboxDraftCreatePrepared struct {
	scope              mailboxDraftCreateScope
	draftID            uuid.UUID
	idempotencyKeyHash [sha256.Size]byte
	canonical          MailboxDraftCanonicalRequest
	effectiveState     MailboxDraftEffectiveState
	hasAttachments     bool
}

// CreateMailboxDraft creates an owner-bound draft only after the durable
// repository reservation has fenced the provider mutation. Attachment-backed
// creates use a private upload/import bridge; ordinary MailboxDraftEngine
// consumers never gain blob or stream access.
func (service *MailboxDraftService) CreateMailboxDraft(ctx context.Context, actorID uuid.UUID, input MailboxDraftCreateInput) (MailboxDraftView, error) {
	if !mailboxDraftCreateServiceValid(service) || ctx == nil || actorID == uuid.Nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if input.calendar != nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
	if ctx.Err() != nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}

	owner, err := service.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxDraftView{}, mailboxDraftGetOwnerError(err)
	}
	scope, err := mailboxDraftCreateScopeFromOwner(actorID, owner)
	if err != nil {
		return MailboxDraftView{}, err
	}

	normalizedKey, err := mailboxDraftCreateNormalizeIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return MailboxDraftView{}, err
	}
	idempotencyKeyHash := sha256.Sum256([]byte(normalizedKey))

	locator, err := service.repository.LookupMailboxDraftCreatePreHashLocator(ctx, MailboxDraftCreatePreHashLocatorRequest{
		AccountID:                scope.accountID,
		DraftClientID:            input.DraftClientID,
		CreateIdempotencyKeyHash: idempotencyKeyHash,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftCreateRepositoryError(err)
	}
	draftID, err := mailboxDraftCreateLocatorDraftID(locator)
	if err != nil {
		return MailboxDraftView{}, err
	}

	prepared, err := service.mailboxDraftCreatePrepare(ctx, scope, draftID, idempotencyKeyHash, input)
	if err != nil {
		return MailboxDraftView{}, err
	}
	if prepared.hasAttachments {
		if _, ok := service.engine.(mailboxDraftAttachmentCreateEngine); !ok {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
		}
	}

	replay, err := service.repository.LookupMailboxDraftCreateReplay(ctx, MailboxDraftCreateReplayLookupRequest{
		AccountID:                prepared.scope.accountID,
		DraftClientID:            prepared.canonical.DraftClientID,
		CreateIdempotencyKeyHash: prepared.idempotencyKeyHash,
		CreateRequestHash:        prepared.canonical.RequestHash,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftCreateRepositoryError(err)
	}
	if replay.Found {
		return service.mailboxDraftCreateReplay(ctx, actorID, replay)
	}
	if replay.DraftID != uuid.Nil || replay.LastMutationID != nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	// Recheck the pre-hash candidate before taking the engine fence. If a
	// concurrent create has appeared, exact replay remains safe; any other
	// candidate is an idempotency collision and reaches no provider call.
	locator, err = service.repository.LookupMailboxDraftCreatePreHashLocator(ctx, MailboxDraftCreatePreHashLocatorRequest{
		AccountID:                prepared.scope.accountID,
		DraftClientID:            prepared.canonical.DraftClientID,
		CreateIdempotencyKeyHash: prepared.idempotencyKeyHash,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftCreateRepositoryError(err)
	}
	if locator.Found {
		if locator.DraftID == uuid.Nil {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
		replay, err = service.repository.LookupMailboxDraftCreateReplay(ctx, MailboxDraftCreateReplayLookupRequest{
			AccountID:                prepared.scope.accountID,
			DraftClientID:            prepared.canonical.DraftClientID,
			CreateIdempotencyKeyHash: prepared.idempotencyKeyHash,
			CreateRequestHash:        prepared.canonical.RequestHash,
		})
		if err != nil {
			return MailboxDraftView{}, mailboxDraftCreateRepositoryError(err)
		}
		if replay.Found {
			return service.mailboxDraftCreateReplay(ctx, actorID, replay)
		}
		if replay.DraftID != uuid.Nil || replay.LastMutationID != nil {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorIdempotencyConflict)
	}
	if locator.DraftID != uuid.Nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	begin, err := service.repository.BeginEngineCommand(ctx, MailboxEngineCommandBeginRequest{AccountID: prepared.scope.accountID})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftCreateRepositoryError(err)
	}
	if !mailboxDraftCreateBeginMatches(begin, prepared.scope) {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	if !begin.Claimed {
		return MailboxDraftView{}, mailboxDraftServiceBusyError(*begin.BusyUntil)
	}

	mutationID := uuid.New()
	reservation, err := service.repository.ReserveMailboxDraftCreate(ctx, MailboxDraftCreateRequest{
		Fence:                    *begin.Fence,
		DraftID:                  prepared.draftID,
		DraftClientID:            prepared.canonical.DraftClientID,
		CreateIdempotencyKeyHash: prepared.idempotencyKeyHash,
		CreateRequestHash:        prepared.canonical.RequestHash,
		Confirmed:                mailboxDraftCreateEffectiveStateCopy(prepared.effectiveState),
		MutationID:               mutationID,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftCreateRepositoryError(err)
	}

	switch reservation.Kind {
	case MailboxDraftReservationReplay:
		if !mailboxDraftCreateReplayReservationValid(reservation, prepared) {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
		}
		return service.GetMailboxDraft(ctx, actorID, reservation.Draft.ID)
	case MailboxDraftReservationReconcileRequired:
		return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, reservation)
	case MailboxDraftReservationDispatch:
		if !mailboxDraftCreateDispatchReservationValid(reservation, prepared, prepared.hasAttachments) {
			return MailboxDraftView{}, service.mailboxDraftCreateClaimedUnknown(ctx, *begin.Fence, prepared.draftID, mutationID)
		}
		if prepared.hasAttachments {
			bridge := service.engine.(mailboxDraftAttachmentCreateEngine)
			return service.mailboxDraftCreateWithAttachments(ctx, actorID, prepared, bridge, reservation)
		}
		return service.mailboxDraftCreateWithoutAttachments(ctx, actorID, prepared, reservation)
	default:
		return MailboxDraftView{}, service.mailboxDraftCreateClaimedUnknown(ctx, *begin.Fence, prepared.draftID, mutationID)
	}
}

func (service *MailboxDraftService) mailboxDraftCreatePrepare(ctx context.Context, scope mailboxDraftCreateScope, draftID uuid.UUID, idempotencyKeyHash [sha256.Size]byte, input MailboxDraftCreateInput) (mailboxDraftCreatePrepared, error) {
	if service == nil || ctx == nil || !mailboxDraftCreateScopeValid(scope) || draftID == uuid.Nil || !mailboxDraftCreateHashPresent(idempotencyKeyHash) || input.DraftClientID == uuid.Nil || input.FromIdentityID == uuid.Nil {
		return mailboxDraftCreatePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	content, err := NormalizeMailboxContent(input.Content)
	if err != nil {
		return mailboxDraftCreatePrepared{}, mailboxDraftCreateContentError(err)
	}
	identity, err := service.identities.ResolveMailboxDraftIdentityProof(ctx, mailboxDraftIdentityProofRequest{
		AccountID:  scope.accountID,
		IdentityID: input.FromIdentityID,
	})
	if err != nil {
		return mailboxDraftCreatePrepared{}, mailboxDraftCreateRepositoryError(err)
	}
	if identity.accountID != scope.accountID || identity.identityID != input.FromIdentityID {
		return mailboxDraftCreatePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	attachmentIDs := make([]uuid.UUID, len(input.Attachments))
	for index, reference := range input.Attachments {
		attachmentIDs[index] = reference.AttachmentID
	}
	attachmentProof, err := service.attachments.ResolveMailboxDraftAttachmentProof(ctx, mailboxDraftAttachmentProofRequest{
		AccountID:     scope.accountID,
		DraftClientID: input.DraftClientID,
		DraftID:       draftID,
		AttachmentIDs: attachmentIDs,
	})
	if err != nil {
		return mailboxDraftCreatePrepared{}, mailboxDraftCreateRepositoryError(err)
	}
	if !mailboxDraftCreateAttachmentProofMatches(attachmentProof, scope.accountID, input.DraftClientID, draftID, attachmentIDs) {
		return mailboxDraftCreatePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	canonical, err := CanonicalizeMailboxDraftRequest(MailboxDraftRequestHashInput{
		AccountID:        scope.accountID,
		DraftClientID:    input.DraftClientID,
		FromIdentityID:   input.FromIdentityID,
		From:             identity.from,
		To:               input.To,
		CC:               input.CC,
		BCC:              input.BCC,
		ReplyTo:          input.ReplyTo,
		Subject:          input.Subject,
		InReplyTo:        input.InReplyTo,
		References:       input.References,
		Priority:         input.Priority,
		ReplyToMessageID: input.ReplyToMessageID,
		ForwardMessageID: input.ForwardMessageID,
		Attachments:      attachmentProof.attachments,
		Body:             content,
	})
	if err != nil {
		return mailboxDraftCreatePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	}

	state := MailboxDraftEffectiveState{
		FromIdentityID:   canonical.FromIdentityID,
		ReplyToMessageID: mailboxDraftGetUUIDPointerCopy(canonical.ReplyToMessageID),
		ForwardMessageID: mailboxDraftGetUUIDPointerCopy(canonical.ForwardMessageID),
		AttachmentIDs:    mailboxDraftCreateSortedAttachmentIDs(canonical.Attachments),
	}
	if !mailboxDraftGetEffectiveStateValid(state) {
		return mailboxDraftCreatePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return mailboxDraftCreatePrepared{
		scope:              scope,
		draftID:            draftID,
		idempotencyKeyHash: idempotencyKeyHash,
		canonical:          canonical,
		effectiveState:     state,
		hasAttachments:     len(canonical.Attachments) != 0,
	}, nil
}

func (service *MailboxDraftService) mailboxDraftCreateReplay(ctx context.Context, actorID uuid.UUID, replay MailboxDraftCreateReplayLookupResult) (MailboxDraftView, error) {
	if service == nil || ctx == nil || actorID == uuid.Nil || !replay.Found || replay.DraftID == uuid.Nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if replay.LastMutationID != nil {
		if *replay.LastMutationID == uuid.Nil {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	return service.GetMailboxDraft(ctx, actorID, replay.DraftID)
}

func (service *MailboxDraftService) mailboxDraftCreateWithoutAttachments(ctx context.Context, actorID uuid.UUID, prepared mailboxDraftCreatePrepared, reservation MailboxDraftReservation) (MailboxDraftView, error) {
	draft, ok := mailboxDraftCreateEngineDraft(prepared, reservation)
	if !ok || len(draft.Message.Attachments) != 0 {
		return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, reservation)
	}
	message, err := service.engine.SaveDraft(ctx, draft)
	if err != nil {
		return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, reservation)
	}
	applied, ok := mailboxDraftCreateApplied(message, prepared, reservation)
	if !ok {
		return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, reservation)
	}
	return service.mailboxDraftCreateFinishApplied(ctx, actorID, prepared, reservation, applied)
}

func (service *MailboxDraftService) mailboxDraftCreateWithAttachments(ctx context.Context, actorID uuid.UUID, prepared mailboxDraftCreatePrepared, bridge mailboxDraftAttachmentCreateEngine, reservation MailboxDraftReservation) (MailboxDraftView, error) {
	if bridge == nil {
		return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, reservation)
	}
	current := reservation
	for {
		if !mailboxDraftCreateDispatchReservationValid(current, prepared, true) {
			return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, current)
		}
		base, ok := mailboxDraftCreateEngineDraft(prepared, current)
		if !ok {
			return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, current)
		}
		materialized, err := service.materializer.MaterializeMailboxDraftAttachments(ctx, mailboxDraftAttachmentMaterializationRequest{
			ActorID:         prepared.scope.actorID,
			OrganizationID:  prepared.scope.organizationID,
			AccountID:       prepared.scope.accountID,
			EngineAccountID: prepared.scope.engineAccountID,
			Fence:           *current.Fence,
			DraftID:         current.Draft.ID,
			DraftClientID:   prepared.canonical.DraftClientID,
			Revision:        current.Work.Revision,
			MutationID:      current.Work.MutationID,
			WorkKind:        MailboxDraftWorkCreate,
			EffectiveState:  mailboxDraftCreateEffectiveStateCopy(current.Work.EffectiveState),
		})
		if err != nil || materialized.capability == nil {
			return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, current)
		}
		draft, ok := mailboxDraftAttachmentCreateDraft(base, materialized.capability)
		if !ok {
			return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, current)
		}
		upload, err := bridge.uploadMailboxDraftCreate(ctx, draft, materialized.capability)
		if err != nil || !mailboxDraftAttachmentCreateUploadValid(upload) {
			return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, current)
		}

		advanced, err := service.repository.AdvanceMailboxDraftCreateToImport(ctx, MailboxDraftAdvanceCreateToImportRequest{
			Fence:      *current.Fence,
			DraftID:    current.Draft.ID,
			MutationID: current.Work.MutationID,
			BlobID:     upload.blobID,
			SizeBytes:  upload.sizeBytes,
		})
		if err != nil || !mailboxDraftCreateImportReservationValid(advanced, prepared, upload) {
			return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, current)
		}

		message, err := bridge.importMailboxDraftCreate(ctx, mailboxDraftAttachmentCreateImport{
			accountID:  prepared.scope.engineAccountID,
			draftKey:   advanced.Work.EngineDraftKey,
			mutationID: advanced.Work.MutationID.String(),
			blobID:     upload.blobID,
			sizeBytes:  upload.sizeBytes,
			receivedAt: advanced.Draft.UpdatedAt,
		})
		if err != nil {
			if mailboxDraftAttachmentCreatePersistedBlobNotFound(err) && advanced.Draft.PendingCreateUploadAttempt < 3 {
				restarted, restartErr := service.repository.RestartMailboxDraftCreateUploadAfterBlobNotFound(ctx, MailboxDraftRestartCreateUploadAfterBlobNotFoundRequest{
					Fence:             *advanced.Fence,
					ActorID:           prepared.scope.actorID,
					AccountID:         prepared.scope.accountID,
					DraftID:           advanced.Draft.ID,
					CurrentMutationID: advanced.Work.MutationID,
					ExpectedBlobID:    upload.blobID,
					ExpectedSizeBytes: upload.sizeBytes,
					NextMutationID:    uuid.New(),
				})
				if restartErr == nil && mailboxDraftCreateDispatchReservationValid(restarted.Reservation, prepared, true) {
					current = restarted.Reservation
					continue
				}
			}
			return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, advanced)
		}
		applied, ok := mailboxDraftCreateApplied(message, prepared, advanced)
		if !ok {
			return MailboxDraftView{}, service.mailboxDraftCreateUnknown(ctx, advanced)
		}
		return service.mailboxDraftCreateFinishApplied(ctx, actorID, prepared, advanced, applied)
	}
}

func (service *MailboxDraftService) mailboxDraftCreateFinishApplied(ctx context.Context, actorID uuid.UUID, prepared mailboxDraftCreatePrepared, reservation MailboxDraftReservation, applied MailboxDraftApplied) (MailboxDraftView, error) {
	finished, err := service.repository.FinishEngineCommand(ctx, MailboxDraftFinishRequest{
		Fence:      *reservation.Fence,
		DraftID:    reservation.Draft.ID,
		MutationID: reservation.Work.MutationID,
		Kind:       MailboxDraftMutationCreate,
		Outcome:    applied,
	})
	if err != nil || !mailboxDraftGetRecordValid(&finished.Draft, prepared.scope.accountID, reservation.Draft.ID) {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	return service.GetMailboxDraft(ctx, actorID, finished.Draft.ID)
}

// mailboxDraftCreateUnknown makes the current exact reservation terminally
// recoverable. If the durable finish itself is indeterminate, the client still
// gets outcome_unknown and no provider mutation is retried here.
func (service *MailboxDraftService) mailboxDraftCreateUnknown(ctx context.Context, reservation MailboxDraftReservation) error {
	if service == nil || ctx == nil || reservation.Fence == nil || reservation.Work == nil || reservation.Draft.ID == uuid.Nil || reservation.Work.MutationID == uuid.Nil || !mailboxEntEngineCommandFenceValid(*reservation.Fence) {
		return mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	return service.mailboxDraftCreateClaimedUnknown(ctx, *reservation.Fence, reservation.Draft.ID, reservation.Work.MutationID)
}

// mailboxDraftCreateClaimedUnknown is the strongest fail-closed terminal
// transition available after a valid command claim. The repository verifies
// the exact fence, draft and mutation before releasing the lease.
func (service *MailboxDraftService) mailboxDraftCreateClaimedUnknown(ctx context.Context, fence MailboxEngineCommandFence, draftID, mutationID uuid.UUID) error {
	if service == nil || ctx == nil || !mailboxEntEngineCommandFenceValid(fence) || draftID == uuid.Nil || mutationID == uuid.Nil {
		return mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	_, _ = service.repository.FinishEngineCommand(ctx, MailboxDraftFinishRequest{
		Fence:      fence,
		DraftID:    draftID,
		MutationID: mutationID,
		Kind:       MailboxDraftMutationCreate,
		Outcome:    MailboxDraftUnknown{},
	})
	return mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
}

func mailboxDraftCreateServiceValid(service *MailboxDraftService) bool {
	return service != nil &&
		!mailboxDraftServiceDependencyNil(service.owners) &&
		!mailboxDraftServiceDependencyNil(service.repository) &&
		!mailboxDraftServiceDependencyNil(service.engine) &&
		!mailboxDraftServiceDependencyNil(service.identities) &&
		!mailboxDraftServiceDependencyNil(service.attachments) &&
		!mailboxDraftServiceDependencyNil(service.materializer)
}

func mailboxDraftCreateScopeFromOwner(actorID uuid.UUID, owner MailboxOwner) (mailboxDraftCreateScope, error) {
	base, err := mailboxDraftGetScopeFromOwner(actorID, owner)
	if err != nil || owner.OrganizationID() == uuid.Nil {
		if err != nil {
			return mailboxDraftCreateScope{}, err
		}
		return mailboxDraftCreateScope{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	scope := mailboxDraftCreateScope{actorID: actorID, organizationID: owner.OrganizationID(), accountID: base.accountID, engineAccountID: base.engineAccountID}
	if !mailboxDraftCreateScopeValid(scope) {
		return mailboxDraftCreateScope{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return scope, nil
}

func mailboxDraftCreateScopeValid(scope mailboxDraftCreateScope) bool {
	return scope.actorID != uuid.Nil && scope.organizationID != uuid.Nil && scope.accountID != uuid.Nil && syncValidOpaque(scope.engineAccountID, MaxEngineOpaqueIDBytes)
}

func mailboxDraftCreateNormalizeIdempotencyKey(value string) (string, error) {
	if !utf8.ValidString(value) {
		return "", mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	normalized := norm.NFC.String(value)
	if len(normalized) < 1 || len(normalized) > mailboxDraftCreateMaxIdempotencyKeyBytes {
		return "", mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	for _, character := range normalized {
		if unicode.IsControl(character) {
			return "", mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
	}
	return strings.Clone(normalized), nil
}

func mailboxDraftCreateLocatorDraftID(locator MailboxDraftCreatePreHashLocatorResult) (uuid.UUID, error) {
	if locator.Found {
		if locator.DraftID == uuid.Nil {
			return uuid.Nil, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
		return locator.DraftID, nil
	}
	if locator.DraftID != uuid.Nil {
		return uuid.Nil, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	draftID := uuid.New()
	if draftID == uuid.Nil {
		return uuid.Nil, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
	return draftID, nil
}

func mailboxDraftCreateAttachmentProofMatches(proof mailboxDraftAttachmentProof, accountID, draftClientID, draftID uuid.UUID, ids []uuid.UUID) bool {
	if proof.accountID != accountID || proof.draftClientID != draftClientID || proof.draftID != draftID || len(proof.attachments) != len(ids) {
		return false
	}
	for index, attachment := range proof.attachments {
		if attachment.AttachmentID != ids[index] {
			return false
		}
	}
	return true
}

func mailboxDraftCreateSortedAttachmentIDs(attachments []MailboxDraftRequestHashAttachment) []uuid.UUID {
	ids := make([]uuid.UUID, len(attachments))
	for index, attachment := range attachments {
		ids[index] = attachment.AttachmentID
	}
	sort.Slice(ids, func(left, right int) bool { return syncLessUUID(ids[left], ids[right]) })
	return ids
}

func mailboxDraftCreateEffectiveStateCopy(state MailboxDraftEffectiveState) MailboxDraftEffectiveState {
	return MailboxDraftEffectiveState{
		FromIdentityID:   state.FromIdentityID,
		ReplyToMessageID: mailboxDraftGetUUIDPointerCopy(state.ReplyToMessageID),
		ForwardMessageID: mailboxDraftGetUUIDPointerCopy(state.ForwardMessageID),
		AttachmentIDs:    append([]uuid.UUID(nil), state.AttachmentIDs...),
	}
}

func mailboxDraftCreateBeginMatches(result MailboxEngineCommandBeginResult, scope mailboxDraftCreateScope) bool {
	if !mailboxDraftCreateScopeValid(scope) {
		return false
	}
	if result.Claimed {
		return result.EngineAccountID == scope.engineAccountID && result.Fence != nil && result.BusyUntil == nil && result.Fence.AccountID == scope.accountID && mailboxEntEngineCommandFenceValid(*result.Fence)
	}
	return result.EngineAccountID == "" && result.Fence == nil && result.BusyUntil != nil && mailboxEntValidTime(*result.BusyUntil)
}

func mailboxDraftCreateReplayReservationValid(reservation MailboxDraftReservation, prepared mailboxDraftCreatePrepared) bool {
	return reservation.Kind == MailboxDraftReservationReplay && reservation.Work == nil && reservation.Fence == nil &&
		reservation.Draft.ID == prepared.draftID && reservation.Draft.AccountID == prepared.scope.accountID &&
		reservation.Draft.DraftClientID == prepared.canonical.DraftClientID && reservation.Draft.Status == MailboxDraftStatusActive &&
		mailboxDraftGetRecordValid(&reservation.Draft, prepared.scope.accountID, prepared.draftID)
}

func mailboxDraftCreateDispatchReservationValid(reservation MailboxDraftReservation, prepared mailboxDraftCreatePrepared, hasAttachments bool) bool {
	if (reservation.Kind != MailboxDraftReservationDispatch && reservation.Kind != MailboxDraftReservationReconcileRequired) || reservation.Work == nil || reservation.Fence == nil || reservation.Draft.ID != prepared.draftID || reservation.Draft.AccountID != prepared.scope.accountID || reservation.Draft.DraftClientID != prepared.canonical.DraftClientID || reservation.Draft.Status != MailboxDraftStatusSaving || reservation.Draft.Revision != 1 || reservation.Draft.EngineMessageID != nil || reservation.Draft.EngineVersion != nil || reservation.Draft.PendingSave != nil || reservation.Draft.LastMutationID == nil || reservation.Draft.PendingMutationKind == nil || *reservation.Draft.PendingMutationKind != MailboxDraftMutationCreate || *reservation.Draft.LastMutationID != reservation.Work.MutationID || reservation.Work.Kind != MailboxDraftWorkCreate || reservation.Work.EngineDraftKey != reservation.Draft.EngineDraftKey || reservation.Work.Revision != reservation.Draft.Revision || reservation.Work.RequestHash != prepared.canonical.RequestHash || !mailboxDraftCreateEffectiveStatesEqual(reservation.Work.EffectiveState, prepared.effectiveState) || !mailboxDraftCreateEffectiveStatesEqual(reservation.Draft.Confirmed, prepared.effectiveState) || reservation.Fence.AccountID != prepared.scope.accountID || !mailboxEntEngineCommandFenceValid(*reservation.Fence) {
		return false
	}
	if hasAttachments {
		return reservation.Draft.PendingCreateAttachmentPhase != nil && *reservation.Draft.PendingCreateAttachmentPhase == MailboxDraftCreateAttachmentPhaseUpload && reservation.Draft.PendingCreateUploadBlobID == nil && reservation.Draft.PendingCreateUploadSizeBytes == nil && reservation.Draft.PendingCreateUploadAttempt >= 1 && reservation.Draft.PendingCreateUploadAttempt <= 3 && reservation.Work.PendingCreateAttachmentPhase != nil && *reservation.Work.PendingCreateAttachmentPhase == MailboxDraftCreateAttachmentPhaseUpload && reservation.Work.PendingCreateUploadBlobID == nil && reservation.Work.PendingCreateUploadSizeBytes == nil && reservation.Work.PendingCreateUploadAttempt == reservation.Draft.PendingCreateUploadAttempt && len(prepared.effectiveState.AttachmentIDs) != 0
	}
	return reservation.Draft.PendingCreateAttachmentPhase == nil && reservation.Draft.PendingCreateUploadBlobID == nil && reservation.Draft.PendingCreateUploadSizeBytes == nil && reservation.Draft.PendingCreateUploadAttempt == 0 && reservation.Work.PendingCreateAttachmentPhase == nil && reservation.Work.PendingCreateUploadBlobID == nil && reservation.Work.PendingCreateUploadSizeBytes == nil && reservation.Work.PendingCreateUploadAttempt == 0 && len(prepared.effectiveState.AttachmentIDs) == 0
}

func mailboxDraftCreateImportReservationValid(reservation MailboxDraftReservation, prepared mailboxDraftCreatePrepared, upload mailboxDraftAttachmentCreateUpload) bool {
	if reservation.Kind != MailboxDraftReservationDispatch || reservation.Work == nil || reservation.Fence == nil || reservation.Draft.ID != prepared.draftID || reservation.Draft.AccountID != prepared.scope.accountID || reservation.Draft.DraftClientID != prepared.canonical.DraftClientID || reservation.Draft.Status != MailboxDraftStatusSaving || reservation.Draft.Revision != 1 || reservation.Draft.EngineMessageID != nil || reservation.Draft.EngineVersion != nil || reservation.Draft.PendingSave != nil || reservation.Draft.LastMutationID == nil || reservation.Draft.PendingMutationKind == nil || *reservation.Draft.PendingMutationKind != MailboxDraftMutationCreate || *reservation.Draft.LastMutationID != reservation.Work.MutationID || reservation.Work.Kind != MailboxDraftWorkCreate || reservation.Work.EngineDraftKey != reservation.Draft.EngineDraftKey || reservation.Work.Revision != reservation.Draft.Revision || reservation.Work.RequestHash != prepared.canonical.RequestHash || !mailboxDraftCreateEffectiveStatesEqual(reservation.Work.EffectiveState, prepared.effectiveState) || !mailboxDraftCreateEffectiveStatesEqual(reservation.Draft.Confirmed, prepared.effectiveState) || reservation.Fence.AccountID != prepared.scope.accountID || !mailboxEntEngineCommandFenceValid(*reservation.Fence) || reservation.Draft.PendingCreateAttachmentPhase == nil || *reservation.Draft.PendingCreateAttachmentPhase != MailboxDraftCreateAttachmentPhaseImport || reservation.Draft.PendingCreateUploadBlobID == nil || *reservation.Draft.PendingCreateUploadBlobID != upload.blobID || reservation.Draft.PendingCreateUploadSizeBytes == nil || *reservation.Draft.PendingCreateUploadSizeBytes != upload.sizeBytes || reservation.Work.PendingCreateAttachmentPhase == nil || *reservation.Work.PendingCreateAttachmentPhase != MailboxDraftCreateAttachmentPhaseImport || reservation.Work.PendingCreateUploadBlobID == nil || *reservation.Work.PendingCreateUploadBlobID != upload.blobID || reservation.Work.PendingCreateUploadSizeBytes == nil || *reservation.Work.PendingCreateUploadSizeBytes != upload.sizeBytes || !mailboxDraftAttachmentCreateUploadValid(upload) || !mailboxDraftAttachmentCreateReceivedAtValid(reservation.Draft.UpdatedAt) {
		return false
	}
	return reservation.Draft.PendingCreateUploadAttempt >= 1 && reservation.Draft.PendingCreateUploadAttempt <= 3 && reservation.Work.PendingCreateUploadAttempt == reservation.Draft.PendingCreateUploadAttempt && len(prepared.effectiveState.AttachmentIDs) != 0
}

func mailboxDraftCreateEngineDraft(prepared mailboxDraftCreatePrepared, reservation MailboxDraftReservation) (EngineDraft, bool) {
	if reservation.Work == nil || reservation.Fence == nil || reservation.Work.MutationID == uuid.Nil || reservation.Work.EngineDraftKey == "" || reservation.Work.EngineMessageID != nil || reservation.Work.EngineVersion != nil {
		return EngineDraft{}, false
	}
	content := EngineMessageContent{
		From:       mailboxDraftGetAddressCopy(prepared.canonical.From),
		To:         mailboxDraftGetAddressesCopy(prepared.canonical.To),
		CC:         mailboxDraftGetAddressesCopy(prepared.canonical.CC),
		BCC:        mailboxDraftGetAddressesCopy(prepared.canonical.BCC),
		ReplyTo:    mailboxDraftGetAddressesCopy(prepared.canonical.ReplyTo),
		Subject:    strings.Clone(prepared.canonical.Subject),
		InReplyTo:  strings.Clone(prepared.canonical.InReplyTo),
		References: mailboxDraftGetStringsCopy(prepared.canonical.References),
		Priority:   prepared.canonical.Priority,
	}
	switch prepared.canonical.Body.SourceFormat {
	case MailboxContentFormatPlain:
		content.Format = EngineBodyPlain
		content.PlainBody = strings.Clone(prepared.canonical.Body.PlainBody)
	case MailboxContentFormatRich:
		content.Format = EngineBodyRich
		content.PlainBody = strings.Clone(prepared.canonical.Body.PlainBody)
		content.RichBody = strings.Clone(prepared.canonical.Body.SanitizedHTML)
	default:
		return EngineDraft{}, false
	}
	draft := EngineDraft{
		AccountID:       prepared.scope.engineAccountID,
		DraftKey:        strings.Clone(reservation.Work.EngineDraftKey),
		MutationID:      reservation.Work.MutationID.String(),
		MaxContentBytes: MaxEngineMessageBytes,
		Message:         content,
	}
	if validateDraftRequest(draft) != nil {
		return EngineDraft{}, false
	}
	return draft, true
}

func mailboxDraftCreateApplied(message EngineMessage, prepared mailboxDraftCreatePrepared, reservation MailboxDraftReservation) (MailboxDraftApplied, bool) {
	if reservation.Work == nil || validateEngineMessage(message) != nil || message.Projection.AccountID != prepared.scope.engineAccountID || !message.Projection.Flags.Draft || message.Projection.Origin != EngineProjectionOriginConfirmedOwnerDraft || !syncValidOpaque(message.Projection.MessageID, MaxEngineOpaqueIDBytes) || !syncValidOpaque(message.Projection.Version, MaxEngineOpaqueIDBytes) {
		return MailboxDraftApplied{}, false
	}
	return MailboxDraftApplied{EngineMessageID: message.Projection.MessageID, EngineVersion: message.Projection.Version}, true
}

func mailboxDraftCreateEffectiveStatesEqual(left, right MailboxDraftEffectiveState) bool {
	if left.FromIdentityID != right.FromIdentityID || !mailboxDraftCreateUUIDPointersEqual(left.ReplyToMessageID, right.ReplyToMessageID) || !mailboxDraftCreateUUIDPointersEqual(left.ForwardMessageID, right.ForwardMessageID) || len(left.AttachmentIDs) != len(right.AttachmentIDs) {
		return false
	}
	for index := range left.AttachmentIDs {
		if left.AttachmentIDs[index] != right.AttachmentIDs[index] {
			return false
		}
	}
	return true
}

func mailboxDraftCreateUUIDPointersEqual(left, right *uuid.UUID) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func mailboxDraftCreateHashPresent(value [sha256.Size]byte) bool {
	var zero [sha256.Size]byte
	return value != zero
}

func mailboxDraftCreateContentError(err error) error {
	switch {
	case errors.Is(err, ErrMailboxContentTooLarge):
		return mailboxDraftServiceError(MailboxDraftServiceErrorPayloadTooLarge)
	case errors.Is(err, ErrMailboxContentInvalid):
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	default:
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
}

func mailboxDraftCreateRepositoryError(err error) error {
	var repositoryError *MailboxDraftRepositoryError
	if !errors.As(err, &repositoryError) || repositoryError == nil {
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
	switch repositoryError.Kind() {
	case MailboxDraftRepositoryErrorInvalid:
		return mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	case MailboxDraftRepositoryErrorNotFound:
		return mailboxDraftServiceError(MailboxDraftServiceErrorNotFound)
	case MailboxDraftRepositoryErrorAccountDeleting:
		return mailboxDraftServiceError(MailboxDraftServiceErrorDeleting)
	case MailboxDraftRepositoryErrorStateConflict, MailboxDraftRepositoryErrorFenceLost:
		return mailboxDraftServiceError(MailboxDraftServiceErrorConflict)
	case MailboxDraftRepositoryErrorIdempotencyConflict:
		return mailboxDraftServiceError(MailboxDraftServiceErrorIdempotencyConflict)
	case MailboxDraftRepositoryErrorAttachmentConflict:
		return mailboxDraftServiceError(MailboxDraftServiceErrorAttachmentConflict)
	case MailboxDraftRepositoryErrorAttachmentLimit:
		return mailboxDraftServiceError(MailboxDraftServiceErrorAttachmentLimit)
	case MailboxDraftRepositoryErrorUnavailable:
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	default:
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
}
