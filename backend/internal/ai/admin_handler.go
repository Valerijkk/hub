package ai

import (
	"errors"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// AdminHandler — admin CRUD AI-провайдеров.
type AdminHandler struct {
	svc       *Service
	isAdminFn func(c fiber.Ctx, userID uuid.UUID) (bool, error)
}

// NewAdminHandler создаёт admin AI handler.
func NewAdminHandler(svc *Service, isAdminFn func(c fiber.Ctx, userID uuid.UUID) (bool, error)) *AdminHandler {
	return &AdminHandler{svc: svc, isAdminFn: isAdminFn}
}

// RegisterRoutes монтирует /admin/ai/*.
func (h *AdminHandler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	admin := api.Group("/admin/ai", authMw, h.requireAdmin())
	admin.Get("/providers", h.ListProviders)
	admin.Post("/providers", h.CreateProvider)
	admin.Patch("/providers/:id", h.UpdateProvider)
	admin.Post("/providers/:id/test", h.TestProvider)
	admin.Get("/usage/summary", h.GetUsageSummary)
	admin.Get("/usage/history", h.GetUsageHistory)
	admin.Get("/settings", h.GetSettings)
	admin.Patch("/settings", h.UpdateSettings)
}

func (h *AdminHandler) requireAdmin() fiber.Handler {
	return func(c fiber.Ctx) error {
		userID, err := userIDFromCtx(c)
		if err != nil {
			return err
		}
		isAdmin, err := h.isAdminFn(c, userID)
		if err != nil || !isAdmin {
			return fiber.NewError(fiber.StatusForbidden, "admin only")
		}
		return c.Next()
	}
}

// ListProviders GET /admin/ai/providers
func (h *AdminHandler) ListProviders(c fiber.Ctx) error {
	items, err := h.svc.ListProviders(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list providers")
	}
	return c.JSON(items)
}

// CreateProvider POST /admin/ai/providers
func (h *AdminHandler) CreateProvider(c fiber.Ctx) error {
	var req CreateProviderRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if req.Name == "" || req.BaseURL == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name and base_url required")
	}
	resp, err := h.svc.CreateProvider(c.Context(), req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateProvider PATCH /admin/ai/providers/:id
func (h *AdminHandler) UpdateProvider(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req UpdateProviderRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	resp, err := h.svc.UpdateProvider(c.Context(), id, req)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}
	return c.JSON(resp)
}

// TestProvider POST /admin/ai/providers/:id/test
func (h *AdminHandler) TestProvider(c fiber.Ctx) error {
	id, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	resp, err := h.svc.TestProvider(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "provider not found")
	}
	return c.JSON(resp)
}

// GetUsageSummary GET /admin/ai/usage/summary
func (h *AdminHandler) GetUsageSummary(c fiber.Ctx) error {
	resp, err := h.svc.GetUsageSummary(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load usage summary")
	}
	return c.JSON(resp)
}

// GetUsageHistory GET /admin/ai/usage/history?days=7
func (h *AdminHandler) GetUsageHistory(c fiber.Ctx) error {
	days := 7
	if raw := c.Query("days"); raw != "" {
		if n, err := strconv.Atoi(raw); err == nil {
			days = n
		}
	}
	resp, err := h.svc.GetUsageHistory(c.Context(), days)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load usage history")
	}
	return c.JSON(resp)
}

// GetSettings GET /admin/ai/settings
func (h *AdminHandler) GetSettings(c fiber.Ctx) error {
	resp, err := h.svc.GetAdminSettings(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to load ai settings")
	}
	return c.JSON(resp)
}

// UpdateSettings PATCH /admin/ai/settings
func (h *AdminHandler) UpdateSettings(c fiber.Ctx) error {
	var req UpdateAdminSettingsRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	resp, err := h.svc.UpdateAdminSettings(c.Context(), req)
	if err != nil {
		if errors.Is(err, ErrSettingsTooLong) {
			return fiber.NewError(fiber.StatusBadRequest, err.Error())
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to save ai settings")
	}
	return c.JSON(resp)
}
