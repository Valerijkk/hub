package tracker_test

// Тесты модуля трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/enttest"
)

type stubMessenger struct {
	posts []string
}

func (s *stubMessenger) PostAutomationMessage(_ context.Context, _, _ uuid.UUID, text string) error {
	s.posts = append(s.posts, text)
	return nil
}

// TestAutomationEngine_StatusChanged_PostToChannel проверяет сценарий: автоматизация engine status changed post to channel.
func TestAutomationEngine_StatusChanged_PostToChannel(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:automation_engine?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	ctx := context.Background()

	org, err := client.Organization.Create().SetName("Test Org").SetSlug("test-org").SetInviteToken("tok").Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().SetName("Auto Proj").SetKey("AUT").SetOrganization(org).Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("a@ex.com").Save(ctx)
	require.NoError(t, err)
	status, err := client.WorkflowStatus.Create().SetName("Done").SetCategory("done").SetProject(proj).Save(ctx)
	require.NoError(t, err)
	ch, err := client.Channel.Create().SetName("general").SetChannelType("public").SetProjectID(proj.ID).SetCreatedBy(user.ID).Save(ctx)
	require.NoError(t, err)
	_, err = client.ChannelMember.Create().SetChannel(ch).SetUser(user).SetRole("owner").Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	_, err = repo.CreateAutomationRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleInput{
		Name:        "Done notify",
		Enabled:     true,
		TriggerType: tracker.TriggerStatusChanged,
		TriggerConfig: map[string]interface{}{
			"to_status_id": status.ID.String(),
		},
		ActionType: tracker.ActionPostToChannel,
		ActionConfig: map[string]interface{}{
			"channel_id": ch.ID.String(),
		},
	})
	require.NoError(t, err)

	messenger := &stubMessenger{}
	engine := tracker.NewAutomationEngine(repo, messenger, nil)
	engine.HandleStatusChanged(ctx, tracker.IssueMovedEvent{
		IssueKey:    "AUT-1",
		IssueID:     uuid.New().String(),
		ProjectID:   proj.ID.String(),
		NewStatusID: status.ID.String(),
		MovedByID:   user.ID.String(),
	})
	assert.Len(t, messenger.posts, 1)
	assert.Contains(t, messenger.posts[0], "AUT-1")
}

// TestAutomationService_MaxRules проверяет бизнес-логику: автоматизация сервис max rules.
func TestAutomationService_MaxRules(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:automation_svc?mode=memory&cache=shared&_fk=1")
	defer client.Close()
	ctx := context.Background()
	proj, err := client.Project.Create().SetName("Max Proj").SetKey("MX").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("b@ex.com").Save(ctx)
	require.NoError(t, err)
	repo := tracker.NewEntRepository(client, nil)
	svc := tracker.NewAutomationService(repo)
	base := tracker.CreateAutomationRuleRequest{
		Name:        "r",
		TriggerType: tracker.TriggerIssueCreated,
		ActionType:  tracker.ActionNotifyUser,
	}
	for i := 0; i < tracker.ExportMaxAutomationRulesPerProject(); i++ {
		base.Name = fmt.Sprintf("rule%d", i+1)
		_, err := svc.CreateRule(ctx, proj.ID, user.ID, base)
		require.NoError(t, err)
	}
	_, err = svc.CreateRule(ctx, proj.ID, user.ID, base)
	require.Error(t, err)
}
