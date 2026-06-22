package github_test

import (
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/internal/github"
)

// TestNewGitHubService_DefaultUsesSafeHTTPClient проверяет, что конструктор
// без явного client не падает на http.DefaultClient и блокирует loopback (SEC-04).
func TestNewGitHubService_DefaultUsesSafeHTTPClient(t *testing.T) {
	svc := github.NewGitHubService(&githubMockRepo{}, nil)
	client := github.ExportGitHubServiceHTTPClient(svc)
	require.NotNil(t, client)

	_, err := client.Get("http://127.0.0.1:65500/")
	require.Error(t, err, "default GitHub HTTP client must block private/loopback addresses")
}
