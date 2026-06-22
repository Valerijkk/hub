package tracker_test

// Тесты сервиса трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/schema"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// TestToCommentResponse проверяет сценарий: to comment response.
func TestToCommentResponse(t *testing.T) {
	now := time.Now().UTC()
	authorID := uuid.New()
	c := &ent.Comment{
		ID:        uuid.New(),
		BodyJSON:  map[string]interface{}{"type": "doc"},
		BodyText:  "Looks good",
		CreatedAt: now,
		UpdatedAt: now,
		Edges: ent.CommentEdges{
			Author: &ent.User{ID: authorID, Name: "Reviewer", Email: "r@test.com", AvatarURL: "r.png"},
		},
	}

	resp := tracker.ExportToCommentResponse(c)
	require.NotNil(t, resp)
	assert.Equal(t, "Looks good", resp.BodyText)
	require.NotNil(t, resp.Author)
	assert.Equal(t, "Reviewer", resp.Author.Name)
}

// TestToActivityLogResponse проверяет сценарий: to activity log response.
func TestToActivityLogResponse(t *testing.T) {
	now := time.Now().UTC()
	userID := uuid.New()
	a := &ent.ActivityLog{
		ID:     uuid.New(),
		Action: "status_changed",
		Changes: []schema.FieldChange{
			{Field: "status", OldValue: "todo", NewValue: "done"},
		},
		CreatedAt: now,
		Edges: ent.ActivityLogEdges{
			User: &ent.User{ID: userID, Name: "Actor", Email: "a@test.com"},
		},
	}

	resp := tracker.ExportToActivityLogResponse(a)
	require.NotNil(t, resp)
	assert.Equal(t, "status_changed", resp.Action)
	assert.Equal(t, "Actor", resp.User.Name)
	require.Len(t, resp.Changes, 1)
}

// TestToTimeLogResponse проверяет сценарий: to time log response.
func TestToTimeLogResponse(t *testing.T) {
	now := time.Now().UTC()
	userID := uuid.New()
	tl := &ent.TimeLog{
		ID:          uuid.New(),
		Minutes:     45,
		Description: "debugging",
		CreatedAt:   now,
		Edges: ent.TimeLogEdges{
			User: &ent.User{ID: userID, Name: "Dev", Email: "d@test.com"},
		},
	}

	resp := tracker.ExportToTimeLogResponse(tl)
	require.NotNil(t, resp)
	assert.Equal(t, 45, resp.Minutes)
	assert.Equal(t, "Dev", resp.User.Name)
	assert.Equal(t, "debugging", resp.Description)
}

// TestToLabelGroupResponse проверяет сценарий: to label group response.
func TestToLabelGroupResponse(t *testing.T) {
	g := &ent.LabelGroup{
		ID:   uuid.New(),
		Name: "Priority",
	}
	resp := tracker.ExportToLabelGroupResponse(g)
	assert.Equal(t, "Priority", resp.Name)
}

// TestToBoardResponse проверяет сценарий: to board response.
func TestToBoardResponse(t *testing.T) {
	now := time.Now().UTC()
	b := &ent.Board{
		ID:        uuid.New(),
		Name:      "Main",
		Filters:   map[string]interface{}{"sprint": "S1"},
		IsDefault: true,
		CreatedAt: now,
	}
	resp := tracker.ExportToBoardResponse(b)
	assert.Equal(t, "Main", resp.Name)
	assert.True(t, resp.IsDefault)
}

// TestBuildSwimlanes_ByPriority проверяет сценарий: build swimlanes by priority.
func TestBuildSwimlanes_ByPriority(t *testing.T) {
	statusID := uuid.New()
	statuses := []*ent.WorkflowStatus{
		{
			ID: statusID, Name: "Todo", Category: workflowstatus.CategoryTodo,
			Edges: ent.WorkflowStatusEdges{
				Issues: []*ent.Issue{
					{ID: uuid.New(), IssueKey: "HUB-1", Title: "A", IssueType: issue.IssueTypeTask, Priority: issue.PriorityHigh},
					{ID: uuid.New(), IssueKey: "HUB-2", Title: "B", IssueType: issue.IssueTypeBug, Priority: issue.PriorityLow},
				},
			},
		},
	}

	lanes := tracker.ExportBuildSwimlanes(statuses, "priority")
	require.NotEmpty(t, lanes)

	keys := make(map[string]bool)
	for _, lane := range lanes {
		keys[lane.Key] = true
	}
	assert.True(t, keys["high"])
	assert.True(t, keys["low"])
}

// TestSwimlaneKeyLabel_AssigneeUnassigned проверяет сценарий: swimlane key label assignee unassigned.
func TestSwimlaneKeyLabel_AssigneeUnassigned(t *testing.T) {
	key, label := tracker.ExportSwimlaneKeyLabel(&ent.Issue{IssueType: issue.IssueTypeTask}, "assignee")
	assert.Equal(t, "unassigned", key)
	assert.Equal(t, "Не назначен", label)
}

// TestSwimlaneKeyLabel_AssigneeWithUser проверяет сценарий: swimlane key label assignee with user.
func TestSwimlaneKeyLabel_AssigneeWithUser(t *testing.T) {
	userID := uuid.New()
	key, label := tracker.ExportSwimlaneKeyLabel(&ent.Issue{
		Edges: ent.IssueEdges{Assignee: &ent.User{ID: userID, Name: "Bob"}},
	}, "assignee")
	assert.Equal(t, userID.String(), key)
	assert.Equal(t, "Bob", label)
}

// TestToIssueLinkResponse_WithoutRelated проверяет сценарий: to issue link response without related.
func TestToIssueLinkResponse_WithoutRelated(t *testing.T) {
	link := &ent.IssueLink{
		ID:       uuid.New(),
		LinkType: issuelink.LinkTypeRelatesTo,
	}
	resp := tracker.ExportToIssueLinkResponse(link, nil)
	require.NotNil(t, resp)
	assert.Equal(t, "relates_to", resp.LinkType)
	assert.Empty(t, resp.Issue.ID)
}

// TestToWorkflowStatusResponse проверяет сценарий: to workflow status response.
func TestToWorkflowStatusResponse(t *testing.T) {
	st := &ent.WorkflowStatus{
		ID: uuid.New(), Name: "Review", Category: workflowstatus.CategoryInProgress,
		Color: "#fff", SortOrder: 2, IsDefault: false,
	}
	resp := tracker.ExportToWorkflowStatusResponse(st)
	require.NotNil(t, resp)
	assert.Equal(t, "Review", resp.Name)
	assert.Equal(t, "in_progress", resp.Category)
}
