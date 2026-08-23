package calendar

import (
	"context"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendareventexception"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/scheduledcall"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/internal/platform/outbox"
)

type eventEntCallTransactionBackend struct {
	tx *ent.Tx
}

var _ EventCallRepository = (*EventEntRepository)(nil)
var _ EventCallTransactionBackend = (*eventEntCallTransactionBackend)(nil)

func newEventEntCallTransactionBackend(transaction *ent.Tx) *eventEntCallTransactionBackend {
	return &eventEntCallTransactionBackend{tx: transaction}
}

// WithEventCallTransaction owns one isolated Ent transaction for the Calendar-to-call bridge.
func (repository *EventEntRepository) WithEventCallTransaction(ctx context.Context, callback func(*EventCallTransaction) error) error {
	if repository == nil || repository.client == nil || callback == nil {
		return ErrEventInvalid
	}
	transaction, err := repository.client.Tx(ctx)
	if err != nil {
		return calendarEntSafeError(err)
	}
	committed := false
	defer func() {
		if recovered := recover(); recovered != nil {
			_ = transaction.Rollback()
			panic(recovered)
		}
		if !committed {
			_ = transaction.Rollback()
		}
	}()
	if err := RunEventCallTransaction(newEventEntCallTransactionBackend(transaction), callback); err != nil {
		return err
	}
	if err := transaction.Commit(); err != nil {
		return calendarEntSafeError(err)
	}
	committed = true
	return nil
}

type eventEntCallRows struct {
	event        *ent.CalendarEvent
	call         *ent.ScheduledCall
	eventVersion uint
	callVersion  uint
}

func (backend *eventEntCallTransactionBackend) lockEventCall(ctx context.Context, scheduledCallID uuid.UUID, expectedEventVersion, expectedCallVersion uint) (eventEntCallRows, error) {
	if backend == nil || backend.tx == nil || scheduledCallID == uuid.Nil || expectedEventVersion == 0 || expectedCallVersion == 0 {
		return eventEntCallRows{}, ErrEventInvalid
	}
	event, err := backend.tx.CalendarEvent.Query().
		Where(calendarevent.ScheduledCallIDEQ(scheduledCallID)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return eventEntCallRows{}, calendarEntSafeError(err)
	}
	call, err := backend.tx.ScheduledCall.Query().
		Where(scheduledcall.IDEQ(scheduledCallID), scheduledcall.HasCalendarEventWith(calendarevent.IDEQ(event.ID))).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return eventEntCallRows{}, calendarEntSafeError(err)
	}
	if event.ScheduledCallID == nil || *event.ScheduledCallID != call.ID {
		return eventEntCallRows{}, ErrEventConflict
	}
	eventVersion, eventVersionOK := eventEntUint(event.Version)
	callVersion, callVersionOK := eventEntUint(call.Version)
	if !eventVersionOK || !callVersionOK {
		return eventEntCallRows{}, ErrEventInvalid
	}
	if eventVersion != expectedEventVersion || callVersion != expectedCallVersion {
		return eventEntCallRows{}, ErrEventConflict
	}
	return eventEntCallRows{event: event, call: call, eventVersion: eventVersion, callVersion: callVersion}, nil
}

func (backend *eventEntCallTransactionBackend) ResolveEventCallProjection(ctx context.Context, request EventCallProjectionRequest) (EventCallProjectionInput, error) {
	if !request.valid() {
		return EventCallProjectionInput{}, ErrEventInvalid
	}
	rows, err := backend.lockEventCall(ctx, request.ScheduledCallID(), request.ExpectedEventVersion(), request.ExpectedCallVersion())
	if err != nil {
		return EventCallProjectionInput{}, err
	}
	return backend.eventCallProjectionInput(ctx, rows, request.Action(), request.Boundary())
}

