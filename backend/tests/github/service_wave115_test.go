package github_test
import (
	"github.com/valerii/hub/backend/internal/github"
	"context"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
)

func TestCreateBranch_IssueNotFound(t *testing.T) {
	projectID := uuid.New()
	svc := github.NewGitHubService(&githubMockRepo{issues: map[string]*ent.Issue{}}, nil)
	_, err := svc.CreateBranch(context.Background(), projectID, "NONE-1", github.CreateBranchRequest{
		IntegrationID: uuid.New().String(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "не найдена")
}

func TestCreateBranch_IntegrationNotFound(t *testing.T) {
	projectID := uuid.New()
	svc := github.NewGitHubService(&githubMockRepo{
		issues: map[string]*ent.Issue{
			"HUB-1": {
				ID: uuid.New(), IssueKey: "HUB-1", Title: "Task",
				Edges: ent.IssueEdges{Project: &ent.Project{ID: projectID}},
			},
		},
	}, nil)
	_, err := svc.CreateBranch(context.Background(), projectID, "HUB-1", github.CreateBranchRequest{
		IntegrationID: uuid.New().String(),
	})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "интеграция")
}
