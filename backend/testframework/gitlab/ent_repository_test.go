package gitlab_test
import (
	"github.com/valerii/hub/backend/internal/gitlab"
	"context"
	"strconv"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

type gitlabRepoEnv struct {
	client    *ent.Client
	repo      *gitlab.EntRepository
	ctx       context.Context
	projectID uuid.UUID
	userID    uuid.UUID
	projectKey string
}

func testGitLabProjectKey() string {
	return "GL" + strings.ToUpper(uuid.New().String()[:6])
}

func newGitLabRepoEnv(t *testing.T) gitlabRepoEnv {
	t.Helper()
	dsn := "file:glrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("gl-" + uuid.New().String()[:8] + "@test.com").
		SetName("GL Tester").
		Save(ctx)
	require.NoError(t, err)

	key := testGitLabProjectKey()
	proj, err := client.Project.Create().
		SetName("GitLab Project").
		SetKey(key).
		Save(ctx)
	require.NoError(t, err)

	return gitlabRepoEnv{
		client:     client,
		repo:       gitlab.NewEntRepository(client, nil),
		ctx:        ctx,
		projectID:  proj.ID,
		userID:     user.ID,
		projectKey: key,
	}
}

func (env gitlabRepoEnv) createIssue(t *testing.T, num int) *ent.Issue {
	t.Helper()
	status, err := env.client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetProjectID(env.projectID).
		SetIsDefault(true).
		Save(env.ctx)
	require.NoError(t, err)

	iss, err := env.client.Issue.Create().
		SetProjectID(env.projectID).
		SetStatusID(status.ID).
		SetReporterID(env.userID).
		SetIssueNumber(num).
		SetIssueKey(env.projectKey + "-" + strconv.Itoa(num)).
		SetTitle("Issue " + strconv.Itoa(num)).
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		Save(env.ctx)
	require.NoError(t, err)
	return iss
}

func TestEntRepository_IntegrationCRUDAndSecrets(t *testing.T) {
	env := newGitLabRepoEnv(t)

	created, err := env.repo.CreateIntegration(env.ctx, env.projectID, gitlab.CreateIntegrationInput{
		ProjectPath:   "group/repo",
		AccessToken:   "glpat-test",
		WebhookSecret: "wh-secret-1",
	})
	require.NoError(t, err)
	assert.Equal(t, "https://gitlab.com", created.GitlabURL)
	assert.Equal(t, "group/repo", created.ProjectPath)
	assert.Equal(t, "glpat-test", created.AccessToken)
	assert.Equal(t, "wh-secret-1", created.WebhookSecret)

	list, err := env.repo.ListIntegrations(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, created.ID, list[0].ID)

	got, err := env.repo.GetIntegration(env.ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "group/repo", got.ProjectPath)

	byPath, err := env.repo.FindIntegrationByProjectPath(env.ctx, "group/repo")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byPath.ID)

	secrets, err := env.repo.ListAllWebhookSecrets(env.ctx)
	require.NoError(t, err)
	assert.Contains(t, secrets, "wh-secret-1")

	cached, err := env.repo.ListAllWebhookSecrets(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, secrets, cached)

	require.NoError(t, env.repo.DeleteIntegration(env.ctx, created.ID))

	afterDelete, err := env.repo.ListAllWebhookSecrets(env.ctx)
	require.NoError(t, err)
	assert.Empty(t, afterDelete)

	_, err = env.repo.GetIntegration(env.ctx, created.ID)
	require.Error(t, err)
}

func TestEntRepository_FindIssueByKey(t *testing.T) {
	env := newGitLabRepoEnv(t)
	iss := env.createIssue(t, 1)

	got, err := env.repo.FindIssueByKey(env.ctx, iss.IssueKey)
	require.NoError(t, err)
	assert.Equal(t, iss.ID, got.ID)
	assert.NotNil(t, got.Edges.Project)
}

func TestEntRepository_CreateIntegration_CustomURL(t *testing.T) {
	env := newGitLabRepoEnv(t)

	created, err := env.repo.CreateIntegration(env.ctx, env.projectID, gitlab.CreateIntegrationInput{
		GitLabURL:   "https://gitlab.example.com",
		ProjectPath: "acme/app",
		AccessToken: "token",
	})
	require.NoError(t, err)
	assert.Equal(t, "https://gitlab.example.com", created.GitlabURL)
}
