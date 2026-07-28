package tracker_test

// Тесты HTTP-обработчика трекера задач: интеграционные и unit-сценарии testframework.
import (
	"github.com/valerii/hub/backend/internal/tracker"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent/enttest"
)

// TestHandler_AutomationRulesCRUD проверяет HTTP-слой: HTTP-обработчик автоматизация rules crud.
func TestHandler_AutomationRulesCRUD(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:automation_crud_"+uuid.New().String()+"?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { _ = client.Close() })
	ctx := t.Context()
	user, err := client.User.Create().SetEmail("crud@test.com").Save(ctx)
	require.NoError(t, err)
	proj, err := client.Project.Create().SetName("Auto Project").SetKey("CRD").Save(ctx)
	require.NoError(t, err)
	status, err := client.WorkflowStatus.Create().SetName("Review").SetCategory("in_progress").SetProject(proj).Save(ctx)
	require.NoError(t, err)
	ch, err := client.Channel.Create().SetName("hub").SetChannelType("public").SetProjectID(proj.ID).SetCreatedBy(user.ID).Save(ctx)
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
	pg.Get("/automation-rules", h.ListAutomationRules)
	pg.Post("/automation-rules", h.CreateAutomationRule)
	pg.Patch("/automation-rules/:ruleId", h.UpdateAutomationRule)
	pg.Delete("/automation-rules/:ruleId", h.DeleteAutomationRule)

	base := "/api/projects/" + proj.ID.String()
	createBody := fmt.Sprintf(`{
		"name":"Notify on review",
		"trigger_type":"status_changed",
		"action_type":"post_to_channel",
		"trigger_config":{"to_status_id":%q},
		"action_config":{"channel_id":%q}
	}`, status.ID, ch.ID)

	req := httptest.NewRequest(http.MethodPost, base+"/automation-rules", bytes.NewReader([]byte(createBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var rule tracker.AutomationRuleResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rule))
	require.NotEmpty(t, rule.ID)

	req = httptest.NewRequest(http.MethodGet, base+"/automation-rules", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var rules []tracker.AutomationRuleResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&rules))
	assert.Len(t, rules, 1)

	patchBody := `{"name":"Renamed rule","enabled":false}`
	req = httptest.NewRequest(http.MethodPatch, base+"/automation-rules/"+rule.ID, bytes.NewReader([]byte(patchBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, base+"/automation-rules/"+rule.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

// TestHandler_UpdateAndDeleteComment проверяет HTTP-слой: HTTP-обработчик обновление and удаление comment.
func TestHandler_UpdateAndDeleteComment(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()

	issue := createIssueViaHandler(t, app, projectID, "Comment CRUD")

	body := `{"body_json":{"text":"first"},"body_text":"first"}`
	req := httptest.NewRequest(http.MethodPost, base+"/issues/"+issue.IssueKey+"/comments", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var comment tracker.CommentResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&comment))

	updateBody := `{"body_json":{"text":"updated"},"body_text":"updated","version":1}`
	req = httptest.NewRequest(http.MethodPut, base+"/issues/"+issue.IssueKey+"/comments/"+comment.ID, bytes.NewReader([]byte(updateBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, base+"/issues/"+issue.IssueKey+"/comments/"+comment.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
