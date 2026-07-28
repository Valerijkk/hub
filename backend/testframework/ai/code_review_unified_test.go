package ai_test
import (
	"github.com/valerii/hub/backend/internal/ai"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	githubmod "github.com/valerii/hub/backend/internal/github"
	gitlabmod "github.com/valerii/hub/backend/internal/gitlab"
	messengermod "github.com/valerii/hub/backend/internal/messenger"
)

func TestUnifiedCodeReviewProvider_FetchReviewGitLab(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/changes") {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"changes": []map[string]interface{}{{"new_path": "app.go", "diff": "+fix"}},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"web_url":       "https://gitlab.example/group/app/-/merge_requests/7",
			"title":         "AI review MR",
			"source_branch": "fix",
			"target_branch": "main",
			"author":        map[string]interface{}{"username": "qa"},
		})
	}))
	t.Cleanup(srv.Close)

	projectID := uuid.New()
	repo := &gitlabReviewRepo{
		integrations: map[string]*ent.GitLabIntegration{
			"group/app": {
				GitlabURL: srv.URL, ProjectPath: "group/app", AccessToken: "tok",
				Edges: ent.GitLabIntegrationEdges{Project: &ent.Project{ID: projectID}},
			},
		},
	}
	glSvc := gitlabmod.NewGitLabServiceWithHTTPClient(repo, nil, srv.Client())
	p := ai.NewUnifiedCodeReviewProvider(nil, glSvc)
	p.SetAccessChecker(stubAccess{allow: map[uuid.UUID]bool{projectID: true}})

	rc, err := p.FetchReview(context.Background(), uuid.New(), srv.URL+"/group/app/-/merge_requests/7")
	require.NoError(t, err)
	require.NotNil(t, rc)
	assert.Equal(t, "AI review MR", rc.Title)
	assert.Equal(t, "fix", rc.HeadBranch)
	assert.Contains(t, rc.ChangedFiles, "app.go")
}

