package calendar

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"

	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

const (
	calendarHTTPMaxBodyBytes = 1 << 20
	calendarHTTPMaxJSONDepth = 16
	calendarHTTPMaxJSONTokens = 32768
)

// EventHandler exposes canonical CalendarEvent mutations without owning
// authorization, persistence, provider, or cross-domain policy.
type EventHandler struct {
	service *EventService
}

// NewEventHandler creates the canonical CalendarEvent mutation handler.
func NewEventHandler(service *EventService) *EventHandler {
	return &EventHandler{service: service}
}

// RegisterRoutes registers only the authenticated canonical mutation surface.
// Legacy Calendar GET routes remain owned by Handler.RegisterRoutes.
func (h *EventHandler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	if !h.ready() || authMw == nil || nilCalendarHTTPValue(api) {
		return
	}
	calendarV2 := api.Group("/calendar/v2", authMw)
	calendarV2.Post("/events", h.CreateCalendarEvent)
	calendarV2.Patch("/events/:eventId", h.UpdateCalendarEvent)
	calendarV2.Delete("/events/:eventId", h.CancelCalendarEvent)
	calendarV2.Post("/events/:eventId/rsvp", h.RespondToCalendarSeries)
	calendarV2.Post("/events/:eventId/instances/:recurrenceId/rsvp", h.RespondToCalendarOccurrence)
	calendarV2.Delete("/events/:eventId/instances/:recurrenceId", h.CancelCalendarEventOccurrence)
}

// CreateCalendarEvent creates one canonical event from an authenticated actor.
func (h *EventHandler) CreateCalendarEvent(c fiber.Ctx) error {
	if !h.ready() {
		return calendarHTTPInternal()
	}
	actorID, err := calendarHTTPActor(c)
	if err != nil {
		return err
	}
	key, err := calendarHTTPIdempotencyKey(c)
	if err != nil {
		return err
	}
	var request calendarHTTPCreateRequest
	if err := calendarHTTPDecode(c, &request); err != nil {
		return err
	}
	command, err := request.command(key)
	if err != nil {
		return calendarHTTPUnprocessable()
	}
	result, err := h.service.CreateCalendarEvent(c.Context(), actorID, command)
	if err != nil {
		return calendarHTTPCreateError(c, err)
	}
	return calendarHTTPSendReceipt(c, result.Response(), fiber.StatusCreated)
}

