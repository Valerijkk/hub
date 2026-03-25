package gitlab

import (
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики для модуля GitLab-интеграции.
type Handler struct {
	service *GitLabService
}

// NewHandler создаёт новый обработчик GitLab-интеграции.
func NewHandler(service *GitLabService) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes регистрирует все маршруты модуля GitLab.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	// Webhook (без auth, валидация по X-Gitlab-Token)
	api.Post("/gitlab/webhook", h.HandleWebhook)

	// Интеграции (auth + permissions)
	proj := api.Group("/projects/:projectId", authMw)
	proj.Post("/gitlab/integrations", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.CreateIntegration)
	proj.Get("/gitlab/integrations", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListIntegrations)
	proj.Delete("/gitlab/integrations/:integrationId", rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc), h.DeleteIntegration)

	// Пайплайны
	proj.Get("/gitlab/pipelines", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListPipelines)

	// Создание ветки из задачи
	api.Post("/issues/:issueKey/gitlab/create-branch", authMw, h.CreateBranch)
}

// HandleWebhook обрабатывает POST /api/gitlab/webhook.
// Валидирует X-Gitlab-Token заголовок, сравнивая с секретами
// всех зарегистрированных GitLab-интеграций.
func (h *Handler) HandleWebhook(c fiber.Ctx) error {
	eventType := utils.CopyString(c.Get("X-Gitlab-Event"))
	webhookToken := utils.CopyString(c.Get("X-Gitlab-Token"))

	body, err := io.ReadAll(c.Request().BodyStream())
	if err != nil {
		body = c.Body()
	}
	if len(body) == 0 {
		body = c.Body()
	}

	// Валидируем X-Gitlab-Token по секретам всех интеграций.
	if err := h.validateWebhookToken(c, webhookToken); err != nil {
		return err
	}

	if err := h.service.HandleWebhook(c.Context(), eventType, body); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// validateWebhookToken проверяет X-Gitlab-Token webhook-запроса.
// GitLab отправляет токен в plain text (не HMAC), поэтому просто
// сравниваем с зарегистрированными секретами.
func (h *Handler) validateWebhookToken(c fiber.Ctx, token string) error {
	if token == "" {
		return fiber.NewError(fiber.StatusForbidden, "отсутствует заголовок X-Gitlab-Token")
	}

	secrets, err := h.service.repo.ListAllWebhookSecrets(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка получения секретов интеграций")
	}

	for _, secret := range secrets {
		if token == secret {
			return nil
		}
	}

	return fiber.NewError(fiber.StatusForbidden, "невалидный X-Gitlab-Token")
}

// CreateIntegration обрабатывает POST /api/projects/:projectId/gitlab/integrations.
func (h *Handler) CreateIntegration(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateIntegrationRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	if req.ProjectPath == "" {
		return fiber.NewError(fiber.StatusBadRequest, "project_path обязателен")
	}
	if req.AccessToken == "" {
		return fiber.NewError(fiber.StatusBadRequest, "access_token обязателен")
	}

	resp, err := h.service.CreateIntegration(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListIntegrations обрабатывает GET /api/projects/:projectId/gitlab/integrations.
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

// DeleteIntegration обрабатывает DELETE /api/projects/:projectId/gitlab/integrations/:integrationId.
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

// ListPipelines обрабатывает GET /api/projects/:projectId/gitlab/pipelines.
// Возвращает последние пайплайны GitLab CI/CD для всех интеграций проекта.
func (h *Handler) ListPipelines(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	pipelines, err := h.service.ListPipelines(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(pipelines)
}

// CreateBranch обрабатывает POST /api/issues/:issueKey/gitlab/create-branch.
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