func (backend *eventEntCallTransactionBackend) ResolveEventCallDue(ctx context.Context, request EventCallDueRequest) (EventCallDueInput, error) {
	if !request.valid() {
		return EventCallDueInput{}, ErrEventInvalid
	}
	rows, err := backend.lockEventCall(ctx, request.ScheduledCallID(), request.ExpectedEventVersion(), request.ExpectedCallVersion())
	if err != nil {
		return EventCallDueInput{}, err
	}
	return backend.eventCallDueInput(ctx, rows, request)
}

func eventEntCallDueSnapshot(rows eventEntCallRows) EventCallDueCallSnapshot {
	return EventCallDueCallSnapshot{
		ScheduledCallID: rows.call.ID, CalendarEventID: rows.event.ID, OrganizationID: rows.event.OrganizationID,
		ProjectID: rows.call.ProjectID, ChannelID: rows.call.ChannelID, CreatorID: rows.call.CreatedBy,
		ScheduledAt: rows.call.ScheduledAt.UTC(), Active: rows.call.IsActive, Deleted: rows.call.DeletedAt != nil,
	}
}

func (backend *eventEntCallTransactionBackend) eventCallDueProjectionInput(ctx context.Context, rows eventEntCallRows) (EventCallDueProjectionInput, error) {
	input, err := backend.eventCallProjectionInput(ctx, rows, EventCallProjectionReconcile, EventCallProjectionBoundary{At: rows.call.ScheduledAt.UTC(), Inclusive: true})
	if err != nil {
		return EventCallDueProjectionInput{}, err
	}
	return EventCallDueProjectionInput{Event: input.Event, Target: input.Target, Creator: input.Creator, Attendees: input.Attendees, Exceptions: input.Exceptions}, nil
}

func (backend *eventEntCallTransactionBackend) eventCallProjectionInput(ctx context.Context, rows eventEntCallRows, action EventCallProjectionAction, boundary EventCallProjectionBoundary) (EventCallProjectionInput, error) {
	current, err := eventEntCalendarEventInput(rows.event)
	if err != nil || current.ScheduledCallID == nil || *current.ScheduledCallID != rows.call.ID || current.OrganizerID == nil ||
		current.OwnerID != *current.OrganizerID || *current.OrganizerID != rows.call.CreatedBy {
		return EventCallProjectionInput{}, ErrEventConflict
	}
	creator, err := backend.eventCallCreatorForUpdate(ctx, rows)
	if err != nil {
		return EventCallProjectionInput{}, err
	}
	attendees, err := backend.eventCallAttendeesForUpdate(ctx, rows)
	if err != nil {
		return EventCallProjectionInput{}, err
	}
	exceptions, err := backend.eventCallExceptionsForUpdate(ctx, rows.event.ID)
	if err != nil {
		return EventCallProjectionInput{}, err
	}
	event := EventCallProjectionEvent{
		EventID: rows.event.ID, ScheduledCallID: rows.call.ID, OrganizationID: current.OrganizationID,
		ProjectID: rows.call.ProjectID, ChannelID: rows.call.ChannelID, OrganizerID: *current.OrganizerID,
		OrganizerEmail: current.OrganizerEmail, Source: current.Source, Status: current.Status,
		Deleted: rows.event.DeletedAt != nil, Title: current.Title, Description: current.Description,
		Temporal: current.Temporal, RRULE: current.RRULE, Exdates: append([]string(nil), current.Exdates...),
		ReminderMinutes: cloneEventCallProjectionInt(current.ReminderMinutes),
	}
	target := EventCallProjectionTarget{
		ScheduledCallID: rows.call.ID, EventID: rows.event.ID, CreatorID: rows.call.CreatedBy,
		OrganizationID: current.OrganizationID, ProjectID: rows.call.ProjectID, ChannelID: rows.call.ChannelID,
		Exists: true, Active: rows.call.IsActive, Deleted: rows.call.DeletedAt != nil,
	}
	return EventCallProjectionInput{Action: action, Event: event, Target: target, Creator: creator, Attendees: attendees, Exceptions: exceptions, Boundary: boundary}, nil
}

