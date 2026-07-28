package ai

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// UserHandler — маршруты BYOK (Phase 17 W4).
type UserHandler struct {
	svc *Service
}

// NewUserHandler создаёт user AI handler.
func NewUserHandler(svc *Service) *UserHandler {
	return &UserHandler{svc: svc}
}

// RegisterRoutes монтирует /ai/my/*.
func (h *UserHandler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	my := api.Group("/ai/my", authMw)
	my.Get("/providers", h.ListProviders)
	my.Post("/providers", h.CreateProvider)
	my.Patch("/providers/:id", h.UpdateProvider)
}

// ListProviders GET /api/ai/my/providers
func (h *UserHandler) ListProviders(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	items, err := h.svc.ListUserProviders(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list providers")
	}
	return c.JSON(items)
}

// CreateProvider POST /api/ai/my/providers
func (h *UserHandler) CreateProvider(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateProviderRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	if req.Name == "" || req.BaseURL == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name and base_url required")
	}
	resp, err := h.svc.CreateUserProvider(c.Context(), userID, req)
	if err != nil {
		// A2W5-A1-003: стабильные коды, без encrypt/base_url internals.
		status, msg := mapProviderMutateError(err)
		return fiber.NewError(status, msg)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// UpdateProvider PATCH /api/ai/my/providers/:id
func (h *UserHandler) UpdateProvider(c fiber.Ctx) error {
	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}
	id, err := uuid.Parse(c.Params("id"))
	if err != nil || id == uuid.Nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid id")
	}
	var req UpdateProviderRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid body")
	}
	resp, err := h.svc.UpdateUserProvider(c.Context(), userID, id, req)
	if err != nil {
		if ent.IsNotFound(err) {
			return fiber.NewError(fiber.StatusNotFound, "provider not found")
		}
		// A2W5-A1-003: стабильные коды, без encrypt/base_url internals.
		status, msg := mapProviderMutateError(err)
		return fiber.NewError(status, msg)
	}
	return c.JSON(resp)
}
