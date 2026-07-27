package calls

import (
	"errors"
	"log/slog"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
	"github.com/valerii/hub/backend/internal/rbac"
)

// ---------------------------------------------------------------------------
// Обработчики записи звонков (CALL-04, CALL-05)
// ---------------------------------------------------------------------------

// StartRecording — Начать запись звонка
// @Summary Начать запись звонка
// @Description Запускает запись текущего звонка через LiveKit Egress.
// @Tags Звонки
// @Produce json
// @Param callId path string true "ID звонка"
// @Success 201 {object} RecordingResponse "Запись начата"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа к каналу звонка"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/recording/start [post]
func (h *Handler) StartRecording(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	resp, err := h.recSvc.StartRecording(c.Context(), callID, userID)
	if err != nil {
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "недостаточно прав для записи звонка")
		}
		slog.Error("ошибка начала записи", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось начать запись")
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// StopRecording — Остановить запись звонка
// @Summary Остановить запись звонка
// @Description Останавливает активную запись звонка.
// @Tags Звонки
// @Accept json
// @Param callId path string true "ID звонка"
// @Param body body object true "recording_id: ID записи"
// @Success 200 "Запись остановлена"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа к записи"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/calls/{callId}/recording/stop [post]
func (h *Handler) StopRecording(c fiber.Ctx) error {
	callID, err := parseCallID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req struct {
		RecordingID uuid.UUID `json:"recording_id"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}
	if req.RecordingID == uuid.Nil {
		return fiber.NewError(fiber.StatusBadRequest, "recording_id обязателен")
	}

	if err := h.recSvc.StopRecording(c.Context(), callID, req.RecordingID, userID); err != nil {
		if errors.Is(err, ErrCallForbidden) {
			return fiber.NewError(fiber.StatusForbidden, "недостаточно прав для остановки записи")
		}
		slog.Error("ошибка остановки записи", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось остановить запись")
	}

	return c.SendStatus(fiber.StatusOK)
}

// ListProjectRecordings — Список записей проекта
// @Summary Список записей проекта
// @Description Возвращает пагинированный список записей звонков проекта.
// @Tags Звонки
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param offset query int false "Смещение (по умолчанию 0)"
// @Param limit query int false "Лимит (по умолчанию 20, макс 100)"
// @Success 200 {object} map[string]interface{} "recordings и total"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/projects/{projectId}/recordings [get]
func (h *Handler) ListProjectRecordings(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("projectId"))
	projectID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный project ID")
	}

	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	// CR-02: проверяем, что пользователь — член проекта и имеет право просмотра.
	if h.rbacSvc != nil {
		effective, permErr := h.rbacSvc.GetEffectivePermissions(c.Context(), userID, projectID)
		if permErr != nil {
			if ent.IsNotFound(permErr) {
				return fiber.NewError(fiber.StatusForbidden, "not a project member")
			}
			slog.Error("rbac check failed", "error", permErr)
			return fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
		}
		if !effective.Has(rbac.PermProjectView) {
			return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
		}
	}

	offset := 0
	limit := 20
	if o := c.Query("offset"); o != "" {
		if parsed, parseErr := strconv.Atoi(o); parseErr == nil && parsed >= 0 {
			offset = parsed
		}
	}
	if l := c.Query("limit"); l != "" {
		if parsed, parseErr := strconv.Atoi(l); parseErr == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	recordings, total, err := h.recSvc.ListProjectRecordings(c.Context(), projectID, offset, limit)
	if err != nil {
		slog.Error("ошибка получения записей проекта", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось получить записи")
	}

	return c.JSON(fiber.Map{
		"recordings": recordings,
		"total":      total,
	})
}

// GetRecordingDownloadURL — Скачать запись
// @Summary Скачать запись
// @Description Возвращает presigned URL для скачивания записи звонка.
// @Tags Звонки
// @Produce json
// @Param recordingId path string true "ID записи"
// @Success 200 {object} RecordingDownloadResponse "URL для скачивания"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 500 {object} fiber.Error "Внутренняя ошибка"
// @Security BearerAuth
// @Router /api/recordings/{recordingId}/download [get]
func (h *Handler) GetRecordingDownloadURL(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("recordingId"))
	recordingID, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный recording ID")
	}

	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	// CR-02: загружаем запись для получения её проекта и проверяем доступ.
	if h.rbacSvc != nil {
		projectID, projErr := h.recSvc.RecordingProjectID(c.Context(), recordingID)
		if projErr != nil {
			slog.Error("recording lookup failed", "error", projErr)
			return fiber.NewError(fiber.StatusNotFound, "запись не найдена")
		}
		effective, permErr := h.rbacSvc.GetEffectivePermissions(c.Context(), userID, projectID)
		if permErr != nil {
			if ent.IsNotFound(permErr) {
				return fiber.NewError(fiber.StatusForbidden, "not a project member")
			}
			slog.Error("rbac check failed", "error", permErr)
			return fiber.NewError(fiber.StatusInternalServerError, "failed to check permissions")
		}
		if !effective.Has(rbac.PermProjectView) {
			return fiber.NewError(fiber.StatusForbidden, "insufficient permissions")
		}
	}

	resp, err := h.recSvc.GetRecordingDownloadURL(c.Context(), recordingID)
	if err != nil {
		slog.Error("ошибка получения URL записи", "error", err)
		return fiber.NewError(fiber.StatusInternalServerError, "не удалось получить URL записи")
	}

	return c.JSON(resp)
}