func (backend *eventEntCallTransactionBackend) eventCallCreatorForUpdate(ctx context.Context, rows eventEntCallRows) (EventCallProjectionCreator, error) {
	if rows.event.OrganizerID == nil || *rows.event.OrganizerID != rows.call.CreatedBy {
		return EventCallProjectionCreator{}, ErrEventConflict
	}
	if _, err := backend.tx.Organization.Query().Where(
		organization.IDEQ(rows.event.OrganizationID),
		organization.HasUsersWith(user.IDEQ(rows.call.CreatedBy), user.StatusEQ(user.StatusActive)),
		organization.HasProjectsWith(project.IDEQ(rows.call.ProjectID)),
	).ForUpdate().Only(ctx); err != nil {
		return EventCallProjectionCreator{}, calendarEntSafeError(err)
	}
	creator, err := backend.tx.User.Query().Where(user.IDEQ(rows.call.CreatedBy), user.StatusEQ(user.StatusActive)).ForUpdate().Only(ctx)
	if err != nil {
		return EventCallProjectionCreator{}, calendarEntSafeError(err)
	}
	if _, err := backend.tx.Project.Query().Where(project.IDEQ(rows.call.ProjectID), project.HasMembersWith(projectmember.HasUserWith(user.IDEQ(creator.ID)))).ForUpdate().Only(ctx); err != nil {
		return EventCallProjectionCreator{}, calendarEntSafeError(err)
	}
	if _, err := backend.tx.Channel.Query().Where(channel.IDEQ(rows.call.ChannelID), channel.ProjectIDEQ(rows.call.ProjectID), channel.DeletedAtIsNil(), channel.HasMembersWith(channelmember.UserIDEQ(creator.ID))).ForUpdate().Only(ctx); err != nil {
		return EventCallProjectionCreator{}, calendarEntSafeError(err)
	}
	if creator.Email != rows.event.OrganizerEmail {
		return EventCallProjectionCreator{}, ErrEventConflict
	}
	return EventCallProjectionCreator{UserID: creator.ID, Email: creator.Email, OrganizationID: rows.event.OrganizationID, ProjectID: rows.call.ProjectID, ChannelID: rows.call.ChannelID, Active: true, ActiveProjectMember: true, ActiveChannelMember: true}, nil
}

func (backend *eventEntCallTransactionBackend) eventCallAttendeesForUpdate(ctx context.Context, rows eventEntCallRows) ([]EventCallProjectionAttendee, error) {
	locked, err := backend.tx.CalendarAttendee.Query().Where(calendarattendee.CalendarEventIDEQ(rows.event.ID), calendarattendee.RemovedAtIsNil()).Order(ent.Asc(calendarattendee.FieldID)).ForUpdate().All(ctx)
	if err != nil {
		return nil, calendarEntSafeError(err)
	}
	attendees := make([]EventCallProjectionAttendee, 0, len(locked))
	for _, attendee := range locked {
		if attendee.UserID == nil {
			return nil, ErrEventInvalid
		}
		var recurrenceID string
		if attendee.RecurrenceID != nil {
			recurrenceID = *attendee.RecurrenceID
		}
		role := CalendarEventAttendeeRole(attendee.Role)
		status := RSVPStatus(attendee.RsvpStatus)
		if !validCalendarEventStoredAttendeeRoleState(role, status, recurrenceID) {
			return nil, ErrEventInvalid
		}
		if role == CalendarEventAttendeeRoleOrganizer && (rows.event.OrganizerID == nil ||
			!validCalendarEventStoredOrganizerAttendeeState(role, status, recurrenceID, attendee.UserID, rows.event.CalendarOwnerID, *rows.event.OrganizerID)) {
			return nil, ErrEventConflict
		}
		member, err := backend.tx.User.Query().Where(user.IDEQ(*attendee.UserID), user.StatusEQ(user.StatusActive)).ForUpdate().Only(ctx)
		if err != nil {
			return nil, calendarEntSafeError(err)
		}
		if _, err := backend.tx.Organization.Query().Where(organization.IDEQ(rows.event.OrganizationID), organization.HasUsersWith(user.IDEQ(member.ID), user.StatusEQ(user.StatusActive))).ForUpdate().Only(ctx); err != nil {
			return nil, calendarEntSafeError(err)
		}
		if _, err := backend.tx.Project.Query().Where(project.IDEQ(rows.call.ProjectID), project.HasMembersWith(projectmember.HasUserWith(user.IDEQ(member.ID)))).ForUpdate().Only(ctx); err != nil {
			return nil, calendarEntSafeError(err)
		}
		if attendee.NormalizedEmail != member.Email {
			return nil, ErrEventConflict
		}
		attendeeID := *attendee.UserID
		attendees = append(attendees, EventCallProjectionAttendee{
			UserID: &attendeeID, Email: member.Email, OrganizationID: rows.event.OrganizationID,
			ProjectID: rows.call.ProjectID, Role: role,
			Status: status, RecurrenceID: recurrenceID,
			Active: true, ActiveProjectMember: true,
		})
	}
	return attendees, nil
}

