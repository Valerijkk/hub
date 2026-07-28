package tms

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// CRUD тест-кейсов
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
		AutomationID:  req.AutomationID,
		CreatedByID:   userID,
	}

	if req.SuiteID != "" {
		id, err := uuid.Parse(req.SuiteID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный suite_id")
		}
		if _, err := s.repo.GetSuiteInProject(ctx, projectID, id); err != nil {
			if ent.IsNotFound(err) {
				return nil, apperrors.NewNotFound("папка не найдена")
			}
			return nil, apperrors.NewInternal("ошибка проверки папки")
		}
		input.SuiteID = &id
	}
	if len(req.Labels) > 0 {
		ids, err := parseUUIDs(req.Labels)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный label ID")
		}
		n, err := s.repo.CountLabelsInProject(ctx, projectID, ids)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка проверки меток")
		}
		if n != len(ids) {
			return nil, apperrors.NewNotFound("метка не найдена")
		}
		input.LabelIDs = ids
	}
	if len(req.Components) > 0 {
		ids, err := parseUUIDs(req.Components)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный component ID")
		}
		n, err := s.repo.CountComponentsInProject(ctx, projectID, ids)
		if err != nil {
			return nil, apperrors.NewInternal("ошибка проверки компонентов")
		}
		if n != len(ids) {
			return nil, apperrors.NewNotFound("компонент не найден")
		}
		input.ComponentIDs = ids
	}

	tc, err := s.repo.CreateTestCase(ctx, projectID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания тест-кейса")
	}

	return toTestCaseResponse(tc), nil
}

// GetTestCase возвращает тест-кейс по ключу (например, "HUB-TC-1").
func (s *TMSService) GetTestCase(ctx context.Context, projectID uuid.UUID, caseKey string) (*TestCaseResponse, error) {
	tc, err := s.repo.FindTestCaseByKey(ctx, projectID, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("тест-кейс не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения тест-кейса")
	}
	return toTestCaseResponse(tc), nil
}

// ListTestCases возвращает пагинированный список тест-кейсов проекта.
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

// UpdateTestCase обновляет тест-кейс, предварительно создавая снимок текущей версии.
func (s *TMSService) UpdateTestCase(ctx context.Context, projectID uuid.UUID, caseKey string, userID uuid.UUID, req UpdateTestCaseRequest) (*TestCaseResponse, error) {
	existing, err := s.repo.FindTestCaseByKey(ctx, projectID, caseKey)
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
		AutomationID:  req.AutomationID,
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
			if _, err := s.repo.GetSuiteInProject(ctx, projectID, id); err != nil {
				if ent.IsNotFound(err) {
					return nil, apperrors.NewNotFound("папка не найдена")
				}
				return nil, apperrors.NewInternal("ошибка проверки папки")
			}
			input.SuiteID = &id
		}
	}
	if req.Labels != nil {
		ids, err := parseUUIDs(*req.Labels)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный label ID")
		}
		if len(ids) > 0 {
			n, err := s.repo.CountLabelsInProject(ctx, projectID, ids)
			if err != nil {
				return nil, apperrors.NewInternal("ошибка проверки меток")
			}
			if n != len(ids) {
				return nil, apperrors.NewNotFound("метка не найдена")
			}
		}
		input.LabelIDs = &ids
	}
	if req.Components != nil {
		ids, err := parseUUIDs(*req.Components)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный component ID")
		}
		if len(ids) > 0 {
			n, err := s.repo.CountComponentsInProject(ctx, projectID, ids)
			if err != nil {
				return nil, apperrors.NewInternal("ошибка проверки компонентов")
			}
			if n != len(ids) {
				return nil, apperrors.NewNotFound("компонент не найден")
			}
		}
		input.ComponentIDs = &ids
	}

	updated, err := s.repo.UpdateTestCase(ctx, existing.ID, input)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return nil, err
		}
		return nil, apperrors.NewInternal("ошибка обновления тест-кейса")
	}

	return toTestCaseResponse(updated), nil
}

