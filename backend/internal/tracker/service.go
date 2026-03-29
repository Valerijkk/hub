// Package tracker реализует трекер задач: CRUD, Kanban-доски, workflow, комментарии, учёт времени.
package tracker

import (
	"context"

	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
)

// WIPChecker проверяет WIP-лимиты для статусов на Kanban-доске.
type WIPChecker interface {
	// CheckWIPLimit проверяет, не превышен ли WIP-лимит для указанного статуса.
	CheckWIPLimit(ctx context.Context, statusID uuid.UUID) error
}

// TrackerService реализует бизнес-логику модуля трекера задач.
// Оркестрирует вызовы репозитория, workflow-движка, логирование активности и outbox-события.
type TrackerService struct {
	repo       Repository
	workflow   *WorkflowEngine
	wipChecker WIPChecker
}

// NewTrackerService создаёт новый сервис трекера.
// wipChecker может быть nil (WIP-лимиты не будут проверяться).
func NewTrackerService(repo Repository, wipChecker WIPChecker) *TrackerService {
	wf := NewWorkflowEngine(repo)
	return &TrackerService{repo: repo, workflow: wf, wipChecker: wipChecker}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// buildFieldChanges сравнивает существующую задачу с запросом обновления и возвращает список изменений.
func buildFieldChanges(existing *ent.Issue, req UpdateIssueRequest) []FieldChange {
	var changes []FieldChange

	if req.Title != nil && *req.Title != existing.Title {
		changes = append(changes, FieldChange{Field: "title", OldValue: existing.Title, NewValue: *req.Title})
	}
	if req.Priority != nil && *req.Priority != string(existing.Priority) {
		changes = append(changes, FieldChange{Field: "priority", OldValue: string(existing.Priority), NewValue: *req.Priority})
	}
	if req.StatusID != nil {
		oldStatus := ""
		if existing.Edges.Status != nil {
			oldStatus = existing.Edges.Status.ID.String()
		}
		if *req.StatusID != oldStatus {
			changes = append(changes, FieldChange{Field: "status", OldValue: oldStatus, NewValue: *req.StatusID})
		}
	}
	if req.AssigneeID != nil {
		oldAssignee := ""
		if existing.Edges.Assignee != nil {
			oldAssignee = existing.Edges.Assignee.ID.String()
		}
		if *req.AssigneeID != oldAssignee {
			changes = append(changes, FieldChange{Field: "assignee", OldValue: oldAssignee, NewValue: *req.AssigneeID})
		}
	}

	return changes
}

// extractPlainText извлекает текст из Plate editor JSON для поискового индекса.
func extractPlainText(descJSON map[string]interface{}) string {
	// Упрощённое извлечение: обходим children и собираем text
	children, ok := descJSON["children"].([]interface{})
	if !ok {
		return ""
	}
	var result string
	for _, child := range children {
		node, ok := child.(map[string]interface{})
		if !ok {
			continue
		}
		if text, ok := node["text"].(string); ok {
			result += text
		}
		// Рекурсивно обходим вложенные children
		if innerChildren, ok := node["children"].([]interface{}); ok {
			for _, inner := range innerChildren {
				if innerNode, ok := inner.(map[string]interface{}); ok {
					if text, ok := innerNode["text"].(string); ok {
						result += text
					}
				}
			}
		}
	}
	return result
}
