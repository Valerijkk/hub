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

// List возвращает пагинированный список уведомлений пользователя.
// Query params: page, per_page, is_read (optional).
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

// UnreadCount возвращает количество непрочитанных уведомлений.
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

// MarkRead помечает одно уведомление как прочитанное.
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

// MarkAllRead помечает все уведомления пользователя как прочитанные.
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

// Delete удаляет уведомление пользователя.
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

// GetPreferences возвращает настройки уведомлений пользователя.
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

// UpdatePreference создаёт или обновляет настройку уведомлений.
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

// GetVAPIDKey возвращает публичный VAPID-ключ для подписки на push-уведомления.
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

// PushSubscribe сохраняет подписку на push-уведомления для текущего пользователя.
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

// PushUnsubscribe удаляет подписку на push-уведомления.
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
