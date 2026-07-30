//go:build hubtest

package ent_test

// TASK-170 / A3W8-4 — sprint active partial unique + OnDelete SetNull issues.
// Runtime unique guard = statements from migrate SQL file (not hardcoded DDL).
// Cascade SetNull = Ent Schema.Create from codegen (migrate/schema.go).
import (
	"context"
	"database/sql"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
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

func task170MigrateSQLPath(t *testing.T) string {
	t.Helper()
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	return filepath.Join(filepath.Dir(thisFile), "..", "..", "migrations", "20260729_task170_sprint_active_unique_setnull.sql")
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

// task170ApplyActiveUniqueFromMigrate — канон A3W8-4-001 из migrate SQL.
// Schema.Create может уже положить IndexWhere; DROP + recreate из файла —
// единственный runtime-источник индекса = канон migrate (не копия в тесте).
// Mutation: удалить/ослабить CREATE UNIQUE INDEX в SQL → constraint IT RED.
func task170ApplyActiveUniqueFromMigrate(t *testing.T, db *sql.DB) {
	t.Helper()
	body, err := os.ReadFile(task170MigrateSQLPath(t))
	require.NoError(t, err)
	stmt := task170ExtractCreateUniqueIndex(string(body))
	require.NotEmpty(t, stmt, "migrate SQL must contain CREATE UNIQUE INDEX sprint_project_sprints")

	_, err = db.Exec(`DROP INDEX IF EXISTS sprint_project_sprints`)
	require.NoError(t, err)
	_, err = db.Exec(stmt)
	require.NoError(t, err, "apply CREATE UNIQUE INDEX from migrate file")
}

var task170CreateUniqueIndexRE = regexp.MustCompile(
	`(?is)CREATE\s+UNIQUE\s+INDEX\s+(?:IF\s+NOT\s+EXISTS\s+)?sprint_project_sprints\s+[^;]+;`,
)

func task170ExtractCreateUniqueIndex(sqlBody string) string {
	m := task170CreateUniqueIndexRE.FindString(sqlBody)
	return strings.TrimSpace(m)
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

// TestSchema_TASK170_MigrateSQLExists — migrate marker (канон для live Postgres).
// Mutation: убрать CREATE UNIQUE INDEX / ON DELETE SET NULL из SQL → FAIL.
func TestSchema_TASK170_MigrateSQLExists(t *testing.T) {
	src, err := os.ReadFile(task170MigrateSQLPath(t))
	require.NoError(t, err)
	body := string(src)
	assert.Contains(t, body, "needs-owner-apply")
	assert.Contains(t, body, "A3W8-4-001")
	assert.Contains(t, body, "sprint_project_sprints")
	assert.Contains(t, body, "status = 'active' AND deleted_at IS NULL")
	assert.Contains(t, body, "CREATE UNIQUE INDEX")
	assert.Contains(t, body, "A3W8-4-002")
	assert.Contains(t, body, "issues_sprints_issues")
	assert.Contains(t, body, "ON DELETE SET NULL")
	require.NotEmpty(t, task170ExtractCreateUniqueIndex(body), "executable CREATE UNIQUE INDEX extractable")
}

// TestConstraint_TASK170_DuplicateActiveSprintRejected — A3W8-4-001.
// Mutation: удалить CREATE UNIQUE INDEX из migrate SQL → второй active проходит → FAIL.
func TestConstraint_TASK170_DuplicateActiveSprintRejected(t *testing.T) {
	env := task170Open(t, "dup")
	task170ApplyActiveUniqueFromMigrate(t, env.db)
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

// TestConstraint_TASK170_SoftDeletedActiveAllowsNewActive — A3W8-4-001b.
// Mutation: WHERE без deleted_at IS NULL в migrate SQL → soft-deleted active блокирует → FAIL.
func TestConstraint_TASK170_SoftDeletedActiveAllowsNewActive(t *testing.T) {
	env := task170Open(t, "softdel")
	task170ApplyActiveUniqueFromMigrate(t, env.db)
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
// Guard: Schema.Create из codegen (issues_sprints_issues OnDelete SetNull).
// Mutation: SetNull→NoAction/Restrict в migrate/schema.go (+ schema sprint.go source) → FK trap → FAIL.
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
