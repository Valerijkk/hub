package calendar

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// EventReadHandler exposes only the authenticated Calendar v2 event-detail
// read route.
type EventReadHandler struct {
	service *EventReadService
}

// NewEventReadHandler creates the isolated Calendar event read handler.
func NewEventReadHandler(service *EventReadService) *EventReadHandler {
	return &EventReadHandler{service: service}
}

// RegisterRoutes registers GET /api/calendar/v2/events/:eventId.
func (handler *EventReadHandler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	if !handler.ready() || authMw == nil || nilCalendarHTTPValue(api) {
		return
	}
	api.Group("/calendar/v2", authMw).Get("/events/:eventId", handler.GetCalendarEvent)
}

// GetCalendarEvent returns one safe event detail for the authenticated actor.
func (handler *EventReadHandler) GetCalendarEvent(c fiber.Ctx) error {
	if !handler.ready() {
		return calendarEventReadInternal()
	}
	actorID, err := calendarHTTPActor(c)
	if err != nil {
		return err
	}
	if len(c.Request().URI().QueryString()) != 0 || len(c.BodyRaw()) != 0 || strings.TrimSpace(c.Get(fiber.HeaderContentType)) != "" {
		return calendarEventReadBadRequest()
	}
	eventID, err := calendarEventReadPathID(c)
	if err != nil {
		return calendarEventReadBadRequest()
	}
	detail, err := handler.service.GetCalendarEvent(c.Context(), actorID, eventID)
	if err != nil {
		return calendarEventReadHTTPError(err)
	}
	return c.Status(fiber.StatusOK).JSON(detail)
}

func calendarEventReadPathID(c fiber.Ctx) (uuid.UUID, error) {
	if nilCalendarHTTPValue(c) {
		return uuid.Nil, ErrEventInvalid
	}
	value := c.Params("eventId")
	eventID, err := uuid.Parse(value)
	if err != nil || eventID == uuid.Nil || value != eventID.String() {
		return uuid.Nil, ErrEventInvalid
	}
	return eventID, nil
}

func calendarEventReadHTTPError(err error) error {
	if errors.Is(err, ErrEventNotFound) {
		return fiber.NewError(fiber.StatusNotFound, "calendar event not found")
	}
	return calendarEventReadInternal()
}

func calendarEventReadBadRequest() error {
	return fiber.NewError(fiber.StatusBadRequest, "invalid calendar event request")
}

func calendarEventReadInternal() error {
	return fiber.NewError(fiber.StatusInternalServerError, "calendar event unavailable")
}

func (handler *EventReadHandler) ready() bool {
	return handler != nil && handler.service != nil && handler.service.ready()
}
