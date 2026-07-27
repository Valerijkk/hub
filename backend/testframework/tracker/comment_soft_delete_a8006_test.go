package tracker_test

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent/comment"
	"github.com/valerii/hub/backend/ent/reaction"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_CommentSoftDelete_A8_006 — soft-delete hides comment from list,
// row survives with deleted_at, reactions are kept (not hard-deleted) but hidden via list.
// Mutation: Comment.Delete()/DeleteOneID → row gone → Get fails → FAIL.
// Mutation: drop DeletedAtIsNil from ListComments → soft-deleted still listed → FAIL.
func TestEntRepository_CommentSoftDelete_A8_006(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Soft-delete comment", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	cmt, err := env.repo.CreateComment(env.ctx, issue.ID, env.userID, tracker.CreateCommentInput{
		BodyJSON: map[string]interface{}{"type": "doc"},
		BodyText: "to soft-delete",
	})
	require.NoError(t, err)

	added, err := env.repo.ToggleReaction(env.ctx, issue.ID, cmt.ID, env.userID, "👍")
	require.NoError(t, err)
	assert.True(t, added)

	require.NoError(t, env.repo.DeleteComment(env.ctx, issue.ID, env.userID, cmt.ID))

	listed, err := env.repo.ListComments(env.ctx, issue.ID)
	require.NoError(t, err)
	assert.Empty(t, listed, "A8-006: soft-deleted comment must not appear in ListComments")

	// Row must still exist (soft, not hard delete).
	row, err := env.client.Comment.Get(env.ctx, cmt.ID)
	require.NoError(t, err, "A8-006: hard Delete would remove the row")
	require.NotNil(t, row.DeletedAt, "A8-006: deleted_at must be set")

	// Reactions kept on soft-deleted comment (no hard cascade).
	reactCount, err := env.client.Reaction.Query().
		Where(reaction.HasCommentWith(comment.IDEQ(cmt.ID))).
		Count(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, 1, reactCount, "A8-006: reactions must not be hard-deleted with comment")

	// Soft-deleted comment is not reactable / updateable.
	_, err = env.repo.ToggleReaction(env.ctx, issue.ID, cmt.ID, env.userID, "🎉")
	require.Error(t, err)
	_, err = env.repo.UpdateComment(env.ctx, issue.ID, env.userID, cmt.ID, tracker.UpdateCommentInput{
		BodyJSON: map[string]interface{}{"type": "doc"},
		BodyText: "hijack",
	})
	require.Error(t, err)

	// Idempotent soft-delete → NotFound (already deleted).
	err = env.repo.DeleteComment(env.ctx, issue.ID, env.userID, cmt.ID)
	require.Error(t, err)
}

// TestDeleteComment_NoHardDelete_Source_A8_006 — DeleteComment must soft-delete via
// SetDeletedAt; hard Comment.Delete / DeleteOneID must not appear.
// Mutation: restore Comment.Delete().Where → FAIL.
func TestDeleteComment_NoHardDelete_Source_A8_006(t *testing.T) {
	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	srcPath := filepath.Join(filepath.Dir(thisFile), "..", "..", "internal", "tracker", "repo_comments.go")
	src, err := os.ReadFile(srcPath)
	require.NoError(t, err)
	body := string(src)

	idx := strings.Index(body, "func (r *EntRepository) DeleteComment")
	require.Greater(t, idx, 0, "DeleteComment missing")
	chunk := body[idx:]
	end := strings.Index(chunk, "\nfunc (")
	if end > 0 {
		chunk = chunk[:end]
	}

	assert.Contains(t, chunk, "SetDeletedAt", "A8-006: DeleteComment must SetDeletedAt")
	assert.Contains(t, chunk, "DeletedAtIsNil", "A8-006: DeleteComment must skip already-deleted")
	assert.NotContains(t, chunk, "Comment.Delete()", "A8-006: hard Comment.Delete() forbidden")
	assert.NotContains(t, chunk, ".Delete().", "A8-006: hard Delete() chain forbidden in DeleteComment")
	assert.NotContains(t, chunk, "DeleteOneID", "A8-006: DeleteOneID forbidden in DeleteComment")
}
