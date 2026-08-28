package calendar

import (
	"context"
	"errors"
	"sort"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendaravailabilityreservation"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendareventexception"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/scheduledcall"
	"github.com/valerii/hub/backend/ent/user"
)

type eventEntTransactionBackend struct {
	tx                          *ent.Tx
	now                         func() time.Time
	inboundMessageDatabaseClock *time.Time
}

var _ EventTransactionBackend = (*eventEntTransactionBackend)(nil)
var _ CalendarInboundMessageTransactionBackend = (*eventEntTransactionBackend)(nil)
var _ MeetingTransactionBackend = (*eventEntTransactionBackend)(nil)
var _ MeetingUpdateTransactionBackend = (*eventEntTransactionBackend)(nil)
var _ MeetingCancelTransactionBackend = (*eventEntTransactionBackend)(nil)

func newEventEntTransactionBackend(transaction *ent.Tx, now func() time.Time) *eventEntTransactionBackend {
	return &eventEntTransactionBackend{tx: transaction, now: now}
}

func (backend *eventEntTransactionBackend) calendarInboundMessageDatabaseClock(ctx context.Context) (time.Time, error) {
	if backend == nil || backend.tx == nil || ctx == nil {
		return time.Time{}, ErrEventInvalid
	}
	if backend.inboundMessageDatabaseClock != nil {
		clock := *backend.inboundMessageDatabaseClock
		if !calendarInboundInviteWholeSecondUTC(clock) {
			return time.Time{}, ErrEventInvalid
		}
		return clock, nil
	}

	clock, err := calendarInboundInviteDatabaseClock(ctx, backend.tx)
	if err != nil {
		return time.Time{}, err
	}
	clockCopy := clock
	backend.inboundMessageDatabaseClock = &clockCopy
	return clockCopy, nil
}

func (backend *eventEntTransactionBackend) ResolveActorOrganization(ctx context.Context, actorID uuid.UUID) (uuid.UUID, error) {
	if backend == nil || backend.tx == nil || actorID == uuid.Nil {
		return uuid.Nil, ErrEventInvalid
	}

	organization, err := backend.activeOrganizationForActorForUpdate(ctx, actorID)
	if err != nil {
		return uuid.Nil, err
	}
	if _, err := backend.activeUserForUpdate(ctx, actorID); err != nil {
		return uuid.Nil, err
	}
	return organization.ID, nil
}

func (backend *eventEntTransactionBackend) ResolveCalendarEventCreateAuthorization(ctx context.Context, scope EventMutationScope, request CalendarEventCreateAuthorizationRequest) (CalendarEventCreateAuthorization, error) {
	if backend == nil || backend.tx == nil || scope.ActorID() == uuid.Nil || scope.OrganizationID() == uuid.Nil || request.Organizer.UserID == nil || *request.Organizer.UserID != scope.ActorID() {
		return CalendarEventCreateAuthorization{}, ErrEventForbidden
	}

	organizer, err := backend.activeOrganizationMemberForUpdate(ctx, scope.ActorID(), scope.OrganizationID())
	if err != nil {
		return CalendarEventCreateAuthorization{}, err
	}
	resolved := CalendarEventCreateAuthorization{
		Organizer: CalendarEventCreateResolvedOrganizer{
			UserID:      organizer.ID,
			Email:       organizer.Email,
			DisplayName: eventEntOptionalDisplayName(organizer.DisplayName),
		},
		Attendees:       make([]CalendarEventCreateResolvedAttendee, len(request.Attendees)),
		ScheduledCallID: eventEntUUIDPointer(request.ScheduledCallID),
	}

	for index, candidate := range request.Attendees {
		attendee := CalendarEventCreateResolvedAttendee{
			Email:       candidate.Email,
			DisplayName: eventEntStringPointer(candidate.DisplayName),
			Role:        candidate.Role,
		}
		if candidate.UserID != nil {
			member, err := backend.activeOrganizationMemberForUpdate(ctx, *candidate.UserID, scope.OrganizationID())
			if err != nil {
				return CalendarEventCreateAuthorization{}, err
			}
			memberID := member.ID
			attendee.UserID = &memberID
			attendee.Email = member.Email
			attendee.DisplayName = eventEntOptionalDisplayName(member.DisplayName)
		}
		resolved.Attendees[index] = attendee
	}
	return resolved, nil
}

// ResolveMeetingCreateAuthorization locks only rows already predicated by the
// validated organization, active project/channel and exact membership set.
func (backend *eventEntTransactionBackend) ResolveMeetingCreateAuthorization(ctx context.Context, scope EventMutationScope, request MeetingCreateAuthorizationRequest) (MeetingCreateAuthorization, error) {
	if backend == nil || backend.tx == nil || scope.ActorID() == uuid.Nil || scope.OrganizationID() == uuid.Nil ||
		scope.EventOperation() != EventMutationMeetingCreate || !request.valid() {
		return MeetingCreateAuthorization{}, ErrEventForbidden
	}
	participants := make([]uuid.UUID, 0, len(request.AttendeeUserIDs)+1)
	participants = append(participants, scope.ActorID())
	participants = append(participants, request.AttendeeUserIDs...)
	sort.Slice(participants, func(left, right int) bool { return participants[left].String() < participants[right].String() })
	if err := backend.acquireCalendarAvailabilityAdvisoryLocks(ctx, participants); err != nil {
		return MeetingCreateAuthorization{}, err
	}
	return backend.resolveMeetingAuthorization(ctx, scope, request)
}

