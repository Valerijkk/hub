package tracker_test

// Тесты сервиса трекера задач (волна 118): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
)

// TestCreateCustomField_DuplicateName проверяет сценарий: создание custom field duplicate name.
func TestCreateCustomField_DuplicateName(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:cf_dup_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("CF").SetKey("CF").Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	svc := tracker.NewTrackerService(repo, nil)

	req := tracker.CreateCustomFieldRequest{Name: "Risk", FieldType: "text"}
	_, err = svc.CreateCustomField(ctx, proj.ID, req)
	require.NoError(t, err)

	_, err = svc.CreateCustomField(ctx, proj.ID, req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "уже существует")
}

// TestAutomationService_ListRules_Empty проверяет бизнес-логику: автоматизация сервис список rules пустой.
func TestAutomationService_ListRules_Empty(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:auto_list_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Empty").SetKey("EMP").Save(ctx)
	require.NoError(t, err)

	svc := tracker.NewAutomationService(tracker.NewEntRepository(client, nil))
	rules, err := svc.ListRules(ctx, proj.ID)
	require.NoError(t, err)
	assert.Empty(t, rules)
}

// TestAutomationService_CreateRule_EmptyName проверяет бизнес-логику: автоматизация сервис создание rule пустой name.
func TestAutomationService_CreateRule_EmptyName(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:auto_name_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	proj, err := client.Project.Create().SetName("Name").SetKey("NAM").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("auto118@test.com").Save(ctx)
	require.NoError(t, err)

	svc := tracker.NewAutomationService(tracker.NewEntRepository(client, nil))
	_, err = svc.CreateRule(ctx, proj.ID, user.ID, tracker.CreateAutomationRuleRequest{
		Name:        "   ",
		TriggerType: tracker.TriggerIssueCreated,
		ActionType:  tracker.ActionNotifyUser,
	})
	require.Error(t, err)
}

// TestListBoards_EmptyProject проверяет сценарий: список boards пустой project.
func TestListBoards_EmptyProject(t *testing.T) {
	svc, _ := newTestService()
	ctx := context.Background()

	boards, err := svc.ListBoards(ctx, uuid.New())
	require.NoError(t, err)
	assert.Empty(t, boards)
}

// TestSwimlaneKeyLabel_UnknownSwimlane проверяет сценарий: swimlane key label unknown swimlane.
func TestSwimlaneKeyLabel_UnknownSwimlane(t *testing.T) {
	key, label := tracker.ExportSwimlaneKeyLabel(&ent.Issue{}, "custom")
	assert.Equal(t, "all", key)
	assert.Equal(t, "Все", label)
}
