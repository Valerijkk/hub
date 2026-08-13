//go:build hubtest

package tracker_test

// TASK-191 / EA-001 — AutomationRule soft-delete вместо hard Delete.
import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/automationrule"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_AutomationRuleSoftDelete_TASK191 — Delete → deleted_at; list hides; row survives.
// Mutation: hard AutomationRule.Delete() → Get fails → FAIL.
// Mutation: drop DeletedAtIsNil from ListAutomationRules → soft-deleted listed → FAIL.
func TestEntRepository_AutomationRuleSoftDelete_TASK191(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	rule, err := env.repo.CreateAutomationRule(env.ctx, env.projectID, env.userID, tracker.CreateAutomationRuleInput{
		Name:        "TrashMe",
		Enabled:     true,
		TriggerType: "issue_created",
		ActionType:  "notify_user",
		ActionConfig: map[string]interface{}{
			"user_id": env.userID.String(),
		},
	})
	require.NoError(t, err)

	require.NoError(t, env.repo.DeleteAutomationRule(env.ctx, env.projectID, rule.ID))

	list, err := env.repo.ListAutomationRules(env.ctx, env.projectID)
	require.NoError(t, err)
	for _, r := range list {
		assert.NotEqual(t, rule.ID, r.ID, "EA-001: soft-deleted rule must not appear in ListAutomationRules")
	}

	raw, err := env.client.AutomationRule.Get(env.ctx, rule.ID)
	require.NoError(t, err, "EA-001: hard Delete would remove the row")
	require.NotNil(t, raw.DeletedAt)
	assert.WithinDuration(t, time.Now().UTC(), *raw.DeletedAt, time.Minute)

	_, err = env.repo.FindAutomationRule(env.ctx, env.projectID, rule.ID)
	require.True(t, ent.IsNotFound(err), "Find must skip trash, got %v", err)

	enabled, err := env.repo.ListEnabledAutomationRules(env.ctx, env.projectID, "issue_created")
	require.NoError(t, err)
	for _, r := range enabled {
		assert.NotEqual(t, rule.ID, r.ID, "engine must not see soft-deleted rules")
	}

	assertAutomationRuleSoftDeleteSource(t)
}

// TestEntRepository_AutomationRuleSoftDelete_Idempotent_TASK191 — повторный delete → not found.
func TestEntRepository_AutomationRuleSoftDelete_Idempotent_TASK191(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	rule, err := env.repo.CreateAutomationRule(env.ctx, env.projectID, env.userID, tracker.CreateAutomationRuleInput{
		Name:        "Once",
		Enabled:     true,
		TriggerType: "issue_assigned",
		ActionType:  "notify_user",
		ActionConfig: map[string]interface{}{
			"user_id": env.userID.String(),
		},
	})
	require.NoError(t, err)
	require.NoError(t, env.repo.DeleteAutomationRule(env.ctx, env.projectID, rule.ID))

	err = env.repo.DeleteAutomationRule(env.ctx, env.projectID, rule.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

// TestEntRepository_AutomationRuleCountIgnoresTrash_TASK191 — count не учитывает корзину.
func TestEntRepository_AutomationRuleCountIgnoresTrash_TASK191(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	rule, err := env.repo.CreateAutomationRule(env.ctx, env.projectID, env.userID, tracker.CreateAutomationRuleInput{
		Name:        "CountTrash",
		Enabled:     true,
		TriggerType: "status_changed",
		ActionType:  "notify_user",
		ActionConfig: map[string]interface{}{
			"user_id": env.userID.String(),
		},
	})
	require.NoError(t, err)
	require.NoError(t, env.repo.DeleteAutomationRule(env.ctx, env.projectID, rule.ID))

	n, err := env.repo.CountAutomationRules(env.ctx, env.projectID)
	require.NoError(t, err)
	assert.Equal(t, 0, n)

	// Sanity: row still in DB without filter.
	rawN, err := env.client.AutomationRule.Query().
		Where(automationrule.ProjectIDEQ(env.projectID)).
		Count(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, rawN)
}

// TestSchema_TASK191_AutomationRuleDeletedAt — schema guard.
// Mutation: убрать deleted_at из automationrule.go → FAIL.
func TestSchema_TASK191_AutomationRuleDeletedAt(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	schemaPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "ent", "schema", "automationrule.go")
	src, err := os.ReadFile(schemaPath)
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, `field.Time("deleted_at")`)
	assert.Contains(t, body, `index.Fields("deleted_at")`)
	assert.Contains(t, body, `index.Fields("project_id", "deleted_at")`)
}

// TestMigrate_TASK191_AutomationRuleSoftDeleteSQL — migrate marker needs-owner-apply.
func TestMigrate_TASK191_AutomationRuleSoftDeleteSQL(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	sqlPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "20260729_task191_automationrule_soft_delete.sql")
	src, err := os.ReadFile(sqlPath)
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, "needs-owner-apply")
	assert.Contains(t, body, "BLOCKED waiting owner")
	assert.Contains(t, body, "ALTER TABLE automation_rules ADD COLUMN IF NOT EXISTS deleted_at")
	assert.Contains(t, body, "automationrule_project_id_deleted_at")
}

func assertAutomationRuleSoftDeleteSource(t *testing.T) {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	root := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	src, err := os.ReadFile(filepath.Join(root, "internal", "tracker", "repo_automation.go"))
	require.NoError(t, err)
	chunk := string(src)
	assert.Contains(t, chunk, "SetDeletedAt", "EA-001: DeleteAutomationRule must SetDeletedAt")
	assert.Contains(t, chunk, "DeletedAtIsNil", "EA-001: list/get must filter DeletedAtIsNil")
	assert.False(t, strings.Contains(chunk, "client.AutomationRule.Delete()"),
		"EA-001: must not hard-delete automation rules")
	assert.Contains(t, chunk, "func (r *EntRepository) DeleteAutomationRule")
	// Hard path would call Delete().Where(...).Exec — soft path uses Update + SetDeletedAt.
	delFnStart := strings.Index(chunk, "func (r *EntRepository) DeleteAutomationRule")
	require.GreaterOrEqual(t, delFnStart, 0)
	delChunk := chunk[delFnStart:]
	if end := strings.Index(delChunk[1:], "\nfunc "); end > 0 {
		delChunk = delChunk[:end+1]
	}
	assert.NotContains(t, delChunk, ".Delete().", "EA-001: DeleteAutomationRule must not use hard Delete builder")
	assert.Contains(t, delChunk, "SetDeletedAt")
}
