package github_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
)

// TestFindIssueByKey_ExcludesTrash_A3W73_002 — soft-deleted issue → NotFound (webhook/branch/link).
// Mutation: убрать issue.DeletedAtIsNil() из FindIssueByKey → trash находится → FAIL.
func TestFindIssueByKey_ExcludesTrash_A3W73_002(t *testing.T) {
	env := newGitHubRepoEnv(t)
	iss := env.createIssue(t)

	_, err := env.client.Issue.UpdateOneID(iss.ID).
		SetDeletedAt(time.Now().UTC()).
		Save(env.ctx)
	require.NoError(t, err)

	_, err = env.repo.FindIssueByKey(env.ctx, iss.IssueKey)
	require.True(t, ent.IsNotFound(err), "trashed issue must be NotFound via FindIssueByKey, got %v", err)

	_, err = env.repo.FindIssueByKeyInProject(env.ctx, env.projectID, iss.IssueKey)
	require.True(t, ent.IsNotFound(err), "trashed issue must be NotFound via FindIssueByKeyInProject, got %v", err)
}
