package github

import (
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/githublink"
)

func TestToIntegrationResponse(t *testing.T) {
	mergeStatusID := uuid.New()
	now := time.Now().UTC()
	ig := &ent.GitHubIntegration{
		ID:            uuid.New(),
		RepoOwner:     "hub-org",
		RepoName:      "hub",
		MergeStatusID: &mergeStatusID,
		CreatedAt:     now,
	}

	resp := toIntegrationResponse(ig)
	require.NotNil(t, resp)
	assert.Equal(t, ig.ID.String(), resp.ID)
	assert.Equal(t, "hub-org", resp.RepoOwner)
	assert.Equal(t, "hub", resp.RepoName)
	assert.Equal(t, mergeStatusID.String(), resp.MergeStatusID)
	assert.Equal(t, now, resp.CreatedAt)
}

func TestToIntegrationResponse_NoMergeStatus(t *testing.T) {
	ig := &ent.GitHubIntegration{
		ID:        uuid.New(),
		RepoOwner: "owner",
		RepoName:  "repo",
		CreatedAt: time.Now(),
	}
	resp := toIntegrationResponse(ig)
	assert.Empty(t, resp.MergeStatusID)
}

func TestToGitHubLinkResponse(t *testing.T) {
	now := time.Now().UTC()
	link := &ent.GitHubLink{
		ID:         uuid.New(),
		LinkType:   githublink.LinkTypePullRequest,
		ExternalID: "42",
		Title:      "Fix login",
		URL:        "https://github.com/org/repo/pull/42",
		Status:     githublink.StatusOpen,
		Author:     "dev",
		Branch:     "feature/login",
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	resp := toGitHubLinkResponse(link)
	require.NotNil(t, resp)
	assert.Equal(t, "pull_request", resp.LinkType)
	assert.Equal(t, "42", resp.ExternalID)
	assert.Equal(t, "dev", resp.Author)
	assert.Equal(t, "feature/login", resp.Branch)
}