// UpdateCalendarEvent applies a presence-preserving patch to one master event.
func (h *EventHandler) UpdateCalendarEvent(c fiber.Ctx) error {
	if !h.ready() {
		return calendarHTTPInternal()
	}
	actorID, err := calendarHTTPActor(c)
	if err != nil {
		return err
	}
	eventID, err := httputil.ParseUUIDParam(c, "eventId")
	if err != nil {
		return err
	}
	expectedVersion, err := calendarHTTPIfMatch(c)
	if err != nil {
		return err
	}
	notifyAttendees, err := calendarHTTPNotifyAttendees(c)
	if err != nil {
		return err
	}
	var request calendarHTTPUpdateRequest
	if err := calendarHTTPDecode(c, &request); err != nil {
		return err
	}
	command, err := request.command(eventID, expectedVersion, notifyAttendees)
	if err != nil {
		return calendarHTTPUnprocessable()
	}
	result, err := h.service.UpdateCalendarEvent(c.Context(), actorID, command)
	if err != nil {
		return calendarHTTPUpdateError(c, err)
	}
	response, err := calendarHTTPMutationResponse(result)
	if err != nil {
		return calendarHTTPInternal()
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

// CancelCalendarEvent cancels one canonical master event.
func (h *EventHandler) CancelCalendarEvent(c fiber.Ctx) error {
	if !h.ready() {
		return calendarHTTPInternal()
	}
	actorID, err := calendarHTTPActor(c)
	if err != nil {
		return err
	}
	eventID, err := httputil.ParseUUIDParam(c, "eventId")
	if err != nil {
		return err
	}
	expectedVersion, err := calendarHTTPIfMatch(c)
	if err != nil {
		return err
	}
	result, err := h.service.CancelCalendarEvent(c.Context(), actorID, eventID, expectedVersion)
	if err != nil {
		return calendarHTTPCancelError(c, err)
	}
	response, err := calendarHTTPMutationResponse(result)
	if err != nil {
		return calendarHTTPInternal()
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

// RespondToCalendarSeries changes only the authenticated attendee's series RSVP.
func (h *EventHandler) RespondToCalendarSeries(c fiber.Ctx) error {
	if !h.ready() {
		return calendarHTTPInternal()
	}
	actorID, err := calendarHTTPActor(c)
	if err != nil {
		return err
	}
	eventID, err := httputil.ParseUUIDParam(c, "eventId")
	if err != nil {
		return err
	}
	key, err := calendarHTTPIdempotencyKey(c)
	if err != nil {
		return err
	}
	var request calendarHTTPRSVPRequest
	if err := calendarHTTPDecode(c, &request); err != nil {
		return err
	}
	command, err := NewCalendarEventSeriesRSVPCommand(CalendarEventSeriesRSVPInput{
		IdempotencyKey: key, EventID: eventID,
		ExpectedAttendeeVersion: request.AttendeeVersion,
		Status: RSVPStatus(request.Status),
	})
	if err != nil {
		return calendarHTTPUnprocessable()
	}
	result, err := h.service.RespondToCalendarSeries(c.Context(), actorID, command)
	if err != nil {
		return calendarHTTPSeriesRSVPError(c, err)
	}
	return calendarHTTPSendReceipt(c, result.Response(), fiber.StatusOK)
}

// RespondToCalendarOccurrence changes only the authenticated attendee's occurrence RSVP.
func (h *EventHandler) RespondToCalendarOccurrence(c fiber.Ctx) error {
	if !h.ready() {
		return calendarHTTPInternal()
	}
	actorID, err := calendarHTTPActor(c)
	if err != nil {
		return err
	}
	eventID, err := httputil.ParseUUIDParam(c, "eventId")
	if err != nil {
		return err
	}
	recurrenceID, err := calendarHTTPRecurrenceID(c)
	if err != nil {
		return err
	}
	key, err := calendarHTTPIdempotencyKey(c)
	if err != nil {
		return err
	}
	var request calendarHTTPRSVPRequest
	if err := calendarHTTPDecode(c, &request); err != nil {
		return err
	}
	command, err := NewCalendarEventOccurrenceRSVPCommand(CalendarEventOccurrenceRSVPInput{
		IdempotencyKey: key, EventID: eventID, RecurrenceID: recurrenceID,
		ExpectedAttendeeVersion: request.AttendeeVersion,
		Status: RSVPStatus(request.Status),
	})
	if err != nil {
		return calendarHTTPUnprocessable()
	}
	result, err := h.service.RespondToCalendarOccurrence(c.Context(), actorID, command)
	if err != nil {
		return calendarHTTPOccurrenceRSVPError(c, err)
	}
	return calendarHTTPSendReceipt(c, result.Response(), fiber.StatusOK)
}

// CancelCalendarEventOccurrence cancels one canonical recurrence instance.
func (h *EventHandler) CancelCalendarEventOccurrence(c fiber.Ctx) error {
	if !h.ready() {
		return calendarHTTPInternal()
	}
	actorID, err := calendarHTTPActor(c)
	if err != nil {
		return err
	}
	eventID, err := httputil.ParseUUIDParam(c, "eventId")
	if err != nil {
		return err
	}
	recurrenceID, err := calendarHTTPRecurrenceID(c)
	if err != nil {
		return err
	}
	expectedVersion, err := calendarHTTPIfMatch(c)
	if err != nil {
		return err
	}
	command, err := NewCalendarEventOccurrenceCancelCommand(CalendarEventOccurrenceCancelInput{
		EventID: eventID, RecurrenceID: recurrenceID, ExpectedVersion: expectedVersion,
	})
	if err != nil {
		return calendarHTTPUnprocessable()
	}
	result, err := h.service.CancelCalendarEventOccurrence(c.Context(), actorID, command)
	if err != nil {
		return calendarHTTPOccurrenceCancelError(c, err)
	}
	if _, err := NewCalendarEventOccurrenceCancelResult(
		result.EventID(), result.RecurrenceID(), result.Status(), result.Sequence(), result.EventVersion(),
	); err != nil {
		return calendarHTTPInternal()
	}
	return c.Status(fiber.StatusOK).JSON(calendarHTTPOccurrenceCancelResponse{
		EventID: result.EventID(), RecurrenceID: result.RecurrenceID(), Status: result.Status(),
		Sequence: result.Sequence(), EventVersion: result.EventVersion(),
	})
}

type calendarHTTPField[T any] struct {
	present bool
	null    bool
	value   T
}

func (field *calendarHTTPField[T]) UnmarshalJSON(data []byte) error {
	if field == nil || field.present {
		return errors.New("invalid calendar field")
	}
	field.present = true
	if bytes.Equal(bytes.TrimSpace(data), []byte("null")) {
		field.null = true
		return nil
	}
	return calendarHTTPDecodeValue(data, &field.value)
}

type calendarHTTPAttendeeRequest struct {
	UserID      calendarHTTPField[uuid.UUID] `json:"user_id"`
	Email       calendarHTTPField[string]    `json:"email"`
	DisplayName calendarHTTPField[string]    `json:"display_name"`
	Role        calendarHTTPField[string]    `json:"role"`
}

func (request calendarHTTPAttendeeRequest) candidate() (CalendarEventCreateAttendeeCandidate, error) {
	if !request.Email.present || request.Email.null || !request.Role.present || request.Role.null {
		return CalendarEventCreateAttendeeCandidate{}, ErrEventInvalid
	}
	var userID *uuid.UUID
	if request.UserID.present && !request.UserID.null {
		value := request.UserID.value
		userID = &value
	}
	var displayName *string
	if request.DisplayName.present && !request.DisplayName.null {
		value := request.DisplayName.value
		displayName = &value
	}
	return CalendarEventCreateAttendeeCandidate{
		UserID: userID, Email: request.Email.value, DisplayName: displayName,
		Role: CalendarEventAttendeeRole(request.Role.value),
	}, nil
}

type calendarHTTPCreateRequest struct {
	Title            calendarHTTPField[string]                        `json:"title"`
	Description      calendarHTTPField[string]                        `json:"description"`
	Location         calendarHTTPField[string]                        `json:"location"`
	StartsAt         calendarHTTPField[time.Time]                     `json:"starts_at"`
	EndsAt           calendarHTTPField[time.Time]                     `json:"ends_at"`
	StartDate        calendarHTTPField[string]                        `json:"start_date"`
	EndDateExclusive calendarHTTPField[string]                        `json:"end_date_exclusive"`
	AllDay           calendarHTTPField[bool]                          `json:"all_day"`
	Timezone         calendarHTTPField[string]                        `json:"timezone"`
	RRULE            calendarHTTPField[string]                        `json:"rrule"`
	Exdates          calendarHTTPField[[]string]                      `json:"exdates"`
	ScheduledCallID  calendarHTTPField[uuid.UUID]                     `json:"scheduled_call_id"`
	ReminderMinutes  calendarHTTPField[int]                           `json:"reminder_minutes"`
	Attendees        calendarHTTPField[[]calendarHTTPAttendeeRequest] `json:"attendees"`
}

func (request calendarHTTPCreateRequest) command(idempotencyKey string) (CalendarEventCreateCommand, error) {
	if !request.validShape() {
		return CalendarEventCreateCommand{}, ErrEventInvalid
	}
	temporal := CalendarEvent{AllDay: request.AllDay.value, Timezone: request.Timezone.value}
	if request.AllDay.value {
		temporal.StartDate = request.StartDate.value
		temporal.EndDateExclusive = request.EndDateExclusive.value
	} else {
		temporal.StartsAt = request.StartsAt.value
		temporal.EndsAt = request.EndsAt.value
	}
	rrule := ""
	if request.RRULE.present && !request.RRULE.null {
		rrule = request.RRULE.value
	}
	exdates := make([]string, 0)
	if request.Exdates.present {
		exdates = append(exdates, request.Exdates.value...)
	}
	var scheduledCallID *uuid.UUID
	if request.ScheduledCallID.present && !request.ScheduledCallID.null {
		value := request.ScheduledCallID.value
		scheduledCallID = &value
	}
	var reminderMinutes *int
	if request.ReminderMinutes.present && !request.ReminderMinutes.null {
		value := request.ReminderMinutes.value
		reminderMinutes = &value
	}
	attendees := make([]CalendarEventCreateAttendeeCandidate, len(request.Attendees.value))
	for index, raw := range request.Attendees.value {
		candidate, err := raw.candidate()
		if err != nil {
			return CalendarEventCreateCommand{}, ErrEventInvalid
		}
		attendees[index] = candidate
	}
	description := ""
	if request.Description.present {
		description = request.Description.value
	}
	location := ""
	if request.Location.present {
		location = request.Location.value
	}
	return NewCalendarEventCreateCommand(CalendarEventCreateInput{
		IdempotencyKey: idempotencyKey, Title: request.Title.value,
		Description: description, Location: location, Temporal: temporal,
		RRULE: rrule, Exdates: exdates, ScheduledCallID: scheduledCallID,
		ReminderMinutes: reminderMinutes, Attendees: attendees,
	})
}

func (request calendarHTTPCreateRequest) validShape() bool {
	if !request.Title.present || request.Title.null ||
		request.Description.present && request.Description.null ||
		request.Location.present && request.Location.null ||
		!request.AllDay.present || request.AllDay.null ||
		!request.Timezone.present || request.Timezone.null ||
		request.Exdates.present && request.Exdates.null ||
		!request.Attendees.present || request.Attendees.null {
		return false
	}
	if !request.StartsAt.present || !request.EndsAt.present ||
		!request.StartDate.present || !request.EndDateExclusive.present {
		return false
	}
	if request.AllDay.value {
		return request.StartsAt.null && request.EndsAt.null &&
			!request.StartDate.null && !request.EndDateExclusive.null
	}
	return !request.StartsAt.null && !request.EndsAt.null &&
		request.StartDate.null && request.EndDateExclusive.null
}

type calendarHTTPUpdateRequest struct {
	Title            calendarHTTPField[string]                        `json:"title"`
	Description      calendarHTTPField[string]                        `json:"description"`
	Location         calendarHTTPField[string]                        `json:"location"`
	StartsAt         calendarHTTPField[time.Time]                     `json:"starts_at"`
	EndsAt           calendarHTTPField[time.Time]                     `json:"ends_at"`
	StartDate        calendarHTTPField[string]                        `json:"start_date"`
	EndDateExclusive calendarHTTPField[string]                        `json:"end_date_exclusive"`
	AllDay           calendarHTTPField[bool]                          `json:"all_day"`
	Timezone         calendarHTTPField[string]                        `json:"timezone"`
	RRULE            calendarHTTPField[string]                        `json:"rrule"`
	Exdates          calendarHTTPField[[]string]                      `json:"exdates"`
	ReminderMinutes  calendarHTTPField[int]                           `json:"reminder_minutes"`
	Attendees        calendarHTTPField[[]calendarHTTPAttendeeRequest] `json:"attendees"`
}

func (request calendarHTTPUpdateRequest) command(eventID uuid.UUID, expectedVersion uint, notifyAttendees bool) (CalendarEventUpdateCommand, error) {
	if !request.validShape() {
		return CalendarEventUpdateCommand{}, ErrEventInvalid
	}
	input := CalendarEventUpdateInput{
		EventID: eventID, ExpectedVersion: expectedVersion, NotifyAttendees: notifyAttendees,
		Title: CalendarEventUpdateString{Present: request.Title.present, Value: request.Title.value},
		Description: CalendarEventUpdateString{Present: request.Description.present, Value: request.Description.value},
		Location: CalendarEventUpdateString{Present: request.Location.present, Value: request.Location.value},
		AllDay: CalendarEventUpdateBool{Present: request.AllDay.present, Value: request.AllDay.value},
		Timezone: CalendarEventUpdateString{Present: request.Timezone.present, Value: request.Timezone.value},
	}
	if request.StartsAt.present {
		input.StartsAt.Present = true
		if !request.StartsAt.null {
			value := request.StartsAt.value
			input.StartsAt.Value = &value
		}
	}
	if request.EndsAt.present {
		input.EndsAt.Present = true
		if !request.EndsAt.null {
			value := request.EndsAt.value
			input.EndsAt.Value = &value
		}
	}
	if request.StartDate.present {
		input.StartDate.Present = true
		if !request.StartDate.null {
			value := request.StartDate.value
			input.StartDate.Value = &value
		}
	}
	if request.EndDateExclusive.present {
		input.EndDateExclusive.Present = true
		if !request.EndDateExclusive.null {
			value := request.EndDateExclusive.value
			input.EndDateExclusive.Value = &value
		}
	}
	if request.RRULE.present {
		input.RRULE.Present = true
		if !request.RRULE.null {
			value := request.RRULE.value
			input.RRULE.Value = &value
		}
	}
	if request.Exdates.present {
		input.Exdates = CalendarEventUpdateExdates{Present: true, Value: append([]string{}, request.Exdates.value...)}
	}
	if request.ReminderMinutes.present {
		input.ReminderMinutes.Present = true
		if !request.ReminderMinutes.null {
			value := request.ReminderMinutes.value
			input.ReminderMinutes.Value = &value
		}
	}
	if request.Attendees.present {
		attendees := make([]CalendarEventCreateAttendeeCandidate, len(request.Attendees.value))
		for index, raw := range request.Attendees.value {
			candidate, err := raw.candidate()
			if err != nil {
				return CalendarEventUpdateCommand{}, ErrEventInvalid
			}
			attendees[index] = candidate
		}
		input.Attendees = CalendarEventUpdateAttendees{Present: true, Value: attendees}
	}
	return NewCalendarEventUpdateCommand(input)
}

func (request calendarHTTPUpdateRequest) validShape() bool {
	if request.Title.present && request.Title.null ||
		request.Description.present && request.Description.null ||
		request.Location.present && request.Location.null ||
		request.AllDay.present && request.AllDay.null ||
		request.Timezone.present && request.Timezone.null ||
		request.Exdates.present && request.Exdates.null ||
		request.Attendees.present && request.Attendees.null {
		return false
	}
	count := 0
	for _, present := range []bool{
		request.Title.present, request.Description.present, request.Location.present,
		request.StartsAt.present, request.EndsAt.present, request.StartDate.present,
		request.EndDateExclusive.present, request.AllDay.present, request.Timezone.present,
		request.RRULE.present, request.Exdates.present, request.ReminderMinutes.present,
		request.Attendees.present,
	} {
		if present {
			count++
		}
	}
	if count == 0 || request.StartsAt.present != request.EndsAt.present ||
		request.StartDate.present != request.EndDateExclusive.present {
		return false
	}
	if !request.AllDay.present {
		return true
	}
	if !request.StartsAt.present || !request.EndsAt.present ||
		!request.StartDate.present || !request.EndDateExclusive.present {
		return false
	}
	if request.AllDay.value {
		return request.StartsAt.null && request.EndsAt.null &&
			!request.StartDate.null && !request.EndDateExclusive.null
	}
	return !request.StartsAt.null && !request.EndsAt.null &&
		request.StartDate.null && request.EndDateExclusive.null
}

type calendarHTTPRSVPRequest struct {
	Status          string `json:"status"`
	AttendeeVersion uint   `json:"attendee_version"`
}

type calendarHTTPMutationResult struct {
	EventID       uuid.UUID                   `json:"event_id"`
	Version       uint                        `json:"version"`
	Sequence      uint                        `json:"sequence"`
	Status        CalendarEventStatus         `json:"status"`
	ChangedFields []CalendarEventChangedField `json:"changed_fields"`
}

type calendarHTTPOccurrenceCancelResponse struct {
	EventID      uuid.UUID           `json:"event_id"`
	RecurrenceID string              `json:"recurrence_id"`
	Status       CalendarEventStatus `json:"status"`
	Sequence     uint                `json:"sequence"`
	EventVersion uint                `json:"event_version"`
}

func calendarHTTPMutationResponse(result CalendarEventUpdateResult) (calendarHTTPMutationResult, error) {
	changedFields := result.ChangedFields()
	if _, err := NewCalendarEventUpdateResult(
		result.EventID(), result.Version(), result.Sequence(), result.Status(), changedFields,
	); err != nil {
		return calendarHTTPMutationResult{}, err
	}
	if changedFields == nil {
		changedFields = make([]CalendarEventChangedField, 0)
	}
	return calendarHTTPMutationResult{
		EventID: result.EventID(), Version: result.Version(), Sequence: result.Sequence(),
		Status: result.Status(), ChangedFields: changedFields,
	}, nil
}

func calendarHTTPActor(c fiber.Ctx) (uuid.UUID, error) {
	actorID, err := httputil.UserIDFromCtx(c)
	if err != nil || actorID == uuid.Nil {
		if err != nil {
			return uuid.Nil, err
		}
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
	}
	return actorID, nil
}

func calendarHTTPDecode(c fiber.Ctx, target any) error {
	body := c.Body()
	if len(body) == 0 || len(body) > calendarHTTPMaxBodyBytes || target == nil {
		return calendarHTTPBadRequest()
	}
	if err := calendarHTTPValidateJSON(body); err != nil {
		return calendarHTTPBadRequest()
	}
	if err := calendarHTTPValidateExactKeys(body, target); err != nil {
		return calendarHTTPBadRequest()
	}
	if err := calendarHTTPDecodeValue(body, target); err != nil {
		return calendarHTTPBadRequest()
	}
	return nil
}

func calendarHTTPValidateExactKeys(data []byte, target any) error {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(data, &object); err != nil || object == nil {
		return errors.New("invalid calendar JSON object")
	}
	var allowed func(string) bool
	validateAttendees := false
	switch target.(type) {
	case *calendarHTTPCreateRequest:
		allowed = calendarHTTPCreateKeyAllowed
		validateAttendees = true
	case *calendarHTTPUpdateRequest:
		allowed = calendarHTTPUpdateKeyAllowed
		validateAttendees = true
	case *calendarHTTPRSVPRequest:
		allowed = calendarHTTPRSVPKeyAllowed
	case *calendarHTTPAvailabilityRequest:
		allowed = calendarHTTPAvailabilityKeyAllowed
	case *calendarHTTPMeetingCreateRequest:
		allowed = calendarHTTPMeetingCreateKeyAllowed
	default:
		return errors.New("unsupported calendar request type")
	}
	for key := range object {
		if !allowed(key) {
			return errors.New("unknown calendar JSON key")
		}
	}
	if validateAttendees {
		if raw, present := object["attendees"]; present {
			return calendarHTTPValidateAttendeeKeys(raw)
		}
	}
	return nil
}

func calendarHTTPCreateKeyAllowed(key string) bool {
	switch key {
	case "title", "description", "location", "starts_at", "ends_at",
		"start_date", "end_date_exclusive", "all_day", "timezone", "rrule",
		"exdates", "scheduled_call_id", "reminder_minutes", "attendees":
		return true
	default:
		return false
	}
}

func calendarHTTPUpdateKeyAllowed(key string) bool {
	if key == "scheduled_call_id" {
		return false
	}
	return calendarHTTPCreateKeyAllowed(key)
}

func calendarHTTPRSVPKeyAllowed(key string) bool {
	return key == "status" || key == "attendee_version"
}

func calendarHTTPAvailabilityKeyAllowed(key string) bool {
	switch key {
	case "user_ids", "from", "to", "timezone", "slot_minutes":
		return true
	default:
		return false
	}
}

func calendarHTTPValidateAttendeeKeys(data []byte) error {
	trimmed := bytes.TrimSpace(data)
	if len(trimmed) == 0 || trimmed[0] != '[' {
		return nil
	}
	var attendees []json.RawMessage
	if err := json.Unmarshal(trimmed, &attendees); err != nil {
		return err
	}
	for _, raw := range attendees {
		element := bytes.TrimSpace(raw)
		if len(element) == 0 || element[0] != '{' {
			return errors.New("invalid calendar attendee object")
		}
		var object map[string]json.RawMessage
		if err := json.Unmarshal(element, &object); err != nil || object == nil {
			return errors.New("invalid calendar attendee object")
		}
		for key := range object {
			switch key {
			case "user_id", "email", "display_name", "role":
			default:
				return errors.New("unknown calendar attendee JSON key")
			}
		}
	}
	return nil
}

func calendarHTTPDecodeValue(data []byte, target any) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("multiple calendar JSON values")
		}
		return err
	}
	return nil
}

type calendarHTTPJSONFrame struct {
	kind         json.Delim
	expectingKey bool
	keys         map[string]struct{}
}

func calendarHTTPValidateJSON(data []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(data))
	frames := make([]calendarHTTPJSONFrame, 0, calendarHTTPMaxJSONDepth)
	rootSeen := false
	rootClosed := false
	tokens := 0
	for {
		token, err := decoder.Token()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		tokens++
		if tokens > calendarHTTPMaxJSONTokens {
			return errors.New("calendar JSON token limit exceeded")
		}
		if !rootSeen {
			delim, ok := token.(json.Delim)
			if !ok || delim != '{' {
				return errors.New("calendar JSON body must be an object")
			}
			rootSeen = true
			frames = append(frames, calendarHTTPJSONFrame{kind: delim, expectingKey: true, keys: make(map[string]struct{})})
			continue
		}
		if rootClosed || len(frames) == 0 {
			return errors.New("multiple calendar JSON values")
		}
		frame := &frames[len(frames)-1]
		if delim, ok := token.(json.Delim); ok {
			switch delim {
			case '{', '[':
				if frame.kind == '{' {
					if frame.expectingKey {
						return errors.New("invalid calendar JSON object")
					}
					frame.expectingKey = true
				}
				if len(frames) >= calendarHTTPMaxJSONDepth {
					return errors.New("calendar JSON depth limit exceeded")
				}
				nested := calendarHTTPJSONFrame{kind: delim}
				if delim == '{' {
					nested.expectingKey = true
					nested.keys = make(map[string]struct{})
				}
				frames = append(frames, nested)
			case '}', ']':
				if delim == '}' && (frame.kind != '{' || !frame.expectingKey) ||
					delim == ']' && frame.kind != '[' {
					return errors.New("invalid calendar JSON delimiter")
				}
				frames = frames[:len(frames)-1]
				if len(frames) == 0 {
					rootClosed = true
				}
			default:
				return errors.New("invalid calendar JSON delimiter")
			}
			continue
		}
		if frame.kind == '{' {
			if frame.expectingKey {
				key, ok := token.(string)
				if !ok {
					return errors.New("invalid calendar JSON key")
				}
				if _, duplicate := frame.keys[key]; duplicate {
					return errors.New("duplicate calendar JSON key")
				}
				frame.keys[key] = struct{}{}
				frame.expectingKey = false
			} else {
				frame.expectingKey = true
			}
		}
	}
	if !rootSeen || !rootClosed || len(frames) != 0 {
		return errors.New("incomplete calendar JSON body")
	}
	return nil
}

