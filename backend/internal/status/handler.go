package status

import (
	"github.com/gofiber/fiber/v3"
)

// Handler — публичный HTTP API status page (без auth).
type Handler struct {
	service *Service
}

// NewHandler создаёт handler status page.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует GET /api/status.
func (h *Handler) RegisterRoutes(api fiber.Router) {
	api.Get("/status", h.GetStatus)
}

// GetStatus godoc
// @Summary Public status page data
// @Tags Infrastructure
// @Produce json
// @Success 200 {object} Response
// @Router /api/status [get]
func (h *Handler) GetStatus(c fiber.Ctx) error {
	resp := h.service.GetStatus(c.Context())
	code := fiber.StatusOK
	if resp.Status != "operational" {
		code = fiber.StatusServiceUnavailable
	}
	return c.Status(code).JSON(resp)
}
