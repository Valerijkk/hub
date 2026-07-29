//go:build hubtest

package ent_test

// TASK-170 / A3W8-4 — sprint active partial unique + OnDelete SetNull issues.
import (
	"context"
	"database/sql"
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
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/sprint"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func task170SchemaDir(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "ent", "schema")
}

type task170Env struct {
	client *ent.Client
	db     *sql.DB
}

func task170Open(t *testing.T, name string) task170Env {
	t.Helper()
	dsn := "file:task170_" + name + "_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	db, err := sql.Open("sqlite3", dsn)
	require.NoError(t, err)
	t.Cleanup(func() { _ = db.Close() })
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	return task170Env{client: client, db: db}
}

// task170ApplyActiveUnique — migrate SQL A3W8-4-001 (тот же DDL, что Atlas SQL).
// Mutation: убрать этот CREATE / IndexWhere из схемы+migrate → второй active проходит → FAIL.
func task170ApplyActiveUnique(t *testing.T, db *sql.DB) {
	t.Helper()
	_, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS sprint_project_sprints
		ON sprints (project_sprints)
		WHERE status = 'active' AND deleted_at IS NULL`)
	require.NoError(t, err)
}

func task170SeedProject(t *testing.T, client *ent.Client) (projectID, userID, statusID uuid.UUID) {
	t.Helper()
	ctx := context.Background()
	u, err := client.User.Create().
		SetEmail("t170-" + uuid.New().String()[:8] + "@t.com").
		SetName("U").
		Save(ctx)
	require.NoError(t, err)
	p, err := client.Project.Create().
		SetName("Sprint170").
		SetKey("S170").
		Save(ctx)
	require.NoError(t, err)
	st, err := client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#000").
		SetProjectID(p.ID).
		Save(ctx)
	require.NoError(t, err)
	return p.ID, u.ID, st.ID
}

// TestSchema_TASK170_SprintActiveUniqueAndSetNull — A3W8-4-001 / 002 markers.
// Mutation: убрать IndexWhere active / OnDelete(SetNull) из sprint.go → FAIL.
func TestSchema_TASK170_SprintActiveUniqueAndSetNull(t *testing.T) {
	src, err := os.ReadFile(filepath.Join(task170SchemaDir(t), "sprint.go"))
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, `entsql.OnDelete(entsql.SetNull)`)
	assert.Contains(t, body, `edge.To("issues"`)
	assert.Contains(t, body, `entsql.IndexWhere("status = 'active' AND deleted_at IS NULL")`)
	assert.Contains(t, body, `Unique()`)
}

// TestSchema_TASK170_MigrateSQLExists — migrate marker.
func TestSchema_TASK170_MigrateSQLExists(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	sqlPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "20260729_task170_sprint_active_unique_setnull.sql")
	src, err := os.ReadFile(sqlPath)
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, "needs-owner-apply")
	assert.Contains(t, body, "A3W8-4-001")
	assert.Contains(t, body, "sprint_project_sprints")
	assert.Contains(t, body, "status = 'active' AND deleted_at IS NULL")
	assert.Contains(t, body, "A3W8-4-002")
	assert.Contains(t, body, "issues_sprints_issues")
	assert.Contains(t, body, "ON DELETE SET NULL")
}

// TestConstraint_TASK170_DuplicateActiveSprintRejected — A3W8-4-001.
// Mutation: убрать CREATE UNIQUE INDEX / IndexWhere → второй active проходит → FAIL.
func TestConstraint_TASK170_DuplicateActiveSprintRejected(t *testing.T) {
	env := task170Open(t, "dup")
	task170ApplyActiveUnique(t, env.db)
	ctx := context.Background()
	projectID, _, _ := task170SeedProject(t, env.client)
	start := time.Now().UTC().Truncate(time.Second)
	end := start.AddDate(0, 0, 14)

	_, err := env.client.Sprint.Create().
		SetName("A").
		SetStatus(sprint.StatusActive).
		SetStartDate(start).
		SetEndDate(end).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	_, err = env.client.Sprint.Create().
		SetName("B").
		SetStatus(sprint.StatusActive).
		SetStartDate(start.AddDate(0, 0, 14)).
		SetEndDate(end.AddDate(0, 0, 14)).
		SetProjectID(projectID).
		Save(ctx)
	require.Error(t, err)
	assert.True(t, ent.IsConstraintError(err), "duplicate active: %v", err)
}

// TestConstraint_TASK170_SoftDeletedActiveAllowsNewActive — A3W8-4-001.
// Mutation: WHERE без deleted_at IS NULL → soft-deleted active блокирует новый → FAIL.
func TestConstraint_TASK170_SoftDeletedActiveAllowsNewActive(t *testing.T) {
	env := task170Open(t, "softdel")
	task170ApplyActiveUnique(t, env.db)
	ctx := context.Background()
	projectID, _, _ := task170SeedProject(t, env.client)
	start := time.Now().UTC().Truncate(time.Second)
	end := start.AddDate(0, 0, 14)

	old, err := env.client.Sprint.Create().
		SetName("OldActive").
		SetStatus(sprint.StatusActive).
		SetStartDate(start).
		SetEndDate(end).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)
	require.NoError(t, env.client.Sprint.UpdateOneID(old.ID).SetDeletedAt(time.Now().UTC()).Exec(ctx))

	_, err = env.client.Sprint.Create().
		SetName("NewActive").
		SetStatus(sprint.StatusActive).
		SetStartDate(start.AddDate(0, 0, 14)).
		SetEndDate(end.AddDate(0, 0, 14)).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err, "soft-deleted active must not block new active")
}

// TestCascade_TASK170_SprintHardDeleteSetsIssueSprintNull — A3W8-4-002.
// Mutation: убрать entsql.OnDelete(SetNull) → hard DELETE sprint с issues → FK trap → FAIL.
func TestCascade_TASK170_SprintHardDeleteSetsIssueSprintNull(t *testing.T) {
	env := task170Open(t, "setnull")
	ctx := context.Background()
	projectID, userID, statusID := task170SeedProject(t, env.client)
	start := time.Now().UTC().Truncate(time.Second)
	end := start.AddDate(0, 0, 14)

	sp, err := env.client.Sprint.Create().
		SetName("WithIssues").
		SetStatus(sprint.StatusPlanning).
		SetStartDate(start).
		SetEndDate(end).
		SetProjectID(projectID).
		Save(ctx)
	require.NoError(t, err)

	iss, err := env.client.Issue.Create().
		SetIssueNumber(1).
		SetIssueKey("S170-1").
		SetTitle("In sprint").
		SetIssueType(issue.IssueTypeTask).
		SetProjectID(projectID).
		SetReporterID(userID).
		SetStatusID(statusID).
		SetSprintID(sp.ID).
		Save(ctx)
	require.NoError(t, err)

	iss, err = env.client.Issue.Get(ctx, iss.ID)
	require.NoError(t, err)
	linked, err := iss.QuerySprint().Exist(ctx)
	require.NoError(t, err)
	require.True(t, linked)

	require.NoError(t, env.client.Sprint.DeleteOneID(sp.ID).Exec(ctx))

	iss, err = env.client.Issue.Get(ctx, iss.ID)
	require.NoError(t, err, "issue must survive sprint hard-delete")
	linked, err = iss.QuerySprint().Exist(ctx)
	require.NoError(t, err)
	assert.False(t, linked, "sprint_issues must be SET NULL")
}
