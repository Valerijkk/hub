package gitlab

import (
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/rbac"
)

// maxWebhookBodyBytes — лимит тела GitLab webhook (memory DoS protection).
const maxWebhookBodyBytes = 10 * 1024 * 1024

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
	proj.Post("/gitlab/integrations", h.CreateIntegration, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))
	proj.Get("/gitlab/integrations", h.ListIntegrations, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	proj.Delete("/gitlab/integrations/:integrationId", h.DeleteIntegration, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))

	// Пайплайны
	proj.Get("/gitlab/pipelines", h.ListPipelines, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	proj.Post("/gitlab/pipelines/trigger", h.TriggerPipeline, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))

	// Создание ветки из задачи (A5-007): RBAC на project + service scope issue↔integration.
	proj.Post("/issues/:issueKey/gitlab/create-branch", h.CreateBranch, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
}

// HandleWebhook godoc
// @Summary GitLab Webhook
// @Description Обрабатывает входящие события от GitLab. Валидирует X-Gitlab-Token.
// @Tags GitLab
// @Accept json
// @Param body body object true "Payload от GitLab"
// @Success 200 "Событие обработано"
// @Failure 403 {object} fiber.Error "Невалидный токен"
// @Router /api/gitlab/webhook [post]
func (h *Handler) HandleWebhook(c fiber.Ctx) error {
	eventType := utils.CopyString(c.Get("X-Gitlab-Event"))
	webhookToken := utils.CopyString(c.Get("X-Gitlab-Token"))

	var body []byte
	if stream := c.Request().BodyStream(); stream != nil {
		var readErr error
		body, readErr = io.ReadAll(io.LimitReader(stream, maxWebhookBodyBytes+1))
		if readErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
		}
	}
	if len(body) == 0 {
		body = c.Body()
	}
	if len(body) > maxWebhookBodyBytes {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, apperr.MsgBadRequest)
	}

	// A5-009: X-Gitlab-Token только секретом интеграции, найденной по project_path.
	if err := h.service.VerifyWebhookToken(c.Context(), body, webhookToken); err != nil {
		return err
	}

	if err := h.service.HandleWebhook(c.Context(), eventType, body); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// CreateIntegration godoc
// @Summary Создать GitLab-интеграцию
// @Description Подключает GitLab-репозиторий к проекту.
// @Tags GitLab
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateIntegrationRequest true "Данные интеграции"
// @Success 201 {object} IntegrationResponse "Интеграция создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/gitlab/integrations [post]
func (h *Handler) CreateIntegration(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateIntegrationRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
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

// ListIntegrations godoc
// @Summary Список GitLab-интеграций
// @Description Возвращает все GitLab-интеграции проекта.
// @Tags GitLab
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} IntegrationResponse "Список интеграций"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/gitlab/integrations [get]
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

// DeleteIntegration godoc
// @Summary Удалить GitLab-интеграцию
// @Description Отключает GitLab-репозиторий от проекта.
// @Tags GitLab
// @Param projectId path string true "ID проекта"
// @Param integrationId path string true "ID интеграции"
// @Success 204 "Интеграция удалена"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/gitlab/integrations/{integrationId} [delete]
func (h *Handler) DeleteIntegration(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	idStr := utils.CopyString(c.Params("integrationId"))
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный integration ID")
	}

	if err := h.service.DeleteIntegration(c.Context(), projectID, id); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ListPipelines godoc
// @Summary Список пайплайнов GitLab
// @Description Возвращает последние CI/CD пайплайны для всех GitLab-интеграций проекта.
// @Tags GitLab
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} PipelineResponse "Список пайплайнов"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/gitlab/pipelines [get]
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

// TriggerPipeline godoc
// @Summary Запустить пайплайн GitLab
// @Description Запускает CI/CD пайплайн в GitLab для указанной интеграции и ветки.
// @Tags GitLab
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body TriggerPipelineRequest true "ID интеграции и ветка"
// @Success 201 {object} PipelineResponse "Пайплайн запущен"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/gitlab/pipelines/trigger [post]
func (h *Handler) TriggerPipeline(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req TriggerPipelineRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	if req.IntegrationID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "integration_id обязателен")
	}

	resp, err := h.service.TriggerPipeline(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// CreateBranch godoc
// @Summary Создать ветку из задачи (GitLab)
// @Description Создаёт Git-ветку в GitLab-репозитории с именем на основе ключа задачи.
// @Tags GitLab
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param issueKey path string true "Ключ задачи"
// @Param body body CreateBranchRequest true "Параметры ветки"
// @Success 201 {object} CreateBranchResponse "Ветка создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Недостаточно прав"
// @Security BearerAuth
// @Router /api/projects/{projectId}/issues/{issueKey}/gitlab/create-branch [post]
func (h *Handler) CreateBranch(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	issueKey := utils.CopyString(c.Params("issueKey"))
	if issueKey == "" {
		return fiber.NewError(fiber.StatusBadRequest, "issue key обязателен")
	}

	var req CreateBranchRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateBranch(c.Context(), projectID, issueKey, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ---------------------------------------------------------------------------
// Вспомогательные функции
// ---------------------------------------------------------------------------

// parseProjectID извлекает и валидирует projectId из параметров пути.
// ME-12: отклоняет uuid.Nil.
func parseProjectID(c fiber.Ctx) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params("projectId"))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "project ID обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil || id == uuid.Nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "некорректный project ID")
	}
	return id, nil
}
