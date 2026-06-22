package storage

import (
	"log/slog"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики для файловых операций: генерация URL загрузки,
// подтверждение загрузки, генерация URL скачивания и удаление файлов.
type Handler struct {
	storage *StorageService
	rbacSvc *rbac.RBACService
}

// NewHandler создаёт новый обработчик файлового хранилища на основе StorageService.
func NewHandler(storage *StorageService) *Handler {
	return &Handler{storage: storage}
}

// RegisterRoutes регистрирует маршруты файлового хранилища.
// authMw — JWT middleware аутентификации (auth.RequireAuth), валидирующий Bearer-токены.
// rbacSvc используется обработчиками для проверки прав доступа по фактическому
// projectID файла (CR-04): rbac.RequirePermission middleware сюда не подходит,
// т.к. пути /files/* не содержат :projectId, а полагаться на ID из body для
// :fileId-маршрутов небезопасно (IDOR — alice's projectId, bob's fileId).
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	h.rbacSvc = rbacSvc

	files := api.Group("/files")
	files.Use(authMw)

	files.Post("/upload-url", h.GenerateUploadURL)
	files.Post("/:fileId/confirm", h.ConfirmUpload)
	files.Get("/:fileId/url", h.GetDownloadURL)
	files.Delete("/:fileId", h.DeleteFile)
}

// authorizeProject проверяет, что текущий пользователь имеет указанное
// разрешение в проекте. Возвращает ошибку Fiber с подходящим статусом
// (401/403/500) или nil, если доступ разрешён.
func (h *Handler) authorizeProject(c fiber.Ctx, userID, projectID uuid.UUID, perm rbac.Permission) error {
	if h.rbacSvc == nil {
		// Безопасный default: без RBAC отказываем — лучше упасть, чем разрешить.
		return fiber.NewError(fiber.StatusInternalServerError, "RBAC service not configured")
	}
	effective, err := h.rbacSvc.GetEffectivePermissions(c.Context(), userID, projectID)
	if err != nil {
		if ent.IsNotFound(err) {
			return fiber.NewError(fiber.StatusForbidden, "not a project member")
		}
		slog.Error("rbac check failed", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
	}
	if !effective.Has(perm) {
		return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
	}
	return nil
}

// userIDFromCtx извлекает userID из Fiber Locals.
func userIDFromCtx(c fiber.Ctx) (uuid.UUID, error) {
	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
	}
	id, err := uuid.Parse(userIDStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "invalid user")
	}
	return id, nil
}

type uploadRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"contentType"`
	ProjectID   string `json:"projectId"`
}

// GenerateUploadURL godoc
// @Summary Получить URL для загрузки файла
// @Description Генерирует presigned S3 URL для загрузки и создаёт pending-запись метаданных.
// @Tags Файлы
// @Accept json
// @Produce json
// @Param body body uploadRequest true "Имя файла, тип и ID проекта"
// @Success 200 {object} UploadResult "URL для загрузки и ID файла"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/files/upload-url [post]
func (h *Handler) GenerateUploadURL(c fiber.Ctx) error {
	var req uploadRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.Filename == "" || req.ContentType == "" || req.ProjectID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "filename, contentType, and projectId are required")
	}

	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project ID")
	}

	// CR-04: проверяем право загрузки в проекте из тела запроса.
	if err := h.authorizeProject(c, userID, projectID, rbac.PermFileUpload); err != nil {
		return err
	}

	result, err := h.storage.GenerateUploadURL(c.Context(), req.Filename, req.ContentType, userID, projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to generate upload URL")
	}

	return c.JSON(result)
}

type confirmRequest struct {
	EntityType string `json:"entityType"`
	EntityID   string `json:"entityId"`
}

// ConfirmUpload godoc
// @Summary Подтвердить загрузку файла
// @Description Подтверждает успешную загрузку файла и привязывает его к сущности.
// @Tags Файлы
// @Accept json
// @Produce json
// @Param fileId path string true "ID файла"
// @Param body body confirmRequest true "Тип и ID сущности"
// @Success 200 {object} map[string]string "status: confirmed"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/files/{fileId}/confirm [post]
func (h *Handler) ConfirmUpload(c fiber.Ctx) error {
	fileID, err := httputil.ParseUUIDParam(c, "fileId")
	if err != nil {
		return err
	}

	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	// CR-04: грузим projectId самого файла, а не доверяем клиенту.
	projectID, err := h.storage.FileProjectID(c.Context(), fileID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "file not found")
	}
	if err := h.authorizeProject(c, userID, projectID, rbac.PermFileUpload); err != nil {
		return err
	}

	var req confirmRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if err := h.storage.ConfirmUpload(c.Context(), fileID, req.EntityType, req.EntityID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to confirm upload")
	}

	return c.JSON(fiber.Map{"status": "confirmed"})
}

// GetDownloadURL godoc
// @Summary Получить URL для скачивания
// @Description Генерирует временный presigned URL для скачивания файла.
// @Tags Файлы
// @Produce json
// @Param fileId path string true "ID файла"
// @Success 200 {object} map[string]string "downloadUrl"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/files/{fileId}/url [get]
func (h *Handler) GetDownloadURL(c fiber.Ctx) error {
	fileID, err := httputil.ParseUUIDParam(c, "fileId")
	if err != nil {
		return err
	}

	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	// CR-04: загружаем проект файла и проверяем PermProjectView у вызывающего.
	projectID, err := h.storage.FileProjectID(c.Context(), fileID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "file not found")
	}
	if err := h.authorizeProject(c, userID, projectID, rbac.PermProjectView); err != nil {
		return err
	}

	downloadURL, err := h.storage.GenerateDownloadURL(c.Context(), fileID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to generate download URL")
	}

	return c.JSON(fiber.Map{"downloadUrl": downloadURL})
}

// DeleteFile godoc
// @Summary Удалить файл
// @Description Мягкое удаление файла (помечает как удалённый).
// @Tags Файлы
// @Produce json
// @Param fileId path string true "ID файла"
// @Success 200 {object} map[string]string "status: deleted"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/files/{fileId} [delete]
func (h *Handler) DeleteFile(c fiber.Ctx) error {
	fileID, err := httputil.ParseUUIDParam(c, "fileId")
	if err != nil {
		return err
	}

	userID, err := userIDFromCtx(c)
	if err != nil {
		return err
	}

	// CR-04: грузим проект файла, проверяем PermFileDelete.
	projectID, err := h.storage.FileProjectID(c.Context(), fileID)
	if err != nil {
		return fiber.NewError(fiber.StatusNotFound, "file not found")
	}
	if err := h.authorizeProject(c, userID, projectID, rbac.PermFileDelete); err != nil {
		return err
	}

	if err := h.storage.DeleteFile(c.Context(), fileID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete file")
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}
