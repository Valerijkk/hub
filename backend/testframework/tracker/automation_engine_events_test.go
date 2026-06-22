package tracker_test

// Тесты модуля трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/enttest"
)

type stubNotifier struct {
	count int
}

func (s *stubNotifier) NotifyAutomation(_ context.Context, _ uuid.UUID, _, _, _, _ string, _, _ *uuid.UUID) error {
	s.count++
	return nil
}

// TestAutomationEngine_IssueCreatedAndAssigned проверяет сценарий: автоматизация engine issue created and assigned.
func TestAutomationEngine_IssueCreatedAndAssigned(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:auto_events_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Events").SetKey("EVT").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("evt@test.com").Save(ctx)
	require.NoError(t, err)
	assignee, err := client.User.Create().SetEmail("assignee@test.com").Save(ctx)
	require.NoError(t, err)
	ch, err := client.Channel.Create().SetName("alerts").SetChannelType("public").SetProjectID(proj.ID).SetCreatedBy(user.ID).Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	_, err = repo.CreateAutomationRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleInput{
		Name: "On create", Enabled: true, TriggerType: tracker.TriggerIssueCreated,
		ActionType: tracker.ActionPostToChannel,
		ActionConfig: map[string]interface{}{"channel_id": ch.ID.String()},
	})
	require.NoError(t, err)
	_, err = repo.CreateAutomationRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleInput{
		Name: "On assign", Enabled: true, TriggerType: tracker.TriggerIssueAssigned,
		ActionType: tracker.ActionNotifyUser,
	})
	require.NoError(t, err)

	messenger := &stubMessenger{}
	notifier := &stubNotifier{}
	engine := tracker.NewAutomationEngine(repo, messenger, notifier)

	issueID := uuid.New()
	engine.HandleIssueCreated(ctx, tracker.IssueCreatedEvent{
		IssueKey: "EVT-1", IssueID: issueID.String(), ProjectID: proj.ID.String(),
		CreatedByID: user.ID.String(), IssueType: "task", Title: "New task",
	})
	assert.Len(t, messenger.posts, 1)

	engine.HandleIssueAssigned(ctx, tracker.IssueAssignedEvent{
		IssueKey: "EVT-1", IssueID: issueID.String(), ProjectID: proj.ID.String(),
		AssignedByID: user.ID.String(), AssigneeID: assignee.ID.String(),
	})
	assert.Equal(t, 1, notifier.count)
}

// TestAutomationEngine_IssueCreated_FiltersByType проверяет сценарий: автоматизация engine issue created filters by type.
func TestAutomationEngine_IssueCreated_FiltersByType(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:auto_filter_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Filter").SetKey("FLT").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("flt@test.com").Save(ctx)
	require.NoError(t, err)
	ch, err := client.Channel.Create().SetName("bugs").SetChannelType("public").SetProjectID(proj.ID).SetCreatedBy(user.ID).Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	_, err = repo.CreateAutomationRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleInput{
		Name: "Bugs only", Enabled: true, TriggerType: tracker.TriggerIssueCreated,
		TriggerConfig: map[string]interface{}{"issue_type": "bug"},
		ActionType:    tracker.ActionPostToChannel,
		ActionConfig:  map[string]interface{}{"channel_id": ch.ID.String()},
	})
	require.NoError(t, err)

	messenger := &stubMessenger{}
	engine := tracker.NewAutomationEngine(repo, messenger, nil)

	engine.HandleIssueCreated(ctx, tracker.IssueCreatedEvent{
		IssueKey: "FLT-1", IssueID: uuid.New().String(), ProjectID: proj.ID.String(),
		CreatedByID: user.ID.String(), IssueType: "task", Title: "Task",
	})
	assert.Empty(t, messenger.posts)

	engine.HandleIssueCreated(ctx, tracker.IssueCreatedEvent{
		IssueKey: "FLT-2", IssueID: uuid.New().String(), ProjectID: proj.ID.String(),
		CreatedByID: user.ID.String(), IssueType: "bug", Title: "Bug",
	})
	assert.Len(t, messenger.posts, 1)
}
