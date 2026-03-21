package backup

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
)

// Handler предоставляет HTTP-обработчики для модуля резервного копирования.
type Handler struct {
	svc *BackupService
}

// NewHandler создаёт новый обработчик бэкапов.
func NewHandler(svc *BackupService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes регистрирует маршруты администрирования бэкапов.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	backups := api.Group("/admin/backups", authMw)
	backups.Post("/", h.TriggerBackup)
	backups.Get("/", h.ListBackups)
	backups.Post("/:id/restore", h.RestoreBackup)
	backups.Delete("/:id", h.DeleteBackup)
}

// TriggerBackup обрабатывает POST /api/admin/backups — запуск бэкапа вручную.
func (h *Handler) TriggerBackup(c fiber.Ctx) error {
	resp, err := h.svc.PerformBackup(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListBackups обрабатывает GET /api/admin/backups.
func (h *Handler) ListBackups(c fiber.Ctx) error {
	backups, err := h.svc.ListBackups(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(backups)
}

// RestoreBackup обрабатывает POST /api/admin/backups/:id/restore.
// Требует подтверждения: тело запроса должно содержать {"confirm": "RESTORE"}.
func (h *Handler) RestoreBackup(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	backupID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID бэкапа")
	}

	var req RestoreRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}
	if req.Confirm != "RESTORE" {
		return fiber.NewError(fiber.StatusBadRequest, "для подтверждения восстановления отправьте {\"confirm\": \"RESTORE\"}")
	}

	if err := h.svc.RestoreBackup(c.Context(), backupID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(fiber.Map{"message": "восстановление запущено"})
}

// DeleteBackup обрабатывает DELETE /api/admin/backups/:id.
func (h *Handler) DeleteBackup(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	backupID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID бэкапа")
	}

	if err := h.svc.DeleteBackup(c.Context(), backupID); err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	return c.SendStatus(fiber.StatusNoContent)
}