func (backend *eventEntCallTransactionBackend) eventCallExceptionsForUpdate(ctx context.Context, eventID uuid.UUID) ([]EventCallProjectionException, error) {
	locked, err := backend.tx.CalendarEventException.Query().
		Where(calendareventexception.CalendarEventIDEQ(eventID)).
		Order(ent.Asc(calendareventexception.FieldID)).
		ForUpdate().
		All(ctx)
	if err != nil {
		return nil, calendarEntSafeError(err)
	}
	exceptions := make([]EventCallProjectionException, 0, len(locked))
	for _, exception := range locked {
		current, err := eventEntCalendarEventExceptionInput(exception)
		if err != nil {
			return nil, ErrEventInvalid
		}
		exceptions = append(exceptions, EventCallProjectionException{
			RecurrenceID: current.RecurrenceID, Temporal: current.Temporal,
			Status: current.Status, Cancelled: current.Cancelled,
		})
	}
	return exceptions, nil
}

func (backend *eventEntCallTransactionBackend) PersistEventCallProjection(ctx context.Context, request EventCallProjectionRequest, projection EventCallProjection) (EventCallProjectionSnapshot, error) {
	if !request.valid() {
		return EventCallProjectionSnapshot{}, ErrEventInvalid
	}
	rows, err := backend.lockEventCall(ctx, request.ScheduledCallID(), request.ExpectedEventVersion(), request.ExpectedCallVersion())
	if err != nil {
		return EventCallProjectionSnapshot{}, err
	}
	input, err := backend.eventCallProjectionInput(ctx, rows, request.Action(), request.Boundary())
	if err != nil {
		return EventCallProjectionSnapshot{}, err
	}
	expected, err := ProjectEventCall(input)
	if err != nil {
		return EventCallProjectionSnapshot{}, err
	}
	snapshot := expected.Snapshot()
	if !reflect.DeepEqual(snapshot, projection.Snapshot()) {
		return EventCallProjectionSnapshot{}, ErrEventConflict
	}
	if err := backend.persistEventCallProjectionSnapshot(ctx, rows, snapshot); err != nil {
		return EventCallProjectionSnapshot{}, err
	}
	return snapshot, nil
}

