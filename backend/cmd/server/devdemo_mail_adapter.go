//go:build !qa && !release

package main

import (
	"context"
	"errors"
	"log/slog"
	"reflect"
	"sync"
	"time"

	"github.com/google/uuid"

	appent "github.com/valerii/hub/backend/ent"
	calendarmod "github.com/valerii/hub/backend/internal/calendar"
	"github.com/valerii/hub/backend/internal/devdemo"
	mailmod "github.com/valerii/hub/backend/internal/mail"
)

const (
	devMailDemoDatasetVersion = 2
	devMailDemoObserveTimeout = 2 * time.Minute
	devMailDemoObservePoll    = 5 * time.Second

	devMailDemoInvitationTitle       = "Mail invitation"
	devMailDemoInvitationDescription = "PM invitation for the dev1 Mail demo"
	devMailDemoInvitationLocation    = "Mail demo"
	devMailDemoInvitationTimezone    = "Europe/Moscow"
)

var (
	errDevMailDemoInvalid     = errors.New("dev mail demo adapter input is invalid")
	errDevMailDemoUnavailable = errors.New("dev mail demo adapter is unavailable")
	errDevMailDemoPending     = errors.New("dev mail demo projection is pending")
)

var _ devdemo.MailDemoPort = (*devMailDemoAdapter)(nil)

type devMailDemoPrepareStage string

const (
	devMailDemoPrepareStageControlPlane devMailDemoPrepareStage = "control-plane"
	devMailDemoPrepareStageSource       devMailDemoPrepareStage = "source"
	devMailDemoPrepareStageRuntime      devMailDemoPrepareStage = "runtime"
)

// devMailDemoPrepareError exposes only the closed Mail control-plane enum
// values; provider and database errors never become part of the dev endpoint
// response.
type devMailDemoPrepareError struct {
	stage          devMailDemoPrepareStage
	controlStage   devMailControlStage
	controlFailure devMailControlFailure
}

func (err devMailDemoPrepareError) Error() string {
	switch err.stage {
	case devMailDemoPrepareStageControlPlane:
		return "dev mail demo control-plane preparation failed: stage=" + devMailDemoControlStageText(err.controlStage) + " failure=" + devMailDemoControlFailureText(err.controlFailure)
	case devMailDemoPrepareStageSource:
		return "dev mail demo source preparation failed"
	case devMailDemoPrepareStageRuntime:
		return "dev mail demo runtime preparation failed"
	default:
		return "dev mail demo preparation failed"
	}
}

func (err devMailDemoPrepareError) Unwrap() error {
	return errDevMailDemoUnavailable
}

func devMailDemoControlStageText(stage devMailControlStage) string {
	switch stage {
	case devMailControlStageInput:
		return "input"
	case devMailControlStageProvider:
		return "provider"
	case devMailControlStageUsers:
		return "users"
	case devMailControlStageDomain:
		return "domain"
	case devMailControlStagePolicy:
		return "policy"
	case devMailControlStageAccount:
		return "account"
	case devMailControlStageIdentity:
		return "identity"
	case devMailControlStageSettings:
		return "settings"
	case devMailControlStageCommit:
		return "commit"
	case devMailControlStageResult:
		return "result"
	default:
		return "unknown"
	}
}

func devMailDemoControlFailureText(failure devMailControlFailure) string {
	switch failure {
	case devMailControlFailureInvalid:
		return "invalid"
	case devMailControlFailureConflict:
		return "conflict"
	case devMailControlFailureUnavailable:
		return "unavailable"
	default:
		return "unknown"
	}
}

// devMailDemoControlDetailText is a deliberate allow-list for the small,
// value-free domain conflict classifier. No provider or database data may
// cross the adapter boundary.
// devMailDemoPrepared is private immutable first-load state. It contains no
// caller-owned scope map or provider/message payload.
type devMailDemoPrepared struct {
	organizationID uuid.UUID
	pmUserID       uuid.UUID
	dev1UserID     uuid.UUID
	control        devMailControlPlaneResult
	source         devMailSourceResult
}

