package tracker_test

// Тесты модуля трекера задач (волна 100): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/automationrule"
	"github.com/valerii/hub/backend/ent/enttest"
)

// TestRenderAutomationTemplate проверяет сценарий: render автоматизация шаблон.
func TestRenderAutomationTemplate(t *testing.T) {
	got := tracker.ExportRenderAutomationTemplate(
		"{{issue_key}} / {{rule_name}} / {{title}}",
		"HUB-1", "Rule A", "Fix bug",
	)
	assert.Equal(t, "HUB-1 / Rule A / Fix bug", got)
}

// TestAutomationEngine_InvalidProjectID_NoOp проверяет отклонение при автоматизация engine некорректный project id no op.
func TestAutomationEngine_InvalidProjectID_NoOp(t *testing.T) {
	engine := tracker.NewAutomationEngine(nil, &stubMessenger{}, &stubNotifier{})
	engine.HandleStatusChanged(context.Background(), tracker.IssueMovedEvent{ProjectID: "bad"})
	engine.HandleIssueCreated(context.Background(), tracker.IssueCreatedEvent{ProjectID: "bad"})
	engine.HandleIssueAssigned(context.Background(), tracker.IssueAssignedEvent{ProjectID: "bad"})
}

// TestAutomationEngine_StatusChanged_WrongStatusSkips проверяет отклонение при автоматизация engine status changed неверный status skips.
func TestAutomationEngine_StatusChanged_WrongStatusSkips(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:auto_skip_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Skip").SetKey("SKP").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("skip@test.com").Save(ctx)
	require.NoError(t, err)
	done, err := client.WorkflowStatus.Create().SetName("Done").SetCategory("done").SetProject(proj).Save(ctx)
	require.NoError(t, err)
	todo, err := client.WorkflowStatus.Create().SetName("Todo").SetCategory("todo").SetProject(proj).Save(ctx)
	require.NoError(t, err)
	ch, err := client.Channel.Create().SetName("ops").SetChannelType("public").SetProjectID(proj.ID).SetCreatedBy(user.ID).Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	_, err = repo.CreateAutomationRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleInput{
		Name: "Done only", Enabled: true, TriggerType: tracker.TriggerStatusChanged,
		TriggerConfig: map[string]interface{}{"to_status_id": done.ID.String()},
		ActionType:    tracker.ActionPostToChannel,
		ActionConfig:  map[string]interface{}{"channel_id": ch.ID.String()},
	})
	require.NoError(t, err)

	messenger := &stubMessenger{}
	engine := tracker.NewAutomationEngine(repo, messenger, nil)
	engine.HandleStatusChanged(ctx, tracker.IssueMovedEvent{
		IssueKey: "SKP-1", IssueID: uuid.New().String(), ProjectID: proj.ID.String(),
		NewStatusID: todo.ID.String(), MovedByID: user.ID.String(),
	})
	assert.Empty(t, messenger.posts)
}

// TestAutomationEngine_InvalidChannelID проверяет отклонение при автоматизация engine некорректный channel id.
func TestAutomationEngine_InvalidChannelID(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:auto_badch_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("BadCh").SetKey("BCH").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("badch@test.com").Save(ctx)
	require.NoError(t, err)
	status, err := client.WorkflowStatus.Create().SetName("Done").SetCategory("done").SetProject(proj).Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	rule, err := client.AutomationRule.Create().
		SetName("Bad channel").
		SetProjectID(proj.ID).
		SetCreatedByID(user.ID).
		SetEnabled(true).
		SetTriggerType(automationrule.TriggerTypeIssueCreated).
		SetActionType(automationrule.ActionTypePostToChannel).
		SetTriggerConfig(map[string]interface{}{}).
		SetActionConfig(map[string]interface{}{"channel_id": "not-uuid"}).
		Save(ctx)
	require.NoError(t, err)

	messenger := &stubMessenger{}
	engine := tracker.NewAutomationEngine(repo, messenger, nil)
	tracker.ExportAutomationEngineExecuteRule(engine, ctx, rule, "BCH-1", uuid.New().String(), proj.ID.String(), user.ID.String(), "")

	// Дополнительно проверяем путь status-changed с невалидным каналом через сохранённое правило
	_, err = repo.CreateAutomationRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleInput{
		Name: "Bad ch rule", Enabled: true, TriggerType: tracker.TriggerStatusChanged,
		TriggerConfig: map[string]interface{}{"to_status_id": status.ID.String()},
		ActionType:    tracker.ActionPostToChannel,
		ActionConfig:  map[string]interface{}{"channel_id": "oops"},
	})
	require.NoError(t, err)
	engine.HandleStatusChanged(ctx, tracker.IssueMovedEvent{
		IssueKey: "BCH-1", IssueID: uuid.New().String(), ProjectID: proj.ID.String(),
		NewStatusID: status.ID.String(), MovedByID: user.ID.String(),
	})
	assert.Empty(t, messenger.posts)
}

