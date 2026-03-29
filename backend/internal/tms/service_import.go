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
// Allure import
// ---------------------------------------------------------------------------

// SubjectTestRunCompleted is the NATS subject for automatic run completion events.
const SubjectTestRunCompleted = "tms.run.completed.%s"

// ImportAllureResults imports Allure results into TMS.
// Creates an automatic run, matches results to test cases,
// stores raw JSON in MongoDB, structured results in PostgreSQL.
func (s *TMSService) ImportAllureResults(ctx context.Context, projectID, userID uuid.UUID, req ImportAllureRequest) (*TestRunDetailResponse, error) {
	// 1. Parse allure-results JSON
	allureResults, err := ParseAllureResults(req.Results)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный allure JSON: " + err.Error())
	}
	if len(allureResults) == 0 {
		return nil, apperrors.NewBadRequest("пустой массив результатов Allure")
	}

	// 2. Create run
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

	// 3. Store raw results in MongoDB
	if err := s.repo.StoreAllureResults(ctx, run.ID, allureResults); err != nil {
		slog.Error("ошибка сохранения allure результатов в MongoDB", "error", err, "run_id", run.ID)
		// Don't abort - PG results are more important
	}

	// 4. Match each result to a test case and create TestResult
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

		// Duration from start/stop
		if ar.Start != nil && ar.Stop != nil {
			duration := *ar.Stop - *ar.Start
			input.DurationMs = &duration
			startTime := time.UnixMilli(*ar.Start)
			stopTime := time.UnixMilli(*ar.Stop)
			input.StartedAt = &startTime
			input.CompletedAt = &stopTime
		}

		// Error message
		if ar.StatusDetails != nil {
			input.ErrorMessage = ar.StatusDetails.Message
			input.ErrorTrace = ar.StatusDetails.Trace
		}

		// Link to test case
		if tc != nil {
			input.TestCaseID = &tc.ID
			input.MatchStrategy = strategy.String()
		} else {
			input.MatchStrategy = "unmapped"
		}

		resultInputs = append(resultInputs, input)
	}

	// 5. Batch insert results
	if len(resultInputs) > 0 {
		_, err = s.repo.BatchCreateTestResults(ctx, run.ID, resultInputs)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка пакетного создания результатов")
		}
	}

	// 6. Recompute stats
	if err := s.recomputeRunStats(ctx, run.ID); err != nil {
		return nil, err
	}

	// 7. Transition run: created -> in_progress -> completed
	now := time.Now()
	_, err = s.repo.UpdateTestRunStatus(ctx, run.ID, "in_progress", &now)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка старта прогона")
	}
	_, err = s.repo.UpdateTestRunStatus(ctx, run.ID, "completed", &now)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка завершения прогона")
	}

	// 8. Publish NATS event
	s.nats.PublishJSON(fmt.Sprintf(SubjectTestRunCompleted, run.ID.String()), map[string]interface{}{
		"run_id":     run.ID,
		"project_id": projectID,
		"title":      title,
		"total":      len(resultInputs),
	})

	// 9. Return detailed response
	return s.GetTestRunDetail(ctx, run.ID)
}

// ImportAllureToExistingRun imports Allure results into an existing run.
func (s *TMSService) ImportAllureToExistingRun(ctx context.Context, projectID, runID, userID uuid.UUID, req ImportAllureRequest) (*TestRunDetailResponse, error) {
	// Verify run exists
	run, err := s.repo.GetTestRun(ctx, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}

	// Parse results
	allureResults, err := ParseAllureResults(req.Results)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный allure JSON: " + err.Error())
	}
	if len(allureResults) == 0 {
		return nil, apperrors.NewBadRequest("пустой массив результатов Allure")
	}

	// Store in MongoDB
	if err := s.repo.StoreAllureResults(ctx, run.ID, allureResults); err != nil {
		slog.Error("ошибка сохранения allure результатов в MongoDB", "error", err, "run_id", run.ID)
	}

	// Match and create results
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
// File import/export
// ---------------------------------------------------------------------------

// ImportTestCases imports test cases from an uploaded file.
func (s *TMSService) ImportTestCases(c fiber.Ctx, projectID, userID uuid.UUID) (*ImportResult, error) {
	return handleImport(s, c, projectID, userID)
}

// ExportTestCases exports project test cases in a specified format.
func (s *TMSService) ExportTestCases(ctx context.Context, projectID uuid.UUID, format string) (io.Reader, string, error) {
	return handleExportTestCases(s, ctx, projectID, format)
}

// ExportTestRunReport exports a run report in a specified format.
func (s *TMSService) ExportTestRunReport(ctx context.Context, runID uuid.UUID, format string) (io.Reader, string, error) {
	return handleExportTestRunReport(s, ctx, runID, format)
}

