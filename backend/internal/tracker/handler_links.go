package tracker

import (
	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Issue link handlers
// ---------------------------------------------------------------------------

// CreateIssueLink обрабатывает POST /api/projects/:projectId/issues/:issueKey/links
func (h *Handler) CreateIssueLink(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateIssueLinkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateIssueLink(c.Context(), issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListIssueLinks обрабатывает GET /api/projects/:projectId/issues/:issueKey/links
func (h *Handler) ListIssueLinks(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.ListIssueLinks(c.Context(), issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteIssueLink обрабатывает DELETE /api/projects/:projectId/issues/:issueKey/links/:linkId
func (h *Handler) DeleteIssueLink(c fiber.Ctx) error {
	linkID, err := httputil.ParseUUIDParam(c, "linkId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteIssueLink(c.Context(), linkID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}
