package github_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/testcase"
	"github.com/valerii/hub/backend/internal/github"
)

// TestGetLinkPreview_TestCaseMembership_CW19_FE_002 — non-member не получает type=test_case
// с title/ключом; member получает TC preview.
// Mutation: убрать IsProjectMember / вернуть Type+Title по regex → foreign enum keys → FAIL.
func TestGetLinkPreview_TestCaseMembership_CW19_FE_002(t *testing.T) {
	projectID := uuid.New()
	memberID := uuid.New()
	foreignID := uuid.New()
	secretTitle := "SECRET-TC-TITLE"
	caseKey := "HUB-TC-77"

	repo := &membershipPreviewRepo{
		previewMockRepo: previewMockRepo{
			testCases: map[string]*ent.TestCase{
				caseKey: {
					ID:      uuid.New(),
					CaseKey: caseKey,
					Title:   secretTitle,
					Status:  testcase.StatusActive,
					Edges:   ent.TestCaseEdges{Project: &ent.Project{ID: projectID}},
				},
			},
		},
		memberIDs: map[uuid.UUID]bool{memberID: true},
		projectID: projectID,
	}
	svc := github.NewLinkPreviewService(repo, nil)
	rawURL := "https://hub.local/projects/x/tms/cases/test-cases/" + caseKey

	foreignResp, err := svc.GetLinkPreview(context.Background(), foreignID, rawURL)
	require.NoError(t, err)
	require.NotNil(t, foreignResp)
	assert.Equal(t, "external", foreignResp.Type)
	assert.NotEqual(t, secretTitle, foreignResp.Title)
	assert.Empty(t, foreignResp.Status)
	assert.NotContains(t, foreignResp.Description, secretTitle)

	memberResp, err := svc.GetLinkPreview(context.Background(), memberID, rawURL)
	require.NoError(t, err)
	require.NotNil(t, memberResp)
	assert.Equal(t, "test_case", memberResp.Type)
	assert.Equal(t, secretTitle, memberResp.Title)
	assert.Equal(t, "active", memberResp.Status)
}