func (backend *eventEntTransactionBackend) resolveMeetingAuthorization(ctx context.Context, scope EventMutationScope, request MeetingCreateAuthorizationRequest) (MeetingCreateAuthorization, error) {
	if backend == nil || backend.tx == nil || scope.ActorID() == uuid.Nil || scope.OrganizationID() == uuid.Nil || !request.valid() {
		return MeetingCreateAuthorization{}, ErrEventForbidden
	}
	participantIDs := make([]uuid.UUID, 0, len(request.AttendeeUserIDs)+1)
	participantIDs = append(participantIDs, scope.ActorID())
	participantIDs = append(participantIDs, request.AttendeeUserIDs...)
	if _, err := backend.tx.Project.Query().
		Where(
			project.IDEQ(request.ProjectID),
			project.StatusEQ(project.StatusActive),
			project.DeletedAtIsNil(),
			project.HasOrganizationWith(organization.IDEQ(scope.OrganizationID())),
			project.HasMembersWith(projectmember.HasUserWith(
				user.IDEQ(scope.ActorID()),
				user.StatusEQ(user.StatusActive),
				user.HasOrganizationWith(organization.IDEQ(scope.OrganizationID())),
			)),
		).
		ForUpdate().
		Only(ctx); err != nil {
		return MeetingCreateAuthorization{}, calendarEntAuthorizationError(err)
	}
	if _, err := backend.tx.Channel.Query().
		Where(
			channel.IDEQ(request.ChannelID),
			channel.ProjectIDEQ(request.ProjectID),
			channel.DeletedAtIsNil(),
			channel.HasMembersWith(
				channelmember.UserIDEQ(scope.ActorID()),
				channelmember.HasUserWith(
					user.StatusEQ(user.StatusActive),
					user.HasOrganizationWith(organization.IDEQ(scope.OrganizationID())),
				),
			),
		).
		ForUpdate().
		Only(ctx); err != nil {
		return MeetingCreateAuthorization{}, calendarEntAuthorizationError(err)
	}

	memberships, err := backend.tx.ChannelMember.Query().
		Where(
			channelmember.ChannelIDEQ(request.ChannelID),
			channelmember.UserIDEQ(scope.ActorID()),
			channelmember.HasChannelWith(
				channel.IDEQ(request.ChannelID),
				channel.ProjectIDEQ(request.ProjectID),
				channel.DeletedAtIsNil(),
				channel.HasProjectWith(
					project.IDEQ(request.ProjectID),
					project.StatusEQ(project.StatusActive),
					project.DeletedAtIsNil(),
					project.HasOrganizationWith(organization.IDEQ(scope.OrganizationID())),
				),
			),
			channelmember.HasUserWith(
				user.StatusEQ(user.StatusActive),
				user.HasOrganizationWith(organization.IDEQ(scope.OrganizationID())),
			),
		).
		Order(ent.Asc(channelmember.FieldID)).
		ForUpdate().
		Limit(2).
		All(ctx)
	if err != nil {
		return MeetingCreateAuthorization{}, calendarEntSafeError(err)
	}
	if len(memberships) != 1 || memberships[0] == nil || memberships[0].ChannelID != request.ChannelID || memberships[0].UserID != scope.ActorID() {
		return MeetingCreateAuthorization{}, ErrEventForbidden
	}

	projectMemberships, err := backend.tx.ProjectMember.Query().
		Where(
			projectmember.HasProjectWith(
				project.IDEQ(request.ProjectID),
				project.StatusEQ(project.StatusActive),
				project.DeletedAtIsNil(),
				project.HasOrganizationWith(organization.IDEQ(scope.OrganizationID())),
			),
			projectmember.HasUserWith(
				user.IDIn(participantIDs...),
				user.StatusEQ(user.StatusActive),
				user.HasOrganizationWith(organization.IDEQ(scope.OrganizationID())),
			),
		).
		WithUser().
		Order(ent.Asc(projectmember.FieldID)).
		ForUpdate().
		Limit(meetingCreateMaxAttendees + 2).
		All(ctx)
	if err != nil {
		return MeetingCreateAuthorization{}, calendarEntSafeError(err)
	}
	if len(projectMemberships) != len(participantIDs) {
		return MeetingCreateAuthorization{}, ErrEventForbidden
	}
	seenProjectMembers := make(map[uuid.UUID]struct{}, len(projectMemberships))
	for _, membership := range projectMemberships {
		if membership == nil || membership.ID == uuid.Nil || membership.Edges.User == nil || membership.Edges.User.ID == uuid.Nil {
			return MeetingCreateAuthorization{}, ErrEventForbidden
		}
		memberID := membership.Edges.User.ID
		if _, duplicate := seenProjectMembers[memberID]; duplicate {
			return MeetingCreateAuthorization{}, ErrEventForbidden
		}
		seenProjectMembers[memberID] = struct{}{}
	}
	for _, participantID := range participantIDs {
		if _, found := seenProjectMembers[participantID]; !found {
			return MeetingCreateAuthorization{}, ErrEventForbidden
		}
	}

	users, err := backend.tx.User.Query().
		Where(
			user.IDIn(participantIDs...),
			user.StatusEQ(user.StatusActive),
			user.HasOrganizationWith(organization.IDEQ(scope.OrganizationID())),
		).
		Order(ent.Asc(user.FieldID)).
		ForUpdate().
		Limit(meetingCreateMaxAttendees + 2).
		All(ctx)
	if err != nil {
		return MeetingCreateAuthorization{}, calendarEntSafeError(err)
	}
	if len(users) != len(participantIDs) {
		return MeetingCreateAuthorization{}, ErrEventForbidden
	}
	resolved := make(map[uuid.UUID]MeetingCreateMember, len(users))
	for _, member := range users {
		if member == nil || member.ID == uuid.Nil || !validCalendarCreateEmail(member.Email) {
			return MeetingCreateAuthorization{}, ErrEventForbidden
		}
		resolved[member.ID] = MeetingCreateMember{
			UserID:      member.ID,
			Email:       normalizeCalendarCreateEmail(member.Email),
			DisplayName: eventEntOptionalDisplayName(member.DisplayName),
		}
	}
	organizer, found := resolved[scope.ActorID()]
	if !found {
		return MeetingCreateAuthorization{}, ErrEventForbidden
	}
	authorization := MeetingCreateAuthorization{
		ProjectID: request.ProjectID,
		ChannelID: request.ChannelID,
		Organizer: organizer,
		Attendees: make([]MeetingCreateMember, len(request.AttendeeUserIDs)),
	}
	for index, attendeeID := range request.AttendeeUserIDs {
		attendee, found := resolved[attendeeID]
		if !found {
			return MeetingCreateAuthorization{}, ErrEventForbidden
		}
		authorization.Attendees[index] = attendee
	}
	if !authorization.valid() {
		return MeetingCreateAuthorization{}, ErrEventForbidden
	}
	return authorization, nil
}