// devMailDemoAdapter is the dev-only bridge between the seeded Mail source
// and the production-owned Mail/Calendar projection workers. It only observes
// public read services after first-load; it never advances a worker itself.
type devMailDemoAdapter struct {
	client        *appent.Client
	engine        *mailmod.FakeMailEngine
	mailboxSvc    *mailmod.MailboxService
	mailboxRepo   *mailmod.MailboxEntRepository
	messageSvc    *mailmod.MailboxMessageService
	eventReadSvc  *calendarmod.EventReadService
	startRuntime  func() error
	projectionSvc *mailmod.MailboxFolderProjectionService
	now           func() time.Time

	guard             sync.Mutex
	prepared          *devMailDemoPrepared
	observedPMEventID uuid.UUID
}

// newDevMailDemoAdapter constructs the server-only Mail demo port. The
// projection service is intentionally built here so the observer uses the
// same owner-gated public read boundary as HTTP.
func newDevMailDemoAdapter(
	client *appent.Client,
	engine *mailmod.FakeMailEngine,
	mailboxSvc *mailmod.MailboxService,
	mailboxRepo *mailmod.MailboxEntRepository,
	eventReadSvc *calendarmod.EventReadService,
	now func() time.Time,
	startRuntime func() error,
) (*devMailDemoAdapter, error) {
	if devMailDemoNil(client) || devMailDemoNil(engine) || devMailDemoNil(mailboxSvc) || devMailDemoNil(mailboxRepo) || devMailDemoNil(eventReadSvc) || now == nil || startRuntime == nil || !validDevMailControlNow(now()) {
		return nil, errDevMailDemoInvalid
	}
	projectionSvc, err := mailmod.NewMailboxFolderProjectionService(mailboxSvc, mailboxRepo)
	if err != nil || devMailDemoNil(projectionSvc) {
		return nil, errDevMailDemoInvalid
	}
	messageSvc, err := mailmod.NewMailboxMessageService(mailboxSvc, mailboxRepo, engine, mailboxRepo)
	if err != nil || devMailDemoNil(messageSvc) {
		return nil, errDevMailDemoInvalid
	}
	return &devMailDemoAdapter{
		client:        client,
		engine:        engine,
		mailboxSvc:    mailboxSvc,
		mailboxRepo:   mailboxRepo,
		messageSvc:    messageSvc,
		eventReadSvc:  eventReadSvc,
		startRuntime:  startRuntime,
		projectionSvc: projectionSvc,
		now:           now,
	}, nil
}

// PrepareMailDemo provisions the provider and immutable static dataset once.
// Later calls must use the same organization and Mail demo identities; they do
// not reload the source and therefore preserve interactive mutations.
func (adapter *devMailDemoAdapter) PrepareMailDemo(ctx context.Context, input devdemo.MailDemoPrepareInput) error {
	if adapter == nil || !adapter.ready() || !validDevMailControlContext(ctx) || input.OrganizationID == uuid.Nil {
		return errDevMailDemoInvalid
	}
	pmUserID, dev1UserID, ok := devMailDemoScopeUsers(input.Scope.UserIDs)
	if !ok {
		return errDevMailDemoInvalid
	}
	if err := ctx.Err(); err != nil {
		return errDevMailDemoUnavailable
	}

	adapter.guard.Lock()
	defer adapter.guard.Unlock()

	if adapter.prepared != nil {
		if adapter.prepared.matches(input.OrganizationID, pmUserID, dev1UserID) {
			if err := adapter.startRuntime(); err != nil {
				return devMailDemoPrepareError{stage: devMailDemoPrepareStageRuntime}
			}
			return nil
		}
		return errDevMailDemoInvalid
	}
	now := adapter.now()
	if !validDevMailControlNow(now) {
		return errDevMailDemoInvalid
	}
	users := map[string]uuid.UUID{
		devMailSourcePMEmail:   pmUserID,
		devMailSourceDev1Email: dev1UserID,
	}
	control, err := reconcileDevMailControlPlane(ctx, adapter.client, adapter.engine, input.OrganizationID, users, now)
	if err != nil {
		prepareErr := devMailDemoPrepareError{stage: devMailDemoPrepareStageControlPlane}
		var controlErr *devMailControlError
		if errors.As(err, &controlErr) && controlErr != nil {
			prepareErr.controlStage = controlErr.stage
			prepareErr.controlFailure = controlErr.failure
		}
		return prepareErr
	}
	if !validDevMailControlResult(control) {
		return devMailDemoPrepareError{
			stage:          devMailDemoPrepareStageControlPlane,
			controlStage:   devMailControlStageResult,
			controlFailure: devMailControlFailureUnavailable,
		}
	}
	source, err := seedDevMailSource(ctx, adapter.engine, control)
	if err != nil {
		return devMailDemoPrepareError{stage: devMailDemoPrepareStageSource}
	}
	if !validDevMailSourceResult(source) {
		return devMailDemoPrepareError{stage: devMailDemoPrepareStageSource}
	}
	adapter.prepared = &devMailDemoPrepared{
		organizationID: input.OrganizationID,
		pmUserID:       pmUserID,
		dev1UserID:     dev1UserID,
		control:        cloneDevMailControlResult(control),
		source:         cloneDevMailSourceResult(source),
	}
	if err := adapter.startRuntime(); err != nil {
		return devMailDemoPrepareError{stage: devMailDemoPrepareStageRuntime}
	}
	return nil
}

