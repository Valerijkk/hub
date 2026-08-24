package calendar

import (
	"context"
	"errors"
	"reflect"
	"time"

	"github.com/google/uuid"
)

// CalendarEventReadRepository is the isolated authorization-aware read port
// for one canonical Calendar event.
type CalendarEventReadRepository interface {
	ReadCalendarEvent(context.Context, uuid.UUID, uuid.UUID) (CalendarEventDetail, error)
}

// CalendarEventDetail is the safe public representation of one master event.
// It deliberately excludes provider state, hashes, deleted rows, iTIP payloads,
// and occurrence overrides.
type CalendarEventDetail struct {
	ID               uuid.UUID                   `json:"id"`
	Title            string                      `json:"title"`
	Description      string                      `json:"description"`
	Location         string                      `json:"location"`
	StartsAt         *time.Time                  `json:"starts_at"`
	EndsAt           *time.Time                  `json:"ends_at"`
	StartDate        *string                     `json:"start_date"`
	EndDateExclusive *string                     `json:"end_date_exclusive"`
	AllDay           bool                        `json:"all_day"`
	Timezone         string                      `json:"timezone"`
	RRULE            *string                     `json:"rrule"`
	Exdates          []string                    `json:"exdates"`
	Sequence         uint                        `json:"sequence"`
	Status           CalendarEventStatus         `json:"status"`
	Version          uint                        `json:"version"`
	OrganizerID      *uuid.UUID                  `json:"organizer_id"`
	OrganizerEmail   string                      `json:"organizer_email"`
	Attendees        []CalendarEventReadAttendee `json:"attendees"`
	ViewerAttendee   *CalendarEventReadAttendee  `json:"viewer_attendee"`
	CanRSVP          bool                        `json:"can_rsvp"`
}

// CalendarEventReadAttendee is one active, master-series attendee. Absence of
// UserID represents an external invitee without exposing any directory data.
type CalendarEventReadAttendee struct {
	ID          uuid.UUID                 `json:"id"`
	UserID      *uuid.UUID                `json:"user_id"`
	Email       string                    `json:"email"`
	Role        CalendarEventAttendeeRole `json:"role"`
	RSVPStatus  RSVPStatus                `json:"rsvp_status"`
	Version     uint                      `json:"version"`
	RespondedAt *time.Time                `json:"responded_at"`
}

// EventReadService validates the repository boundary before returning a safe
// detail object to HTTP.
type EventReadService struct {
	repository CalendarEventReadRepository
}

// NewEventReadService creates the isolated Calendar event read service.
func NewEventReadService(repository CalendarEventReadRepository) *EventReadService {
	return &EventReadService{repository: repository}
}

// GetCalendarEvent returns one event visible to the authenticated actor.
func (service *EventReadService) GetCalendarEvent(ctx context.Context, actorID, eventID uuid.UUID) (CalendarEventDetail, error) {
	if !service.ready() || ctx == nil || actorID == uuid.Nil || eventID == uuid.Nil {
		return CalendarEventDetail{}, ErrEventInvalid
	}
	detail, err := service.repository.ReadCalendarEvent(ctx, actorID, eventID)
	if err != nil {
		if errors.Is(err, ErrEventNotFound) {
			return CalendarEventDetail{}, ErrEventNotFound
		}
		return CalendarEventDetail{}, ErrEventInternal
	}
	if !calendarEventDetailValid(detail, actorID, eventID) {
		return CalendarEventDetail{}, ErrEventInternal
	}
	return cloneCalendarEventDetail(detail), nil
}

func (service *EventReadService) ready() bool {
	return service != nil && !calendarEventReadDependencyNil(service.repository)
}

func calendarEventDetailValid(detail CalendarEventDetail, actorID, eventID uuid.UUID) bool {
	if detail.ID != eventID || actorID == uuid.Nil || detail.Title == "" || detail.Version == 0 ||
		detail.OrganizerEmail != normalizeCalendarCreateEmail(detail.OrganizerEmail) ||
		!validCalendarCreateEmail(detail.OrganizerEmail) || !validCalendarEventStatus(detail.Status) ||
		len(detail.Exdates) > calendarCreateMaxExdates || len(detail.Attendees) > calendarCreateMaxAttendees+1 {
		return false
	}
	if detail.OrganizerID != nil && *detail.OrganizerID == uuid.Nil {
		return false
	}
	temporal := CalendarEvent{AllDay: detail.AllDay, Timezone: detail.Timezone}
	if detail.AllDay {
		if detail.StartsAt != nil || detail.EndsAt != nil || detail.StartDate == nil || detail.EndDateExclusive == nil {
			return false
		}
		temporal.StartDate = *detail.StartDate
		temporal.EndDateExclusive = *detail.EndDateExclusive
	} else {
		if detail.StartsAt == nil || detail.EndsAt == nil || detail.StartDate != nil || detail.EndDateExclusive != nil {
			return false
		}
		temporal.StartsAt = detail.StartsAt.UTC()
		temporal.EndsAt = detail.EndsAt.UTC()
	}
	if temporal.Validate() != nil || detail.RRULE != nil && *detail.RRULE == "" {
		return false
	}

	viewerCount := 0
	for index := range detail.Attendees {
		attendee := detail.Attendees[index]
		if !calendarEventReadAttendeeValid(attendee) {
			return false
		}
		if index > 0 && detail.Attendees[index-1].ID.String() >= attendee.ID.String() {
			return false
		}
		if attendee.UserID != nil && *attendee.UserID == actorID {
			viewerCount++
		}
	}
	if !calendarEventReadOrganizerGraphValid(detail) {
		return false
	}
	if detail.ViewerAttendee == nil {
		return viewerCount == 0 && !detail.CanRSVP
	}
	if viewerCount != 1 || detail.ViewerAttendee.UserID == nil || *detail.ViewerAttendee.UserID != actorID ||
		!calendarEventReadAttendeeInSlice(*detail.ViewerAttendee, detail.Attendees) {
		return false
	}
	canRSVP := detail.Status != CalendarEventStatusCancelled && detail.ViewerAttendee.Role != CalendarEventAttendeeRoleOrganizer
	if detail.OrganizerID != nil && *detail.OrganizerID == actorID {
		canRSVP = false
	}
	return detail.CanRSVP == canRSVP
}

