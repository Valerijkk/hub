package tms

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Отчёты
// ---------------------------------------------------------------------------

// GetCoverageReport возвращает отчёт о покрытии задач тест-кейсами.
func (s *TMSService) GetCoverageReport(ctx context.Context, projectID uuid.UUID) (*CoverageReport, error) {
	issues, err := s.repo.GetProjectIssuesWithTestCases(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения задач для покрытия")
	}

	report := &CoverageReport{
		Issues: make([]IssueCoverage, 0),
	}

	for _, issue := range issues {
		if len(issue.Edges.LinkedTestCases) == 0 {
			continue
		}

		cov := IssueCoverage{
			IssueID:    issue.ID.String(),
			IssueKey:   issue.IssueKey,
			IssueTitle: issue.Title,
			TestCases:  make([]TestCaseCoverage, 0, len(issue.Edges.LinkedTestCases)),
		}

		for _, tc := range issue.Edges.LinkedTestCases {
			lastStatus := "no_results"
			lastResult, err := s.repo.GetLastResultForTestCase(ctx, tc.ID)
			if err == nil && lastResult != nil {
				lastStatus = string(lastResult.Status)
			}

			cov.TestCases = append(cov.TestCases, TestCaseCoverage{
				CaseKey:    tc.CaseKey,
				CaseTitle:  tc.Title,
				LastStatus: lastStatus,
			})
		}

		report.Issues = append(report.Issues, cov)
	}

	return report, nil
}

// GetTestTrends возвращает тренды прогонов (pass/fail/skip) за последние N прогонов.
func (s *TMSService) GetTestTrends(ctx context.Context, projectID uuid.UUID, lastN int) ([]TrendPoint, error) {
	if lastN <= 0 || lastN > 100 {
		lastN = 20
	}

	params := ListTestRunsParams{Offset: 0, Limit: lastN}
	runs, _, err := s.repo.ListTestRuns(ctx, projectID, params)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения прогонов для трендов")
	}

	trends := make([]TrendPoint, 0, len(runs))
	// Обратный порядок (от старых к новым для графиков)
	for i := len(runs) - 1; i >= 0; i-- {
		run := runs[i]
		passRate := float64(0)
		if run.TotalCount > 0 {
			passRate = float64(run.PassedCount) / float64(run.TotalCount) * 100
		}
		trends = append(trends, TrendPoint{
			RunID:     run.ID.String(),
			Title:     run.Title,
			Total:     run.TotalCount,
			Passed:    run.PassedCount,
			Failed:    run.FailedCount,
			Skipped:   run.SkippedCount,
			Broken:    run.BrokenCount,
			PassRate:  passRate,
			CreatedAt: run.CreatedAt,
		})
	}

	return trends, nil
}

// GetRunComparison сравнивает два прогона, показывая изменения статусов.
func (s *TMSService) GetRunComparison(ctx context.Context, projectID, runID1, runID2 uuid.UUID) (*RunComparison, error) {
	// Оба прогона должны принадлежать projectID (W4-run-export / A3-012).
	run1, err := s.repo.GetTestRunInProject(ctx, projectID, runID1)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон 1 не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона 1")
	}
	run2, err := s.repo.GetTestRunInProject(ctx, projectID, runID2)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон 2 не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона 2")
	}

	// Получаем результаты обоих прогонов
	results1, err := s.repo.ListTestResults(ctx, runID1)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения результатов прогона 1")
	}
	results2, err := s.repo.ListTestResults(ctx, runID2)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения результатов прогона 2")
	}

	// Строим карту: caseKey -> статус для каждого прогона
	map1 := make(map[string]statusInfo)
	for _, r := range results1 {
		if r.Edges.TestCase != nil {
			map1[r.Edges.TestCase.CaseKey] = statusInfo{
				status: string(r.Status),
				title:  r.Edges.TestCase.Title,
			}
		}
	}
	map2 := make(map[string]statusInfo)
	for _, r := range results2 {
		if r.Edges.TestCase != nil {
			map2[r.Edges.TestCase.CaseKey] = statusInfo{
				status: string(r.Status),
				title:  r.Edges.TestCase.Title,
			}
		}
	}

	// Собираем изменения
	changes := make([]ResultChange, 0)
	allKeys := make(map[string]bool)
	for k := range map1 {
		allKeys[k] = true
	}
	for k := range map2 {
		allKeys[k] = true
	}

	for key := range allKeys {
		s1 := map1[key]
		s2 := map2[key]
		if s1.status != s2.status {
			title := s1.title
			if title == "" {
				title = s2.title
			}
			changes = append(changes, ResultChange{
				CaseKey:    key,
				CaseTitle:  title,
				StatusRun1: s1.status,
				StatusRun2: s2.status,
			})
		}
	}

	passRate1 := float64(0)
	if run1.TotalCount > 0 {
		passRate1 = float64(run1.PassedCount) / float64(run1.TotalCount) * 100
	}
	passRate2 := float64(0)
	if run2.TotalCount > 0 {
		passRate2 = float64(run2.PassedCount) / float64(run2.TotalCount) * 100
	}

	return &RunComparison{
		Run1: TestRunSummary{
			RunID:     run1.ID.String(),
			Title:     run1.Title,
			Total:     run1.TotalCount,
			Passed:    run1.PassedCount,
			Failed:    run1.FailedCount,
			PassRate:  passRate1,
			CreatedAt: run1.CreatedAt,
		},
		Run2: TestRunSummary{
			RunID:     run2.ID.String(),
			Title:     run2.Title,
			Total:     run2.TotalCount,
			Passed:    run2.PassedCount,
			Failed:    run2.FailedCount,
			PassRate:  passRate2,
			CreatedAt: run2.CreatedAt,
		},
		Changes: changes,
	}, nil
}

