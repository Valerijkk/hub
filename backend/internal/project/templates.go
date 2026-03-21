package project

import (
	"context"
	"fmt"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/projecttemplate"
)

// TemplateResponse — DTO ответа для шаблона проекта.
type TemplateResponse struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	TemplateType string                 `json:"template_type"`
	Config       map[string]interface{} `json:"config"`
	IsSystem     bool                   `json:"is_system"`
	CreatedAt    time.Time              `json:"created_at"`
}

// CreateTemplateRequest — запрос на сохранение проекта как шаблона.
type CreateTemplateRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=255"`
	Description string `json:"description,omitempty"`
}

// TemplateService управляет шаблонами проектов: системные + пользовательские.
type TemplateService struct {
	db *ent.Client
}

// NewTemplateService создаёт новый сервис шаблонов.
func NewTemplateService(db *ent.Client) *TemplateService {
	return &TemplateService{db: db}
}

// ListTemplates возвращает все доступные шаблоны проектов.
func (s *TemplateService) ListTemplates(ctx context.Context) ([]TemplateResponse, error) {
	templates, err := s.db.ProjectTemplate.Query().
		Order(ent.Asc(projecttemplate.FieldCreatedAt)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить шаблоны: %w", err)
	}

	result := make([]TemplateResponse, 0, len(templates))
	for _, t := range templates {
		result = append(result, *toTemplateResponse(t))
	}
	return result, nil
}

// GetTemplate возвращает шаблон по ID.
func (s *TemplateService) GetTemplate(ctx context.Context, id uuid.UUID) (*TemplateResponse, error) {
	t, err := s.db.ProjectTemplate.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fiber.NewError(fiber.StatusNotFound, "шаблон не найден")
		}
		return nil, fmt.Errorf("ошибка получения шаблона: %w", err)
	}
	return toTemplateResponse(t), nil
}

// SaveProjectAsTemplate сохраняет текущий проект как custom шаблон.
// Создаёт снимок конфигурации: statuses, labels, boards, custom_fields.
func (s *TemplateService) SaveProjectAsTemplate(ctx context.Context, projectID, userID uuid.UUID, req CreateTemplateRequest) (*TemplateResponse, error) {
	// Собираем конфигурацию проекта
	proj, err := s.db.Project.Get(ctx, projectID)
	if err != nil {
		return nil, fmt.Errorf("проект не найден: %w", err)
	}

	// Получаем статусы
	statuses, _ := s.db.Project.QueryWorkflowStatuses(proj).All(ctx)
	statusNames := make([]string, 0, len(statuses))
	for _, st := range statuses {
		statusNames = append(statusNames, st.Name)
	}

	// Получаем метки
	labels, _ := s.db.Project.QueryLabels(proj).All(ctx)
	labelNames := make([]string, 0, len(labels))
	for _, l := range labels {
		labelNames = append(labelNames, l.Name)
	}

	config := map[string]interface{}{
		"statuses": statusNames,
		"labels":   labelNames,
	}

	t, err := s.db.ProjectTemplate.Create().
		SetName(req.Name).
		SetDescription(req.Description).
		SetTemplateType(projecttemplate.TemplateTypeCustom).
		SetConfig(config).
		SetCreatedByID(userID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось сохранить шаблон: %w", err)
	}

	return toTemplateResponse(t), nil
}

// DeleteTemplate удаляет пользовательский шаблон (системные нельзя).
func (s *TemplateService) DeleteTemplate(ctx context.Context, id uuid.UUID) error {
	t, err := s.db.ProjectTemplate.Get(ctx, id)
	if err != nil {
		if ent.IsNotFound(err) {
			return fiber.NewError(fiber.StatusNotFound, "шаблон не найден")
		}
		return fmt.Errorf("ошибка получения шаблона: %w", err)
	}

	if t.IsSystem {
		return fiber.NewError(fiber.StatusForbidden, "системные шаблоны нельзя удалять")
	}

	return s.db.ProjectTemplate.DeleteOneID(id).Exec(ctx)
}

