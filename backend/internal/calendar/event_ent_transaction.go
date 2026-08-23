package calendar

import (
	"context"
	"errors"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/lib/pq"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/channel"
	"github.com/valerii/hub/backend/ent/channelmember"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/projectmember"
	"github.com/valerii/hub/backend/ent/user"
)

type eventEntTransactionBackend struct {
	tx  *ent.Tx
	now func() time.Time
}

var _ EventTransactionBackend = (*eventEntTransactionBackend)(nil)
var _ MeetingTransactionBackend = (*eventEntTransactionBackend)(nil)

func newEventEntTransactionBackend(transaction *ent.Tx, now func() time.Time) *eventEntTransactionBackend {
	return &eventEntTransactionBackend{tx: transaction, now: now}
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

	participantIDs := make([]uuid.UUID, 0, len(request.AttendeeUserIDs)+1)
	participantIDs = append(participantIDs, scope.ActorID())
	participantIDs = append(participantIDs, request.AttendeeUserIDs...)
	memberships, err := backend.tx.ChannelMember.Query().
		Where(
			channelmember.ChannelIDEQ(request.ChannelID),
			channelmember.UserIDIn(participantIDs...),
			channelmember.HasChannelWith(
				channel.IDEQ(request.ChannelID),
				channel.ProjectIDEQ(request.ProjectID),
				channel.DeletedAtIsNil(),
			),
			channelmember.HasUserWith(
				user.StatusEQ(user.StatusActive),
				user.HasOrganizationWith(organization.IDEQ(scope.OrganizationID())),
			),
		).
		Order(ent.Asc(channelmember.FieldUserID), ent.Asc(channelmember.FieldID)).
		ForUpdate().
		Limit(meetingCreateMaxAttendees + 2).
		All(ctx)
	if err != nil {
		return MeetingCreateAuthorization{}, calendarEntSafeError(err)
	}
	if len(memberships) != len(participantIDs) {
		return MeetingCreateAuthorization{}, ErrEventForbidden
	}
	seenMemberships := make(map[uuid.UUID]struct{}, len(memberships))
	for _, membership := range memberships {
		if membership == nil || membership.ChannelID != request.ChannelID || membership.UserID == uuid.Nil {
			return MeetingCreateAuthorization{}, ErrEventForbidden
		}
		if _, duplicate := seenMemberships[membership.UserID]; duplicate {
			return MeetingCreateAuthorization{}, ErrEventForbidden
		}
		seenMemberships[membership.UserID] = struct{}{}
	}
	for _, participantID := range participantIDs {
		if _, found := seenMemberships[participantID]; !found {
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
