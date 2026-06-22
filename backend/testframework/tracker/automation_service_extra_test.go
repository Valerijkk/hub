package tracker_test

// Тесты сервиса трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/enttest"
)

// TestAutomationService_DeleteRule_NotFound проверяет бизнес-логику: автоматизация сервис удаление rule not found.
func TestAutomationService_DeleteRule_NotFound(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:autodel_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	proj, err := client.Project.Create().SetName("Del").SetKey("DEL").Save(ctx)
	require.NoError(t, err)
	repo := tracker.NewEntRepository(client, nil)
	svc := tracker.NewAutomationService(repo)

	err = svc.DeleteRule(ctx, proj.ID, uuid.New())
	require.Error(t, err)
}

// TestAutomationService_CreateRule_InvalidTrigger проверяет отклонение при автоматизация сервис создание rule некорректный trigger.
func TestAutomationService_CreateRule_InvalidTrigger(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:autoval_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	proj, err := client.Project.Create().SetName("Val").SetKey("VAL").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("val@test.com").Save(ctx)
	require.NoError(t, err)
	svc := tracker.NewAutomationService(tracker.NewEntRepository(client, nil))

	_, err = svc.CreateRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleRequest{
		Name:        "bad",
		TriggerType: "unknown_trigger",
		ActionType:  tracker.ActionNotifyUser,
	})
	require.Error(t, err)

	_, err = svc.CreateRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleRequest{
		Name:        "bad2",
		TriggerType: tracker.TriggerIssueCreated,
		ActionType:  "unknown_action",
	})
	require.Error(t, err)
}

// TestAutomationService_UpdateRule_NotFound проверяет бизнес-логику: автоматизация сервис обновление rule not found.
func TestAutomationService_UpdateRule_NotFound(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:autoup_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()
	proj, err := client.Project.Create().SetName("Up").SetKey("UP").Save(ctx)
	require.NoError(t, err)
	svc := tracker.NewAutomationService(tracker.NewEntRepository(client, nil))

	name := "renamed"
	_, err = svc.UpdateRule(ctx, proj.ID, uuid.New(), tracker.UpdateAutomationRuleRequest{Name: &name})
	require.Error(t, err)
}

// TestValidateAutomationRuleRequest_MissingChannel проверяет отклонение при validate автоматизация rule request отсутствующий channel.
func TestValidateAutomationRuleRequest_MissingChannel(t *testing.T) {
	err := tracker.ExportValidateAutomationRuleRequest(tracker.TriggerIssueCreated, tracker.ActionPostToChannel, nil, map[string]interface{}{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "channel_id")
}

// TestValidateAutomationRuleRequest_StatusChangedNeedsStatus проверяет сценарий: validate автоматизация rule request status changed needs status.
func TestValidateAutomationRuleRequest_StatusChangedNeedsStatus(t *testing.T) {
	err := tracker.ExportValidateAutomationRuleRequest(tracker.TriggerStatusChanged, tracker.ActionNotifyUser, map[string]interface{}{}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "to_status_id")
}
