package status

import (
	"github.com/gofiber/fiber/v3"
)

// Handler — HTTP API status page.
type Handler struct {
	service *Service
}

// NewHandler создаёт handler status page.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует публичный aggregate и детальный (auth) статус.
// W7-A5-002: anon не видит per-component outage (recon).
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	api.Get("/status", h.GetStatus)
	// Fiber v3: middleware после handler в сигнатуре выполняется ДО handler.
	api.Get("/status/details", h.GetStatusDetails, authMw)
}

// GetStatus godoc
// @Summary Public status page (aggregate only)
// @Tags Infrastructure
// @Produce json
// @Success 200 {object} Response
// @Router /api/status [get]
func (h *Handler) GetStatus(c fiber.Ctx) error {
	resp := h.service.GetPublicStatus(c.Context())
	code := fiber.StatusOK
	if resp.Status != "operational" {
		code = fiber.StatusServiceUnavailable
	}
	return c.Status(code).JSON(resp)
}

// GetStatusDetails godoc
// @Summary Detailed status page (auth)
// @Tags Infrastructure
// @Produce json
// @Success 200 {object} Response
// @Router /api/status/details [get]
// @Security BearerAuth
func (h *Handler) GetStatusDetails(c fiber.Ctx) error {
	resp := h.service.GetDetailedStatus(c.Context())
	code := fiber.StatusOK
	if resp.Status != "operational" {
		code = fiber.StatusServiceUnavailable
	}
	return c.Status(code).JSON(resp)
}