func calendarHTTPIdempotencyKey(c fiber.Ctx) (string, error) {
	key := strings.Clone(c.Get("Idempotency-Key"))
	if !validCalendarCreateIdempotencyKey(key) {
		return "", calendarHTTPBadRequest()
	}
	return key, nil
}

func calendarHTTPIfMatch(c fiber.Ctx) (uint, error) {
	raw := strings.Clone(c.Get("If-Match"))
	if len(raw) < 3 || raw[0] != '"' || raw[len(raw)-1] != '"' {
		return 0, calendarHTTPBadRequest()
	}
	digits := raw[1 : len(raw)-1]
	if digits == "" || digits[0] == '0' {
		return 0, calendarHTTPBadRequest()
	}
	for index := range digits {
		if digits[index] < '0' || digits[index] > '9' {
			return 0, calendarHTTPBadRequest()
		}
	}
	parsed, err := strconv.ParseUint(digits, 10, strconv.IntSize)
	if err != nil || parsed == 0 {
		return 0, calendarHTTPBadRequest()
	}
	return uint(parsed), nil
}

func calendarHTTPNotifyAttendees(c fiber.Ctx) (bool, error) {
	if !c.Request().URI().QueryArgs().Has("notify_attendees") {
		return true, nil
	}
	switch c.Query("notify_attendees") {
	case "true":
		return true, nil
	case "false":
		return false, nil
	default:
		return false, calendarHTTPBadRequest()
	}
}

