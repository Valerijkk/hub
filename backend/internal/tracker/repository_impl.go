package tracker

import (
	"context"
	"fmt"
	"math"
	"time"

	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/activitylog"
	"github.com/valerii/hub/backend/ent/board"
	"github.com/valerii/hub/backend/ent/comment"
	"github.com/valerii/hub/backend/ent/customfielddefinition"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/label"
	"github.com/valerii/hub/backend/ent/labelgroup"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/reaction"
	"github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/ent/timelog"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/ent/workflowtransition"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
)

// ---------------------------------------------------------------------------
// Issue CRUD
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

// FindIssueByKey возвращает задачу по её ключу (например "HUB-1") с eager loading связей.
// Исключает мягко удалённые задачи.
func (r *EntRepository) FindIssueByKey(ctx context.Context, issueKey string) (*ent.Issue, error) {
	return r.client.Issue.Query().
		Where(issue.IssueKeyEQ(issueKey), issue.DeletedAtIsNil()).
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

// ListIssues возвращает пагинированный список задач проекта с фильтрами.
// Использует курсорную пагинацию по issue_key.
func (r *EntRepository) ListIssues(ctx context.Context, projectID uuid.UUID, params ListIssuesParams) ([]*ent.Issue, int, error) {
	// Базовый запрос с фильтром по проекту (исключаем мягко удалённые)
	query := r.client.Issue.Query().
		Where(issue.HasProjectWith(project.IDEQ(projectID)), issue.DeletedAtIsNil())

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
	if params.Assignee != nil {
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

	// Курсорная пагинация
	if params.Cursor != "" {
		query = query.Where(issue.IssueKeyGT(params.Cursor))
	}

	// Лимит
	limit := params.Limit
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	// Сортировка
	switch params.OrderBy {
	case "priority":
		query = query.Order(ent.Asc(issue.FieldPriority))
	case "updated_at":
		query = query.Order(ent.Desc(issue.FieldUpdatedAt))
	default:
		query = query.Order(ent.Desc(issue.FieldCreatedAt))
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
		SetDeletedAt(time.Now()).
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

// ---------------------------------------------------------------------------
// Issue links
// ---------------------------------------------------------------------------

// CreateIssueLink создаёт направленную связь между двумя задачами.
// Проверяет отсутствие дубликата. Публикует outbox-событие project.{projectID}.issue.link_created.
func (r *EntRepository) CreateIssueLink(ctx context.Context, sourceID, targetID uuid.UUID, linkType string, customType string) (*ent.IssueLink, error) {
	// Проверяем дубликат
	exists, err := r.client.IssueLink.Query().
		Where(
			issuelink.HasSourceWith(issue.IDEQ(sourceID)),
			issuelink.HasTargetWith(issue.IDEQ(targetID)),
		).
		Exist(ctx)
	if err != nil {
		return nil, fmt.Errorf("проверка дубликата связи: %w", err)
	}
	if exists {
		return nil, apperrors.NewConflict("связь между этими задачами уже существует")
	}

	create := r.client.IssueLink.Create().
		SetSourceID(sourceID).
		SetTargetID(targetID).
		SetLinkType(issuelink.LinkType(linkType))

	if customType != "" {
		create = create.SetCustomType(customType)
	}

	link, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("создание связи: %w", err)
	}

	// Публикуем outbox-событие создания связи.
	if r.publisher != nil {
		source, qErr := r.client.Issue.Query().
			Where(issue.IDEQ(sourceID)).
			WithProject().
			Only(ctx)
		if qErr == nil && source.Edges.Project != nil {
			projectID := source.Edges.Project.ID.String()
			_ = r.publisher.Publish(ctx, r.client, fmt.Sprintf("project.%s.issue.link_created", projectID), IssueLinkCreatedEvent{
				SourceID:  sourceID.String(),
				TargetID:  targetID.String(),
				LinkType:  linkType,
				ProjectID: projectID,
			})
		}
	}

	return link, nil
}

// ListIssueLinks возвращает все связи (входящие и исходящие) для задачи.
func (r *EntRepository) ListIssueLinks(ctx context.Context, issueID uuid.UUID) ([]*ent.IssueLink, error) {
	// Исходящие связи
	outgoing, err := r.client.IssueLink.Query().
		Where(issuelink.HasSourceWith(issue.IDEQ(issueID))).
		WithTarget().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("получение исходящих связей: %w", err)
	}

	// Входящие связи
	incoming, err := r.client.IssueLink.Query().
		Where(issuelink.HasTargetWith(issue.IDEQ(issueID))).
		WithSource().
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("получение входящих связей: %w", err)
	}

	// Объединяем
	all := make([]*ent.IssueLink, 0, len(outgoing)+len(incoming))
	all = append(all, outgoing...)
	all = append(all, incoming...)
	return all, nil
}

// DeleteIssueLink удаляет связь между задачами.
func (r *EntRepository) DeleteIssueLink(ctx context.Context, linkID uuid.UUID) error {
	return r.client.IssueLink.DeleteOneID(linkID).Exec(ctx)
}

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

// ---------------------------------------------------------------------------
// Comments
// ---------------------------------------------------------------------------

// CreateComment добавляет комментарий к задаче.
// Публикует outbox-событие project.{projectID}.issue.comment_created.
func (r *EntRepository) CreateComment(ctx context.Context, issueID, authorID uuid.UUID, input CreateCommentInput) (*ent.Comment, error) {
	create := r.client.Comment.Create().
		SetIssueID(issueID).
		SetAuthorID(authorID).
		SetBodyJSON(input.BodyJSON).
		SetBodyText(input.BodyText).
		SetIsInternal(input.IsInternal)

	if len(input.Mentions) > 0 {
		create = create.SetMentions(input.Mentions)
	}

	c, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("создание комментария: %w", err)
	}

	// Публикуем outbox-событие создания комментария.
	if r.publisher != nil {
		iss, err := r.client.Issue.Query().
			Where(issue.IDEQ(issueID)).
			WithProject().
			Only(ctx)
		if err == nil && iss.Edges.Project != nil {
			projectID := iss.Edges.Project.ID.String()
			_ = r.publisher.Publish(ctx, r.client, fmt.Sprintf("project.%s.issue.comment_created", projectID), CommentCreatedEvent{
				CommentID: c.ID.String(),
				IssueID:   issueID.String(),
				IssueKey:  iss.IssueKey,
				ProjectID: projectID,
				AuthorID:  authorID.String(),
				Mentions:  input.Mentions,
			})
		}
	}

	// Перезагружаем с автором
	return r.client.Comment.Query().
		Where(comment.IDEQ(c.ID)).
		WithAuthor().
		WithReactions().
		Only(ctx)
}