// SeedSystemTemplates создаёт системные шаблоны если их ещё нет.
func (s *TemplateService) SeedSystemTemplates(ctx context.Context) error {
	count, err := s.db.ProjectTemplate.Query().
		Where(projecttemplate.IsSystem(true)).
		Count(ctx)
	if err != nil {
		return err
	}
	if count > 0 {
		return nil // Уже созданы
	}

	templates := []struct {
		name       string
		ttype      projecttemplate.TemplateType
		desc       string
		config     map[string]interface{}
	}{
		{
			name:  "Empty",
			ttype: projecttemplate.TemplateTypeEmpty,
			desc:  "Минимальный шаблон с базовыми статусами",
			config: map[string]interface{}{
				"statuses": []string{"Open", "In Progress", "Done"},
				"labels":   []string{},
			},
		},
		{
			name:  "Software Dev",
			ttype: projecttemplate.TemplateTypeSoftwareDev,
			desc:  "Полный workflow для разработки ПО",
			config: map[string]interface{}{
				"statuses": []string{"Backlog", "To Do", "In Progress", "Code Review", "Testing", "Done"},
				"labels":   []string{"frontend", "backend", "infra", "urgent", "tech-debt"},
			},
		},
		{
			name:  "QA",
			ttype: projecttemplate.TemplateTypeQa,
			desc:  "Шаблон для команды тестирования",
			config: map[string]interface{}{
				"statuses": []string{"Backlog", "Ready for Testing", "In Testing", "Failed", "Passed", "Done"},
				"labels":   []string{"regression", "smoke", "integration", "e2e", "manual"},
			},
		},
		{
			name:  "Kanban Simple",
			ttype: projecttemplate.TemplateTypeKanbanSimple,
			desc:  "Простой Kanban с тремя колонками",
			config: map[string]interface{}{
				"statuses": []string{"To Do", "Doing", "Done"},
				"labels":   []string{},
			},
		},
	}

	for _, tmpl := range templates {
		_, err := s.db.ProjectTemplate.Create().
			SetName(tmpl.name).
			SetDescription(tmpl.desc).
			SetTemplateType(tmpl.ttype).
			SetConfig(tmpl.config).
			SetIsSystem(true).
			Save(ctx)
		if err != nil {
			return fmt.Errorf("не удалось создать системный шаблон %s: %w", tmpl.name, err)
		}
	}

	return nil
}

// toTemplateResponse конвертирует Ent-сущность в DTO ответа.
func toTemplateResponse(t *ent.ProjectTemplate) *TemplateResponse {
	return &TemplateResponse{
		ID:           t.ID.String(),
		Name:         t.Name,
		Description:  t.Description,
		TemplateType: string(t.TemplateType),
		Config:       t.Config,
		IsSystem:     t.IsSystem,
		CreatedAt:    t.CreatedAt,
	}
}

// --- HTTP handlers ---

// TemplateHandler предоставляет HTTP-обработчики для шаблонов проектов.
type TemplateHandler struct {
	svc *TemplateService
}

// NewTemplateHandler создаёт обработчик шаблонов.
func NewTemplateHandler(svc *TemplateService) *TemplateHandler {
	return &TemplateHandler{svc: svc}
}

// RegisterRoutes регистрирует маршруты шаблонов проектов.
func (h *TemplateHandler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	templates := api.Group("/project-templates", authMw)
	templates.Get("/", h.ListTemplates)
	templates.Post("/", h.SaveAsTemplate)
	templates.Delete("/:id", h.DeleteTemplate)
}

// ListTemplates обрабатывает GET /api/project-templates.
func (h *TemplateHandler) ListTemplates(c fiber.Ctx) error {
	templates, err := h.svc.ListTemplates(c.Context())
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}
	return c.JSON(templates)
}

// SaveAsTemplate обрабатывает POST /api/project-templates.
func (h *TemplateHandler) SaveAsTemplate(c fiber.Ctx) error {
	userID, err := templateUserID(c)
	if err != nil {
		return err
	}

	var req struct {
		CreateTemplateRequest
		ProjectID string `json:"project_id" validate:"required"`
	}
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный project_id")
	}

	resp, err := h.svc.SaveProjectAsTemplate(c.Context(), projectID, userID, req.CreateTemplateRequest)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// DeleteTemplate обрабатывает DELETE /api/project-templates/:id.
func (h *TemplateHandler) DeleteTemplate(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID шаблона")
	}

	if err := h.svc.DeleteTemplate(c.Context(), id); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// templateUserID извлекает user ID из Fiber Locals.
func templateUserID(c fiber.Ctx) (uuid.UUID, error) {
	idStr, ok := c.Locals("userID").(string)
	if !ok || idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusUnauthorized, "пользователь не аутентифицирован")
	}
	return uuid.Parse(idStr)
}
