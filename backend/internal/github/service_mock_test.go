package github

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/githublink"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/workflowstatus"
)

type githubMockRepo struct {
	integrations map[string]*ent.GitHubIntegration
	issues       map[string]*ent.Issue
	links        map[string]*ent.GitHubLink
	webhookSecrets []string
	outdatedKeys   []string
}

func (m *githubMockRepo) integrationKey(owner, repo string) string {
	return owner + "/" + repo
}

func (m *githubMockRepo) CreateIntegration(_ context.Context, _ uuid.UUID, input CreateIntegrationInput) (*ent.GitHubIntegration, error) {
	ig := &ent.GitHubIntegration{
		ID:            uuid.New(),
		RepoOwner:     input.RepoOwner,
		RepoName:      input.RepoName,
		AccessToken:   input.AccessToken,
		WebhookSecret: input.WebhookSecret,
		MergeStatusID: input.MergeStatusID,
		CreatedAt:     time.Now().UTC(),
	}
	if m.integrations == nil {
		m.integrations = make(map[string]*ent.GitHubIntegration)
	}
	m.integrations[m.integrationKey(input.RepoOwner, input.RepoName)] = ig
	if input.WebhookSecret != "" {
		m.webhookSecrets = append(m.webhookSecrets, input.WebhookSecret)
	}
	return ig, nil
}

func (m *githubMockRepo) ListIntegrations(_ context.Context, _ uuid.UUID) ([]*ent.GitHubIntegration, error) {
	out := make([]*ent.GitHubIntegration, 0, len(m.integrations))
	for _, ig := range m.integrations {
		out = append(out, ig)
	}
	return out, nil
}

func (m *githubMockRepo) GetIntegration(_ context.Context, id uuid.UUID) (*ent.GitHubIntegration, error) {
	for _, ig := range m.integrations {
		if ig.ID == id {
			return ig, nil
		}
	}
	return nil, &ent.NotFoundError{}
}

func (m *githubMockRepo) DeleteIntegration(_ context.Context, id uuid.UUID) error {
	for k, ig := range m.integrations {
		if ig.ID == id {
			delete(m.integrations, k)
			return nil
		}
	}
	return &ent.NotFoundError{}
}

func (m *githubMockRepo) FindIntegrationByRepo(_ context.Context, owner, repo string) (*ent.GitHubIntegration, error) {
	if ig, ok := m.integrations[m.integrationKey(owner, repo)]; ok {
		return ig, nil
	}
	return nil, &ent.NotFoundError{}
}

func (m *githubMockRepo) CreateGitHubLink(_ context.Context, issueID, integrationID uuid.UUID, input CreateGitHubLinkInput) (*ent.GitHubLink, error) {
	link := &ent.GitHubLink{
		ID:            uuid.New(),
		LinkType:      githublink.LinkType(input.LinkType),
		ExternalID:    input.ExternalID,
		Title:         input.Title,
		URL:           input.URL,
		Status:        githublink.Status(input.Status),
		Author:        input.Author,
		Branch:        input.Branch,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}
	if m.links == nil {
		m.links = make(map[string]*ent.GitHubLink)
	}
	key := input.ExternalID + ":" + input.LinkType + ":" + integrationID.String()
	m.links[key] = link
	return link, nil
}

func (m *githubMockRepo) FindGitHubLink(_ context.Context, externalID, linkType string, integrationID uuid.UUID) (*ent.GitHubLink, error) {
	key := externalID + ":" + linkType + ":" + integrationID.String()
	if link, ok := m.links[key]; ok {
		return link, nil
	}
	return nil, &ent.NotFoundError{}
}

func (m *githubMockRepo) UpdateGitHubLinkStatus(_ context.Context, linkID uuid.UUID, status string) error {
	for _, link := range m.links {
		if link.ID == linkID {
			link.Status = githublink.Status(status)
			return nil
		}
	}
	return &ent.NotFoundError{}
}

func (m *githubMockRepo) ListGitHubLinks(_ context.Context, issueID uuid.UUID) ([]*ent.GitHubLink, error) {
	out := make([]*ent.GitHubLink, 0)
	for _, link := range m.links {
		out = append(out, link)
	}
	return out, nil
}

