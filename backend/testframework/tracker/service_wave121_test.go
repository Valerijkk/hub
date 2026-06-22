package tracker_test

// Тесты сервиса трекера задач (волна 121): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/enttest"
)

// Фаза 121 — automation engine post-to-channel path (coverage bump)..

func TestAutomationEngine_Wave121_StatusChanged_PostToChannel(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:wave121_auto_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Auto121").SetKey("A121").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("a121@test.com").Save(ctx)
	require.NoError(t, err)
	status, err := client.WorkflowStatus.Create().SetName("Done").SetCategory("done").SetProject(proj).Save(ctx)
	require.NoError(t, err)
	ch, err := client.Channel.Create().SetName("alerts").SetChannelType("public").SetProjectID(proj.ID).SetCreatedBy(user.ID).Save(ctx)
	require.NoError(t, err)
	_, err = client.ChannelMember.Create().SetChannel(ch).SetUser(user).SetRole("owner").Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	_, err = repo.CreateAutomationRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleInput{
		Name: "Done notify", Enabled: true, TriggerType: tracker.TriggerStatusChanged,
		TriggerConfig: map[string]interface{}{"to_status_id": status.ID.String()},
		ActionType:    tracker.ActionPostToChannel,
		ActionConfig:  map[string]interface{}{"channel_id": ch.ID.String()},
	})
	require.NoError(t, err)

	messenger := &stubMessenger{}
	engine := tracker.NewAutomationEngine(repo, messenger, nil)
	engine.HandleStatusChanged(ctx, tracker.IssueMovedEvent{
		IssueKey: "A121-1", IssueID: uuid.New().String(), ProjectID: proj.ID.String(),
		NewStatusID: status.ID.String(), MovedByID: user.ID.String(),
	})
	assert.Len(t, messenger.posts, 1)
}