func calendarEventReadOrganizerGraphValid(detail CalendarEventDetail) bool {
	if detail.OrganizerID == nil {
		for _, attendee := range detail.Attendees {
			if attendee.Role == CalendarEventAttendeeRoleOrganizer {
				return false
			}
		}
		return true
	}

	organizerID := *detail.OrganizerID
	matches := 0
	for _, attendee := range detail.Attendees {
		idMatches := attendee.UserID != nil && *attendee.UserID == organizerID
		emailMatches := attendee.Email == detail.OrganizerEmail
		organizerRole := attendee.Role == CalendarEventAttendeeRoleOrganizer
		if !idMatches && !emailMatches && !organizerRole {
			continue
		}
		if !idMatches || !emailMatches || attendee.RSVPStatus != RSVPAccepted ||
			attendee.Role != CalendarEventAttendeeRoleOrganizer && attendee.Role != CalendarEventAttendeeRoleRequired {
			return false
		}
		// Before the organizer-role cutover, the canonical organizer row used
		// required. It remains acceptable only under this exact ID+email+accepted
		// graph; every new row uses the dedicated organizer role.
		matches++
	}
	return matches == 1
}

func calendarEventReadAttendeeValid(attendee CalendarEventReadAttendee) bool {
	if attendee.ID == uuid.Nil || attendee.Version == 0 || attendee.Email != normalizeCalendarCreateEmail(attendee.Email) ||
		!validCalendarCreateEmail(attendee.Email) || !validCalendarEventStoredAttendeeRoleState(attendee.Role, attendee.RSVPStatus, "") {
		return false
	}
	if attendee.UserID != nil && *attendee.UserID == uuid.Nil {
		return false
	}
	if attendee.Role == CalendarEventAttendeeRoleOrganizer && (attendee.UserID == nil || attendee.RSVPStatus != RSVPAccepted) {
		return false
	}
	if attendee.RSVPStatus == RSVPNeedsAction {
		return attendee.RespondedAt == nil
	}
	return attendee.RespondedAt == nil || !attendee.RespondedAt.IsZero() && attendee.RespondedAt.Location() == time.UTC
}

func calendarEventReadAttendeeInSlice(expected CalendarEventReadAttendee, attendees []CalendarEventReadAttendee) bool {
	for _, attendee := range attendees {
		if attendee.ID == expected.ID && attendee.Version == expected.Version && attendee.Email == expected.Email && attendee.Role == expected.Role && attendee.RSVPStatus == expected.RSVPStatus &&
			calendarEventReadUUIDPointersEqual(attendee.UserID, expected.UserID) && calendarEventReadTimePointersEqual(attendee.RespondedAt, expected.RespondedAt) {
			return true
		}
	}
	return false
}

func cloneCalendarEventDetail(detail CalendarEventDetail) CalendarEventDetail {
	detail.StartsAt = cloneCalendarEventReadTime(detail.StartsAt)
	detail.EndsAt = cloneCalendarEventReadTime(detail.EndsAt)
	detail.StartDate = cloneCalendarEventReadString(detail.StartDate)
	detail.EndDateExclusive = cloneCalendarEventReadString(detail.EndDateExclusive)
	detail.RRULE = cloneCalendarEventReadString(detail.RRULE)
	detail.OrganizerID = copyUUID(detail.OrganizerID)
	detail.Exdates = append([]string{}, detail.Exdates...)
	detail.Attendees = append([]CalendarEventReadAttendee{}, detail.Attendees...)
	for index := range detail.Attendees {
		detail.Attendees[index] = cloneCalendarEventReadAttendee(detail.Attendees[index])
	}
	if detail.ViewerAttendee != nil {
		viewer := cloneCalendarEventReadAttendee(*detail.ViewerAttendee)
		detail.ViewerAttendee = &viewer
	}
	return detail
}

func cloneCalendarEventReadAttendee(attendee CalendarEventReadAttendee) CalendarEventReadAttendee {
	attendee.UserID = copyUUID(attendee.UserID)
	attendee.RespondedAt = cloneCalendarEventReadTime(attendee.RespondedAt)
	return attendee
}

func cloneCalendarEventReadTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func cloneCalendarEventReadString(value *string) *string {
	if value == nil {
		return nil
	}
	copied := *value
	return &copied
}

func calendarEventReadUUIDPointersEqual(left, right *uuid.UUID) bool {
	return left == nil && right == nil || left != nil && right != nil && *left == *right
}

func calendarEventReadTimePointersEqual(left, right *time.Time) bool {
	return left == nil && right == nil || left != nil && right != nil && left.Equal(*right)
}

func calendarEventReadDependencyNil(value any) bool {
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
