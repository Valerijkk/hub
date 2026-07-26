package tms

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Импорт Allure
// ---------------------------------------------------------------------------

// SubjectTestRunCompleted — NATS-subject для событий автозавершения прогона.
const SubjectTestRunCompleted = "tms.run.completed.%s"

// ImportAllureResults импортирует результаты Allure в TMS.
// Создаёт автоматический прогон, матчит результаты к тест-кейсам,
// сохраняет raw JSON в MongoDB, структурированные результаты в PostgreSQL.
func (s *TMSService) ImportAllureResults(ctx context.Context, projectID, userID uuid.UUID, req ImportAllureRequest) (*TestRunDetailResponse, error) {
	// 1. Парсим JSON allure-results
	allureResults, err := ParseAllureResults(req.Results)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный allure JSON: " + err.Error())
	}
	if len(allureResults) == 0 {
		return nil, apperrors.NewBadRequest("пустой массив результатов Allure")
	}

	// 2. Создаём прогон
	title := req.Title
	if title == "" {
		title = fmt.Sprintf("Allure Import %s", time.Now().Format("2006-01-02 15:04"))
	}
	runInput := CreateTestRunInput{
		Title:       title,
		RunType:     "automatic",
		CreatedByID: userID,
		AllureRunID: req.AllureRunID,
	}
	run, err := s.repo.CreateTestRun(ctx, projectID, runInput)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания прогона")
	}

	// 3. Сохраняем сырые результаты в MongoDB
	if err := s.repo.StoreAllureResults(ctx, run.ID, allureResults); err != nil {
		slog.Error("ошибка сохранения allure результатов в MongoDB", "error", err, "run_id", run.ID)
		// Не прерываем импорт — PG-результаты важнее Mongo raw dump.
	}

	// 4. Сопоставляем каждый результат с тест-кейсом и создаём TestResult
	resultInputs := make([]CreateTestResultInput, 0, len(allureResults))
	for _, ar := range allureResults {
		tc, strategy, matchErr := s.matcher.MatchTestCase(ctx, projectID, &ar)
		if matchErr != nil {
			slog.Error("ошибка матчинга allure результата", "error", matchErr, "uuid", ar.UUID)
			continue
		}

		input := CreateTestResultInput{
			Status:     allureStatusToTMS(ar.Status),
			AllureUUID: ar.UUID,
		}

		// Длительность из start/stop — валидируем epoch, чтобы не записать мусор.
		const maxEpochMs = 4102444800000 // 2100-01-01 00:00:00 UTC в миллисекундах
		if ar.Start != nil && ar.Stop != nil &&
			*ar.Start > 0 && *ar.Stop > 0 &&
			*ar.Start < maxEpochMs && *ar.Stop < maxEpochMs &&
			*ar.Start <= *ar.Stop {
			duration := *ar.Stop - *ar.Start
			input.DurationMs = &duration
			startTime := time.UnixMilli(*ar.Start)
			stopTime := time.UnixMilli(*ar.Stop)
			input.StartedAt = &startTime
			input.CompletedAt = &stopTime
		}

		// Сообщение об ошибке
		if ar.StatusDetails != nil {
			input.ErrorMessage = ar.StatusDetails.Message
			input.ErrorTrace = ar.StatusDetails.Trace
		}

		// Привязка к тест-кейсу
		if tc != nil {
			input.TestCaseID = &tc.ID
			input.MatchStrategy = strategy.String()
		} else {
			input.MatchStrategy = "unmapped"
		}

		resultInputs = append(resultInputs, input)
	}

	// 5. Пакетная вставка результатов
	if len(resultInputs) > 0 {
		_, err = s.repo.BatchCreateTestResults(ctx, run.ID, resultInputs)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка пакетного создания результатов")
		}
	}

	// 6. Пересчитываем статистику
	if err := s.recomputeRunStats(ctx, run.ID); err != nil {
		return nil, err
	}

	// 7. Переход прогона: created -> in_progress -> completed
	now := time.Now()
	_, err = s.repo.UpdateTestRunStatus(ctx, run.ID, "in_progress", &now)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка старта прогона")
	}
	_, err = s.repo.UpdateTestRunStatus(ctx, run.ID, "completed", &now)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка завершения прогона")
	}

	// 8. Публикуем событие NATS
	s.nats.PublishJSON(fmt.Sprintf(SubjectTestRunCompleted, run.ID.String()), map[string]interface{}{
		"run_id":     run.ID,
		"project_id": projectID,
		"title":      title,
		"total":      len(resultInputs),
	})

	// 9. Возвращаем детальный ответ
	return s.GetTestRunDetail(ctx, run.ID)
}

