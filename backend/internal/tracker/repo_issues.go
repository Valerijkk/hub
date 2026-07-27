package tracker

import (
	"context"
	"fmt"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/predicate"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// issueProjectIDEQ — A8-001 query rewrite: filter by FK column `project_issues`
// instead of EXISTS via HasProjectWith. Same semantics for M2O required edge;
// once composite index (project_issues, deleted_at) lands (NEEDS_APPROVAL),
// ListIssues / board hot path can use Index Scan.
func issueProjectIDEQ(projectID uuid.UUID) predicate.Issue {
	return predicate.Issue(func(s *sql.Selector) {
		s.Where(sql.EQ(s.C(issue.ProjectColumn), projectID))
	})
}

// ---------------------------------------------------------------------------
// CRUD задач
// ---------------------------------------------------------------------------

// CreateIssue создаёт новую задачу в проекте. Атомарно инкрементирует счётчик
// задач проекта, вычисляет issue_key, публикует outbox-событие.
func (r *EntRepository) CreateIssue(ctx context.Context, projectID uuid.UUID, input CreateIssueInput) (*ent.Issue, error) {
	tx, err := r.client.Tx(ctx)
	if err != nil {
		return nil, fmt.Errorf("начало транзакции: %w", err)
	}

	// Атомарный инкремент счётчика задач проекта
	p, err := tx.Project.UpdateOneID(projectID).
		AddTaskCounter(1).
		Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("инкремент счётчика задач: %w", err)
	}

	issueNumber := p.TaskCounter
	issueKey := fmt.Sprintf("%s-%d", p.Key, issueNumber)

	// Определяем статус: если не указан, берём дефолтный
	statusID := input.StatusID
	if statusID == nil {
		def, err := r.findDefaultStatusTx(ctx, tx, projectID)
		if err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("получение дефолтного статуса: %w", err)
		}
		statusID = &def.ID
	}

	// Создание задачи
	create := tx.Issue.Create().
		SetIssueNumber(issueNumber).
		SetIssueKey(issueKey).
		SetTitle(input.Title).
		SetIssueType(issue.IssueType(input.IssueType)).
		SetProjectID(projectID).
		SetStatusID(*statusID).
		SetReporterID(input.ReporterID)

	if input.Priority != "" {
		create = create.SetPriority(issue.Priority(input.Priority))
	}
	if input.DescriptionJSON != nil {
		create = create.SetDescriptionJSON(input.DescriptionJSON)
	}
	if input.DescriptionText != "" {
		create = create.SetDescriptionText(input.DescriptionText)
	}
	if input.AssigneeID != nil {
		create = create.SetAssigneeID(*input.AssigneeID)
	}
	if input.ParentID != nil {
		create = create.SetParentID(*input.ParentID)
	}
	if len(input.Labels) > 0 {
		create = create.AddLabelIDs(input.Labels...)
	}
	if input.CustomFields != nil {
		create = create.SetCustomFields(input.CustomFields)
	}
	if input.EstimateMinutes != nil {
		create = create.SetEstimateMinutes(*input.EstimateMinutes)
	}
	if input.StoryPoints != nil {
		create = create.SetStoryPoints(*input.StoryPoints)
	}
	if input.Checklist != nil {
		create = create.SetChecklist(input.Checklist)
	}
	if input.DueDate != nil {
		create = create.SetDueDate(*input.DueDate)
	}

	created, err := create.Save(ctx)
	if err != nil {
		_ = tx.Rollback()
		return nil, fmt.Errorf("создание задачи: %w", err)
	}

	// Публикуем outbox-событие
	if r.publisher != nil {
		err = r.publisher.PublishInTx(ctx, tx, fmt.Sprintf("project.%s.issue.created", projectID.String()), IssueCreatedEvent{
			IssueID:     created.ID.String(),
			IssueKey:    issueKey,
			ProjectID:   projectID.String(),
			Title:       input.Title,
			IssueType:   input.IssueType,
			CreatedByID: input.ReporterID.String(),
		})
		if err != nil {
			_ = tx.Rollback()
			return nil, fmt.Errorf("публикация события: %w", err)
		}
	}

	if err = tx.Commit(); err != nil {
		return nil, fmt.Errorf("фиксация транзакции: %w", err)
	}

	// Перезагружаем с eager loading
	return r.FindIssueByID(ctx, created.ID)
}

