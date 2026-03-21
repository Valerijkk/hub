package tms

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/schema"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	natsclient "github.com/valerii/hub/backend/internal/platform/nats"
)

// TMSService реализует бизнес-логику модуля управления тестированием.
// Оркестрирует вызовы репозитория, версионирование, стейт-машину прогонов,
// импорт результатов Allure и публикацию NATS-событий.
type TMSService struct {
	repo    Repository
	nats    *natsclient.Client
	matcher *Matcher
}

// NewTMSService создаёт новый сервис TMS.
func NewTMSService(repo Repository, nats *natsclient.Client) *TMSService {
	return &TMSService{
		repo:    repo,
		nats:    nats,
		matcher: NewMatcher(repo),
	}
}

// ---------------------------------------------------------------------------
// Тест-кейсы
// ---------------------------------------------------------------------------

// CreateTestCase создаёт новый тест-кейс в проекте.
func (s *TMSService) CreateTestCase(ctx context.Context, projectID, userID uuid.UUID, req CreateTestCaseRequest) (*TestCaseResponse, error) {
	input := CreateTestCaseInput{
		Title:         req.Title,
		Description:   req.Description,
		Precondition:  req.Precondition,
		Postcondition: req.Postcondition,
		CaseType:      req.CaseType,
		Priority:      req.Priority,
		Status:        req.Status,
		Steps:         req.Steps,
		FullName:      req.FullName,
		ClassName:     req.ClassName,
		MethodName:    req.MethodName,
		ExternalID:    req.ExternalID,
		CreatedByID:   userID,
	}

	if req.SuiteID != "" {
		id, err := uuid.Parse(req.SuiteID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный suite_id")
		}
		input.SuiteID = &id
	}
	if len(req.Labels) > 0 {
		ids, err := parseUUIDs(req.Labels)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный label ID")
		}
		input.LabelIDs = ids
	}
	if len(req.Components) > 0 {
		ids, err := parseUUIDs(req.Components)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный component ID")
		}
		input.ComponentIDs = ids
	}

	tc, err := s.repo.CreateTestCase(ctx, projectID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания тест-кейса")
	}

	return toTestCaseResponse(tc), nil
}

// GetTestCase возвращает тест-кейс по ключу.
func (s *TMSService) GetTestCase(ctx context.Context, caseKey string) (*TestCaseResponse, error) {
	tc, err := s.repo.FindTestCaseByKey(ctx, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("тест-кейс не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения тест-кейса")
	}
	return toTestCaseResponse(tc), nil
}

// ListTestCases возвращает пагинированный список тест-кейсов.
func (s *TMSService) ListTestCases(ctx context.Context, projectID uuid.UUID, req ListTestCasesRequest) (*PaginatedResponse[TestCaseResponse], error) {
	params := ListTestCasesParams{
		Offset:   req.Offset,
		Limit:    req.Limit,
		Status:   req.Status,
		Priority: req.Priority,
		Search:   req.Search,
	}
	if req.SuiteID != "" {
		id, err := uuid.Parse(req.SuiteID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный suite_id")
		}
		params.SuiteID = &id
	}

	items, total, err := s.repo.ListTestCases(ctx, projectID, params)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения списка тест-кейсов")
	}

	limit := req.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	responses := make([]TestCaseResponse, 0, len(items))
	for _, item := range items {
		responses = append(responses, *toTestCaseResponse(item))
	}

	return &PaginatedResponse[TestCaseResponse]{
		Items:  responses,
		Total:  total,
		Offset: req.Offset,
		Limit:  limit,
	}, nil
}