// PersistMeetingScheduledCall creates the forward event link first, then fills
// the immutable reverse link with one parameterized statement on this tx.
func (backend *eventEntTransactionBackend) PersistMeetingScheduledCall(ctx context.Context, scope EventMutationScope, command MeetingScheduledCallCommand) (MeetingScheduledCallPersistenceResult, error) {
	if backend == nil || backend.tx == nil || scope.EventOperation() != EventMutationMeetingCreate || !command.valid() ||
		command.organizationID != scope.OrganizationID() || command.creatorID != scope.ActorID() {
		return MeetingScheduledCallPersistenceResult{}, ErrEventForbidden
	}
	create := backend.tx.ScheduledCall.Create().
		SetID(command.callID).
		SetTitle(command.title).
		SetScheduledAt(command.startsAt).
		SetDurationMinutes(command.DurationMinutes()).
		SetChannelID(command.channelID).
		SetProjectID(command.projectID).
		SetCreatedBy(command.creatorID).
		SetCalendarEventID(command.eventID).
		AddParticipantIDs(command.attendeeUserIDs...)
	if command.description != "" {
		create.SetDescription(command.description)
	}
	if command.rrule != "" {
		create.SetRrule(command.rrule)
	}
	call, err := create.Save(ctx)
	if err != nil {
		return MeetingScheduledCallPersistenceResult{}, calendarEntSafeError(err)
	}
	if call == nil || call.ID != command.callID || call.CalendarEventID == nil || *call.CalendarEventID != command.eventID ||
		call.ProjectID != command.projectID || call.ChannelID != command.channelID || call.CreatedBy != command.creatorID ||
		!call.ScheduledAt.Equal(command.startsAt) || call.DurationMinutes != command.DurationMinutes() || call.Version < 1 {
		return MeetingScheduledCallPersistenceResult{}, ErrEventInternal
	}
	driver := backend.tx.Driver()
	if driver == nil {
		return MeetingScheduledCallPersistenceResult{}, ErrEventInternal
	}
	var linkResult sql.Result
	if err := driver.Exec(
		ctx,
		"UPDATE calendar_events SET scheduled_call_id = $1 WHERE id = $2 AND organization_id = $3 AND calendar_owner_id = $4 AND organizer_id = $4 AND scheduled_call_id IS NULL",
		[]any{command.callID, command.eventID, command.organizationID, command.creatorID},
		&linkResult,
	); err != nil {
		return MeetingScheduledCallPersistenceResult{}, calendarEntSafeError(err)
	}
	affected, err := linkResult.RowsAffected()
	if err != nil || affected != 1 {
		return MeetingScheduledCallPersistenceResult{}, ErrEventInternal
	}
	return NewMeetingScheduledCallPersistenceResult(call.ID, uint(call.Version))
}

type meetingEntMutationRows struct {
	event     *ent.CalendarEvent
	call      *ent.ScheduledCall
	attendees []*ent.CalendarAttendee
	snapshot  MeetingMutationSnapshot
}

func (backend *eventEntTransactionBackend) ResolveMeetingUpdateAuthorization(ctx context.Context, scope EventMutationScope, request MeetingUpdateAuthorizationRequest) (MeetingUpdateAuthorization, error) {
	if backend == nil || backend.tx == nil || scope.EventOperation() != EventMutationMeetingUpdate || !request.valid() {
		return MeetingUpdateAuthorization{}, ErrEventForbidden
	}
	rows, err := backend.lockMeetingMutationRows(ctx, scope, request.EventID)
	if err != nil {
		return MeetingUpdateAuthorization{}, err
	}
	if request.Desired.ProjectID != rows.snapshot.ProjectID || request.Desired.ChannelID != rows.snapshot.ChannelID {
		return MeetingUpdateAuthorization{}, newMeetingMutationConflict(MeetingMutationConflictScopeImmutable)
	}
	participants, ok := meetingMutationLockParticipants(rows.snapshot, request.Desired.AttendeeUserIDs)
	if !ok {
		return MeetingUpdateAuthorization{}, ErrEventInternal
	}
	if err := backend.acquireCalendarAvailabilityAdvisoryLocks(ctx, participants); err != nil {
		return MeetingUpdateAuthorization{}, err
	}
	members, err := backend.lockMeetingMutationUsers(ctx, scope.OrganizationID(), participants)
	if err != nil {
		return MeetingUpdateAuthorization{}, err
	}
	if !meetingMutationCurrentMembersMatch(rows.snapshot, members) {
		return MeetingUpdateAuthorization{}, ErrEventForbidden
	}
	desired, err := backend.resolveMeetingAuthorization(ctx, scope, request.Desired)
	if err != nil {
		return MeetingUpdateAuthorization{}, err
	}
	authorization := MeetingUpdateAuthorization{Current: rows.snapshot, Desired: desired}
	if !authorization.valid() {
		return MeetingUpdateAuthorization{}, ErrEventInternal
	}
	return authorization, nil
}

func (backend *eventEntTransactionBackend) ResolveMeetingCancelAuthorization(ctx context.Context, scope EventMutationScope, eventID uuid.UUID) (MeetingCancelAuthorization, error) {
	if backend == nil || backend.tx == nil || scope.EventOperation() != EventMutationMeetingCancel || eventID == uuid.Nil {
		return MeetingCancelAuthorization{}, ErrEventForbidden
	}
	rows, err := backend.lockMeetingMutationRows(ctx, scope, eventID)
	if err != nil {
		return MeetingCancelAuthorization{}, err
	}
	participants, ok := meetingMutationLockParticipants(rows.snapshot, nil)
	if !ok {
		return MeetingCancelAuthorization{}, ErrEventInternal
	}
	if err := backend.acquireCalendarAvailabilityAdvisoryLocks(ctx, participants); err != nil {
		return MeetingCancelAuthorization{}, err
	}
	members, err := backend.lockMeetingMutationUsers(ctx, scope.OrganizationID(), participants)
	if err != nil {
		return MeetingCancelAuthorization{}, err
	}
	if !meetingMutationCurrentMembersMatch(rows.snapshot, members) {
		return MeetingCancelAuthorization{}, ErrEventForbidden
	}
	authorization := MeetingCancelAuthorization{Current: rows.snapshot}
	if !authorization.valid() {
		return MeetingCancelAuthorization{}, ErrEventInternal
	}
	return authorization, nil
}

