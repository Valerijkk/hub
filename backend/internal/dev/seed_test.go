package dev

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDemoUsersValid(t *testing.T) {
	assert.NotEmpty(t, demoUsers, "demoUsers should not be empty")

	emails := make(map[string]bool)
	for _, u := range demoUsers {
		t.Run(u.Email, func(t *testing.T) {
			assert.NotEmpty(t, u.Email, "email must not be empty")
			assert.NotEmpty(t, u.Name, "name must not be empty")
			assert.NotEmpty(t, u.DisplayName, "display name must not be empty")
			assert.NotEmpty(t, u.SystemRole, "system role must not be empty")
			assert.False(t, emails[u.Email], "duplicate email: %s", u.Email)
			emails[u.Email] = true
		})
	}
}

func TestDemoStatusesValid(t *testing.T) {
	assert.NotEmpty(t, demoStatuses, "demoStatuses should not be empty")

	names := make(map[string]bool)
	for _, s := range demoStatuses {
		t.Run(s.Name, func(t *testing.T) {
			assert.NotEmpty(t, s.Name, "status name must not be empty")
			assert.NotEmpty(t, s.Color, "status color must not be empty")
			assert.NotEmpty(t, s.Category, "status category must not be empty")
			assert.Greater(t, s.Order, 0, "sort order must be positive")
			assert.False(t, names[s.Name], "duplicate status name: %s", s.Name)
			names[s.Name] = true
		})
	}
}

func TestDemoIssuesValid(t *testing.T) {
	assert.NotEmpty(t, demoIssues, "demoIssues should not be empty")

	keys := make(map[string]bool)
	for _, iss := range demoIssues {
		t.Run(iss.Key, func(t *testing.T) {
			assert.NotEmpty(t, iss.Key, "issue key must not be empty")
			assert.NotEmpty(t, iss.Title, "issue title must not be empty")
			assert.NotEmpty(t, iss.IssueType, "issue type must not be empty")
			assert.NotEmpty(t, iss.Priority, "issue priority must not be empty")
			assert.NotEmpty(t, iss.StatusName, "issue status name must not be empty")
			assert.NotEmpty(t, iss.AssigneeEmail, "issue assignee email must not be empty")
			assert.False(t, keys[iss.Key], "duplicate issue key: %s", iss.Key)
			keys[iss.Key] = true
		})
	}
}

func TestDemoTestCasesValid(t *testing.T) {
	assert.NotEmpty(t, demoTestCases, "demoTestCases should not be empty")

	for _, tc := range demoTestCases {
		t.Run(tc.CaseKey, func(t *testing.T) {
			assert.NotEmpty(t, tc.CaseKey, "test case key must not be empty")
			assert.NotEmpty(t, tc.Title, "test case title must not be empty")
			assert.NotEmpty(t, tc.Steps, "test case must have steps")
			for i, step := range tc.Steps {
				assert.NotEmpty(t, step.Action, "step %d action must not be empty", i)
				assert.NotEmpty(t, step.ExpectedResult, "step %d expected result must not be empty", i)
			}
		})
	}
}

func TestDemoChannelsValid(t *testing.T) {
	assert.NotEmpty(t, demoChannels, "demoChannels should not be empty")

	for _, ch := range demoChannels {
		t.Run(ch.Name, func(t *testing.T) {
			assert.NotEmpty(t, ch.Name, "channel name must not be empty")
			assert.NotEmpty(t, ch.Type, "channel type must not be empty")
		})
	}
}
