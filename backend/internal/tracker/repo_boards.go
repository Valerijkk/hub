package tracker

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/board"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/label"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// ---------------------------------------------------------------------------
// Доски
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