func (backend *eventEntTransactionBackend) lockMeetingMutationRows(ctx context.Context, scope EventMutationScope, eventID uuid.UUID) (meetingEntMutationRows, error) {
	if backend == nil || backend.tx == nil || eventID == uuid.Nil || scope.ActorID() == uuid.Nil || scope.OrganizationID() == uuid.Nil {
		return meetingEntMutationRows{}, ErrEventForbidden
	}
	event, err := backend.tx.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(eventID),
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.CalendarOwnerIDEQ(scope.ActorID()),
			calendarevent.OrganizerIDEQ(scope.ActorID()),
			calendarevent.SourceEQ(calendarevent.Source(CalendarEventSourceHub)),
			calendarevent.StatusIn(calendarevent.Status(CalendarEventStatusConfirmed), calendarevent.Status(CalendarEventStatusTentative)),
			calendarevent.AllDay(false),
			calendarevent.DeletedAtIsNil(),
			calendarevent.ScheduledCallIDNotNil(),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return meetingEntMutationRows{}, calendarEntMeetingMutationAuthorizationError(err)
	}
	if event == nil || event.ScheduledCallID == nil || event.OrganizerID == nil || *event.OrganizerID != scope.ActorID() ||
		event.StartsAt == nil || event.EndsAt == nil || event.StartDate != nil || event.EndDateExclusive != nil ||
		event.Location != "" || event.ReminderMinutes != nil || len(event.Exdates) != 0 ||
		event.OrganizerEmail != normalizeCalendarCreateEmail(event.OrganizerEmail) {
		return meetingEntMutationRows{}, ErrEventInternal
	}
	call, err := backend.tx.ScheduledCall.Query().
		Where(
			scheduledcall.IDEQ(*event.ScheduledCallID),
			scheduledcall.CalendarEventIDEQ(event.ID),
			scheduledcall.CreatedByEQ(scope.ActorID()),
			scheduledcall.IsActive(true),
			scheduledcall.DeletedAtIsNil(),
		).
		WithParticipants().
		ForUpdate().
		Only(ctx)
	if err != nil {
		return meetingEntMutationRows{}, calendarEntMeetingMutationAuthorizationError(err)
	}
	if call == nil || call.CalendarEventID == nil || *call.CalendarEventID != event.ID || call.ID != *event.ScheduledCallID ||
		call.ProjectID == uuid.Nil || call.ChannelID == uuid.Nil || call.CreatedBy != scope.ActorID() {
		return meetingEntMutationRows{}, ErrEventForbidden
	}
	attendees, err := backend.tx.CalendarAttendee.Query().
		Where(
			calendarattendee.CalendarEventIDEQ(event.ID),
			calendarattendee.RecurrenceIDIsNil(),
		).
		Order(ent.Asc(calendarattendee.FieldID)).
		ForUpdate().
		Limit(meetingMutationMaxStoredAttendees + 1).
		All(ctx)
	if err != nil {
		return meetingEntMutationRows{}, calendarEntSafeError(err)
	}
	if len(attendees) == 0 || len(attendees) > meetingMutationMaxStoredAttendees {
		return meetingEntMutationRows{}, ErrEventInternal
	}
	hasMaterializedExceptions := false
	if scope.EventOperation() == EventMutationMeetingUpdate {
		exceptions, err := backend.tx.CalendarEventException.Query().
			Where(calendareventexception.CalendarEventIDEQ(event.ID)).
			Order(ent.Asc(calendareventexception.FieldRecurrenceID), ent.Asc(calendareventexception.FieldID)).
			ForUpdate().
			Limit(calendarAvailabilityReservationMaxOccurrences + 1).
			All(ctx)
		if err != nil {
			return meetingEntMutationRows{}, calendarEntSafeError(err)
		}
		if len(exceptions) > calendarAvailabilityReservationMaxOccurrences {
			return meetingEntMutationRows{}, ErrAvailabilityLimitExceeded
		}
		hasMaterializedExceptions = len(exceptions) != 0
	}
	eventVersion, eventVersionOK := eventEntUint(event.Version)
	callVersion, callVersionOK := eventEntUint(call.Version)
	sequence, sequenceOK := eventEntUint(event.Sequence)
	startsAt, startsOK := canonicalAvailabilityTime(*event.StartsAt)
	endsAt, endsOK := canonicalAvailabilityTime(*event.EndsAt)
	callStartsAt, callStartsOK := canonicalAvailabilityTime(call.ScheduledAt)
	callEndsAt, callEndsOK := canonicalAvailabilityTime(callStartsAt.Add(time.Duration(call.DurationMinutes) * time.Minute))
	rule := ""
	if event.Rrule != nil {
		rule = *event.Rrule
	}
	if !eventVersionOK || !callVersionOK || !sequenceOK || !startsOK || !endsOK || !callStartsOK || !callEndsOK ||
		call.DurationMinutes < 1 || call.DurationMinutes > int(meetingCreateMaxDuration/time.Minute) ||
		call.Title != event.Title || call.Description != event.Description || call.Rrule != rule ||
		!meetingMutationCallTemporalMatches(event.ID, startsAt, endsAt, event.Timezone, rule, callStartsAt, callEndsAt) {
		return meetingEntMutationRows{}, ErrEventInternal
	}
	stored := make([]MeetingStoredAttendee, len(attendees))
	for index, attendee := range attendees {
		if attendee == nil || attendee.UserID == nil || *attendee.UserID == uuid.Nil ||
			attendee.NormalizedEmail != normalizeCalendarCreateEmail(attendee.NormalizedEmail) {
			return meetingEntMutationRows{}, ErrEventInternal
		}
		version, ok := eventEntUint(attendee.Version)
		if !ok {
			return meetingEntMutationRows{}, ErrEventInternal
		}
		displayName := ""
		if attendee.DisplayName != nil {
			displayName = *attendee.DisplayName
		}
		stored[index] = MeetingStoredAttendee{
			AttendeeID: attendee.ID, UserID: *attendee.UserID, Email: normalizeCalendarCreateEmail(attendee.NormalizedEmail),
			DisplayName: displayName, Role: CalendarEventAttendeeRole(attendee.Role), Status: RSVPStatus(attendee.RsvpStatus),
			Version: version, RemovedAt: eventEntTimePointer(attendee.RemovedAt),
		}
	}
	snapshot := MeetingMutationSnapshot{
		EventID: event.ID, ScheduledCallID: call.ID, OrganizationID: event.OrganizationID, OwnerID: event.CalendarOwnerID,
		UID: event.UID, Title: event.Title, Description: event.Description, ProjectID: call.ProjectID, ChannelID: call.ChannelID,
		StartsAt: startsAt, EndsAt: endsAt, Timezone: event.Timezone, RRULE: rule, Sequence: sequence,
		EventVersion: eventVersion, ScheduledCallVersion: callVersion, HasMaterializedExceptions: hasMaterializedExceptions,
		OrganizerEmail: normalizeCalendarCreateEmail(event.OrganizerEmail),
		Attendees: stored,
	}
	if !snapshot.valid() || !meetingMutationCallParticipantsMatch(snapshot, call.Edges.Participants) {
		return meetingEntMutationRows{}, ErrEventInternal
	}
	return meetingEntMutationRows{event: event, call: call, attendees: attendees, snapshot: snapshot}, nil
}