func calendarHTTPRecurrenceID(c fiber.Ctx) (string, error) {
	raw := strings.Clone(c.Params("recurrenceId"))
	if len(raw) == 0 || len(raw) > 342 {
		return "", calendarHTTPBadRequest()
	}
	for index := range raw {
		character := raw[index]
		if character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z' ||
			character >= '0' && character <= '9' || character == '_' || character == '-' {
			continue
		}
		return "", calendarHTTPBadRequest()
	}
	decoded, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil || len(decoded) == 0 || len(decoded) > calendarEventOccurrenceCancelMaxRecurrenceBytes ||
		base64.RawURLEncoding.EncodeToString(decoded) != raw {
		return "", calendarHTTPBadRequest()
	}
	recurrenceID := strings.Clone(string(decoded))
	if !validCalendarEventOccurrenceCancelCanonicalKey(recurrenceID) {
		return "", calendarHTTPBadRequest()
	}
	return recurrenceID, nil
}

func calendarHTTPSendReceipt(c fiber.Ctx, response MutationReceiptResponse, expectedStatus int) error {
	body := response.JSON()
	if response.Status() != expectedStatus || len(body) == 0 || !json.Valid(body) {
		return calendarHTTPInternal()
	}
	c.Set("Content-Type", "application/json")
	return c.Status(response.Status()).Send(body)
}

