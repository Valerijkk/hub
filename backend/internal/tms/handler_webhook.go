package tms

import (
	"encoding/json"
	"io"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// ---------------------------------------------------------------------------
// CI/CD Webhook handlers
// ---------------------------------------------------------------------------
// Упрощённые эндпоинты для CI/CD пайплайнов.
// Аутентификация через X-Hub-Token заголовок или query-параметр token.
// ---------------------------------------------------------------------------

// WebhookResponse — ответ вебхука с ID прогона и сводкой.
type WebhookResponse struct {
	RunID   string `json:"run_id"`
	Title   string `json:"title"`
	Total   int    `json:"total"`
	Passed  int    `json:"passed"`
	Failed  int    `json:"failed"`
	Skipped int    `json:"skipped"`
	Broken  int    `json:"broken"`
	Status  string `json:"status"`
}

// WebhookAllure handles POST /api/projects/:projectId/tms/webhook
// @Summary CI/CD webhook для импорта результатов Allure
// @Description Принимает multipart (zip-архив allure-results) или JSON-массив результатов.
// @Description Аутентификация через X-Hub-Token заголовок или query-параметр token.
// @Tags TMS
// @Accept multipart/form-data,application/json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param X-Hub-Token header string false "API-ключ"
// @Param token query string false "API-ключ (альтернативный способ)"
// @Param title query string false "Заголовок прогона"
// @Success 201 {object} WebhookResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/webhook [post]
func (h *Handler) WebhookAllure(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	title := c.Query("title", "")

	// Определяем формат: multipart или JSON
	contentType := c.Get("Content-Type")
	var allureData []byte

	if isMultipart(contentType) {
		// Multipart: читаем файл "file" или "results"
		file, fileErr := c.FormFile("file")
		if fileErr != nil {
			file, fileErr = c.FormFile("results")
		}
		if fileErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "в multipart-запросе ожидается поле 'file' или 'results'")
		}

		f, openErr := file.Open()
		if openErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "не удалось открыть загруженный файл")
		}
		defer f.Close()

		allureData, err = io.ReadAll(f)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "не удалось прочитать загруженный файл")
		}

		// Если title не задан в query, пробуем из form field
		if title == "" {
			title = c.FormValue("title", "")
		}
	} else {
		// JSON body
		allureData = c.Body()
	}

	if len(allureData) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "пустое тело запроса")
	}

	// Пытаемся распарсить как JSON-массив результатов
	var rawResults json.RawMessage
	if err := json.Unmarshal(allureData, &rawResults); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "невалидный JSON: "+err.Error())
	}

	req := ImportAllureRequest{
		Results: rawResults,
		Title:   title,
	}

	detail, err := h.service.ImportAllureResults(c.Context(), projectID, userID, req)
	if err != nil {
		return err
	}

	resp := WebhookResponse{
		RunID:   detail.ID,
		Title:   detail.Title,
		Total:   detail.TotalCount,
		Passed:  detail.PassedCount,
		Failed:  detail.FailedCount,
		Skipped: detail.SkippedCount,
		Broken:  detail.BrokenCount,
		Status:  detail.Status,
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// WebhookJUnit handles POST /api/projects/:projectId/tms/webhook/junit
// @Summary CI/CD webhook для импорта результатов JUnit XML
// @Description Принимает multipart (XML-файл) или raw XML body.
// @Description Парсит JUnit XML, конвертирует в Allure-формат и создаёт прогон.
// @Tags TMS
// @Accept multipart/form-data,application/xml
// @Produce json
// @Param projectId path string true "Project ID"
// @Param X-Hub-Token header string false "API-ключ"
// @Param token query string false "API-ключ (альтернативный способ)"
// @Param title query string false "Заголовок прогона"
// @Success 201 {object} WebhookResponse
// @Failure 400 {object} errors.AppError
// @Failure 401 {object} errors.AppError
// @Router /api/projects/{projectId}/tms/webhook/junit [post]
func (h *Handler) WebhookJUnit(c fiber.Ctx) error {
	projectID, err := httputil.ParseProjectID(c)
	if err != nil {
		return err
	}
	userID, err := httputil.UserIDFromCtx(c)
	if err != nil {
		return err
	}

	title := c.Query("title", "")

	contentType := c.Get("Content-Type")
	var xmlData []byte

	if isMultipart(contentType) {
		file, fileErr := c.FormFile("file")
		if fileErr != nil {
			file, fileErr = c.FormFile("results")
		}
		if fileErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "в multipart-запросе ожидается поле 'file' или 'results'")
		}

		f, openErr := file.Open()
		if openErr != nil {
			return fiber.NewError(fiber.StatusBadRequest, "не удалось открыть загруженный файл")
		}
		defer f.Close()

		xmlData, err = io.ReadAll(f)
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "не удалось прочитать загруженный файл")
		}

		if title == "" {
			title = c.FormValue("title", "")
		}
	} else {
		xmlData = c.Body()
	}

	if len(xmlData) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "пустое тело запроса")
	}

	// Парсим JUnit XML -> AllureResult[]
	allureResults, err := ParseJUnitXML(xmlData)
	if err != nil {
		slog.Error("ошибка парсинга JUnit XML", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, err.Error())
	}

	// Конвертируем обратно в JSON для ImportAllureRequest
	resultsJSON, err := json.Marshal(allureResults)
	if err != nil {
		return fiber.NewError(fiber.StatusInternalServerError, "ошибка сериализации результатов")
	}

	if title == "" {
		title = "JUnit Import"
	}

	req := ImportAllureRequest{
		Results: resultsJSON,
		Title:   title,
	}

	detail, err := h.service.ImportAllureResults(c.Context(), projectID, userID, req)
	if err != nil {
		return err
	}

	resp := WebhookResponse{
		RunID:   detail.ID,
		Title:   detail.Title,
		Total:   detail.TotalCount,
		Passed:  detail.PassedCount,
		Failed:  detail.FailedCount,
		Skipped: detail.SkippedCount,
		Broken:  detail.BrokenCount,
		Status:  detail.Status,
	}

	return c.Status(fiber.StatusCreated).JSON(resp)
}

// isMultipart проверяет, является ли Content-Type multipart.
func isMultipart(contentType string) bool {
	return len(contentType) > 9 && contentType[:9] == "multipart"
}
