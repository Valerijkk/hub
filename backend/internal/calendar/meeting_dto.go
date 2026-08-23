package calendar

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type calendarHTTPMeetingCreateRequest struct {
	Title             string      `json:"title"`
	Description       string      `json:"description"`
	ProjectID         uuid.UUID   `json:"project_id"`
	ChannelID         uuid.UUID   `json:"channel_id"`
	StartsAt          time.Time   `json:"starts_at"`
	EndsAt            time.Time   `json:"ends_at"`
	Timezone          string      `json:"timezone"`
	RRULE             string      `json:"rrule"`
	AttendeeUserIDs   []uuid.UUID `json:"attendee_user_ids"`
}

func (request calendarHTTPMeetingCreateRequest) command(idempotencyKey string) (MeetingCreateCommand, error) {
	return NewMeetingCreateCommand(MeetingCreateInput{
		IdempotencyKey:  idempotencyKey,
		Title:           request.Title,
		Description:     request.Description,
		ProjectID:       request.ProjectID,
		ChannelID:       request.ChannelID,
		StartsAt:        request.StartsAt,
		EndsAt:          request.EndsAt,
		Timezone:        request.Timezone,
		RRULE:           request.RRULE,
		AttendeeUserIDs: request.AttendeeUserIDs,
	})
}

type meetingCreateResponseBody struct {
	EventID           uuid.UUID           `json:"event_id"`
	ScheduledCallID   uuid.UUID           `json:"scheduled_call_id"`
	ProjectID         uuid.UUID           `json:"project_id"`
	ChannelID         uuid.UUID           `json:"channel_id"`
	OrganizerID       uuid.UUID           `json:"organizer_id"`
	AttendeeUserIDs   []uuid.UUID         `json:"attendee_user_ids"`
	Title             string              `json:"title"`
	Description       string              `json:"description,omitempty"`
	StartsAt          time.Time           `json:"starts_at"`
	EndsAt            time.Time           `json:"ends_at"`
	Timezone          string              `json:"timezone"`
	RRULE             string              `json:"rrule,omitempty"`
	Status            CalendarEventStatus `json:"status"`
	EventVersion      uint                `json:"event_version"`
	ScheduledCallVersion uint             `json:"scheduled_call_version"`
}

func meetingCreateResponse(eventID, scheduledCallID uuid.UUID, organizerID uuid.UUID, input MeetingCreateInput, eventVersion, callVersion uint) (MutationReceiptResponse, error) {
	if eventID == uuid.Nil || scheduledCallID == uuid.Nil || organizerID == uuid.Nil || eventVersion == 0 || callVersion == 0 {
		return MutationReceiptResponse{}, ErrEventInvalid
	}
	encoded, err := json.Marshal(meetingCreateResponseBody{
		EventID:             eventID,
		ScheduledCallID:     scheduledCallID,
		ProjectID:           input.ProjectID,
		ChannelID:           input.ChannelID,
		OrganizerID:         organizerID,
		AttendeeUserIDs:     append([]uuid.UUID(nil), input.AttendeeUserIDs...),
		Title:               input.Title,
		Description:         input.Description,
		StartsAt:            input.StartsAt,
		EndsAt:              input.EndsAt,
		Timezone:            input.Timezone,
		RRULE:               input.RRULE,
		Status:              CalendarEventStatusConfirmed,
		EventVersion:        eventVersion,
		ScheduledCallVersion: callVersion,
	})
	if err != nil {
		return MutationReceiptResponse{}, ErrEventInvalid
	}
	return NewMutationReceiptResponse(201, encoded)
}

type meetingCalendarCreatedPayload struct {
	EventID         uuid.UUID   `json:"event_id"`
	ScheduledCallID uuid.UUID   `json:"scheduled_call_id"`
	ProjectID       uuid.UUID   `json:"project_id"`
	ChannelID       uuid.UUID   `json:"channel_id"`
	OrganizerID     uuid.UUID   `json:"organizer_id"`
	AttendeeUserIDs []uuid.UUID `json:"attendee_user_ids"`
}

type meetingCallScheduledPayload struct {
	ScheduledCallID uuid.UUID   `json:"scheduled_call_id"`
	CalendarEventID uuid.UUID   `json:"calendar_event_id"`
	ProjectID       uuid.UUID   `json:"project_id"`
	ChannelID       uuid.UUID   `json:"channel_id"`
	CreatorID       uuid.UUID   `json:"creator_id"`
	ParticipantIDs  []uuid.UUID `json:"participant_ids"`
	Title           string      `json:"title"`
	ScheduledAt     time.Time   `json:"scheduled_at"`
	DurationMinutes int         `json:"duration_minutes"`
	RRULE           string      `json:"rrule,omitempty"`
}
