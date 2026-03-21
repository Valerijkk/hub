package apigateway

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
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

// CreateAPIKey обрабатывает POST /api/user/api-keys.
func (h *Handler) CreateAPIKey(c fiber.Ctx) error {
	userID, err := getUserID(c)
	if err != nil {
		return err
	}

	var req CreateAPIKeyRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
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

// ListAPIKeys обрабатывает GET /api/user/api-keys.
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

// RevokeAPIKey обрабатывает DELETE /api/user/api-keys/:id.
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

// CreateEndpoint обрабатывает POST /api/projects/:projectId/webhooks.
func (h *Handler) CreateEndpoint(c fiber.Ctx) error {
	projectID, err := parseID(c, "projectId")
	if err != nil {
		return err
	}

	var req CreateWebhookRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.webhookSvc.CreateEndpoint(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListEndpoints обрабатывает GET /api/projects/:projectId/webhooks.
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

// UpdateEndpoint обрабатывает PUT /api/webhooks/:id.
func (h *Handler) UpdateEndpoint(c fiber.Ctx) error {
	epID, err := parseID(c, "id")
	if err != nil {
		return err
	}

	var req UpdateWebhookRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.webhookSvc.UpdateEndpoint(c.Context(), epID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteEndpoint обрабатывает DELETE /api/webhooks/:id.
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

// ListDeliveries обрабатывает GET /api/webhooks/:id/deliveries.
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
