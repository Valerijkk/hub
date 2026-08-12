package github_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/internal/github"
)

// membershipPreviewRepo — mock с управлением IsProjectMember для A3W7-3-003.
type membershipPreviewRepo struct {
	previewMockRepo
	memberIDs map[uuid.UUID]bool
	projectID uuid.UUID
}

func (m *membershipPreviewRepo) IsProjectMember(_ context.Context, userID, projectID uuid.UUID) (bool, error) {
	if projectID != m.projectID {
		return false, nil
	}
	return m.memberIDs[userID], nil
}

// TestGetLinkPreview_IssueMembership_A3W73_003 — non-member не видит title/status/assignee;
// member получает issue preview. Mutation: убрать IsProjectMember → foreign утекает → FAIL.
func TestGetLinkPreview_IssueMembership_A3W73_003(t *testing.T) {
	projectID := uuid.New()
	memberID := uuid.New()
	foreignID := uuid.New()
	secretTitle := "SECRET-FOREIGN-TITLE"
	secretStatus := "In Progress"
	secretAssignee := "Bob Secret"

	repo := &membershipPreviewRepo{
		previewMockRepo: previewMockRepo{
			issues: map[string]*ent.Issue{
				"HUB-99": {
					ID:        uuid.New(),
					IssueKey:  "HUB-99",
					Title:     secretTitle,
					IssueType: issue.IssueTypeBug,
					Edges: ent.IssueEdges{
						Project:  &ent.Project{ID: projectID},
						Status:   &ent.WorkflowStatus{Name: secretStatus},
						Assignee: &ent.User{Name: secretAssignee, AvatarURL: "b.png"},
					},
				},
			},
		},
		memberIDs: map[uuid.UUID]bool{memberID: true},
		projectID: projectID,
	}
	svc := github.NewLinkPreviewService(repo, nil)
	rawURL := "https://hub.local/projects/x/issues/HUB-99"

	// Foreign / non-member → generic external, без утечки полей issue.
	foreignResp, err := svc.GetLinkPreview(context.Background(), foreignID, rawURL)
	require.NoError(t, err)
	require.NotNil(t, foreignResp)
	assert.Equal(t, "external", foreignResp.Type)
	assert.NotEqual(t, secretTitle, foreignResp.Title)
	assert.Empty(t, foreignResp.IssueKey)
	assert.Empty(t, foreignResp.Status)
	assert.Empty(t, foreignResp.AssigneeName)
	assert.NotContains(t, foreignResp.Title, secretTitle)
	assert.NotContains(t, foreignResp.Description, secretTitle)

	// Member → полный issue preview.
	memberResp, err := svc.GetLinkPreview(context.Background(), memberID, rawURL)
	require.NoError(t, err)
	require.NotNil(t, memberResp)
	assert.Equal(t, "issue", memberResp.Type)
	assert.Equal(t, secretTitle, memberResp.Title)
	assert.Equal(t, "HUB-99", memberResp.IssueKey)
	assert.Equal(t, secretStatus, memberResp.Status)
	assert.Equal(t, secretAssignee, memberResp.AssigneeName)
}
