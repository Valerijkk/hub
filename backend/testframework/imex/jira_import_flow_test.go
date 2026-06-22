package imex_test
import (
	"github.com/valerii/hub/backend/internal/imex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/issue"
)

func TestImportJiraJSON_ParentChildAndComments(t *testing.T) {
	svc, client, ctx, projectID, userID := newImexTestEnv(t)

	payload := []imex.JiraIssue{
		{
			Key: "EPIC-1",
			Fields: imex.JiraFields{
				Summary:   "Epic title",
				IssueType: imex.JiraNamedField{Name: "Epic"},
				Priority:  imex.JiraNamedField{Name: "Medium"},
				Subtasks: []imex.JiraIssue{
					{Key: "SUB-1"},
				},
			},
		},
		{
			Key: "SUB-1",
			Fields: imex.JiraFields{
				Summary:   "Subtask title",
				IssueType: imex.JiraNamedField{Name: "Sub-task"},
				Priority:  imex.JiraNamedField{Name: "Low"},
				Parent:    &imex.JiraIssue{Key: "EPIC-1"},
				Comment: &imex.JiraComments{
					Comments: []imex.JiraComment{
						{Body: "Imported comment"},
					},
				},
			},
		},
	}
	data, err := json.Marshal(payload)
	require.NoError(t, err)

	result, err := svc.ImportJiraJSON(ctx, projectID, userID, data)
	require.NoError(t, err)
	assert.Equal(t, 2, result.Imported)
	assert.Zero(t, result.Skipped)

	epic, err := client.Issue.Query().
		Where(issue.TitleEQ("Epic title")).
		Only(ctx)
	require.NoError(t, err)

	sub, err := client.Issue.Query().
		Where(issue.TitleEQ("Subtask title")).
		WithParent().
		Only(ctx)
	require.NoError(t, err)
	require.NotNil(t, sub.Edges.Parent)
	assert.Equal(t, epic.ID, sub.Edges.Parent.ID)

	comments, err := sub.QueryComments().All(ctx)
	require.NoError(t, err)
	require.Len(t, comments, 1)
	assert.Equal(t, "Imported comment", comments[0].BodyText)
}

func TestImportJiraJSON_SkipMissingSummary(t *testing.T) {
	svc, _, ctx, projectID, userID := newImexTestEnv(t)
	data := []byte(`[{"key":"BAD-1","fields":{"issuetype":{"name":"Task"},"priority":{"name":"Medium"}}}]`)

	result, err := svc.ImportJiraJSON(ctx, projectID, userID, data)
	require.NoError(t, err)
	assert.Zero(t, result.Imported)
	assert.Equal(t, 1, result.Skipped)
	require.Len(t, result.Errors, 1)
}

func TestImportJiraJSON_EmptyArray(t *testing.T) {
	svc, _, ctx, projectID, userID := newImexTestEnv(t)

	result, err := svc.ImportJiraJSON(ctx, projectID, userID, []byte(`[]`))
	require.NoError(t, err)
	assert.Zero(t, result.Imported)
	assert.Zero(t, result.Skipped)
}

func TestMapJiraIssueType_Extended(t *testing.T) {
	assert.Equal(t, string(issue.IssueTypeImprovement), imex.ExportMapJiraIssueType("Improvement"))
	assert.Equal(t, string(issue.IssueTypeSpike), imex.ExportMapJiraIssueType("Spike"))
	assert.Equal(t, string(issue.IssueTypeSubTask), imex.ExportMapJiraIssueType("subtask"))
}

func TestJiraDescriptionToText_DefaultBranch(t *testing.T) {
	assert.Equal(t, "42", imex.ExportJiraDescriptionToText(42))
}
