package gitlab_test
import (
	"github.com/valerii/hub/backend/internal/gitlab"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

func TestFetchMRForReview_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "token", r.Header.Get("PRIVATE-TOKEN"))
		if strings.Contains(r.URL.Path, "/changes") {
			require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{
				"changes": []map[string]interface{}{
					{"new_path": "main.go", "diff": "+line"},
				},
			}))
			return
		}
		require.NoError(t, json.NewEncoder(w).Encode(map[string]interface{}{
			"web_url":       "https://gitlab.example/group/project/-/merge_requests/1",
			"title":         "Fix login",
			"description":   "details",
			"source_branch": "feat",
			"target_branch": "main",
			"author":        map[string]interface{}{"username": "dev"},
		}))
	}))
	t.Cleanup(srv.Close)

	repo := &gitlabMockRepo{
		integrations: map[string]*ent.GitLabIntegration{
			"group/project": {
				GitlabURL:   srv.URL,
				ProjectPath: "group/project",
				AccessToken: "token",
			},
		},
	}
	svc := gitlab.NewGitLabServiceWithHTTPClient(repo, nil, srv.Client())

	mrURL := srv.URL + "/group/project/-/merge_requests/1"
	out, err := svc.FetchMRForReview(context.Background(), mrURL)
	require.NoError(t, err)
	require.NotNil(t, out)
	assert.Equal(t, "Fix login", out.Title)
	assert.Equal(t, "dev", out.Author)
	assert.Equal(t, "feat", out.SourceBranch)
	assert.Contains(t, out.ChangedFiles, "main.go")
	assert.Contains(t, out.PatchSummary, "main.go")
}

func TestFetchMRForReview_InvalidURL(t *testing.T) {
	svc := gitlab.NewGitLabService(&gitlabMockRepo{}, nil)
	_, err := svc.FetchMRForReview(context.Background(), "not-a-url")
	require.Error(t, err)
}

func TestFetchMRForReview_NoIntegration(t *testing.T) {
	svc := gitlab.NewGitLabService(&gitlabMockRepo{}, nil)
	_, err := svc.FetchMRForReview(context.Background(), "https://gitlab.com/missing/project/-/merge_requests/1")
	require.Error(t, err)
}

func TestParseGitLabMRURL(t *testing.T) {
	host, path, iid, err := gitlab.ExportParseGitLabMRURL("https://gitlab.com/group/sub/-/merge_requests/42")
	require.NoError(t, err)
	assert.Equal(t, "https://gitlab.com", host)
	assert.Equal(t, "group/sub", path)
	assert.Equal(t, 42, iid)
}
