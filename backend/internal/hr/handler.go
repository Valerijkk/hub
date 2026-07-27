package hr

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// Handler — HTTP-обработчики HR-lite.
type Handler struct {
	service *Service
}

// NewHandler создаёт handler HR.
func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует маршруты /api/hr.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	hrGroup := api.Group("/hr", authMw)
	hrGroup.Get("/leave", h.ListMyLeave)
	hrGroup.Get("/leave/pending", h.ListPending)
	hrGroup.Post("/leave", h.CreateLeave)
	hrGroup.Post("/leave/:leaveId/cancel", h.CancelLeave)
	hrGroup.Post("/leave/:leaveId/approve", h.ApproveLeave)
	hrGroup.Post("/leave/:leaveId/reject", h.RejectLeave)
}

func (h *Handler) ListMyLeave(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	items, err := h.service.ListMyLeave(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list leave requests")
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *Handler) ListPending(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	items, err := h.service.ListPending(c.Context(), userID)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(fiber.Map{"items": items})
}

func (h *Handler) CreateLeave(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	var req CreateLeaveRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if req.LeaveType == "" {
		req.LeaveType = LeaveVacation
	}
	resp, err := h.service.CreateLeave(c.Context(), userID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

func (h *Handler) ApproveLeave(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	leaveID, err := parseLeaveID(c)
	if err != nil {
		return err
	}
	var req ReviewLeaveRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.ApproveLeave(c.Context(), userID, leaveID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) RejectLeave(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	leaveID, err := parseLeaveID(c)
	if err != nil {
		return err
	}
	var req ReviewLeaveRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.RejectLeave(c.Context(), userID, leaveID, req)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func (h *Handler) CancelLeave(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}
	leaveID, err := parseLeaveID(c)
	if err != nil {
		return err
	}
	var body struct {
		Version int `json:"version"`
	}
	// A6-006: bind errors → 400 (same as Approve/Reject; do not swallow).
	if err := c.Bind().JSON(&body); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	resp, err := h.service.CancelLeave(c.Context(), userID, leaveID, body.Version)
	if err != nil {
		return mapServiceError(err)
	}
	return c.JSON(resp)
}

func parseLeaveID(c fiber.Ctx) (uuid.UUID, error) {
	raw := c.Params("leaveId")
	id, err := uuid.Parse(raw)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "invalid leave id")
	}
	return id, nil
}

func mapServiceError(err error) error {
	switch {
	case errors.Is(err, ErrNotFound):
		return fiber.NewError(fiber.StatusNotFound, "leave request not found")
	case errors.Is(err, ErrForbidden):
		return fiber.NewError(fiber.StatusForbidden, "admin access required")
	case errors.Is(err, ErrInvalidInput):
		return fiber.NewError(fiber.StatusBadRequest, "invalid leave request input")
	case errors.Is(err, ErrVersionConflict):
		return fiber.NewError(fiber.StatusConflict, "leave request version conflict")
	case errors.Is(err, ErrOverlap):
		return fiber.NewError(fiber.StatusConflict, "leave dates overlap")
	default:
		return fiber.NewError(fiber.StatusInternalServerError, "hr operation failed")
	}
}
