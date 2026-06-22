package imex_test
import (
	"github.com/valerii/hub/backend/internal/imex"
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/issue"
)

func TestMapJiraIssueType(t *testing.T) {
	assert.Equal(t, string(issue.IssueTypeBug), imex.ExportMapJiraIssueType("Bug"))
	assert.Equal(t, string(issue.IssueTypeStory), imex.ExportMapJiraIssueType("Story"))
	assert.Equal(t, string(issue.IssueTypeEpic), imex.ExportMapJiraIssueType("Epic"))
	assert.Equal(t, string(issue.IssueTypeSubTask), imex.ExportMapJiraIssueType("Sub-task"))
	assert.Equal(t, string(issue.IssueTypeTechDebt), imex.ExportMapJiraIssueType("Technical Debt"))
	assert.Equal(t, string(issue.IssueTypeTask), imex.ExportMapJiraIssueType("Unknown"))
}

func TestMapJiraPriority(t *testing.T) {
	assert.Equal(t, string(issue.PriorityCritical), imex.ExportMapJiraPriority("Highest"))
	assert.Equal(t, string(issue.PriorityHigh), imex.ExportMapJiraPriority("High"))
	assert.Equal(t, string(issue.PriorityMedium), imex.ExportMapJiraPriority("Normal"))
	assert.Equal(t, string(issue.PriorityLow), imex.ExportMapJiraPriority("Low"))
	assert.Equal(t, string(issue.PriorityTrivial), imex.ExportMapJiraPriority("Lowest"))
	assert.Equal(t, string(issue.PriorityMedium), imex.ExportMapJiraPriority(""))
}

func TestJiraDescriptionToText(t *testing.T) {
	assert.Equal(t, "plain text", imex.ExportJiraDescriptionToText("plain text"))
	assert.Equal(t, "", imex.ExportJiraDescriptionToText(nil))

	adf := map[string]interface{}{
		"type": "doc",
		"content": []interface{}{
			map[string]interface{}{
				"type": "paragraph",
				"content": []interface{}{
					map[string]interface{}{"text": "Hello"},
					map[string]interface{}{"text": " Jira"},
				},
			},
		},
	}
	assert.Contains(t, imex.ExportJiraDescriptionToText(adf), "Hello Jira")
}

func TestParseJiraJSON_Formats(t *testing.T) {
	issueObj := `{"key":"PROJ-1","fields":{"summary":"Fix bug","issuetype":{"name":"Bug"},"priority":{"name":"High"}}}`

	t.Run("array", func(t *testing.T) {
		data := `[` + issueObj + `]`
		issues, err := imex.ExportParseJiraJSON([]byte(data))
		require.NoError(t, err)
		require.Len(t, issues, 1)
		assert.Equal(t, "PROJ-1", issues[0].Key)
	})

	t.Run("search response", func(t *testing.T) {
		data := `{"issues":[` + issueObj + `]}`
		issues, err := imex.ExportParseJiraJSON([]byte(data))
		require.NoError(t, err)
		require.Len(t, issues, 1)
	})

	t.Run("single issue", func(t *testing.T) {
		issues, err := imex.ExportParseJiraJSON([]byte(issueObj))
		require.NoError(t, err)
		require.Len(t, issues, 1)
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := imex.ExportParseJiraJSON([]byte(`{"foo":1}`))
		require.Error(t, err)
	})

	t.Run("empty array", func(t *testing.T) {
		issues, err := imex.ExportParseJiraJSON([]byte(`[]`))
		require.NoError(t, err)
		assert.Empty(t, issues)
	})
}

func TestParseJiraJSON_RoundTrip(t *testing.T) {
	raw := []imex.JiraIssue{{
		Key: "HUB-99",
		Fields: imex.JiraFields{
			Summary:     "Import me",
			IssueType:   imex.JiraNamedField{Name: "Task"},
			Priority:    imex.JiraNamedField{Name: "Medium"},
			Description: "desc",
		},
	}}
	data, err := json.Marshal(raw)
	require.NoError(t, err)

	parsed, err := imex.ExportParseJiraJSON(data)
	require.NoError(t, err)
	assert.Equal(t, "HUB-99", parsed[0].Key)
	assert.Equal(t, "Import me", parsed[0].Fields.Summary)
}
