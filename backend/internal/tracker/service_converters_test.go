package tracker

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/customfielddefinition"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func TestToIssueResponse_FullMapping(t *testing.T) {
	t.Helper()
	now := time.Now().UTC()
	statusID := uuid.New()
	assigneeID := uuid.New()
	reporterID := uuid.New()
	parentID := uuid.New()
	labelID := uuid.New()
	desc := map[string]interface{}{"type": "doc"}

	i := &ent.Issue{
		ID:               uuid.New(),
		IssueNumber:      42,
		IssueKey:         "HUB-42",
		Title:            "Full issue",
		IssueType:        issue.IssueTypeStory,
		Priority:         issue.PriorityHigh,
		Rank:             "aaa|zzz",
		Version:          3,
		EstimateMinutes:  intPtr(120),
		StoryPoints:      intPtr(5),
		TimeSpentMinutes: 30,
		CustomFields:     map[string]interface{}{"sprint": "S1"},
		Checklist: []map[string]interface{}{
			{"id": "c1", "text": "Do thing", "checked": true},
		},
		DescriptionJSON: desc,
		DescriptionText: "plain text",
		DueDate:         &now,
		CreatedAt:       now,
		UpdatedAt:       now,
		Edges: ent.IssueEdges{
			Status: &ent.WorkflowStatus{
				ID: statusID, Name: "In Progress", Category: workflowstatus.CategoryInProgress,
				Color: "#3B82F6", SortOrder: 1, IsDefault: false,
			},
			Assignee: &ent.User{ID: assigneeID, Name: "Assignee", Email: "a@test.com", AvatarURL: "a.png"},
			Reporter: &ent.User{ID: reporterID, Name: "Reporter", Email: "r@test.com"},
			Parent:   &ent.Issue{ID: parentID, IssueKey: "HUB-1", Title: "Parent"},
			Children: []*ent.Issue{{ID: uuid.New()}, {ID: uuid.New()}},
			Labels: []*ent.Label{
				{ID: labelID, Name: "urgent", Color: "#FF0000", Edges: ent.LabelEdges{
					Group: &ent.LabelGroup{Name: "Priority"},
				}},
			},
		},
	}

	resp := toIssueResponse(i)
	require.NotNil(t, resp)
	assert.Equal(t, i.ID.String(), resp.ID)
	assert.Equal(t, "HUB-42", resp.IssueKey)
	assert.Equal(t, "story", resp.IssueType)
	assert.Equal(t, "high", resp.Priority)
	assert.Equal(t, 3, resp.Version)
	assert.Equal(t, 2, resp.ChildrenCount)
	assert.Len(t, resp.Checklist, 1)
	assert.True(t, resp.Checklist[0].Checked)
	assert.Equal(t, "plain text", resp.DescriptionText)
	require.NotNil(t, resp.DescriptionJSON)
	assert.Equal(t, parentID.String(), resp.Parent.ID)
	assert.Len(t, resp.Labels, 1)
	assert.Equal(t, "Priority", resp.Labels[0].GroupName)
}

func TestToIssueBriefResponse_ChecklistAndStatus(t *testing.T) {
	i := &ent.Issue{
		ID:        uuid.New(),
		IssueKey:  "TST-7",
		Title:     "Brief",
		IssueType: issue.IssueTypeBug,
		Priority:  issue.PriorityCritical,
		StoryPoints: intPtr(8),
		Checklist: []map[string]interface{}{
			{"id": "x", "text": "step", "checked": false},
		},
		Edges: ent.IssueEdges{
			Status: &ent.WorkflowStatus{Name: "Done"},
		},
	}

	resp := toIssueBriefResponse(i)
	require.NotNil(t, resp)
	assert.Equal(t, "Done", resp.Status)
	require.NotNil(t, resp.StoryPoints)
	assert.Equal(t, 8, *resp.StoryPoints)
	assert.Len(t, resp.Checklist, 1)
}

func TestAggregateReactions_GroupsByEmoji(t *testing.T) {
	u1 := uuid.New()
	u2 := uuid.New()
	reactions := []*ent.Reaction{
		{Emoji: "thumbsup", Edges: ent.ReactionEdges{User: &ent.User{ID: u1, Name: "A", Email: "a@test.com"}}},
		{Emoji: "thumbsup", Edges: ent.ReactionEdges{User: &ent.User{ID: u2, Name: "B", Email: "b@test.com"}}},
		{Emoji: "heart", Edges: ent.ReactionEdges{User: &ent.User{ID: u1, Name: "A", Email: "a@test.com"}}},
	}

	result := aggregateReactions(reactions)
	require.Len(t, result, 2)

	byEmoji := make(map[string]ReactionResponse, len(result))
	for _, r := range result {
		byEmoji[r.Emoji] = r
	}
	assert.Equal(t, 2, byEmoji["thumbsup"].Count)
	assert.Len(t, byEmoji["thumbsup"].Users, 2)
	assert.Equal(t, 1, byEmoji["heart"].Count)
}

func TestToBoardDataResponse_BuildsColumns(t *testing.T) {
	statusID := uuid.New()
	issueID := uuid.New()
	board := &ent.Board{
		ID:        uuid.New(),
		Name:      "Sprint",
		Filters:   map[string]interface{}{"assignee": "me"},
		IsDefault: true,
		CreatedAt: time.Now(),
		Edges: ent.BoardEdges{
			Project: &ent.Project{
				ID:  uuid.New(),
				Key: "HUB",
				Edges: ent.ProjectEdges{
					WorkflowStatuses: []*ent.WorkflowStatus{
						{
							ID: statusID, Name: "To Do", Category: workflowstatus.CategoryTodo,
							Color: "#000", SortOrder: 0,
							Edges: ent.WorkflowStatusEdges{
								Issues: []*ent.Issue{
									{ID: issueID, IssueKey: "HUB-1", Title: "Task", IssueType: issue.IssueTypeTask, Priority: issue.PriorityMedium},
								},
							},
						},
					},
				},
			},
		},
	}

	resp := toBoardDataResponse(board)
	require.NotNil(t, resp)
	assert.Equal(t, "Sprint", resp.Board.Name)
	require.Len(t, resp.Columns, 1)
	assert.Equal(t, "To Do", resp.Columns[0].Status.Name)
	require.Len(t, resp.Columns[0].Issues, 1)
	assert.Equal(t, "HUB-1", resp.Columns[0].Issues[0].IssueKey)
}

func TestToIssueLinkResponse_WithRelatedIssue(t *testing.T) {
	link := &ent.IssueLink{
		ID:         uuid.New(),
		LinkType:   issuelink.LinkTypeBlocks,
		CustomType: "custom",
	}
	related := &ent.Issue{ID: uuid.New(), IssueKey: "HUB-9", Title: "Blocked", IssueType: issue.IssueTypeTask, Priority: issue.PriorityLow}

	resp := toIssueLinkResponse(link, related)
	require.NotNil(t, resp)
	assert.Equal(t, "blocks", resp.LinkType)
	assert.Equal(t, "HUB-9", resp.Issue.IssueKey)
}

func intPtr(v int) *int { return &v }

func TestToCustomFieldResponse(t *testing.T) {
	cf := &ent.CustomFieldDefinition{
		ID:        uuid.New(),
		Name:      "Environment",
		FieldType: customfielddefinition.FieldTypeSelect,
		Options:   []string{"dev", "prod"},
		Required:  true,
		SortOrder: 2,
	}

	resp := toCustomFieldResponse(cf)
	assert.Equal(t, "Environment", resp.Name)
	assert.Equal(t, "select", resp.FieldType)
	assert.True(t, resp.Required)
}
