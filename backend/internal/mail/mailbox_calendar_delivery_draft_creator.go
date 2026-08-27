package mail

import (
	"context"
	"crypto/sha256"
	"errors"

	"github.com/google/uuid"
)

// mailboxCalendarDeliveryDraftCreator is the dedicated Calendar draft path.
// It accepts only a sealed calendarMIMEPart and never delegates to the public
// generic create, save, attachment, reply, or forward paths.
type mailboxCalendarDeliveryDraftCreator struct {
	owners     mailboxOwnerResolver
	identities MailboxDraftIdentityProofResolver
	locator    mailboxCalendarDeliveryDraftLocator
	calendar   mailboxDraftCalendarCreateRepository
	engine     mailboxDraftCalendarCreateEngine
	reader     mailboxCalendarDeliveryDraftReader
}

// mailboxCalendarDeliveryDraftLocator is the narrow durable capability needed
// before Calendar MIME reservation. Calendar transitions remain separate so a
// caller cannot accidentally enter the generic attachment state machine.
type mailboxCalendarDeliveryDraftLocator interface {
	BeginEngineCommand(context.Context, MailboxEngineCommandBeginRequest) (MailboxEngineCommandBeginResult, error)
	LookupMailboxDraftCreatePreHashLocator(context.Context, MailboxDraftCreatePreHashLocatorRequest) (MailboxDraftCreatePreHashLocatorResult, error)
	LookupMailboxDraftCreateReplay(context.Context, MailboxDraftCreateReplayLookupRequest) (MailboxDraftCreateReplayLookupResult, error)
}

// mailboxCalendarDeliveryDraftReader preserves the existing owner-bound read
// boundary after a durable replay or successful Calendar import.
type mailboxCalendarDeliveryDraftReader interface {
	GetMailboxDraft(context.Context, uuid.UUID, uuid.UUID) (MailboxDraftView, error)
}

