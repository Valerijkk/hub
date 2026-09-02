//go:build !qa && !release

package main

import (
	"context"
	"errors"
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
	devMailDemoDatasetVersion = 1
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
	client         *appent.Client
	engine         *mailmod.FakeMailEngine
	mailboxSvc     *mailmod.MailboxService
	mailboxRepo    *mailmod.MailboxEntRepository
	eventReadSvc   *calendarmod.EventReadService
	startRuntime   func() error
	projectionSvc  *mailmod.MailboxFolderProjectionService
	now            func() time.Time

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
	return &devMailDemoAdapter{
		client:        client,
		engine:        engine,
		mailboxSvc:    mailboxSvc,
		mailboxRepo:   mailboxRepo,
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
// dev1 receipt. It deliberately does not call worker step/run APIs.
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
	if err := adapter.observe(ctx, *adapter.prepared, input.CalendarEvent.EventID); err != nil {
		return devdemo.MailDemoResult{}, err
	}
	replayed := 0
	if adapter.observedPMEventID == input.CalendarEvent.EventID {
		replayed = 1
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
		return nil
	}
	return errDevMailDemoUnavailable
}

// devMailDemoSeededInvitationAttendeeID admits only the immutable Calendar
// fixture reconciled by the dev seeder. Recovery must never act on an
// arbitrary PM-to-dev1 event merely because it has the same participants.
func devMailDemoSeededInvitationAttendeeID(prepared devMailDemoPrepared, eventID uuid.UUID, event calendarmod.CalendarEventDetail) (uuid.UUID, bool) {
	startsAt := time.Date(2026, time.September, 3, 10, 0, 0, 0, time.UTC)
	endsAt := time.Date(2026, time.September, 3, 10, 30, 0, 0, time.UTC)
	if eventID == uuid.Nil || event.ID != eventID || event.Title != devMailDemoInvitationTitle || event.Description != devMailDemoInvitationDescription || event.Location != devMailDemoInvitationLocation ||
		event.AllDay || event.Timezone != devMailDemoInvitationTimezone || event.StartsAt == nil || event.EndsAt == nil || event.StartDate != nil || event.EndDateExclusive != nil ||
		event.StartsAt.Location() != time.UTC || event.EndsAt.Location() != time.UTC || !event.StartsAt.Equal(startsAt) || !event.EndsAt.Equal(endsAt) || event.RRULE != nil || len(event.Exdates) != 0 ||
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

func (adapter *devMailDemoAdapter) observe(ctx context.Context, prepared devMailDemoPrepared, pmEventID uuid.UUID) error {
	if ctx == nil {
		return errDevMailDemoUnavailable
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	deadlineCtx, cancel := context.WithTimeout(ctx, devMailDemoObserveTimeout)
	defer cancel()
	for {
		err := adapter.observeOnce(deadlineCtx, prepared, pmEventID)
		if err == nil {
			return nil
		}
		if contextErr := ctx.Err(); contextErr != nil {
			return contextErr
		}
		if deadlineErr := deadlineCtx.Err(); deadlineErr != nil {
			return deadlineErr
		}
		if !errors.Is(err, errDevMailDemoPending) {
			return errDevMailDemoUnavailable
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
			return deadlineCtx.Err()
		case <-timer.C:
		}
	}
}

func devMailDemoResult(prepared devMailDemoPrepared, replayed int) (devdemo.MailDemoResult, error) {
	source := prepared.source
	if replayed < 0 || replayed > 1 || !validDevMailSourceResult(source) ||
		source.accountsTotal != 2 || source.foldersTotal != 12 || source.threadsTotal != 3 ||
		source.messagesTotal != 4 || source.attachmentsTotal != 1 {
		return devdemo.MailDemoResult{}, errDevMailDemoUnavailable
	}
	return devdemo.MailDemoResult{
		DatasetVersion:       devMailDemoDatasetVersion,
		AccountsTotal:        source.accountsTotal,
		FoldersTotal:         source.foldersTotal,
		ThreadsTotal:         source.threadsTotal,
		MessagesTotal:        source.messagesTotal,
		AttachmentsTotal:     source.attachmentsTotal,
		InvitationsTotal:     1,
		InvitationsReplayed: replayed,
	}, nil
}

func devMailDemoSourceResultEqual(left, right devMailSourceResult) bool {
	return reflect.DeepEqual(left, right)
}

func (adapter *devMailDemoAdapter) ready() bool {
	return adapter != nil && !devMailDemoNil(adapter.client) && !devMailDemoNil(adapter.engine) &&
		!devMailDemoNil(adapter.mailboxSvc) && !devMailDemoNil(adapter.mailboxRepo) &&
		!devMailDemoNil(adapter.eventReadSvc) && !devMailDemoNil(adapter.projectionSvc) &&
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