func meetingMutationCallTemporalMatches(eventID uuid.UUID, startsAt, endsAt time.Time, timezone, rule string, callStartsAt, callEndsAt time.Time) bool {
	if eventID == uuid.Nil || callEndsAt.Sub(callStartsAt) != endsAt.Sub(startsAt) {
		return false
	}
	if rule == "" {
		return startsAt.Equal(callStartsAt) && endsAt.Equal(callEndsAt)
	}
	occurrences, _, err := availabilityTimedEventOccurrences(
		CalendarEvent{StartsAt: startsAt, EndsAt: endsAt, Timezone: timezone}, rule, nil, nil, eventID,
	)
	if err != nil || len(occurrences) == 0 || len(occurrences) > calendarAvailabilityReservationMaxOccurrences {
		return false
	}
	for _, occurrence := range occurrences {
		if occurrence.from.Equal(callStartsAt) && occurrence.to.Equal(callEndsAt) {
			return true
		}
	}
	return false
}

func meetingMutationCallParticipantsMatch(snapshot MeetingMutationSnapshot, participants []*ent.User) bool {
	actual := make([]uuid.UUID, len(participants))
	for index, participant := range participants {
		if participant == nil || participant.ID == uuid.Nil {
			return false
		}
		actual[index] = participant.ID
	}
	sort.Slice(actual, func(left, right int) bool { return actual[left].String() < actual[right].String() })
	expected := snapshot.activeInviteeUserIDs()
	if len(actual) != len(expected) {
		return false
	}
	for index := range actual {
		if actual[index] != expected[index] || index > 0 && actual[index] == actual[index-1] {
			return false
		}
	}
	return true
}

func meetingMutationLockParticipants(snapshot MeetingMutationSnapshot, desired []uuid.UUID) ([]uuid.UUID, bool) {
	if !snapshot.valid() {
		return nil, false
	}
	participants := make([]uuid.UUID, 0, len(snapshot.Attendees)+len(desired))
	participants = append(participants, snapshot.OwnerID)
	participants = append(participants, snapshot.activeInviteeUserIDs()...)
	participants = append(participants, desired...)
	sort.Slice(participants, func(left, right int) bool { return participants[left].String() < participants[right].String() })
	unique := participants[:0]
	for _, participantID := range participants {
		if participantID == uuid.Nil {
			return nil, false
		}
		if len(unique) == 0 || unique[len(unique)-1] != participantID {
			unique = append(unique, participantID)
		}
	}
	if len(unique) == 0 || len(unique) > meetingCreateMaxAttendees*2+1 {
		return nil, false
	}
	return append([]uuid.UUID(nil), unique...), true
}

func (backend *eventEntTransactionBackend) lockMeetingMutationUsers(ctx context.Context, organizationID uuid.UUID, participants []uuid.UUID) (map[uuid.UUID]*ent.User, error) {
	if backend == nil || backend.tx == nil || organizationID == uuid.Nil || len(participants) == 0 || len(participants) > meetingCreateMaxAttendees*2+1 {
		return nil, ErrEventInternal
	}
	rows, err := backend.tx.User.Query().
		Where(
			user.IDIn(participants...),
			user.HasOrganizationWith(organization.IDEQ(organizationID)),
		).
		Order(ent.Asc(user.FieldID)).
		ForUpdate().
		Limit(len(participants) + 1).
		All(ctx)
	if err != nil {
		return nil, calendarEntSafeError(err)
	}
	if len(rows) != len(participants) {
		return nil, ErrEventForbidden
	}
	resolved := make(map[uuid.UUID]*ent.User, len(rows))
	for _, row := range rows {
		if row == nil || row.ID == uuid.Nil || !validCalendarCreateEmail(row.Email) {
			return nil, ErrEventForbidden
		}
		if _, duplicate := resolved[row.ID]; duplicate {
			return nil, ErrEventForbidden
		}
		resolved[row.ID] = row
	}
	return resolved, nil
}

func meetingMutationCurrentMembersMatch(snapshot MeetingMutationSnapshot, members map[uuid.UUID]*ent.User) bool {
	if !snapshot.valid() || len(members) == 0 {
		return false
	}
	for _, attendee := range snapshot.Attendees {
		if attendee.RemovedAt != nil {
			continue
		}
		member, found := members[attendee.UserID]
		if !found || member == nil || normalizeCalendarCreateEmail(member.Email) != attendee.Email {
			return false
		}
	}
	return true
}