func (backend *eventEntCallTransactionBackend) persistEventCallProjectionSnapshot(ctx context.Context, rows eventEntCallRows, snapshot EventCallProjectionSnapshot) error {
	expectedCallVersion, nextCallVersion, valid := eventEntCallNextVersion(rows.callVersion)
	if !valid {
		return ErrEventInvalid
	}
	switch snapshot.Disposition {
	case EventCallProjectionUpsert:
		if snapshot.DeletedAt != nil || snapshot.Occurrence == nil || !eventCallProjectionExactUTCSecond(snapshot.Occurrence.StartsAt) {
			return ErrEventInvalid
		}
		update := backend.tx.ScheduledCall.UpdateOneID(rows.call.ID).
			Where(scheduledcall.VersionEQ(expectedCallVersion)).
			SetTitle(snapshot.Title).
			SetDescription(snapshot.Description).
			SetScheduledAt(snapshot.Occurrence.StartsAt.UTC()).
			SetDurationMinutes(snapshot.DurationMinutes).
			SetParticipantIDs(append([]uuid.UUID(nil), snapshot.ParticipantIDs...)...).
			SetIsActive(true).
			ClearDeletedAt().
			SetVersion(nextCallVersion)
		if len(snapshot.RRULE) == 0 {
			update.ClearRrule()
		} else {
			update.SetRrule(snapshot.RRULE)
		}
		if snapshot.ReminderMinutes == nil {
			update.ClearReminderMinutes()
		} else {
			update.SetReminderMinutes(*snapshot.ReminderMinutes)
		}
		if err := update.Exec(ctx); err != nil {
			return calendarEntCASFailure(err)
		}
		return nil
	case EventCallProjectionPaused, EventCallProjectionCompleted:
		if snapshot.DeletedAt != nil {
			return ErrEventInvalid
		}
		if err := backend.tx.ScheduledCall.UpdateOneID(rows.call.ID).
			Where(scheduledcall.VersionEQ(expectedCallVersion)).
			SetIsActive(false).
			SetVersion(nextCallVersion).
			Exec(ctx); err != nil {
			return calendarEntCASFailure(err)
		}
		return nil
	case EventCallProjectionDeleted:
		if snapshot.DeletedAt == nil || !eventCallProjectionExactUTCSecond(*snapshot.DeletedAt) {
			return ErrEventInvalid
		}
		if err := backend.tx.ScheduledCall.UpdateOneID(rows.call.ID).
			Where(scheduledcall.VersionEQ(expectedCallVersion)).
			SetIsActive(false).
			SetDeletedAt(snapshot.DeletedAt.UTC()).
			SetVersion(nextCallVersion).
			Exec(ctx); err != nil {
			return calendarEntCASFailure(err)
		}
		return nil
	default:
		return ErrEventInvalid
	}
}

func eventEntCallNextVersion(version uint) (int, int, bool) {
	expected, valid := eventEntInt(version)
	if !valid || expected < 1 || expected == int(^uint(0)>>1) {
		return 0, 0, false
	}
	return expected, expected + 1, true
}

func (backend *eventEntCallTransactionBackend) PersistEventCallDue(ctx context.Context, request EventCallDueRequest, decision EventCallDueDecision) (EventCallDueSnapshot, error) {
	if !request.valid() {
		return EventCallDueSnapshot{}, ErrEventInvalid
	}
	rows, err := backend.lockEventCall(ctx, request.ScheduledCallID(), request.ExpectedEventVersion(), request.ExpectedCallVersion())
	if err != nil {
		return EventCallDueSnapshot{}, err
	}
	input, err := backend.eventCallDueInput(ctx, rows, request)
	if err != nil {
		return EventCallDueSnapshot{}, err
	}
	expected, err := ProjectEventCallDue(input)
	if err != nil {
		return EventCallDueSnapshot{}, err
	}
	snapshot := expected.Snapshot()
	if !reflect.DeepEqual(snapshot, decision.Snapshot()) || !eventEntCallOutboxEqual(expected, decision) {
		return EventCallDueSnapshot{}, ErrEventConflict
	}
	if err := backend.persistEventCallDueSnapshot(ctx, rows, snapshot); err != nil {
		return EventCallDueSnapshot{}, err
	}
	if descriptor, found := expected.Outbox(); found {
		if err := backend.appendEventCallDueOutbox(ctx, rows.call.ChannelID, descriptor); err != nil {
			return EventCallDueSnapshot{}, err
		}
	}
	return snapshot, nil
}