// ListComments возвращает все комментарии к задаче с авторами и реакциями.
func (r *EntRepository) ListComments(ctx context.Context, issueID uuid.UUID) ([]*ent.Comment, error) {
	return r.client.Comment.Query().
		Where(comment.HasIssueWith(issue.IDEQ(issueID))).
		WithAuthor().
		WithReactions(func(q *ent.ReactionQuery) {
			q.WithUser()
		}).
		Order(ent.Asc(comment.FieldCreatedAt)).
		All(ctx)
}

// UpdateComment обновляет содержимое комментария.
func (r *EntRepository) UpdateComment(ctx context.Context, commentID uuid.UUID, input UpdateCommentInput) (*ent.Comment, error) {
	c, err := r.client.Comment.UpdateOneID(commentID).
		SetBodyJSON(input.BodyJSON).
		SetBodyText(input.BodyText).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление комментария: %w", err)
	}
	return c, nil
}

// DeleteComment удаляет комментарий по ID.
func (r *EntRepository) DeleteComment(ctx context.Context, commentID uuid.UUID) error {
	return r.client.Comment.DeleteOneID(commentID).Exec(ctx)
}

// ---------------------------------------------------------------------------
// Reactions
// ---------------------------------------------------------------------------

// ToggleReaction добавляет или удаляет реакцию (toggle). Возвращает true если добавлена, false если удалена.
func (r *EntRepository) ToggleReaction(ctx context.Context, commentID, userID uuid.UUID, emoji string) (bool, error) {
	// Ищем существующую реакцию
	existing, err := r.client.Reaction.Query().
		Where(
			reaction.HasCommentWith(comment.IDEQ(commentID)),
			reaction.HasUserWith(user.IDEQ(userID)),
			reaction.EmojiEQ(emoji),
		).
		Only(ctx)

	if err != nil && !ent.IsNotFound(err) {
		return false, fmt.Errorf("поиск реакции: %w", err)
	}

	if existing != nil {
		// Удаляем (toggle off)
		if err := r.client.Reaction.DeleteOneID(existing.ID).Exec(ctx); err != nil {
			return false, fmt.Errorf("удаление реакции: %w", err)
		}
		return false, nil
	}

	// Создаём (toggle on)
	_, err = r.client.Reaction.Create().
		SetCommentID(commentID).
		SetUserID(userID).
		SetEmoji(emoji).
		Save(ctx)
	if err != nil {
		return false, fmt.Errorf("создание реакции: %w", err)
	}
	return true, nil
}

