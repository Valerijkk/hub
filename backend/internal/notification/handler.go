package notification

import (
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
)

// Handler предоставляет HTTP-обработчики для модуля уведомлений.
type Handler struct {
	service *NotificationService
}

// NewHandler создаёт новый обработчик уведомлений.
func NewHandler(service *NotificationService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует маршруты уведомлений под /api/notifications.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	notifs := api.Group("/notifications", authMw)

	notifs.Get("/", h.List)
	notifs.Get("/unread-count", h.UnreadCount)
	notifs.Patch("/:id/read", h.MarkRead)
	notifs.Patch("/read-all", h.MarkAllRead)
	notifs.Delete("/:id", h.Delete)
	notifs.Get("/preferences", h.GetPreferences)
	notifs.Put("/preferences", h.UpdatePreference)

	// Push-уведомления: подписка, отписка, VAPID-ключ.
	notifs.Get("/push/vapid-key", h.GetVAPIDKey)
	notifs.Post("/push/subscribe", h.PushSubscribe)
	notifs.Delete("/push/unsubscribe", h.PushUnsubscribe)
}

// List godoc
// @Summary Список уведомлений
// @Description Возвращает пагинированный список уведомлений текущего пользователя.
// @Tags Уведомления
// @Produce json
// @Param page query int false "Страница (по умолчанию 1)"
// @Param per_page query int false "Кол-во на странице (по умолчанию 20, макс 100)"
// @Param is_read query bool false "Фильтр по статусу прочтения"
// @Success 200 {object} NotificationListResponse "Список уведомлений"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/notifications [get]
func (h *Handler) List(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	page, _ := strconv.Atoi(c.Query("page", "1"))
	perPage, _ := strconv.Atoi(c.Query("per_page", "20"))
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	var isRead *bool
	if v := c.Query("is_read"); v != "" {
		b, err := strconv.ParseBool(v)
		if err == nil {
			isRead = &b
		}
	}

	result, err := h.service.ListNotifications(c.Context(), userID, page, perPage, isRead)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка получения уведомлений")
	}

	return c.JSON(result)
}

// UnreadCount godoc
// @Summary Счётчик непрочитанных
// @Description Возвращает количество непрочитанных уведомлений пользователя.
// @Tags Уведомления
// @Produce json
// @Success 200 {object} UnreadCountResponse "count: число"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/notifications/unread-count [get]
func (h *Handler) UnreadCount(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	count, err := h.service.GetUnreadCount(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка подсчёта уведомлений")
	}

	return c.JSON(UnreadCountResponse{Count: count})
}