func calendarHTTPBadRequest() error {
	return fiber.NewError(fiber.StatusBadRequest, "invalid calendar request")
}

func calendarHTTPUnprocessable() error {
	return fiber.NewError(fiber.StatusUnprocessableEntity, "calendar event is invalid")
}

func calendarHTTPInternal() error {
	return fiber.NewError(fiber.StatusInternalServerError, "calendar operation failed")
}

func calendarHTTPCreateError(c fiber.Ctx, err error) error {
	var conflict *CalendarEventCreateConflictError
	if errors.As(err, &conflict) && conflict != nil {
		switch conflict.Kind() {
		case CalendarEventCreateConflictIdempotencyKeyReused:
			return calendarHTTPConflict(c, "IDEMPOTENCY_KEY_REUSED", "calendar request conflicts with an existing receipt")
		case CalendarEventCreateConflictScheduledCallLinkMismatch:
			return calendarHTTPConflict(c, "SCHEDULED_CALL_LINK_MISMATCH", "scheduled call does not match the calendar event")
		case CalendarEventCreateConflictScheduledCallNotAttachable:
			return calendarHTTPConflict(c, "SCHEDULED_CALL_NOT_ATTACHABLE", "scheduled call cannot be attached")
		default:
			return calendarHTTPInternal()
		}
	}
	return calendarHTTPCommonError(err)
}