// ---------------------------------------------------------------------------
// Activity log
// ---------------------------------------------------------------------------

// LogActivity записывает запись в лог активности задачи.
func (r *EntRepository) LogActivity(ctx context.Context, issueID, userID uuid.UUID, action string, changes []FieldChange) (*ent.ActivityLog, error) {
	// Конвертируем FieldChange из пакета tracker в schema.FieldChange
	schemaChanges := make([]schema.FieldChange, len(changes))
	for i, c := range changes {
		schemaChanges[i] = schema.FieldChange{
			Field:    c.Field,
			OldValue: c.OldValue,
			NewValue: c.NewValue,
		}
	}

	create := r.client.ActivityLog.Create().
		SetIssueID(issueID).
		SetUserID(userID).
		SetAction(action)

	if len(schemaChanges) > 0 {
		create = create.SetChanges(schemaChanges)
	}

	return create.Save(ctx)
}

// ListActivities возвращает все записи лога активности задачи.
func (r *EntRepository) ListActivities(ctx context.Context, issueID uuid.UUID) ([]*ent.ActivityLog, error) {
	return r.client.ActivityLog.Query().
		Where(activitylog.HasIssueWith(issue.IDEQ(issueID))).
		WithUser().
		Order(ent.Desc(activitylog.FieldCreatedAt)).
		All(ctx)
}

// ---------------------------------------------------------------------------
// Time tracking
// ---------------------------------------------------------------------------

// LogTime записывает вручную введённое время на задачу.
func (r *EntRepository) LogTime(ctx context.Context, issueID, userID uuid.UUID, minutes int, description string) (*ent.TimeLog, error) {
	create := r.client.TimeLog.Create().
		SetIssueID(issueID).
		SetUserID(userID).
		SetMinutes(minutes)

	if description != "" {
		create = create.SetDescription(description)
	}

	tl, err := create.Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("запись времени: %w", err)
	}

	// Обновляем суммарное время на задаче
	err = r.client.Issue.UpdateOneID(issueID).
		AddTimeSpentMinutes(minutes).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление суммарного времени: %w", err)
	}

	return tl, nil
}

// StartTimer запускает таймер на задаче для пользователя.
func (r *EntRepository) StartTimer(ctx context.Context, issueID, userID uuid.UUID) (*ent.TimeLog, error) {
	now := time.Now()
	// Создаём TimeLog с started_at, но без ended_at и минимальным minutes=1
	// (minutes будет пересчитан при StopTimer)
	tl, err := r.client.TimeLog.Create().
		SetIssueID(issueID).
		SetUserID(userID).
		SetMinutes(1). // placeholder, будет пересчитано при остановке
		SetStartedAt(now).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("запуск таймера: %w", err)
	}
	return tl, nil
}

