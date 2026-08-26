package tracker

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ListIssueMessageLinks — Связи задачи с тредами
// @Summary Связи задачи с тредами
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Success 200 {array} IssueMessageLinkResponse
// @Security BearerAuth
// @Router /projects/{projectId}/issues/{issueKey}/message-links [get]
func (h *Handler) ListIssueMessageLinks(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	issueKey := parseIssueKey(c)
	resp, err := h.service.ListIssueMessageLinks(c.Context(), projectID, issueKey)
	if err != nil {
		return err
	}
	return c.JSON(resp)
}

// CreateIssueMessageLink — Связать задачу с тредом
// @Summary Связать задачу с тредом
// @Tags Трекер
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Param body body CreateIssueMessageLinkRequest true "Link"
// @Success 201 {object} IssueMessageLinkResponse
// @Security BearerAuth
// @Router /projects/{projectId}/issues/{issueKey}/message-links [post]
func (h *Handler) CreateIssueMessageLink(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}
	issueKey := parseIssueKey(c)
	var req CreateIssueMessageLinkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}
	resp, err := h.service.CreateIssueMessageLink(c.Context(), projectID, issueKey, userID, req)
	if err != nil {
		return err
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// DeleteIssueMessageLink — Удалить связь задачи с тредом
// @Summary Удалить связь задачи с тредом
// @Tags Трекер
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Param linkId path string true "ID связи"
// @Success 204 "No Content"
// @Security BearerAuth
// @Router /projects/{projectId}/issues/{issueKey}/message-links/{linkId} [delete]
func (h *Handler) DeleteIssueMessageLink(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	issueKey := parseIssueKey(c)
	linkID, err := uuid.Parse(c.Params("linkId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid link id")
	}
	if err := h.service.DeleteIssueMessageLink(c.Context(), projectID, issueKey, linkID); err != nil {
		return err
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// ListThreadIssueLinks — Связи треда с задачами
// @Summary Связи треда с задачами
// @Tags Трекер
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param channelId query string true "Channel ID"
// @Param messageId query string true "Thread root message ID"
// @Success 200 {array} IssueMessageLinkResponse
// @Security BearerAuth
// @Router /projects/{projectId}/message-links [get]
func (h *Handler) ListThreadIssueLinks(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	actorID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, apperr.MsgUnauthorized)
	}
	channelStr := c.Query("channel_id")
	messageID := c.Query("message_id")
	if channelStr == "" || messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "channel_id и message_id обязательны")
	}
	channelID, err := uuid.Parse(channelStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный channel_id")
	}
	resp, err := h.service.ListThreadIssueLinks(c.Context(), projectID, channelID, actorID, messageID)
	if err != nil {
		return err
	}
	return c.JSON(resp)
}
