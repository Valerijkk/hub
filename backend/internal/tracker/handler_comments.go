package tracker

import (
	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Comment handlers
// ---------------------------------------------------------------------------

// CreateComment обрабатывает POST /api/projects/:projectId/issues/:issueKey/comments
func (h *Handler) CreateComment(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateCommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateComment(c.Context(), issueKey, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListComments обрабатывает GET /api/projects/:projectId/issues/:issueKey/comments
func (h *Handler) ListComments(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.ListComments(c.Context(), issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateComment обрабатывает PUT /api/projects/:projectId/issues/:issueKey/comments/:commentId
func (h *Handler) UpdateComment(c fiber.Ctx) error {
	commentID, err := httputil.ParseUUIDParam(c, "commentId")
	if err != nil {
		return err
	}

	var req UpdateCommentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateComment(c.Context(), commentID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteComment обрабатывает DELETE /api/projects/:projectId/issues/:issueKey/comments/:commentId
func (h *Handler) DeleteComment(c fiber.Ctx) error {
	commentID, err := httputil.ParseUUIDParam(c, "commentId")
	if err != nil {
		return err
	}

	if err := h.service.DeleteComment(c.Context(), commentID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ToggleReaction обрабатывает POST /api/projects/:projectId/issues/:issueKey/comments/:commentId/reactions
func (h *Handler) ToggleReaction(c fiber.Ctx) error {
	commentID, err := httputil.ParseUUIDParam(c, "commentId")
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreateReactionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	added, err := h.service.ToggleReaction(c.Context(), commentID, userID, req.Emoji)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"added": added})
}

// ---------------------------------------------------------------------------
// Activity handler
// ---------------------------------------------------------------------------

// GetActivity обрабатывает GET /api/projects/:projectId/issues/:issueKey/activity
func (h *Handler) GetActivity(c fiber.Ctx) error {
	issueKey := parseIssueKey(c)
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.GetActivity(c.Context(), issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
