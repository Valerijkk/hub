//go:build hubtest

package tracker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_UpdateAutomationRule_RequiresProjectID_P3BE05 — Update без ProjectIDEQ
// обновлял чужой rule. Mutation: UpdateOneID(ruleID) без Where ProjectIDEQ → FAIL.
func TestEntRepository_UpdateAutomationRule_RequiresProjectID_P3BE05(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	rule, err := env.repo.CreateAutomationRule(env.ctx, env.projectID, env.userID, tracker.CreateAutomationRuleInput{
		Name: "Scoped", Enabled: true,
		TriggerType: "issue_created", ActionType: "notify_user",
		TriggerConfig: map[string]interface{}{},
		ActionConfig:  map[string]interface{}{"user_id": env.userID.String()},
	})
	require.NoError(t, err)

	foreignProj, err := env.client.Project.Create().
		SetName("Foreign Auto").
		SetKey("FAU").
		Save(env.ctx)
	require.NoError(t, err)

	name := "hijacked"
	_, err = env.repo.UpdateAutomationRule(env.ctx, foreignProj.ID, rule.ID, tracker.UpdateAutomationRuleInput{
		Name: &name,
	})
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))

	still, err := env.repo.FindAutomationRule(env.ctx, env.projectID, rule.ID)
	require.NoError(t, err)
	assert.Equal(t, "Scoped", still.Name)

	okName := "renamed"
	updated, err := env.repo.UpdateAutomationRule(env.ctx, env.projectID, rule.ID, tracker.UpdateAutomationRuleInput{
		Name: &okName,
	})
	require.NoError(t, err)
	assert.Equal(t, "renamed", updated.Name)
}
