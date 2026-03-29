package project

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent/projecttemplate"
)

// ---------------------------------------------------------------------------
// Структуры данных для экспорта/импорта шаблонов
// ---------------------------------------------------------------------------

// TemplateExport — JSON-представление шаблона для обмена между инстансами.
type TemplateExport struct {
	Name         string                 `json:"name"`
	Description  string                 `json:"description,omitempty"`
	TemplateType string                 `json:"template_type"`
	Config       map[string]interface{} `json:"config"`
	Version      int                    `json:"version"` // schema version for forward compatibility
}

const templateExportVersion = 1

// ---------------------------------------------------------------------------
// Методы сервиса
// ---------------------------------------------------------------------------

// ExportTemplate экспортирует шаблон проекта как JSON для обмена между инстансами.
func (s *TemplateService) ExportTemplate(ctx context.Context, templateID uuid.UUID) ([]byte, error) {
	t, err := s.db.ProjectTemplate.Get(ctx, templateID)
	if err != nil {
		return nil, fmt.Errorf("template not found: %w", err)
	}

	export := TemplateExport{
		Name:         t.Name,
		Description:  t.Description,
		TemplateType: string(t.TemplateType),
		Config:       t.Config,
		Version:      templateExportVersion,
	}

	data, err := json.MarshalIndent(export, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("error serializing template: %w", err)
	}

	return data, nil
}

// ImportTemplate импортирует шаблон из JSON. Создаёт его как кастомный шаблон.
func (s *TemplateService) ImportTemplate(ctx context.Context, userID uuid.UUID, data []byte) (*TemplateResponse, error) {
	var export TemplateExport
	if err := json.Unmarshal(data, &export); err != nil {
		return nil, fmt.Errorf("invalid template JSON: %w", err)
	}

	if export.Name == "" {
		return nil, fmt.Errorf("template name is required")
	}

	// Imported templates are always "custom" type, regardless of source.
	templateType := projecttemplate.TemplateTypeCustom

	config := export.Config
	if config == nil {
		config = make(map[string]interface{})
	}

	t, err := s.db.ProjectTemplate.Create().
		SetName(export.Name).
		SetDescription(export.Description).
		SetTemplateType(templateType).
		SetConfig(config).
		SetCreatedByID(userID).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("error creating template: %w", err)
	}

	return toTemplateResponse(t), nil
}

// ---------------------------------------------------------------------------
// HTTP-обработчики
// ---------------------------------------------------------------------------

// ExportTemplateHandler обрабатывает GET /api/project-templates/:id/export.
// @Summary Экспортировать шаблон проекта
// @Description Выгружает шаблон в формате JSON для обмена между инстансами
// @Tags Шаблоны проектов
// @Produce application/json
// @Security BearerAuth
// @Param id path string true "ID шаблона"
// @Success 200 {object} TemplateExport "JSON-файл шаблона"
// @Failure 400 {object} map[string]string "Некорректный ID"
// @Failure 500 {object} map[string]string "Ошибка экспорта"
// @Router /project-templates/{id}/export [get]
func (h *TemplateHandler) ExportTemplateHandler(c fiber.Ctx) error {
	idStr := utils.CopyString(c.Params("id"))
	id, err := uuid.Parse(idStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "invalid template ID")
	}

	data, err := h.svc.ExportTemplate(c.Context(), id)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Content-Type", "application/json")
	c.Set("Content-Disposition", "attachment; filename=template.json")
	return c.Send(data)
}

// ImportTemplateHandler обрабатывает POST /api/project-templates/import.
// @Summary Импортировать шаблон проекта
// @Description Загружает JSON-файл шаблона и создаёт его как пользовательский шаблон
// @Tags Шаблоны проектов
// @Accept multipart/form-data
// @Produce json
// @Security BearerAuth
// @Param file formData file true "JSON-файл шаблона"
// @Success 201 {object} TemplateResponse "Импортированный шаблон"
// @Failure 400 {object} map[string]string "Некорректный файл"
// @Router /project-templates/import [post]
func (h *TemplateHandler) ImportTemplateHandler(c fiber.Ctx) error {
	userID, err := templateUserID(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "file is required")
	}

	f, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "error opening file")
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "error reading file")
	}

	resp, err := h.svc.ImportTemplate(c.Context(), userID, data)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}