// statusInfo — вспомогательная структура для сравнения прогонов.
type statusInfo struct {
	status string
	title  string
}

// GetTestRunHistory возвращает историю последних N прогонов проекта.
func (s *TMSService) GetTestRunHistory(ctx context.Context, projectID uuid.UUID, limit int) ([]TestRunSummary, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}
	params := ListTestRunsParams{
		Offset: 0,
		Limit:  limit,
	}
	runs, _, err := s.repo.ListTestRuns(ctx, projectID, params)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения истории прогонов")
	}

	summaries := make([]TestRunSummary, 0, len(runs))
	for _, run := range runs {
		passRate := float64(0)
		if run.TotalCount > 0 {
			passRate = float64(run.PassedCount) / float64(run.TotalCount) * 100
		}
		summaries = append(summaries, TestRunSummary{
			RunID:     run.ID.String(),
			Title:     run.Title,
			Total:     run.TotalCount,
			Passed:    run.PassedCount,
			Failed:    run.FailedCount,
			PassRate:  passRate,
			CreatedAt: run.CreatedAt,
		})
	}
	return summaries, nil
}

// GetFlakyTests возвращает тест-кейсы с нестабильными (flaky) результатами из последних 5 прогонов.
func (s *TMSService) GetFlakyTests(ctx context.Context, projectID uuid.UUID) (*FlakyTestReport, error) {
	const runWindow = 5
	params := ListTestRunsParams{Offset: 0, Limit: runWindow}
	runs, _, err := s.repo.ListTestRuns(ctx, projectID, params)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения прогонов для flaky-анализа")
	}

	type tcMeta struct {
		statuses    []string
		markedFlaky bool
	}

	tcStatuses := make(map[uuid.UUID]*tcMeta)
	tcMap := make(map[uuid.UUID]*ent.TestCase)
	for _, run := range runs {
		results, err := s.repo.ListTestResults(ctx, run.ID)
		if err != nil {
			continue
		}
		for _, r := range results {
			if r.Edges.TestCase == nil {
				continue
			}
			tcID := r.Edges.TestCase.ID
			meta, ok := tcStatuses[tcID]
			if !ok {
				meta = &tcMeta{}
				tcStatuses[tcID] = meta
			}
			meta.statuses = append(meta.statuses, string(r.Status))
			if r.IsFlaky {
				meta.markedFlaky = true
			}
			if _, ok := tcMap[tcID]; !ok {
				tcMap[tcID] = r.Edges.TestCase
			}
		}
	}

	items := make([]FlakyTestItem, 0)
	for tcID, meta := range tcStatuses {
		if len(meta.statuses) < 2 && !meta.markedFlaky {
			continue
		}
		isFlaky := meta.markedFlaky
		if !isFlaky && len(meta.statuses) >= 2 {
			first := meta.statuses[0]
			for _, st := range meta.statuses[1:] {
				if st != first {
					isFlaky = true
					break
				}
			}
		}
		if !isFlaky {
			continue
		}
		tc, ok := tcMap[tcID]
		if !ok {
			continue
		}
		passed := 0
		for _, st := range meta.statuses {
			if st == "passed" {
				passed++
			}
		}
		flakeRate := 0.0
		if n := len(meta.statuses); n > 0 {
			flakeRate = float64(n-passed) / float64(n)
		}
		base := toTestCaseResponse(tc)
		items = append(items, FlakyTestItem{
			TestCaseResponse: *base,
			StatusHistory:    meta.statuses,
			FlakeRate:        flakeRate,
			MarkedFlaky:      meta.markedFlaky,
		})
	}

	return &FlakyTestReport{
		Items:     items,
		RunWindow: runWindow,
	}, nil
}

