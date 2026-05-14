package messenger

import (
	"time"

	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Message handlers
// ---------------------------------------------------------------------------

// SendMessage godoc
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
// @Router /api/messenger/channels/{channelId}/messages [post]
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

// ListMessages godoc
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
// @Router /api/messenger/channels/{channelId}/messages [get]
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

// EditMessage godoc
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
// @Router /api/messenger/channels/{channelId}/messages/{messageId} [put]
func (h *Handler) EditMessage(c fiber.Ctx) error {
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

	resp, err := h.service.EditMessage(c.Context(), messageID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteMessage godoc
// @Summary Удалить сообщение
// @Description Удаляет сообщение из канала. Доступно автору или администратору канала.
// @Tags Мессенджер
// @Param channelId path string true "ID канала"
// @Param messageId path string true "ID сообщения"
// @Success 204 "Сообщение удалено"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/messages/{messageId} [delete]
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

// ForwardMessage godoc
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
// @Router /api/messenger/channels/{channelId}/messages/{messageId}/forward [post]
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

// ListThreadReplies godoc
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
// @Router /api/messenger/channels/{channelId}/messages/{messageId}/replies [get]
func (h *Handler) ListThreadReplies(c fiber.Ctx) error {
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

	resp, err := h.service.ListThreadReplies(c.Context(), messageID, userID, before, limit)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