// ReconcileMailDemo observes the production projection path until the seeded
// Calendar invitation has an outbound PM message and an independent inbound
// dev1 receipt. Occurrence work remains bound to the dev1 owner event, not
// the PM master event. It deliberately does not call worker step/run APIs.
func (adapter *devMailDemoAdapter) ReconcileMailDemo(ctx context.Context, input devdemo.MailDemoInput) (devdemo.MailDemoResult, error) {
	if adapter == nil || !adapter.ready() || !validDevMailControlContext(ctx) || input.CalendarEvent.Key != devdemo.MailInvitationEventKey || input.CalendarEvent.EventID == uuid.Nil {
		return devdemo.MailDemoResult{}, errDevMailDemoInvalid
	}
	if err := ctx.Err(); err != nil {
		return devdemo.MailDemoResult{}, err
	}
	pmUserID, dev1UserID, ok := devMailDemoScopeUsers(input.Scope.UserIDs)
	if !ok {
		return devdemo.MailDemoResult{}, errDevMailDemoInvalid
	}

	adapter.guard.Lock()
	defer adapter.guard.Unlock()
	if adapter.prepared == nil || !adapter.prepared.matches(adapter.prepared.organizationID, pmUserID, dev1UserID) {
		return devdemo.MailDemoResult{}, errDevMailDemoInvalid
	}
	if err := adapter.recoverCalendarInviteBeforeObserve(ctx, *adapter.prepared, input.CalendarEvent.EventID); err != nil {
		return devdemo.MailDemoResult{}, err
	}
	dev1EventID, err := adapter.observe(ctx, *adapter.prepared, input.CalendarEvent.EventID)
	if err != nil {
		return devdemo.MailDemoResult{}, err
	}
	if dev1EventID == uuid.Nil || dev1EventID == input.CalendarEvent.EventID {
		return devdemo.MailDemoResult{}, errDevMailDemoUnavailable
	}
	if err := seedDevMailDemoInvitationOccurrence(ctx, adapter.engine, *adapter.prepared, input.CalendarEvent.EventID); err != nil {
		return devdemo.MailDemoResult{}, err
	}
	if err := adapter.observeOccurrence(ctx, *adapter.prepared, dev1EventID); err != nil {
		return devdemo.MailDemoResult{}, err
	}
	if err := adapter.recoverDevMailDemoBrandKitAttachment(ctx, *adapter.prepared); err != nil {
		return devdemo.MailDemoResult{}, err
	}
	replayed := 0
	if adapter.observedPMEventID == input.CalendarEvent.EventID {
		replayed = 2
	} else {
		adapter.observedPMEventID = input.CalendarEvent.EventID
	}
	return devMailDemoResult(*adapter.prepared, replayed)
}

func (adapter *devMailDemoAdapter) recoverCalendarInviteBeforeObserve(ctx context.Context, prepared devMailDemoPrepared, eventID uuid.UUID) error {
	if ctx == nil || eventID == uuid.Nil || !prepared.matches(prepared.organizationID, prepared.pmUserID, prepared.dev1UserID) {
		return errDevMailDemoInvalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	event, err := adapter.eventReadSvc.GetCalendarEvent(ctx, prepared.pmUserID, eventID)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		return errDevMailDemoUnavailable
	}
	attendeeID, matches := devMailDemoSeededInvitationAttendeeID(prepared, eventID, event)
	if !matches {
		return errDevMailDemoInvalid
	}
	eventSequence, err := calendarDemoUintToInt(calendarDemoStageCreate, event.Sequence)
	if err != nil {
		return errDevMailDemoInvalid
	}
	err = adapter.mailboxRepo.RecoverDevDemoCalendarInviteDelivery(ctx, mailmod.DevDemoCalendarInviteRecoveryRequest{
		OrganizationID: prepared.organizationID,
		EventID:        eventID,
		AttendeeID:     attendeeID,
		OrganizerID:    prepared.pmUserID,
		EventSequence:  eventSequence,
		Kind:           mailmod.CalendarInviteDeliveryKindCreated,
		RecurrenceID:   "",
	})
	if err == nil {
		return nil
	}
	if contextErr := ctx.Err(); contextErr != nil {
		return contextErr
	}
	var repositoryError *mailmod.CalendarMailDeliveryRepositoryError
	if errors.As(err, &repositoryError) && repositoryError != nil && repositoryError.Kind() == mailmod.CalendarMailDeliveryRepositoryErrorConflict {
		return adapter.recoverDevMailDemoSentCalendarInviteDelivery(ctx, prepared, event, attendeeID, eventSequence)
	}
	return errDevMailDemoUnavailable
}

