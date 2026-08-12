package tracker

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Кастомные поля
// ---------------------------------------------------------------------------

// CreateCustomField создаёт определение кастомного поля.
func (s *TrackerService) CreateCustomField(ctx context.Context, projectID uuid.UUID, req CreateCustomFieldRequest) (*CustomFieldDefinitionResponse, error) {
	input := CreateCustomFieldInput{
		Name:      req.Name,
		FieldType: req.FieldType,
		Options:   req.Options,
		Required:  req.Required,
	}

	cf, err := s.repo.CreateCustomField(ctx, projectID, input)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("поле с таким именем уже существует")
		}
		return nil, apperrors.NewInternal("ошибка создания кастомного поля")
	}
	return toCustomFieldResponse(cf), nil
}

// ListCustomFields возвращает все определения кастомных полей проекта.
func (s *TrackerService) ListCustomFields(ctx context.Context, projectID uuid.UUID) ([]CustomFieldDefinitionResponse, error) {
	fields, err := s.repo.ListCustomFields(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения кастомных полей")
	}
	result := make([]CustomFieldDefinitionResponse, 0, len(fields))
	for _, f := range fields {
		result = append(result, *toCustomFieldResponse(f))
	}
	return result, nil
}

// DeleteCustomField удаляет определение кастомного поля в рамках проекта (P1-BE-02).
func (s *TrackerService) DeleteCustomField(ctx context.Context, projectID, fieldID uuid.UUID) error {
	if err := s.repo.DeleteCustomField(ctx, projectID, fieldID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("поле не найдено")
		}
		return apperrors.NewInternal("ошибка удаления кастомного поля")
	}
	return nil
}
