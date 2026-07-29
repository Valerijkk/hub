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

// maxEpochMs — верхняя граница epoch-ms (2100-01-01 UTC) для Allure start/stop.
const maxEpochMs = 4102444800000

// automaticImporter — Ent-путь с атомарным create+batch+status (A2W2-2-004).
type automaticImporter interface {
	CreateAutomaticImportRun(ctx context.Context, projectID uuid.UUID, input CreateTestRunInput, results []CreateTestResultInput, stats TestRunStats) (*ent.TestRun, error)
}

// existingRunImporter — A3W2-2-003: batch+stats в одной Ent tx для ExistingRun.
type existingRunImporter interface {
	ImportResultsToExistingRun(ctx context.Context, runID uuid.UUID, results []CreateTestResultInput) error
}

// ImportAllureResults импортирует результаты Allure в TMS.
// Создаёт автоматический прогон, матчит результаты к тест-кейсам,
// сохраняет raw JSON в MongoDB, структурированные результаты в PostgreSQL.
func (s *TMSService) ImportAllureResults(ctx context.Context, projectID, userID uuid.UUID, req ImportAllureRequest) (*TestRunDetailResponse, error) {
	// 1. Парсим JSON allure-results
	allureResults, err := ParseAllureResults(req.Results)
	if err != nil {
		// A2W2-2-011: фиксированное сообщение; детали парсера только в slog.
		slog.Error("некорректный allure JSON", "error", err)
		return nil, apperrors.NewBadRequest("некорректный allure JSON")
	}
	if len(allureResults) == 0 {
		return nil, apperrors.NewBadRequest("пустой массив результатов Allure")
	}

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

	// 2. Сопоставляем результаты (до PG-записи)
	resultInputs, stats := s.buildAllureResultInputs(ctx, projectID, allureResults)

	var run *ent.TestRun
	if ai, ok := s.repo.(automaticImporter); ok {
		// A2W2-2-004: PG-часть в одной Ent tx.
		run, err = ai.CreateAutomaticImportRun(ctx, projectID, runInput, resultInputs, stats)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка импорта прогона")
		}
	} else {
		// Mock fallback: create → batch → status с abort-cleanup.
		run, err = s.repo.CreateTestRun(ctx, projectID, runInput)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка создания прогона")
		}
		if len(resultInputs) > 0 {
			if _, err = s.repo.BatchCreateTestResults(ctx, run.ID, resultInputs); err != nil {
				s.abortImportRun(ctx, run.ID)
				return nil, apperrors.NewInternal("ошибка пакетного создания результатов")
			}
		}
		if err := s.recomputeRunStats(ctx, run.ID); err != nil {
			s.abortImportRun(ctx, run.ID)
			return nil, err
		}
		now := time.Now()
		if _, err = s.repo.UpdateTestRunStatus(ctx, run.ID, "in_progress", &now); err != nil {
			s.abortImportRun(ctx, run.ID)
			return nil, apperrors.NewInternal("ошибка старта прогона")
		}
		if _, err = s.repo.UpdateTestRunStatus(ctx, run.ID, "completed", &now); err != nil {
			s.abortImportRun(ctx, run.ID)
			return nil, apperrors.NewInternal("ошибка завершения прогона")
		}
	}

	// 3. Mongo raw dump (best-effort, вне PG tx)
	if err := s.repo.StoreAllureResults(ctx, run.ID, allureResults); err != nil {
		slog.Error("ошибка сохранения allure результатов в MongoDB", "error", err, "run_id", run.ID)
	}

	// 4. NATS — единый subject через Event* (A2W2-2-007).
	s.nats.PublishJSON(EventTestRunCompleted, map[string]interface{}{
		"run_id":     run.ID,
		"project_id": projectID,
		"title":      title,
		"total":      len(resultInputs),
	})

	return s.GetTestRunDetail(ctx, projectID, run.ID)
}

// ImportAllureToExistingRun импортирует результаты Allure в существующий прогон.
func (s *TMSService) ImportAllureToExistingRun(ctx context.Context, projectID, runID, userID uuid.UUID, req ImportAllureRequest) (*TestRunDetailResponse, error) {
	run, err := s.repo.GetTestRunInProject(ctx, projectID, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}

	allureResults, err := ParseAllureResults(req.Results)
	if err != nil {
		slog.Error("некорректный allure JSON", "error", err, "run_id", runID)
		return nil, apperrors.NewBadRequest("некорректный allure JSON")
	}
	if len(allureResults) == 0 {
		return nil, apperrors.NewBadRequest("пустой массив результатов Allure")
	}

	if err := s.repo.StoreAllureResults(ctx, run.ID, allureResults); err != nil {
		slog.Error("ошибка сохранения allure результатов в MongoDB", "error", err, "run_id", run.ID)
	}

	resultInputs, _ := s.buildAllureResultInputs(ctx, projectID, allureResults)

	if eri, ok := s.repo.(existingRunImporter); ok {
		// A3W2-2-003: PG batch+stats в одной Ent tx.
		if err := eri.ImportResultsToExistingRun(ctx, run.ID, resultInputs); err != nil {
			return nil, apperrors.NewInternal("ошибка импорта результатов")
		}
	} else {
		// Mock fallback: batch → recompute с compensating delete при fail.
		var created []*ent.TestResult
		if len(resultInputs) > 0 {
			var batchErr error
			created, batchErr = s.repo.BatchCreateTestResults(ctx, run.ID, resultInputs)
			if batchErr != nil {
				return nil, apperrors.NewInternal("ошибка пакетного создания результатов")
			}
		}
		if err := s.recomputeRunStats(ctx, run.ID); err != nil {
			if len(created) > 0 {
				ids := make([]uuid.UUID, len(created))
				for i, c := range created {
					ids[i] = c.ID
				}
				if delErr := s.repo.DeleteTestResults(ctx, ids); delErr != nil {
					slog.Error("не удалось откатить partial allure results", "error", delErr, "run_id", run.ID)
				}
			}
			return nil, err
		}
	}

	return s.GetTestRunDetail(ctx, projectID, run.ID)
}

func (s *TMSService) buildAllureResultInputs(ctx context.Context, projectID uuid.UUID, allureResults []AllureResult) ([]CreateTestResultInput, TestRunStats) {
	resultInputs := make([]CreateTestResultInput, 0, len(allureResults))
	stats := TestRunStats{}
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

		// A2W2-2-005: тот же maxEpochMs guard, что в ImportAllureResults.
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
		stats.TotalCount++
		switch input.Status {
		case "passed":
			stats.PassedCount++
		case "failed":
			stats.FailedCount++
		case "broken":
			stats.BrokenCount++
		case "skipped":
			stats.SkippedCount++
		default:
			stats.UnknownCount++
		}
	}
	return resultInputs, stats
}

func (s *TMSService) abortImportRun(ctx context.Context, runID uuid.UUID) {
	now := time.Now()
	if _, err := s.repo.UpdateTestRunStatus(ctx, runID, "aborted", &now); err != nil {
		slog.Error("не удалось abort orphan allure-run", "error", err, "run_id", runID)
	}
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
func (s *TMSService) ExportTestRunReport(ctx context.Context, projectID, runID uuid.UUID, format string) (io.Reader, string, error) {
	return handleExportTestRunReport(s, ctx, projectID, runID, format)
}
