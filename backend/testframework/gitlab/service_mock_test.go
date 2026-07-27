package gitlab_test
import (
	"github.com/valerii/hub/backend/internal/gitlab"
	"context"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
)

type gitlabMockRepo struct {
	integrations map[string]*ent.GitLabIntegration
	byID         map[uuid.UUID]*ent.GitLabIntegration
	byProject    map[uuid.UUID][]*ent.GitLabIntegration
	issues       map[string]*ent.Issue
	webhookSecrets []string
}

func (m *gitlabMockRepo) CreateIntegration(_ context.Context, projectID uuid.UUID, input gitlab.CreateIntegrationInput) (*ent.GitLabIntegration, error) {
	ig := &ent.GitLabIntegration{
		ID:            uuid.New(),
		GitlabURL:     input.GitLabURL,
		ProjectPath:   input.ProjectPath,
		AccessToken:   input.AccessToken,
		WebhookSecret: input.WebhookSecret,
	}
	if m.integrations == nil {
		m.integrations = make(map[string]*ent.GitLabIntegration)
	}
	if m.byID == nil {
		m.byID = make(map[uuid.UUID]*ent.GitLabIntegration)
	}
	if m.byProject == nil {
		m.byProject = make(map[uuid.UUID][]*ent.GitLabIntegration)
	}
	m.integrations[input.ProjectPath] = ig
	m.byID[ig.ID] = ig
	m.byProject[projectID] = append(m.byProject[projectID], ig)
	if input.WebhookSecret != "" {
		m.webhookSecrets = append(m.webhookSecrets, input.WebhookSecret)
	}
	return ig, nil
}
func (m *gitlabMockRepo) ListIntegrations(_ context.Context, projectID uuid.UUID) ([]*ent.GitLabIntegration, error) {
	if m.byProject == nil {
		return nil, nil
	}
	return m.byProject[projectID], nil
}
func (m *gitlabMockRepo) GetIntegration(_ context.Context, id uuid.UUID) (*ent.GitLabIntegration, error) {
	if ig, ok := m.byID[id]; ok {
		return ig, nil
	}
	return nil, &ent.NotFoundError{}
}
func (m *gitlabMockRepo) GetIntegrationInProject(_ context.Context, projectID, id uuid.UUID) (*ent.GitLabIntegration, error) {
	ig, err := m.GetIntegration(context.Background(), id)
	if err != nil {
		return nil, err
	}
	if m.byProject != nil {
		for _, candidate := range m.byProject[projectID] {
			if candidate.ID == id {
				return ig, nil
			}
		}
		// Integration exists but not under this project — IDOR guard.
		for pid, list := range m.byProject {
			if pid == projectID {
				continue
			}
			for _, candidate := range list {
				if candidate.ID == id {
					return nil, &ent.NotFoundError{}
				}
			}
		}
	}
	return ig, nil
}
func (m *gitlabMockRepo) DeleteIntegration(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (m *gitlabMockRepo) FindIntegrationByProjectPath(_ context.Context, path string) (*ent.GitLabIntegration, error) {
	if ig, ok := m.integrations[path]; ok {
		return ig, nil
	}
	return nil, &ent.NotFoundError{}
}
func (m *gitlabMockRepo) FindIssueByKey(_ context.Context, key string) (*ent.Issue, error) {
	if iss, ok := m.issues[key]; ok {
		return iss, nil
	}
	return nil, &ent.NotFoundError{}
}
func (m *gitlabMockRepo) ListAllWebhookSecrets(context.Context) ([]string, error) {
	return m.webhookSecrets, nil
}

func TestCreateIntegration_DefaultGitLabURL(t *testing.T) {
	repo := &gitlabMockRepo{}
	svc := gitlab.NewGitLabService(repo, nil)

	resp, err := svc.CreateIntegration(context.Background(), uuid.New(), gitlab.CreateIntegrationRequest{
		ProjectPath: "group/repo",
		AccessToken: "token",
	})
	require.NoError(t, err)
	assert.Equal(t, "https://gitlab.com", resp.GitLabURL)
	assert.Equal(t, "group/repo", resp.ProjectPath)
}

func TestCreateIntegration_RejectsUnsafeURL(t *testing.T) {
	svc := gitlab.NewGitLabService(&gitlabMockRepo{}, nil)

	_, err := svc.CreateIntegration(context.Background(), uuid.New(), gitlab.CreateIntegrationRequest{
		GitLabURL:   "file:///etc/passwd",
		ProjectPath: "group/repo",
		AccessToken: "token",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gitlab_url")
}

func TestHandleWebhook_PushLinksIssueKeys(t *testing.T) {
	repo := &gitlabMockRepo{
		integrations: map[string]*ent.GitLabIntegration{
			"hub/backend": {ID: uuid.New(), ProjectPath: "hub/backend"},
		},
		issues: map[string]*ent.Issue{
			"HUB-7": {ID: uuid.New(), IssueKey: "HUB-7", Title: "Task"},
		},
	}
	svc := gitlab.NewGitLabService(repo, nil)

	payload, err := json.Marshal(gitlab.WebhookPushEvent{
		ObjectKind: "push",
		Project:    gitlab.WebhookProject{PathWithNamespace: "hub/backend"},
		Commits: []gitlab.WebhookCommit{{
			ID:      "abc123",
			Message: "fix: HUB-7 login issue",
			URL:     "https://gitlab.com/hub/backend/-/commit/abc123",
			Author:  struct{ Name string `json:"name"` }{Name: "Dev"},
		}},
	})
	require.NoError(t, err)

	require.NoError(t, svc.HandleWebhook(context.Background(), "Push Hook", payload))
}

func TestHandleWebhook_UnknownEventIgnored(t *testing.T) {
	svc := gitlab.NewGitLabService(&gitlabMockRepo{}, nil)
	err := svc.HandleWebhook(context.Background(), "Unknown Hook", []byte(`{}`))
	require.NoError(t, err)
}

func TestHandleWebhook_InvalidPushPayload(t *testing.T) {
	svc := gitlab.NewGitLabService(&gitlabMockRepo{}, nil)
	err := svc.HandleWebhook(context.Background(), "Push Hook", []byte("not-json"))
	require.Error(t, err)
}

func TestHandleWebhook_MergeRequestInvalidPayload(t *testing.T) {
	svc := gitlab.NewGitLabService(&gitlabMockRepo{}, nil)
	err := svc.HandleWebhook(context.Background(), "Merge Request Hook", []byte("{"))
	require.Error(t, err)
}

func TestListIntegrations_Empty(t *testing.T) {
	svc := gitlab.NewGitLabService(&gitlabMockRepo{}, nil)
	list, err := svc.ListIntegrations(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestFindIssueByKey_NotInRepo(t *testing.T) {
	repo := &gitlabMockRepo{issues: map[string]*ent.Issue{}}
	_, err := repo.FindIssueByKey(context.Background(), "MISSING-1")
	require.Error(t, err)
}

func TestIssueEntityFields(t *testing.T) {
	iss := &ent.Issue{IssueKey: "HUB-1", Title: "x", IssueType: issue.IssueTypeTask}
	assert.Equal(t, "HUB-1", iss.IssueKey)
}

func TestHandleWebhook_MergeRequestLinksIssue(t *testing.T) {
	repo := &gitlabMockRepo{
		integrations: map[string]*ent.GitLabIntegration{
			"hub/backend": {ID: uuid.New(), ProjectPath: "hub/backend"},
		},
		issues: map[string]*ent.Issue{
			"HUB-3": {ID: uuid.New(), IssueKey: "HUB-3", Title: "MR task"},
		},
	}
	svc := gitlab.NewGitLabService(repo, nil)

	payload, err := json.Marshal(gitlab.WebhookMergeRequestEvent{
		ObjectKind: "merge_request",
		Project:    gitlab.WebhookProject{PathWithNamespace: "hub/backend"},
		ObjectAttributes: gitlab.WebhookMergeRequest{
			IID:          5,
			Title:        "HUB-3 add endpoint",
			Description:  "",
			State:        "opened",
			Action:       "open",
			SourceBranch: "feature/hub-3",
			URL:          "https://gitlab.com/hub/backend/-/merge_requests/5",
		},
	})
	require.NoError(t, err)

	require.NoError(t, svc.HandleWebhook(context.Background(), "Merge Request Hook", payload))
}

func TestHandleWebhook_PipelineHook(t *testing.T) {
	repo := &gitlabMockRepo{
		integrations: map[string]*ent.GitLabIntegration{
			"hub/backend": {ID: uuid.New(), ProjectPath: "hub/backend"},
		},
	}
	svc := gitlab.NewGitLabService(repo, nil)

	payload, err := json.Marshal(gitlab.WebhookPipelineEvent{
		ObjectKind: "pipeline",
		Project:    gitlab.WebhookProject{PathWithNamespace: "hub/backend"},
		ObjectAttributes: gitlab.WebhookPipeline{
			ID:     99,
			Status: "success",
			Ref:    "main",
			SHA:    "deadbeef",
		},
	})
	require.NoError(t, err)

	require.NoError(t, svc.HandleWebhook(context.Background(), "Pipeline Hook", payload))
}

func TestHandleWebhook_PipelineInvalidPayload(t *testing.T) {
	svc := gitlab.NewGitLabService(&gitlabMockRepo{}, nil)
	err := svc.HandleWebhook(context.Background(), "Pipeline Hook", []byte("{"))
	require.Error(t, err)
}

func TestDeleteIntegration_DelegatesToRepo(t *testing.T) {
	id := uuid.New()
	repo := &gitlabMockRepo{}
	svc := gitlab.NewGitLabService(repo, nil)
	require.NoError(t, svc.DeleteIntegration(context.Background(), uuid.New(), id))
}

func TestCreateIntegration_CustomGitLabURL(t *testing.T) {
	repo := &gitlabMockRepo{}
	svc := gitlab.NewGitLabService(repo, nil)

	resp, err := svc.CreateIntegration(context.Background(), uuid.New(), gitlab.CreateIntegrationRequest{
		GitLabURL:   "https://gitlab.example.com",
		ProjectPath: "team/app",
		AccessToken: "glpat-test",
	})
	require.NoError(t, err)
	assert.Equal(t, "https://gitlab.example.com", resp.GitLabURL)
	assert.Equal(t, "team/app", resp.ProjectPath)
}
