package calendar

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

// AvailabilityHandler exposes the privacy-preserving Calendar v2 availability
// endpoint. Authorization remains owned by AvailabilitySourceRepository.
type AvailabilityHandler struct {
	service *AvailabilityService
}

// NewAvailabilityHandler creates the Calendar v2 availability handler.
func NewAvailabilityHandler(service *AvailabilityService) *AvailabilityHandler {
	return &AvailabilityHandler{service: service}
}

// RegisterRoutes registers only the authenticated Calendar v2 availability route.
func (handler *AvailabilityHandler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	if !handler.ready() || authMw == nil || nilCalendarHTTPValue(api) {
		return
	}
	calendarV2 := api.Group("/calendar/v2", authMw)
	calendarV2.Post("/availability", handler.QueryAvailability)
}

// QueryAvailability returns source-redacted availability for one complete
// authorized user set.
func (handler *AvailabilityHandler) QueryAvailability(c fiber.Ctx) error {
	if !handler.ready() {
		return calendarHTTPInternal()
	}
	actorID, err := calendarHTTPActor(c)
	if err != nil {
		return err
	}
	var request calendarHTTPAvailabilityRequest
	if err := calendarHTTPDecode(c, &request); err != nil {
		return err
	}
	query, err := request.query()
	if err != nil {
		return calendarHTTPBadRequest()
	}
	result, err := handler.service.QueryAvailability(c.Context(), actorID, query)
	if err != nil {
		return calendarHTTPAvailabilityError(c, err)
	}
	response, err := calendarHTTPAvailabilityResponseFor(result, query.Timezone)
	if err != nil {
		return calendarHTTPInternal()
	}
	return c.Status(fiber.StatusOK).JSON(response)
}

func calendarHTTPAvailabilityError(c fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, ErrAvailabilityNotFound):
		return fiber.NewError(fiber.StatusNotFound, "calendar availability not found")
	case errors.Is(err, ErrAvailabilityLimitExceeded):
		return calendarHTTPAvailabilityUnprocessable(c, "AVAILABILITY_LIMIT_EXCEEDED")
	case errors.Is(err, ErrAvailabilitySourceInvalid):
		return calendarHTTPAvailabilityUnprocessable(c, "AVAILABILITY_SOURCE_INVALID")
	default:
		return calendarHTTPInternal()
	}
}

func calendarHTTPAvailabilityUnprocessable(c fiber.Ctx, code string) error {
	if code == "" {
		return calendarHTTPInternal()
	}
	c.Set("X-Hub-Error-Code", code)
	return fiber.NewError(fiber.StatusUnprocessableEntity, "calendar availability cannot be resolved")
}

func (handler *AvailabilityHandler) ready() bool {
	return handler != nil && handler.service != nil && handler.service.ready()
}
