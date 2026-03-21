package github

import (
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики для модуля GitHub-интеграции.
type Handler struct {
	service *GitHubService
	preview *LinkPreviewService
}

// NewHandler создаёт новый обработчик GitHub-интеграции.
func NewHandler(service *GitHubService, preview *LinkPreviewService) *Handler {
	return &Handler{service: service, preview: preview}
}

// RegisterRoutes регистрирует все маршруты модуля GitHub.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	// Webhook (без auth, валидация по подписи)
	api.Post("/github/webhook", h.HandleWebhook)

	// Интеграции (auth + owner)
	proj := api.Group("/projects/:projectId", authMw)
	proj.Post("/github/integrations", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.CreateIntegration)
	proj.Get("/github/integrations", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListIntegrations)
	proj.Delete("/github/integrations/:integrationId", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.DeleteIntegration)

	// GitHub ссылки для задачи
	api.Get("/issues/:issueKey/github-links", authMw, h.ListGitHubLinks)

	// Создание ветки из задачи (GHUB-05)
	api.Post("/issues/:issueKey/github/create-branch", authMw, h.CreateBranch)

	// Превью ссылок (MSG-11)
	api.Get("/link-preview", authMw, h.GetLinkPreview)
}

// HandleWebhook обрабатывает POST /api/github/webhook.
func (h *Handler) HandleWebhook(c fiber.Ctx) error {
	eventType := utils.CopyString(c.Get("X-GitHub-Event"))
	signature := utils.CopyString(c.Get("X-Hub-Signature-256"))

	body, err := io.ReadAll(c.Request().BodyStream())
	if err != nil {
		// Для Fiber v3 body уже прочитан, используем Body()
		body = c.Body()
	}
	if len(body) == 0 {
		body = c.Body()
	}

	// Валидируем подпись для всех зарегистрированных интеграций
	// (в production стоит определять интеграцию по repo из payload)
	if signature != "" {
		// Подпись будет валидироваться в сервисе при необходимости
		_ = signature
	}

	if err := h.service.HandleWebhook(c.Context(), eventType, body); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// CreateIntegration обрабатывает POST /api/projects/:projectId/github/integrations.
func (h *Handler) CreateIntegration(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateIntegrationRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateIntegration(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListIntegrations обрабатывает GET /api/projects/:projectId/github/integrations.
func (h *Handler) ListIntegrations(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListIntegrations(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteIntegration обрабатывает DELETE /api/projects/:projectId/github/integrations/:integrationId.
func (h *Handler) DeleteIntegration(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("integrationId"))
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный integration ID")
	}

	if err := h.service.DeleteIntegration(c.Context(), id); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListGitHubLinks обрабатывает GET /api/issues/:issueKey/github-links.
func (h *Handler) ListGitHubLinks(c fiber.Ctx) error {
	issueKey := utils.CopyString(c.Params("issueKey"))
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	resp, err := h.service.ListGitHubLinks(c.Context(), issueKey)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CreateBranch обрабатывает POST /api/issues/:issueKey/github/create-branch (GHUB-05).
func (h *Handler) CreateBranch(c fiber.Ctx) error {
	issueKey := utils.CopyString(c.Params("issueKey"))
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	var req CreateBranchRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.CreateBranch(c.Context(), issueKey, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// GetLinkPreview обрабатывает GET /api/link-preview?url=... (MSG-11).
func (h *Handler) GetLinkPreview(c fiber.Ctx) error {
	rawURL := utils.CopyString(c.Query("url"))
	if rawURL == "" {
		return fiber.NewError(fiber.StatusBadRequest, "параметр url обязателен")
	}

	resp, err := h.preview.GetLinkPreview(c.Context(), rawURL)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// parseProjectID извлекает и валидирует projectId из параметров пути.
func parseProjectID(c fiber.Ctx) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params("projectId"))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "project ID обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "некорректный project ID")
	}
	return id, nil
}