func (backend *eventEntCallTransactionBackend) eventCallDueInput(ctx context.Context, rows eventEntCallRows, request EventCallDueRequest) (EventCallDueInput, error) {
	input := EventCallDueInput{Call: eventEntCallDueSnapshot(rows), Now: request.Now()}
	if !rows.call.IsActive || rows.call.DeletedAt != nil {
		return input, nil
	}
	projection, err := backend.eventCallDueProjectionInput(ctx, rows)
	if err != nil {
		return EventCallDueInput{}, err
	}
	input.Projection = &projection
	input.Envelope = request.Envelope()
	return input, nil
}

func eventEntCallOutboxEqual(left, right EventCallDueDecision) bool {
	leftDescriptor, leftFound := left.Outbox()
	rightDescriptor, rightFound := right.Outbox()
	if leftFound != rightFound {
		return false
	}
	return !leftFound || reflect.DeepEqual(leftDescriptor.Envelope(), rightDescriptor.Envelope())
}

func (backend *eventEntCallTransactionBackend) persistEventCallDueSnapshot(ctx context.Context, rows eventEntCallRows, snapshot EventCallDueSnapshot) error {
	if !eventEntCallDueSnapshotMatchesRows(rows, snapshot) {
		return ErrEventInvalid
	}
	switch snapshot.Disposition {
	case EventCallDueNotDue, EventCallDueNoop:
		return nil
	case EventCallDueReschedule, EventCallDueAdvance:
		if snapshot.ScheduledAt == nil || !snapshot.Active || !eventCallProjectionExactUTCSecond(*snapshot.ScheduledAt) {
			return ErrEventInvalid
		}
		return backend.persistEventCallDueMutation(ctx, rows, snapshot.ScheduledAt.UTC(), true)
	case EventCallDueDeactivate:
		if snapshot.ScheduledAt != nil || snapshot.Active {
			return ErrEventInvalid
		}
		return backend.persistEventCallDueMutation(ctx, rows, time.Time{}, false)
	default:
		return ErrEventInvalid
	}
}

func eventEntCallDueSnapshotMatchesRows(rows eventEntCallRows, snapshot EventCallDueSnapshot) bool {
	return snapshot.ScheduledCallID == rows.call.ID && snapshot.CalendarEventID == rows.event.ID &&
		snapshot.OrganizationID == rows.event.OrganizationID && snapshot.ProjectID == rows.call.ProjectID &&
		snapshot.ChannelID == rows.call.ChannelID && snapshot.CreatorID == rows.call.CreatedBy
}

func (backend *eventEntCallTransactionBackend) persistEventCallDueMutation(ctx context.Context, rows eventEntCallRows, scheduledAt time.Time, active bool) error {
	expectedCallVersion, nextCallVersion, valid := eventEntCallNextVersion(rows.callVersion)
	if !valid {
		return ErrEventInvalid
	}
	update := backend.tx.ScheduledCall.UpdateOneID(rows.call.ID).
		Where(scheduledcall.VersionEQ(expectedCallVersion)).
		SetIsActive(active).
		SetVersion(nextCallVersion)
	if !scheduledAt.IsZero() {
		update.SetScheduledAt(scheduledAt.UTC())
	}
	if err := update.Exec(ctx); err != nil {
		return calendarEntCASFailure(err)
	}
	return nil
}

func (backend *eventEntCallTransactionBackend) appendEventCallDueOutbox(ctx context.Context, channelID uuid.UUID, descriptor EventOutboxDescriptor) error {
	if backend == nil || backend.tx == nil || channelID == uuid.Nil || !descriptor.valid() {
		return ErrEventInvalid
	}
	envelope := descriptor.Envelope()
	if envelope.Subject() != eventEntCallDueSubject(channelID) {
		return ErrEventInvalid
	}
	return calendarEntSafeError(outbox.NewPublisher().PublishInTx(ctx, backend.tx, envelope.Subject(), envelope))
}

func eventEntCallDueSubject(channelID uuid.UUID) string {
	return "calls.channel." + channelID.String() + ".scheduled_due.v1"
}
