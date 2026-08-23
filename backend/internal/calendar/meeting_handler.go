package calendar

import (
	"errors"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"

	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// MeetingHandler exposes authenticated create and whole-meeting mutation routes.
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
	meetings := api.Group("/calendar/v2", authMw)
	meetings.Post("/meetings", handler.CreateMeeting)
	meetings.Patch("/meetings/:eventId", handler.UpdateMeeting)
	meetings.Delete("/meetings/:eventId", handler.CancelMeeting)
}

func (handler *MeetingHandler) UpdateMeeting(c fiber.Ctx) error {
	if !handler.ready() {
		return meetingHTTPInternal()
	}
	actorID, err := calendarHTTPActor(c)
	if err != nil { return err }
	eventID, err := httputil.ParseUUIDParam(c, "eventId")
	if err != nil { return err }
	expectedVersion, err := calendarHTTPIfMatch(c)
	if err != nil { return err }
	idempotencyKey, err := calendarHTTPIdempotencyKey(c)
	if err != nil { return err }
	var request calendarHTTPMeetingCreateRequest
	if err := calendarHTTPDecode(c, &request); err != nil { return err }
	command, err := request.updateCommand(eventID, expectedVersion, idempotencyKey)
	if err != nil { return meetingHTTPUnprocessable() }
	result, err := handler.service.UpdateMeeting(c.Context(), actorID, command)
	if err != nil { return meetingHTTPMutationError(c, err) }
	return calendarHTTPSendReceipt(c, result.Response(), fiber.StatusOK)
}

func (handler *MeetingHandler) CancelMeeting(c fiber.Ctx) error {
	if !handler.ready() {
		return meetingHTTPInternal()
	}
	if len(c.Body()) != 0 {
		return calendarHTTPBadRequest()
	}
	actorID, err := calendarHTTPActor(c)
	if err != nil { return err }
	eventID, err := httputil.ParseUUIDParam(c, "eventId")
	if err != nil { return err }
	expectedVersion, err := calendarHTTPIfMatch(c)
	if err != nil { return err }
	idempotencyKey, err := calendarHTTPIdempotencyKey(c)
	if err != nil { return err }
	command, err := NewMeetingCancelCommand(MeetingCancelInput{EventID: eventID, ExpectedVersion: expectedVersion, IdempotencyKey: idempotencyKey})
	if err != nil { return meetingHTTPUnprocessable() }
	result, err := handler.service.CancelMeeting(c.Context(), actorID, command)
	if err != nil { return meetingHTTPMutationError(c, err) }
	return calendarHTTPSendReceipt(c, result.Response(), fiber.StatusOK)
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

func meetingHTTPMutationError(c fiber.Ctx, err error) error {
	conflict, found := meetingHTTPMutationConflict(err)
	if found {
		switch conflict.Kind() {
		case MeetingMutationConflictIdempotencyKeyReused:
			return calendarHTTPConflict(c, "IDEMPOTENCY_KEY_REUSED", "meeting request conflicts with an existing receipt")
		case MeetingMutationConflictSlotUnavailable:
			return calendarHTTPConflict(c, "SLOT_UNAVAILABLE", "meeting slot is unavailable")
		case MeetingMutationConflictVersion:
			eventID, eventOK := conflict.EventID()
			version, versionOK := conflict.CurrentVersion()
			if !eventOK || !versionOK {
				return meetingHTTPInternal()
			}
			c.Set("ETag", strconv.Quote(strconv.FormatUint(uint64(version), 10)))
			c.Set("Location", fmt.Sprintf("/api/calendar/v2/meetings/%s", eventID.String()))
			return calendarHTTPConflict(c, "MEETING_VERSION_CONFLICT", "meeting version is stale")
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

func meetingHTTPMutationConflict(err error) (MeetingMutationConflictError, bool) {
	var value MeetingMutationConflictError
	if errors.As(err, &value) { return value, true }
	var pointer *MeetingMutationConflictError
	if errors.As(err, &pointer) && pointer != nil { return *pointer, true }
	return MeetingMutationConflictError{}, false
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
