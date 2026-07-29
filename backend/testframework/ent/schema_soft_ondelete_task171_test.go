//go:build hubtest

package ent_test

// TASK-171 / A3W8-5 — soft-delete Team/WorkflowStatus + OnDelete Cascade guards.
import (
	"context"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/team"
	"github.com/valerii/hub/backend/ent/teammember"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/testcaseversion"
	"github.com/valerii/hub/backend/ent/testresult"
	"github.com/valerii/hub/backend/ent/testrun"
	"github.com/valerii/hub/backend/ent/webhookdelivery"
	"github.com/valerii/hub/backend/ent/webhookendpoint"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/ent/workflowtransition"
)

func task171SchemaDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "ent", "schema")
}

// TestSchema_TASK171_SoftDeleteFields — A3W8-5-001 / 002.
// Mutation: убрать deleted_at / updated_at из team.go или workflowstatus.go → FAIL.
func TestSchema_TASK171_SoftDeleteFields(t *testing.T) {
	dir := task171SchemaDir(t)
	for _, file := range []string{"team.go", "workflowstatus.go"} {
		src, err := os.ReadFile(filepath.Join(dir, file))
		require.NoError(t, err, file)
		body := string(src)
		assert.Contains(t, body, `field.Time("deleted_at")`, file)
		assert.Contains(t, body, `field.Time("updated_at")`, file)
		assert.Contains(t, body, `index.Fields("deleted_at")`, file)
	}
}

