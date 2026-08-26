package messenger

import (
	"time"

	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчики сообщений
// ---------------------------------------------------------------------------

// SendMessage — Отправить сообщение
// @Summary Отправить сообщение
// @Description Отправляет новое сообщение в указанный канал. Поддерживает ответы в тредах.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID канала"
// @Param body body SendMessageRequest true "Данные сообщения"
// @Success 201 {object} MessageResponse "Сообщение отправлено"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/channels/{channelId}/messages [post]
func (h *Handler) SendMessage(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req SendMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.SendMessage(c.Context(), channelID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListMessages — Список сообщений канала
// @Summary Список сообщений канала
// @Description Возвращает сообщения канала с пагинацией по курсору (before) и лимиту.
// @Tags Мессенджер
// @Produce json
// @Param channelId path string true "ID канала"
// @Param before query string false "Курсор — дата RFC3339, сообщения до этой даты"
// @Param limit query int false "Количество сообщений (по умолчанию 50)"
// @Success 200 {array} MessageResponse "Список сообщений"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/channels/{channelId}/messages [get]
func (h *Handler) ListMessages(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var before *time.Time
	beforeStr := c.Query("before")
	if beforeStr != "" {
		t, err := time.Parse(time.RFC3339, beforeStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "некорректный формат before (ожидается RFC3339)")
		}
		before = &t
	}

	limit := queryInt(c, "limit", 50)

	resp, err := h.service.ListMessages(c.Context(), channelID, userID, before, limit)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// EditMessage — Редактировать сообщение
// @Summary Редактировать сообщение
// @Description Редактирует текст ранее отправленного сообщения. Доступно только автору.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID канала"
// @Param messageId path string true "ID сообщения"
// @Param body body EditMessageRequest true "Новый текст"
// @Success 200 {object} MessageResponse "Сообщение обновлено"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /messenger/channels/{channelId}/messages/{messageId} [put]
func (h *Handler) EditMessage(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req EditMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.EditMessage(c.Context(), channelID, messageID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteMessage — Удалить сообщение
// @Summary Удалить сообщение
// @Description Удаляет сообщение из канала. Доступно автору либо owner/admin/moderator канала.
// @Tags Мессенджер
// @Param channelId path string true "ID канала"
// @Param messageId path string true "ID сообщения"
// @Success 204 "Сообщение удалено"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /messenger/channels/{channelId}/messages/{messageId} [delete]
func (h *Handler) DeleteMessage(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.DeleteMessage(c.Context(), messageID, userID, channelID.String(), false); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ForwardMessage — Переслать сообщение
// @Summary Переслать сообщение
// @Description Пересылает сообщение в другой канал.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID исходного канала"
// @Param messageId path string true "ID сообщения"
// @Param body body ForwardMessageRequest true "ID целевого канала"
// @Success 201 {object} MessageResponse "Сообщение переслано"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/channels/{channelId}/messages/{messageId}/forward [post]
func (h *Handler) ForwardMessage(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req ForwardMessageRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.ForwardMessage(c.Context(), channelID, userID, messageID, req.TargetChannelID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// PromoteToIssue — Промоутнуть сообщение в задачу трекера
// @Summary Промоутнуть сообщение в задачу трекера
// @Description Создаёт новую задачу в указанном проекте на основе содержимого сообщения и его треда,
// @Description постит в исходный тред backlink-ссылку на созданную задачу.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param messageId path string true "ID сообщения"
// @Param body body PromoteToIssueRequest true "Параметры промоушена"
// @Success 201 {object} PromoteToIssueResponse "Задача создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Failure 404 {object} fiber.Error "Сообщение не найдено"
// @Security BearerAuth
// @Router /messenger/channels/{channelId}/messages/{messageId}/promote-to-issue [post]
func (h *Handler) PromoteToIssue(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req PromoteToIssueRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.PromoteToIssue(c.Context(), channelID, messageID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListThreadReplies — Список ответов в треде
// @Summary Список ответов в треде
// @Description Возвращает ответы на указанное сообщение (тред) с пагинацией.
// @Tags Мессенджер
// @Produce json
// @Param channelId path string true "ID канала"
// @Param messageId path string true "ID родительского сообщения"
// @Param before query string false "Курсор — дата RFC3339"
// @Param limit query int false "Количество (по умолчанию 50)"
// @Success 200 {array} MessageResponse "Ответы в треде"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/channels/{channelId}/messages/{messageId}/replies [get]
func (h *Handler) ListThreadReplies(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	messageID := parseMessageID(c)
	if messageID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "message ID обязателен")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var before *time.Time
	beforeStr := c.Query("before")
	if beforeStr != "" {
		t, err := time.Parse(time.RFC3339, beforeStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "некорректный формат before (ожидается RFC3339)")
		}
		before = &t
	}

	limit := queryInt(c, "limit", 50)

	resp, err := h.service.ListThreadReplies(c.Context(), channelID, messageID, userID, before, limit)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
