package calendar

import (
	"errors"
	"time"

	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// Handler — HTTP API единого календаря.
type Handler struct {
	service *Service
}

// NewHandler создаёт handler календаря.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует маршруты /api/calendar.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	cal := api.Group("/calendar", authMw)
	cal.Get("/events", h.ListEvents)
	cal.Get("/export.ics", h.ExportICal)
}

// ListEvents godoc
// @Summary События единого календаря
// @Description Канонические события, звонки, спринты и дедлайны задач за период (MOD-04)
// @Tags Calendar
// @Produce json
// @Param from query string true "Начало периода (RFC3339)"
// @Param to query string true "Конец периода (RFC3339)"
// @Success 200 {object} EventsResponse
// @Security BearerAuth
// @Router /api/calendar/events [get]
func (h *Handler) ListEvents(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	from, to, err := parseRange(c)
	if err != nil {
		return err
	}
	resp, err := h.service.ListEvents(c.Context(), userID, from, to)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

// ExportICal godoc
// @Summary Экспорт iCalendar
// @Tags Calendar
// @Produce text/calendar
// @Param from query string true "Начало периода (RFC3339)"
// @Param to query string true "Конец периода (RFC3339)"
// @Success 200 {string} string "text/calendar"
// @Security BearerAuth
// @Router /api/calendar/export.ics [get]
func (h *Handler) ExportICal(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	from, to, err := parseRange(c)
	if err != nil {
		return err
	}
	resp, err := h.service.ListEvents(c.Context(), userID, from, to)
	if err != nil {
		return mapServiceError(err)
	}
	body := RenderICalendar(resp.Events)
	c.Set("Content-Type", "text/calendar; charset=utf-8")
	c.Set("Content-Disposition", `attachment; filename="hub-calendar.ics"`)
	return c.SendString(body)
}

func parseRange(c fiber.Ctx) (time.Time, time.Time, error) {
	fromRaw := c.Query("from")
	toRaw := c.Query("to")
	if fromRaw == "" || toRaw == "" {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "from and to query params required")
	}
	from, err := time.Parse(time.RFC3339, fromRaw)
	if err != nil {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "invalid from timestamp")
	}
	to, err := time.Parse(time.RFC3339, toRaw)
	if err != nil {
		return time.Time{}, time.Time{}, fiber.NewError(fiber.StatusBadRequest, "invalid to timestamp")
	}
	return from, to, nil
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, ErrInvalidRange):
		return fiber.NewError(fiber.StatusBadRequest, "invalid calendar range")
	case errors.Is(err, ErrRangeTooWide):
		return fiber.NewError(fiber.StatusBadRequest, "calendar range too wide")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "calendar operation failed")
	}
}
