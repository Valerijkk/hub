package imex

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/issue"
)

func TestMapJiraIssueType(t *testing.T) {
	assert.Equal(t, string(issue.IssueTypeBug), mapJiraIssueType("Bug"))
	assert.Equal(t, string(issue.IssueTypeStory), mapJiraIssueType("Story"))
	assert.Equal(t, string(issue.IssueTypeEpic), mapJiraIssueType("Epic"))
	assert.Equal(t, string(issue.IssueTypeSubTask), mapJiraIssueType("Sub-task"))
	assert.Equal(t, string(issue.IssueTypeTechDebt), mapJiraIssueType("Technical Debt"))
	assert.Equal(t, string(issue.IssueTypeTask), mapJiraIssueType("Unknown"))
}

func TestMapJiraPriority(t *testing.T) {
	assert.Equal(t, string(issue.PriorityCritical), mapJiraPriority("Highest"))
	assert.Equal(t, string(issue.PriorityHigh), mapJiraPriority("High"))
	assert.Equal(t, string(issue.PriorityMedium), mapJiraPriority("Normal"))
	assert.Equal(t, string(issue.PriorityLow), mapJiraPriority("Low"))
	assert.Equal(t, string(issue.PriorityTrivial), mapJiraPriority("Lowest"))
	assert.Equal(t, string(issue.PriorityMedium), mapJiraPriority(""))
}

func TestJiraDescriptionToText(t *testing.T) {
	assert.Equal(t, "plain text", jiraDescriptionToText("plain text"))
	assert.Equal(t, "", jiraDescriptionToText(nil))

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
	assert.Contains(t, jiraDescriptionToText(adf), "Hello Jira")
}

func TestParseJiraJSON_Formats(t *testing.T) {
	issueObj := `{"key":"PROJ-1","fields":{"summary":"Fix bug","issuetype":{"name":"Bug"},"priority":{"name":"High"}}}`

	t.Run("array", func(t *testing.T) {
		data := `[` + issueObj + `]`
		issues, err := parseJiraJSON([]byte(data))
		require.NoError(t, err)
		require.Len(t, issues, 1)
		assert.Equal(t, "PROJ-1", issues[0].Key)
	})

	t.Run("search response", func(t *testing.T) {
		data := `{"issues":[` + issueObj + `]}`
		issues, err := parseJiraJSON([]byte(data))
		require.NoError(t, err)
		require.Len(t, issues, 1)
	})

	t.Run("single issue", func(t *testing.T) {
		issues, err := parseJiraJSON([]byte(issueObj))
		require.NoError(t, err)
		require.Len(t, issues, 1)
	})

	t.Run("invalid", func(t *testing.T) {
		_, err := parseJiraJSON([]byte(`{"foo":1}`))
		require.Error(t, err)
	})
}

func TestParseJiraJSON_RoundTrip(t *testing.T) {
	raw := []JiraIssue{{
		Key: "HUB-99",
		Fields: JiraFields{
			Summary:     "Import me",
			IssueType:   JiraNamedField{Name: "Task"},
			Priority:    JiraNamedField{Name: "Medium"},
			Description: "desc",
		},
	}}
	data, err := json.Marshal(raw)
	require.NoError(t, err)

	parsed, err := parseJiraJSON(data)
	require.NoError(t, err)
	assert.Equal(t, "HUB-99", parsed[0].Key)
	assert.Equal(t, "Import me", parsed[0].Fields.Summary)
}