// UpdateTestCase обновляет тест-кейс, предварительно создавая снимок версии.
func (s *TMSService) UpdateTestCase(ctx context.Context, caseKey string, userID uuid.UUID, req UpdateTestCaseRequest) (*TestCaseResponse, error) {
	existing, err := s.repo.FindTestCaseByKey(ctx, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("тест-кейс не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения тест-кейса")
	}

	// Создаём снимок текущей версии перед обновлением
	snapshot := buildSnapshot(existing)
	_, err = s.repo.CreateTestCaseVersion(ctx, existing.ID, userID, existing.Version, snapshot, "")
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания версии")
	}

	// Формируем input для обновления
	input := UpdateTestCaseInput{
		Title:         req.Title,
		Description:   req.Description,
		Precondition:  req.Precondition,
		Postcondition: req.Postcondition,
		CaseType:      req.CaseType,
		Priority:      req.Priority,
		Status:        req.Status,
		Steps:         req.Steps,
		FullName:      req.FullName,
		ClassName:     req.ClassName,
		MethodName:    req.MethodName,
		ExternalID:    req.ExternalID,
		Version:       req.Version,
	}

	if req.SuiteID != nil {
		if *req.SuiteID == "" {
			input.ClearSuite = true
		} else {
			id, err := uuid.Parse(*req.SuiteID)
			if err != nil {
				return nil, apperrors.NewBadRequest("некорректный suite_id")
			}
			input.SuiteID = &id
		}
	}
	if req.Labels != nil {
		ids, err := parseUUIDs(*req.Labels)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный label ID")
		}
		input.LabelIDs = &ids
	}
	if req.Components != nil {
		ids, err := parseUUIDs(*req.Components)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный component ID")
		}
		input.ComponentIDs = &ids
	}

	updated, err := s.repo.UpdateTestCase(ctx, existing.ID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка обновления тест-кейса")
	}

	return toTestCaseResponse(updated), nil
}

// DeleteTestCase удаляет тест-кейс по ключу.
func (s *TMSService) DeleteTestCase(ctx context.Context, caseKey string) error {
	existing, err := s.repo.FindTestCaseByKey(ctx, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("тест-кейс не найден")
		}
		return apperrors.NewInternal("ошибка получения тест-кейса")
	}
	return s.repo.DeleteTestCase(ctx, existing.ID)
}

// RollbackTestCase откатывает тест-кейс к указанной версии.
func (s *TMSService) RollbackTestCase(ctx context.Context, caseKey string, versionID, userID uuid.UUID) (*TestCaseResponse, error) {
	existing, err := s.repo.FindTestCaseByKey(ctx, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("тест-кейс не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения тест-кейса")
	}

	version, err := s.repo.GetTestCaseVersion(ctx, versionID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("версия не найдена")
		}
		return nil, apperrors.NewInternal("ошибка получения версии")
	}

	// Создаём снимок текущего состояния перед откатом
	currentSnapshot := buildSnapshot(existing)
	_, err = s.repo.CreateTestCaseVersion(ctx, existing.ID, userID, existing.Version, currentSnapshot, "перед откатом к версии")
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания снимка перед откатом")
	}

	// Применяем snapshot как обновление
	input := snapshotToUpdateInput(version.Snapshot)
	updated, err := s.repo.UpdateTestCase(ctx, existing.ID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка применения отката")
	}

	return toTestCaseResponse(updated), nil
}

// ---------------------------------------------------------------------------
// Версии
// ---------------------------------------------------------------------------

// ListTestCaseVersions возвращает все версии тест-кейса.
func (s *TMSService) ListTestCaseVersions(ctx context.Context, caseKey string) ([]TestCaseVersionResponse, error) {
	tc, err := s.repo.FindTestCaseByKey(ctx, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("тест-кейс не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения тест-кейса")
	}

	versions, err := s.repo.ListTestCaseVersions(ctx, tc.ID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения версий")
	}

	result := make([]TestCaseVersionResponse, 0, len(versions))
	for _, v := range versions {
		result = append(result, *toTestCaseVersionResponse(v))
	}
	return result, nil
}

// ---------------------------------------------------------------------------
// Привязка к задачам
// ---------------------------------------------------------------------------

// LinkTestCaseToIssue привязывает тест-кейс к задаче.
func (s *TMSService) LinkTestCaseToIssue(ctx context.Context, caseKey string, issueID uuid.UUID) error {
	tc, err := s.repo.FindTestCaseByKey(ctx, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("тест-кейс не найден")
		}
		return apperrors.NewInternal("ошибка получения тест-кейса")
	}
	return s.repo.LinkTestCaseToIssue(ctx, tc.ID, issueID)
}

// UnlinkTestCaseFromIssue отвязывает тест-кейс от задачи.
func (s *TMSService) UnlinkTestCaseFromIssue(ctx context.Context, caseKey string, issueID uuid.UUID) error {
	tc, err := s.repo.FindTestCaseByKey(ctx, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("тест-кейс не найден")
		}
		return apperrors.NewInternal("ошибка получения тест-кейса")
	}
	return s.repo.UnlinkTestCaseFromIssue(ctx, tc.ID, issueID)
}