// findDefaultStatusTx ищет дефолтный статус проекта внутри транзакции.
func (r *EntRepository) findDefaultStatusTx(ctx context.Context, tx *ent.Tx, projectID uuid.UUID) (*ent.WorkflowStatus, error) {
	return tx.WorkflowStatus.Query().
		Where(
			workflowstatus.HasProjectWith(project.IDEQ(projectID)),
			workflowstatus.IsDefaultEQ(true),
		).
		Only(ctx)
}

// FindIssueByKey возвращает задачу по ключу внутри projectID (A3-002 IDOR guard).
// Исключает мягко удалённые задачи.
func (r *EntRepository) FindIssueByKey(ctx context.Context, projectID uuid.UUID, issueKey string) (*ent.Issue, error) {
	return r.client.Issue.Query().
		Where(
			issue.IssueKeyEQ(issueKey),
			issue.DeletedAtIsNil(),
			issue.HasProjectWith(project.ID(projectID)),
		).
		WithAssignee().
		WithReporter().
		WithStatus().
		WithLabels().
		WithParent().
		WithChildren().
		WithProject().
		Only(ctx)
}

// FindIssueByID возвращает задачу по UUID с eager loading связей.
// Исключает мягко удалённые задачи.
func (r *EntRepository) FindIssueByID(ctx context.Context, id uuid.UUID) (*ent.Issue, error) {
	return r.client.Issue.Query().
		Where(issue.IDEQ(id), issue.DeletedAtIsNil()).
		WithAssignee().
		WithReporter().
		WithStatus().
		WithLabels().
		WithParent().
		WithChildren().
		WithProject().
		Only(ctx)
}

// FindIssueInProject возвращает задачу по UUID только внутри projectID (A3-011-target IDOR guard).
// Исключает мягко удалённые задачи. Mutation: снять issueProjectIDEQ → foreign issue находится.
func (r *EntRepository) FindIssueInProject(ctx context.Context, projectID, id uuid.UUID) (*ent.Issue, error) {
	return r.client.Issue.Query().
		Where(
			issue.IDEQ(id),
			issue.DeletedAtIsNil(),
			issueProjectIDEQ(projectID),
		).
		WithAssignee().
		WithReporter().
		WithStatus().
		WithLabels().
		WithParent().
		WithChildren().
		WithProject().
		Only(ctx)
}

