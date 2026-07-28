package tms

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/platform/observability"
)

// ---------------------------------------------------------------------------
// Прогоны тестов
// ---------------------------------------------------------------------------

// CreateTestRun создаёт новый прогон тестов.
func (s *TMSService) CreateTestRun(ctx context.Context, projectID, userID uuid.UUID, req CreateTestRunRequest) (*TestRunResponse, error) {
	return observability.Timed(ctx, "tms.CreateTestRun", func(ctx context.Context) (*TestRunResponse, error) {
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
	})
}

// GetTestRun возвращает прогон по ID в рамках projectID (A3-012).
func (s *TMSService) GetTestRun(ctx context.Context, projectID, runID uuid.UUID) (*TestRunResponse, error) {
	run, err := s.repo.GetTestRunInProject(ctx, projectID, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}
	return toTestRunResponse(run), nil
}

// ListTestRuns возвращает пагинированный список прогонов проекта.
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

// UpdateRunStatus обновляет статус прогона с валидацией переходов конечного автомата.
func (s *TMSService) UpdateRunStatus(ctx context.Context, projectID, runID uuid.UUID, newStatus string) (*TestRunResponse, error) {
	run, err := s.repo.GetTestRunInProject(ctx, projectID, runID)
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

// ReopenTestRun переоткрывает завершённый или прерванный прогон.
func (s *TMSService) ReopenTestRun(ctx context.Context, projectID, runID uuid.UUID) (*TestRunResponse, error) {
	run, err := s.repo.GetTestRunInProject(ctx, projectID, runID)
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
// Результаты тестов
// ---------------------------------------------------------------------------

// CreateTestResult создаёт результат теста в прогоне (прогон должен принадлежать projectID).
func (s *TMSService) CreateTestResult(ctx context.Context, projectID, runID uuid.UUID, req CreateTestResultRequest) (*TestResultResponse, error) {
	if _, err := s.repo.GetTestRunInProject(ctx, projectID, runID); err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}

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
		n, err := s.repo.CountTestCasesInProject(ctx, projectID, []uuid.UUID{id})
		if err != nil {
			return nil, apperrors.NewInternal("ошибка проверки тест-кейса")
		}
		if n != 1 {
			return nil, apperrors.NewNotFound("тест-кейс не найден")
		}
		input.TestCaseID = &id
	}

	result, err := s.repo.CreateTestResult(ctx, runID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания результата")
	}

	return toTestResultResponse(result), nil
}

// UpdateTestResult обновляет результат теста (требует принадлежности к run через Get).
func (s *TMSService) UpdateTestResult(ctx context.Context, resultID uuid.UUID, req UpdateTestResultRequest) (*TestResultResponse, error) {
	existing, err := s.repo.GetTestResult(ctx, resultID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("результат не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения результата")
	}
	if existing.Edges.Run == nil {
		return nil, apperrors.NewNotFound("результат не найден")
	}

	input := UpdateTestResultInput{
		Status:      req.Status,
		Comment:     req.Comment,
		StepResults: req.StepResults,
		IsFlaky:     req.IsFlaky,
	}

	result, err := s.repo.UpdateTestResult(ctx, existing.Edges.Run.ID, resultID, input)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("результат не найден")
		}
		return nil, apperrors.NewInternal("ошибка обновления результата")
	}

	return toTestResultResponse(result), nil
}

// ListTestResults возвращает все результаты прогона в рамках projectID (A3-012).
func (s *TMSService) ListTestResults(ctx context.Context, projectID, runID uuid.UUID) ([]TestResultResponse, error) {
	if _, err := s.repo.GetTestRunInProject(ctx, projectID, runID); err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}

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
// Расширенные операции прогона
// ---------------------------------------------------------------------------

// AddTestCasesToRun добавляет тест-кейсы в прогон, создавая записи TestResult со статусом unknown.
func (s *TMSService) AddTestCasesToRun(ctx context.Context, projectID, runID uuid.UUID, req AddTestCasesToRunRequest) (*TestRunDetailResponse, error) {
	if _, err := s.repo.GetTestRunInProject(ctx, projectID, runID); err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}

	caseIDs, err := parseUUIDs(req.CaseIDs)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный ID тест-кейса")
	}
	if len(caseIDs) > 0 {
		n, err := s.repo.CountTestCasesInProject(ctx, projectID, caseIDs)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка проверки тест-кейсов")
		}
		if n != len(caseIDs) {
			return nil, apperrors.NewNotFound("тест-кейс не найден")
		}
	}

	if err := s.repo.AddTestCasesToRun(ctx, runID, caseIDs); err != nil {
		return nil, apperrors.NewInternal("ошибка добавления тест-кейсов в прогон")
	}

	// Пересчитываем статистику прогона
	if err := s.recomputeRunStats(ctx, runID); err != nil {
		return nil, err
	}

	return s.GetTestRunDetail(ctx, projectID, runID)
}