// newMailboxCalendarDeliveryDraftCreator constructs the private production
// Calendar creator. Its dependencies are deliberately the existing owner,
// identity, locator, Calendar repository, Calendar engine, and owner-bound
// reader capabilities; construction performs no I/O.
func newMailboxCalendarDeliveryDraftCreator(
	owners mailboxOwnerResolver,
	identities MailboxDraftIdentityProofResolver,
	locator mailboxCalendarDeliveryDraftLocator,
	calendar mailboxDraftCalendarCreateRepository,
	engine mailboxDraftCalendarCreateEngine,
	reader mailboxCalendarDeliveryDraftReader,
) (*mailboxCalendarDeliveryDraftCreator, error) {
	if mailboxDraftServiceDependencyNil(owners) ||
		mailboxDraftServiceDependencyNil(identities) ||
		mailboxDraftServiceDependencyNil(locator) ||
		mailboxDraftServiceDependencyNil(calendar) ||
		mailboxDraftServiceDependencyNil(engine) ||
		mailboxDraftServiceDependencyNil(reader) {
		return nil, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return &mailboxCalendarDeliveryDraftCreator{
		owners:     owners,
		identities: identities,
		locator:    locator,
		calendar:   calendar,
		engine:     engine,
		reader:     reader,
	}, nil
}

// CreateMailboxDraft implements the intentionally narrow Calendar delivery
// draft capability. The unexported calendar field makes a sealed
// calendarMIMEPart the only MIME input that can reach the Calendar engine.
func (creator *mailboxCalendarDeliveryDraftCreator) CreateMailboxDraft(ctx context.Context, actorID uuid.UUID, input MailboxDraftCreateInput) (MailboxDraftView, error) {
	if !mailboxCalendarDeliveryDraftCreatorValid(creator) || ctx == nil || actorID == uuid.Nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if ctx.Err() != nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
	if !mailboxCalendarDeliveryDraftInputSealed(input) {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	owner, err := creator.owners.ResolveOwner(ctx, actorID)
	if err != nil {
		return MailboxDraftView{}, mailboxDraftGetOwnerError(err)
	}
	scope, err := mailboxDraftCreateScopeFromOwner(actorID, owner)
	if err != nil {
		return MailboxDraftView{}, err
	}
	if input.calendar.organization() != scope.organizationID {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	normalizedKey, err := mailboxDraftCreateNormalizeIdempotencyKey(input.IdempotencyKey)
	if err != nil {
		return MailboxDraftView{}, err
	}
	idempotencyKeyHash := sha256.Sum256([]byte(normalizedKey))

	locator, err := creator.locator.LookupMailboxDraftCreatePreHashLocator(ctx, MailboxDraftCreatePreHashLocatorRequest{
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

	prepared, err := creator.prepare(ctx, owner, scope, draftID, idempotencyKeyHash, input)
	if err != nil {
		return MailboxDraftView{}, err
	}

	replay, err := creator.locator.LookupMailboxDraftCreateReplay(ctx, MailboxDraftCreateReplayLookupRequest{
		AccountID:                prepared.scope.accountID,
		DraftClientID:            prepared.canonical.DraftClientID,
		CreateIdempotencyKeyHash: prepared.idempotencyKeyHash,
		CreateRequestHash:        prepared.canonical.RequestHash,
	})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftCreateRepositoryError(err)
	}
	if replay.Found {
		return creator.replay(ctx, actorID, input.DraftClientID, replay)
	}
	if replay.DraftID != uuid.Nil || replay.LastMutationID != nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	// Recheck under the same pre-hash locator immediately before acquiring the
	// account fence. A competing create is an exact replay or a closed conflict;
	// it never reaches upload or import.
	locator, err = creator.locator.LookupMailboxDraftCreatePreHashLocator(ctx, MailboxDraftCreatePreHashLocatorRequest{
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
		replay, err = creator.locator.LookupMailboxDraftCreateReplay(ctx, MailboxDraftCreateReplayLookupRequest{
			AccountID:                prepared.scope.accountID,
			DraftClientID:            prepared.canonical.DraftClientID,
			CreateIdempotencyKeyHash: prepared.idempotencyKeyHash,
			CreateRequestHash:        prepared.canonical.RequestHash,
		})
		if err != nil {
			return MailboxDraftView{}, mailboxDraftCreateRepositoryError(err)
		}
		if replay.Found {
			return creator.replay(ctx, actorID, input.DraftClientID, replay)
		}
		if replay.DraftID != uuid.Nil || replay.LastMutationID != nil {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorIdempotencyConflict)
	}
	if locator.DraftID != uuid.Nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	begin, err := creator.locator.BeginEngineCommand(ctx, MailboxEngineCommandBeginRequest{AccountID: prepared.scope.accountID})
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
	bindingDigest, ok := mailboxCalendarDeliveryDraftBindingDigest(prepared, input.calendar, mutationID)
	if !ok {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	reservation, err := creator.calendar.ReserveMailboxDraftCalendarCreate(ctx, mailboxDraftCalendarCreateReserveRequest{
		Create: MailboxDraftCreateRequest{
			Fence:                    *begin.Fence,
			DraftID:                  prepared.draftID,
			DraftClientID:            prepared.canonical.DraftClientID,
			CreateIdempotencyKeyHash: prepared.idempotencyKeyHash,
			CreateRequestHash:        prepared.canonical.RequestHash,
			Confirmed:                mailboxDraftCreateEffectiveStateCopy(prepared.effectiveState),
			MutationID:               mutationID,
		},
		State: MailboxDraftCreateMIMEState{
			Kind:          MailboxDraftCreateMIMEKindCalendar,
			Phase:         MailboxDraftCreateMIMEPhaseUpload,
			BindingDigest: bindingDigest,
			Attempt:       1,
		},
	})
	if err != nil {
		return MailboxDraftView{}, mailboxDraftCreateRepositoryError(err)
	}

	switch reservation.Kind {
	case MailboxDraftReservationReplay:
		if !mailboxCalendarDeliveryDraftReplayReservationValid(reservation, prepared) {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
		}
		return creator.get(ctx, actorID, prepared.canonical.DraftClientID, reservation.Draft.ID)
	case MailboxDraftReservationReconcileRequired:
		// A prior upload/import outcome is already uncertain. This creator never
		// attempts an external reconciliation or an unfenced retry. A still-saving
		// exact state is first closed through its dedicated durable transition so
		// it does not retain the account lease until expiry.
		creator.closeReconcileRequired(ctx, prepared, input.calendar, reservation)
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	case MailboxDraftReservationDispatch:
		if !mailboxCalendarDeliveryDraftDispatchReservationValid(reservation, prepared, bindingDigest) {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
		}
		return creator.create(ctx, actorID, prepared, input.calendar, reservation)
	default:
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
}

func (creator *mailboxCalendarDeliveryDraftCreator) prepare(ctx context.Context, owner MailboxOwner, scope mailboxDraftCreateScope, draftID uuid.UUID, idempotencyKeyHash [sha256.Size]byte, input MailboxDraftCreateInput) (mailboxDraftCreatePrepared, error) {
	if creator == nil || ctx == nil || !mailboxDraftCreateScopeValid(scope) || draftID == uuid.Nil || !mailboxDraftCreateHashPresent(idempotencyKeyHash) || !mailboxCalendarDeliveryDraftInputSealed(input) {
		return mailboxDraftCreatePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	primaryAddress, hasPrimaryAddress := owner.PrimaryAddress()
	primary, err := NormalizeAddress(EngineAddress{Email: primaryAddress})
	if !hasPrimaryAddress || err != nil || primary.Email != primaryAddress {
		return mailboxDraftCreatePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	content, err := NormalizeMailboxContent(input.Content)
	if err != nil {
		return mailboxDraftCreatePrepared{}, mailboxDraftCreateContentError(err)
	}
	identity, err := creator.identities.ResolveMailboxDraftIdentityProof(ctx, mailboxDraftIdentityProofRequest{
		AccountID:  scope.accountID,
		IdentityID: input.FromIdentityID,
	})
	if err != nil {
		return mailboxDraftCreatePrepared{}, mailboxDraftCreateRepositoryError(err)
	}
	if identity.accountID != scope.accountID || identity.identityID != input.FromIdentityID || identity.from.Email != primary.Email {
		return mailboxDraftCreatePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}

	canonical, err := CanonicalizeMailboxDraftRequest(MailboxDraftRequestHashInput{
		AccountID:      scope.accountID,
		DraftClientID:  input.DraftClientID,
		FromIdentityID: input.FromIdentityID,
		From:           identity.from,
		To:             input.To,
		Subject:        input.Subject,
		Priority:       input.Priority,
		Body:           content,
	})
	if err != nil {
		return mailboxDraftCreatePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnprocessable)
	}
	requestHash, ok := mailboxDraftCreateRequestHashWithCalendar(canonical.RequestHash, input.calendar)
	if !ok {
		return mailboxDraftCreatePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	canonical.RequestHash = requestHash

	effectiveState := MailboxDraftEffectiveState{FromIdentityID: canonical.FromIdentityID, AttachmentIDs: []uuid.UUID{}}
	if !mailboxDraftGetEffectiveStateValid(effectiveState) || len(canonical.To) != 1 || len(canonical.CC) != 0 || len(canonical.BCC) != 0 || len(canonical.Attachments) != 0 {
		return mailboxDraftCreatePrepared{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	return mailboxDraftCreatePrepared{
		scope:              scope,
		draftID:            draftID,
		idempotencyKeyHash: idempotencyKeyHash,
		canonical:          canonical,
		effectiveState:     effectiveState,
		hasAttachments:     false,
	}, nil
}

func (creator *mailboxCalendarDeliveryDraftCreator) replay(ctx context.Context, actorID, draftClientID uuid.UUID, replay MailboxDraftCreateReplayLookupResult) (MailboxDraftView, error) {
	if creator == nil || ctx == nil || actorID == uuid.Nil || draftClientID == uuid.Nil || !replay.Found || replay.DraftID == uuid.Nil {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
	}
	if replay.LastMutationID != nil {
		if *replay.LastMutationID == uuid.Nil {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorInvalid)
		}
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	return creator.get(ctx, actorID, draftClientID, replay.DraftID)
}

func (creator *mailboxCalendarDeliveryDraftCreator) create(ctx context.Context, actorID uuid.UUID, prepared mailboxDraftCreatePrepared, calendar *calendarMIMEPart, reservation MailboxDraftReservation) (MailboxDraftView, error) {
	current := reservation
	for {
		bindingDigest, ok := mailboxCalendarDeliveryDraftBindingDigest(prepared, calendar, current.Work.MutationID)
		if !ok || !mailboxCalendarDeliveryDraftDispatchReservationValid(current, prepared, bindingDigest) {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
		}
		state := current.Draft.PendingCreateMIME
		if state == nil {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
		}

		base, ok := mailboxDraftCreateEngineDraft(prepared, current)
		if !ok || len(base.Message.Attachments) != 0 {
			return MailboxDraftView{}, creator.markUnknown(ctx, current, bindingDigest, mailboxDraftCalendarCreateUnknownStageUpload, "", 0)
		}
		upload, err := creator.engine.uploadCalendarDraftCreateMIME(ctx, base, calendar)
		if err != nil || !validMailboxDraftCalendarCreateUpload(upload) {
			return MailboxDraftView{}, creator.markUnknown(ctx, current, bindingDigest, mailboxDraftCalendarCreateUnknownStageUpload, "", 0)
		}

		advanced, err := creator.calendar.AdvanceMailboxDraftCalendarCreateToImport(ctx, mailboxDraftCalendarCreateAdvanceToImportRequest{
			Fence:                 *current.Fence,
			DraftID:               current.Draft.ID,
			MutationID:            current.Work.MutationID,
			ExpectedBindingDigest: bindingDigest,
			BlobID:                upload.blobID,
			SizeBytes:             upload.sizeBytes,
		})
		if err != nil || !mailboxCalendarDeliveryDraftImportReservationValid(advanced, prepared, bindingDigest, upload) {
			return MailboxDraftView{}, creator.markUnknown(ctx, current, bindingDigest, mailboxDraftCalendarCreateUnknownStageAdvance, upload.blobID, upload.sizeBytes)
		}

		message, err := creator.engine.importCalendarDraftCreateMIME(ctx, mailboxDraftCalendarCreateImport{
			accountID:  prepared.scope.engineAccountID,
			draftKey:   advanced.Work.EngineDraftKey,
			mutationID: advanced.Work.MutationID.String(),
			blobID:     upload.blobID,
			sizeBytes:  upload.sizeBytes,
			receivedAt: advanced.Draft.UpdatedAt,
		})
		if err != nil {
			if errors.Is(err, errMailboxDraftCalendarCreatePersistedBlobNotFound) && state.Attempt < 3 {
				nextMutationID := uuid.New()
				nextBindingDigest, nextOK := mailboxCalendarDeliveryDraftBindingDigest(prepared, calendar, nextMutationID)
				if !nextOK {
					return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
				}
				restarted, restartErr := creator.calendar.RestartMailboxDraftCalendarCreateUploadAfterBlobNotFound(ctx, mailboxDraftCalendarCreateRestartAfterBlobNotFoundRequest{
					Fence:                 *advanced.Fence,
					DraftID:               advanced.Draft.ID,
					ExpectedMutationID:    advanced.Work.MutationID,
					ExpectedAttempt:       state.Attempt,
					ExpectedBindingDigest: bindingDigest,
					ExpectedBlobID:        upload.blobID,
					ExpectedSizeBytes:     upload.sizeBytes,
					NextMutationID:        nextMutationID,
					NextBindingDigest:     nextBindingDigest,
					IdempotencyKeyHash:    prepared.idempotencyKeyHash,
					RequestHash:           prepared.canonical.RequestHash,
				})
				if restartErr == nil && mailboxCalendarDeliveryDraftDispatchReservationValid(restarted, prepared, nextBindingDigest) {
					current = restarted
					continue
				}
				return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
			}
			return MailboxDraftView{}, creator.markUnknown(ctx, advanced, bindingDigest, mailboxDraftCalendarCreateUnknownStageImport, upload.blobID, upload.sizeBytes)
		}
		applied, ok := mailboxDraftCreateApplied(message, prepared, advanced)
		if !ok {
			return MailboxDraftView{}, creator.markUnknown(ctx, advanced, bindingDigest, mailboxDraftCalendarCreateUnknownStageImport, upload.blobID, upload.sizeBytes)
		}

		finished, err := creator.calendar.FinishMailboxDraftCalendarCreateApplied(ctx, mailboxDraftCalendarCreateFinishAppliedRequest{
			Fence:                 *advanced.Fence,
			DraftID:               advanced.Draft.ID,
			MutationID:            advanced.Work.MutationID,
			ExpectedBindingDigest: bindingDigest,
			ExpectedBlobID:        upload.blobID,
			ExpectedSizeBytes:     upload.sizeBytes,
			Applied:               applied,
		})
		if err != nil || !mailboxCalendarDeliveryDraftFinishedRecordValid(finished, prepared, advanced.Work.MutationID) {
			return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
		}
		return creator.get(ctx, actorID, prepared.canonical.DraftClientID, finished.ID)
	}
}

func (creator *mailboxCalendarDeliveryDraftCreator) markUnknown(ctx context.Context, reservation MailboxDraftReservation, bindingDigest [sha256.Size]byte, stage mailboxDraftCalendarCreateUnknownStage, blobID string, sizeBytes int64) error {
	if creator == nil || ctx == nil || reservation.Fence == nil || reservation.Work == nil || reservation.Work.PendingCreateMIME == nil || reservation.Draft.ID == uuid.Nil || reservation.Work.MutationID == uuid.Nil {
		return mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
	}
	_, _ = creator.calendar.MarkMailboxDraftCalendarCreateUnknown(ctx, mailboxDraftCalendarCreateMarkUnknownRequest{
		Fence:                 *reservation.Fence,
		DraftID:               reservation.Draft.ID,
		MutationID:            reservation.Work.MutationID,
		ExpectedAttempt:       reservation.Work.PendingCreateMIME.Attempt,
		ExpectedBindingDigest: bindingDigest,
		Stage:                 stage,
		BlobID:                blobID,
		SizeBytes:             sizeBytes,
	})
	return mailboxDraftServiceError(MailboxDraftServiceErrorOutcomeUnknown)
}

func (creator *mailboxCalendarDeliveryDraftCreator) closeReconcileRequired(ctx context.Context, prepared mailboxDraftCreatePrepared, calendar *calendarMIMEPart, reservation MailboxDraftReservation) {
	if creator == nil || reservation.Draft.Status != MailboxDraftStatusSaving || reservation.Work == nil || reservation.Draft.PendingCreateMIME == nil {
		return
	}
	bindingDigest, ok := mailboxCalendarDeliveryDraftBindingDigest(prepared, calendar, reservation.Work.MutationID)
	if !ok {
		return
	}
	state := reservation.Draft.PendingCreateMIME
	switch state.Phase {
	case MailboxDraftCreateMIMEPhaseUpload:
		_ = creator.markUnknown(ctx, reservation, bindingDigest, mailboxDraftCalendarCreateUnknownStageUpload, "", 0)
	case MailboxDraftCreateMIMEPhaseImport:
		if state.BlobID != nil && state.SizeBytes != nil {
			_ = creator.markUnknown(ctx, reservation, bindingDigest, mailboxDraftCalendarCreateUnknownStageImport, *state.BlobID, *state.SizeBytes)
		}
	}
}

func (creator *mailboxCalendarDeliveryDraftCreator) get(ctx context.Context, actorID, draftClientID, draftID uuid.UUID) (MailboxDraftView, error) {
	view, err := creator.reader.GetMailboxDraft(ctx, actorID, draftID)
	if err != nil {
		return MailboxDraftView{}, mailboxCalendarDeliveryDraftSafeServiceError(err)
	}
	if view.ID != draftID || view.DraftClientID != draftClientID || view.Revision < 1 {
		return MailboxDraftView{}, mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
	return view, nil
}

func mailboxCalendarDeliveryDraftInputSealed(input MailboxDraftCreateInput) bool {
	if input.calendar == nil || !input.calendar.valid() || input.DraftClientID == uuid.Nil || input.FromIdentityID == uuid.Nil ||
		input.DraftClientID != calendarMailDeliveryDraftClientID(input.calendar.deliveryID()) ||
		input.IdempotencyKey != calendarMailDeliveryDraftKey(input.calendar.deliveryID()) ||
		len(input.To) != 1 || len(input.CC) != 0 || len(input.BCC) != 0 || len(input.ReplyTo) != 0 ||
		input.InReplyTo != "" || len(input.References) != 0 || input.ReplyToMessageID != nil || input.ForwardMessageID != nil || len(input.Attachments) != 0 {
		return false
	}
	recipient, err := NormalizeAddress(input.To[0])
	if err != nil {
		return false
	}
	return recipient == input.calendar.recipientAddress()
}

func mailboxCalendarDeliveryDraftBindingDigest(prepared mailboxDraftCreatePrepared, calendar *calendarMIMEPart, mutationID uuid.UUID) ([sha256.Size]byte, bool) {
	if !mailboxDraftCreateScopeValid(prepared.scope) || prepared.draftID == uuid.Nil || calendar == nil || !calendar.valid() {
		return [sha256.Size]byte{}, false
	}
	return mailboxDraftCreateMIMEBindingDigest(prepared.scope.organizationID, prepared.scope.accountID, prepared.draftID, mutationID, MailboxDraftCreateMIMEKindCalendar, prepared.canonical.RequestHash, calendar)
}

func mailboxCalendarDeliveryDraftReplayReservationValid(reservation MailboxDraftReservation, prepared mailboxDraftCreatePrepared) bool {
	return reservation.Kind == MailboxDraftReservationReplay && reservation.Work == nil && reservation.Fence == nil &&
		reservation.Draft.ID == prepared.draftID && reservation.Draft.AccountID == prepared.scope.accountID &&
		reservation.Draft.DraftClientID == prepared.canonical.DraftClientID && reservation.Draft.Status == MailboxDraftStatusActive &&
		len(reservation.Draft.Confirmed.AttachmentIDs) == 0 && mailboxDraftGetRecordValid(&reservation.Draft, prepared.scope.accountID, prepared.draftID)
}

func mailboxCalendarDeliveryDraftDispatchReservationValid(reservation MailboxDraftReservation, prepared mailboxDraftCreatePrepared, expectedBindingDigest [sha256.Size]byte) bool {
	if reservation.Kind != MailboxDraftReservationDispatch || reservation.Work == nil || reservation.Fence == nil ||
		reservation.Draft.ID != prepared.draftID || reservation.Draft.AccountID != prepared.scope.accountID ||
		reservation.Draft.DraftClientID != prepared.canonical.DraftClientID || reservation.Draft.Status != MailboxDraftStatusSaving ||
		reservation.Draft.Revision != 1 || reservation.Draft.EngineMessageID != nil || reservation.Draft.EngineVersion != nil ||
		reservation.Draft.PendingSave != nil || reservation.Draft.LastMutationID == nil || reservation.Draft.PendingMutationKind == nil ||
		*reservation.Draft.PendingMutationKind != MailboxDraftMutationCreate || *reservation.Draft.LastMutationID != reservation.Work.MutationID ||
		reservation.Work.Kind != MailboxDraftWorkCreate || reservation.Work.EngineDraftKey != reservation.Draft.EngineDraftKey ||
		reservation.Work.Revision != reservation.Draft.Revision || reservation.Work.RequestHash != prepared.canonical.RequestHash ||
		!mailboxDraftCreateEffectiveStatesEqual(reservation.Work.EffectiveState, prepared.effectiveState) ||
		!mailboxDraftCreateEffectiveStatesEqual(reservation.Draft.Confirmed, prepared.effectiveState) ||
		reservation.Fence.AccountID != prepared.scope.accountID || !mailboxEntEngineCommandFenceValid(*reservation.Fence) ||
		len(reservation.Draft.Confirmed.AttachmentIDs) != 0 || len(reservation.Work.EffectiveState.AttachmentIDs) != 0 {
		return false
	}
	state := reservation.Draft.PendingCreateMIME
	workState := reservation.Work.PendingCreateMIME
	return state != nil && workState != nil && mailboxDraftCreateMIMEStatesEqual(state, workState) &&
		state.Kind == MailboxDraftCreateMIMEKindCalendar && state.Phase == MailboxDraftCreateMIMEPhaseUpload &&
		state.Attempt >= 1 && state.Attempt <= 3 && state.BindingDigest == expectedBindingDigest &&
		state.BlobID == nil && state.SizeBytes == nil && state.RetryAt == nil && state.ExhaustedAt == nil &&
		validMailboxDraftCreateMIMEStateBridge(state, reservation.Draft.Status, reservation.Draft.PendingMutationKind,
			reservation.Draft.PendingCreateAttachmentPhase, reservation.Draft.PendingCreateUploadBlobID,
			reservation.Draft.PendingCreateUploadSizeBytes, reservation.Draft.PendingCreateUploadAttempt,
			reservation.Draft.PendingCreateUploadRetryAt, reservation.Draft.PendingCreateUploadExhaustedAt) &&
		validMailboxDraftCreateMIMEStateBridge(workState, reservation.Draft.Status, reservation.Draft.PendingMutationKind,
			reservation.Work.PendingCreateAttachmentPhase, reservation.Work.PendingCreateUploadBlobID,
			reservation.Work.PendingCreateUploadSizeBytes, reservation.Work.PendingCreateUploadAttempt,
			reservation.Work.PendingCreateUploadRetryAt, reservation.Work.PendingCreateUploadExhaustedAt)
}

func mailboxCalendarDeliveryDraftImportReservationValid(reservation MailboxDraftReservation, prepared mailboxDraftCreatePrepared, expectedBindingDigest [sha256.Size]byte, upload mailboxDraftCalendarCreateUpload) bool {
	if reservation.Kind != MailboxDraftReservationDispatch || reservation.Work == nil || reservation.Fence == nil ||
		reservation.Draft.ID != prepared.draftID || reservation.Draft.AccountID != prepared.scope.accountID ||
		reservation.Draft.DraftClientID != prepared.canonical.DraftClientID || reservation.Draft.Status != MailboxDraftStatusSaving ||
		reservation.Draft.Revision != 1 || reservation.Draft.EngineMessageID != nil || reservation.Draft.EngineVersion != nil ||
		reservation.Draft.PendingSave != nil || reservation.Draft.LastMutationID == nil || reservation.Draft.PendingMutationKind == nil ||
		*reservation.Draft.PendingMutationKind != MailboxDraftMutationCreate || *reservation.Draft.LastMutationID != reservation.Work.MutationID ||
		reservation.Work.Kind != MailboxDraftWorkCreate || reservation.Work.EngineDraftKey != reservation.Draft.EngineDraftKey ||
		reservation.Work.Revision != reservation.Draft.Revision || reservation.Work.RequestHash != prepared.canonical.RequestHash ||
		!mailboxDraftCreateEffectiveStatesEqual(reservation.Work.EffectiveState, prepared.effectiveState) ||
		!mailboxDraftCreateEffectiveStatesEqual(reservation.Draft.Confirmed, prepared.effectiveState) ||
		reservation.Fence.AccountID != prepared.scope.accountID || !mailboxEntEngineCommandFenceValid(*reservation.Fence) ||
		!validMailboxDraftCalendarCreateUpload(upload) || !mailboxDraftCalendarCreateReceivedAtValid(reservation.Draft.UpdatedAt) {
		return false
	}
	state := reservation.Draft.PendingCreateMIME
	workState := reservation.Work.PendingCreateMIME
	return state != nil && workState != nil && mailboxDraftCreateMIMEStatesEqual(state, workState) &&
		state.Kind == MailboxDraftCreateMIMEKindCalendar && state.Phase == MailboxDraftCreateMIMEPhaseImport &&
		state.Attempt >= 1 && state.Attempt <= 3 && state.BindingDigest == expectedBindingDigest &&
		state.BlobID != nil && *state.BlobID == upload.blobID && state.SizeBytes != nil && *state.SizeBytes == upload.sizeBytes &&
		state.RetryAt == nil && state.ExhaustedAt == nil &&
		validMailboxDraftCreateMIMEStateBridge(state, reservation.Draft.Status, reservation.Draft.PendingMutationKind,
			reservation.Draft.PendingCreateAttachmentPhase, reservation.Draft.PendingCreateUploadBlobID,
			reservation.Draft.PendingCreateUploadSizeBytes, reservation.Draft.PendingCreateUploadAttempt,
			reservation.Draft.PendingCreateUploadRetryAt, reservation.Draft.PendingCreateUploadExhaustedAt) &&
		validMailboxDraftCreateMIMEStateBridge(workState, reservation.Draft.Status, reservation.Draft.PendingMutationKind,
			reservation.Work.PendingCreateAttachmentPhase, reservation.Work.PendingCreateUploadBlobID,
			reservation.Work.PendingCreateUploadSizeBytes, reservation.Work.PendingCreateUploadAttempt,
			reservation.Work.PendingCreateUploadRetryAt, reservation.Work.PendingCreateUploadExhaustedAt)
}

func mailboxCalendarDeliveryDraftFinishedRecordValid(record MailboxDraftRecord, prepared mailboxDraftCreatePrepared, mutationID uuid.UUID) bool {
	return mutationID != uuid.Nil && record.ID == prepared.draftID && record.AccountID == prepared.scope.accountID &&
		record.DraftClientID == prepared.canonical.DraftClientID && record.ConfirmedMutationID != nil &&
		*record.ConfirmedMutationID == mutationID && len(record.Confirmed.AttachmentIDs) == 0 &&
		mailboxDraftGetRecordValid(&record, prepared.scope.accountID, prepared.draftID)
}

func mailboxCalendarDeliveryDraftCreatorValid(creator *mailboxCalendarDeliveryDraftCreator) bool {
	return creator != nil && !mailboxDraftServiceDependencyNil(creator.owners) &&
		!mailboxDraftServiceDependencyNil(creator.identities) && !mailboxDraftServiceDependencyNil(creator.locator) &&
		!mailboxDraftServiceDependencyNil(creator.calendar) && !mailboxDraftServiceDependencyNil(creator.engine) &&
		!mailboxDraftServiceDependencyNil(creator.reader)
}

func mailboxCalendarDeliveryDraftSafeServiceError(err error) error {
	kind, ok := MailboxDraftServiceErrorKindOf(err)
	if !ok {
		return mailboxDraftServiceError(MailboxDraftServiceErrorUnavailable)
	}
	return mailboxDraftServiceError(kind)
}
