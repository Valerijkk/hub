package github_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/github"
)

// TestEntRepository_IntegrationSoftDelete_W7_A3_010 — DeleteIntegration soft-deletes; list hides.
// Mutation: hard Delete().Exec → row gone → FAIL.
func TestEntRepository_IntegrationSoftDelete_W7_A3_010(t *testing.T) {
	env := newGitHubRepoEnv(t)

	ig, err := env.repo.CreateIntegration(env.ctx, env.projectID, github.CreateIntegrationInput{
		RepoOwner: "acme", RepoName: "hub-trash",
		AccessToken: "tok", WebhookSecret: "sec",
	})
	require.NoError(t, err)

	require.NoError(t, env.repo.DeleteIntegration(env.ctx, env.projectID, ig.ID))

	list, err := env.repo.ListIntegrations(env.ctx, env.projectID)
	require.NoError(t, err)
	for _, item := range list {
		assert.NotEqual(t, ig.ID, item.ID)
	}

	_, err = env.repo.GetIntegration(env.ctx, ig.ID)
	require.Error(t, err)

	raw, err := env.client.GitHubIntegration.Get(env.ctx, ig.ID)
	require.NoError(t, err)
	require.NotNil(t, raw.DeletedAt)

	_, thisFile, _, ok := runtime.Caller(0)
	require.True(t, ok)
	backendRoot := filepath.Clean(filepath.Join(filepath.Dir(thisFile), "..", ".."))
	src, err := os.ReadFile(filepath.Join(backendRoot, "internal", "github", "repository.go"))
	require.NoError(t, err)
	assert.Contains(t, string(src), "SetDeletedAt")
	assert.NotContains(t, string(src), "GitHubIntegration.Delete()")
}