func (backend *eventEntTransactionBackend) PersistMeetingUpdate(ctx context.Context, scope EventMutationScope, command MeetingUpdatePersistenceCommand) (MeetingMutationPersistenceResult, error) {
	if backend == nil || backend.tx == nil || backend.now == nil || scope.EventOperation() != EventMutationMeetingUpdate || !command.valid() ||
		command.current.OrganizationID != scope.OrganizationID() || command.current.OwnerID != scope.ActorID() {
		return MeetingMutationPersistenceResult{}, ErrEventForbidden
	}
	rows, err := backend.lockMeetingMutationRows(ctx, scope, command.current.EventID)
	if err != nil {
		return MeetingMutationPersistenceResult{}, err
	}
	if !meetingMutationSnapshotsEqual(rows.snapshot, command.current) {
		return MeetingMutationPersistenceResult{}, ErrEventConflict
	}
	eventVersion, callVersion, sequence, ok := meetingMutationUpdateVersions(command.current, command.plan)
	if !ok {
		return MeetingMutationPersistenceResult{}, ErrEventInternal
	}
	expectedEventVersion, _ := eventEntInt(command.current.EventVersion)
	eventUpdate := backend.tx.CalendarEvent.UpdateOneID(command.current.EventID).
		Where(
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.CalendarOwnerIDEQ(scope.ActorID()),
			calendarevent.OrganizerIDEQ(scope.ActorID()),
			calendarevent.VersionEQ(expectedEventVersion),
			calendarevent.StatusIn(calendarevent.Status(CalendarEventStatusConfirmed), calendarevent.Status(CalendarEventStatusTentative)),
			calendarevent.DeletedAtIsNil(),
		).
		SetTitle(command.desired.Title).
		SetDescription(command.desired.Description).
		SetStartsAt(command.desired.StartsAt).
		SetEndsAt(command.desired.EndsAt).
		SetTimezone(command.desired.Timezone).
		SetOrganizerID(scope.ActorID()).
		SetOrganizerEmail(normalizeCalendarCreateEmail(command.authorization.Organizer.Email)).
		SetSequence(sequence).
		SetVersion(eventVersion)
	if command.desired.RRULE == "" {
		eventUpdate.ClearRrule()
	} else {
		eventUpdate.SetRrule(command.desired.RRULE)
	}
	if err := eventUpdate.Exec(ctx); err != nil {
		return MeetingMutationPersistenceResult{}, calendarEntCASFailure(err)
	}
	now, ok := canonicalAvailabilityTime(backend.now().UTC().Truncate(time.Second))
	if !ok {
		return MeetingMutationPersistenceResult{}, ErrEventInternal
	}
	activeIDs, removedIDs, err := backend.persistMeetingUpdateAttendees(ctx, rows, command, now)
	if err != nil {
		return MeetingMutationPersistenceResult{}, err
	}
	if command.plan.reservationsChanged {
		if err := backend.ReconcileCalendarAvailabilityReservations(ctx, scope, command.reservations); err != nil {
			return MeetingMutationPersistenceResult{}, err
		}
	}
	if command.plan.callChanged {
		expectedCallVersion, _ := eventEntInt(command.current.ScheduledCallVersion)
		callUpdate := backend.tx.ScheduledCall.UpdateOneID(command.current.ScheduledCallID).
			Where(
				scheduledcall.CalendarEventIDEQ(command.current.EventID),
				scheduledcall.CreatedByEQ(scope.ActorID()),
				scheduledcall.VersionEQ(expectedCallVersion),
				scheduledcall.DeletedAtIsNil(),
			).
			SetTitle(command.desired.Title).
			SetDescription(command.desired.Description).
			SetScheduledAt(command.desired.StartsAt).
			SetDurationMinutes(int(command.desired.EndsAt.Sub(command.desired.StartsAt) / time.Minute)).
			SetProjectID(command.desired.ProjectID).
			SetChannelID(command.desired.ChannelID).
			ClearParticipants().
			AddParticipantIDs(append([]uuid.UUID(nil), command.desired.AttendeeUserIDs...)...).
			SetIsActive(true).
			ClearDeletedAt().
			SetVersion(callVersion)
		if command.desired.RRULE == "" {
			callUpdate.ClearRrule()
		} else {
			callUpdate.SetRrule(command.desired.RRULE)
		}
		if err := callUpdate.Exec(ctx); err != nil {
			return MeetingMutationPersistenceResult{}, calendarEntCASFailure(err)
		}
	}
	return newMeetingMutationPersistenceResult(uint(eventVersion), uint(callVersion), uint(sequence), activeIDs, removedIDs)
}

func (backend *eventEntTransactionBackend) persistMeetingUpdateAttendees(ctx context.Context, rows meetingEntMutationRows, command MeetingUpdatePersistenceCommand, now time.Time) ([]uuid.UUID, []uuid.UUID, error) {
	existing := make(map[uuid.UUID]*ent.CalendarAttendee, len(rows.attendees))
	for _, attendee := range rows.attendees {
		if attendee == nil || attendee.UserID == nil || *attendee.UserID == uuid.Nil {
			return nil, nil, ErrEventInternal
		}
		existing[*attendee.UserID] = attendee
	}
	organizer := existing[command.current.OwnerID]
	if organizer == nil || organizer.RemovedAt != nil || organizer.Version < 1 || organizer.Version == int(^uint(0)>>1) ||
		organizer.RsvpStatus != calendarattendee.RsvpStatus(RSVPAccepted) {
		return nil, nil, ErrEventInternal
	}
	desired := make(map[uuid.UUID]MeetingCreateMember, len(command.authorization.Attendees))
	activeIDs := make([]uuid.UUID, 0, len(command.authorization.Attendees))
	for _, member := range command.authorization.Attendees {
		desired[member.UserID] = member
		row := existing[member.UserID]
		if row == nil {
			memberID := member.UserID
			recordID := calendarCreateUUID("hub.calendar.meeting.attendee.id.v1", command.current.EventID[:], memberID[:])
			create := backend.tx.CalendarAttendee.Create().
				SetID(recordID).
				SetCalendarEventID(command.current.EventID).
				SetUserID(memberID).
				SetNormalizedEmail(normalizeCalendarCreateEmail(member.Email)).
				SetRole(calendarattendee.Role(CalendarEventAttendeeRoleRequired)).
				SetRsvpStatus(calendarattendee.RsvpStatus(RSVPNeedsAction)).
				SetVersion(1)
			if member.DisplayName != nil {
				create.SetDisplayName(*member.DisplayName)
			}
			if _, err := create.Save(ctx); err != nil {
				return nil, nil, calendarEntSafeError(err)
			}
			activeIDs = append(activeIDs, recordID)
			continue
		}
		if row.Version < 1 || row.Version == int(^uint(0)>>1) {
			return nil, nil, ErrEventInternal
		}
		if row.RemovedAt != nil {
			update := backend.tx.CalendarAttendee.UpdateOneID(row.ID).
				Where(
					calendarattendee.CalendarEventIDEQ(command.current.EventID),
					calendarattendee.RecurrenceIDIsNil(),
					calendarattendee.VersionEQ(row.Version),
				).
				SetUserID(member.UserID).
				SetNormalizedEmail(normalizeCalendarCreateEmail(member.Email)).
				SetRole(calendarattendee.Role(CalendarEventAttendeeRoleRequired)).
				SetRsvpStatus(calendarattendee.RsvpStatus(RSVPNeedsAction)).
				ClearRespondedAt().
				ClearRemovedAt().
				SetVersion(row.Version + 1)
			if member.DisplayName == nil {
				update.ClearDisplayName()
			} else {
				update.SetDisplayName(*member.DisplayName)
			}
			if err := update.Exec(ctx); err != nil {
				return nil, nil, calendarEntCASFailure(err)
			}
		} else if calendarEventUpdateResetsNonOrganizers(command.plan.policy) {
			if err := backend.tx.CalendarAttendee.UpdateOneID(row.ID).
				Where(
					calendarattendee.CalendarEventIDEQ(command.current.EventID),
					calendarattendee.RecurrenceIDIsNil(),
					calendarattendee.RemovedAtIsNil(),
					calendarattendee.VersionEQ(row.Version),
				).
				SetRsvpStatus(calendarattendee.RsvpStatus(RSVPNeedsAction)).
				ClearRespondedAt().
				SetVersion(row.Version + 1).
				Exec(ctx); err != nil {
				return nil, nil, calendarEntCASFailure(err)
			}
		}
		activeIDs = append(activeIDs, row.ID)
	}
	removedIDs := make([]uuid.UUID, 0, len(rows.attendees))
	for _, attendee := range rows.attendees {
		if attendee.UserID == nil || *attendee.UserID == command.current.OwnerID || attendee.RemovedAt != nil {
			continue
		}
		if _, retained := desired[*attendee.UserID]; retained {
			continue
		}
		if attendee.Version < 1 || attendee.Version == int(^uint(0)>>1) {
			return nil, nil, ErrEventInternal
		}
		if err := backend.tx.CalendarAttendee.UpdateOneID(attendee.ID).
			Where(
				calendarattendee.CalendarEventIDEQ(command.current.EventID),
				calendarattendee.RecurrenceIDIsNil(),
				calendarattendee.RemovedAtIsNil(),
				calendarattendee.VersionEQ(attendee.Version),
			).
			SetRemovedAt(now).
			SetVersion(attendee.Version + 1).
			Exec(ctx); err != nil {
			return nil, nil, calendarEntCASFailure(err)
		}
		removedIDs = append(removedIDs, attendee.ID)
	}
	sort.Slice(activeIDs, func(left, right int) bool { return activeIDs[left].String() < activeIDs[right].String() })
	sort.Slice(removedIDs, func(left, right int) bool { return removedIDs[left].String() < removedIDs[right].String() })
	return activeIDs, removedIDs, nil
}

