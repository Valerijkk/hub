package storage

import (
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler provides HTTP handlers for file operations.
type Handler struct {
	storage *StorageService
}

// NewHandler creates a new storage handler.
func NewHandler(storage *StorageService) *Handler {
	return &Handler{storage: storage}
}

// RegisterRoutes registers file storage routes on the given router group.
// All routes require authentication; upload/delete routes require file permissions.
func (h *Handler) RegisterRoutes(api fiber.Router, rbacSvc *rbac.RBACService) {
	files := api.Group("/files")
	files.Use(rbac.RequireAuth())

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

// GenerateUploadURL handles POST /api/files/upload-url
func (h *Handler) GenerateUploadURL(c fiber.Ctx) error {
	var req uploadRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid request body")
	}

	if req.Filename == "" || req.ContentType == "" || req.ProjectID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "filename, contentType, and projectId are required")
	}

	userID, err := uuid.Parse(rbac.UserIDFromCtx(c))
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

// ConfirmUpload handles POST /api/files/:fileId/confirm
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

// GetDownloadURL handles GET /api/files/:fileId/url
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

// DeleteFile handles DELETE /api/files/:fileId
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