// ImportAllureToExistingRun импортирует результаты Allure в существующий прогон.
func (s *TMSService) ImportAllureToExistingRun(ctx context.Context, projectID, runID, userID uuid.UUID, req ImportAllureRequest) (*TestRunDetailResponse, error) {
	// Проверяем существование прогона в рамках projectID (A3-012).
	run, err := s.repo.GetTestRun(ctx, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}
	if run.Edges.Project == nil || run.Edges.Project.ID != projectID {
		return nil, apperrors.NewNotFound("прогон не найден")
	}

	// Парсим результаты
	allureResults, err := ParseAllureResults(req.Results)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный allure JSON: " + err.Error())
	}
	if len(allureResults) == 0 {
		return nil, apperrors.NewBadRequest("пустой массив результатов Allure")
	}

	// Сохраняем в MongoDB
	if err := s.repo.StoreAllureResults(ctx, run.ID, allureResults); err != nil {
		slog.Error("ошибка сохранения allure результатов в MongoDB", "error", err, "run_id", run.ID)
	}

	// Сопоставляем и создаём результаты
	resultInputs := make([]CreateTestResultInput, 0, len(allureResults))
	for _, ar := range allureResults {
		tc, strategy, matchErr := s.matcher.MatchTestCase(ctx, projectID, &ar)
		if matchErr != nil {
			slog.Error("ошибка матчинга allure результата", "error", matchErr, "uuid", ar.UUID)
			continue
		}

		input := CreateTestResultInput{
			Status:     allureStatusToTMS(ar.Status),
			AllureUUID: ar.UUID,
		}
		if ar.Start != nil && ar.Stop != nil {
			duration := *ar.Stop - *ar.Start
			input.DurationMs = &duration
			startTime := time.UnixMilli(*ar.Start)
			stopTime := time.UnixMilli(*ar.Stop)
			input.StartedAt = &startTime
			input.CompletedAt = &stopTime
		}
		if ar.StatusDetails != nil {
			input.ErrorMessage = ar.StatusDetails.Message
			input.ErrorTrace = ar.StatusDetails.Trace
		}
		if tc != nil {
			input.TestCaseID = &tc.ID
			input.MatchStrategy = strategy.String()
		} else {
			input.MatchStrategy = "unmapped"
		}
		resultInputs = append(resultInputs, input)
	}

	if len(resultInputs) > 0 {
		_, err = s.repo.BatchCreateTestResults(ctx, run.ID, resultInputs)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка пакетного создания результатов")
		}
	}

	if err := s.recomputeRunStats(ctx, run.ID); err != nil {
		return nil, err
	}

	return s.GetTestRunDetail(ctx, run.ID)
}

// ---------------------------------------------------------------------------
// Импорт/экспорт файлов
// ---------------------------------------------------------------------------

// ImportTestCases импортирует тест-кейсы из загруженного файла.
func (s *TMSService) ImportTestCases(c fiber.Ctx, projectID, userID uuid.UUID) (*ImportResult, error) {
	return handleImport(s, c, projectID, userID)
}

// ExportTestCases экспортирует тест-кейсы проекта в указанном формате.
func (s *TMSService) ExportTestCases(ctx context.Context, projectID uuid.UUID, format string) (io.Reader, string, error) {
	return handleExportTestCases(s, ctx, projectID, format)
}

// ExportTestRunReport экспортирует отчёт прогона в указанном формате.
func (s *TMSService) ExportTestRunReport(ctx context.Context, runID uuid.UUID, format string) (io.Reader, string, error) {
	return handleExportTestRunReport(s, ctx, runID, format)
}

