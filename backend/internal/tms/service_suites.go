package tms

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Suites
// ---------------------------------------------------------------------------

// CreateSuite creates a new test case folder/suite.
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

// ListSuites returns the suite tree for a project.
func (s *TMSService) ListSuites(ctx context.Context, projectID uuid.UUID) ([]SuiteResponse, error) {
	suites, err := s.repo.ListSuites(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения папок")
	}

	// Build tree: return only root nodes (no parent)
	return buildSuiteTree(suites), nil
}

// UpdateSuite updates a suite.
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

// DeleteSuite deletes a suite.
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
// Components
// ---------------------------------------------------------------------------

// CreateComponent creates a new component.
func (s *TMSService) CreateComponent(ctx context.Context, projectID uuid.UUID, req CreateComponentRequest) (*ComponentResponse, error) {
	comp, err := s.repo.CreateComponent(ctx, projectID, req.Name, req.Description)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка создания компонента")
	}
	return toComponentResponse(comp), nil
}

// ListComponents returns all components for a project.
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

// DeleteComponent deletes a component.
func (s *TMSService) DeleteComponent(ctx context.Context, componentID uuid.UUID) error {
	if err := s.repo.DeleteComponent(ctx, componentID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("компонент не найден")
		}
		return apperrors.NewInternal("ошибка удаления компонента")
	}
	return nil
}