func (backend *eventEntTransactionBackend) PersistMeetingCancel(ctx context.Context, scope EventMutationScope, command MeetingCancelPersistenceCommand) (MeetingMutationPersistenceResult, error) {
	if backend == nil || backend.tx == nil || backend.now == nil || scope.EventOperation() != EventMutationMeetingCancel || !command.valid() ||
		command.current.OrganizationID != scope.OrganizationID() || command.current.OwnerID != scope.ActorID() {
		return MeetingMutationPersistenceResult{}, ErrEventForbidden
	}
	rows, err := backend.lockMeetingMutationRows(ctx, scope, command.current.EventID)
	if err != nil {
		return MeetingMutationPersistenceResult{}, err
	}
	if !meetingMutationSnapshotsEqual(rows.snapshot, command.current) {
		return MeetingMutationPersistenceResult{}, ErrEventConflict
	}
	eventVersion, callVersion, sequence, ok := meetingMutationNextVersions(command.current)
	if !ok {
		return MeetingMutationPersistenceResult{}, ErrEventInternal
	}
	now, ok := canonicalAvailabilityTime(backend.now().UTC().Truncate(time.Second))
	if !ok {
		return MeetingMutationPersistenceResult{}, ErrEventInternal
	}
	expectedEventVersion, _ := eventEntInt(command.current.EventVersion)
	if err := backend.tx.CalendarEvent.UpdateOneID(command.current.EventID).
		Where(
			calendarevent.OrganizationIDEQ(scope.OrganizationID()),
			calendarevent.CalendarOwnerIDEQ(scope.ActorID()),
			calendarevent.OrganizerIDEQ(scope.ActorID()),
			calendarevent.VersionEQ(expectedEventVersion),
			calendarevent.StatusIn(calendarevent.Status(CalendarEventStatusConfirmed), calendarevent.Status(CalendarEventStatusTentative)),
			calendarevent.DeletedAtIsNil(),
		).
		SetStatus(calendarevent.Status(CalendarEventStatusCancelled)).
		SetSequence(sequence).
		SetVersion(eventVersion).
		Exec(ctx); err != nil {
		return MeetingMutationPersistenceResult{}, calendarEntCASFailure(err)
	}
	reservations, err := backend.tx.CalendarAvailabilityReservation.Query().
		Where(
			calendaravailabilityreservation.OrganizationIDEQ(scope.OrganizationID()),
			calendaravailabilityreservation.CalendarEventIDEQ(command.current.EventID),
			calendaravailabilityreservation.StateEQ(calendaravailabilityreservation.StateActive),
			calendaravailabilityreservation.ReleasedAtIsNil(),
		).
		Order(ent.Asc(calendaravailabilityreservation.FieldUserID), ent.Asc(calendaravailabilityreservation.FieldOccurrenceKey)).
		ForUpdate().
		Limit(calendarAvailabilityReservationMaxRows + 1).
		All(ctx)
	if err != nil {
		return MeetingMutationPersistenceResult{}, calendarEntSafeError(err)
	}
	if len(reservations) > calendarAvailabilityReservationMaxRows {
		return MeetingMutationPersistenceResult{}, ErrAvailabilityLimitExceeded
	}
	expectedParticipants := make(map[uuid.UUID]struct{}, len(command.current.Attendees))
	expectedParticipants[command.current.OwnerID] = struct{}{}
	for _, attendeeID := range command.current.activeInviteeUserIDs() {
		expectedParticipants[attendeeID] = struct{}{}
	}
	for _, reservation := range reservations {
		if reservation == nil || reservation.State != calendaravailabilityreservation.StateActive || reservation.ReleasedAt != nil {
			return MeetingMutationPersistenceResult{}, ErrAvailabilitySourceInvalid
		}
		if _, expected := expectedParticipants[reservation.UserID]; !expected {
			return MeetingMutationPersistenceResult{}, ErrAvailabilitySourceInvalid
		}
		startsAt, endsAt, valid := canonicalCalendarAvailabilityReservationInterval(reservation.StartsAt, reservation.EndsAt)
		if !valid || reservation.OccurrenceKey != startsAt.Format(time.RFC3339Nano) || !endsAt.Equal(reservation.EndsAt) {
			return MeetingMutationPersistenceResult{}, ErrAvailabilitySourceInvalid
		}
	}
	if err := backend.releaseObsoleteCalendarAvailabilityReservations(ctx, reservations, map[calendarAvailabilityReservationIdentity]CalendarAvailabilityReservationOccurrence{}, now); err != nil {
		return MeetingMutationPersistenceResult{}, err
	}
	expectedCallVersion, _ := eventEntInt(command.current.ScheduledCallVersion)
	if err := backend.tx.ScheduledCall.UpdateOneID(command.current.ScheduledCallID).
		Where(
			scheduledcall.CalendarEventIDEQ(command.current.EventID),
			scheduledcall.CreatedByEQ(scope.ActorID()),
			scheduledcall.VersionEQ(expectedCallVersion),
			scheduledcall.DeletedAtIsNil(),
		).
		SetIsActive(false).
		SetDeletedAt(now).
		SetVersion(callVersion).
		Exec(ctx); err != nil {
		return MeetingMutationPersistenceResult{}, calendarEntCASFailure(err)
	}
	return newMeetingMutationPersistenceResult(uint(eventVersion), uint(callVersion), uint(sequence), command.current.activeInviteeRecordIDs(), nil)
}

