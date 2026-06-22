package tracker_test

// Тесты сервиса трекера задач (волна 119): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/enttest"
)

// TestAutomationService_UpdateRule_Success проверяет бизнес-логику: автоматизация сервис обновление rule успех.
func TestAutomationService_UpdateRule_Success(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:autoup_ok_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Up OK").SetKey("UPK").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("upok@test.com").Save(ctx)
	require.NoError(t, err)

	svc := tracker.NewAutomationService(tracker.NewEntRepository(client, nil))
	created, err := svc.CreateRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleRequest{
		Name:        "Notify on create",
		TriggerType: tracker.TriggerIssueCreated,
		ActionType:  tracker.ActionNotifyUser,
	})
	require.NoError(t, err)

	enabled := false
	newName := "Renamed rule"
	ruleID, err := uuid.Parse(created.ID)
	require.NoError(t, err)
	updated, err := svc.UpdateRule(ctx, proj.ID, ruleID, tracker.UpdateAutomationRuleRequest{
		Name:    &newName,
		Enabled: &enabled,
	})
	require.NoError(t, err)
	assert.Equal(t, newName, updated.Name)
	assert.False(t, updated.Enabled)
}

// TestValidateAutomationRuleRequest_ConfigStringNonString проверяет сценарий: validate автоматизация rule request config string non string.
func TestValidateAutomationRuleRequest_ConfigStringNonString(t *testing.T) {
	err := tracker.ExportValidateAutomationRuleRequest(
		tracker.TriggerStatusChanged,
		tracker.ActionPostToChannel,
		map[string]interface{}{"to_status_id": 42},
		map[string]interface{}{"channel_id": uuid.New().String()},
	)
	require.NoError(t, err)

	err = tracker.ExportValidateAutomationRuleRequest(
		tracker.TriggerIssueCreated,
		tracker.ActionPostToChannel,
		nil,
		map[string]interface{}{"channel_id": 12345},
	)
	require.NoError(t, err)
}

// TestConfigString_NilMap проверяет сценарий: config string nil map.
func TestConfigString_NilMap(t *testing.T) {
	assert.Equal(t, "", tracker.ExportConfigString(nil, "any"))
}
