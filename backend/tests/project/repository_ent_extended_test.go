package project_test

// Тесты репозитория проектов: интеграционные и unit-сценарии testframework.
import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issuelink"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/projecttemplate"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

// TestEntRepository_FindByID проверяет слой репозитория: ent репозиторий find by id.
func TestEntRepository_FindByID(t *testing.T) {
	env := newProjectRepoEnv(t)

	created, err := env.repo.Create(env.ctx, "FindMe", "FND", "", env.owner.ID)
	require.NoError(t, err)

	found, err := env.repo.FindByID(env.ctx, created.ID)
	require.NoError(t, err)
	assert.Equal(t, created.ID, found.ID)
	assert.Equal(t, "FindMe", found.Name)

	_, err = env.repo.FindByID(env.ctx, uuid.New())
	require.Error(t, err)
}

// TestEntRepository_TemplateCRUD проверяет слой репозитория: ent репозиторий шаблон crud.
func TestEntRepository_TemplateCRUD(t *testing.T) {
	env := newProjectRepoEnv(t)
	cfg := map[string]interface{}{"statuses": []string{"Open", "Done"}}

	custom, err := env.repo.CreateTemplate(env.ctx, "Custom", "desc", projecttemplate.TemplateTypeCustom, cfg, &env.owner.ID, false)
	require.NoError(t, err)
	assert.Equal(t, "Custom", custom.Name)

	system, err := env.repo.CreateTemplate(env.ctx, "System", "", projecttemplate.TemplateTypeEmpty, cfg, nil, true)
	require.NoError(t, err)
	assert.True(t, system.IsSystem)

	list, err := env.repo.ListTemplates(env.ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, len(list), 2)

	got, err := env.repo.GetTemplate(env.ctx, custom.ID)
	require.NoError(t, err)
	assert.Equal(t, custom.ID, got.ID)

	count, err := env.repo.CountSystemTemplates(env.ctx)
	require.NoError(t, err)
	assert.GreaterOrEqual(t, count, 1)

	require.NoError(t, env.repo.DeleteTemplate(env.ctx, custom.ID))
	_, err = env.repo.GetTemplate(env.ctx, custom.ID)
	require.Error(t, err)
}

// TestEntRepository_GetProjectWithWorkflowStatusesAndLabels проверяет слой репозитория: ent репозиторий получение project with workflow statuses and метки.
func TestEntRepository_GetProjectWithWorkflowStatusesAndLabels(t *testing.T) {
	env := newProjectRepoEnv(t)

	proj, err := env.repo.Create(env.ctx, "TplSrc", "TPS", "", env.owner.ID)
	require.NoError(t, err)

	_, err = env.client.WorkflowStatus.Create().
		SetName("In Progress").
		SetCategory(workflowstatus.CategoryInProgress).
		SetColor("#3B82F6").
		SetSortOrder(1).
		SetProjectID(proj.ID).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = env.client.Label.Create().
		SetName("backend").
		SetColor("#EF4444").
		SetProjectID(proj.ID).
		Save(env.ctx)
	require.NoError(t, err)

	got, statuses, labels, err := env.repo.GetProjectWithWorkflowStatusesAndLabels(env.ctx, proj.ID)
	require.NoError(t, err)
	assert.Equal(t, proj.ID, got.ID)
	assert.Contains(t, statuses, "In Progress")
	assert.Contains(t, labels, "backend")
}

// TestEntRepository_CrossProjectLinks проверяет слой репозитория: ent репозиторий cross project links.
func TestEntRepository_CrossProjectLinks(t *testing.T) {
	env := newProjectRepoEnv(t)

	projA, err := env.repo.Create(env.ctx, "Alpha", "ALP", "", env.owner.ID)
	require.NoError(t, err)
	projB, err := env.repo.Create(env.ctx, "Beta", "BET", "", env.owner.ID)
	require.NoError(t, err)

	statusA, err := env.client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(projA.ID).
		Save(env.ctx)
	require.NoError(t, err)

	statusB, err := env.client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory(workflowstatus.CategoryTodo).
		SetColor("#6B7280").
		SetSortOrder(0).
		SetIsDefault(true).
		SetProjectID(projB.ID).
		Save(env.ctx)
	require.NoError(t, err)

	src := createTestIssue(t, env, projA.ID, 1, "ALP-1", statusA.ID)
	tgt := createTestIssue(t, env, projB.ID, 1, "BET-1", statusB.ID)

	link, err := env.client.IssueLink.Create().
		SetLinkType(issuelink.LinkTypeRelatesTo).
		SetSourceID(src.ID).
		SetTargetID(tgt.ID).
		Save(env.ctx)
	require.NoError(t, err)

	outgoing, err := env.repo.ListOutgoingCrossProjectLinks(env.ctx, projA.ID)
	require.NoError(t, err)
	require.Len(t, outgoing, 1)
	assert.Equal(t, link.ID, outgoing[0].ID)
	require.NotNil(t, outgoing[0].Edges.Source)
	require.NotNil(t, outgoing[0].Edges.Target)

	incoming, err := env.repo.ListIncomingCrossProjectLinks(env.ctx, projB.ID)
	require.NoError(t, err)
	require.Len(t, incoming, 1)
	assert.Equal(t, link.ID, incoming[0].ID)

	require.NoError(t, env.repo.DeleteIssueLink(env.ctx, link.ID))
	outgoing, err = env.repo.ListOutgoingCrossProjectLinks(env.ctx, projA.ID)
	require.NoError(t, err)
	assert.Empty(t, outgoing)
}

func createTestIssue(t *testing.T, env projectRepoEnv, projectID uuid.UUID, num int, key string, statusID uuid.UUID) *ent.Issue {
	t.Helper()
	iss, err := env.client.Issue.Create().
		SetIssueNumber(num).
		SetIssueKey(key).
		SetTitle("Link issue").
		SetIssueType(issue.IssueTypeTask).
		SetProjectID(projectID).
		SetReporterID(env.owner.ID).
		SetStatusID(statusID).
		Save(env.ctx)
	require.NoError(t, err)
	return iss
}
