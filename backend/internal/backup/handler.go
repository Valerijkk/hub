package backup

import (
	"errors"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
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
// Требует системную роль admin — операции backup/restore/delete = total-data-loss primitive.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, adminMw fiber.Handler) {
	backups := api.Group("/admin/backups", authMw, adminMw)
	backups.Post("/", h.TriggerBackup)
	backups.Get("/", h.ListBackups)
	backups.Post("/:id/restore", h.RestoreBackup)
	backups.Delete("/:id", h.DeleteBackup)
}

// TriggerBackup godoc
// @Summary Запустить бэкап
// @Description Запускает создание резервной копии базы данных вручную.
// @Tags Резервное копирование
// @Produce json
// @Success 201 {object} BackupResponse "Бэкап создан"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/admin/backups [post]
func (h *Handler) TriggerBackup(c fiber.Ctx) error {
	resp, err := h.svc.PerformBackup(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "backup failed")
	}
	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListBackups godoc
// @Summary Список бэкапов
// @Description Возвращает все созданные резервные копии.
// @Tags Резервное копирование
// @Produce json
// @Success 200 {array} BackupResponse "Список бэкапов"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/admin/backups [get]
func (h *Handler) ListBackups(c fiber.Ctx) error {
	backups, err := h.svc.ListBackups(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "failed to list backups")
	}
	return c.JSON(backups)
}

// RestoreBackup godoc
// @Summary Восстановить из бэкапа
// @Description Восстанавливает базу данных из указанной резервной копии. Требует подтверждения.
// @Tags Резервное копирование
// @Accept json
// @Produce json
// @Param id path string true "ID бэкапа"
// @Param body body RestoreRequest true "Подтверждение: {confirm: RESTORE}"
// @Success 200 {object} map[string]string "Восстановление запущено"
// @Failure 400 {object} fiber.Error "Неверный запрос или нет подтверждения"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/admin/backups/{id}/restore [post]
func (h *Handler) RestoreBackup(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	backupID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID бэкапа")
	}

	var req RestoreRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}
	if req.Confirm != "RESTORE" {
		return fiber.NewError(fiber.StatusBadRequest, "для подтверждения восстановления отправьте {\"confirm\": \"RESTORE\"}")
	}

	if err := h.svc.RestoreBackup(c.Context(), backupID); err != nil {
		// W7-A4-011: missing id → 404; иначе generic 500 без stderr/URI (A7-007).
		if errors.Is(err, ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "backup not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "restore failed")
	}

	return c.JSON(fiber.Map{"message": "восстановление запущено"})
}

// DeleteBackup godoc
// @Summary Удалить бэкап
// @Description Удаляет резервную копию из хранилища.
// @Tags Резервное копирование
// @Param id path string true "ID бэкапа"
// @Success 204 "Бэкап удалён"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Только для админа"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/admin/backups/{id} [delete]
func (h *Handler) DeleteBackup(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	backupID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID бэкапа")
	}

	if err := h.svc.DeleteBackup(c.Context(), backupID); err != nil {
		// W7-A4-011: missing id → 404.
		if errors.Is(err, ErrNotFound) {
			return fiber.NewError(fiber.StatusNotFound, "backup not found")
		}
		return fiber.NewError(fiber.StatusInternalServerError, "delete failed")
	}

	return c.SendStatus(fiber.StatusNoContent)
}
