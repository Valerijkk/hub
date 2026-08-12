package tms_test

import (
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestEntRepository_CreateBugIssue_NilTestCase_A3_009 — uuid.Nil не вызывает UpdateOneID(Nil).
// Mutation: всегда UpdateOneID(testCaseID) → ошибка/rollback.
func TestEntRepository_CreateBugIssue_NilTestCase_A3_009(t *testing.T) {
	repo, ctx, projectID, userID := newEntRepositoryTestEnv(t)

	bug, err := repo.CreateBugIssue(ctx, projectID, userID, uuid.Nil, "Orphan bug", "no tc", "high")
	require.NoError(t, err)
	require.NotNil(t, bug)
	assert.Equal(t, "Orphan bug", bug.Title)
	assert.NotEmpty(t, bug.IssueKey)
}
