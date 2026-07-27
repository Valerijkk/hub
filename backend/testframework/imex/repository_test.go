package imex_test
import (
	"github.com/valerii/hub/backend/internal/imex"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

func newImexRepoTestEnv(t *testing.T) (*imex.EntRepository, context.Context, uuid.UUID, uuid.UUID) {
	t.Helper()
	dsn := "file:imexrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("repo-" + uuid.New().String()[:8] + "@test.com").
		SetName("Repo User").
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Repo Project").
		SetKey("RP").
		Save(ctx)
	require.NoError(t, err)

	status, err := client.WorkflowStatus.Create().
		SetName("To Do").
		SetCategory(workflowstatus.CategoryTodo).
		SetIsDefault(true).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)
	_ = status

	return imex.NewEntRepository(client), ctx, proj.ID, user.ID
}

func TestEntRepository_AllocateIssueKey(t *testing.T) {
	repo, ctx, projectID, _ := newImexRepoTestEnv(t)

	num, key, statusID, err := repo.AllocateIssueKey(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, 1, num)
	assert.Equal(t, "RP-1", key)
	assert.NotEqual(t, uuid.Nil, statusID)

	num2, key2, _, err := repo.AllocateIssueKey(ctx, projectID)
	require.NoError(t, err)
	assert.Equal(t, 2, num2)
	assert.Equal(t, "RP-2", key2)
}

func TestEntRepository_CreateIssueAndList(t *testing.T) {
	repo, ctx, projectID, userID := newImexRepoTestEnv(t)

	num, key, statusID, err := repo.AllocateIssueKey(ctx, projectID)
	require.NoError(t, err)

	desc := "imported body"
	created, err := repo.CreateIssue(ctx, imex.CreateIssueInput{
		Title:           "Test issue",
		IssueType:       issue.IssueTypeTask,
		Priority:        issue.PriorityMedium,
		IssueNumber:     num,
		IssueKey:        key,
		ProjectID:       projectID,
		StatusID:        statusID,
		ReporterID:      &userID,
		DescriptionText: &desc,
	})
	require.NoError(t, err)
	assert.Equal(t, key, created.IssueKey)

	err = repo.CreateIssueComment(ctx, created.ID, userID, "hello")
	require.NoError(t, err)

	issues, err := repo.ListProjectIssuesForExport(ctx, projectID)
	require.NoError(t, err)
	require.Len(t, issues, 1)
	assert.Equal(t, "Test issue", issues[0].Title)

	plain, err := repo.ListProjectIssuesPlain(ctx, projectID)
	require.NoError(t, err)
	assert.Len(t, plain, 1)
}

func TestEntRepository_SetIssueParent(t *testing.T) {
	repo, ctx, projectID, _ := newImexRepoTestEnv(t)

	parentNum, parentKey, statusID, err := repo.AllocateIssueKey(ctx, projectID)
	require.NoError(t, err)
	parent, err := repo.CreateIssue(ctx, imex.CreateIssueInput{
		Title:       "Parent",
		IssueType:   issue.IssueTypeEpic,
		Priority:    issue.PriorityMedium,
		IssueNumber: parentNum,
		IssueKey:    parentKey,
		ProjectID:   projectID,
		StatusID:    statusID,
	})
	require.NoError(t, err)

	childNum, childKey, _, err := repo.AllocateIssueKey(ctx, projectID)
	require.NoError(t, err)
	child, err := repo.CreateIssue(ctx, imex.CreateIssueInput{
		Title:       "Child",
		IssueType:   issue.IssueTypeSubTask,
		Priority:    issue.PriorityLow,
		IssueNumber: childNum,
		IssueKey:    childKey,
		ProjectID:   projectID,
		StatusID:    statusID,
	})
	require.NoError(t, err)

	err = repo.SetIssueParent(ctx, child.ID, parent.ID)
	require.NoError(t, err)
}

func TestEntRepository_AllocateIssueKey_UnknownProject(t *testing.T) {
	repo, ctx, _, _ := newImexRepoTestEnv(t)

	_, _, _, err := repo.AllocateIssueKey(ctx, uuid.New())
	require.Error(t, err)
}

func TestIsNotFound(t *testing.T) {
	repo, ctx, projectID, _ := newImexRepoTestEnv(t)
	_, err := repo.GetTestRunInProject(ctx, projectID, uuid.New())
	assert.True(t, imex.IsNotFound(err))
}
