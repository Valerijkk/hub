package tms

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Test runs
// ---------------------------------------------------------------------------

// CreateTestRun creates a new test run.
func (s *TMSService) CreateTestRun(ctx context.Context, projectID, userID uuid.UUID, req CreateTestRunRequest) (*TestRunResponse, error) {
	input := CreateTestRunInput{
		Title:       req.Title,
		Description: req.Description,
		RunType:     req.RunType,
		CreatedByID: userID,
	}

	run, err := s.repo.CreateTestRun(ctx, projectID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания прогона")
	}

	return toTestRunResponse(run), nil
}

// GetTestRun returns a test run by ID.
func (s *TMSService) GetTestRun(ctx context.Context, runID uuid.UUID) (*TestRunResponse, error) {
	run, err := s.repo.GetTestRun(ctx, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}
	return toTestRunResponse(run), nil
}

// ListTestRuns returns a paginated list of test runs.
func (s *TMSService) ListTestRuns(ctx context.Context, projectID uuid.UUID, req ListTestRunsRequest) (*PaginatedResponse[TestRunResponse], error) {
	params := ListTestRunsParams{
		Offset: req.Offset,
		Limit:  req.Limit,
		Status: req.Status,
	}

	items, total, err := s.repo.ListTestRuns(ctx, projectID, params)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения списка прогонов")
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	responses := make([]TestRunResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, *toTestRunResponse(item))
	}

	return &PaginatedResponse[TestRunResponse]{
		Items:  responses,
		Total:  total,
		Offset: req.Offset,
		Limit:  limit,
	}, nil
}

// UpdateRunStatus updates the status of a test run with state machine validation.
func (s *TMSService) UpdateRunStatus(ctx context.Context, runID uuid.UUID, newStatus string) (*TestRunResponse, error) {
	run, err := s.repo.GetTestRun(ctx, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}

	if !isValidRunTransition(string(run.Status), newStatus) {
		return nil, apperrors.NewBadRequest("недопустимый переход статуса: " + string(run.Status) + " -> " + newStatus)
	}

	now := time.Now()
	updated, err := s.repo.UpdateTestRunStatus(ctx, runID, newStatus, &now)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка обновления статуса прогона")
	}

	return toTestRunResponse(updated), nil
}

// ReopenTestRun reopens a completed or aborted test run.
func (s *TMSService) ReopenTestRun(ctx context.Context, runID uuid.UUID) (*TestRunResponse, error) {
	run, err := s.repo.GetTestRun(ctx, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}

	if string(run.Status) != "completed" && string(run.Status) != "aborted" {
		return nil, apperrors.NewBadRequest("переоткрыть можно только завершённый или прерванный прогон")
	}

	reopened, err := s.repo.ReopenTestRun(ctx, runID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка переоткрытия прогона")
	}

	return toTestRunResponse(reopened), nil
}

// ---------------------------------------------------------------------------
// Test results
// ---------------------------------------------------------------------------

// CreateTestResult creates a test result in a run.
func (s *TMSService) CreateTestResult(ctx context.Context, runID uuid.UUID, req CreateTestResultRequest) (*TestResultResponse, error) {
	input := CreateTestResultInput{
		Status:        req.Status,
		Comment:       req.Comment,
		StepResults:   req.StepResults,
		DurationMs:    req.DurationMs,
		MatchStrategy: req.MatchStrategy,
		IsFlaky:       req.IsFlaky,
	}

	if req.TestCaseID != "" {
		id, err := uuid.Parse(req.TestCaseID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный test_case_id")
		}
		input.TestCaseID = &id
	}

	result, err := s.repo.CreateTestResult(ctx, runID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания результата")
	}

	return toTestResultResponse(result), nil
}

// UpdateTestResult updates a test result.
func (s *TMSService) UpdateTestResult(ctx context.Context, resultID uuid.UUID, req UpdateTestResultRequest) (*TestResultResponse, error) {
	input := UpdateTestResultInput{
		Status:      req.Status,
		Comment:     req.Comment,
		StepResults: req.StepResults,
		IsFlaky:     req.IsFlaky,
	}

	result, err := s.repo.UpdateTestResult(ctx, resultID, input)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("результат не найден")
		}
		return nil, apperrors.NewInternal("ошибка обновления результата")
	}

	return toTestResultResponse(result), nil
}

// ListTestResults returns all results of a test run.
func (s *TMSService) ListTestResults(ctx context.Context, runID uuid.UUID) ([]TestResultResponse, error) {
	results, err := s.repo.ListTestResults(ctx, runID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения результатов")
	}

	responses := make([]TestResultResponse, 0, len(results))
	for _, r := range results {
		responses = append(responses, *toTestResultResponse(r))
	}
	return responses, nil
}