// TestSchema_TASK171_OnDeleteCascades — A3W8-5-003…007.
// Mutation: убрать entsql.OnDelete(Cascade) с указанного edge → FAIL.
func TestSchema_TASK171_OnDeleteCascades(t *testing.T) {
	dir := task171SchemaDir(t)
	mustContain := map[string][]string{
		"team.go": {
			`edge.To("members"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
		"testcase.go": {
			`edge.To("versions"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
		"testrun.go": {
			`edge.To("results"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
		"webhookendpoint.go": {
			`edge.To("deliveries"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
		"workflowstatus.go": {
			`edge.To("from_transitions"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
		"workflowtransition.go": {
			`edge.To("to_status"`,
			`entsql.OnDelete(entsql.Cascade)`,
		},
	}
	for file, needles := range mustContain {
		src, err := os.ReadFile(filepath.Join(dir, file))
		require.NoError(t, err, file)
		body := string(src)
		for _, n := range needles {
			assert.Contains(t, body, n, "%s missing %q", file, n)
		}
	}
}

// TestSchema_TASK171_MigrateSQLExists — migrate marker.
func TestSchema_TASK171_MigrateSQLExists(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	sqlPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "20260729_task171_soft_delete_ondelete_cascade.sql")
	src, err := os.ReadFile(sqlPath)
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, "needs-owner-apply")
	assert.Contains(t, body, "A3W8-5-001")
	assert.Contains(t, body, "team_members_teams_members")
	assert.Contains(t, body, "test_case_versions_test_cases_versions")
	assert.Contains(t, body, "test_results_test_runs_results")
	assert.Contains(t, body, "webhook_deliveries_webhook_endpoints_deliveries")
	assert.Contains(t, body, "workflow_transitions_workflow_status_from_transitions")
	assert.Contains(t, body, "ON DELETE CASCADE")
}

func task171Open(t *testing.T, name string) *ent.Client {
	t.Helper()
	dsn := "file:task171_" + name + "_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return client
}

// TestCascade_TASK171_TeamHardDeleteRemovesMembers — A3W8-5-003.
// Mutation: убрать Cascade с members → FAIL (FK / orphan).
func TestCascade_TASK171_TeamHardDeleteRemovesMembers(t *testing.T) {
	client := task171Open(t, "team")
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("t171-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		Save(ctx)
	require.NoError(t, err)

	tm, err := client.Team.Create().SetName("CascadeTeam").Save(ctx)
	require.NoError(t, err)

	_, err = client.TeamMember.Create().SetTeamID(tm.ID).SetUserID(u.ID).Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.Team.DeleteOneID(tm.ID).Exec(ctx))

	n, err := client.TeamMember.Query().Where(teammember.HasTeamWith(team.IDEQ(tm.ID))).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

// TestSoft_TASK171_TeamDeletedAtListFilter — A3W8-5-001.
// Mutation: list без DeletedAtIsNil → soft-deleted в списке → FAIL.
func TestSoft_TASK171_TeamDeletedAtListFilter(t *testing.T) {
	client := task171Open(t, "teamsoft")
	ctx := context.Background()

	tm, err := client.Team.Create().SetName("SoftTeam").Save(ctx)
	require.NoError(t, err)
	require.NoError(t, client.Team.UpdateOneID(tm.ID).SetDeletedAt(time.Now().UTC()).Exec(ctx))

	active, err := client.Team.Query().Where(team.DeletedAtIsNil()).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, active)

	all, err := client.Team.Query().Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, all)
}

// TestCascade_TASK171_TestCaseHardDeleteRemovesVersions — A3W8-5-004.
func TestCascade_TASK171_TestCaseHardDeleteRemovesVersions(t *testing.T) {
	client := task171Open(t, "tcver")
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("tc171-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		Save(ctx)
	require.NoError(t, err)
	p, err := client.Project.Create().SetName("TC").SetKey("TC1").Save(ctx)
	require.NoError(t, err)

	tc, err := client.TestCase.Create().
		SetCaseNumber(1).
		SetCaseKey("TC1-TC-1").
		SetTitle("Case").
		SetProjectID(p.ID).
		SetCreatedByID(u.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.TestCaseVersion.Create().
		SetTestCaseID(tc.ID).
		SetVersionNumber(1).
		SetSnapshot(map[string]interface{}{"title": "Case"}).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.TestCase.DeleteOneID(tc.ID).Exec(ctx))

	n, err := client.TestCaseVersion.Query().
		Where(testcaseversion.HasTestCaseWith(testcase.IDEQ(tc.ID))).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

// TestCascade_TASK171_TestRunHardDeleteRemovesResults — A3W8-5-005.
func TestCascade_TASK171_TestRunHardDeleteRemovesResults(t *testing.T) {
	client := task171Open(t, "trun")
	ctx := context.Background()

	u, err := client.User.Create().
		SetEmail("tr171-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		Save(ctx)
	require.NoError(t, err)
	p, err := client.Project.Create().SetName("TR").SetKey("TR1").Save(ctx)
	require.NoError(t, err)

	run, err := client.TestRun.Create().
		SetTitle("Run").
		SetProjectID(p.ID).
		SetCreatedByID(u.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.TestResult.Create().SetRunID(run.ID).Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.TestRun.DeleteOneID(run.ID).Exec(ctx))

	n, err := client.TestResult.Query().Where(testresult.HasRunWith(testrun.IDEQ(run.ID))).Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

// TestCascade_TASK171_WebhookEndpointHardDeleteRemovesDeliveries — A3W8-5-006.
func TestCascade_TASK171_WebhookEndpointHardDeleteRemovesDeliveries(t *testing.T) {
	client := task171Open(t, "wh")
	ctx := context.Background()

	p, err := client.Project.Create().SetName("WH").SetKey("WH1").Save(ctx)
	require.NoError(t, err)

	ep, err := client.WebhookEndpoint.Create().
		SetURL("https://example.com/hook").
		SetSecret("s3cret-value-ok").
		SetEvents([]string{"issue.created"}).
		SetProjectID(p.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.WebhookDelivery.Create().
		SetEndpointID(ep.ID).
		SetEventType("issue.created").
		SetPayload(map[string]interface{}{"ok": true}).
		SetStatusCode(200).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.WebhookEndpoint.DeleteOneID(ep.ID).Exec(ctx))

	n, err := client.WebhookDelivery.Query().
		Where(webhookdelivery.HasEndpointWith(webhookendpoint.IDEQ(ep.ID))).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}

// TestCascade_TASK171_WorkflowStatusHardDeleteRemovesTransitions — A3W8-5-007.
func TestCascade_TASK171_WorkflowStatusHardDeleteRemovesTransitions(t *testing.T) {
	client := task171Open(t, "ws")
	ctx := context.Background()

	p, err := client.Project.Create().SetName("WS").SetKey("WS1").Save(ctx)
	require.NoError(t, err)

	from, err := client.WorkflowStatus.Create().
		SetName("From").
		SetCategory(workflowstatus.CategoryTodo).
		SetProjectID(p.ID).
		Save(ctx)
	require.NoError(t, err)
	to, err := client.WorkflowStatus.Create().
		SetName("To").
		SetCategory(workflowstatus.CategoryDone).
		SetProjectID(p.ID).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.WorkflowTransition.Create().
		SetFromStatusID(from.ID).
		SetToStatusID(to.ID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.WorkflowStatus.DeleteOneID(from.ID).Exec(ctx))

	n, err := client.WorkflowTransition.Query().
		Where(workflowtransition.HasFromStatusWith(workflowstatus.IDEQ(from.ID))).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)

	// to_status Cascade: удаление target чистит входящие переходы.
	from2, err := client.WorkflowStatus.Create().
		SetName("From2").
		SetCategory(workflowstatus.CategoryInProgress).
		SetProjectID(p.ID).
		Save(ctx)
	require.NoError(t, err)
	_, err = client.WorkflowTransition.Create().
		SetFromStatusID(from2.ID).
		SetToStatusID(to.ID).
		Save(ctx)
	require.NoError(t, err)

	require.NoError(t, client.WorkflowStatus.DeleteOneID(to.ID).Exec(ctx))
	n, err = client.WorkflowTransition.Query().
		Where(workflowtransition.HasToStatusWith(workflowstatus.IDEQ(to.ID))).
		Count(ctx)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
}
