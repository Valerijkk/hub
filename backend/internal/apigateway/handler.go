// Package apigateway управляет API-ключами и вебхуками для внешних интеграций.
package apigateway

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
)

// Handler предоставляет HTTP-обработчики для модуля API Gateway (ключи + вебхуки).
type Handler struct {
	apiKeySvc  *APIKeyService
	webhookSvc *WebhookService
}

// NewHandler создаёт новый обработчик API Gateway.
func NewHandler(apiKeySvc *APIKeyService, webhookSvc *WebhookService) *Handler {
	return &Handler{apiKeySvc: apiKeySvc, webhookSvc: webhookSvc}
}

// RegisterRoutes регистрирует маршруты API ключей и вебхуков.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	// API Keys (пользовательские)
	keys := api.Group("/user/api-keys", authMw)
	keys.Post("/", h.CreateAPIKey)
	keys.Get("/", h.ListAPIKeys)
	keys.Delete("/:id", h.RevokeAPIKey)

	// Webhooks (проектные)
	webhooks := api.Group("/projects/:projectId/webhooks", authMw)
	webhooks.Post("/", h.CreateEndpoint)
	webhooks.Get("/", h.ListEndpoints)

	// Webhooks (по ID)
	wh := api.Group("/webhooks", authMw)
	wh.Put("/:id", h.UpdateEndpoint)
	wh.Delete("/:id", h.DeleteEndpoint)
	wh.Get("/:id/deliveries", h.ListDeliveries)
}

// --- API Key handlers ---

// CreateAPIKey godoc
// @Summary Создать API-ключ
// @Description Генерирует новый API-ключ для текущего пользователя.
// @Tags API Ключи
// @Accept json
// @Produce json
// @Param body body CreateAPIKeyRequest true "Название и опциональная ��ата истечения"
// @Success 201 {object} APIKeyResponse "Ключ созд��н (токен по��азывается единожды)"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/user/api-keys [post]
func (h *Handler) CreateAPIKey(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req CreateAPIKeyRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name обязателен")
	}

	resp, err := h.apiKeySvc.CreateAPIKey(c.Context(), userID, req.Name, req.ExpiresAt)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListAPIKeys godoc
// @Summary Список API-ключей
// @Description Возвращает все API-ключи текущего пользователя.
// @Tags API Ключи
// @Produce json
// @Success 200 {array} APIKeyResponse "Список ключей"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/user/api-keys [get]
func (h *Handler) ListAPIKeys(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	keys, err := h.apiKeySvc.ListAPIKeys(c.Context(), userID)
	if err != nil {
		return err
	}

	return c.JSON(keys)
}

// RevokeAPIKey godoc
// @Summary Отозвать API-ключ
// @Description Деактивирует API-ключ пользователя.
// @Tags API Ключи
// @Param id path string true "ID ключа"
// @Success 204 "Ключ отозван"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/user/api-keys/{id} [delete]
func (h *Handler) RevokeAPIKey(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	keyID, err := parseID(c, "id")
	if err != nil {
		return err
	}

	if err := h.apiKeySvc.RevokeAPIKey(c.Context(), userID, keyID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// --- Webhook handlers ---

// CreateEndpoint godoc
// @Summary Создать webhook-эндпоинт
// @Description Регистрирует новый webhook-эндпоинт для проекта.
// @Tags API Ключи
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateWebhookRequest true "URL и события"
// @Success 201 {object} WebhookEndpointResponse "Эндпоинт создан"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/webhooks [post]
func (h *Handler) CreateEndpoint(c fiber.Ctx) error {
	projectID, err := parseID(c, "projectId")
	if err != nil {
		return err
	}

	var req CreateWebhookRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.webhookSvc.CreateEndpoint(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListEndpoints godoc
// @Summary Список webhook-эндпоинтов
// @Description Возвращает все webhook-эндпоинты проекта.
// @Tags API Ключи
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} WebhookEndpointResponse "Список эндпоинтов"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/webhooks [get]
func (h *Handler) ListEndpoints(c fiber.Ctx) error {
	projectID, err := parseID(c, "projectId")
	if err != nil {
		return err
	}

	eps, err := h.webhookSvc.ListEndpoints(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(eps)
}

// UpdateEndpoint godoc
// @Summary Обновить webhook-эндпоинт
// @Description Обновляет URL, события или статус webhook-эндпоинта.
// @Tags API Ключи
// @Accept json
// @Produce json
// @Param id path string true "ID эндпоинта"
// @Param body body UpdateWebhookRequest true "Обновляемые данные"
// @Success 200 {object} WebhookEndpointResponse "Эндпоинт обновлён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/webhooks/{id} [put]
func (h *Handler) UpdateEndpoint(c fiber.Ctx) error {
	epID, err := parseID(c, "id")
	if err != nil {
		return err
	}

	var req UpdateWebhookRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.webhookSvc.UpdateEndpoint(c.Context(), epID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteEndpoint godoc
// @Summary Удалить webhook-эндпоинт
// @Description Удаляет webhook-эндпоинт.
// @Tags API Ключи
// @Param id path string true "ID эндпоинта"
// @Success 204 "Эндпоинт удалён"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/webhooks/{id} [delete]
func (h *Handler) DeleteEndpoint(c fiber.Ctx) error {
	epID, err := parseID(c, "id")
	if err != nil {
		return err
	}

	if err := h.webhookSvc.DeleteEndpoint(c.Context(), epID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListDeliveries godoc
// @Summary Список доставок webhook
// @Description Возвращает историю доставок для указанного webhook-эндпоинта.
// @Tags API Ключи
// @Produce json
// @Param id path string true "ID эндпоинта"
// @Success 200 {array} WebhookDeliveryResponse "История доставок"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/webhooks/{id}/deliveries [get]
func (h *Handler) ListDeliveries(c fiber.Ctx) error {
	epID, err := parseID(c, "id")
	if err != nil {
		return err
	}

	deliveries, err := h.webhookSvc.ListDeliveries(c.Context(), epID)
	if err != nil {
		return err
	}

	return c.JSON(deliveries)
}

// --- Helpers ---

// getUserID извлекает user ID из Fiber Locals.
func getUserID(c fiber.Ctx) (uuid.UUID, error) {
	idStr, ok := c.Locals("userID").(string)
	if !ok || idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "пользователь не аутентифицирован")
	}
	return uuid.Parse(idStr)
}

// parseID извлекает UUID из параметра пути.
func parseID(c fiber.Ctx, name string) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params(name))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, name+" обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "некорректный "+name)
	}
	return id, nil
}