// ---------------------------------------------------------------------------
// Extended run operations
// ---------------------------------------------------------------------------

// AddTestCasesToRun adds test cases to a run, creating TestResult entries with unknown status.
func (s *TMSService) AddTestCasesToRun(ctx context.Context, runID uuid.UUID, req AddTestCasesToRunRequest) (*TestRunDetailResponse, error) {
	caseIDs, err := parseUUIDs(req.CaseIDs)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный ID тест-кейса")
	}

	if err := s.repo.AddTestCasesToRun(ctx, runID, caseIDs); err != nil {
		return nil, apperrors.NewInternal("ошибка добавления тест-кейсов в прогон")
	}

	// Recompute run stats
	if err := s.recomputeRunStats(ctx, runID); err != nil {
		return nil, err
	}

	return s.GetTestRunDetail(ctx, runID)
}

// ExecuteTestResult updates a test result with run status validation.
func (s *TMSService) ExecuteTestResult(ctx context.Context, runID, resultID, userID uuid.UUID, req UpdateTestResultRequest) (*TestResultResponse, error) {
	// Check that run is in_progress
	run, err := s.repo.GetTestRun(ctx, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}
	if string(run.Status) != "in_progress" {
		return nil, apperrors.NewBadRequest("выполнять результаты можно только в прогоне со статусом in_progress")
	}

	// Update result
	input := UpdateTestResultInput{
		Status:      req.Status,
		Comment:     req.Comment,
		StepResults: req.StepResults,
		IsFlaky:     req.IsFlaky,
	}

	result, err := s.repo.UpdateTestResult(ctx, resultID, input)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("результат не найден")
		}
		return nil, apperrors.NewInternal("ошибка обновления результата")
	}

	// Attach files
	if len(req.FileIDs) > 0 {
		fileUUIDs, err := parseUUIDs(req.FileIDs)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный ID файла")
		}
		if err := s.repo.AddFilesToResult(ctx, resultID, fileUUIDs); err != nil {
			return nil, apperrors.NewInternal("ошибка привязки файлов к результату")
		}
	}

	// Recompute run stats
	if err := s.recomputeRunStats(ctx, runID); err != nil {
		return nil, err
	}

	_ = result
	// Reload result with full edges
	updated, err := s.repo.GetTestResult(ctx, resultID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения обновлённого результата")
	}

	return toTestResultResponse(updated), nil
}

// AssignTestersToRun assigns testers to a test run.
func (s *TMSService) AssignTestersToRun(ctx context.Context, runID uuid.UUID, req AssignTestersRequest) (*TestRunDetailResponse, error) {
	userIDs, err := parseUUIDs(req.UserIDs)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный ID пользователя")
	}

	if err := s.repo.AssignTestersToRun(ctx, runID, userIDs); err != nil {
		return nil, apperrors.NewInternal("ошибка назначения тестировщиков")
	}

	return s.GetTestRunDetail(ctx, runID)
}

// GetTestRunDetail returns full run information with results and assignees.
func (s *TMSService) GetTestRunDetail(ctx context.Context, runID uuid.UUID) (*TestRunDetailResponse, error) {
	run, err := s.repo.GetTestRunWithResults(ctx, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона с результатами")
	}

	return toTestRunDetailResponse(run), nil
}

// StartTestRun transitions a run from created to in_progress.
func (s *TMSService) StartTestRun(ctx context.Context, runID uuid.UUID) (*TestRunResponse, error) {
	return s.UpdateRunStatus(ctx, runID, "in_progress")
}

// CompleteTestRun transitions a run from in_progress to completed.
func (s *TMSService) CompleteTestRun(ctx context.Context, runID uuid.UUID) (*TestRunResponse, error) {
	return s.UpdateRunStatus(ctx, runID, "completed")
}

// recomputeRunStats recomputes run statistics based on current results.
func (s *TMSService) recomputeRunStats(ctx context.Context, runID uuid.UUID) error {
	results, err := s.repo.ListTestResults(ctx, runID)
	if err != nil {
		return apperrors.NewInternal("ошибка получения результатов для пересчёта статистики")
	}

	stats := TestRunStats{TotalCount: len(results)}
	for _, r := range results {
		switch string(r.Status) {
		case "passed":
			stats.PassedCount++
		case "failed":
			stats.FailedCount++
		case "broken":
			stats.BrokenCount++
		case "skipped":
			stats.SkippedCount++
		case "unknown":
			stats.UnknownCount++
		}
	}

	_, err = s.repo.UpdateTestRunStats(ctx, runID, stats)
	if err != nil {
		return apperrors.NewInternal("ошибка обновления статистики прогона")
	}

	return nil
}