// DeleteTestCase удаляет тест-кейс по ключу.
func (s *TMSService) DeleteTestCase(ctx context.Context, projectID uuid.UUID, caseKey string) error {
	existing, err := s.repo.FindTestCaseByKey(ctx, projectID, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("тест-кейс не найден")
		}
		return apperrors.NewInternal("ошибка получения тест-кейса")
	}
	return s.repo.DeleteTestCase(ctx, existing.ID)
}

// RollbackTestCase откатывает тест-кейс к указанной версии.
func (s *TMSService) RollbackTestCase(ctx context.Context, projectID uuid.UUID, caseKey string, versionID, userID uuid.UUID) (*TestCaseResponse, error) {
	existing, err := s.repo.FindTestCaseByKey(ctx, projectID, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("тест-кейс не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения тест-кейса")
	}

	// Ownership enforced in EntRepository via HasTestCaseWith (A3-007).
	version, err := s.repo.GetTestCaseVersion(ctx, existing.ID, versionID)
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

	// Применяем снимок как обновление (optimistic lock на текущей version)
	input := snapshotToUpdateInput(version.Snapshot)
	input.Version = existing.Version
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
func (s *TMSService) ListTestCaseVersions(ctx context.Context, projectID uuid.UUID, caseKey string) ([]TestCaseVersionResponse, error) {
	tc, err := s.repo.FindTestCaseByKey(ctx, projectID, caseKey)
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
// Связи тест-кейс↔задача
// ---------------------------------------------------------------------------

// LinkTestCaseToIssue привязывает тест-кейс к задаче трекера.
// Issue должна принадлежать тому же projectID (A3-003-link) — иначе NotFound.
func (s *TMSService) LinkTestCaseToIssue(ctx context.Context, projectID uuid.UUID, caseKey string, issueID uuid.UUID) error {
	tc, err := s.repo.FindTestCaseByKey(ctx, projectID, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("тест-кейс не найден")
		}
		return apperrors.NewInternal("ошибка получения тест-кейса")
	}
	if _, err := s.repo.FindIssueInProject(ctx, projectID, issueID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("задача не найдена")
		}
		return apperrors.NewInternal("ошибка получения задачи")
	}
	return s.repo.LinkTestCaseToIssue(ctx, tc.ID, issueID)
}

// UnlinkTestCaseFromIssue отвязывает тест-кейс от задачи трекера.
// Симметрично Link: foreign issue → NotFound (A3-003-link).
func (s *TMSService) UnlinkTestCaseFromIssue(ctx context.Context, projectID uuid.UUID, caseKey string, issueID uuid.UUID) error {
	tc, err := s.repo.FindTestCaseByKey(ctx, projectID, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("тест-кейс не найден")
		}
		return apperrors.NewInternal("ошибка получения тест-кейса")
	}
	if _, err := s.repo.FindIssueInProject(ctx, projectID, issueID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("задача не найдена")
		}
		return apperrors.NewInternal("ошибка получения задачи")
	}
	return s.repo.UnlinkTestCaseFromIssue(ctx, tc.ID, issueID)
}

// GetTestCaseLinks возвращает задачи, привязанные к тест-кейсу.
func (s *TMSService) GetTestCaseLinks(ctx context.Context, projectID uuid.UUID, caseKey string) ([]IssueBriefResponse, error) {
	tc, err := s.repo.FindTestCaseByKey(ctx, projectID, caseKey)
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
// runID из URL сверяется с принадлежностью result к прогону (A2W2-1-001 / A3-008).
func (s *TMSService) CreateBugFromFailedResult(ctx context.Context, projectID, runID, userID, resultID uuid.UUID, req CreateBugRequest) (*IssueBriefResponse, error) {
	// Project scope на URL-прогоне до чтения result (не доверяем произвольному resultId).
	if _, err := s.repo.GetTestRunInProject(ctx, projectID, runID); err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("результат не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения прогона")
	}

	// HasRunWith(runID): result из другого прогона того же проекта → 404.
	result, err := s.repo.GetTestResultInRun(ctx, runID, resultID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("результат не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения результата")
	}

	// Определяем тест-кейс для заголовка бага
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
		return nil, apperrors.NewInternal("ошибка создания бага")
	}

	return &IssueBriefResponse{
		ID:       issue.ID.String(),
		IssueKey: issue.IssueKey,
		Title:    issue.Title,
	}, nil
}
