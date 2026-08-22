package calendar

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

type eventEntTransactionBackend struct {
	tx  *ent.Tx
	now func() time.Time
}

var _ EventTransactionBackend = (*eventEntTransactionBackend)(nil)

func newEventEntTransactionBackend(transaction *ent.Tx, now func() time.Time) *eventEntTransactionBackend {
	return &eventEntTransactionBackend{tx: transaction, now: now}
}

func (backend *eventEntTransactionBackend) ResolveActorOrganization(ctx context.Context, actorID uuid.UUID) (uuid.UUID, error) {
	if backend == nil || backend.tx == nil || actorID == uuid.Nil {
		return uuid.Nil, ErrEventInvalid
	}

	actor, err := backend.activeUserForUpdate(ctx, actorID)
	if err != nil {
		return uuid.Nil, err
	}
	organization, err := backend.tx.Organization.Query().
		Where(organization.HasUsersWith(user.IDEQ(actor.ID))).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return uuid.Nil, calendarEntAuthorizationError(err)
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

func (backend *eventEntTransactionBackend) activeOrganizationMemberForUpdate(ctx context.Context, memberID, organizationID uuid.UUID) (*ent.User, error) {
	if memberID == uuid.Nil || organizationID == uuid.Nil {
		return nil, ErrEventForbidden
	}
	member, err := backend.activeUserForUpdate(ctx, memberID)
	if err != nil {
		return nil, err
	}
	_, err = backend.tx.Organization.Query().
		Where(
			organization.IDEQ(organizationID),
			organization.HasUsersWith(user.IDEQ(member.ID)),
		).
		ForUpdate().
		Only(ctx)
	if err != nil {
		return nil, calendarEntAuthorizationError(err)
	}
	return member, nil
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
	case ent.IsNotFound(err):
		return ErrEventNotFound
	case ent.IsConstraintError(err):
		return ErrEventConflict
	default:
		return ErrEventInvalid
	}
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
