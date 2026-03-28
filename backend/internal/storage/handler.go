package storage

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики для файловых операций: генерация URL загрузки,
// подтверждение загрузки, генерация URL скачивания и удаление файлов.
type Handler struct {
	storage *StorageService
}

// NewHandler создаёт новый обработчик файлового хранилища на основе StorageService.
func NewHandler(storage *StorageService) *Handler {
	return &Handler{storage: storage}
}

// RegisterRoutes регистрирует маршруты файлового хранилища.
// authMw — JWT middleware аутентификации (auth.RequireAuth), валидирующий Bearer-токены.
// rbacSvc используется для проверки прав на загрузку и удаление файлов.
// Все маршруты /files/* требуют аутентификации; загрузка и удаление дополнительно
// требуют разрешений PermFileUpload и PermFileDelete соответственно.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	files := api.Group("/files")
	files.Use(authMw)

	files.Post("/upload-url", rbac.RequirePermission(rbac.PermFileUpload, rbacSvc), h.GenerateUploadURL)
	files.Post("/:fileId/confirm", h.ConfirmUpload)
	files.Get("/:fileId/url", h.GetDownloadURL)
	files.Delete("/:fileId", rbac.RequirePermission(rbac.PermFileDelete, rbacSvc), h.DeleteFile)
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
// @Success 200 {object} UploadURLResponse "URL для загрузки и ID файла"
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

	userIDStr, ok := c.Locals("userID").(string)
	if !ok || userIDStr == "" {
		return fiber.NewError(fiber.StatusUnauthorized, "user not authenticated")
	}

	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusUnauthorized, "invalid user")
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid project ID")
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
	fileID, err := uuid.Parse(c.Params("fileId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid file ID")
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
	fileID, err := uuid.Parse(c.Params("fileId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid file ID")
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
	fileID, err := uuid.Parse(c.Params("fileId"))
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid file ID")
	}

	if err := h.storage.DeleteFile(c.Context(), fileID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to delete file")
	}

	return c.JSON(fiber.Map{"status": "deleted"})
}
