package messenger

import (
	"github.com/gofiber/fiber/v3"
)

// ---------------------------------------------------------------------------
// Search handlers
// ---------------------------------------------------------------------------

// SearchMessages godoc
// @Summary Поиск сообщений
// @Description Полнотекстовый поиск по сообщениям мессенджера с опциональной фильтрацией по каналу.
// @Tags Мессенджер
// @Produce json
// @Param query query string true "Поисковый запрос"
// @Param channel_id query string false "ID канала для фильтрации"
// @Param limit query int false "Лимит результатов (по умолчанию 20)"
// @Success 200 {array} MessageResponse "Результаты поиска"
// @Failure 400 {object} fiber.Error "Параметр query обязателен"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/messenger/search [get]
func (h *Handler) SearchMessages(c fiber.Ctx) error {
	query := c.Query("query")
	if query == "" {
		return fiber.NewError(fiber.StatusBadRequest, "параметр query обязателен")
	}
	channelID := c.Query("channel_id")
	limit := queryInt(c, "limit", 20)

	resp, err := h.service.SearchMessages(c.Context(), query, channelID, limit)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}