// GetTestCaseLinks возвращает задачи, привязанные к тест-кейсу.
func (s *TMSService) GetTestCaseLinks(ctx context.Context, caseKey string) ([]IssueBriefResponse, error) {
	tc, err := s.repo.FindTestCaseByKey(ctx, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("тест-кейс не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения тест-кейса")
	}

	issues, err := s.repo.GetTestCaseLinkedIssues(ctx, tc.ID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения привязанных задач")
	}

	result := make([]IssueBriefResponse, 0, len(issues))
	for _, i := range issues {
		result = append(result, IssueBriefResponse{
			ID:       i.ID.String(),
			IssueKey: i.IssueKey,
			Title:    i.Title,
		})
	}
	return result, nil
}

// CreateBugFromFailedResult создаёт баг из упавшего результата теста.
// Создаёт задачу типа bug через shared Ent client (без импорта tracker пакета),
// привязывает к тест-кейсу и возвращает краткую информацию о созданном баге.
func (s *TMSService) CreateBugFromFailedResult(ctx context.Context, projectID, userID uuid.UUID, resultID uuid.UUID, req CreateBugRequest) (*IssueBriefResponse, error) {
	// Получаем результат с тест-кейсом
	result, err := s.repo.GetTestResult(ctx, resultID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("результат не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения результата")
	}

	// Определяем тест-кейс
	var testCaseID uuid.UUID
	title := req.Title
	if result.Edges.TestCase != nil {
		testCaseID = result.Edges.TestCase.ID
		if title == "" {
			title = "FAILED: " + result.Edges.TestCase.Title
		}
	} else {
		if title == "" {
			title = "FAILED: тест без привязки (результат " + resultID.String()[:8] + ")"
		}
	}

	description := req.Description
	if description == "" && result.ErrorMessage != "" {
		description = result.ErrorMessage
	}

	issue, err := s.repo.CreateBugIssue(ctx, projectID, userID, testCaseID, title, description, req.Priority)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания бага: " + err.Error())
	}

	return &IssueBriefResponse{
		ID:       issue.ID.String(),
		IssueKey: issue.IssueKey,
		Title:    issue.Title,
	}, nil
}

// GetCoverageReport возвращает отчёт о покрытии задач тестами.
// Для каждой задачи проекта, имеющей привязанные тест-кейсы,
// показывает последний статус результата каждого тест-кейса.
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
	// Реверсируем порядок (от старых к новым для графика)
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

// GetRunComparison сравнивает два прогона, показывая изменения статусов тест-кейсов.
func (s *TMSService) GetRunComparison(ctx context.Context, runID1, runID2 uuid.UUID) (*RunComparison, error) {
	// Получаем оба прогона
	run1, err := s.repo.GetTestRun(ctx, runID1)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон 1 не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона 1")
	}
	run2, err := s.repo.GetTestRun(ctx, runID2)
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

	// Строим карту: caseKey -> status для каждого прогона
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

// ---------------------------------------------------------------------------
// Папки/сьюты
// ---------------------------------------------------------------------------

// CreateSuite создаёт новую папку тест-кейсов.
func (s *TMSService) CreateSuite(ctx context.Context, projectID uuid.UUID, req CreateSuiteRequest) (*SuiteResponse, error) {
	input := CreateSuiteInput{
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}
	if req.ParentID != "" {
		id, err := uuid.Parse(req.ParentID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный parent_id")
		}
		input.ParentID = &id
	}

	suite, err := s.repo.CreateSuite(ctx, projectID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания папки")
	}

	return toSuiteResponse(suite), nil
}

// ListSuites возвращает дерево папок проекта.
func (s *TMSService) ListSuites(ctx context.Context, projectID uuid.UUID) ([]SuiteResponse, error) {
	suites, err := s.repo.ListSuites(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения папок")
	}

	// Строим дерево: возвращаем только корневые (без parent)
	return buildSuiteTree(suites), nil
}

// UpdateSuite обновляет папку.
func (s *TMSService) UpdateSuite(ctx context.Context, suiteID uuid.UUID, req UpdateSuiteRequest) (*SuiteResponse, error) {
	input := UpdateSuiteInput{
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}

	suite, err := s.repo.UpdateSuite(ctx, suiteID, input)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("папка не найдена")
		}
		return nil, apperrors.NewInternal("ошибка обновления папки")
	}

	return toSuiteResponse(suite), nil
}

// DeleteSuite удаляет папку.
func (s *TMSService) DeleteSuite(ctx context.Context, suiteID uuid.UUID) error {
	if err := s.repo.DeleteSuite(ctx, suiteID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("папка не найдена")
		}
		return apperrors.NewInternal("ошибка удаления папки")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Компоненты
// ---------------------------------------------------------------------------

// CreateComponent создаёт новый компонент.
func (s *TMSService) CreateComponent(ctx context.Context, projectID uuid.UUID, req CreateComponentRequest) (*ComponentResponse, error) {
	comp, err := s.repo.CreateComponent(ctx, projectID, req.Name, req.Description)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания компонента")
	}
	return toComponentResponse(comp), nil
}