// ListIssues возвращает пагинированный список задач проекта с фильтрами.
// Использует курсорную пагинацию по issue_key.
func (r *EntRepository) ListIssues(ctx context.Context, projectID uuid.UUID, params ListIssuesParams) ([]*ent.Issue, int, error) {
	// Базовый запрос с фильтром по проекту (исключаем мягко удалённые)
	query := r.client.Issue.Query().
		Where(issueProjectIDEQ(projectID), issue.DeletedAtIsNil())

	// Фильтры
	if len(params.Status) > 0 {
		query = query.Where(issue.HasStatusWith(workflowstatus.NameIn(params.Status...)))
	}
	if len(params.Priority) > 0 {
		priorities := make([]issue.Priority, len(params.Priority))
		for i, p := range params.Priority {
			priorities[i] = issue.Priority(p)
		}
		query = query.Where(issue.PriorityIn(priorities...))
	}
	if params.UnassignedOnly {
		query = query.Where(issue.Not(issue.HasAssignee()))
	} else if params.Assignee != nil {
		query = query.Where(issue.HasAssigneeWith(user.IDEQ(*params.Assignee)))
	}
	if params.IssueType != "" {
		query = query.Where(issue.IssueTypeEQ(issue.IssueType(params.IssueType)))
	}
	if params.Search != "" {
		// Используем tsvector полнотекстовый поиск (search_vector создаётся индексатором).
		// Fallback на ContainsFold если search_vector отсутствует.
		query = query.Where(func(s *sql.Selector) {
			s.Where(sql.ExprP(
				"to_tsvector('simple', title || ' ' || COALESCE(description_text, '')) @@ plainto_tsquery('simple', $1)",
				params.Search,
			))
		})
	}

	// Общее количество (до пагинации)
	total, err := query.Clone().Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("подсчёт задач: %w", err)
	}

	// Курсорная пагинация: курсор — issue_key последней строки, фильтр согласован с order_by.
	if params.Cursor != "" {
		cursorIssue, err := r.client.Issue.Query().
			Where(
				issue.IssueKeyEQ(params.Cursor),
				issueProjectIDEQ(projectID),
				issue.DeletedAtIsNil(),
			).
			Only(ctx)
		if err != nil {
			if ent.IsNotFound(err) {
				return []*ent.Issue{}, total, nil
			}
			return nil, 0, fmt.Errorf("курсорная задача: %w", err)
		}
		query = applyListIssuesCursorFilter(query, cursorIssue, params.OrderBy, params.Cursor)
	}

	// Лимит
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Сортировка (issue_key ASC — стабильный tie-breaker для курсора)
	switch params.OrderBy {
	case "priority":
		query = query.Order(issue.ByPriority(), issue.ByIssueKey())
	case "updated_at":
		query = query.Order(issue.ByUpdatedAt(sql.OrderDesc()), issue.ByIssueKey())
	default:
		query = query.Order(issue.ByCreatedAt(sql.OrderDesc()), issue.ByIssueKey())
	}

	items, err := query.
		Limit(limit).
		WithAssignee().
		WithStatus().
		WithLabels().
		All(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("список задач: %w", err)
	}

	return items, total, nil
}

// applyListIssuesCursorFilter возвращает строки, идущие после курсора в порядке orderBy.
// Курсор API — issue_key последней выданной строки; tie-breaker — issue_key ASC.
func applyListIssuesCursorFilter(query *ent.IssueQuery, cursorIssue *ent.Issue, orderBy, cursor string) *ent.IssueQuery {
	keyAfter := issue.IssueKeyGT(cursor)
	switch orderBy {
	case "priority":
		return query.Where(issue.Or(
			predicate.Issue(sql.FieldGT(issue.FieldPriority, cursorIssue.Priority)),
			issue.And(issue.PriorityEQ(cursorIssue.Priority), keyAfter),
		))
	case "updated_at":
		return query.Where(issue.Or(
			issue.UpdatedAtLT(cursorIssue.UpdatedAt),
			issue.And(issue.UpdatedAtEQ(cursorIssue.UpdatedAt), keyAfter),
		))
	default:
		return query.Where(issue.Or(
			issue.CreatedAtLT(cursorIssue.CreatedAt),
			issue.And(issue.CreatedAtEQ(cursorIssue.CreatedAt), keyAfter),
		))
	}
}

