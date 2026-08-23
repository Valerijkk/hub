package gitlab_test

import (
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/valerii/hub/backend/internal/gitlab"
)

var _ gitlab.Repository = (*gitlabMockRepo)(nil)

func testGitLabHandlerApp(t *testing.T, repo gitlab.Repository) *fiber.App {
	t.Helper()

	app := fiber.New()
	api := app.Group("/api")
	handler := gitlab.NewHandler(gitlab.NewGitLabService(repo, nil))

	api.Post("/gitlab/webhook", handler.HandleWebhook)
	api.Post("/projects/:projectId/gitlab/pipelines/trigger", handler.TriggerPipeline)
	api.Post("/projects/:projectId/issues/:issueKey/gitlab/create-branch", handler.CreateBranch)

	return app
}
