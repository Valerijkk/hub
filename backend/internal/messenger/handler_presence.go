package messenger

import (
	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// Обработчик набора текста
// ---------------------------------------------------------------------------

// SendTyping — Отправить индикатор набора
// @Summary Отправить индикатор набора
// @Description Публикует событие "пользователь печатает" для участников канала.
// @Tags Мессенджер
// @Param channelId path string true "ID канала"
// @Success 204 "Событие отправлено"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/channels/{channelId}/typing [post]
func (h *Handler) SendTyping(c fiber.Ctx) error {
	channelID, err := parseChannelID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.PublishTyping(c.Context(), channelID.String(), userID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Обработчики присутствия
// ---------------------------------------------------------------------------

// GetPresence — Получить статус присутствия
// @Summary Получить статус присутствия
// @Description Возвращает статус онлайн/оффлайн для указанных пользователей.
// @Tags Мессенджер
// @Produce json
// @Param user_ids query string true "Список ID пользователей через запятую"
// @Success 200 {object} map[string]string "Статусы пользователей"
// @Failure 400 {object} fiber.Error "Параметр user_ids обязателен"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/presence [get]
func (h *Handler) GetPresence(c fiber.Ctx) error {
	userIDsStr := c.Query("user_ids")
	if userIDsStr == "" {
		return fiber.NewError(fiber.StatusBadRequest, "параметр user_ids обязателен")
	}

	userIDs := splitComma(userIDsStr)
	statuses, err := h.service.GetPresence(c.Context(), userIDs)
	if err != nil {
		return err
	}

	return c.JSON(statuses)
}

// SetCustomStatus — Установить пользовательский статус
// @Summary Установить пользовательский статус
// @Description Устанавливает текстовый статус пользователя с опциональной длительностью.
// @Tags Мессенджер
// @Accept json
// @Param body body object true "text и duration статуса"
// @Success 200 "Статус обновлён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/presence/status [put]
func (h *Handler) SetCustomStatus(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req struct {
		Text     string `json:"text"`
		Duration string `json:"duration"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	if err := h.service.SetCustomStatus(c.Context(), userID.String(), req.Text, req.Duration); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// Heartbeat — Heartbeat присутствия
// @Summary Heartbeat присутствия
// @Description Обновляет таймстамп активности пользователя для отслеживания онлайн-статуса.
// @Tags Мессенджер
// @Success 200 "Heartbeat принят"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/presence/heartbeat [post]
func (h *Handler) Heartbeat(c fiber.Ctx) error {
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	if err := h.service.Heartbeat(c.Context(), userID.String()); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}
