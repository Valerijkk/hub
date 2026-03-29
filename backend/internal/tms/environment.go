package tms

import (
	"context"
	"fmt"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/environment"
	"github.com/valerii/hub/backend/ent/project"
)

// EnvironmentResponse — DTO ответа для тестового окружения.
type EnvironmentResponse struct {
	ID        string                 `json:"id"`
	Name      string                 `json:"name"`
	Label     string                 `json:"label"`
	Config    map[string]interface{} `json:"config,omitempty"`
	CreatedAt string                 `json:"created_at"`
}

// CreateEnvironmentRequest — запрос на создание окружения.
type CreateEnvironmentRequest struct {
	Name   string                 `json:"name" validate:"required,min=1,max=100"`
	Label  string                 `json:"label" validate:"required,oneof=dev staging prod custom"`
	Config map[string]interface{} `json:"config,omitempty"`
}

// UpdateEnvironmentRequest — запрос на обновление окружения.
type UpdateEnvironmentRequest struct {
	Name   *string                 `json:"name,omitempty"`
	Label  *string                 `json:"label,omitempty"`
	Config *map[string]interface{} `json:"config,omitempty"`
}

// --- Сервисные методы ---

// CreateEnvironment создаёт тестовое окружение в проекте.
func (s *TMSService) CreateEnvironment(ctx context.Context, projectID uuid.UUID, req CreateEnvironmentRequest) (*EnvironmentResponse, error) {
	builder := s.repo.EntClient().Environment.Create().
		SetName(req.Name).
		SetEnvLabel(environment.EnvLabel(req.Label)).
		SetProjectID(projectID)

	if req.Config != nil {
		builder.SetConfig(req.Config)
	}

	env, err := builder.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать окружение: %w", err)
	}

	return toEnvironmentResponse(env), nil
}

// ListEnvironments возвращает все окружения проекта.
func (s *TMSService) ListEnvironments(ctx context.Context, projectID uuid.UUID) ([]EnvironmentResponse, error) {
	envs, err := s.repo.EntClient().Project.Query().
		Where(project.ID(projectID)).
		QueryEnvironments().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("не удалось получить окружения: %w", err)
	}

	result := make([]EnvironmentResponse, 0, len(envs))
	for _, e := range envs {
		result = append(result, *toEnvironmentResponse(e))
	}
	return result, nil
}

// UpdateEnvironment обновляет тестовое окружение.
func (s *TMSService) UpdateEnvironment(ctx context.Context, envID uuid.UUID, req UpdateEnvironmentRequest) (*EnvironmentResponse, error) {
	builder := s.repo.EntClient().Environment.UpdateOneID(envID)
	if req.Name != nil {
		builder.SetName(*req.Name)
	}
	if req.Label != nil {
		builder.SetEnvLabel(environment.EnvLabel(*req.Label))
	}
	if req.Config != nil {
		builder.SetConfig(*req.Config)
	}

	env, err := builder.Save(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, fiber.NewError(fiber.StatusNotFound, "окружение не найдено")
		}
		return nil, fmt.Errorf("не удалось обновить окружение: %w", err)
	}

	return toEnvironmentResponse(env), nil
}

// DeleteEnvironment удаляет тестовое окружение.
func (s *TMSService) DeleteEnvironment(ctx context.Context, envID uuid.UUID) error {
	err := s.repo.EntClient().Environment.DeleteOneID(envID).Exec(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return fiber.NewError(fiber.StatusNotFound, "окружение не найдено")
		}
		return fmt.Errorf("не удалось удалить окружение: %w", err)
	}
	return nil
}

// toEnvironmentResponse конвертирует Ent-сущность в DTO ответа.
func toEnvironmentResponse(e *ent.Environment) *EnvironmentResponse {
	return &EnvironmentResponse{
		ID:        e.ID.String(),
		Name:      e.Name,
		Label:     string(e.EnvLabel),
		Config:    e.Config,
		CreatedAt: e.CreatedAt.Format("2006-01-02T15:04:05Z"),
	}
}