// recoverDevMailDemoSentCalendarInviteDelivery permits a second, exact
// dev-only recovery only after the first recovery proved this is not its
// rejected-delivery shape. Both durable correlation paths and both owner-gated
// message bodies must demonstrate the same restart loss before a sent delivery
// can be requeued; mixed states never risk a duplicate PM invitation.
func (adapter *devMailDemoAdapter) recoverDevMailDemoSentCalendarInviteDelivery(ctx context.Context, prepared devMailDemoPrepared, pmEvent calendarmod.CalendarEventDetail, attendeeID uuid.UUID, eventSequence int) error {
	pmOwner, err := adapter.observeOwner(ctx, prepared, devMailSourcePMEmail, prepared.pmUserID)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		if errors.Is(err, errDevMailDemoPending) {
			return nil
		}
		return errDevMailDemoUnavailable
	}
	dev1Owner, err := adapter.observeOwner(ctx, prepared, devMailSourceDev1Email, prepared.dev1UserID)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		if errors.Is(err, errDevMailDemoPending) {
			return nil
		}
		return errDevMailDemoUnavailable
	}
	if pmOwner.accountID == dev1Owner.accountID {
		return errDevMailDemoUnavailable
	}
	pmMessageID, err := adapter.observePMInvitation(ctx, prepared, pmOwner, pmEvent.ID)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		if errors.Is(err, errDevMailDemoPending) {
			return nil
		}
		return errDevMailDemoUnavailable
	}
	dev1EventID, dev1MessageID, err := adapter.observeDev1Invitation(ctx, prepared, dev1Owner, pmEvent)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		if errors.Is(err, errDevMailDemoPending) {
			return nil
		}
		return errDevMailDemoUnavailable
	}
	if pmMessageID == uuid.Nil || dev1EventID == uuid.Nil || dev1MessageID == uuid.Nil || pmMessageID == dev1MessageID || pmEvent.ID == dev1EventID {
		return errDevMailDemoUnavailable
	}
	if err := adapter.observeInvitationSemantics(ctx, prepared, pmEvent, dev1EventID); err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		return errDevMailDemoUnavailable
	}
	pmUnavailable, err := adapter.devMailDemoCalendarInviteMessageUnavailable(ctx, pmOwner.userID, pmMessageID, pmEvent.ID)
	if err != nil {
		return err
	}
	dev1Unavailable, err := adapter.devMailDemoCalendarInviteMessageUnavailable(ctx, dev1Owner.userID, dev1MessageID, dev1EventID)
	if err != nil {
		return err
	}
	if !pmUnavailable && !dev1Unavailable {
		return nil
	}
	if !pmUnavailable || !dev1Unavailable {
		return errDevMailDemoUnavailable
	}
	result, err := adapter.mailboxRepo.RecoverDevDemoCalendarInviteSentDelivery(ctx, mailmod.DevDemoCalendarInviteSentRecoveryRequest{
		OrganizationID:     prepared.organizationID,
		EventID:            pmEvent.ID,
		AttendeeID:         attendeeID,
		OrganizerID:        prepared.pmUserID,
		EventSequence:      eventSequence,
		Kind:               mailmod.CalendarInviteDeliveryKindCreated,
		RecurrenceID:       "",
		RecipientUserID:    dev1Owner.userID,
		RecipientAccountID: dev1Owner.accountID,
		RecipientEventID:   dev1EventID,
		RecipientMessageID: dev1MessageID,
		ExpectedUID:        "hub-" + pmEvent.ID.String(),
	})
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		slog.Warn("dev demo calendar invite sent-delivery recovery failed", "recovery_error", err)
		return errDevMailDemoUnavailable
	}
	switch result.Outcome {
	case mailmod.DevDemoCalendarInviteSentRecoveryOutcomeNoop, mailmod.DevDemoCalendarInviteSentRecoveryOutcomeRequeued:
		return nil
	default:
		return errDevMailDemoUnavailable
	}
}

