package github

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"strconv"
	"strings"

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

	// CI/CD Pipeline Status (workflow runs)
	proj.Get("/github/workflows", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.ListWorkflowRuns)
	proj.Get("/github/workflows/:runId", rbac.RequirePermission(rbac.PermProjectView, rbacSvc), h.GetWorkflowRunDetail)

	// GitHub ссылки для задачи
	api.Get("/issues/:issueKey/github-links", authMw, h.ListGitHubLinks)

	// Создание ветки из задачи (GHUB-05)
	api.Post("/issues/:issueKey/github/create-branch", authMw, h.CreateBranch)

	// Превью ссылок (MSG-11)
	api.Get("/link-preview", authMw, h.GetLinkPreview)
}

// HandleWebhook обрабатывает POST /api/github/webhook.
// Валидирует HMAC-SHA256 подпись из заголовка X-Hub-Signature-256,
// сравнивая с секретами всех зарегистрированных интеграций.
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

	// Валидируем HMAC-SHA256 подпись по секретам всех зарегистрированных интеграций.
	if err := h.validateWebhookSignature(c, body, signature); err != nil {
		return err
	}

	if err := h.service.HandleWebhook(c.Context(), eventType, body); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusOK)
}

// validateWebhookSignature проверяет HMAC-SHA256 подпись webhook-запроса.
// Перебирает все зарегистрированные секреты интеграций и возвращает 403,
// если ни один не совпал. Формат подписи: "sha256=<hex>".
func (h *Handler) validateWebhookSignature(c fiber.Ctx, body []byte, signature string) error {
	if signature == "" {
		return fiber.NewError(fiber.StatusForbidden, "отсутствует заголовок X-Hub-Signature-256")
	}

	// Извлекаем hex-часть из "sha256=<hex>".
	const prefix = "sha256="
	if !strings.HasPrefix(signature, prefix) {
		return fiber.NewError(fiber.StatusForbidden, "некорректный формат подписи")
	}
	receivedMAC, err := hex.DecodeString(signature[len(prefix):])
	if err != nil {
		return fiber.NewError(fiber.StatusForbidden, "некорректная hex-подпись")
	}

	secrets, err := h.service.repo.ListAllWebhookSecrets(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка получения секретов интеграций")
	}

	// Проверяем подпись по каждому секрету — хотя бы один должен совпасть.
	for _, secret := range secrets {
		mac := hmac.New(sha256.New, []byte(secret))
		mac.Write(body)
		expectedMAC := mac.Sum(nil)
		if hmac.Equal(receivedMAC, expectedMAC) {
			return nil
		}
	}

	return fiber.NewError(fiber.StatusForbidden, "невалидная подпись webhook")
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

// ListWorkflowRuns обрабатывает GET /api/projects/:projectId/github/workflows.
// Возвращает последние запуски CI/CD пайплайнов для всех интеграций проекта.
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

// GetWorkflowRunDetail обрабатывает GET /api/projects/:projectId/github/workflows/:runId.
// Возвращает детали запуска workflow, включая список джобов.
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
