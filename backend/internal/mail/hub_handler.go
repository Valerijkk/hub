package mail

import (
	"github.com/gofiber/fiber/v3"
)

// HubHandler — HTTP handlers микросервиса hub-mail.
type HubHandler struct {
	svc *Service
}

// NewHubHandler создаёт handler hub-mail.
func NewHubHandler(svc *Service) *HubHandler {
	return &HubHandler{svc: svc}
}

// RegisterRoutes регистрирует маршруты hub-mail (без /api prefix).
func (h *HubHandler) RegisterRoutes(app *fiber.App) {
	app.Get("/health", h.Health)
	app.Get("/v1/status", h.Status)
	v1 := app.Group("/v1", InternalTokenAuth(h.svc.cfg.InternalToken))
	v1.Post("/send", h.Send)
}

// Health — liveness probe для оркестратора.
func (h *HubHandler) Health(c fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "service": "hub-mail"})
}

// Status — конфигурация без секретов.
func (h *HubHandler) Status(c fiber.Ctx) error {
	return c.JSON(h.svc.Status())
}

// Send — транзакционная отправка (внутренний мост).
func (h *HubHandler) Send(c fiber.Ctx) error {
	var req SendRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if err := h.svc.SendAlert(c.Context(), req); err != nil {
		if err == ErrInvalidRequest {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "send failed")
	}
	return c.JSON(fiber.Map{"message": "queued"})
}