// UpdateIssue обновляет задачу с оптимистичной блокировкой по полю version.
// Возвращает ошибку конфликта при несовпадении версий.
func (r *EntRepository) UpdateIssue(ctx context.Context, issueID uuid.UUID, input UpdateIssueInput) (*ent.Issue, error) {
	update := r.client.Issue.Update().
		Where(
			issue.IDEQ(issueID),
			issue.VersionEQ(input.Version),
		).
		AddVersion(1)

	if input.Title != nil {
		update = update.SetTitle(*input.Title)
	}
	if input.DescriptionJSON != nil {
		update = update.SetDescriptionJSON(*input.DescriptionJSON)
	}
	if input.DescriptionText != nil {
		update = update.SetDescriptionText(*input.DescriptionText)
	}
	if input.Priority != nil {
		update = update.SetPriority(issue.Priority(*input.Priority))
	}
	if input.AssigneeID != nil {
		update = update.SetAssigneeID(*input.AssigneeID)
	}
	if input.StatusID != nil {
		update = update.SetStatusID(*input.StatusID)
	}
	if input.Labels != nil {
		update = update.ClearLabels().AddLabelIDs(*input.Labels...)
	}
	if input.CustomFields != nil {
		update = update.SetCustomFields(*input.CustomFields)
	}
	if input.EstimateMinutes != nil {
		update = update.SetEstimateMinutes(*input.EstimateMinutes)
	}
	if input.StoryPoints != nil {
		update = update.SetStoryPoints(*input.StoryPoints)
	}
	if input.Checklist != nil {
		update = update.SetChecklist(*input.Checklist)
	}
	if input.DueDate != nil {
		update = update.SetDueDate(*input.DueDate)
	}

	affected, err := update.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление задачи: %w", err)
	}
	if affected == 0 {
		return nil, apperrors.NewConflict("version mismatch: задача была изменена другим пользователем")
	}

	updated, err := r.FindIssueByID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	// Публикуем outbox-событие обновления задачи.
	if r.publisher != nil && updated.Edges.Project != nil {
		projectID := updated.Edges.Project.ID.String()
		_ = r.publisher.Publish(ctx, r.client, fmt.Sprintf("project.%s.issue.updated", projectID), IssueUpdatedEvent{
			IssueID:   issueID.String(),
			IssueKey:  updated.IssueKey,
			ProjectID: projectID,
		})

		// Отдельное событие при смене назначенного.
		if input.AssigneeID != nil {
			_ = r.publisher.Publish(ctx, r.client, fmt.Sprintf("project.%s.issue.assigned", projectID), IssueAssignedEvent{
				IssueID:    issueID.String(),
				IssueKey:   updated.IssueKey,
				ProjectID:  projectID,
				AssigneeID: input.AssigneeID.String(),
			})
		}
	}

	return updated, nil
}

// DeleteIssue выполняет мягкое удаление задачи (устанавливает deleted_at).
func (r *EntRepository) DeleteIssue(ctx context.Context, issueID uuid.UUID) error {
	return r.client.Issue.UpdateOneID(issueID).
		SetDeletedAt(time.Now().UTC()).
		Exec(ctx)
}

// MoveIssue перемещает задачу в другой статус и обновляет ранг (для Kanban drag-and-drop).
// Публикует outbox-событие project.{projectID}.issue.moved.
func (r *EntRepository) MoveIssue(ctx context.Context, issueID uuid.UUID, statusID uuid.UUID, rank string) (*ent.Issue, error) {
	// Получаем текущий статус до обновления для события.
	current, err := r.FindIssueByID(ctx, issueID)
	if err != nil {
		return nil, fmt.Errorf("получение текущей задачи: %w", err)
	}
	oldStatusID := ""
	if current.Edges.Status != nil {
		oldStatusID = current.Edges.Status.ID.String()
	}

	err = r.client.Issue.UpdateOneID(issueID).
		SetStatusID(statusID).
		SetRank(rank).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("перемещение задачи: %w", err)
	}

	moved, err := r.FindIssueByID(ctx, issueID)
	if err != nil {
		return nil, err
	}

	// Публикуем outbox-событие перемещения задачи.
	if r.publisher != nil && moved.Edges.Project != nil {
		projectID := moved.Edges.Project.ID.String()
		_ = r.publisher.Publish(ctx, r.client, fmt.Sprintf("project.%s.issue.moved", projectID), IssueMovedEvent{
			IssueID:     issueID.String(),
			IssueKey:    moved.IssueKey,
			ProjectID:   projectID,
			OldStatusID: oldStatusID,
			NewStatusID: statusID.String(),
			NewRank:     rank,
		})
	}

	return moved, nil
}
