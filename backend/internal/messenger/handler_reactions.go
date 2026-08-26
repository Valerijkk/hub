package messenger

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчики реакций
// ---------------------------------------------------------------------------

// ToggleReaction — Переключить реакцию
// @Summary Переключить реакцию
// @Description Добавляет или убирает реакцию (эмодзи) на сообщение.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID канала"
// @Param messageId path string true "ID сообщения"
// @Param body body ToggleReactionRequest true "Эмодзи реакции"
// @Success 200 {object} map[string]bool "added: true/false"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/channels/{channelId}/messages/{messageId}/reactions [post]
func (h *Handler) ToggleReaction(c fiber.Ctx) error {
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

	var req ToggleReactionRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	added, err := h.service.ToggleReaction(c.Context(), channelID, messageID, userID, req.Emoji)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"added": added})
}

// ---------------------------------------------------------------------------
// Обработчики закрепления
// ---------------------------------------------------------------------------

// TogglePin — Закрепить/открепить сообщение
// @Summary Закрепить/открепить сообщение
// @Description Переключает закрепление сообщения в канале.
// @Tags Мессенджер
// @Produce json
// @Param channelId path string true "ID канала"
// @Param messageId path string true "ID сообщения"
// @Success 200 {object} map[string]bool "pinned: true/false"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/channels/{channelId}/messages/{messageId}/pin [post]
func (h *Handler) TogglePin(c fiber.Ctx) error {
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

	pinned, err := h.service.TogglePin(c.Context(), channelID, messageID, userID)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{"pinned": pinned})
}

// ListPins — Список закреплённых сообщений
// @Summary Список закреплённых сообщений
// @Description Возвращает все закреплённые сообщения в указанном канале.
// @Tags Мессенджер
// @Produce json
// @Param channelId path string true "ID канала"
// @Success 200 {array} MessageResponse "Закреплённые сообщения"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/channels/{channelId}/pins [get]
func (h *Handler) ListPins(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	pins, err := h.service.ListPins(c.Context(), channelID, userID)
	if err != nil {
		return err
	}

	return c.JSON(pins)
}

// ---------------------------------------------------------------------------
// Обработчики закладок
// ---------------------------------------------------------------------------

// ListBookmarks — Список закладок
// @Summary Список закладок
// @Description Возвращает все сохранённые закладки на сообщения текущего пользователя.
// @Tags Мессенджер
// @Produce json
// @Success 200 {array} BookmarkResponse "Список закладок"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/bookmarks [get]
func (h *Handler) ListBookmarks(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListBookmarks(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// AddBookmark — Добавить закладку
// @Summary Добавить закладку
// @Description Сохраняет сообщение в закладки текущего пользователя.
// @Tags Мессенджер
// @Accept json
// @Param body body AddBookmarkRequest true "ID сообщения и канала"
// @Success 201 "Закладка добавлена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/bookmarks [post]
func (h *Handler) AddBookmark(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req AddBookmarkRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	chID, err := uuid.Parse(req.ChannelID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный channel_id")
	}

	if err := h.service.AddBookmark(c.Context(), userID, req.MessageID, chID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusCreated)
}

// RemoveBookmark — Удалить закладку
// @Summary Удалить закладку
// @Description Удаляет закладку на сообщение.
// @Tags Мессенджер
// @Param bookmarkId path string true "ID закладки"
// @Success 204 "Закладка удалена"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/bookmarks/{bookmarkId} [delete]
func (h *Handler) RemoveBookmark(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("bookmarkId"))
	if idStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "bookmark ID обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный bookmark ID")
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.RemoveBookmark(c.Context(), userID, id); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Обработчики непрочитанного
// ---------------------------------------------------------------------------

// GetUnreadCounts — Счётчики непрочитанных
// @Summary Счётчики непрочитанных
// @Description Возвращает количество непрочитанных сообщений по каналам для текущего пользователя.
// @Tags Мессенджер
// @Produce json
// @Success 200 {object} map[string]int "Счётчики непрочитанных по каналам"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/unread [get]
func (h *Handler) GetUnreadCounts(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	counts, err := h.service.GetUnreadResponse(c.Context(), userID.String())
	if err != nil {
		return err
	}

	return c.JSON(counts)
}

// MarkChannelRead — Отметить канал прочитанным
// @Summary Отметить канал прочитанным
// @Description Сбрасывает счётчик непрочитанных сообщений для указанного канала.
// @Tags Мессенджер
// @Param channelId path string true "ID канала"
// @Success 200 "Канал отмечен прочитанным"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/channels/{channelId}/read [post]
func (h *Handler) MarkChannelRead(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.MarkChannelRead(c.Context(), userID.String(), channelID.String()); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// ---------------------------------------------------------------------------
// Обработчики опросов
// ---------------------------------------------------------------------------

// parsePollID извлекает и валидирует pollId из параметров пути.
// ME-12: отклоняет uuid.Nil.
func parsePollID(c fiber.Ctx) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params("pollId"))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "poll ID обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "некорректный poll ID")
	}
	return id, nil
}

// CreatePoll — Создать опрос
// @Summary Создать опрос
// @Description Создаёт новый опрос в указанном канале мессенджера.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param channelId path string true "ID канала"
// @Param body body CreatePollRequest true "Данные опроса"
// @Success 201 {object} PollResponse "Опрос создан"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/channels/{channelId}/polls [post]
func (h *Handler) CreatePoll(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req CreatePollRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreatePoll(c.Context(), channelID, userID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// VotePoll — Проголосовать в опросе
// @Summary Проголосовать в опросе
// @Description Отправляет голос текущего пользователя за выбранный вариант.
// @Tags Мессенджер
// @Accept json
// @Produce json
// @Param pollId path string true "ID опроса"
// @Param body body VotePollRequest true "Выбранный вариант"
// @Success 200 {object} PollResponse "Голос учтён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/polls/{pollId}/vote [post]
func (h *Handler) VotePoll(c fiber.Ctx) error {
	pollID, err := parsePollID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req VotePollRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.VotePoll(c.Context(), pollID, userID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ClosePoll — Закрыть опрос
// @Summary Закрыть опрос
// @Description Завершает опрос и фиксирует результаты. Доступно только создателю.
// @Tags Мессенджер
// @Produce json
// @Param pollId path string true "ID опроса"
// @Success 200 {object} PollResponse "Опрос закрыт"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /messenger/polls/{pollId}/close [post]
func (h *Handler) ClosePoll(c fiber.Ctx) error {
	pollID, err := parsePollID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ClosePoll(c.Context(), pollID, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetPollResults — Результаты опроса
// @Summary Результаты опроса
// @Description Возвращает текущие результаты опроса, включая голоса и процентное распределение.
// @Tags Мессенджер
// @Produce json
// @Param pollId path string true "ID опроса"
// @Success 200 {object} PollResponse "Результаты опроса"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /messenger/polls/{pollId} [get]
func (h *Handler) GetPollResults(c fiber.Ctx) error {
	pollID, err := parsePollID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetPollResults(c.Context(), pollID, userID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
