package tracker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/ent/workflowtransition"
)

// ---------------------------------------------------------------------------
// Workflow
// ---------------------------------------------------------------------------

// ListWorkflowStatuses возвращает все статусы workflow проекта, отсортированные по sort_order.
func (r *EntRepository) ListWorkflowStatuses(ctx context.Context, projectID uuid.UUID) ([]*ent.WorkflowStatus, error) {
	return r.client.WorkflowStatus.Query().
		Where(workflowstatus.HasProjectWith(project.IDEQ(projectID))).
		Order(ent.Asc(workflowstatus.FieldSortOrder)).
		All(ctx)
}

// CreateWorkflowStatus создаёт новый статус workflow в проекте.
func (r *EntRepository) CreateWorkflowStatus(ctx context.Context, projectID uuid.UUID, name, category, color string, sortOrder int) (*ent.WorkflowStatus, error) {
	create := r.client.WorkflowStatus.Create().
		SetName(name).
		SetCategory(workflowstatus.Category(category)).
		SetColor(color).
		SetSortOrder(sortOrder).
		SetProjectID(projectID)

	return create.Save(ctx)
}

// UpdateWorkflowStatus обновляет статус workflow (имя, цвет, порядок).
func (r *EntRepository) UpdateWorkflowStatus(ctx context.Context, statusID uuid.UUID, name, color *string, sortOrder *int) (*ent.WorkflowStatus, error) {
	update := r.client.WorkflowStatus.UpdateOneID(statusID)
	if name != nil {
		update = update.SetName(*name)
	}
	if color != nil {
		update = update.SetColor(*color)
	}
	if sortOrder != nil {
		update = update.SetSortOrder(*sortOrder)
	}
	return update.Save(ctx)
}

// DeleteWorkflowStatus удаляет статус workflow по ID.
func (r *EntRepository) DeleteWorkflowStatus(ctx context.Context, statusID uuid.UUID) error {
	return r.client.WorkflowStatus.DeleteOneID(statusID).Exec(ctx)
}

// FindDefaultStatus возвращает дефолтный статус workflow для проекта.
func (r *EntRepository) FindDefaultStatus(ctx context.Context, projectID uuid.UUID) (*ent.WorkflowStatus, error) {
	return r.client.WorkflowStatus.Query().
		Where(
			workflowstatus.HasProjectWith(project.IDEQ(projectID)),
			workflowstatus.IsDefaultEQ(true),
		).
		Only(ctx)
}

// ListWorkflowTransitions возвращает все разрешённые переходы для проекта.
func (r *EntRepository) ListWorkflowTransitions(ctx context.Context, projectID uuid.UUID) ([]*ent.WorkflowTransition, error) {
	return r.client.WorkflowTransition.Query().
		Where(
			workflowtransition.HasFromStatusWith(
				workflowstatus.HasProjectWith(project.IDEQ(projectID)),
			),
		).
		WithFromStatus().
		WithToStatus().
		All(ctx)
}

// CreateWorkflowTransition создаёт новый разрешённый переход между двумя статусами.
func (r *EntRepository) CreateWorkflowTransition(ctx context.Context, fromStatusID, toStatusID uuid.UUID) (*ent.WorkflowTransition, error) {
	return r.client.WorkflowTransition.Create().
		SetFromStatusID(fromStatusID).
		SetToStatusID(toStatusID).
		Save(ctx)
}

// DeleteWorkflowTransition удаляет переход workflow по ID.
func (r *EntRepository) DeleteWorkflowTransition(ctx context.Context, transitionID uuid.UUID) error {
	return r.client.WorkflowTransition.DeleteOneID(transitionID).Exec(ctx)
}

// FindTransition проверяет существование разрешённого перехода между двумя статусами.
func (r *EntRepository) FindTransition(ctx context.Context, fromStatusID, toStatusID uuid.UUID) (*ent.WorkflowTransition, error) {
	return r.client.WorkflowTransition.Query().
		Where(
			workflowtransition.HasFromStatusWith(workflowstatus.IDEQ(fromStatusID)),
			workflowtransition.HasToStatusWith(workflowstatus.IDEQ(toStatusID)),
		).
		Only(ctx)
}

// SeedDefaultWorkflow создаёт дефолтные статусы и переходы для нового проекта.
// Статусы: To Do (todo), In Progress (in_progress), In Review (in_progress), Done (done).
// Переходы: ToDo->InProgress, InProgress->InReview, InReview->Done,
// InProgress->ToDo (назад), InReview->InProgress (назад).
func (r *EntRepository) SeedDefaultWorkflow(ctx context.Context, projectID uuid.UUID) error {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return fmt.Errorf("начало транзакции: %w", err)
	}

	// Создаём статусы
	todo, err := tx.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(projectID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("создание статуса To Do: %w", err)
	}

	inProgress, err := tx.WorkflowStatus.Create().
		SetName("In Progress").
		SetCategory(workflowstatus.CategoryInProgress).
		SetColor("#3B82F6").
		SetSortOrder(1).
		SetProjectID(projectID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("создание статуса In Progress: %w", err)
	}

	inReview, err := tx.WorkflowStatus.Create().
		SetName("In Review").
		SetCategory(workflowstatus.CategoryInProgress).
		SetColor("#8B5CF6").
		SetSortOrder(2).
		SetProjectID(projectID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("создание статуса In Review: %w", err)
	}

	done, err := tx.WorkflowStatus.Create().
		SetName("Done").
		SetCategory(workflowstatus.CategoryDone).
		SetColor("#10B981").
		SetSortOrder(3).
		SetProjectID(projectID).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("создание статуса Done: %w", err)
	}

	// Создаём переходы any-to-any (Kanban — свободное перемещение между любыми статусами)
	allStatuses := []*ent.WorkflowStatus{todo, inProgress, inReview, done}
	for _, from := range allStatuses {
		for _, to := range allStatuses {
			if from.ID != to.ID {
				_, err := tx.WorkflowTransition.Create().
					SetFromStatusID(from.ID).
					SetToStatusID(to.ID).
					Save(ctx)
				if err != nil {
					_ = tx.Rollback()
					return fmt.Errorf("создание перехода: %w", err)
				}
			}
		}
	}

	return tx.Commit()
}
