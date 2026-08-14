package github_test

import (
	"context"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/valerii/hub/backend/ent"
	hubgithub "github.com/valerii/hub/backend/internal/github"
)

type githubMockRepo struct {
	hubgithub.Repository
	integrations   map[string]*ent.GitHubIntegration
	issues         map[string]*ent.Issue
	webhookSecrets []string
}

func (m *githubMockRepo) FindIntegrationByRepo(_ context.Context, owner, repo string) (*ent.GitHubIntegration, error) {
	integration := m.integrations[owner+"/"+repo]
	if integration == nil {
		return nil, &ent.NotFoundError{}
	}
	return integration, nil
}

func (m *githubMockRepo) FindIssueByKeyInProject(_ context.Context, projectID uuid.UUID, issueKey string) (*ent.Issue, error) {
	issue := m.issues[issueKey]
	if issue == nil || issue.Edges.Project == nil || issue.Edges.Project.ID != projectID {
		return nil, &ent.NotFoundError{}
	}
	return issue, nil
}

func (m *githubMockRepo) GetIntegrationInProject(_ context.Context, _ uuid.UUID, integrationID uuid.UUID) (*ent.GitHubIntegration, error) {
	for _, integration := range m.integrations {
		if integration.ID == integrationID {
			return integration, nil
		}
	}
	return nil, &ent.NotFoundError{}
}

func testGitHubHandlerApp(t *testing.T, repo hubgithub.Repository) *fiber.App {
	t.Helper()
	app := fiber.New()
	handler := hubgithub.NewHandler(hubgithub.NewGitHubService(repo, nil), nil)
	app.Post("/github/webhook", handler.HandleWebhook)
	return app
}

var _ hubgithub.Repository = (*githubMockRepo)(nil)
