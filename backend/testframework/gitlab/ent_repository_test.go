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

	require.NoError(t, env.repo.DeleteIntegration(env.ctx, env.projectID, created.ID))

	afterDelete, err := env.repo.ListAllWebhookSecrets(env.ctx)
	require.NoError(t, err)
	assert.Empty(t, afterDelete)

	_, err = env.repo.GetIntegration(env.ctx, created.ID)
	require.Error(t, err)
}

// TestEntRepository_DeleteIntegration_RejectsForeignProject_A5_004 — интеграция проекта B
// не удаляется через project A. Mutation: снять HasProjectWith → чужой delete проходит → FAIL.
func TestEntRepository_DeleteIntegration_RejectsForeignProject_A5_004(t *testing.T) {
	env := newGitLabRepoEnv(t)

	foreign, err := env.client.Project.Create().
		SetName("Foreign GL Project").
		SetKey("FGL" + strings.ToUpper(uuid.New().String()[:3])).
		Save(env.ctx)
	require.NoError(t, err)

	foreignRepo := gitlab.NewEntRepository(env.client, nil)
	created, err := foreignRepo.CreateIntegration(env.ctx, foreign.ID, gitlab.CreateIntegrationInput{
		ProjectPath:   "other/secret",
		AccessToken:   "glpat-foreign",
		WebhookSecret: "foreign-wh",
	})
	require.NoError(t, err)

	err = env.repo.DeleteIntegration(env.ctx, env.projectID, created.ID)
	require.True(t, ent.IsNotFound(err), "foreign integration must not delete via project A, got %v", err)

	stillThere, err := foreignRepo.GetIntegration(env.ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, stillThere.ID)

	require.NoError(t, foreignRepo.DeleteIntegration(env.ctx, foreign.ID, created.ID))
	_, err = foreignRepo.GetIntegration(env.ctx, created.ID)
	require.True(t, ent.IsNotFound(err))
}

// TestEntRepository_GetIntegrationInProject_RejectsForeign_A5_005 — интеграция проекта B
// не находится через project A. Mutation: снять HasProjectWith → FAIL.
func TestEntRepository_GetIntegrationInProject_RejectsForeign_A5_005(t *testing.T) {
	env := newGitLabRepoEnv(t)

	foreign, err := env.client.Project.Create().
		SetName("Foreign GL Trigger").
		SetKey("FGT" + strings.ToUpper(uuid.New().String()[:3])).
		Save(env.ctx)
	require.NoError(t, err)

	foreignRepo := gitlab.NewEntRepository(env.client, nil)
	created, err := foreignRepo.CreateIntegration(env.ctx, foreign.ID, gitlab.CreateIntegrationInput{
		ProjectPath: "other/ci",
		AccessToken: "glpat-ci",
	})
	require.NoError(t, err)

	_, err = env.repo.GetIntegrationInProject(env.ctx, env.projectID, created.ID)
	require.True(t, ent.IsNotFound(err), "foreign integration must not resolve via project A, got %v", err)

	got, err := foreignRepo.GetIntegrationInProject(env.ctx, foreign.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

func TestEntRepository_FindIssueByKey(t *testing.T) {
	env := newGitLabRepoEnv(t)
	iss := env.createIssue(t, 1)

	got, err := env.repo.FindIssueByKey(env.ctx, iss.IssueKey)
	require.NoError(t, err)
	assert.Equal(t, iss.ID, got.ID)
	assert.NotNil(t, got.Edges.Project)
}

// TestEntRepository_FindIssueByKeyInProject_RejectsForeign_A5_007 — issue чужого проекта
// не находится по ключу. Mutation: снять HasProjectWith → тест краснеет.
func TestEntRepository_FindIssueByKeyInProject_RejectsForeign_A5_007(t *testing.T) {
	env := newGitLabRepoEnv(t)
	iss := env.createIssue(t, 1)

	foreign, err := env.client.Project.Create().
		SetName("Foreign GL Issue Project A5007").
		SetKey("FIG" + strings.ToUpper(uuid.New().String()[:3])).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = env.repo.FindIssueByKeyInProject(env.ctx, foreign.ID, iss.IssueKey)
	require.True(t, ent.IsNotFound(err), "issue must not be visible via foreign project, got %v", err)

	got, err := env.repo.FindIssueByKeyInProject(env.ctx, env.projectID, iss.IssueKey)
	require.NoError(t, err)
	assert.Equal(t, iss.ID, got.ID)
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