func calendarHTTPUpdateError(c fiber.Ctx, err error) error {
	conflict, found := calendarHTTPUpdateConflict(err)
	if !found {
		return calendarHTTPCommonError(err)
	}
	switch conflict.Kind() {
	case CalendarEventUpdateVersionConflict:
		return calendarHTTPVersionConflict(c, conflict, "CALENDAR_EVENT_VERSION_CONFLICT", "calendar event version is stale")
	case CalendarEventUpdateNotMutable:
		return calendarHTTPConflict(c, "CALENDAR_EVENT_NOT_MUTABLE", "calendar event cannot be changed")
	case CalendarEventUpdateRecurrenceExceptionConflict:
		return calendarHTTPConflict(c, "CALENDAR_EVENT_RECURRENCE_CONFLICT", "calendar event recurrence cannot be changed")
	default:
		return calendarHTTPInternal()
	}
}

func calendarHTTPCancelError(c fiber.Ctx, err error) error {
	conflict, found := calendarHTTPUpdateConflict(err)
	if !found {
		return calendarHTTPCommonError(err)
	}
	switch conflict.Kind() {
	case CalendarEventUpdateVersionConflict:
		return calendarHTTPVersionConflict(c, conflict, "CALENDAR_EVENT_VERSION_CONFLICT", "calendar event version is stale")
	case CalendarEventUpdateNotMutable:
		return calendarHTTPConflict(c, "CALENDAR_EVENT_NOT_MUTABLE", "calendar event cannot be cancelled")
	default:
		return calendarHTTPInternal()
	}
}