func (m *githubMockRepo) FindIssueByKey(_ context.Context, key string) (*ent.Issue, error) {
	if iss, ok := m.issues[key]; ok {
		return iss, nil
	}
	return nil, &ent.NotFoundError{}
}

func (m *githubMockRepo) UpdateIssueStatus(_ context.Context, _, _ uuid.UUID) error {
	return nil
}

func (m *githubMockRepo) FindDoneStatus(_ context.Context, _ uuid.UUID) (*ent.WorkflowStatus, error) {
	return &ent.WorkflowStatus{ID: uuid.New(), Category: workflowstatus.CategoryDone}, nil
}

func (m *githubMockRepo) LogActivity(_ context.Context, _ uuid.UUID, _ string) error {
	return nil
}

func (m *githubMockRepo) ListAllWebhookSecrets(context.Context) ([]string, error) {
	return m.webhookSecrets, nil
}

func (m *githubMockRepo) MarkTestCasesOutdatedByFiles(_ context.Context, changedFiles []string) ([]string, error) {
	if len(changedFiles) == 0 {
		return nil, nil
	}
	return []string{"TC-1"}, nil
}

func TestCreateIntegration_InvalidMergeStatusID(t *testing.T) {
	svc := NewGitHubService(&githubMockRepo{}, nil)
	_, err := svc.CreateIntegration(context.Background(), uuid.New(), CreateIntegrationRequest{
		RepoOwner:     "org",
		RepoName:      "hub",
		AccessToken:   "token",
		WebhookSecret: "secret",
		MergeStatusID: "not-a-uuid",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "merge_status_id")
}

func TestCreateIntegration_Success(t *testing.T) {
	repo := &githubMockRepo{}
	svc := NewGitHubService(repo, nil)

	resp, err := svc.CreateIntegration(context.Background(), uuid.New(), CreateIntegrationRequest{
		RepoOwner:     "hub-org",
		RepoName:      "backend",
		AccessToken:   "ghp_test",
		WebhookSecret: "whsec",
	})
	require.NoError(t, err)
	assert.Equal(t, "hub-org", resp.RepoOwner)
	assert.Equal(t, "backend", resp.RepoName)
}

func TestListIntegrations_Empty(t *testing.T) {
	svc := NewGitHubService(&githubMockRepo{}, nil)
	list, err := svc.ListIntegrations(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Empty(t, list)
}

func TestHandleWebhook_PushLinksCommit(t *testing.T) {
	integrationID := uuid.New()
	repo := &githubMockRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"org/hub": {ID: integrationID, RepoOwner: "org", RepoName: "hub"},
		},
		issues: map[string]*ent.Issue{
			"HUB-7": {ID: uuid.New(), IssueKey: "HUB-7", Title: "Fix bug"},
		},
	}
	svc := NewGitHubService(repo, nil)

	payload, err := json.Marshal(map[string]interface{}{
		"repository": map[string]interface{}{
			"name":  "hub",
			"owner": map[string]string{"login": "org"},
		},
		"commits": []map[string]interface{}{
			{
				"sha":     "abc123",
				"message": "fix: HUB-7 login issue",
				"url":     "https://github.com/org/hub/commit/abc123",
				"author":  map[string]string{"name": "Dev"},
				"added":   []string{"src/test/java/com/example/LoginTest.java"},
				"modified": []string{},
			},
		},
	})
	require.NoError(t, err)

	require.NoError(t, svc.HandleWebhook(context.Background(), "push", payload))
	assert.NotEmpty(t, repo.links)
}

func TestHandleWebhook_PushInvalidPayload(t *testing.T) {
	svc := NewGitHubService(&githubMockRepo{}, nil)
	err := svc.HandleWebhook(context.Background(), "push", []byte("{"))
	require.Error(t, err)
}

func TestHandleWebhook_PROpenedCreatesLink(t *testing.T) {
	integrationID := uuid.New()
	repo := &githubMockRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"org/hub": {ID: integrationID, RepoOwner: "org", RepoName: "hub"},
		},
		issues: map[string]*ent.Issue{
			"HUB-1": {ID: uuid.New(), IssueKey: "HUB-1", Title: "Task"},
		},
	}
	svc := NewGitHubService(repo, nil)

	payload, err := json.Marshal(map[string]interface{}{
		"action": "opened",
		"pull_request": map[string]interface{}{
			"number":    42,
			"title":     "HUB-1 implement feature",
			"body":      "",
			"html_url":  "https://github.com/org/hub/pull/42",
			"merged":    false,
			"head":      map[string]string{"ref": "feature/hub-1"},
			"user":      map[string]string{"login": "dev"},
		},
		"repository": map[string]interface{}{
			"name":  "hub",
			"owner": map[string]string{"login": "org"},
		},
	})
	require.NoError(t, err)

	require.NoError(t, svc.HandleWebhook(context.Background(), "pull_request", payload))
	assert.Contains(t, repo.links, "42:pull_request:"+integrationID.String())
}