// devMailDemoCalendarInviteMessageUnavailable classifies only the one
// restart-loss signature eligible for sent-delivery recovery. Any other read
// error, or a readable message that no longer satisfies the invitation
// projection contract, fails closed.
func (adapter *devMailDemoAdapter) devMailDemoCalendarInviteMessageUnavailable(ctx context.Context, actorID, messageID, eventID uuid.UUID) (bool, error) {
	if ctx == nil || actorID == uuid.Nil || messageID == uuid.Nil || eventID == uuid.Nil {
		return false, errDevMailDemoUnavailable
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	detail, err := adapter.messageSvc.GetMailboxMessage(ctx, actorID, messageID)
	if err != nil {
		if contextErr := ctx.Err(); contextErr != nil {
			return false, contextErr
		}
		var readError *mailmod.MailboxMessageReadError
		if errors.As(err, &readError) && readError != nil && readError.Kind() == mailmod.MailboxMessageReadErrorUnavailable {
			return true, nil
		}
		return false, errDevMailDemoUnavailable
	}
	if detail.ID != messageID || detail.ThreadID == uuid.Nil || detail.CalendarMethod == nil || *detail.CalendarMethod != mailmod.MailCalendarMethodRequest || detail.CalendarEventID == nil || *detail.CalendarEventID != eventID || detail.PlainBody == "" || detail.SanitizedHTML != nil || !detail.RemoteImagesBlocked || detail.IsDraft || detail.Subject == "" {
		return false, errDevMailDemoUnavailable
	}
	return false, nil
}

// devMailDemoSeededInvitationAttendeeID admits only the immutable Calendar
// fixture reconciled by the dev seeder. Recovery must never act on an
// arbitrary PM-to-dev1 event merely because it has the same participants.
func devMailDemoSeededInvitationAttendeeID(prepared devMailDemoPrepared, eventID uuid.UUID, event calendarmod.CalendarEventDetail) (uuid.UUID, bool) {
	startsAt := time.Date(2026, time.September, 3, 10, 0, 0, 0, time.UTC)
	endsAt := time.Date(2026, time.September, 3, 10, 30, 0, 0, time.UTC)
	if eventID == uuid.Nil || event.ID != eventID || event.Title != devMailDemoInvitationTitle || event.Description != devMailDemoInvitationDescription || event.Location != devMailDemoInvitationLocation ||
		event.AllDay || event.Timezone != devMailDemoInvitationTimezone || event.StartsAt == nil || event.EndsAt == nil || event.StartDate != nil || event.EndDateExclusive != nil ||
		event.StartsAt.Location() != time.UTC || event.EndsAt.Location() != time.UTC || !event.StartsAt.Equal(startsAt) || !event.EndsAt.Equal(endsAt) || event.RRULE == nil || *event.RRULE != devMailDemoInvitationRRULE || len(event.Exdates) != 0 ||
		event.Status != calendarmod.CalendarEventStatusConfirmed || event.Version != 1 || event.OrganizerID == nil || *event.OrganizerID != prepared.pmUserID || event.OrganizerEmail != devMailSourcePMEmail || len(event.Attendees) != 2 {
		return uuid.Nil, false
	}

	attendeeID := uuid.Nil
	organizerCount := 0
	for _, attendee := range event.Attendees {
		if attendee.ID == uuid.Nil || attendee.UserID == nil || attendee.Version != 1 || attendee.RespondedAt != nil {
			return uuid.Nil, false
		}
		switch *attendee.UserID {
		case prepared.pmUserID:
			if attendee.Email != devMailSourcePMEmail || attendee.Role != calendarmod.CalendarEventAttendeeRoleOrganizer || attendee.RSVPStatus != calendarmod.RSVPAccepted {
				return uuid.Nil, false
			}
			organizerCount++
		case prepared.dev1UserID:
			if attendee.Email != devMailSourceDev1Email || attendee.Role != calendarmod.CalendarEventAttendeeRoleRequired || attendee.RSVPStatus != calendarmod.RSVPNeedsAction || attendeeID != uuid.Nil {
				return uuid.Nil, false
			}
			attendeeID = attendee.ID
		default:
			return uuid.Nil, false
		}
	}
	return attendeeID, organizerCount == 1 && attendeeID != uuid.Nil
}

func (adapter *devMailDemoAdapter) observe(ctx context.Context, prepared devMailDemoPrepared, pmEventID uuid.UUID) (uuid.UUID, error) {
	if ctx == nil {
		return uuid.Nil, errDevMailDemoUnavailable
	}
	if err := ctx.Err(); err != nil {
		return uuid.Nil, err
	}
	deadlineCtx, cancel := context.WithTimeout(ctx, devMailDemoObserveTimeout)
	defer cancel()
	for {
		dev1EventID, err := adapter.observeOnce(deadlineCtx, prepared, pmEventID)
		if err == nil {
			if dev1EventID == uuid.Nil || dev1EventID == pmEventID {
				return uuid.Nil, errDevMailDemoUnavailable
			}
			return dev1EventID, nil
		}
		if contextErr := ctx.Err(); contextErr != nil {
			return uuid.Nil, contextErr
		}
		if deadlineErr := deadlineCtx.Err(); deadlineErr != nil {
			return uuid.Nil, deadlineErr
		}
		if !errors.Is(err, errDevMailDemoPending) {
			return uuid.Nil, errDevMailDemoUnavailable
		}
		timer := time.NewTimer(devMailDemoObservePoll)
		select {
		case <-deadlineCtx.Done():
			if !timer.Stop() {
				select {
				case <-timer.C:
				default:
				}
			}
			return uuid.Nil, deadlineCtx.Err()
		case <-timer.C:
		}
	}
}

func devMailDemoResult(prepared devMailDemoPrepared, replayed int) (devdemo.MailDemoResult, error) {
	source := prepared.source
	if (replayed != 0 && replayed != 2) || !validDevMailSourceResult(source) ||
		source.accountsTotal != 2 || source.foldersTotal != 12 || source.threadsTotal != 3 ||
		source.messagesTotal != 4 || source.attachmentsTotal != 1 {
		return devdemo.MailDemoResult{}, errDevMailDemoUnavailable
	}
	return devdemo.MailDemoResult{
		DatasetVersion:      devMailDemoDatasetVersion,
		AccountsTotal:       source.accountsTotal,
		FoldersTotal:        source.foldersTotal,
		ThreadsTotal:        source.threadsTotal,
		MessagesTotal:       source.messagesTotal,
		AttachmentsTotal:    source.attachmentsTotal,
		InvitationsTotal:    2,
		InvitationsReplayed: replayed,
	}, nil
}

func devMailDemoSourceResultEqual(left, right devMailSourceResult) bool {
	return reflect.DeepEqual(left, right)
}

func (adapter *devMailDemoAdapter) ready() bool {
	return adapter != nil && !devMailDemoNil(adapter.client) && !devMailDemoNil(adapter.engine) &&
		!devMailDemoNil(adapter.mailboxSvc) && !devMailDemoNil(adapter.mailboxRepo) &&
		!devMailDemoNil(adapter.messageSvc) && !devMailDemoNil(adapter.eventReadSvc) && !devMailDemoNil(adapter.projectionSvc) &&
		adapter.now != nil && adapter.startRuntime != nil
}

func (prepared devMailDemoPrepared) matches(organizationID, pmUserID, dev1UserID uuid.UUID) bool {
	return organizationID != uuid.Nil && pmUserID != uuid.Nil && dev1UserID != uuid.Nil &&
		prepared.organizationID == organizationID && prepared.pmUserID == pmUserID && prepared.dev1UserID == dev1UserID
}

func devMailDemoScopeUsers(userIDs map[string]uuid.UUID) (uuid.UUID, uuid.UUID, bool) {
	if userIDs == nil {
		return uuid.Nil, uuid.Nil, false
	}
	pmUserID, pmOK := userIDs[devMailSourcePMEmail]
	dev1UserID, dev1OK := userIDs[devMailSourceDev1Email]
	if !pmOK || !dev1OK || pmUserID == uuid.Nil || dev1UserID == uuid.Nil || pmUserID == dev1UserID {
		return uuid.Nil, uuid.Nil, false
	}
	return pmUserID, dev1UserID, true
}

func devMailDemoNil(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