func calendarHTTPSeriesRSVPError(c fiber.Ctx, err error) error {
	conflict, found := calendarHTTPSeriesRSVPConflict(err)
	if !found {
		return calendarHTTPCommonError(err)
	}
	switch conflict.Kind() {
	case CalendarEventSeriesRSVPAttendeeVersionConflict:
		return calendarHTTPVersionConflict(c, conflict, "ATTENDEE_VERSION_CONFLICT", "calendar attendee version is stale")
	case CalendarEventSeriesRSVPIdempotencyKeyReused:
		return calendarHTTPConflict(c, "IDEMPOTENCY_KEY_REUSED", "calendar request conflicts with an existing receipt")
	default:
		return calendarHTTPInternal()
	}
}

func calendarHTTPOccurrenceRSVPError(c fiber.Ctx, err error) error {
	conflict, found := calendarHTTPOccurrenceRSVPConflict(err)
	if !found {
		return calendarHTTPCommonError(err)
	}
	switch conflict.Kind() {
	case CalendarEventOccurrenceRSVPAttendeeVersionConflict:
		return calendarHTTPVersionConflict(c, conflict, "ATTENDEE_VERSION_CONFLICT", "calendar attendee version is stale")
	case CalendarEventOccurrenceRSVPIdempotencyKeyReused:
		return calendarHTTPConflict(c, "IDEMPOTENCY_KEY_REUSED", "calendar request conflicts with an existing receipt")
	default:
		return calendarHTTPInternal()
	}
}

