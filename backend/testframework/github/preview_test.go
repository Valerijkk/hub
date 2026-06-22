package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/issue"
)

type previewMockRepo struct {
	issues map[string]*ent.Issue
}

func (m *previewMockRepo) CreateIntegration(context.Context, uuid.UUID, github.CreateIntegrationInput) (*ent.GitHubIntegration, error) {
	return nil, nil
}
func (m *previewMockRepo) ListIntegrations(context.Context, uuid.UUID) ([]*ent.GitHubIntegration, error) {
	return nil, nil
}
func (m *previewMockRepo) GetIntegration(context.Context, uuid.UUID) (*ent.GitHubIntegration, error) {
	return nil, nil
}
func (m *previewMockRepo) DeleteIntegration(context.Context, uuid.UUID) error { return nil }
func (m *previewMockRepo) FindIntegrationByRepo(context.Context, string, string) (*ent.GitHubIntegration, error) {
	return nil, nil
}
func (m *previewMockRepo) CreateGitHubLink(context.Context, uuid.UUID, uuid.UUID, github.CreateGitHubLinkInput) (*ent.GitHubLink, error) {
	return nil, nil
}
func (m *previewMockRepo) FindGitHubLink(context.Context, string, string, uuid.UUID) (*ent.GitHubLink, error) {
	return nil, nil
}
func (m *previewMockRepo) UpdateGitHubLinkStatus(context.Context, uuid.UUID, string) error { return nil }
func (m *previewMockRepo) ListGitHubLinks(context.Context, uuid.UUID) ([]*ent.GitHubLink, error) {
	return nil, nil
}
func (m *previewMockRepo) FindIssueByKey(_ context.Context, issueKey string) (*ent.Issue, error) {
	if iss, ok := m.issues[issueKey]; ok {
		return iss, nil
	}
	return nil, &ent.NotFoundError{}
}
func (m *previewMockRepo) UpdateIssueStatus(context.Context, uuid.UUID, uuid.UUID) error { return nil }
func (m *previewMockRepo) FindDoneStatus(context.Context, uuid.UUID) (*ent.WorkflowStatus, error) {
	return nil, nil
}
func (m *previewMockRepo) LogActivity(context.Context, uuid.UUID, string) error { return nil }
func (m *previewMockRepo) ListAllWebhookSecrets(context.Context) ([]string, error) {
	return nil, nil
}
func (m *previewMockRepo) MarkTestCasesOutdatedByFiles(context.Context, []string) ([]string, error) {
	return nil, nil
}

func TestGetLinkPreview_IssueURL(t *testing.T) {
	repo := &previewMockRepo{
		issues: map[string]*ent.Issue{
			"HUB-42": {
				ID:        uuid.New(),
				IssueKey:  "HUB-42",
				Title:     "Login bug",
				IssueType: issue.IssueTypeBug,
				Edges: ent.IssueEdges{
					Status:   &ent.WorkflowStatus{Name: "Open"},
					Assignee: &ent.User{Name: "Alice", AvatarURL: "a.png"},
				},
			},
		},
	}
	svc := github.NewLinkPreviewService(repo, nil)
	ctx := context.Background()

	resp, err := svc.GetLinkPreview(ctx, "https://hub.local/projects/x/issues/HUB-42")
	require.NoError(t, err)
	assert.Equal(t, "issue", resp.Type)
	assert.Equal(t, "Login bug", resp.Title)
	assert.Equal(t, "HUB-42", resp.IssueKey)
	assert.Equal(t, "Open", resp.Status)
	assert.Equal(t, "Alice", resp.AssigneeName)
}

func TestGetLinkPreview_TestCaseURL(t *testing.T) {
	svc := github.NewLinkPreviewService(&previewMockRepo{issues: map[string]*ent.Issue{}}, nil)
	resp, err := svc.GetLinkPreview(context.Background(), "https://hub.local/projects/x/tms/cases/test-cases/HUB-TC-99")
	require.NoError(t, err)
	assert.Equal(t, "test_case", resp.Type)
	assert.Equal(t, "HUB-TC-99", resp.Title)
}

func TestGetLinkPreview_CacheHit(t *testing.T) {
	mr, err := miniredis.Run()
	require.NoError(t, err)
	t.Cleanup(mr.Close)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	t.Cleanup(func() { _ = rdb.Close() })

	rawURL := "https://hub.local/projects/x/tms/cases/test-cases/HUB-TC-1"
	cached := github.LinkPreviewResponse{URL: rawURL, Type: "test_case", Title: "HUB-TC-1"}
	data, _ := json.Marshal(cached)
	require.NoError(t, rdb.Set(context.Background(), "link_preview:"+rawURL, string(data), time.Hour).Err())

	svc := github.NewLinkPreviewService(&previewMockRepo{}, rdb)
	resp, err := svc.GetLinkPreview(context.Background(), rawURL)
	require.NoError(t, err)
	assert.Equal(t, "test_case", resp.Type)
}

func TestGetLinkPreview_FallbackExternal(t *testing.T) {
	svc := github.NewLinkPreviewService(&previewMockRepo{}, nil)
	resp, err := svc.GetLinkPreview(context.Background(), "not-a-valid-url-scheme")
	require.NoError(t, err)
	assert.Equal(t, "external", resp.Type)
	assert.NotEmpty(t, resp.Title)
}
