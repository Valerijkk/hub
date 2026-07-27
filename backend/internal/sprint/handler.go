package sprint

import (
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
	"github.com/valerii/hub/backend/internal/rbac"
)

// Handler предоставляет HTTP-обработчики для модуля спринтов.
type Handler struct {
	service *SprintService
	poker   *PokerService
}

// NewHandler создаёт новый обработчик спринтов.
func NewHandler(service *SprintService, poker *PokerService) *Handler {
	return &Handler{service: service, poker: poker}
}

// RegisterRoutes регистрирует все маршруты модуля спринтов.
// ВАЖНО (Fiber v3): RBAC-гард — второй variadic после handler (см. tms/handler.go).
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler, rbacSvc *rbac.RBACService) {
	sprints := api.Group("/projects/:projectId/sprints", authMw)

	sprints.Post("/", h.CreateSprint, rbac.RequirePermission(rbac.PermTaskCreate, rbacSvc))
	sprints.Get("/", h.ListSprints, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	sprints.Get("/:sprintId", h.GetSprint, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	sprints.Put("/:sprintId", h.UpdateSprint, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	sprints.Post("/:sprintId/start", h.StartSprint, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	sprints.Post("/:sprintId/complete", h.CompleteSprint, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	sprints.Put("/:sprintId/retro", h.UpdateRetroNotes, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))
	sprints.Get("/:sprintId/burndown", h.GetBurndown, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	sprints.Post("/:sprintId/poker", h.StartPokerSession, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))
	sprints.Post("/:sprintId/issues", h.AddIssuesToSprint, rbac.RequirePermission(rbac.PermTaskEdit, rbacSvc))

	// Velocity на уровне проекта
	velocity := api.Group("/projects/:projectId/velocity", authMw)
	velocity.Get("/", h.GetVelocity, rbac.RequirePermission(rbac.PermProjectView, rbacSvc))

	// Poker маршруты без projectId
	poker := api.Group("/sprints/poker", authMw)
	poker.Post("/:sessionId/vote", h.CastVote)
	poker.Post("/:sessionId/reveal", h.RevealVotes)
	poker.Get("/:sessionId", h.GetPokerSession)
}

// ---------------------------------------------------------------------------
// Обработчики спринтов
// ---------------------------------------------------------------------------

// CreateSprint — Создать спринт
// @Summary Создать спринт
// @Description Создаёт новый спринт в проекте.
// @Tags Спринты
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param body body CreateSprintRequest true "Данные спринта"
// @Success 201 {object} SprintResponse "Спринт создан"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/sprints [post]
func (h *Handler) CreateSprint(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	var req CreateSprintRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.CreateSprint(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListSprints — Список спринтов
// @Summary Список спринтов
// @Description Возвращает все спринты проекта.
// @Tags Спринты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} SprintResponse "Список спринтов"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/sprints [get]
func (h *Handler) ListSprints(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.ListSprints(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetSprint — Получить спринт
// @Summary Получить спринт
// @Description Возвращает детальную информацию о спринте.
// @Tags Спринты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param sprintId path string true "ID спринта"
// @Success 200 {object} SprintResponse "Данные спринта"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/sprints/{sprintId} [get]
func (h *Handler) GetSprint(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	sprintID, err := httputil.ParseUUIDParam(c, "sprintId")
	if err != nil {
		return err
	}

	resp, err := h.service.GetSprint(c.Context(), projectID, sprintID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateSprint — Обновить спринт
// @Summary Обновить спринт
// @Description Обновляет название, цель или даты спринта.
// @Tags Спринты
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param sprintId path string true "ID спринта"
// @Param body body UpdateSprintRequest true "Обновляемые данные"
// @Success 200 {object} SprintResponse "Спринт обновлён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/sprints/{sprintId} [put]
func (h *Handler) UpdateSprint(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	sprintID, err := httputil.ParseUUIDParam(c, "sprintId")
	if err != nil {
		return err
	}

	var req UpdateSprintRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.UpdateSprint(c.Context(), projectID, sprintID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// StartSprint — Запустить спринт
// @Summary Запустить спринт
// @Description Переводит спринт в статус "active". В проекте может быть только один активный спринт.
// @Tags Спринты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param sprintId path string true "ID спринта"
// @Success 200 {object} SprintResponse "Спринт запущен"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/sprints/{sprintId}/start [post]
func (h *Handler) StartSprint(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	sprintID, err := httputil.ParseUUIDParam(c,"sprintId")
	if err != nil {
		return err
	}

	resp, err := h.service.StartSprint(c.Context(), projectID, sprintID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// CompleteSprint — Завершить спринт
// @Summary Завершить спринт
// @Description Завершает спринт и формирует итоговую статистику.
// @Tags Спринты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param sprintId path string true "ID спринта"
// @Success 200 {object} SprintResponse "Спринт завершён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/sprints/{sprintId}/complete [post]
func (h *Handler) CompleteSprint(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	sprintID, err := httputil.ParseUUIDParam(c,"sprintId")
	if err != nil {
		return err
	}

	resp, err := h.service.CompleteSprint(c.Context(), projectID, sprintID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// UpdateRetroNotes — Обновить ретроспективу
// @Summary Обновить ретроспективу
// @Description Сохраняет заметки ретроспективы спринта.
// @Tags Спринты
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param sprintId path string true "ID спринта"
// @Param body body RetroNotesRequest true "Заметки ретро"
// @Success 200 {object} SprintResponse "Ретро обновлено"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/sprints/{sprintId}/retro [put]
func (h *Handler) UpdateRetroNotes(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	sprintID, err := httputil.ParseUUIDParam(c, "sprintId")
	if err != nil {
		return err
	}

	var req RetroNotesRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	resp, err := h.service.UpdateRetroNotes(c.Context(), projectID, sprintID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetBurndown — Burndown-диаграмма спринта
// @Summary Burndown-диаграмма спринта
// @Description Возвращает данные для построения burndown-графика спринта.
// @Tags Спринты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param sprintId path string true "ID спринта"
// @Success 200 {array} BurndownPoint "Данные burndown"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/sprints/{sprintId}/burndown [get]
func (h *Handler) GetBurndown(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	sprintID, err := httputil.ParseUUIDParam(c, "sprintId")
	if err != nil {
		return err
	}

	resp, err := h.service.GetBurndown(c.Context(), projectID, sprintID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetVelocity — Velocity проекта
// @Summary Velocity проекта
// @Description Возвращает метрики velocity по завершённым спринтам проекта.
// @Tags Спринты
// @Produce json
// @Param projectId path string true "ID проекта"
// @Success 200 {array} VelocityPoint "Данные velocity"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/velocity [get]
func (h *Handler) GetVelocity(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}

	resp, err := h.service.GetVelocity(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// AddIssuesToSprint — Добавить задачи в спринт
// @Summary Добавить задачи в спринт
// @Description Привязывает одну или несколько задач к указанному спринту.
// @Tags Спринты
// @Accept json
// @Param projectId path string true "ID проекта"
// @Param sprintId path string true "ID спринта"
// @Param body body MoveIssueToSprintRequest true "Список ID задач"
// @Success 204 "Задачи добавлены"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Failure 403 {object} fiber.Error "Нет доступа"
// @Security BearerAuth
// @Router /api/projects/{projectId}/sprints/{sprintId}/issues [post]
func (h *Handler) AddIssuesToSprint(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	sprintID, err := httputil.ParseUUIDParam(c, "sprintId")
	if err != nil {
		return err
	}

	var req MoveIssueToSprintRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	issueIDs := make([]uuid.UUID, 0, len(req.IssueIDs))
	for _, idStr := range req.IssueIDs {
		id, err := uuid.Parse(idStr)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "некорректный issue_id: "+idStr)
		}
		issueIDs = append(issueIDs, id)
	}

	if err := h.service.AddIssuesToSprint(c.Context(), projectID, sprintID, issueIDs); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// ---------------------------------------------------------------------------
// Обработчики Planning Poker
// ---------------------------------------------------------------------------

// StartPokerSession — Начать Planning Poker сессию
// @Summary Начать Planning Poker сессию
// @Description Создаёт сессию Planning Poker для оценки задачи в рамках спринта.
// @Tags Спринты
// @Accept json
// @Produce json
// @Param projectId path string true "ID проекта"
// @Param sprintId path string true "ID спринта"
// @Param body body StartPokerRequest true "ID задачи для оценки"
// @Success 201 {object} PokerSessionResponse "Сессия создана"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/projects/{projectId}/sprints/{sprintId}/poker [post]
func (h *Handler) StartPokerSession(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	sprintID, err := httputil.ParseUUIDParam(c, "sprintId")
	if err != nil {
		return err
	}

	// RT-005 sibling: poker session must not bind to a foreign sprint UUID.
	if _, err := h.service.GetSprint(c.Context(), projectID, sprintID); err != nil {
		return err
	}

	var req StartPokerRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	issueID, err := uuid.Parse(req.IssueID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный issue_id")
	}

	resp, err := h.poker.StartPokerSession(c.Context(), sprintID, issueID)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// CastVote — Проголосовать в Poker
// @Summary Проголосовать в Poker
// @Description Отправляет голос (оценку) участника в сессии Planning Poker.
// @Tags Спринты
// @Accept json
// @Param sessionId path string true "ID сессии Poker"
// @Param body body VoteRequest true "Значение оценки"
// @Success 204 "Голос учтён"
// @Failure 400 {object} fiber.Error "Неверный запрос"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/sprints/poker/{sessionId}/vote [post]
func (h *Handler) CastVote(c fiber.Ctx) error {
	sessionID := utils.CopyString(c.Params("sessionId"))
	if sessionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "session_id обязателен")
	}

	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	var req VoteRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, apperr.MsgBadRequest)
	}

	if err := h.poker.CastVote(c.Context(), sessionID, userID, req.Value); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// RevealVotes — Раскрыть голоса Poker
// @Summary Раскрыть голоса Poker
// @Description Раскрывает все голоса сессии Planning Poker и показывает результат.
// @Tags Спринты
// @Produce json
// @Param sessionId path string true "ID сессии Poker"
// @Success 200 {object} PokerSessionResponse "Результаты голосования"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/sprints/poker/{sessionId}/reveal [post]
func (h *Handler) RevealVotes(c fiber.Ctx) error {
	sessionID := utils.CopyString(c.Params("sessionId"))
	if sessionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "session_id обязателен")
	}

	resp, err := h.poker.RevealVotes(c.Context(), sessionID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// GetPokerSession — Получить сессию Poker
// @Summary Получить сессию Poker
// @Description Возвращает текущее состояние сессии Planning Poker.
// @Tags Спринты
// @Produce json
// @Param sessionId path string true "ID сессии Poker"
// @Success 200 {object} PokerSessionResponse "Данные сессии"
// @Failure 400 {object} fiber.Error "Неверный ID"
// @Failure 401 {object} fiber.Error "Не авторизован"
// @Security BearerAuth
// @Router /api/sprints/poker/{sessionId} [get]
func (h *Handler) GetPokerSession(c fiber.Ctx) error {
	sessionID := utils.CopyString(c.Params("sessionId"))
	if sessionID == "" {
		return fiber.NewError(fiber.StatusBadRequest, "session_id обязателен")
	}

	resp, err := h.poker.GetSession(c.Context(), sessionID)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

