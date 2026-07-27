package github

import (
	"io"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/rbac"
)

// maxWebhookBodyBytes — лимит тела GitHub webhook (memory DoS protection).
const maxWebhookBodyBytes = 10 * 1024 * 1024

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
	proj.Post("/github/integrations", h.CreateIntegration, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))
	proj.Get("/github/integrations", h.ListIntegrations, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	proj.Delete("/github/integrations/:integrationId", h.DeleteIntegration, rbac.RequirePermission(rbac.PermProjectManageWorkflow, rbacSvc))

	// Статус CI/CD pipeline (workflow runs)
	proj.Get("/github/workflows", h.ListWorkflowRuns, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	proj.Get("/github/workflows/:runId", h.GetWorkflowRunDetail, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Создание ветки из задачи (GHUB-05 / A5-006): RBAC на project + service scope issue↔integration.
	proj.Post("/issues/:issueKey/github/create-branch", h.CreateBranch, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))

	// GitHub ссылки для задачи
	api.Get("/issues/:issueKey/github-links", authMw, h.ListGitHubLinks)

	// Превью ссылок (MSG-11)
	api.Get("/link-preview", authMw, h.GetLinkPreview)
}

// HandleWebhook godoc
// @Summary GitHub Webhook
// @Description Обрабатывает входящие события от GitHub. Валидирует HMAC-SHA256 подпись.
// @Tags GitHub
// @Accept json
// @Param body body object true "Payload от GitHub"
// @Success 200 "Событие обработано"
// @Failure 403 {object} fiber.Error "Невалидная подпись"
// @Router /api/github/webhook [post]
func (h *Handler) HandleWebhook(c fiber.Ctx) error {
	eventType := utils.CopyString(c.Get("X-GitHub-Event"))
	signature := utils.CopyString(c.Get("X-Hub-Signature-256"))

	body := c.Body()
	if len(body) == 0 {
		if stream := c.Request().BodyStream(); stream != nil {
			var err error
			body, err = io.ReadAll(io.LimitReader(stream, maxWebhookBodyBytes+1))
			if err != nil {
				return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
			}
		}
	}
	if len(body) > maxWebhookBodyBytes {
		return fiber.NewError(fiber.StatusRequestEntityTooLarge, apperr.MsgBadRequest)
	}

	// A5-008: HMAC только секретом интеграции, найденной по repo из payload.
	if err := h.service.VerifyWebhookSignature(c.Context(), body, signature); err != nil {
		return err
	}

	if err := h.service.HandleWebhook(c.Context(), eventType, body); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// CreateIntegration godoc
// @Summary Создать GitHub-интеграцию
// @Description Подключает GitHub-репозиторий к проекту.
// @Tags GitHub
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateIntegrationRequest true "Данные интеграции"
// @Success 201 {object} IntegrationResponse "Интеграция создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/github/integrations [post]
func (h *Handler) CreateIntegration(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateIntegrationRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateIntegration(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListIntegrations godoc
// @Summary Список GitHub-интеграций
// @Description Возвращает все GitHub-интеграции проекта.
// @Tags GitHub
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} IntegrationResponse "Список интеграций"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/github/integrations [get]
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
// @Summary Удалить GitHub-интеграцию
// @Description Отключает GitHub-репозиторий от проекта.
// @Tags GitHub
// @Param projectId path string true "ID проекта"
// @Param integrationId path string true "ID интеграции"
// @Success 204 "Интеграция удалена"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/github/integrations/{integrationId} [delete]
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

// ListGitHubLinks godoc
// @Summary GitHub-ссылки задачи
// @Description Возвращает коммиты, PR и ветки, связанные с задачей через GitHub.
// @Tags GitHub
// @Produce json
// @Param issueKey path string true "Ключ задачи"
// @Success 200 {array} GitHubLinkResponse "Связанные ссылки"
// @Failure 400 {object} fiber.Error "Неверный ключ"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/issues/{issueKey}/github-links [get]
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

// CreateBranch godoc
// @Summary Создать ветку из задачи
// @Description Создаёт Git-ветку в GitHub-репозитории с именем на основе ключа задачи.
// @Tags GitHub
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
// @Router /api/projects/{projectId}/issues/{issueKey}/github/create-branch [post]
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

// ListWorkflowRuns godoc
// @Summary Список запусков CI/CD
// @Description Возвращает последние запуски GitHub Actions workflows для проекта.
// @Tags GitHub
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} WorkflowRunResponse "Список запусков"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/github/workflows [get]
func (h *Handler) ListWorkflowRuns(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	runs, err := h.service.ListWorkflowRuns(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(runs)
}

// GetWorkflowRunDetail godoc
// @Summary Детали запуска workflow
// @Description Возвращает подробности запуска GitHub Actions, включая список джобов.
// @Tags GitHub
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param runId path int true "ID запуска"
// @Success 200 {object} WorkflowRunDetailResponse "Детали запуска"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/github/workflows/{runId} [get]
func (h *Handler) GetWorkflowRunDetail(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	runIDStr := utils.CopyString(c.Params("runId"))
	runID, err := strconv.ParseInt(runIDStr, 10, 64)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный run ID")
	}

	detail, err := h.service.GetWorkflowRunDetail(c.Context(), projectID, runID)
	if err != nil {
		return err
	}

	return c.JSON(detail)
}

// GetLinkPreview godoc
// @Summary Превью ссылки
// @Description Возвращает Open Graph метаданные для указанного URL (заголовок, описание, изображение).
// @Tags GitHub
// @Produce json
// @Param url query string true "URL для получения превью"
// @Success 200 {object} LinkPreviewResponse "Превью ссылки"
// @Failure 400 {object} fiber.Error "URL не указан"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/link-preview [get]
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
