package imex

import (
	"io"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/utils/v2"
	"github.com/google/uuid"
)

// Handler предоставляет HTTP-обработчики для импорта/экспорта задач.
type Handler struct {
	svc *ImExService
}

// NewHandler создаёт новый обработчик импорта/экспорта.
func NewHandler(svc *ImExService) *Handler {
	return &Handler{svc: svc}
}

// RegisterRoutes регистрирует маршруты импорта/экспорта.
func (h *Handler) RegisterRoutes(api fiber.Router, authMw fiber.Handler) {
	projects := api.Group("/projects/:projectId", authMw)
	projects.Post("/import/issues", h.ImportIssues)
	projects.Post("/import/jira", h.ImportJira)
	projects.Get("/export/issues", h.ExportIssues)
	projects.Get("/export/full", h.ExportProjectZIP)

	// PDF-экспорт прогонов и спринтов
	api.Get("/projects/:projectId/test-runs/:runId/export/pdf", authMw, h.ExportRunPDF)
	api.Get("/projects/:projectId/sprints/:sprintId/export/pdf", authMw, h.ExportSprintPDF)
}

// ImportIssues обрабатывает POST /api/projects/:projectId/import/issues (multipart).
func (h *Handler) ImportIssues(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	file, err := c.FormFile("file")
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "файл обязателен")
	}

	f, err := file.Open()
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ошибка открытия файла")
	}
	defer f.Close()

	data, err := io.ReadAll(f)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "ошибка чтения файла")
	}

	preview := c.Query("mode", "preview") == "preview"

	// CSV по умолчанию
	previewResp, importResp, err := h.svc.ImportIssuesCSV(c.Context(), projectID, data, nil, preview)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	if previewResp != nil {
		return c.JSON(previewResp)
	}
	return c.JSON(importResp)
}

// ExportIssues обрабатывает GET /api/projects/:projectId/export/issues?format=csv|excel.
func (h *Handler) ExportIssues(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	format := c.Query("format", "csv")

	var data []byte
	var contentType, filename string

	switch format {
	case "excel", "xlsx":
		data, err = h.svc.ExportIssuesExcel(c.Context(), projectID)
		contentType = "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"
		filename = "issues.xlsx"
	default:
		data, err = h.svc.ExportIssuesCSV(c.Context(), projectID)
		contentType = "text/csv"
		filename = "issues.csv"
	}

	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Content-Type", contentType)
	c.Set("Content-Disposition", "attachment; filename="+filename)
	return c.Send(data)
}

// ExportProjectZIP обрабатывает GET /api/projects/:projectId/export/full.
func (h *Handler) ExportProjectZIP(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	data, err := h.svc.ExportProjectZIP(c.Context(), projectID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Content-Type", "application/zip")
	c.Set("Content-Disposition", "attachment; filename=project-export.zip")
	return c.Send(data)
}

// ExportRunPDF обрабатывает GET /api/projects/:projectId/test-runs/:runId/export/pdf.
func (h *Handler) ExportRunPDF(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	runIDStr := utils.CopyString(c.Params("runId"))
	runID, err := uuid.Parse(runIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный run ID")
	}

	data, err := h.svc.ExportRunPDF(c.Context(), projectID, runID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=test-run-report.pdf")
	return c.Send(data)
}

// ExportSprintPDF обрабатывает GET /api/projects/:projectId/sprints/:sprintId/export/pdf.
func (h *Handler) ExportSprintPDF(c fiber.Ctx) error {
	projectID, err := parseProjectID(c)
	if err != nil {
		return err
	}

	sprintIDStr := utils.CopyString(c.Params("sprintId"))
	sprintID, err := uuid.Parse(sprintIDStr)
	if err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "некорректный sprint ID")
	}

	data, err := h.svc.ExportSprintPDF(c.Context(), projectID, sprintID)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, err.Error())
	}

	c.Set("Content-Type", "application/pdf")
	c.Set("Content-Disposition", "attachment; filename=sprint-report.pdf")
	return c.Send(data)
}

// parseProjectID извлекает projectId из параметров пути.
func parseProjectID(c fiber.Ctx) (uuid.UUID, error) {
	idStr := utils.CopyString(c.Params("projectId"))
	if idStr == "" {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "project ID обязателен")
	}
	id, err := uuid.Parse(idStr)
	if err != nil {
		return uuid.Nil, fiber.NewError(fiber.StatusBadRequest, "некорректный project ID")
	}
	return id, nil
}
