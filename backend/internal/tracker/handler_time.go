package tracker

import (
	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Time tracking handlers
// ---------------------------------------------------------------------------

// LogTime обрабатывает POST /api/projects/:projectId/issues/:issueKey/time
func (h *Handler) LogTime(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req LogTimeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.LogTime(c.Context(), issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// StartTimer обрабатывает POST /api/projects/:projectId/issues/:issueKey/timer/start
func (h *Handler) StartTimer(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.StartTimer(c.Context(), issueKey, userID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// StopTimer обрабатывает POST /api/projects/:projectId/issues/:issueKey/timer/stop
func (h *Handler) StopTimer(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.StopTimer(c.Context(), issueKey, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ListTimeLogs обрабатывает GET /api/projects/:projectId/issues/:issueKey/time-logs
func (h *Handler) ListTimeLogs(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.ListTimeLogs(c.Context(), issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
