package calendar

import (
	"context"
	"database/sql"

	"github.com/google/uuid"

	hubent "github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarattendee"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/organization"
	"github.com/valerii/hub/backend/ent/user"
)

var _ CalendarEventReadRepository = (*EventEntRepository)(nil)

// ReadCalendarEvent resolves one active actor organization, applies the event
// visibility predicate in SQL, and returns only active master attendees.
func (repository *EventEntRepository) ReadCalendarEvent(ctx context.Context, actorID, eventID uuid.UUID) (CalendarEventDetail, error) {
	if repository == nil || repository.client == nil || ctx == nil || actorID == uuid.Nil || eventID == uuid.Nil {
		return CalendarEventDetail{}, ErrEventInvalid
	}
	transaction, err := repository.client.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelRepeatableRead,
		ReadOnly:   true,
	})
	if err != nil {
		return CalendarEventDetail{}, ErrEventInternal
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

	detail, err := repository.readCalendarEventSnapshot(ctx, transaction, actorID, eventID)
	if err != nil {
		return CalendarEventDetail{}, err
	}
	if err := transaction.Commit(); err != nil {
		return CalendarEventDetail{}, ErrEventInternal
	}
	committed = true
	return detail, nil
}

func (repository *EventEntRepository) readCalendarEventSnapshot(ctx context.Context, transaction *hubent.Tx, actorID, eventID uuid.UUID) (CalendarEventDetail, error) {
	if repository == nil || transaction == nil || ctx == nil || actorID == uuid.Nil || eventID == uuid.Nil {
		return CalendarEventDetail{}, ErrEventInvalid
	}

	organizations, err := transaction.Organization.Query().
		Where(organization.HasUsersWith(user.IDEQ(actorID), user.StatusEQ(user.StatusActive))).
		Limit(2).
		All(ctx)
	if err != nil {
		return CalendarEventDetail{}, ErrEventInternal
	}
	if len(organizations) == 0 {
		return CalendarEventDetail{}, ErrEventNotFound
	}
	if len(organizations) != 1 || organizations[0] == nil || organizations[0].ID == uuid.Nil {
		return CalendarEventDetail{}, ErrEventInternal
	}
	organizationID := organizations[0].ID

	event, err := transaction.CalendarEvent.Query().
		Where(
			calendarevent.IDEQ(eventID),
			calendarevent.OrganizationIDEQ(organizationID),
			calendarevent.DeletedAtIsNil(),
			calendarevent.Or(
				calendarevent.CalendarOwnerIDEQ(actorID),
				calendarevent.OrganizerIDEQ(actorID),
				calendarevent.HasAttendeesWith(
					calendarattendee.UserIDEQ(actorID),
					calendarattendee.RemovedAtIsNil(),
					calendarattendee.RecurrenceIDIsNil(),
				),
			),
		).
		Only(ctx)
	if err != nil {
		if hubent.IsNotFound(err) {
			return CalendarEventDetail{}, ErrEventNotFound
		}
		return CalendarEventDetail{}, ErrEventInternal
	}
	input, err := eventEntCalendarEventInput(event)
	if err != nil || input.EventID != eventID || input.OrganizationID != organizationID || !validCalendarEventUpdateAuthoritativeCurrent(input) {
		return CalendarEventDetail{}, ErrEventInternal
	}

	rows, err := transaction.CalendarAttendee.Query().
		Where(
			calendarattendee.CalendarEventIDEQ(eventID),
			calendarattendee.RemovedAtIsNil(),
			calendarattendee.RecurrenceIDIsNil(),
		).
		Order(hubent.Asc(calendarattendee.FieldID)).
		Limit(calendarCreateMaxAttendees + 2).
		All(ctx)
	if err != nil || len(rows) > calendarCreateMaxAttendees+1 {
		return CalendarEventDetail{}, ErrEventInternal
	}
	return calendarEventDetailFromEnt(input, rows, actorID)
}

func calendarEventDetailFromEnt(input CalendarEventCommandInput, rows []*hubent.CalendarAttendee, actorID uuid.UUID) (CalendarEventDetail, error) {
	actorOwnsOrOrganizes := input.OwnerID == actorID || input.OrganizerID != nil && *input.OrganizerID == actorID
	detail := CalendarEventDetail{
		ID:             input.EventID,
		Title:          input.Title,
		Description:    input.Description,
		Location:       input.Location,
		AllDay:         input.Temporal.AllDay,
		Timezone:       input.Temporal.Timezone,
		Exdates:        append([]string{}, input.Exdates...),
		Sequence:       input.Sequence,
		Status:         input.Status,
		Version:        input.Version,
		OrganizerID:    copyUUID(input.OrganizerID),
		OrganizerEmail: input.OrganizerEmail,
		Attendees:      make([]CalendarEventReadAttendee, 0, len(rows)),
	}
	if input.Temporal.AllDay {
		startDate := input.Temporal.StartDate
		endDateExclusive := input.Temporal.EndDateExclusive
		detail.StartDate = &startDate
		detail.EndDateExclusive = &endDateExclusive
	} else {
		startsAt := input.Temporal.StartsAt.UTC()
		endsAt := input.Temporal.EndsAt.UTC()
		detail.StartsAt = &startsAt
		detail.EndsAt = &endsAt
	}
	if input.RRULE != "" {
		rrule := input.RRULE
		detail.RRULE = &rrule
	}

	for _, row := range rows {
		if row == nil || row.CalendarEventID != input.EventID || row.RemovedAt != nil || row.RecurrenceID != nil {
			return CalendarEventDetail{}, ErrEventInternal
		}
		snapshot, err := eventEntUpdateAttendeeSnapshot(row)
		if err != nil {
			return CalendarEventDetail{}, ErrEventInternal
		}
		attendee := CalendarEventReadAttendee{
			ID:          snapshot.AttendeeID,
			UserID:      copyUUID(snapshot.UserID),
			Email:       snapshot.Email,
			Role:        snapshot.Role,
			RSVPStatus:  snapshot.Status,
			Version:     snapshot.Version,
			RespondedAt: cloneCalendarEventReadTime(snapshot.RespondedAt),
		}
		detail.Attendees = append(detail.Attendees, attendee)
		if attendee.UserID != nil && *attendee.UserID == actorID {
			if detail.ViewerAttendee != nil {
				if actorOwnsOrOrganizes {
					return CalendarEventDetail{}, ErrEventInternal
				}
				return CalendarEventDetail{}, ErrEventNotFound
			}
			viewer := cloneCalendarEventReadAttendee(attendee)
			detail.ViewerAttendee = &viewer
		}
	}
	if !actorOwnsOrOrganizes && detail.ViewerAttendee == nil {
		return CalendarEventDetail{}, ErrEventNotFound
	}
	detail.CanRSVP = detail.ViewerAttendee != nil && detail.Status != CalendarEventStatusCancelled && detail.ViewerAttendee.Role != CalendarEventAttendeeRoleOrganizer
	if detail.OrganizerID != nil && *detail.OrganizerID == actorID {
		detail.CanRSVP = false
	}
	if !calendarEventDetailValid(detail, actorID, input.EventID) {
		return CalendarEventDetail{}, ErrEventInternal
	}
	return detail, nil
}
