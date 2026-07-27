package tracker_test

// Тесты репозитория трекера задач (волна 100): интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/automationrule"
	"github.com/valerii/hub/backend/ent/enttest"
)

type stubOutboxPublisher struct {
	events []string
}

func (s *stubOutboxPublisher) PublishInTx(_ context.Context, _ *ent.Tx, eventType string, _ any) error {
	s.events = append(s.events, eventType)
	return nil
}

func (s *stubOutboxPublisher) Publish(_ context.Context, _ *ent.Client, eventType string, _ any) error {
	s.events = append(s.events, eventType)
	return nil
}

// TestEntRepository_UpdateIssue_PublishesOutboxEvents проверяет слой репозитория: ent репозиторий обновление issue publishes outbox events.
func TestEntRepository_UpdateIssue_PublishesOutboxEvents(t *testing.T) {
	env := newTrackerRepoTestEnv(t)
	pub := &stubOutboxPublisher{}
	env.repo = tracker.NewEntRepository(env.client, pub)

	assignee, err := env.client.User.Create().
		SetEmail("assignee-" + uuid.New().String()[:8] + "@test.com").
		Save(env.ctx)
	require.NoError(t, err)

	created, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Outbox issue", IssueType: "task", ReporterID: env.userID,
	})
	require.NoError(t, err)

	newTitle := "Updated with assignee"
	_, err = env.repo.UpdateIssue(env.ctx, created.ID, tracker.UpdateIssueInput{
		Title:      &newTitle,
		AssigneeID: &assignee.ID,
		Version:    created.Version,
	})
	require.NoError(t, err)
	assert.Contains(t, pub.events, "project."+env.projectID.String()+".issue.updated")
	assert.Contains(t, pub.events, "project."+env.projectID.String()+".issue.assigned")
}

// TestToAutomationRuleResponse_NilAndEmptyConfig проверяет сценарий: to автоматизация rule response nil and пустой config.
func TestToAutomationRuleResponse_NilAndEmptyConfig(t *testing.T) {
	assert.Empty(t, tracker.ExportToAutomationRuleResponse(nil).ID)

	client := enttest.Open(t, "sqlite3", "file:auto_dto_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()
	proj, err := client.Project.Create().SetName("DTO").SetKey("DTO").Save(ctx)
	require.NoError(t, err)
	user, err := client.User.Create().SetEmail("dto@test.com").Save(ctx)
	require.NoError(t, err)
	rule, err := client.AutomationRule.Create().
		SetName("empty cfg").
		SetProjectID(proj.ID).
		SetCreatedByID(user.ID).
		SetTriggerType(automationrule.TriggerTypeIssueCreated).
		SetActionType(automationrule.ActionTypeNotifyUser).
		Save(ctx)
	require.NoError(t, err)

	resp := tracker.ExportToAutomationRuleResponse(rule)
	assert.NotNil(t, resp.TriggerConfig)
	assert.NotNil(t, resp.ActionConfig)
}

// TestHandler_Automation_NotConfigured проверяет HTTP-слой: HTTP-обработчик автоматизация not configured.
func TestHandler_Automation_NotConfigured(t *testing.T) {
	svc, _ := newTestService()
	h := tracker.NewHandler(svc, nil)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", uuid.New().String())
		return c.Next()
	}
	pg := api.Group("/projects/:projectId", authMw)
	pg.Get("/automation-rules", h.ListAutomationRules)
	pg.Post("/automation-rules", h.CreateAutomationRule)
	pg.Patch("/automation-rules/:ruleId", h.UpdateAutomationRule)
	pg.Delete("/automation-rules/:ruleId", h.DeleteAutomationRule)

	base := "/api/projects/" + uuid.New().String()

	req := httptest.NewRequest(http.MethodGet, base+"/automation-rules", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, base+"/automation-rules", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPatch, base+"/automation-rules/"+uuid.New().String(), bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, base+"/automation-rules/"+uuid.New().String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotImplemented, resp.StatusCode)
}