// ExecuteTestResult обновляет результат теста с валидацией статуса прогона.
func (s *TMSService) ExecuteTestResult(ctx context.Context, projectID, runID, resultID, userID uuid.UUID, req UpdateTestResultRequest) (*TestResultResponse, error) {
	// Проверяем что прогон принадлежит проекту и в статусе in_progress (A3-012).
	run, err := s.repo.GetTestRunInProject(ctx, projectID, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}
	if string(run.Status) != "in_progress" {
		return nil, apperrors.NewBadRequest("выполнять результаты можно только в прогоне со статусом in_progress")
	}

	// Обновляем результат
	input := UpdateTestResultInput{
		Status:       req.Status,
		Comment:      req.Comment,
		StepResults:  req.StepResults,
		IsFlaky:      req.IsFlaky,
		ExecutedByID: &userID, // A3-017: audit who executed
	}

	if _, err := s.repo.UpdateTestResult(ctx, runID, resultID, input); err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("результат не найден")
		}
		return nil, apperrors.NewInternal("ошибка обновления результата")
	}

	// Привязываем файлы к результату
	if len(req.FileIDs) > 0 {
		fileUUIDs, err := parseUUIDs(req.FileIDs)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный ID файла")
		}
		if err := s.repo.AddFilesToResult(ctx, runID, resultID, fileUUIDs); err != nil {
			return nil, apperrors.NewInternal("ошибка привязки файлов к результату")
		}
	}

	// Пересчитываем статистику прогона
	if err := s.recomputeRunStats(ctx, runID); err != nil {
		return nil, err
	}

	// Перезагружаем результат с полными связями (Update не возвращает edges).
	updated, err := s.repo.GetTestResult(ctx, resultID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения обновлённого результата")
	}

	return toTestResultResponse(updated), nil
}

// AssignTestersToRun назначает тестировщиков на прогон.
func (s *TMSService) AssignTestersToRun(ctx context.Context, projectID, runID uuid.UUID, req AssignTestersRequest) (*TestRunDetailResponse, error) {
	if _, err := s.repo.GetTestRunInProject(ctx, projectID, runID); err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}

	userIDs, err := parseUUIDs(req.UserIDs)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный ID пользователя")
	}

	if err := s.repo.AssignTestersToRun(ctx, runID, userIDs); err != nil {
		return nil, apperrors.NewInternal("ошибка назначения тестировщиков")
	}

	return s.GetTestRunDetail(ctx, projectID, runID)
}

// GetTestRunDetail возвращает полную информацию о прогоне с результатами и назначенными.
func (s *TMSService) GetTestRunDetail(ctx context.Context, projectID, runID uuid.UUID) (*TestRunDetailResponse, error) {
	run, err := s.repo.GetTestRunWithResultsInProject(ctx, projectID, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона с результатами")
	}

	return toTestRunDetailResponse(run), nil
}

// StartTestRun переводит прогон из created в in_progress.
func (s *TMSService) StartTestRun(ctx context.Context, projectID, runID uuid.UUID) (*TestRunResponse, error) {
	return s.UpdateRunStatus(ctx, projectID, runID, "in_progress")
}

// CompleteTestRun переводит прогон из in_progress в completed и публикует NATS-событие.
func (s *TMSService) CompleteTestRun(ctx context.Context, projectID, runID uuid.UUID) (*TestRunResponse, error) {
	resp, err := s.UpdateRunStatus(ctx, projectID, runID, "completed")
	if err != nil {
		return nil, err
	}

	// Публикуем событие завершения прогона для уведомлений.
	s.nats.PublishJSON(fmt.Sprintf(SubjectTestRunCompleted, runID.String()), map[string]interface{}{
		"run_id": runID,
		"title":  resp.Title,
		"total":  resp.TotalCount,
		"passed": resp.PassedCount,
	})

	return resp, nil
}

// recomputeRunStats пересчитывает статистику прогона по текущим результатам.
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
