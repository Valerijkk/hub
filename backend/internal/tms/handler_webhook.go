package tms

import (
	"encoding/json"
	"io"
	"log/slog"

	"github.com/gofiber/fiber/v3"
	apperr "github.com/valerii/hub/backend/internal/platform/errors"
	httputil "github.com/valerii/hub/backend/internal/platform/http"
)

// apperr используется в swag @Failure аннотациях.
var _ *apperr.AppError
// ---------------------------------------------------------------------------
// Обработчики вебхуков CI/CD
// ---------------------------------------------------------------------------
// Упрощённые эндпоинты для CI/CD пайплайнов.
// Аутентификация через X-Hub-Token заголовок или query-параметр token.
// ---------------------------------------------------------------------------

// maxAllureFileBytes — лимит размера на каждый загружаемый файл в multipart-вебхуках.
// Глобальный Fiber BodyLimit покрывает совокупный размер запроса, но не отдельный part —
// без этой защиты multipart с большим числом крупных файлов может уронить процесс по памяти.
const maxAllureFileBytes = 5 * 1024 * 1024 // 5 МБ на каждый загружаемый файл

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

// WebhookAllure обрабатывает POST /api/projects/:projectId/tms/webhook
// @Summary CI/CD webhook для импорта результатов Allure
// @Description Принимает multipart (JSON-файл allure-results) или raw JSON-массив результатов.
// @Description ZIP-архивы не поддерживаются — передавайте JSON-массив или один .json файл.
// @Description Аутентификация через X-Hub-Token заголовок или query-параметр token.
// @Tags TMS
// @Accept multipart/form-data,application/json
// @Produce json
// @Param projectId path string true "Project ID"
// @Param X-Hub-Token header string false "API-ключ"
// @Param token query string false "API-ключ (альтернативный способ)"
// @Param title query string false "Заголовок прогона"
// @Success 201 {object} WebhookResponse
// @Failure 400 {object} apperr.AppError
// @Failure 401 {object} apperr.AppError
// @Failure 413 {object} apperr.AppError
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

		allureData, err = io.ReadAll(io.LimitReader(f, maxAllureFileBytes+1))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "не удалось прочитать загруженный файл")
		}
		if len(allureData) > maxAllureFileBytes {
			return fiber.NewError(fiber.StatusRequestEntityTooLarge, "загруженный файл превышает лимит 5 MB")
		}

		// Если title не задан в query, пробуем из form field
		if title == "" {
			title = c.FormValue("title", "")
		}
	} else {
		// A2W2-1-005: raw body — тот же cap, что и multipart (не только Fiber BodyLimit 25 MB).
		allureData = c.Body()
		if len(allureData) > maxAllureFileBytes {
			return fiber.NewError(fiber.StatusRequestEntityTooLarge, "тело запроса превышает лимит 5 MB")
		}
	}

	if len(allureData) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "пустое тело запроса")
	}

	// Пытаемся распарсить как JSON-массив результатов
	var rawResults json.RawMessage
	if err := json.Unmarshal(allureData, &rawResults); err != nil {
		// P3-BE-08 / W2-A1-011: не отдаём internals парсера клиенту.
		slog.Error("невалидный Allure JSON в webhook", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "невалидный JSON")
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

// WebhookJUnit обрабатывает POST /api/projects/:projectId/tms/webhook/junit
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
// @Failure 400 {object} apperr.AppError
// @Failure 401 {object} apperr.AppError
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

		xmlData, err = io.ReadAll(io.LimitReader(f, maxAllureFileBytes+1))
		if err != nil {
			return fiber.NewError(fiber.StatusBadRequest, "не удалось прочитать загруженный файл")
		}
		if len(xmlData) > maxAllureFileBytes {
			return fiber.NewError(fiber.StatusRequestEntityTooLarge, "загруженный файл превышает лимит 5 MB")
		}

		if title == "" {
			title = c.FormValue("title", "")
		}
	} else {
		// A2W2-1-005: raw XML body — тот же 5 MB cap, что multipart.
		xmlData = c.Body()
		if len(xmlData) > maxAllureFileBytes {
			return fiber.NewError(fiber.StatusRequestEntityTooLarge, "тело запроса превышает лимит 5 MB")
		}
	}

	if len(xmlData) == 0 {
		return fiber.NewError(fiber.StatusBadRequest, "пустое тело запроса")
	}

	// Парсим JUnit XML -> AllureResult[]
	allureResults, err := ParseJUnitXML(xmlData)
	if err != nil {
		// P3-BE-08 / W2-A1-011: детали парсера только в slog.
		slog.Error("ошибка парсинга JUnit XML", "error", err)
		return fiber.NewError(fiber.StatusBadRequest, "невалидный JUnit XML")
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