func meetingMutationUpdateVersions(snapshot MeetingMutationSnapshot, plan meetingUpdatePlan) (int, int, int, bool) {
	if !snapshot.valid() || !plan.hasChanges() || plan.policy.VersionIncrement != 1 || plan.policy.SequenceIncrement > 1 ||
		snapshot.EventVersion > ^uint(0)-plan.policy.VersionIncrement || snapshot.Sequence > ^uint(0)-plan.policy.SequenceIncrement {
		return 0, 0, 0, false
	}
	callIncrement := uint(0)
	if plan.callChanged {
		callIncrement = 1
	}
	if snapshot.ScheduledCallVersion > ^uint(0)-callIncrement {
		return 0, 0, 0, false
	}
	eventVersion, eventOK := eventEntInt(snapshot.EventVersion + plan.policy.VersionIncrement)
	callVersion, callOK := eventEntInt(snapshot.ScheduledCallVersion + callIncrement)
	sequence, sequenceOK := eventEntInt(snapshot.Sequence + plan.policy.SequenceIncrement)
	return eventVersion, callVersion, sequence, eventOK && callOK && sequenceOK
}

func meetingMutationNextVersions(snapshot MeetingMutationSnapshot) (int, int, int, bool) {
	if !snapshot.valid() || snapshot.EventVersion == ^uint(0) || snapshot.ScheduledCallVersion == ^uint(0) || snapshot.Sequence == ^uint(0) {
		return 0, 0, 0, false
	}
	eventVersion, eventOK := eventEntInt(snapshot.EventVersion + 1)
	callVersion, callOK := eventEntInt(snapshot.ScheduledCallVersion + 1)
	sequence, sequenceOK := eventEntInt(snapshot.Sequence + 1)
	return eventVersion, callVersion, sequence, eventOK && callOK && sequenceOK
}

func calendarEntMeetingMutationAuthorizationError(err error) error {
	if ent.IsNotFound(err) {
		return ErrEventForbidden
	}
	return calendarEntSafeError(err)
}

func (backend *eventEntTransactionBackend) activeOrganizationMemberForUpdate(ctx context.Context, memberID, organizationID uuid.UUID) (*ent.User, error) {
	if memberID == uuid.Nil || organizationID == uuid.Nil {
		return nil, ErrEventForbidden
	}
	if _, err := backend.activeOrganizationForMemberForUpdate(ctx, memberID, organizationID); err != nil {
		return nil, err
	}
	member, err := backend.activeUserForUpdate(ctx, memberID)
	if err != nil {
		return nil, err
	}
	return member, nil
}

func (backend *eventEntTransactionBackend) activeOrganizationForActorForUpdate(ctx context.Context, actorID uuid.UUID) (*ent.Organization, error) {
	organization, err := backend.tx.Organization.Query().
		Where(organization.HasUsersWith(
			user.IDEQ(actorID),
			user.StatusEQ(user.StatusActive),
		)).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, calendarEntAuthorizationError(err)
	}
	return organization, nil
}

func (backend *eventEntTransactionBackend) activeOrganizationForMemberForUpdate(ctx context.Context, memberID, organizationID uuid.UUID) (*ent.Organization, error) {
	organization, err := backend.tx.Organization.Query().
		Where(
			organization.IDEQ(organizationID),
			organization.HasUsersWith(
				user.IDEQ(memberID),
				user.StatusEQ(user.StatusActive),
			),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, calendarEntAuthorizationError(err)
	}
	return organization, nil
}

func (backend *eventEntTransactionBackend) activeUserForUpdate(ctx context.Context, memberID uuid.UUID) (*ent.User, error) {
	member, err := backend.tx.User.Query().
		Where(
			user.IDEQ(memberID),
			user.StatusEQ(user.StatusActive),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, calendarEntAuthorizationError(err)
	}
	return member, nil
}

func calendarEntAuthorizationError(err error) error {
	if ent.IsNotFound(err) {
		return ErrEventForbidden
	}
	return calendarEntSafeError(err)
}

func calendarEntSafeError(err error) error {
	switch {
	case err == nil:
		return nil
	case calendarEntAvailabilityReservationExclusion(err):
		return ErrAvailabilitySlotUnavailable
	case errors.Is(err, ErrEventNotFound):
		return ErrEventNotFound
	case errors.Is(err, ErrEventForbidden):
		return ErrEventForbidden
	case errors.Is(err, ErrEventConflict):
		return ErrEventConflict
	case errors.Is(err, ErrEventInvalid):
		return ErrEventInvalid
	case errors.Is(err, ErrEventTransactionClosed):
		return ErrEventTransactionClosed
	case errors.Is(err, ErrEventInternal):
		return ErrEventInternal
	case ent.IsNotFound(err):
		return ErrEventNotFound
	case ent.IsConstraintError(err):
		return ErrEventConflict
	default:
		return ErrEventInvalid
	}
}

func calendarEntAvailabilityReservationExclusion(err error) bool {
	var postgresError *pq.Error
	return errors.As(err, &postgresError) && postgresError.Code == "23P01" && postgresError.Constraint == "calendar_availability_reservations_active_window_excl"
}

func eventEntUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func eventEntStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func eventEntOptionalDisplayName(value string) *string {
	if value == "" {
		return nil
	}
	return eventEntStringPointer(&value)
}
