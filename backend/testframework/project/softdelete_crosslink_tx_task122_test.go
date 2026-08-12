//go:build hubtest

package project_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/issuelink"
	entproject "github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/workflowstatus"
	"github.com/valerii/hub/backend/internal/project"
)

// TestSoftDeleteWithCrossLinkCleanup_TxRollbackOnMidFail_A2W1_A5_001 —
// Ent twin: mid-fail после DeleteIssueLink → rollback, project не soft-deleted, link остаётся.
// Mutation: SoftDeleteWithCrossLinkCleanup без общей Ent-tx → после inject-fail link удалён.
func TestSoftDeleteWithCrossLinkCleanup_TxRollbackOnMidFail_A2W1_A5_001(t *testing.T) {
	env := newProjectRepoEnv(t)

	projA, err := env.repo.Create(env.ctx, "AlphaTx", "ATX", "", env.owner.ID)
	require.NoError(t, err)
	projB, err := env.repo.Create(env.ctx, "BetaTx", "BTX", "", env.owner.ID)
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

	src := createTestIssue(t, env, projA.ID, 1, "ATX-1", statusA.ID)
	tgt := createTestIssue(t, env, projB.ID, 1, "BTX-1", statusB.ID)
	link, err := env.client.IssueLink.Create().
		SetLinkType(issuelink.LinkTypeRelatesTo).
		SetSourceID(src.ID).
		SetTargetID(tgt.ID).
		Save(env.ctx)
	require.NoError(t, err)

	project.ExportSetAfterCrossLinkDeleteHook(func(context.Context) error {
		return errors.New("injected mid-fail")
	})
	t.Cleanup(func() { project.ExportSetAfterCrossLinkDeleteHook(nil) })

	_, _, err = env.repo.SoftDeleteWithCrossLinkCleanup(env.ctx, projA.ID)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "injected mid-fail")

	got, err := env.client.Project.Get(env.ctx, projA.ID)
	require.NoError(t, err)
	assert.Equal(t, entproject.StatusActive, got.Status, "mid-fail must not soft-delete project")
	assert.Nil(t, got.DeletedAt)

	_, err = env.client.IssueLink.Get(env.ctx, link.ID)
	require.NoError(t, err, "mid-fail must roll back DeleteIssueLink")
}

// TestSoftDeleteWithCrossLinkCleanup_AtomicSuccess_A2W1_A5_001 — happy path Ent.
func TestSoftDeleteWithCrossLinkCleanup_AtomicSuccess_A2W1_A5_001(t *testing.T) {
	env := newProjectRepoEnv(t)

	projA, err := env.repo.Create(env.ctx, "AlphaOK", "AOK", "", env.owner.ID)
	require.NoError(t, err)
	projB, err := env.repo.Create(env.ctx, "BetaOK", "BOK", "", env.owner.ID)
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

	src := createTestIssue(t, env, projA.ID, 1, "AOK-1", statusA.ID)
	tgt := createTestIssue(t, env, projB.ID, 1, "BOK-1", statusB.ID)
	link, err := env.client.IssueLink.Create().
		SetLinkType(issuelink.LinkTypeRelatesTo).
		SetSourceID(src.ID).
		SetTargetID(tgt.ID).
		Save(env.ctx)
	require.NoError(t, err)

	outgoing, incoming, err := env.repo.SoftDeleteWithCrossLinkCleanup(env.ctx, projA.ID)
	require.NoError(t, err)
	assert.Len(t, outgoing, 1)
	assert.Empty(t, incoming)

	got, err := env.client.Project.Get(env.ctx, projA.ID)
	require.NoError(t, err)
	assert.Equal(t, entproject.StatusDeleted, got.Status)
	require.NotNil(t, got.DeletedAt)

	_, err = env.client.IssueLink.Get(env.ctx, link.ID)
	require.Error(t, err)
}