// TestAutomationEngine_CustomTemplate проверяет сценарий: автоматизация engine custom шаблон.
func TestAutomationEngine_CustomTemplate(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:auto_tmpl_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Tmpl").SetKey("TMP").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("tmpl@test.com").Save(ctx)
	require.NoError(t, err)
	ch, err := client.Channel.Create().SetName("feed").SetChannelType("public").SetProjectID(proj.ID).SetCreatedBy(user.ID).Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	_, err = repo.CreateAutomationRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleInput{
		Name: "Custom", Enabled: true, TriggerType: tracker.TriggerIssueCreated,
		ActionType: tracker.ActionPostToChannel,
		ActionConfig: map[string]interface{}{
			"channel_id":        ch.ID.String(),
			"message_template":  "{{title}} for {{issue_key}} via {{rule_name}}",
		},
	})
	require.NoError(t, err)

	messenger := &stubMessenger{}
	engine := tracker.NewAutomationEngine(repo, messenger, nil)
	engine.HandleIssueCreated(ctx, tracker.IssueCreatedEvent{
		IssueKey: "TMP-9", IssueID: uuid.New().String(), ProjectID: proj.ID.String(),
		CreatedByID: user.ID.String(), IssueType: "task", Title: "Ship it",
	})
	require.Len(t, messenger.posts, 1)
	assert.Contains(t, messenger.posts[0], "Ship it for TMP-9 via Custom")
}

type stubMessengerFail struct{}

func (stubMessengerFail) PostAutomationMessage(_ context.Context, _, _ uuid.UUID, _ string) error {
	return fmt.Errorf("post failed")
}

// TestAutomationEngine_PostMessageError проверяет сценарий: автоматизация engine post message ошибка.
func TestAutomationEngine_PostMessageError(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:auto_postfail_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("PostFail").SetKey("PF").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("pf@test.com").Save(ctx)
	require.NoError(t, err)
	ch, err := client.Channel.Create().SetName("err").SetChannelType("public").SetProjectID(proj.ID).SetCreatedBy(user.ID).Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	_, err = repo.CreateAutomationRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleInput{
		Name: "Post", Enabled: true, TriggerType: tracker.TriggerIssueCreated,
		ActionType:   tracker.ActionPostToChannel,
		ActionConfig: map[string]interface{}{"channel_id": ch.ID.String()},
	})
	require.NoError(t, err)

	engine := tracker.NewAutomationEngine(repo, stubMessengerFail{}, nil)
	engine.HandleIssueCreated(ctx, tracker.IssueCreatedEvent{
		IssueKey: "PF-1", IssueID: uuid.New().String(), ProjectID: proj.ID.String(),
		CreatedByID: user.ID.String(), IssueType: "task", Title: "x",
	})
}

// TestAutomationEngine_NotifyInvalidAssignee проверяет отклонение при автоматизация engine notify некорректный assignee.
func TestAutomationEngine_NotifyInvalidAssignee(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:auto_notify_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Notify").SetKey("NTF").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("ntf@test.com").Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	rule, err := client.AutomationRule.Create().
		SetName("Notify").
		SetProjectID(proj.ID).
		SetCreatedByID(user.ID).
		SetEnabled(true).
		SetTriggerType(automationrule.TriggerTypeIssueAssigned).
		SetActionType(automationrule.ActionTypeNotifyUser).
		SetTriggerConfig(map[string]interface{}{}).
		SetActionConfig(map[string]interface{}{}).
		Save(ctx)
	require.NoError(t, err)

	notifier := &stubNotifier{}
	engine := tracker.NewAutomationEngine(repo, nil, notifier)
	tracker.ExportAutomationEngineExecuteRule(engine, ctx, rule, "NTF-1", uuid.New().String(), proj.ID.String(), user.ID.String(), "not-uuid")
	assert.Equal(t, 0, notifier.count)

	tracker.ExportAutomationEngineExecuteRule(engine, ctx, rule, "NTF-1", uuid.New().String(), proj.ID.String(), user.ID.String(), "")
	assert.Equal(t, 0, notifier.count)
}

// TestAutomationEngine_NilMessengerAndNotifier проверяет сценарий: автоматизация engine nil мессенджер and notifier.
func TestAutomationEngine_NilMessengerAndNotifier(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:auto_nil_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Nil").SetKey("NIL").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("nil@test.com").Save(ctx)
	require.NoError(t, err)
	assignee, err := client.User.Create().SetEmail("a@test.com").Save(ctx)
	require.NoError(t, err)
	ch, err := client.Channel.Create().SetName("nil").SetChannelType("public").SetProjectID(proj.ID).SetCreatedBy(user.ID).Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	_, err = repo.CreateAutomationRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleInput{
		Name: "Post", Enabled: true, TriggerType: tracker.TriggerIssueCreated,
		ActionType: tracker.ActionPostToChannel, ActionConfig: map[string]interface{}{"channel_id": ch.ID.String()},
	})
	require.NoError(t, err)
	_, err = repo.CreateAutomationRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleInput{
		Name: "Notify", Enabled: true, TriggerType: tracker.TriggerIssueAssigned, ActionType: tracker.ActionNotifyUser,
	})
	require.NoError(t, err)

	engine := tracker.NewAutomationEngine(repo, nil, nil)
	engine.HandleIssueCreated(ctx, tracker.IssueCreatedEvent{
		IssueKey: "NIL-1", IssueID: uuid.New().String(), ProjectID: proj.ID.String(),
		CreatedByID: user.ID.String(), IssueType: "task", Title: "x",
	})
	engine.HandleIssueAssigned(ctx, tracker.IssueAssignedEvent{
		IssueKey: "NIL-1", IssueID: uuid.New().String(), ProjectID: proj.ID.String(),
		AssignedByID: user.ID.String(), AssigneeID: assignee.ID.String(),
	})
}
