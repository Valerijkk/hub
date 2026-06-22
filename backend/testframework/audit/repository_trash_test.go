package audit_test

// Тесты репозитория аудита: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/audit"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

type auditTrashEnv struct {
	repo      audit.Repository
	ctx       context.Context
	client    *ent.Client
	projectID uuid.UUID
	userA     uuid.UUID
	userB     uuid.UUID
	statusID  uuid.UUID
}

func newAuditTrashEnv(t *testing.T) auditTrashEnv {
	t.Helper()
	dsn := "file:audittrash_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	reporter, err := client.User.Create().
		SetEmail("rep-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)
	assignee, err := client.User.Create().
		SetEmail("asg-" + uuid.New().String()[:8] + "@test.com").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Trash Project").
		SetKey("TRH").
		Save(ctx)
	require.NoError(t, err)

	status, err := client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#000").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	return auditTrashEnv{
		repo:      audit.NewRepository(client),
		ctx:       ctx,
		client:    client,
		projectID: proj.ID,
		userA:     reporter.ID,
		userB:     assignee.ID,
		statusID:  status.ID,
	}
}

func (e auditTrashEnv) createIssue(t *testing.T, assigneeID *uuid.UUID) *ent.Issue {
	t.Helper()
	create := e.client.Issue.Create().
		SetProjectID(e.projectID).
		SetStatusID(e.statusID).
		SetReporterID(e.userA).
		SetIssueNumber(1).
		SetIssueKey("TRH-1").
		SetTitle("Trash issue").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium)
	if assigneeID != nil {
		create = create.SetAssigneeID(*assigneeID)
	}
	iss, err := create.Save(e.ctx)
	require.NoError(t, err)
	return iss
}

// TestRepository_SoftDeleteAndRestoreIssue проверяет слой репозитория: репозиторий soft удаление and restore issue.
func TestRepository_SoftDeleteAndRestoreIssue(t *testing.T) {
	env := newAuditTrashEnv(t)
	iss := env.createIssue(t, nil)
	when := time.Now().UTC()

	require.NoError(t, env.repo.SoftDeleteIssue(env.ctx, iss.ID, when))

	items, total, err := env.repo.ListDeletedIssues(env.ctx, &env.projectID, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)
	require.Len(t, items, 1)
	assert.Equal(t, iss.ID, items[0].ID)

	require.NoError(t, env.repo.RestoreIssue(env.ctx, iss.ID))

	items, total, err = env.repo.ListDeletedIssues(env.ctx, &env.projectID, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 0, total)
}

// TestRepository_SoftDeleteAndRestoreProject проверяет слой репозитория: репозиторий soft удаление and restore project.
func TestRepository_SoftDeleteAndRestoreProject(t *testing.T) {
	env := newAuditTrashEnv(t)
	when := time.Now().UTC()

	require.NoError(t, env.repo.SoftDeleteProject(env.ctx, env.projectID, when))

	_, total, err := env.repo.ListDeletedProjects(env.ctx, when.Add(-time.Hour), 0, 10)
	require.NoError(t, err)
	assert.Equal(t, 1, total)

	require.NoError(t, env.repo.RestoreProject(env.ctx, env.projectID))

	updated, err := env.client.Project.Get(env.ctx, env.projectID)
	require.NoError(t, err)
	assert.Equal(t, project.StatusActive, updated.Status)
	assert.Nil(t, updated.DeletedAt)
}

// TestRepository_PurgeExpiredIssues проверяет отклонение при репозиторий очистка просроченный issues.
func TestRepository_PurgeExpiredIssues(t *testing.T) {
	env := newAuditTrashEnv(t)
	iss := env.createIssue(t, nil)
	old := time.Now().UTC().Add(-40 * 24 * time.Hour)

	require.NoError(t, env.repo.SoftDeleteIssue(env.ctx, iss.ID, old))

	n, err := env.repo.PurgeExpiredIssues(env.ctx, time.Now().UTC().Add(-30*24*time.Hour))
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	_, err = env.client.Issue.Get(env.ctx, iss.ID)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}

// TestRepository_ReassignAssignedIssues проверяет слой репозитория: репозиторий reassign assigned issues.
func TestRepository_ReassignAssignedIssues(t *testing.T) {
	env := newAuditTrashEnv(t)
	iss := env.createIssue(t, &env.userA)

	n, err := env.repo.ReassignAssignedIssues(env.ctx, env.userA, env.userB)
	require.NoError(t, err)
	assert.Equal(t, 1, n)

	updated, err := env.client.Issue.Query().
		Where(issue.IDEQ(iss.ID)).
		WithAssignee().
		Only(env.ctx)
	require.NoError(t, err)
	require.NotNil(t, updated.Edges.Assignee)
	assert.Equal(t, env.userB, updated.Edges.Assignee.ID)
}

// TestRepository_DeactivateUserAndSessions проверяет слой репозитория: репозиторий deactivate user and sessions.
func TestRepository_DeactivateUserAndSessions(t *testing.T) {
	env := newAuditTrashEnv(t)

	_, err := env.client.Session.Create().
		SetUserID(env.userA).
		SetRefreshToken("hash-" + uuid.New().String()).
		SetDevice("browser").
		SetIPAddress("127.0.0.1").
		SetExpiresAt(time.Now().UTC().Add(24 * time.Hour)).
		Save(env.ctx)
	require.NoError(t, err)

	require.NoError(t, env.repo.DeactivateUser(env.ctx, env.userA))

	u, err := env.client.User.Get(env.ctx, env.userA)
	require.NoError(t, err)
	assert.Equal(t, user.StatusDeactivated, u.Status)

	n, err := env.repo.DeleteUserSessions(env.ctx, env.userA)
	require.NoError(t, err)
	assert.Equal(t, 1, n)
}

// TestRepository_HardDeleteUser проверяет слой репозитория: репозиторий hard удаление user.
func TestRepository_HardDeleteUser(t *testing.T) {
	env := newAuditTrashEnv(t)

	require.NoError(t, env.repo.HardDeleteUser(env.ctx, env.userA))

	_, err := env.client.User.Get(env.ctx, env.userA)
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
}