func TestUnifiedCodeReviewProvider_FetchReviewGitHub(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch {
		case strings.HasSuffix(r.URL.Path, "/files"):
			_, _ = w.Write([]byte(`[{"filename":"main.go","status":"added","patch":"+package main"}]`))
		case strings.Contains(r.URL.Path, "/pulls/"):
			_, _ = w.Write([]byte(`{
				"html_url":"https://github.com/acme/app/pull/9",
				"title":"Add feature",
				"user":{"login":"bot"},
				"base":{"ref":"main"},
				"head":{"ref":"feat"}
			}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	projectID := uuid.New()
	ghRepo := &githubReviewRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"acme/app": {
				RepoOwner: "acme", RepoName: "app", AccessToken: "ghp_x",
				Edges: ent.GitHubIntegrationEdges{Project: &ent.Project{ID: projectID}},
			},
		},
	}
	ghSvc := githubmod.NewGitHubServiceWithHTTPClient(ghRepo, nil, srv.Client(), srv.URL)
	p := ai.NewUnifiedCodeReviewProvider(ghSvc, nil)
	p.SetAccessChecker(stubAccess{allow: map[uuid.UUID]bool{projectID: true}})

	rc, err := p.FetchReview(context.Background(), uuid.New(), "https://github.com/acme/app/pull/9")
	require.NoError(t, err)
	require.NotNil(t, rc)
	assert.Equal(t, "Add feature", rc.Title)
	assert.Contains(t, rc.ChangedFiles, "main.go")
}

func TestMessengerBot_HandleCommand_ReviewUnifiedGitLab(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/changes") {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{
				"changes": []map[string]interface{}{{"new_path": "x.go", "diff": "+x"}},
			})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{
			"title": "Bot review MR", "source_branch": "b", "target_branch": "main",
		})
	}))
	t.Cleanup(srv.Close)

	projectID := uuid.New()
	repo := &gitlabReviewRepo{
		integrations: map[string]*ent.GitLabIntegration{
			"o/r": {
				GitlabURL: srv.URL, ProjectPath: "o/r", AccessToken: "k",
				Edges: ent.GitLabIntegrationEdges{Project: &ent.Project{ID: projectID}},
			},
		},
	}
	glSvc := gitlabmod.NewGitLabServiceWithHTTPClient(repo, nil, srv.Client())
	adapter := &streamingMockAdapter{chunks: []string{"review done"}}
	bot, _, userID, channelID, _ := setupHubAIBot(t, adapter)
	review := ai.NewUnifiedCodeReviewProvider(nil, glSvc)
	review.SetAccessChecker(stubAccess{allow: map[uuid.UUID]bool{projectID: true}})
	bot.SetCodeReview(review)

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "/ai review " + srv.URL + "/o/r/-/merge_requests/3",
		MessageID:   "m-unified-review",
	})
	require.NoError(t, err)
}

func TestMessengerBot_HandleCommand_ReviewUnifiedGitHub(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if strings.HasSuffix(r.URL.Path, "/files") {
			_, _ = w.Write([]byte(`[{"filename":"x.go","status":"modified","patch":"+x"}]`))
			return
		}
		_, _ = w.Write([]byte(`{"title":"Bot PR","user":{"login":"dev"},"base":{"ref":"main"},"head":{"ref":"fix"}}`))
	}))
	t.Cleanup(srv.Close)

	projectID := uuid.New()
	ghRepo := &githubReviewRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"o/r": {
				RepoOwner: "o", RepoName: "r", AccessToken: "tok",
				Edges: ent.GitHubIntegrationEdges{Project: &ent.Project{ID: projectID}},
			},
		},
	}
	ghSvc := githubmod.NewGitHubServiceWithHTTPClient(ghRepo, nil, srv.Client(), srv.URL)
	adapter := &streamingMockAdapter{chunks: []string{"github review ok"}}
	bot, _, userID, channelID, _ := setupHubAIBot(t, adapter)
	review := ai.NewUnifiedCodeReviewProvider(ghSvc, nil)
	review.SetAccessChecker(stubAccess{allow: map[uuid.UUID]bool{projectID: true}})
	bot.SetCodeReview(review)

	err := bot.OnMessageSent(context.Background(), messengermod.MessageHookEvent{
		ChannelID:   channelID,
		SenderID:    userID,
		ContentText: "/ai review https://github.com/o/r/pull/5",
		MessageID:   "m-gh-review",
	})
	require.NoError(t, err)
}

type gitlabReviewRepo struct {
	integrations map[string]*ent.GitLabIntegration
}

func (m *gitlabReviewRepo) CreateIntegration(context.Context, uuid.UUID, gitlabmod.CreateIntegrationInput) (*ent.GitLabIntegration, error) {
	return nil, nil
}
func (m *gitlabReviewRepo) ListIntegrations(context.Context, uuid.UUID) ([]*ent.GitLabIntegration, error) {
	return nil, nil
}
func (m *gitlabReviewRepo) GetIntegration(context.Context, uuid.UUID) (*ent.GitLabIntegration, error) {
	return nil, nil
}
func (m *gitlabReviewRepo) GetIntegrationInProject(context.Context, uuid.UUID, uuid.UUID) (*ent.GitLabIntegration, error) {
	return nil, nil
}
func (m *gitlabReviewRepo) DeleteIntegration(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (m *gitlabReviewRepo) FindIntegrationByProjectPath(_ context.Context, path string) (*ent.GitLabIntegration, error) {
	if ig, ok := m.integrations[path]; ok {
		return ig, nil
	}
	return nil, &ent.NotFoundError{}
}
func (m *gitlabReviewRepo) FindIssueByKey(context.Context, string) (*ent.Issue, error) {
	return nil, nil
}
func (m *gitlabReviewRepo) FindIssueByKeyInProject(context.Context, uuid.UUID, string) (*ent.Issue, error) {
	return nil, &ent.NotFoundError{}
}
func (m *gitlabReviewRepo) ListAllWebhookSecrets(context.Context) ([]string, error) {
	return nil, nil
}

type githubReviewRepo struct {
	integrations map[string]*ent.GitHubIntegration
}

func (m *githubReviewRepo) CreateIntegration(_ context.Context, _ uuid.UUID, input githubmod.CreateIntegrationInput) (*ent.GitHubIntegration, error) {
	return nil, nil
}
func (m *githubReviewRepo) ListIntegrations(context.Context, uuid.UUID) ([]*ent.GitHubIntegration, error) {
	return nil, nil
}
func (m *githubReviewRepo) GetIntegration(context.Context, uuid.UUID) (*ent.GitHubIntegration, error) {
	return nil, nil
}
func (m *githubReviewRepo) GetIntegrationInProject(context.Context, uuid.UUID, uuid.UUID) (*ent.GitHubIntegration, error) {
	return nil, nil
}
func (m *githubReviewRepo) DeleteIntegration(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (m *githubReviewRepo) FindIntegrationByRepo(_ context.Context, owner, repo string) (*ent.GitHubIntegration, error) {
	if ig, ok := m.integrations[owner+"/"+repo]; ok {
		return ig, nil
	}
	return nil, &ent.NotFoundError{}
}
func (m *githubReviewRepo) CreateGitHubLink(context.Context, uuid.UUID, uuid.UUID, githubmod.CreateGitHubLinkInput) (*ent.GitHubLink, error) {
	return nil, nil
}
func (m *githubReviewRepo) FindGitHubLink(context.Context, string, string, uuid.UUID) (*ent.GitHubLink, error) {
	return nil, nil
}
func (m *githubReviewRepo) UpdateGitHubLinkStatus(context.Context, uuid.UUID, string) error { return nil }
func (m *githubReviewRepo) ListGitHubLinks(context.Context, uuid.UUID) ([]*ent.GitHubLink, error) {
	return nil, nil
}
func (m *githubReviewRepo) FindIssueByKey(context.Context, string) (*ent.Issue, error) { return nil, nil }
func (m *githubReviewRepo) FindIssueByKeyInProject(context.Context, uuid.UUID, string) (*ent.Issue, error) {
	return nil, &ent.NotFoundError{}
}
func (m *githubReviewRepo) UpdateIssueStatus(context.Context, uuid.UUID, uuid.UUID) error {
	return nil
}
func (m *githubReviewRepo) FindDoneStatus(context.Context, uuid.UUID) (*ent.WorkflowStatus, error) {
	return nil, nil
}
func (m *githubReviewRepo) LogActivity(context.Context, uuid.UUID, string) error { return nil }
func (m *githubReviewRepo) ListAllWebhookSecrets(context.Context) ([]string, error) {
	return nil, nil
}
func (m *githubReviewRepo) IsProjectMember(context.Context, uuid.UUID, uuid.UUID) (bool, error) {
	return true, nil
}
func (m *githubReviewRepo) MarkTestCasesOutdatedByFiles(context.Context, []string) ([]string, error) {
	return nil, nil
}