// --- HTTP handlers для окружений ---

// CreateEnvironmentHandler обрабатывает POST /api/projects/:projectId/environments.
// @Summary Создать тестовое окружение
// @Description Создаёт новое тестовое окружение (dev/staging/prod/custom) в проекте
// @Tags Окружения
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Param body body CreateEnvironmentRequest true "Данные окружения"
// @Success 201 {object} EnvironmentResponse "Созданное окружение"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Router /projects/{projectId}/environments [post]
func (h *Handler) CreateEnvironmentHandler(c fiber.Ctx) error {
	projectID, err := parseTMSProjectID(c)
	if err != nil {
		return err
	}

	var req CreateEnvironmentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}
	if req.Name == "" {
		return fiber.NewError(fiber.StatusBadRequest, "name обязателен")
	}

	resp, err := h.service.CreateEnvironment(c.Context(), projectID, req)
	if err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// ListEnvironmentsHandler обрабатывает GET /api/projects/:projectId/environments.
// @Summary Получить список окружений
// @Description Возвращает все тестовые окружения указанного проекта
// @Tags Окружения
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Success 200 {array} EnvironmentResponse "Список окружений"
// @Failure 400 {object} map[string]string "Некорректный ID проекта"
// @Router /projects/{projectId}/environments [get]
func (h *Handler) ListEnvironmentsHandler(c fiber.Ctx) error {
	projectID, err := parseTMSProjectID(c)
	if err != nil {
		return err
	}

	envs, err := h.service.ListEnvironments(c.Context(), projectID)
	if err != nil {
		return err
	}

	return c.JSON(envs)
}

// UpdateEnvironmentHandler обрабатывает PUT /api/projects/:projectId/environments/:envId.
// @Summary Обновить тестовое окружение
// @Description Обновляет название, метку или конфигурацию окружения
// @Tags Окружения
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Param envId path string true "ID окружения"
// @Param body body UpdateEnvironmentRequest true "Поля для обновления"
// @Success 200 {object} EnvironmentResponse "Обновлённое окружение"
// @Failure 400 {object} map[string]string "Некорректный запрос"
// @Failure 404 {object} map[string]string "Окружение не найдено"
// @Router /projects/{projectId}/environments/{envId} [put]
func (h *Handler) UpdateEnvironmentHandler(c fiber.Ctx) error {
	envIDStr := utils.CopyString(c.Params("envId"))
	envID, err := uuid.Parse(envIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID окружения")
	}

	var req UpdateEnvironmentRequest
	if err := c.Bind().JSON(&req); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректное тело запроса")
	}

	resp, err := h.service.UpdateEnvironment(c.Context(), envID, req)
	if err != nil {
		return err
	}

	return c.JSON(resp)
}

// DeleteEnvironmentHandler обрабатывает DELETE /api/projects/:projectId/environments/:envId.
// @Summary Удалить тестовое окружение
// @Description Удаляет окружение из проекта. Действие необратимо
// @Tags Окружения
// @Security BearerAuth
// @Param projectId path string true "ID проекта"
// @Param envId path string true "ID окружения"
// @Success 204 "Окружение удалено"
// @Failure 400 {object} map[string]string "Некорректный ID"
// @Failure 404 {object} map[string]string "Окружение не найдено"
// @Router /projects/{projectId}/environments/{envId} [delete]
func (h *Handler) DeleteEnvironmentHandler(c fiber.Ctx) error {
	envIDStr := utils.CopyString(c.Params("envId"))
	envID, err := uuid.Parse(envIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный ID окружения")
	}

	if err := h.service.DeleteEnvironment(c.Context(), envID); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// parseTMSProjectID извлекает projectId из параметров пути.
func parseTMSProjectID(c fiber.Ctx) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params("projectId"))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "project ID обязателен")
	}
	return uuid.Parse(idStr)
}
