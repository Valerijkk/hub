package tracker

import (
	"context"

	"github.com/google/uuid"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// WorkflowEngine валидирует переходы между статусами задач.
// Использует репозиторий для проверки существования разрешённого перехода.
type WorkflowEngine struct {
	repo Repository
}

// NewWorkflowEngine создаёт новый движок workflow.
func NewWorkflowEngine(repo Repository) *WorkflowEngine {
	return &WorkflowEngine{repo: repo}
}

// ValidateTransition проверяет, разрешён ли переход из одного статуса в другой.
// Возвращает nil если переход допустим, ошибку если нет.
func (w *WorkflowEngine) ValidateTransition(ctx context.Context, fromStatusID, toStatusID uuid.UUID) error {
	_, err := w.repo.FindTransition(ctx, fromStatusID, toStatusID)
	if err != nil {
		return apperrors.NewBadRequest("переход из текущего статуса не разрешён")
	}
	return nil
}
