package crm_test
import (
	"github.com/valerii/hub/backend/internal/crm"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
)

func setupCRMTest(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:crm?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("sales@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Client Alpha").
		SetKey("ALPHA").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(owner.ID).
		Save(ctx)
	require.NoError(t, err)

	repo := crm.NewEntRepository(client)
	svc := crm.NewService(repo)
	h := crm.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", owner.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, owner.ID, proj.ID
}

func TestHandler_CreateDealAndPipeline(t *testing.T) {
	app, _, projID := setupCRMTest(t)

	body := `{"name":"Acme Corp","email":"buyer@acme.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/crm/contacts", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var contact crm.ContactResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&contact))

	dealBody, err := json.Marshal(map[string]interface{}{
		"title":        "Website redesign",
		"stage":        "lead",
		"amount_cents": 15000000,
		"contact_id":   contact.ID,
		"project_id":   projID,
	})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPost, "/api/crm/deals", bytes.NewReader(dealBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/crm/pipeline", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var pipeline crm.PipelineResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&pipeline))
	require.NotEmpty(t, pipeline.Stages)
	assert.Equal(t, 1, len(pipeline.Stages[0].Deals))
	assert.Equal(t, "Website redesign", pipeline.Stages[0].Deals[0].Title)
}

func TestHandler_MoveDealStage(t *testing.T) {
	app, _, projID := setupCRMTest(t)

	dealBody, err := json.Marshal(crm.CreateDealRequest{
		Title:     "Support contract",
		Stage:     crm.StageLead,
		ProjectID: &projID,
	})
	require.NoError(t, err)

	req := httptest.NewRequest(http.MethodPost, "/api/crm/deals", bytes.NewReader(dealBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var deal crm.DealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))

	stage := crm.StageQualified
	updateBody, err := json.Marshal(crm.UpdateDealRequest{
		Stage:   &stage,
		Version: deal.Version,
	})
	require.NoError(t, err)

	req = httptest.NewRequest(http.MethodPut, "/api/crm/deals/"+deal.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated crm.DealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	assert.Equal(t, crm.StageQualified, updated.Stage)
}

func TestHandler_UpdateDeleteContact(t *testing.T) {
	app, _, _ := setupCRMTest(t)

	body := `{"name":"Beta LLC","email":"beta@corp.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/crm/contacts", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	var contact crm.ContactResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&contact))

	updateBody, _ := json.Marshal(map[string]any{
		"name":    "Beta Corp",
		"version": contact.Version,
	})
	req = httptest.NewRequest(http.MethodPut, "/api/crm/contacts/"+contact.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, "/api/crm/contacts/"+contact.ID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandler_DeleteDeal(t *testing.T) {
	app, _, projID := setupCRMTest(t)

	dealBody, _ := json.Marshal(crm.CreateDealRequest{
		Title:     "To delete",
		ProjectID: &projID,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/crm/deals", bytes.NewReader(dealBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var deal crm.DealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))

	req = httptest.NewRequest(http.MethodDelete, "/api/crm/deals/"+deal.ID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}
