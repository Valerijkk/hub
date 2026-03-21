package tms

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/schema"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// TMSService реализует бизнес-логику модуля управления тестированием.
// Оркестрирует вызовы репозитория, версионирование, стейт-машину прогонов.
type TMSService struct {
	repo Repository
}

// NewTMSService создаёт новый сервис TMS.
func NewTMSService(repo Repository) *TMSService {
	return &TMSService{repo: repo}
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
