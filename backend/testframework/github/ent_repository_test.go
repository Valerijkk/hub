package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/githublink"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

type githubRepoEnv struct {
	client     *ent.Client
	repo       *github.EntRepository
	ctx        context.Context
	projectID  uuid.UUID
	userID     uuid.UUID
	projectKey string
}

func testGitHubProjectKey() string {
	return "GH" + strings.ToUpper(uuid.New().String()[:6])
}

func newGitHubRepoEnv(t *testing.T) githubRepoEnv {
	t.Helper()
	dsn := "file:ghrepo_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })

	ctx := context.Background()
	user, err := client.User.Create().
		SetEmail("gh-" + uuid.New().String()[:8] + "@test.com").
		SetName("GH Tester").
		Save(ctx)
	require.NoError(t, err)

	key := testGitHubProjectKey()
	proj, err := client.Project.Create().
		SetName("GitHub Project").
		SetKey(key).
		Save(ctx)
	require.NoError(t, err)

	return githubRepoEnv{
		client:     client,
		repo:       github.NewEntRepository(client, nil),
		ctx:        ctx,
		projectID:  proj.ID,
		userID:     user.ID,
		projectKey: key,
	}
}

func (env githubRepoEnv) createIssue(t *testing.T) *ent.Issue {
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
		SetIssueNumber(1).
		SetIssueKey(env.projectKey + "-1").
		SetTitle("Linked issue").
		SetIssueType(issue.IssueTypeTask).
		SetPriority(issue.PriorityMedium).
		Save(env.ctx)
	require.NoError(t, err)
	return iss
}

func TestEntRepository_IntegrationCRUDAndSecrets(t *testing.T) {
	env := newGitHubRepoEnv(t)

	created, err := env.repo.CreateIntegration(env.ctx, env.projectID, github.CreateIntegrationInput{
		RepoOwner:     "acme",
		RepoName:      "hub",
		AccessToken:   "ghp_test",
		WebhookSecret: "gh-secret",
	})
	require.NoError(t, err)
	assert.Equal(t, "acme", created.RepoOwner)
	assert.Equal(t, "hub", created.RepoName)

	list, err := env.repo.ListIntegrations(env.ctx, env.projectID)
	require.NoError(t, err)
	require.Len(t, list, 1)

	got, err := env.repo.GetIntegration(env.ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, "acme", got.RepoOwner)

	byRepo, err := env.repo.FindIntegrationByRepo(env.ctx, "acme", "hub")
	require.NoError(t, err)
	assert.Equal(t, created.ID, byRepo.ID)

	secrets, err := env.repo.ListAllWebhookSecrets(env.ctx)
	require.NoError(t, err)
	assert.Contains(t, secrets, "gh-secret")

	require.NoError(t, env.repo.DeleteIntegration(env.ctx, env.projectID, created.ID))
	after, err := env.repo.ListAllWebhookSecrets(env.ctx)
	require.NoError(t, err)
	assert.Empty(t, after)
}