// StopTimer останавливает запущенный таймер и записывает прошедшее время.
func (r *EntRepository) StopTimer(ctx context.Context, issueID, userID uuid.UUID) (*ent.TimeLog, error) {
	// Ищем открытый таймер (ended_at is nil)
	openTimer, err := r.client.TimeLog.Query().
		Where(
			timelog.HasIssueWith(issue.IDEQ(issueID)),
			timelog.HasUserWith(user.IDEQ(userID)),
			timelog.EndedAtIsNil(),
		).
		Only(ctx)
	if err != nil {
		if ent.IsNotFound(err) {
			return nil, apperrors.NewNotFound("активный таймер не найден")
		}
		return nil, fmt.Errorf("поиск открытого таймера: %w", err)
	}

	now := time.Now()
	elapsed := int(math.Ceil(now.Sub(*openTimer.StartedAt).Minutes()))
	if elapsed < 1 {
		elapsed = 1
	}

	// Обновляем таймер
	tl, err := r.client.TimeLog.UpdateOneID(openTimer.ID).
		SetEndedAt(now).
		SetMinutes(elapsed).
		Save(ctx)
	if err != nil {
		return nil, fmt.Errorf("остановка таймера: %w", err)
	}

	// Обновляем суммарное время на задаче
	err = r.client.Issue.UpdateOneID(issueID).
		AddTimeSpentMinutes(elapsed).
		Exec(ctx)
	if err != nil {
		return nil, fmt.Errorf("обновление суммарного времени: %w", err)
	}

	return tl, nil
}

// ListTimeLogs возвращает все записи времени для задачи.
func (r *EntRepository) ListTimeLogs(ctx context.Context, issueID uuid.UUID) ([]*ent.TimeLog, error) {
	return r.client.TimeLog.Query().
		Where(timelog.HasIssueWith(issue.IDEQ(issueID))).
		WithUser().
		Order(ent.Desc(timelog.FieldCreatedAt)).
		All(ctx)
}

// ---------------------------------------------------------------------------
// Labels
// ---------------------------------------------------------------------------

// CreateLabel создаёт новую метку в проекте, опционально привязывая к группе.
func (r *EntRepository) CreateLabel(ctx context.Context, projectID uuid.UUID, name, color string, groupID *uuid.UUID) (*ent.Label, error) {
	create := r.client.Label.Create().
		SetProjectID(projectID).
		SetName(name).
		SetColor(color)

	if groupID != nil {
		create = create.SetGroupID(*groupID)
	}

	return create.Save(ctx)
}

// ListLabels возвращает все метки проекта.
func (r *EntRepository) ListLabels(ctx context.Context, projectID uuid.UUID) ([]*ent.Label, error) {
	return r.client.Label.Query().
		Where(label.HasProjectWith(project.IDEQ(projectID))).
		WithGroup().
		All(ctx)
}

// DeleteLabel удаляет метку по ID.
func (r *EntRepository) DeleteLabel(ctx context.Context, labelID uuid.UUID) error {
	return r.client.Label.DeleteOneID(labelID).Exec(ctx)
}

// CreateLabelGroup создаёт новую группу меток в проекте.
func (r *EntRepository) CreateLabelGroup(ctx context.Context, projectID uuid.UUID, name string) (*ent.LabelGroup, error) {
	return r.client.LabelGroup.Create().
		SetProjectID(projectID).
		SetName(name).
		Save(ctx)
}

// ListLabelGroups возвращает все группы меток проекта с их метками.
func (r *EntRepository) ListLabelGroups(ctx context.Context, projectID uuid.UUID) ([]*ent.LabelGroup, error) {
	return r.client.LabelGroup.Query().
		Where(labelgroup.HasProjectWith(project.IDEQ(projectID))).
		WithLabels().
		All(ctx)
}

// ---------------------------------------------------------------------------
// Boards
// ---------------------------------------------------------------------------

// CreateBoard создаёт новую доску в проекте с сохранёнными фильтрами.
func (r *EntRepository) CreateBoard(ctx context.Context, projectID uuid.UUID, name string, filters map[string]interface{}) (*ent.Board, error) {
	create := r.client.Board.Create().
		SetProjectID(projectID).
		SetName(name)

	if filters != nil {
		create = create.SetFilters(filters)
	}

	return create.Save(ctx)
}

// ListBoards возвращает все доски проекта.
func (r *EntRepository) ListBoards(ctx context.Context, projectID uuid.UUID) ([]*ent.Board, error) {
	return r.client.Board.Query().
		Where(board.HasProjectWith(project.IDEQ(projectID))).
		All(ctx)
}