func calendarHTTPOccurrenceCancelError(c fiber.Ctx, err error) error {
	conflict, found := calendarHTTPOccurrenceCancelConflict(err)
	if !found {
		return calendarHTTPCommonError(err)
	}
	switch conflict.Kind() {
	case CalendarEventOccurrenceCancelVersionConflict:
		return calendarHTTPVersionConflict(c, conflict, "CALENDAR_EVENT_VERSION_CONFLICT", "calendar event version is stale")
	case CalendarEventOccurrenceCancelNotMutable:
		return calendarHTTPConflict(c, "CALENDAR_OCCURRENCE_NOT_MUTABLE", "calendar occurrence cannot be cancelled")
	default:
		return calendarHTTPInternal()
	}
}

type calendarHTTPVersionConflictMetadata interface {
	EventID() (uuid.UUID, bool)
	CurrentVersion() (uint, bool)
}

func calendarHTTPVersionConflict(c fiber.Ctx, conflict calendarHTTPVersionConflictMetadata, code, message string) error {
	eventID, hasEventID := conflict.EventID()
	currentVersion, hasVersion := conflict.CurrentVersion()
	if !hasEventID || eventID == uuid.Nil || !hasVersion || currentVersion == 0 {
		return calendarHTTPInternal()
	}
	c.Set("ETag", strconv.Quote(strconv.FormatUint(uint64(currentVersion), 10)))
	c.Set("Location", fmt.Sprintf("/api/calendar/v2/events/%s", eventID.String()))
	return calendarHTTPConflict(c, code, message)
}

func calendarHTTPConflict(c fiber.Ctx, code, message string) error {
	if code == "" || message == "" {
		return calendarHTTPInternal()
	}
	c.Set("X-Hub-Error-Code", code)
	return fiber.NewError(fiber.StatusConflict, message)
}

func calendarHTTPCommonError(err error) error {
	switch {
	case errors.Is(err, ErrEventNotFound), errors.Is(err, ErrEventForbidden):
		return fiber.NewError(fiber.StatusNotFound, "calendar event not found")
	case errors.Is(err, ErrEventInvalid):
		return calendarHTTPUnprocessable()
	default:
		return calendarHTTPInternal()
	}
}

func calendarHTTPUpdateConflict(err error) (CalendarEventUpdateConflictError, bool) {
	var value CalendarEventUpdateConflictError
	if errors.As(err, &value) {
		return value, true
	}
	var pointer *CalendarEventUpdateConflictError
	if errors.As(err, &pointer) && pointer != nil {
		return *pointer, true
	}
	return CalendarEventUpdateConflictError{}, false
}

func calendarHTTPSeriesRSVPConflict(err error) (CalendarEventSeriesRSVPConflictError, bool) {
	var value CalendarEventSeriesRSVPConflictError
	if errors.As(err, &value) {
		return value, true
	}
	var pointer *CalendarEventSeriesRSVPConflictError
	if errors.As(err, &pointer) && pointer != nil {
		return *pointer, true
	}
	return CalendarEventSeriesRSVPConflictError{}, false
}

func calendarHTTPOccurrenceRSVPConflict(err error) (CalendarEventOccurrenceRSVPConflictError, bool) {
	var value CalendarEventOccurrenceRSVPConflictError
	if errors.As(err, &value) {
		return value, true
	}
	var pointer *CalendarEventOccurrenceRSVPConflictError
	if errors.As(err, &pointer) && pointer != nil {
		return *pointer, true
	}
	return CalendarEventOccurrenceRSVPConflictError{}, false
}

func calendarHTTPOccurrenceCancelConflict(err error) (CalendarEventOccurrenceCancelConflictError, bool) {
	var value CalendarEventOccurrenceCancelConflictError
	if errors.As(err, &value) {
		return value, true
	}
	var pointer *CalendarEventOccurrenceCancelConflictError
	if errors.As(err, &pointer) && pointer != nil {
		return *pointer, true
	}
	return CalendarEventOccurrenceCancelConflictError{}, false
}

func nilCalendarHTTPValue(value any) bool {
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

func (h *EventHandler) ready() bool {
	return h != nil && h.service != nil &&
		!nilCalendarHTTPValue(h.service.repository) &&
		!nilCalendarHTTPValue(h.service.clock) &&
		!nilCalendarHTTPValue(h.service.ids)
}
