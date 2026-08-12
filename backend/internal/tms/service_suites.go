package tms

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Папки (suites)
// ---------------------------------------------------------------------------

// CreateSuite создаёт новую папку/секцию тест-кейсов.
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
		if _, err := s.repo.GetSuiteInProject(ctx, projectID, id); err != nil {
			if ent.IsNotFound(err) {
				return nil, apperrors.NewNotFound("родительская папка не найдена")
			}
			return nil, apperrors.NewInternal("ошибка проверки родительской папки")
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

	// Возвращаем только корневые узлы (без parent) — UI строит дерево рекурсивно.
	return buildSuiteTree(suites), nil
}

// UpdateSuite обновляет папку только внутри projectID (P1-BE-10).
// A3W2-1-001: VersionEQ conflict → 409, не маскировать в 500.
func (s *TMSService) UpdateSuite(ctx context.Context, projectID, suiteID uuid.UUID, req UpdateSuiteRequest) (*SuiteResponse, error) {
	input := UpdateSuiteInput{
		Name:        req.Name,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		Version:     req.Version,
	}

	suite, err := s.repo.UpdateSuite(ctx, projectID, suiteID, input)
	if err != nil {
		var appErr *apperrors.AppError
		if errors.As(err, &appErr) {
			return nil, err
		}
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("папка не найдена")
		}
		return nil, apperrors.NewInternal("ошибка обновления папки")
	}

	return toSuiteResponse(suite), nil
}

// DeleteSuite удаляет папку только внутри projectID (P1-BE-10).
func (s *TMSService) DeleteSuite(ctx context.Context, projectID, suiteID uuid.UUID) error {
	if err := s.repo.DeleteSuite(ctx, projectID, suiteID); err != nil {
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

// DeleteComponent удаляет компонент только внутри projectID (P1-BE-10).
func (s *TMSService) DeleteComponent(ctx context.Context, projectID, componentID uuid.UUID) error {
	if err := s.repo.DeleteComponent(ctx, projectID, componentID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("компонент не найден")
		}
		return apperrors.NewInternal("ошибка удаления компонента")
	}
	return nil
}