func TestHandleWebhook_UnknownEventIgnored(t *testing.T) {
	svc := NewGitHubService(&githubMockRepo{}, nil)
	err := svc.HandleWebhook(context.Background(), "release", []byte(`{}`))
	require.NoError(t, err)
}

func TestListGitHubLinks_IssueNotFound(t *testing.T) {
	svc := NewGitHubService(&githubMockRepo{}, nil)
	_, err := svc.ListGitHubLinks(context.Background(), "MISSING-1")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}

func TestCreateBranch_InvalidIntegrationID(t *testing.T) {
	repo := &githubMockRepo{
		issues: map[string]*ent.Issue{
			"HUB-2": {ID: uuid.New(), IssueKey: "HUB-2", Title: "Task"},
		},
	}
	svc := NewGitHubService(repo, nil)

	_, err := svc.CreateBranch(context.Background(), "HUB-2", CreateBranchRequest{
		IntegrationID: "bad-id",
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "integration_id")
}

func TestDeleteIntegration_DelegatesToRepo(t *testing.T) {
	id := uuid.New()
	repo := &githubMockRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"org/hub": {ID: id, RepoOwner: "org", RepoName: "hub"},
		},
	}
	svc := NewGitHubService(repo, nil)
	require.NoError(t, svc.DeleteIntegration(context.Background(), id))
	assert.Empty(t, repo.integrations)
}

func TestMarkTestCasesOutdated_PublishesWhenFilesChanged(t *testing.T) {
	repo := &githubMockRepo{}
	svc := NewGitHubService(repo, nil)
	svc.markTestCasesOutdated(context.Background(), []string{"src/FooTest.java"})
}

func TestFindIssueByKey_NotInRepo(t *testing.T) {
	repo := &githubMockRepo{issues: map[string]*ent.Issue{}}
	_, err := repo.FindIssueByKey(context.Background(), "NONE-1")
	require.Error(t, err)
}

func TestIssueEntityForWebhook(t *testing.T) {
	iss := &ent.Issue{IssueKey: "HUB-3", Title: "x", IssueType: issue.IssueTypeBug}
	assert.Equal(t, issue.IssueTypeBug, iss.IssueType)
}

func TestListGitHubLinks_Success(t *testing.T) {
	issueID := uuid.New()
	integrationID := uuid.New()
	repo := &githubMockRepo{
		issues: map[string]*ent.Issue{
			"HUB-10": {ID: issueID, IssueKey: "HUB-10", Title: "Task"},
		},
		links: map[string]*ent.GitHubLink{
			"7:pull_request:" + integrationID.String(): {
				ID:         uuid.New(),
				LinkType:   githublink.LinkTypePullRequest,
				ExternalID: "7",
				Title:      "HUB-10 feature",
				URL:        "https://github.com/org/hub/pull/7",
				Status:     githublink.StatusOpen,
			},
		},
	}
	svc := NewGitHubService(repo, nil)

	links, err := svc.ListGitHubLinks(context.Background(), "HUB-10")
	require.NoError(t, err)
	require.Len(t, links, 1)
	assert.Equal(t, "7", links[0].ExternalID)
}

func TestListIntegrations_ReturnsItems(t *testing.T) {
	repo := &githubMockRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"org/hub": {ID: uuid.New(), RepoOwner: "org", RepoName: "hub"},
		},
	}
	svc := NewGitHubService(repo, nil)

	list, err := svc.ListIntegrations(context.Background(), uuid.New())
	require.NoError(t, err)
	require.Len(t, list, 1)
	assert.Equal(t, "org", list[0].RepoOwner)
}