// TestHandler_Automation_InvalidRuleID проверяет отклонение при HTTP-обработчик автоматизация некорректный rule id.
func TestHandler_Automation_InvalidRuleID(t *testing.T) {
	client := enttestOpenAutomationHandler(t)
	base := "/api/projects/" + client.projectID.String()

	req := httptest.NewRequest(http.MethodPatch, base+"/automation-rules/not-uuid", bytes.NewReader([]byte(`{"name":"x"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func enttestOpenAutomationHandler(t *testing.T) struct {
	app       *fiber.App
	projectID uuid.UUID
} {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:auto_handler_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()
	user, err := client.User.Create().SetEmail("auto-h@test.com").Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().SetName("tracker.Handler Proj").SetKey("HND").Save(ctx)
	require.NoError(t, err)

	repo := tracker.NewEntRepository(client, nil)
	h := tracker.NewHandler(tracker.NewTrackerService(repo, nil), tracker.NewAutomationService(repo))
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", user.ID.String())
		return c.Next()
	}
	pg := api.Group("/projects/:projectId", authMw)
	pg.Patch("/automation-rules/:ruleId", h.UpdateAutomationRule)
	pg.Delete("/automation-rules/:ruleId", h.DeleteAutomationRule)

	return struct {
		app       *fiber.App
		projectID uuid.UUID
	}{app: app, projectID: proj.ID}
}

// TestEntRepository_GetBoardData_AssigneeAndLabelFilters проверяет слой репозитория: ent репозиторий получение board data assignee and label filters.
func TestEntRepository_GetBoardData_AssigneeAndLabelFilters(t *testing.T) {
	env := newTrackerRepoTestEnv(t)

	assignee, err := env.client.User.Create().
		SetEmail("board-assignee@test.com").
		SetName("Assignee").
		Save(env.ctx)
	require.NoError(t, err)

	lbl, err := env.client.Label.Create().
		SetName("critical").
		SetColor("#FF0000").
		SetProjectID(env.projectID).
		Save(env.ctx)
	require.NoError(t, err)

	matching, err := env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Match", IssueType: "bug", Priority: "high",
		ReporterID: env.userID, AssigneeID: &assignee.ID,
	})
	require.NoError(t, err)
	_, err = env.client.Issue.UpdateOneID(matching.ID).AddLabels(lbl).Save(env.ctx)
	require.NoError(t, err)

	_, err = env.repo.CreateIssue(env.ctx, env.projectID, tracker.CreateIssueInput{
		Title: "Other", IssueType: "task", Priority: "low", ReporterID: env.userID,
	})
	require.NoError(t, err)

	board, err := env.repo.CreateBoard(env.ctx, env.projectID, "Full filter", map[string]interface{}{
		"assignee_id": assignee.ID.String(),
		"issue_type":  "bug",
		"priority":    "high",
		"label_ids":   []interface{}{lbl.ID.String()},
	})
	require.NoError(t, err)

	data, err := env.repo.GetBoardData(env.ctx, env.projectID, board.ID)
	require.NoError(t, err)
	require.NotNil(t, data.Edges.Project)
	statuses := data.Edges.Project.Edges.WorkflowStatuses
	require.NotEmpty(t, statuses)
	var issueCount int
	for _, st := range statuses {
		issueCount += len(st.Edges.Issues)
	}
	assert.Equal(t, 1, issueCount)
}

// TestHandler_Automation_DeleteInvalidRuleID проверяет отклонение при HTTP-обработчик автоматизация удаление некорректный rule id.
func TestHandler_Automation_DeleteInvalidRuleID(t *testing.T) {
	client := enttestOpenAutomationHandler(t)
	base := "/api/projects/" + client.projectID.String()

	req := httptest.NewRequest(http.MethodDelete, base+"/automation-rules/not-uuid", nil)
	resp, err := client.app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
