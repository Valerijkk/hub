//go:build hubtest

package tracker_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestEntRepository_ListActivities_Caps_P3BE06 — N>MaxIssueActivities → Limit.
// Mutation: снять Limit → FAIL (len > MaxIssueActivities).
func TestEntRepository_ListActivities_Caps_P3BE06(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Act cap", IssueType: "task", Priority: "medium", ReporterID: env.userID,
	})
	require.NoError(t, err)

	capN := tracker.MaxIssueActivities
	builders := make([]*ent.ActivityLogCreate, 0, capN+3)
	for i := 0; i < capN+3; i++ {
		builders = append(builders, env.client.ActivityLog.Create().
			SetIssueID(issue.ID).
			SetUserID(env.userID).
			SetAction(fmt.Sprintf("act-%d", i)))
	}
	_, err = env.client.ActivityLog.CreateBulk(builders...).Save(env.ctx)
	require.NoError(t, err)

	list, err := env.repo.ListActivities(env.ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, capN, len(list))
}

// TestEntRepository_ListComments_Caps_P3BE06 — N>MaxIssueComments → Limit.
func TestEntRepository_ListComments_Caps_P3BE06(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Cmt cap", IssueType: "task", Priority: "medium", ReporterID: env.userID,
	})
	require.NoError(t, err)

	capN := tracker.MaxIssueComments
	builders := make([]*ent.CommentCreate, 0, capN+3)
	for i := 0; i < capN+3; i++ {
		builders = append(builders, env.client.Comment.Create().
			SetIssueID(issue.ID).
			SetAuthorID(env.userID).
			SetBodyText(fmt.Sprintf("c-%d", i)).
			SetBodyJSON(map[string]interface{}{"type": "doc", "n": i}))
	}
	_, err = env.client.Comment.CreateBulk(builders...).Save(env.ctx)
	require.NoError(t, err)

	list, err := env.repo.ListComments(env.ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, capN, len(list))
}

// TestEntRepository_ListTimeLogs_Caps_P3BE06 — N>MaxIssueTimeLogs → Limit.
func TestEntRepository_ListTimeLogs_Caps_P3BE06(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	issue, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "TL cap", IssueType: "task", Priority: "medium", ReporterID: env.userID,
	})
	require.NoError(t, err)

	capN := tracker.MaxIssueTimeLogs
	builders := make([]*ent.TimeLogCreate, 0, capN+3)
	for i := 0; i < capN+3; i++ {
		builders = append(builders, env.client.TimeLog.Create().
			SetIssueID(issue.ID).
			SetUserID(env.userID).
			SetMinutes(1+i%5).
			SetDescription(fmt.Sprintf("tl-%d", i)))
	}
	_, err = env.client.TimeLog.CreateBulk(builders...).Save(env.ctx)
	require.NoError(t, err)

	list, err := env.repo.ListTimeLogs(env.ctx, issue.ID)
	require.NoError(t, err)
	assert.Equal(t, capN, len(list))
}