// GetBoardData возвращает доску по ID для рендеринга (доска + данные).
// Применяет сохранённые фильтры доски (assignee_id, issue_type, label_ids, priority)
// к запросу задач, чтобы вернуть только релевантные задачи.
func (r *EntRepository) GetBoardData(ctx context.Context, boardID uuid.UUID) (*ent.Board, error) {
	// Загружаем доску отдельно для доступа к фильтрам.
	b, err := r.client.Board.Get(ctx, boardID)
	if err != nil {
		return nil, fmt.Errorf("загрузка доски: %w", err)
	}

	return r.client.Board.Query().
		Where(board.IDEQ(boardID)).
		WithProject(func(q *ent.ProjectQuery) {
			q.WithWorkflowStatuses(func(sq *ent.WorkflowStatusQuery) {
				sq.Order(ent.Asc(workflowstatus.FieldSortOrder))
				sq.WithIssues(func(iq *ent.IssueQuery) {
					iq.Where(issue.DeletedAtIsNil())
					applyBoardFilters(iq, b.Filters)
					iq.WithAssignee()
					iq.WithLabels()
					iq.Order(ent.Asc(issue.FieldRank))
				})
			})
		}).
		Only(ctx)
}

// applyBoardFilters применяет сохранённые фильтры доски к запросу задач.
// Поддерживаемые ключи фильтров:
//   - assignee_id (string)   — UUID назначенного пользователя
//   - issue_type  (string)   — тип задачи (task, bug, story, epic)
//   - priority    (string)   — приоритет (highest, high, medium, low, lowest)
//   - label_ids   ([]string) — массив UUID меток
func applyBoardFilters(iq *ent.IssueQuery, filters map[string]interface{}) {
	if len(filters) == 0 {
		return
	}

	if assigneeID, ok := filters["assignee_id"].(string); ok && assigneeID != "" {
		uid, err := uuid.Parse(assigneeID)
		if err == nil {
			iq.Where(issue.HasAssigneeWith(user.IDEQ(uid)))
		}
	}

	if issueType, ok := filters["issue_type"].(string); ok && issueType != "" {
		iq.Where(issue.IssueTypeEQ(issue.IssueType(issueType)))
	}

	if priority, ok := filters["priority"].(string); ok && priority != "" {
		iq.Where(issue.PriorityEQ(issue.Priority(priority)))
	}

	if labelIDs, ok := filters["label_ids"].([]interface{}); ok && len(labelIDs) > 0 {
		uids := make([]uuid.UUID, 0, len(labelIDs))
		for _, raw := range labelIDs {
			if s, ok := raw.(string); ok {
				uid, err := uuid.Parse(s)
				if err == nil {
					uids = append(uids, uid)
				}
			}
		}
		if len(uids) > 0 {
			iq.Where(issue.HasLabelsWith(label.IDIn(uids...)))
		}
	}
}

// ---------------------------------------------------------------------------
// Custom fields
// ---------------------------------------------------------------------------

// CreateCustomField создаёт новое определение кастомного поля в проекте.
func (r *EntRepository) CreateCustomField(ctx context.Context, projectID uuid.UUID, input CreateCustomFieldInput) (*ent.CustomFieldDefinition, error) {
	create := r.client.CustomFieldDefinition.Create().
		SetProjectID(projectID).
		SetName(input.Name).
		SetFieldType(customfielddefinition.FieldType(input.FieldType)).
		SetRequired(input.Required)

	if len(input.Options) > 0 {
		create = create.SetOptions(input.Options)
	}

	return create.Save(ctx)
}

// ListCustomFields возвращает все определения кастомных полей проекта.
func (r *EntRepository) ListCustomFields(ctx context.Context, projectID uuid.UUID) ([]*ent.CustomFieldDefinition, error) {
	return r.client.CustomFieldDefinition.Query().
		Where(customfielddefinition.HasProjectWith(project.IDEQ(projectID))).
		Order(ent.Asc(customfielddefinition.FieldSortOrder)).
		All(ctx)
}

// DeleteCustomField удаляет определение кастомного поля по ID.
func (r *EntRepository) DeleteCustomField(ctx context.Context, fieldID uuid.UUID) error {
	return r.client.CustomFieldDefinition.DeleteOneID(fieldID).Exec(ctx)
}
