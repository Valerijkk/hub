package tracker

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Метки
// ---------------------------------------------------------------------------

// CreateLabel создаёт метку в проекте.
func (s *TrackerService) CreateLabel(ctx context.Context, projectID uuid.UUID, req CreateLabelRequest) (*LabelResponse, error) {
	var groupID *uuid.UUID
	if req.GroupID != "" {
		id, err := uuid.Parse(req.GroupID)
		if err != nil {
			return nil, apperrors.NewBadRequest("некорректный group_id")
		}
		groupID = &id
	}

	l, err := s.repo.CreateLabel(ctx, projectID, req.Name, req.Color, groupID)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("метка с таким именем уже существует")
		}
		return nil, apperrors.NewInternal("ошибка создания метки")
	}

	return toLabelResponse(l), nil
}

// ListLabels возвращает все метки проекта.
func (s *TrackerService) ListLabels(ctx context.Context, projectID uuid.UUID) ([]LabelResponse, error) {
	labels, err := s.repo.ListLabels(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения меток")
	}
	result := make([]LabelResponse, 0, len(labels))
	for _, l := range labels {
		result = append(result, *toLabelResponse(l))
	}
	return result, nil
}

// DeleteLabel удаляет метку в рамках проекта (Block F IDOR).
func (s *TrackerService) DeleteLabel(ctx context.Context, projectID, labelID uuid.UUID) error {
	if err := s.repo.DeleteLabel(ctx, projectID, labelID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("метка не найдена")
		}
		return apperrors.NewInternal("ошибка удаления метки")
	}
	return nil
}

// CreateLabelGroup создаёт группу меток.
func (s *TrackerService) CreateLabelGroup(ctx context.Context, projectID uuid.UUID, req CreateLabelGroupRequest) (*LabelGroupResponse, error) {
	g, err := s.repo.CreateLabelGroup(ctx, projectID, req.Name)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("группа меток с таким именем уже существует")
		}
		return nil, apperrors.NewInternal("ошибка создания группы меток")
	}
	return toLabelGroupResponse(g), nil
}

// ListLabelGroups возвращает все группы меток проекта.
func (s *TrackerService) ListLabelGroups(ctx context.Context, projectID uuid.UUID) ([]LabelGroupResponse, error) {
	groups, err := s.repo.ListLabelGroups(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения групп меток")
	}
	result := make([]LabelGroupResponse, 0, len(groups))
	for _, g := range groups {
		result = append(result, *toLabelGroupResponse(g))
	}
	return result, nil
}