// ListComponents возвращает все компоненты проекта.
func (s *TMSService) ListComponents(ctx context.Context, projectID uuid.UUID) ([]ComponentResponse, error) {
	comps, err := s.repo.ListComponents(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения компонентов")
	}
	result := make([]ComponentResponse, 0, len(comps))
	for _, c := range comps {
		result = append(result, *toComponentResponse(c))
	}
	return result, nil
}

// DeleteComponent удаляет компонент.
func (s *TMSService) DeleteComponent(ctx context.Context, componentID uuid.UUID) error {
	if err := s.repo.DeleteComponent(ctx, componentID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("компонент не найден")
		}
		return apperrors.NewInternal("ошибка удаления компонента")
	}
	return nil
}

// ---------------------------------------------------------------------------
// Прогоны
// ---------------------------------------------------------------------------

// CreateTestRun создаёт новый прогон тестов.
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

// GetTestRun возвращает прогон по ID.
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

// ListTestRuns возвращает пагинированный список прогонов.
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

// UpdateRunStatus обновляет статус прогона с проверкой стейт-машины.
// Допустимые переходы: created->in_progress, in_progress->completed, in_progress->aborted.
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

// ReopenTestRun переоткрывает завершённый прогон.
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
// Результаты
// ---------------------------------------------------------------------------

// CreateTestResult создаёт результат теста в прогоне.
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

// UpdateTestResult обновляет результат теста.
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

// ListTestResults возвращает все результаты прогона.
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
// Расширенные операции прогонов
// ---------------------------------------------------------------------------

// AddTestCasesToRun добавляет тест-кейсы в прогон, создавая записи TestResult со статусом unknown.
// Обновляет total_count и unknown_count прогона.
func (s *TMSService) AddTestCasesToRun(ctx context.Context, runID uuid.UUID, req AddTestCasesToRunRequest) (*TestRunDetailResponse, error) {
	caseIDs, err := parseUUIDs(req.CaseIDs)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный ID тест-кейса")
	}

	if err := s.repo.AddTestCasesToRun(ctx, runID, caseIDs); err != nil {
		return nil, apperrors.NewInternal("ошибка добавления тест-кейсов в прогон")
	}

	// Пересчитываем статистику прогона
	if err := s.recomputeRunStats(ctx, runID); err != nil {
		return nil, err
	}

	return s.GetTestRunDetail(ctx, runID)
}

// ExecuteTestResult обновляет результат теста с проверкой статуса прогона.
// Устанавливает статус, комментарий, результаты шагов, файлы-вложения и исполнителя.
func (s *TMSService) ExecuteTestResult(ctx context.Context, runID, resultID, userID uuid.UUID, req UpdateTestResultRequest) (*TestResultResponse, error) {
	// Проверяем что прогон в статусе in_progress
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

	// Обновляем результат
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

	// Привязываем файлы
	if len(req.FileIDs) > 0 {
		fileUUIDs, err := parseUUIDs(req.FileIDs)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный ID файла")
		}
		if err := s.repo.AddFilesToResult(ctx, resultID, fileUUIDs); err != nil {
			return nil, apperrors.NewInternal("ошибка привязки файлов к результату")
		}
	}

	// Пересчитываем статистику прогона
	if err := s.recomputeRunStats(ctx, runID); err != nil {
		return nil, err
	}

	_ = result
	// Перезагружаем результат с полными связями
	updated, err := s.repo.GetTestResult(ctx, resultID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения обновлённого результата")
	}

	return toTestResultResponse(updated), nil
}

// AssignTestersToRun назначает тестировщиков на прогон.
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

// GetTestRunDetail возвращает полную информацию о прогоне с результатами и назначенными.
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

// StartTestRun переводит прогон из статуса created в in_progress, устанавливая started_at.
func (s *TMSService) StartTestRun(ctx context.Context, runID uuid.UUID) (*TestRunResponse, error) {
	return s.UpdateRunStatus(ctx, runID, "in_progress")
}

// CompleteTestRun переводит прогон из in_progress в completed, устанавливая completed_at.
func (s *TMSService) CompleteTestRun(ctx context.Context, runID uuid.UUID) (*TestRunResponse, error) {
	return s.UpdateRunStatus(ctx, runID, "completed")
}