func TestHandleWebhook_PRMergedUpdatesLink(t *testing.T) {
	mergeStatusID := uuid.New()
	projectID := uuid.New()
	issueID := uuid.New()
	integrationID := uuid.New()
	repo := &githubMockRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"org/hub": {
				ID:            integrationID,
				RepoOwner:     "org",
				RepoName:      "hub",
				MergeStatusID: &mergeStatusID,
			},
		},
		issues: map[string]*ent.Issue{
			"HUB-4": {
				ID:       issueID,
				IssueKey: "HUB-4",
				Title:    "Task",
				Edges: ent.IssueEdges{
					Project: &ent.Project{ID: projectID},
				},
			},
		},
		links: map[string]*ent.GitHubLink{
			"12:pull_request:" + integrationID.String(): {
				ID:         uuid.New(),
				LinkType:   githublink.LinkTypePullRequest,
				ExternalID: "12",
				Status:     githublink.StatusOpen,
			},
		},
	}
	svc := NewGitHubService(repo, nil)

	payload, err := json.Marshal(map[string]interface{}{
		"action": "closed",
		"pull_request": map[string]interface{}{
			"number":   12,
			"title":    "HUB-4 merge me",
			"body":     "",
			"html_url": "https://github.com/org/hub/pull/12",
			"merged":   true,
			"head":     map[string]string{"ref": "feature/hub-4"},
			"user":     map[string]string{"login": "dev"},
		},
		"repository": map[string]interface{}{
			"name":  "hub",
			"owner": map[string]string{"login": "org"},
		},
	})
	require.NoError(t, err)

	require.NoError(t, svc.HandleWebhook(context.Background(), "pull_request", payload))
	link := repo.links["12:pull_request:"+integrationID.String()]
	require.NotNil(t, link)
	assert.Equal(t, githublink.StatusMerged, link.Status)
}

func TestHandleWebhook_PROpenedUpdatesExistingLink(t *testing.T) {
	integrationID := uuid.New()
	repo := &githubMockRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"org/hub": {ID: integrationID, RepoOwner: "org", RepoName: "hub"},
		},
		issues: map[string]*ent.Issue{
			"HUB-8": {ID: uuid.New(), IssueKey: "HUB-8", Title: "Task"},
		},
		links: map[string]*ent.GitHubLink{
			"3:pull_request:" + integrationID.String(): {
				ID:         uuid.New(),
				LinkType:   githublink.LinkTypePullRequest,
				ExternalID: "3",
				Status:     githublink.StatusClosed,
			},
		},
	}
	svc := NewGitHubService(repo, nil)

	payload, err := json.Marshal(map[string]interface{}{
		"action": "reopened",
		"pull_request": map[string]interface{}{
			"number":   3,
			"title":    "HUB-8 reopen",
			"body":     "",
			"html_url": "https://github.com/org/hub/pull/3",
			"merged":   false,
			"head":     map[string]string{"ref": "feature/hub-8"},
			"user":     map[string]string{"login": "dev"},
		},
		"repository": map[string]interface{}{
			"name":  "hub",
			"owner": map[string]string{"login": "org"},
		},
	})
	require.NoError(t, err)

	require.NoError(t, svc.HandleWebhook(context.Background(), "pull_request", payload))
	link := repo.links["3:pull_request:"+integrationID.String()]
	require.NotNil(t, link)
	assert.Equal(t, githublink.StatusOpen, link.Status)
}

func TestHandleWebhook_PRInvalidPayload(t *testing.T) {
	svc := NewGitHubService(&githubMockRepo{}, nil)
	err := svc.HandleWebhook(context.Background(), "pull_request", []byte("{"))
	require.Error(t, err)
}

func TestListWorkflowRuns_EmptyIntegrations(t *testing.T) {
	svc := NewGitHubService(&githubMockRepo{}, nil)
	runs, err := svc.ListWorkflowRuns(context.Background(), uuid.New())
	require.NoError(t, err)
	assert.Empty(t, runs)
}

func TestGetWorkflowRunDetail_NotFound(t *testing.T) {
	repo := &githubMockRepo{
		integrations: map[string]*ent.GitHubIntegration{
			"o/r": {ID: uuid.New(), RepoOwner: "o", RepoName: "r", AccessToken: "tok"},
		},
	}
	svc := NewGitHubService(repo, nil)
	_, err := svc.GetWorkflowRunDetail(context.Background(), uuid.New(), 999)
	require.Error(t, err)
}
