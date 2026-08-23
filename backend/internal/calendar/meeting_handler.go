package calendar

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

// MeetingHandler exposes only the authenticated create vertical in this atom.
type MeetingHandler struct {
	service *MeetingService
}

func NewMeetingHandler(service *MeetingService) *MeetingHandler {
	return &MeetingHandler{service: service}
}

func (handler *MeetingHandler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	if !handler.ready() || authMw == nil || nilCalendarHTTPValue(api) {
		return
	}
	api.Group("/calendar/v2", authMw).Post("/meetings", handler.CreateMeeting)
}

// CreateMeeting accepts actor identity only from the authenticated context.
func (handler *MeetingHandler) CreateMeeting(c fiber.Ctx) error {
	if !handler.ready() {
		return meetingHTTPInternal()
	}
	actorID, err := calendarHTTPActor(c)
	if err != nil {
		return err
	}
	idempotencyKey, err := calendarHTTPIdempotencyKey(c)
	if err != nil {
		return err
	}
	var request calendarHTTPMeetingCreateRequest
	if err := calendarHTTPDecode(c, &request); err != nil {
		return err
	}
	command, err := request.command(idempotencyKey)
	if err != nil {
		return meetingHTTPUnprocessable()
	}
	result, err := handler.service.CreateMeeting(c.Context(), actorID, command)
	if err != nil {
		return meetingHTTPCreateError(c, err)
	}
	return calendarHTTPSendReceipt(c, result.Response(), fiber.StatusCreated)
}

func calendarHTTPMeetingCreateKeyAllowed(key string) bool {
	switch key {
	case "title", "description", "project_id", "channel_id", "starts_at", "ends_at", "timezone", "rrule", "attendee_user_ids":
		return true
	default:
		return false
	}
}

func meetingHTTPCreateError(c fiber.Ctx, err error) error {
	conflict, found := meetingHTTPCreateConflict(err)
	if found {
		switch conflict.Kind() {
		case MeetingCreateConflictIdempotencyKeyReused:
			return calendarHTTPConflict(c, "IDEMPOTENCY_KEY_REUSED", "meeting request conflicts with an existing receipt")
		case MeetingCreateConflictSlotUnavailable:
			return calendarHTTPConflict(c, "SLOT_UNAVAILABLE", "meeting slot is unavailable")
		default:
			return meetingHTTPInternal()
		}
	}
	switch {
	case errors.Is(err, ErrEventNotFound), errors.Is(err, ErrEventForbidden):
		return fiber.NewError(fiber.StatusNotFound, "meeting not found")
	case errors.Is(err, ErrEventInvalid):
		return meetingHTTPUnprocessable()
	default:
		return meetingHTTPInternal()
	}
}

func meetingHTTPCreateConflict(err error) (MeetingCreateConflictError, bool) {
	var value MeetingCreateConflictError
	if errors.As(err, &value) {
		return value, true
	}
	var pointer *MeetingCreateConflictError
	if errors.As(err, &pointer) && pointer != nil {
		return *pointer, true
	}
	return MeetingCreateConflictError{}, false
}

func meetingHTTPUnprocessable() error {
	return fiber.NewError(fiber.StatusUnprocessableEntity, "meeting is invalid")
}

func meetingHTTPInternal() error {
	return fiber.NewError(fiber.StatusInternalServerError, "meeting operation failed")
}

func (handler *MeetingHandler) ready() bool {
	return handler != nil && handler.service != nil &&
		!nilCalendarHTTPValue(handler.service.repository) &&
		!nilCalendarHTTPValue(handler.service.clock) &&
		!nilCalendarHTTPValue(handler.service.ids)
}
