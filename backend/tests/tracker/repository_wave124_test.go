package tracker_test

// Тесты репозитория трекера задач (волна 124): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Фаза 124 — CreateIssue optional fields + ListIssues filter matrix..

func TestEntRepository_Wave124_CreateIssue_AllOptionalFields(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	pub := &stubOutboxPublisher{}
	env.repo = tracker.NewEntRepository(env.client, pub)

	label, err := env.repo.CreateLabel(env.ctx, env.projectID, "wave124", "#112233", nil)
	require.NoError(t, err)

	parent, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Parent", IssueType: "epic", ReporterID: env.userID,
	})
	require.NoError(t, err)

	due := time.Now().UTC().Add(48 * time.Hour)
	est := 120
	sp := 3
	created, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title:           "Full issue",
		IssueType:       "story",
		Priority:        "high",
		ReporterID:      env.userID,
		AssigneeID:      &env.userID,
		ParentID:        &parent.ID,
		Labels:          []uuid.UUID{label.ID},
		DescriptionText: "desc",
		DescriptionJSON: map[string]interface{}{"type": "doc"},
		CustomFields:    map[string]interface{}{"env": "prod"},
		EstimateMinutes: &est,
		StoryPoints:     &sp,
		Checklist:       []map[string]interface{}{{"text": "item", "done": false}},
		DueDate:         &due,
	})
	require.NoError(t, err)
	assert.Equal(t, "Full issue", created.Title)
	assert.Contains(t, pub.events, "project."+env.projectID.String()+".issue.created")
}

// TestEntRepository_Wave124_ListIssues_FiltersAndPagination проверяет слой репозитория: ent репозиторий wave 124 список issues filters and pagination.
func TestEntRepository_Wave124_ListIssues_FiltersAndPagination(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	otherUser, err := env.client.User.Create().
		SetEmail("other-" + uuid.New().String()[:8] + "@test.com").
		SetName("Other").
		Save(env.ctx)
	require.NoError(t, err)

	inProgress, err := env.repo.CreateWorkflowStatus(env.ctx, env.projectID, "In Progress", "in_progress", "#3B82F6", 2)
	require.NoError(t, err)

	for i, title := range []string{"Alpha task", "Beta bug", "Gamma task"} {
		it := "task"
		if i == 1 {
			it = "bug"
		}
		_, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
			Title: title, IssueType: it, Priority: "high", ReporterID: env.userID,
			AssigneeID: func() *uuid.UUID {
				if i == 0 {
					return &env.userID
				}
				return nil
			}(),
			StatusID: func() *uuid.UUID {
				if i == 2 {
					return &inProgress.ID
				}
				return nil
			}(),
		})
		require.NoError(t, err)
	}

	unassigned, total, err := env.repo.ListIssues(env.ctx, env.projectID, tracker.ListIssuesParams{
		UnassignedOnly: true, Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Len(t, unassigned, 2)

	assigned, total, err := env.repo.ListIssues(env.ctx, env.projectID, tracker.ListIssuesParams{
		Assignee: &env.userID, Limit: 10,
	})
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	assert.Len(t, assigned, 1)

	byStatus, _, err := env.repo.ListIssues(env.ctx, env.projectID, tracker.ListIssuesParams{
		Status: []string{"In Progress"}, Limit: 10,
	})
	require.NoError(t, err)
	assert.Len(t, byStatus, 1)

	page1, total, err := env.repo.ListIssues(env.ctx, env.projectID, tracker.ListIssuesParams{
		Limit: 2, OrderBy: "updated_at",
	})
	require.NoError(t, err)
	assert.Equal(t, 3, total)
	require.Len(t, page1, 2)

	page2, _, err := env.repo.ListIssues(env.ctx, env.projectID, tracker.ListIssuesParams{
		Cursor: page1[len(page1)-1].IssueKey, Limit: 10, OrderBy: "updated_at",
	})
	require.NoError(t, err)
	assert.Len(t, page2, 1)
	page1Keys := make(map[string]struct{}, len(page1))
	for _, item := range page1 {
		page1Keys[item.IssueKey] = struct{}{}
	}
	for _, item := range page2 {
		assert.NotContains(t, page1Keys, item.IssueKey)
	}

	byPriority, _, err := env.repo.ListIssues(env.ctx, env.projectID, tracker.ListIssuesParams{
		Priority: []string{"high"}, Limit: 10, OrderBy: "priority",
	})
	require.NoError(t, err)
	assert.Len(t, byPriority, 3)

	_ = otherUser
	_ = inProgress
}