// recomputeRunStats пересчитывает статистику прогона на основе текущих результатов.
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

// ---------------------------------------------------------------------------
// Импорт Allure
// ---------------------------------------------------------------------------

// SubjectTestRunCompleted — NATS-тема для события завершения автоматического прогона.
const SubjectTestRunCompleted = "tms.run.completed.%s"

// ImportAllureResults импортирует результаты Allure в TMS.
// Создаёт автоматический прогон, сопоставляет результаты с тест-кейсами,
// сохраняет сырой JSON в MongoDB, структурированные результаты в PostgreSQL.
func (s *TMSService) ImportAllureResults(ctx context.Context, projectID, userID uuid.UUID, req ImportAllureRequest) (*TestRunDetailResponse, error) {
	// 1. Парсим allure-results JSON
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
		// Не прерываем — PG-результаты важнее
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

		// Длительность из start/stop
		if ar.Start != nil && ar.Stop != nil {
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

	// 7. Переводим прогон: created -> in_progress -> completed
	now := time.Now()
	_, err = s.repo.UpdateTestRunStatus(ctx, run.ID, "in_progress", &now)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка старта прогона")
	}
	_, err = s.repo.UpdateTestRunStatus(ctx, run.ID, "completed", &now)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка завершения прогона")
	}

	// 8. Публикуем NATS-событие завершения
	s.publishEvent(fmt.Sprintf(SubjectTestRunCompleted, run.ID.String()), map[string]interface{}{
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
	// Проверяем существование прогона
	run, err := s.repo.GetTestRun(ctx, runID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("прогон не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
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

	// Матчим и создаём результаты
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

// GetFlakyTests возвращает тест-кейсы с flaky-результатами из последних прогонов.
func (s *TMSService) GetFlakyTests(ctx context.Context, projectID uuid.UUID) ([]TestCaseResponse, error) {
	// Получаем последние прогоны
	params := ListTestRunsParams{Offset: 0, Limit: 10}
	runs, _, err := s.repo.ListTestRuns(ctx, projectID, params)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения прогонов для flaky-анализа")
	}

	// Собираем историю статусов по тест-кейсам
	tcStatuses := make(map[uuid.UUID][]string)
	tcMap := make(map[uuid.UUID]*ent.TestCase)
	for _, run := range runs {
		results, err := s.repo.ListTestResults(ctx, run.ID)
		if err != nil {
			continue
		}
		for _, r := range results {
			if r.Edges.TestCase != nil {
				tcID := r.Edges.TestCase.ID
				tcStatuses[tcID] = append(tcStatuses[tcID], string(r.Status))
				if _, ok := tcMap[tcID]; !ok {
					tcMap[tcID] = r.Edges.TestCase
				}
			}
		}
	}

	// Определяем flaky: если у тест-кейса есть разные статусы
	var flakyTests []TestCaseResponse
	for tcID, statuses := range tcStatuses {
		if len(statuses) < 2 {
			continue
		}
		first := statuses[0]
		isFlaky := false
		for _, s := range statuses[1:] {
			if s != first {
				isFlaky = true
				break
			}
		}
		if isFlaky {
			if tc, ok := tcMap[tcID]; ok {
				flakyTests = append(flakyTests, *toTestCaseResponse(tc))
			}
		}
	}

	if flakyTests == nil {
		flakyTests = make([]TestCaseResponse, 0)
	}
	return flakyTests, nil
}

// ---------------------------------------------------------------------------
// Импорт/экспорт (реализация в importer.go / exporter.go)
// ---------------------------------------------------------------------------

// ImportTestCases импортирует тест-кейсы из загруженного файла.
// Парсит файл, автоматически определяет формат, выполняет маппинг полей и создаёт тест-кейсы.
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

// publishEvent публикует событие в NATS (fire-and-forget).
// Безопасен при nil NATS-клиенте (пропускает публикацию).
func (s *TMSService) publishEvent(subject string, payload interface{}) {
	if s.nats == nil {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		slog.Error("не удалось сериализовать TMS событие", "subject", subject, "error", err)
		return
	}
	if err := s.nats.Publish(subject, data); err != nil {
		slog.Error("не удалось опубликовать TMS событие", "subject", subject, "error", err)
	}
}

// ---------------------------------------------------------------------------
// Стейт-машина прогонов
// ---------------------------------------------------------------------------

// isValidRunTransition проверяет допустимость перехода между статусами прогона.
func isValidRunTransition(from, to string) bool {
	validTransitions := map[string][]string{
		"created":     {"in_progress"},
		"in_progress": {"completed", "aborted"},
	}
	allowed, ok := validTransitions[from]
	if !ok {
		return false
	}
	for _, s := range allowed {
		if s == to {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Вспомогательные функции
// ---------------------------------------------------------------------------

// parseUUIDs преобразует слайс строк в слайс UUID.
func parseUUIDs(strs []string) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0, len(strs))
	for _, s := range strs {
		id, err := uuid.Parse(s)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// buildSnapshot создаёт снимок тест-кейса для версионирования.
func buildSnapshot(tc *ent.TestCase) map[string]interface{} {
	snapshot := map[string]interface{}{
		"title":         tc.Title,
		"description":   tc.Description,
		"precondition":  tc.Precondition,
		"postcondition": tc.Postcondition,
		"case_type":     string(tc.CaseType),
		"priority":      string(tc.Priority),
		"status":        string(tc.Status),
		"version":       tc.Version,
	}
	if len(tc.Steps) > 0 {
		stepsJSON, _ := json.Marshal(tc.Steps)
		var steps interface{}
		_ = json.Unmarshal(stepsJSON, &steps)
		snapshot["steps"] = steps
	}
	if tc.FullName != "" {
		snapshot["full_name"] = tc.FullName
	}
	if tc.ClassName != "" {
		snapshot["class_name"] = tc.ClassName
	}
	if tc.MethodName != "" {
		snapshot["method_name"] = tc.MethodName
	}
	if tc.ExternalID != "" {
		snapshot["external_id"] = tc.ExternalID
	}
	return snapshot
}

// snapshotToUpdateInput преобразует снимок версии в UpdateTestCaseInput для отката.
func snapshotToUpdateInput(snapshot map[string]interface{}) UpdateTestCaseInput {
	input := UpdateTestCaseInput{}

	if v, ok := snapshot["title"].(string); ok {
		input.Title = &v
	}
	if v, ok := snapshot["description"].(string); ok {
		input.Description = &v
	}
	if v, ok := snapshot["precondition"].(string); ok {
		input.Precondition = &v
	}
	if v, ok := snapshot["postcondition"].(string); ok {
		input.Postcondition = &v
	}
	if v, ok := snapshot["case_type"].(string); ok {
		input.CaseType = &v
	}
	if v, ok := snapshot["priority"].(string); ok {
		input.Priority = &v
	}
	if v, ok := snapshot["status"].(string); ok {
		input.Status = &v
	}
	if v, ok := snapshot["steps"]; ok {
		stepsJSON, _ := json.Marshal(v)
		var steps []schema.TestStep
		if err := json.Unmarshal(stepsJSON, &steps); err == nil {
			input.Steps = &steps
		}
	}
	if v, ok := snapshot["full_name"].(string); ok {
		input.FullName = &v
	}
	if v, ok := snapshot["class_name"].(string); ok {
		input.ClassName = &v
	}
	if v, ok := snapshot["method_name"].(string); ok {
		input.MethodName = &v
	}
	if v, ok := snapshot["external_id"].(string); ok {
		input.ExternalID = &v
	}

	return input
}

// ---------------------------------------------------------------------------
// Конвертеры Ent -> DTO
// ---------------------------------------------------------------------------

// toTestCaseResponse конвертирует Ent TestCase в DTO.
func toTestCaseResponse(tc *ent.TestCase) *TestCaseResponse {
	resp := &TestCaseResponse{
		ID:            tc.ID.String(),
		CaseNumber:    tc.CaseNumber,
		CaseKey:       tc.CaseKey,
		Title:         tc.Title,
		Description:   tc.Description,
		Precondition:  tc.Precondition,
		Postcondition: tc.Postcondition,
		CaseType:      string(tc.CaseType),
		Priority:      string(tc.Priority),
		Status:        string(tc.Status),
		Steps:         tc.Steps,
		FullName:      tc.FullName,
		ClassName:     tc.ClassName,
		MethodName:    tc.MethodName,
		ExternalID:    tc.ExternalID,
		Version:       tc.Version,
		CreatedAt:     tc.CreatedAt,
		UpdatedAt:     tc.UpdatedAt,
	}

	if tc.Edges.Suite != nil {
		resp.Suite = &SuiteBriefResponse{
			ID:   tc.Edges.Suite.ID.String(),
			Name: tc.Edges.Suite.Name,
		}
	}
	if tc.Edges.CreatedBy != nil {
		resp.CreatedBy = &UserBriefResponse{
			ID:        tc.Edges.CreatedBy.ID.String(),
			Name:      tc.Edges.CreatedBy.Name,
			Email:     tc.Edges.CreatedBy.Email,
			AvatarURL: tc.Edges.CreatedBy.AvatarURL,
		}
	}
	if tc.Edges.Labels != nil {
		labels := make([]LabelResponse, 0, len(tc.Edges.Labels))
		for _, l := range tc.Edges.Labels {
			labels = append(labels, LabelResponse{
				ID:    l.ID.String(),
				Name:  l.Name,
				Color: l.Color,
			})
		}
		resp.Labels = labels
	}
	if tc.Edges.Components != nil {
		components := make([]ComponentResponse, 0, len(tc.Edges.Components))
		for _, c := range tc.Edges.Components {
			components = append(components, ComponentResponse{
				ID:          c.ID.String(),
				Name:        c.Name,
				Description: c.Description,
				CreatedAt:   c.CreatedAt,
			})
		}
		resp.Components = components
	}
	if tc.Edges.LinkedIssues != nil {
		issues := make([]IssueBriefResponse, 0, len(tc.Edges.LinkedIssues))
		for _, i := range tc.Edges.LinkedIssues {
			issues = append(issues, IssueBriefResponse{
				ID:       i.ID.String(),
				IssueKey: i.IssueKey,
				Title:    i.Title,
			})
		}
		resp.LinkedIssues = issues
	}

	return resp
}

// toTestCaseVersionResponse конвертирует Ent TestCaseVersion в DTO.
func toTestCaseVersionResponse(v *ent.TestCaseVersion) *TestCaseVersionResponse {
	resp := &TestCaseVersionResponse{
		ID:                v.ID.String(),
		VersionNumber:     v.VersionNumber,
		Snapshot:          v.Snapshot,
		ChangeDescription: v.ChangeDescription,
		CreatedAt:         v.CreatedAt,
	}
	if v.Edges.CreatedBy != nil {
		resp.CreatedBy = &UserBriefResponse{
			ID:        v.Edges.CreatedBy.ID.String(),
			Name:      v.Edges.CreatedBy.Name,
			Email:     v.Edges.CreatedBy.Email,
			AvatarURL: v.Edges.CreatedBy.AvatarURL,
		}
	}
	return resp
}

// toSuiteResponse конвертирует Ent TestSuite в DTO.
func toSuiteResponse(suite *ent.TestSuite) *SuiteResponse {
	resp := &SuiteResponse{
		ID:          suite.ID.String(),
		Name:        suite.Name,
		Description: suite.Description,
		SortOrder:   suite.SortOrder,
		CreatedAt:   suite.CreatedAt,
		UpdatedAt:   suite.UpdatedAt,
	}
	if suite.Edges.Parent != nil {
		resp.ParentID = suite.Edges.Parent.ID.String()
	}
	return resp
}

// toComponentResponse конвертирует Ent Component в DTO.
func toComponentResponse(c *ent.Component) *ComponentResponse {
	return &ComponentResponse{
		ID:          c.ID.String(),
		Name:        c.Name,
		Description: c.Description,
		CreatedAt:   c.CreatedAt,
	}
}

// toTestRunResponse конвертирует Ent TestRun в DTO.
func toTestRunResponse(run *ent.TestRun) *TestRunResponse {
	resp := &TestRunResponse{
		ID:           run.ID.String(),
		Title:        run.Title,
		Description:  run.Description,
		RunType:      string(run.RunType),
		Status:       string(run.Status),
		TotalCount:   run.TotalCount,
		PassedCount:  run.PassedCount,
		FailedCount:  run.FailedCount,
		SkippedCount: run.SkippedCount,
		BrokenCount:  run.BrokenCount,
		UnknownCount: run.UnknownCount,
		AllureRunID:  run.AllureRunID,
		StartedAt:    run.StartedAt,
		CompletedAt:  run.CompletedAt,
		CreatedAt:    run.CreatedAt,
		UpdatedAt:    run.UpdatedAt,
	}
	if run.Edges.CreatedBy != nil {
		resp.CreatedBy = &UserBriefResponse{
			ID:        run.Edges.CreatedBy.ID.String(),
			Name:      run.Edges.CreatedBy.Name,
			Email:     run.Edges.CreatedBy.Email,
			AvatarURL: run.Edges.CreatedBy.AvatarURL,
		}
	}
	return resp
}

// toTestResultResponse конвертирует Ent TestResult в DTO.
func toTestResultResponse(r *ent.TestResult) *TestResultResponse {
	resp := &TestResultResponse{
		ID:            r.ID.String(),
		Status:        string(r.Status),
		Comment:       r.Comment,
		StepResults:   r.StepResults,
		DurationMs:    r.DurationMs,
		AllureUUID:    r.AllureUUID,
		ErrorMessage:  r.ErrorMessage,
		ErrorTrace:    r.ErrorTrace,
		IsFlaky:       r.IsFlaky,
		MatchStrategy: string(r.MatchStrategy),
		StartedAt:     r.StartedAt,
		CompletedAt:   r.CompletedAt,
		CreatedAt:     r.CreatedAt,
	}
	if r.Edges.TestCase != nil {
		resp.TestCaseKey = r.Edges.TestCase.CaseKey
	}
	if r.Edges.ExecutedBy != nil {
		resp.ExecutedBy = &UserBriefResponse{
			ID:        r.Edges.ExecutedBy.ID.String(),
			Name:      r.Edges.ExecutedBy.Name,
			Email:     r.Edges.ExecutedBy.Email,
			AvatarURL: r.Edges.ExecutedBy.AvatarURL,
		}
	}
	return resp
}

// toTestRunDetailResponse конвертирует Ent TestRun (с результатами и назначенными) в DTO.
func toTestRunDetailResponse(run *ent.TestRun) *TestRunDetailResponse {
	resp := &TestRunDetailResponse{
		TestRunResponse: *toTestRunResponse(run),
	}

	// Результаты
	resp.Results = make([]TestResultDetailResponse, 0, len(run.Edges.Results))
	for _, r := range run.Edges.Results {
		detail := TestResultDetailResponse{
			ID:          r.ID.String(),
			Status:      string(r.Status),
			Comment:     r.Comment,
			StepResults: r.StepResults,
			DurationMs:  r.DurationMs,
			IsFlaky:     r.IsFlaky,
			StartedAt:   r.StartedAt,
			CompletedAt: r.CompletedAt,
			CreatedAt:   r.CreatedAt,
		}
		if r.Edges.TestCase != nil {
			tc := r.Edges.TestCase
			detail.TestCase = &TestCaseBriefDTO{
				ID:       tc.ID.String(),
				CaseKey:  tc.CaseKey,
				Title:    tc.Title,
				CaseType: string(tc.CaseType),
				Priority: string(tc.Priority),
			}
		}
		if r.Edges.ExecutedBy != nil {
			detail.ExecutedBy = &UserBriefResponse{
				ID:        r.Edges.ExecutedBy.ID.String(),
				Name:      r.Edges.ExecutedBy.Name,
				Email:     r.Edges.ExecutedBy.Email,
				AvatarURL: r.Edges.ExecutedBy.AvatarURL,
			}
		}
		if r.Edges.Files != nil {
			detail.Files = make([]FileBriefResponse, 0, len(r.Edges.Files))
			for _, f := range r.Edges.Files {
				detail.Files = append(detail.Files, FileBriefResponse{
					ID:       f.ID.String(),
					FileName: f.Filename,
					FileSize: f.Size,
					MimeType: f.ContentType,
				})
			}
		}
		resp.Results = append(resp.Results, detail)
	}

	// Назначенные тестировщики
	resp.Assignees = make([]UserBriefResponse, 0, len(run.Edges.Assignees))
	for _, u := range run.Edges.Assignees {
		resp.Assignees = append(resp.Assignees, UserBriefResponse{
			ID:        u.ID.String(),
			Name:      u.Name,
			Email:     u.Email,
			AvatarURL: u.AvatarURL,
		})
	}

	return resp
}

// buildSuiteTree строит дерево папок из плоского списка.
func buildSuiteTree(suites []*ent.TestSuite) []SuiteResponse {
	// Карта id -> response
	byID := make(map[uuid.UUID]*SuiteResponse, len(suites))
	for _, s := range suites {
		resp := toSuiteResponse(s)
		resp.Children = make([]SuiteResponse, 0)
		byID[s.ID] = resp
	}

	// Строим дерево
	var roots []SuiteResponse
	for _, s := range suites {
		if s.Edges.Parent == nil {
			roots = append(roots, *byID[s.ID])
		} else {
			parentResp, ok := byID[s.Edges.Parent.ID]
			if ok {
				parentResp.Children = append(parentResp.Children, *byID[s.ID])
			}
		}
	}

	if roots == nil {
		roots = make([]SuiteResponse, 0)
	}
	return roots
}
