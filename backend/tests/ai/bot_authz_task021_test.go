package ai_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/internal/ai"
	githubmod "github.com/valerii/hub/backend/internal/github"
	gitlabmod "github.com/valerii/hub/backend/internal/gitlab"
	trackermod "github.com/valerii/hub/backend/internal/tracker"
)

type stubAccess struct {
	allow map[uuid.UUID]bool
	err   error
}

func (s stubAccess) CanViewProject(_ context.Context, _, projectID uuid.UUID) (bool, error) {
	if s.err != nil {
		return false, s.err
	}
	if s.allow == nil {
		return false, nil
	}
	return s.allow[projectID], nil
}

type gitlabReviewRoundTripper struct {
	paths  []string
	tokens []string
}

type failFastRoundTripper struct {
	calls int
}

func (r *failFastRoundTripper) RoundTrip(*http.Request) (*http.Response, error) {
	r.calls++
	return nil, errors.New("unexpected HTTP request")
}

func (r *gitlabReviewRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	r.paths = append(r.paths, req.URL.EscapedPath())
	r.tokens = append(r.tokens, req.Header.Get("PRIVATE-TOKEN"))

	var body string
	status := http.StatusOK
	switch req.URL.EscapedPath() {
	case "/api/v4/projects/group%2Fapp/merge_requests/1":
		body = `{"web_url":"https://gitlab.example/group/app/-/merge_requests/1","title":"Allowed review","description":"Review body","source_branch":"feature","target_branch":"main","author":{"username":"alice"}}`
	case "/api/v4/projects/group%2Fapp/merge_requests/1/changes":
		body = `{"changes":[{"new_path":"main.go","diff":"+package main"}]}`
	default:
		status = http.StatusNotFound
		body = `{"message":"unexpected GitLab API path"}`
	}
	return &http.Response{
		StatusCode: status,
		Header:     make(http.Header),
		Body:       io.NopCloser(strings.NewReader(body)),
		Request:    req,
	}, nil
}

func TestUnifiedCodeReviewProvider_DeniesForeignProject(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	ghRepo := &githubReviewRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"acme/app": {
				RepoOwner: "acme",
				RepoName:  "app",
				AccessToken: "tok",
				Edges: ent.GitHubIntegrationEdges{
					Project: &ent.Project{ID: projectID},
				},
			},
		},
	}
	transport := &failFastRoundTripper{}
	ghSvc := githubmod.NewGitHubServiceWithHTTPClient(ghRepo, nil, &http.Client{Transport: transport}, "")
	p := ai.NewUnifiedCodeReviewProvider(ghSvc, nil)
	p.SetAccessChecker(stubAccess{allow: map[uuid.UUID]bool{projectID: false}})

	_, err := p.FetchReview(context.Background(), userID, "https://github.com/acme/app/pull/1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unavailable")
	assert.Zero(t, transport.calls)
}

func TestUnifiedCodeReviewProvider_AllowsProjectMember(t *testing.T) {
	projectID := uuid.New()
	userID := uuid.New()
	srvPath := "group/app"
	glRepo := &gitlabReviewRepo{
		integrations: map[string]*ent.GitLabIntegration{
			srvPath: {
				GitlabURL:   "https://gitlab.example",
				ProjectPath: srvPath,
				AccessToken: "tok",
				Edges: ent.GitLabIntegrationEdges{
					Project: &ent.Project{ID: projectID},
				},
			},
		},
	}
	transport := &gitlabReviewRoundTripper{}
	glSvc := gitlabmod.NewGitLabServiceWithHTTPClient(glRepo, nil, &http.Client{Transport: transport})
	p := ai.NewUnifiedCodeReviewProvider(nil, glSvc)
	p.SetAccessChecker(stubAccess{allow: map[uuid.UUID]bool{projectID: true}})

	review, err := p.FetchReview(context.Background(), userID, "https://gitlab.example/group/app/-/merge_requests/1")
	require.NoError(t, err)
	require.NotNil(t, review)
	assert.Equal(t, "Allowed review", review.Title)
	assert.Equal(t, []string{"main.go"}, review.ChangedFiles)
	assert.Equal(t, []string{
		"/api/v4/projects/group%2Fapp/merge_requests/1",
		"/api/v4/projects/group%2Fapp/merge_requests/1/changes",
	}, transport.paths)
	assert.Equal(t, []string{"tok", "tok"}, transport.tokens)
}

func TestRBACProjectAccess_NilSvcDenies(t *testing.T) {
	a := ai.RBACProjectAccess{}
	ok, err := a.CanViewProject(context.Background(), uuid.New(), uuid.New())
	require.NoError(t, err)
	assert.False(t, ok)
}

type stubProjectByKey struct {
	proj *ent.Project
	err  error
}

func (s stubProjectByKey) FindByKey(context.Context, string) (*ent.Project, error) {
	if s.err != nil {
		return nil, s.err
	}
	return s.proj, nil
}

func TestTrackerIssueLookup_DeniesForeignProject(t *testing.T) {
	// P1-BE-31 / W5-A3-002: guard на ProjectAccessChecker до GetIssue.
	projID := uuid.New()
	l := ai.NewTrackerIssueLookup(&trackermod.TrackerService{}, stubProjectByKey{
		proj: &ent.Project{ID: projID, Key: "HUB"},
	})
	l.SetAccessChecker(stubAccess{allow: map[uuid.UUID]bool{projID: false}})

	title, desc, err := l.IssueContext(context.Background(), uuid.New(), "HUB-42")
	require.Error(t, err)
	assert.True(t, ent.IsNotFound(err))
	assert.Empty(t, title)
	assert.Empty(t, desc)
}