// MarkRead godoc
// @Summary Пометить прочитанным
// @Description Помечает одно уведомление как прочитанное.
// @Tags Уведомления
// @Param id path string true "ID уведомления"
// @Success 204 "Помечено прочитанным"
// @Failure 400 {object} fiber.Error "Невалидный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/notifications/{id}/read [patch]
func (h *Handler) MarkRead(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	notifID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "невалидный ID уведомления")
	}

	if err := h.service.MarkRead(c.Context(), notifID, userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка отметки прочитанным")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// MarkAllRead godoc
// @Summary Пометить все прочитанными
// @Description Помечает все уведомления текущего пользователя как прочитанные.
// @Tags Уведомления
// @Success 204 "Все помечены прочитанными"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/notifications/read-all [patch]
func (h *Handler) MarkAllRead(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	if err := h.service.MarkAllRead(c.Context(), userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка отметки всех прочитанными")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// Delete godoc
// @Summary Удалить уведомление
// @Description Удаляет одно уведомление пользователя.
// @Tags Уведомления
// @Param id path string true "ID уведомления"
// @Success 204 "Уведомление удалено"
// @Failure 400 {object} fiber.Error "Невалидный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/notifications/{id} [delete]
func (h *Handler) Delete(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	notifID, err := uuid.Parse(c.Params("id"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "невалидный ID уведомления")
	}

	if err := h.service.DeleteNotification(c.Context(), notifID, userID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка удаления уведомления")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// GetPreferences godoc
// @Summary Настройки уведомлений
// @Description Возвращает настройки уведомлений текущего пользователя по типам и каналам.
// @Tags Уведомления
// @Produce json
// @Success 200 {array} PreferenceResponse "Настройки уведомлений"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/notifications/preferences [get]
func (h *Handler) GetPreferences(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	prefs, err := h.service.GetPreferences(c.Context(), userID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка получения настроек")
	}

	return c.JSON(prefs)
}

// UpdatePreference godoc
// @Summary Обновить настройку уведомлений
// @Description Создаёт или обновляет настройку уведомлений для указанного типа и канала.
// @Tags Уведомления
// @Accept json
// @Produce json
// @Param body body PreferenceUpdateRequest true "Тип уведомления, канал, enabled"
// @Success 200 {object} PreferenceResponse "Настройка обновлена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/notifications/preferences [put]
func (h *Handler) UpdatePreference(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	var req PreferenceUpdateRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "невалидное тело запроса")
	}

	if req.NotificationType == "" || req.Channel == "" {
		return fiber.NewError(fiber.StatusBadRequest, "notification_type и channel обязательны")
	}

	pref, err := h.service.UpdatePreference(c.Context(), userID, req)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка обновления настройки")
	}

	return c.JSON(pref)
}

// GetVAPIDKey godoc
// @Summary Получить VAPID-ключ
// @Description Возвращает публичный VAPID-ключ для подписки на Web Push уведомления.
// @Tags Уведомления
// @Produce json
// @Success 200 {object} map[string]string "vapid_public_key"
// @Security BearerAuth
// @Router /api/notifications/push/vapid-key [get]
func (h *Handler) GetVAPIDKey(c fiber.Ctx) error {
	push := h.service.PushSender()
	key := ""
	if push != nil {
		key = push.VAPIDPublicKey()
	}
	return c.JSON(fiber.Map{"vapid_public_key": key})
}

// PushSubscribeRequest — запрос на подписку push-уведомлений.
type PushSubscribeRequest struct {
	Endpoint  string `json:"endpoint"`
	P256dhKey string `json:"p256dh_key"`
	AuthKey   string `json:"auth_key"`
}

// PushSubscribe godoc
// @Summary Подписаться на push-уведомления
// @Description Сохраняет подписку Web Push для текущего пользователя.
// @Tags Уведомления
// @Accept json
// @Param body body PushSubscribeRequest true "Данные подписки (endpoint, p256dh_key, auth_key)"
// @Success 201 "Подписка сохранена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 503 {object} fiber.Error "Push не настроен"
// @Security BearerAuth
// @Router /api/notifications/push/subscribe [post]
func (h *Handler) PushSubscribe(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	var req PushSubscribeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "невалидное тело запроса")
	}

	if req.Endpoint == "" || req.P256dhKey == "" || req.AuthKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "endpoint, p256dh_key и auth_key обязательны")
	}

	push := h.service.PushSender()
	if push == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "push-уведомления не настроены")
	}

	userAgent := c.Get("User-Agent")
	if err := push.SaveSubscription(c.Context(), userID, req.Endpoint, req.P256dhKey, req.AuthKey, userAgent); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка сохранения подписки")
	}

	return c.SendStatus(fiber.StatusCreated)
}

// PushUnsubscribeRequest — запрос на отписку от push-уведомлений.
type PushUnsubscribeRequest struct {
	Endpoint string `json:"endpoint"`
}

// PushUnsubscribe godoc
// @Summary Отписаться от push-уведомлений
// @Description Удаляет подписку Web Push по endpoint.
// @Tags Уведомления
// @Accept json
// @Param body body PushUnsubscribeRequest true "endpoint для удаления"
// @Success 204 "Подписка удалена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 503 {object} fiber.Error "Push не настроен"
// @Security BearerAuth
// @Router /api/notifications/push/unsubscribe [delete]
func (h *Handler) PushUnsubscribe(c fiber.Ctx) error {
	_, err := getUserID(c)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	var req PushUnsubscribeRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "невалидное тело запроса")
	}

	if req.Endpoint == "" {
		return fiber.NewError(fiber.StatusBadRequest, "endpoint обязателен")
	}

	push := h.service.PushSender()
	if push == nil {
		return fiber.NewError(fiber.StatusServiceUnavailable, "push-уведомления не настроены")
	}

	if err := push.RemoveSubscription(c.Context(), req.Endpoint); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка удаления подписки")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// getUserID извлекает ID пользователя из контекста Fiber.
func getUserID(c fiber.Ctx) (uuid.UUID, error) {
	raw := c.Locals("userID")
	if raw == nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}

	switch v := raw.(type) {
	case string:
		return uuid.Parse(v)
	case uuid.UUID:
		return v, nil
	default:
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "не авторизован")
	}
}
