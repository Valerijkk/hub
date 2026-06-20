package tracker

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func TestBuildFieldChanges_DetectsDiffs(t *testing.T) {
	statusID := uuid.New()
	assigneeID := uuid.New()
	existing := &ent.Issue{
		Title:    "Old title",
		Priority: issue.PriorityMedium,
		Edges: ent.IssueEdges{
			Status:   &ent.WorkflowStatus{ID: statusID},
			Assignee: &ent.User{ID: assigneeID},
		},
	}

	newTitle := "New title"
	newPriority := "high"
	newStatus := uuid.New().String()
	newAssignee := uuid.New().String()
	req := UpdateIssueRequest{
		Title:      &newTitle,
		Priority:   &newPriority,
		StatusID:   &newStatus,
		AssigneeID: &newAssignee,
		Version:    1,
	}

	changes := buildFieldChanges(existing, req)
	require.Len(t, changes, 4)

	fields := make(map[string]FieldChange, len(changes))
	for _, c := range changes {
		fields[c.Field] = c
	}
	assert.Equal(t, "Old title", fields["title"].OldValue)
	assert.Equal(t, "New title", fields["title"].NewValue)
	assert.Equal(t, string(issue.PriorityMedium), fields["priority"].OldValue)
	assert.Equal(t, statusID.String(), fields["status"].OldValue)
	assert.Equal(t, assigneeID.String(), fields["assignee"].OldValue)
}

func TestBuildFieldChanges_NoChangesWhenSame(t *testing.T) {
	statusID := uuid.New()
	statusStr := statusID.String()
	existing := &ent.Issue{
		Title:    "Same",
		Priority: issue.PriorityLow,
		Edges: ent.IssueEdges{
			Status: &ent.WorkflowStatus{ID: statusID, Category: workflowstatus.CategoryTodo},
		},
	}
	priority := "low"
	req := UpdateIssueRequest{Title: &existing.Title, Priority: &priority, StatusID: &statusStr, Version: 1}

	assert.Empty(t, buildFieldChanges(existing, req))
}

func TestExtractPlainText_PlateJSON(t *testing.T) {
	desc := map[string]interface{}{
		"type": "doc",
		"children": []interface{}{
			map[string]interface{}{
				"type": "p",
				"children": []interface{}{
					map[string]interface{}{"text": "Hello "},
					map[string]interface{}{"text": "world"},
				},
			},
			map[string]interface{}{"text": "!"},
		},
	}

	assert.Equal(t, "Hello world!", extractPlainText(desc))
}

func TestExtractPlainText_InvalidJSONReturnsEmpty(t *testing.T) {
	assert.Equal(t, "", extractPlainText(map[string]interface{}{"type": "doc"}))
	assert.Equal(t, "", extractPlainText(nil))
}
