package tracker

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Workflow
// ---------------------------------------------------------------------------

// ListWorkflowStatuses возвращает все статусы workflow проекта.
func (s *TrackerService) ListWorkflowStatuses(ctx context.Context, projectID uuid.UUID) ([]WorkflowStatusResponse, error) {
	statuses, err := s.repo.ListWorkflowStatuses(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения статусов")
	}
	result := make([]WorkflowStatusResponse, 0, len(statuses))
	for _, st := range statuses {
		result = append(result, *toWorkflowStatusResponse(st))
	}
	return result, nil
}

// CreateWorkflowStatus создаёт новый статус workflow.
func (s *TrackerService) CreateWorkflowStatus(ctx context.Context, projectID uuid.UUID, req CreateWorkflowStatusRequest) (*WorkflowStatusResponse, error) {
	st, err := s.repo.CreateWorkflowStatus(ctx, projectID, req.Name, req.Category, req.Color, req.SortOrder)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("статус с таким именем уже существует")
		}
		return nil, apperrors.NewInternal("ошибка создания статуса")
	}
	return toWorkflowStatusResponse(st), nil
}

// UpdateWorkflowStatus обновляет статус workflow.
func (s *TrackerService) UpdateWorkflowStatus(ctx context.Context, statusID uuid.UUID, req CreateWorkflowStatusRequest) (*WorkflowStatusResponse, error) {
	st, err := s.repo.UpdateWorkflowStatus(ctx, statusID, &req.Name, &req.Color, &req.SortOrder)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("статус не найден")
		}
		return nil, apperrors.NewInternal("ошибка обновления статуса")
	}
	return toWorkflowStatusResponse(st), nil
}

// DeleteWorkflowStatus удаляет статус workflow.
func (s *TrackerService) DeleteWorkflowStatus(ctx context.Context, statusID uuid.UUID) error {
	if err := s.repo.DeleteWorkflowStatus(ctx, statusID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("статус не найден")
		}
		return apperrors.NewInternal("ошибка удаления статуса")
	}
	return nil
}

// ListWorkflowTransitions возвращает все переходы workflow проекта.
func (s *TrackerService) ListWorkflowTransitions(ctx context.Context, projectID uuid.UUID) ([]WorkflowTransitionResponse, error) {
	transitions, err := s.repo.ListWorkflowTransitions(ctx, projectID)
	if err != nil {
		return nil, apperrors.NewInternal("ошибка получения переходов")
	}
	result := make([]WorkflowTransitionResponse, 0, len(transitions))
	for _, t := range transitions {
		result = append(result, WorkflowTransitionResponse{
			ID:           t.ID.String(),
			FromStatusID: t.Edges.FromStatus.ID.String(),
			ToStatusID:   t.Edges.ToStatus.ID.String(),
		})
	}
	return result, nil
}

// CreateWorkflowTransition создаёт разрешённый переход между статусами.
func (s *TrackerService) CreateWorkflowTransition(ctx context.Context, req CreateWorkflowTransitionRequest) (*WorkflowTransitionResponse, error) {
	fromID, err := uuid.Parse(req.FromStatusID)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный from_status_id")
	}
	toID, err := uuid.Parse(req.ToStatusID)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный to_status_id")
	}

	t, err := s.repo.CreateWorkflowTransition(ctx, fromID, toID)
	if err != nil {
		if ent.IsConstraintError(err) {
			return nil, apperrors.NewConflict("такой переход уже существует")
		}
		return nil, apperrors.NewInternal("ошибка создания перехода")
	}
	return &WorkflowTransitionResponse{
		ID:           t.ID.String(),
		FromStatusID: req.FromStatusID,
		ToStatusID:   req.ToStatusID,
	}, nil
}

// DeleteWorkflowTransition удаляет переход workflow.
func (s *TrackerService) DeleteWorkflowTransition(ctx context.Context, transitionID uuid.UUID) error {
	if err := s.repo.DeleteWorkflowTransition(ctx, transitionID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("переход не найден")
		}
		return apperrors.NewInternal("ошибка удаления перехода")
	}
	return nil
}
