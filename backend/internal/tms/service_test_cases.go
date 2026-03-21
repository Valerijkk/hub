package tms

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Test cases CRUD
// ---------------------------------------------------------------------------

// CreateTestCase creates a new test case in a project.
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

// GetTestCase returns a test case by key.
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

// ListTestCases returns a paginated list of test cases.
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

// UpdateTestCase updates a test case, creating a version snapshot first.
func (s *TMSService) UpdateTestCase(ctx context.Context, caseKey string, userID uuid.UUID, req UpdateTestCaseRequest) (*TestCaseResponse, error) {
	existing, err := s.repo.FindTestCaseByKey(ctx, caseKey)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("тест-кейс не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения тест-кейса")
	}

	// Create snapshot of current version before updating
	snapshot := buildSnapshot(existing)
	_, err = s.repo.CreateTestCaseVersion(ctx, existing.ID, userID, existing.Version, snapshot, "")
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания версии")
	}

	// Build update input
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

// DeleteTestCase deletes a test case by key.
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

// RollbackTestCase rolls back a test case to a specified version.
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

	// Create snapshot of current state before rollback
	currentSnapshot := buildSnapshot(existing)
	_, err = s.repo.CreateTestCaseVersion(ctx, existing.ID, userID, existing.Version, currentSnapshot, "перед откатом к версии")
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания снимка перед откатом")
	}

	// Apply snapshot as update
	input := snapshotToUpdateInput(version.Snapshot)
	updated, err := s.repo.UpdateTestCase(ctx, existing.ID, input)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка применения отката")
	}

	return toTestCaseResponse(updated), nil
}

// ---------------------------------------------------------------------------
// Versions
// ---------------------------------------------------------------------------

// ListTestCaseVersions returns all versions of a test case.
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
// Issue linking
// ---------------------------------------------------------------------------

// LinkTestCaseToIssue links a test case to an issue.
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

// UnlinkTestCaseFromIssue unlinks a test case from an issue.
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

// GetTestCaseLinks returns issues linked to a test case.
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

// CreateBugFromFailedResult creates a bug from a failed test result.
func (s *TMSService) CreateBugFromFailedResult(ctx context.Context, projectID, userID uuid.UUID, resultID uuid.UUID, req CreateBugRequest) (*IssueBriefResponse, error) {
	// Get result with test case
	result, err := s.repo.GetTestResult(ctx, resultID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("результат не найден")
		}
		return nil, apperrors.NewInternal("ошибка получения результата")
	}

	// Determine test case
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
