package tracker_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/tracker"
)

// TestCreateIssue_RejectsNonMemberAssignee_A2W1_A2_002 — assignee вне проекта → 400.
// Mutation: убрать IsProjectMember check → create успешен с чужим UUID.
func TestCreateIssue_RejectsNonMemberAssignee_A2W1_A2_002(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	reporter := uuid.New()
	outsider := uuid.New()
	member := uuid.New()

	repo.projectMembers = map[uuid.UUID]map[uuid.UUID]bool{
		projectID: {member: true, reporter: true},
	}

	_, err := svc.CreateIssue(ctx, projectID, reporter, tracker.CreateIssueRequest{
		Title: "X", IssueType: "task", AssigneeID: outsider.String(),
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)

	ok, err := svc.CreateIssue(ctx, projectID, reporter, tracker.CreateIssueRequest{
		Title: "Y", IssueType: "task", AssigneeID: member.String(),
	})
	require.NoError(t, err)
	require.NotNil(t, ok)
}

// TestUpdateIssue_RejectsNonMemberAssignee_A2W1_A2_002 — update assignee вне проекта → 400.
// Mutation: убрать IsProjectMember в UpdateIssue → update принимает чужой UUID.
func TestUpdateIssue_RejectsNonMemberAssignee_A2W1_A2_002(t *testing.T) {
	svc, repo := newTestService()
	ctx := context.Background()
	projectID := uuid.New()
	reporter := uuid.New()
	outsider := uuid.New()
	member := uuid.New()

	repo.projectMembers = map[uuid.UUID]map[uuid.UUID]bool{
		projectID: {member: true, reporter: true},
	}

	created, err := svc.CreateIssue(ctx, projectID, reporter, tracker.CreateIssueRequest{
		Title: "Z", IssueType: "task",
	})
	require.NoError(t, err)

	outsiderStr := outsider.String()
	_, err = svc.UpdateIssue(ctx, projectID, created.IssueKey, reporter, tracker.UpdateIssueRequest{
		AssigneeID: &outsiderStr,
		Version:    created.Version,
	})
	require.Error(t, err)
	var appErr *apperrors.AppError
	require.ErrorAs(t, err, &appErr)
	assert.Equal(t, 400, appErr.Code)

	memberStr := member.String()
	updated, err := svc.UpdateIssue(ctx, projectID, created.IssueKey, reporter, tracker.UpdateIssueRequest{
		AssigneeID: &memberStr,
		Version:    created.Version,
	})
	require.NoError(t, err)
	require.NotNil(t, updated)
}