// GetTraceabilityMatrix строит матрицу трассировки: строки = задачи, столбцы = тест-кейсы, ячейки = последний статус.
func (s *TMSService) GetTraceabilityMatrix(ctx context.Context, projectID uuid.UUID) (*TraceabilityMatrix, error) {
	issues, err := s.repo.GetProjectIssuesWithTestCases(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения задач для матрицы трассировки")
	}

	rows := make([]TraceabilityRow, 0, len(issues))
	for _, iss := range issues {
		if iss.Edges.LinkedTestCases == nil || len(iss.Edges.LinkedTestCases) == 0 {
			continue
		}
		row := TraceabilityRow{
			IssueKey:   iss.IssueKey,
			IssueTitle: iss.Title,
		}
		for _, tc := range iss.Edges.LinkedTestCases {
			lastStatus := "no_results"
			lastResult, err := s.repo.GetLastResultForTestCase(ctx, tc.ID)
			if err == nil && lastResult != nil {
				lastStatus = string(lastResult.Status)
			}
			row.TestCases = append(row.TestCases, TraceabilityCell{
				CaseKey:    tc.CaseKey,
				CaseTitle:  tc.Title,
				LastStatus: lastStatus,
			})
		}
		rows = append(rows, row)
	}

	return &TraceabilityMatrix{Rows: rows}, nil
}

// GetDefectCategories группирует упавшие тесты из последних прогонов по паттерну сообщения об ошибке.
func (s *TMSService) GetDefectCategories(ctx context.Context, projectID uuid.UUID) ([]DefectCategory, error) {
	params := ListTestRunsParams{Offset: 0, Limit: 10}
	runs, _, err := s.repo.ListTestRuns(ctx, projectID, params)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения прогонов для анализа дефектов")
	}

	// Собираем упавшие результаты и группируем по паттерну ошибки
	categories := make(map[string][]string) // паттерн -> []caseKey
	for _, run := range runs {
		results, err := s.repo.ListTestResults(ctx, run.ID)
		if err != nil {
			continue
		}
		for _, r := range results {
			if string(r.Status) != "failed" && string(r.Status) != "broken" {
				continue
			}
			pattern := "unknown"
			if r.ErrorMessage != "" {
				pattern = normalizeErrorPattern(r.ErrorMessage)
			}
			caseKey := ""
			if r.Edges.TestCase != nil {
				caseKey = r.Edges.TestCase.CaseKey
			}
			if caseKey != "" {
				categories[pattern] = append(categories[pattern], caseKey)
			}
		}
	}

	result := make([]DefectCategory, 0, len(categories))
	for pattern, keys := range categories {
		// Дедупликация ключей
		seen := make(map[string]bool)
		unique := make([]string, 0)
		for _, k := range keys {
			if !seen[k] {
				seen[k] = true
				unique = append(unique, k)
			}
		}
		result = append(result, DefectCategory{
			Pattern:  pattern,
			Count:    len(unique),
			CaseKeys: unique,
		})
	}

	return result, nil
}

// CompareEnvironments сравнивает результаты тестов между двумя окружениями.
// Для каждого окружения получает последний прогон и его результаты, затем объединяет по ключу тест-кейса.
func (s *TMSService) CompareEnvironments(ctx context.Context, projectID uuid.UUID, envID1, envID2 uuid.UUID) (*EnvironmentComparisonResponse, error) {
	env1, err := s.repo.GetEnvironment(ctx, envID1)
	if err != nil {
		return nil, apperrors.NewNotFound("окружение 1 не найдено")
	}
	env2, err := s.repo.GetEnvironment(ctx, envID2)
	if err != nil {
		return nil, apperrors.NewNotFound("окружение 2 не найдено")
	}

	// Вспомогательная функция: получить последние результаты для окружения
	getLatestResults := func(envID uuid.UUID) (map[string]string, map[string]string, error) {
		run, err := s.repo.GetLatestTestRunForEnvironment(ctx, projectID, envID)
		if err != nil || run == nil {
			return make(map[string]string), make(map[string]string), nil
		}

		results, err := s.repo.ListTestResults(ctx, run.ID)
		if err != nil {
			return nil, nil, err
		}

		statusMap := make(map[string]string)  // case_key -> статус
		titleMap := make(map[string]string)    // case_key -> заголовок
		for _, r := range results {
			if r.Edges.TestCase != nil {
				statusMap[r.Edges.TestCase.CaseKey] = string(r.Status)
				titleMap[r.Edges.TestCase.CaseKey] = r.Edges.TestCase.Title
			}
		}
		return statusMap, titleMap, nil
	}

	status1, titles1, err := getLatestResults(envID1)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения результатов окружения 1")
	}
	status2, titles2, err := getLatestResults(envID2)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения результатов окружения 2")
	}

	// Объединяем все ключи тест-кейсов
	allKeys := make(map[string]bool)
	for k := range status1 {
		allKeys[k] = true
	}
	for k := range status2 {
		allKeys[k] = true
	}

	items := make([]EnvironmentComparisonItem, 0, len(allKeys))
	for key := range allKeys {
		title := titles1[key]
		if title == "" {
			title = titles2[key]
		}
		s1 := status1[key]
		if s1 == "" {
			s1 = "no_results"
		}
		s2 := status2[key]
		if s2 == "" {
			s2 = "no_results"
		}
		items = append(items, EnvironmentComparisonItem{
			CaseKey:    key,
			CaseTitle:  title,
			StatusEnv1: s1,
			StatusEnv2: s2,
		})
	}

	return &EnvironmentComparisonResponse{
		Env1:  *toEnvironmentResponse(env1),
		Env2:  *toEnvironmentResponse(env2),
		Items: items,
	}, nil
}
