package onec

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// AdminHandler — admin API коннектора 1С (MOD-07).
type AdminHandler struct {
	svc       *Service
	isAdminFn func(c fiber.Ctx, userID uuid.UUID) (bool, error)
}

// NewAdminHandler создаёт admin-обработчик.
func NewAdminHandler(svc *Service, isAdminFn func(c fiber.Ctx, userID uuid.UUID) (bool, error)) *AdminHandler {
	return &AdminHandler{svc: svc, isAdminFn: isAdminFn}
}

// RegisterRoutes регистрирует /admin/integrations/onec/*.
func (h *AdminHandler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	g := api.Group("/admin/integrations/onec", authMw, h.requireAdmin())
	g.Get("/settings", h.GetSettings)
	g.Put("/settings", h.UpdateSettings)
	g.Post("/test", h.TestWebhook)
}

func (h *AdminHandler) requireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		userIDStr, ok := c.Locals("userID").(string)
		if !ok || userIDStr == "" {
			return fiber.NewError(fiber.StatusUnauthorized, "unauthorized")
		}
		userID, err := uuid.Parse(userIDStr)
		if err != nil {
			return fiber.NewError(fiber.StatusUnauthorized, "invalid user id")
		}
		isAdmin, err := h.isAdminFn(c, userID)
		if err != nil || !isAdmin {
			return fiber.NewError(fiber.StatusForbidden, "admin only")
		}
		return c.Next()
	}
}

func (h *AdminHandler) GetSettings(c fiber.Ctx) error {
	resp, err := h.svc.GetSettings(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load settings")
	}
	return c.JSON(resp)
}

func (h *AdminHandler) UpdateSettings(c fiber.Ctx) error {
	var req UpdateSettingsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	resp, err := h.svc.UpdateSettings(c.Context(), req)
	if err != nil {
		if err == ErrInvalidConfig {
			return fiber.NewError(fiber.StatusBadRequest, "invalid 1c connector config")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to save settings")
	}
	return c.JSON(resp)
}

func (h *AdminHandler) TestWebhook(c fiber.Ctx) error {
	resp, err := h.svc.TestWebhook(c.Context())
	if err != nil {
		if err == ErrNotConfigured {
			return fiber.NewError(fiber.StatusBadRequest, "webhook url not configured")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "test failed")
	}
	return c.JSON(resp)
}
