package tracker

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Рабочий процесс
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

// UpdateWorkflowStatus обновляет статус workflow в рамках проекта (P1-BE-01).
func (s *TrackerService) UpdateWorkflowStatus(ctx context.Context, projectID, statusID uuid.UUID, req CreateWorkflowStatusRequest) (*WorkflowStatusResponse, error) {
	st, err := s.repo.UpdateWorkflowStatus(ctx, projectID, statusID, &req.Name, &req.Color, &req.SortOrder)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("статус не найден")
		}
		return nil, apperrors.NewInternal("ошибка обновления статуса")
	}
	return toWorkflowStatusResponse(st), nil
}

// DeleteWorkflowStatus мягко удаляет статус workflow в рамках проекта (P1-BE-01 / A3W8-5-002).
func (s *TrackerService) DeleteWorkflowStatus(ctx context.Context, projectID, statusID uuid.UUID) error {
	if err := s.repo.DeleteWorkflowStatus(ctx, projectID, statusID); err != nil {
		var ae *apperrors.AppError
		if errors.As(err, &ae) {
			return ae
		}
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
		// W1-A2-015 / P3-BE-07: без eager edges — skip, не nil deref.
		if t.Edges.FromStatus == nil || t.Edges.ToStatus == nil {
			continue
		}
		result = append(result, WorkflowTransitionResponse{
			ID:           t.ID.String(),
			FromStatusID: t.Edges.FromStatus.ID.String(),
			ToStatusID:   t.Edges.ToStatus.ID.String(),
		})
	}
	return result, nil
}

// CreateWorkflowTransition создаёт разрешённый переход между статусами проекта (P1-BE-01).
func (s *TrackerService) CreateWorkflowTransition(ctx context.Context, projectID uuid.UUID, req CreateWorkflowTransitionRequest) (*WorkflowTransitionResponse, error) {
	fromID, err := uuid.Parse(req.FromStatusID)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный from_status_id")
	}
	toID, err := uuid.Parse(req.ToStatusID)
	if err != nil {
		return nil, apperrors.NewBadRequest("некорректный to_status_id")
	}

	t, err := s.repo.CreateWorkflowTransition(ctx, projectID, fromID, toID)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("статус не найден")
		}
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

// DeleteWorkflowTransition удаляет переход workflow в рамках проекта (P1-BE-01).
func (s *TrackerService) DeleteWorkflowTransition(ctx context.Context, projectID, transitionID uuid.UUID) error {
	if err := s.repo.DeleteWorkflowTransition(ctx, projectID, transitionID); err != nil {
		if ent.IsNotFound(err) {
			return apperrors.NewNotFound("переход не найден")
		}
		return apperrors.NewInternal("ошибка удаления перехода")
	}
	return nil
}
