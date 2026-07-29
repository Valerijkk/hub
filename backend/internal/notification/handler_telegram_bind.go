package notification

import (
	"errors"

	"github.com/gofiber/fiber/v3"
)

// TelegramBind godoc
// @Summary Создать токен привязки Telegram
// @Description Генерирует bind_token и deep link для привязки Telegram к уведомлениям (NOTIF-02)
// @Tags Уведомления
// @Produce json
// @Security BearerAuth
// @Success 200 {object} TelegramBindResponse "Bind token"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 503 {object} fiber.Error "Telegram bind не настроен"
// @Router /notifications/telegram/bind [post]
func (h *Handler) TelegramBind(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	resp, err := h.telegramBind.CreateBindToken(c.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrTelegramBindUnavailable) {
			return fiber.NewError(fiber.StatusServiceUnavailable, "telegram bind not configured")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to create bind token")
	}
	return c.JSON(resp)
}

// TelegramConfirm godoc
// @Summary Подтвердить привязку Telegram по коду
// @Description Подтверждает link code из бота (NOTIF-02)
// @Tags Уведомления
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param body body TelegramConfirmRequest true "6-значный код"
// @Success 200 {object} map[string]string "telegram linked"
// @Failure 400 {object} fiber.Error "Неверный или просроченный код"
// @Failure 409 {object} fiber.Error "Telegram уже привязан к другому аккаунту"
// @Failure 429 {object} fiber.Error "Слишком много попыток confirm"
// @Router /notifications/telegram/confirm [post]
func (h *Handler) TelegramConfirm(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	var req TelegramConfirmRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}
	if err := h.telegramBind.ConfirmLinkCode(c.Context(), userID, req.Code); err != nil {
		if errors.Is(err, ErrTelegramInvalidCode) {
			return fiber.NewError(fiber.StatusBadRequest, "invalid or expired code")
		}
		if errors.Is(err, ErrTelegramConfirmLimited) {
			return fiber.NewError(fiber.StatusTooManyRequests, "too many verification attempts, try again later")
		}
		if errors.Is(err, ErrTelegramIDTaken) {
			return fiber.NewError(fiber.StatusConflict, "telegram already linked to another account")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to confirm link")
	}
	return c.JSON(fiber.Map{"message": "telegram linked"})
}

// TelegramUnlink godoc
// @Summary Отвязать Telegram
// @Description Удаляет привязку Telegram для уведомлений (NOTIF-02)
// @Tags Уведомления
// @Security BearerAuth
// @Success 204 "No Content"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Router /notifications/telegram [delete]
func (h *Handler) TelegramUnlink(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	if err := h.telegramBind.Unlink(c.Context(), userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to unlink telegram")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

// TelegramTest godoc
// @Summary Отправить тестовое Telegram-уведомление
// @Description Отправляет тестовое сообщение в привязанный Telegram (NOTIF-02)
// @Tags Уведомления
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]string "test sent"
// @Failure 400 {object} fiber.Error "Telegram не привязан"
// @Router /notifications/telegram/test [post]
func (h *Handler) TelegramTest(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	if err := h.telegramBind.SendTest(c.Context(), userID); err != nil {
		if errors.Is(err, ErrTelegramNotLinked) {
			return fiber.NewError(fiber.StatusBadRequest, "telegram not linked")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "failed to send test")
	}
	return c.JSON(fiber.Map{"message": "test sent"})
}

// TelegramStatus godoc
// @Summary Статус привязки Telegram
// @Description Возвращает linked/telegram_id для текущего пользователя (NOTIF-02)
// @Tags Уведомления
// @Produce json
// @Security BearerAuth
// @Success 200 {object} TelegramStatusResponse "Статус привязки"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Router /notifications/telegram/status [get]
func (h *Handler) TelegramStatus(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}
	resp, err := h.telegramBind.GetStatus(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to get status")
	}
	return c.JSON(resp)
}