// TestEntRepository_DeleteIntegration_RejectsForeignProject_A5_003 — интеграция проекта B
// не удаляется через project A. Mutation: снять HasProjectWith → чужой delete проходит → FAIL.
func TestEntRepository_DeleteIntegration_RejectsForeignProject_A5_003(t *testing.T) {
	env := newGitHubRepoEnv(t)

	foreign, err := env.client.Project.Create().
		SetName("Foreign GH Delete A5003").
		SetKey("FGD" + strings.ToUpper(uuid.New().String()[:3])).
		Save(env.ctx)
	require.NoError(t, err)

	foreignRepo := github.NewEntRepository(env.client, nil)
	created, err := foreignRepo.CreateIntegration(env.ctx, foreign.ID, github.CreateIntegrationInput{
		RepoOwner:     "other",
		RepoName:      "todelete",
		AccessToken:   "ghp_foreign",
		WebhookSecret: "foreign-secret",
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

// TestEntRepository_GetIntegrationInProject_RejectsForeign_A5_006 — интеграция проекта B
// не читается через project A. Mutation: снять HasProjectWith → тест краснеет.
func TestEntRepository_GetIntegrationInProject_RejectsForeign_A5_006(t *testing.T) {
	env := newGitHubRepoEnv(t)

	foreign, err := env.client.Project.Create().
		SetName("Foreign GH Project A5006").
		SetKey("FGH" + strings.ToUpper(uuid.New().String()[:3])).
		Save(env.ctx)
	require.NoError(t, err)

	foreignRepo := github.NewEntRepository(env.client, nil)
	created, err := foreignRepo.CreateIntegration(env.ctx, foreign.ID, github.CreateIntegrationInput{
		RepoOwner:     "other",
		RepoName:      "secret",
		AccessToken:   "ghp_foreign",
		WebhookSecret: "foreign-secret",
	})
	require.NoError(t, err)

	_, err = env.repo.GetIntegrationInProject(env.ctx, env.projectID, created.ID)
	require.True(t, ent.IsNotFound(err), "foreign integration must not be visible via project A, got %v", err)

	got, err := foreignRepo.GetIntegrationInProject(env.ctx, foreign.ID, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, got.ID)
}

// TestEntRepository_FindIssueByKeyInProject_RejectsForeign_A5_006 — issue чужого проекта
// не находится по ключу. Mutation: снять HasProjectWith → тест краснеет.
func TestEntRepository_FindIssueByKeyInProject_RejectsForeign_A5_006(t *testing.T) {
	env := newGitHubRepoEnv(t)
	iss := env.createIssue(t)

	foreign, err := env.client.Project.Create().
		SetName("Foreign Issue Project A5006").
		SetKey("FIS" + strings.ToUpper(uuid.New().String()[:3])).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = env.repo.FindIssueByKeyInProject(env.ctx, foreign.ID, iss.IssueKey)
	require.True(t, ent.IsNotFound(err), "issue must not be visible via foreign project, got %v", err)

	got, err := env.repo.FindIssueByKeyInProject(env.ctx, env.projectID, iss.IssueKey)
	require.NoError(t, err)
	assert.Equal(t, iss.ID, got.ID)
}

func TestEntRepository_GitHubLinksAndIssueLookup(t *testing.T) {
	env := newGitHubRepoEnv(t)
	iss := env.createIssue(t)

	ig, err := env.repo.CreateIntegration(env.ctx, env.projectID, github.CreateIntegrationInput{
		RepoOwner:   "acme",
		RepoName:    "app",
		AccessToken: "token",
	})
	require.NoError(t, err)

	link, err := env.repo.CreateGitHubLink(env.ctx, iss.ID, ig.ID, github.CreateGitHubLinkInput{
		LinkType:   "pull_request",
		ExternalID: "42",
		Title:      "Fix bug",
		URL:        "https://github.com/acme/app/pull/42",
		Status:     "open",
		Author:     "dev",
		Branch:     "fix/bug",
	})
	require.NoError(t, err)

	found, err := env.repo.FindGitHubLink(env.ctx, "42", "pull_request", ig.ID)
	require.NoError(t, err)
	assert.Equal(t, link.ID, found.ID)

	links, err := env.repo.ListGitHubLinks(env.ctx, iss.ID)
	require.NoError(t, err)
	require.Len(t, links, 1)

	require.NoError(t, env.repo.UpdateGitHubLinkStatus(env.ctx, link.ID, string(githublink.StatusMerged)))

	gotIssue, err := env.repo.FindIssueByKey(env.ctx, iss.IssueKey)
	require.NoError(t, err)
	assert.Equal(t, iss.ID, gotIssue.ID)

	err = env.repo.LogActivity(env.ctx, iss.ID, "github.pr.linked")
	require.Error(t, err) // ActivityLog.user is required — github repo omits userID today
}

func TestEntRepository_MarkTestCasesOutdatedByFiles(t *testing.T) {
	env := newGitHubRepoEnv(t)

	tc, err := env.client.TestCase.Create().
		SetProjectID(env.projectID).
		SetTitle("Login test").
		SetCaseKey("TC-1").
		SetCaseNumber(1).
		SetCaseType(testcase.CaseTypeClassic).
		SetPriority(testcase.PriorityMedium).
		SetStatus(testcase.StatusDraft).
		SetSyncStatus(testcase.SyncStatusSynced).
		SetClassName("com.example.LoginTest").
		SetCreatedByID(env.userID).
		Save(env.ctx)
	require.NoError(t, err)

	keys, err := env.repo.MarkTestCasesOutdatedByFiles(env.ctx, []string{
		"src/test/java/com/example/LoginTest.java",
	})
	require.NoError(t, err)
	require.Len(t, keys, 1)
	assert.Equal(t, tc.CaseKey, keys[0])

	updated, err := env.client.TestCase.Get(env.ctx, tc.ID)
	require.NoError(t, err)
	assert.Equal(t, testcase.SyncStatusOutdated, updated.SyncStatus)

	empty, err := env.repo.MarkTestCasesOutdatedByFiles(env.ctx, nil)
	require.NoError(t, err)
	assert.Nil(t, empty)
}

func TestEntRepository_FindDoneStatus(t *testing.T) {
	env := newGitHubRepoEnv(t)

	_, err := env.client.WorkflowStatus.Create().
		SetName("Done").
		SetCategory(workflowstatus.CategoryDone).
		SetProjectID(env.projectID).
		Save(env.ctx)
	require.NoError(t, err)

	done, err := env.repo.FindDoneStatus(env.ctx, env.projectID)
	require.NoError(t, err)
	assert.Equal(t, workflowstatus.CategoryDone, done.Category)
}
