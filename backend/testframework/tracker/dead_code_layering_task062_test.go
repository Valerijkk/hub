//go:build hubtest

package tracker_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestDeleteIssue_LogsDeletedActivity_Task062 — W1-A2-014 / P3-BE-07:
// userID больше не мёртвый; soft-delete пишет activity «deleted».
// Мутация: убрать LogActivity из DeleteIssue → assert на action краснеет.
func TestDeleteIssue_LogsDeletedActivity_Task062(t *testing.T) {
	repo := newMockRepo()
	svc := tracker.NewTrackerService(repo, nil)
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()
	iss := &ent.Issue{
		ID:       uuid.New(),
		IssueKey: "HUB-62",
		Edges:    ent.IssueEdges{Project: &ent.Project{ID: projectID}},
	}
	repo.issues[iss.ID] = iss
	repo.issuesByKey[iss.IssueKey] = iss

	require.NoError(t, svc.DeleteIssue(ctx, projectID, "HUB-62", userID))

	require.Len(t, repo.activities, 1)
	assert.Equal(t, "deleted", repo.activities[0].Action)
	require.NotNil(t, repo.activities[0].Edges.User)
	assert.Equal(t, userID, repo.activities[0].Edges.User.ID)
	_, still := repo.issues[iss.ID]
	assert.False(t, still, "issue must be soft-deleted after activity")
}

// TestListWorkflowTransitions_SkipsNilEdges_Task062 — W1-A2-015 / P3-BE-07:
// nil FromStatus/ToStatus не паникует, строка пропускается.
func TestListWorkflowTransitions_SkipsNilEdges_Task062(t *testing.T) {
	repo := newMockRepo()
	svc := tracker.NewTrackerService(repo, nil)
	projectID := uuid.New()
	from := &ent.WorkflowStatus{ID: uuid.New()}
	to := &ent.WorkflowStatus{ID: uuid.New()}
	good := &ent.WorkflowTransition{ID: uuid.New()}
	good.Edges.FromStatus = from
	good.Edges.ToStatus = to
	bad := &ent.WorkflowTransition{ID: uuid.New()} // edges nil
	repo.transitions["bad"] = bad
	repo.transitions[from.ID.String()+":"+to.ID.String()] = good

	list, err := svc.ListWorkflowTransitions(context.Background(), projectID)
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, good.ID.String(), list[0].ID)
	assert.Equal(t, from.ID.String(), list[0].FromStatusID)
	assert.Equal(t, to.ID.String(), list[0].ToStatusID)
}
