package ai_test

import (
	"context"
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
	ghSvc := githubmod.NewGitHubServiceWithHTTPClient(ghRepo, nil, nil, "")
	p := ai.NewUnifiedCodeReviewProvider(ghSvc, nil)
	p.SetAccessChecker(stubAccess{allow: map[uuid.UUID]bool{projectID: false}})

	_, err := p.FetchReview(context.Background(), userID, "https://github.com/acme/app/pull/1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unavailable")
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
	// Without HTTP server FetchMR will fail after authz — check authz alone via deny vs allow path.
	glSvc := gitlabmod.NewGitLabServiceWithHTTPClient(glRepo, nil, nil)
	p := ai.NewUnifiedCodeReviewProvider(nil, glSvc)
	p.SetAccessChecker(stubAccess{allow: map[uuid.UUID]bool{projectID: true}})

	_, err := p.FetchReview(context.Background(), userID, "https://gitlab.example/group/app/-/merge_requests/1")
	// Authz passed; fetch fails because no HTTP — not "unavailable".
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "unavailable")
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
