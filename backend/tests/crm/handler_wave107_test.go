package crm_test
import (
	"github.com/valerii/hub/backend/internal/crm"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_UpdateDeal_VersionConflict(t *testing.T) {
	app, _, projID := setupCRMTest(t)

	dealBody, _ := json.Marshal(crm.CreateDealRequest{
		Title:     "Conflict deal",
		Stage:     crm.StageLead,
		ProjectID: &projID,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/crm/deals", bytes.NewReader(dealBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var deal crm.DealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))
	resp.Body.Close()

	stage := crm.StageQualified
	updateBody, _ := json.Marshal(crm.UpdateDealRequest{Stage: &stage, Version: deal.Version + 99})
	req = httptest.NewRequest(http.MethodPut, "/api/crm/deals/"+deal.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestHandler_CreateDeal_InvalidJSON(t *testing.T) {
	app, _, _ := setupCRMTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/crm/deals", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_DeleteDeal_NotFound(t *testing.T) {
	app, _, _ := setupCRMTest(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/crm/deals/"+uuid.New().String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_UpdateContact_VersionConflict(t *testing.T) {
	app, _, _ := setupCRMTest(t)

	body := `{"name":"Gamma Inc","email":"g@test.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/crm/contacts", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var contact crm.ContactResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&contact))
	resp.Body.Close()

	updateBody, _ := json.Marshal(map[string]any{"name": "Gamma LLC", "version": contact.Version + 50})
	req = httptest.NewRequest(http.MethodPut, "/api/crm/contacts/"+contact.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestHandler_UpdateDeal_InvalidStage(t *testing.T) {
	app, _, projID := setupCRMTest(t)

	dealBody, _ := json.Marshal(crm.CreateDealRequest{
		Title:     "Bad stage",
		ProjectID: &projID,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/crm/deals", bytes.NewReader(dealBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var deal crm.DealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))
	resp.Body.Close()

	badStage := crm.DealStage("bogus")
	updateBody, _ := json.Marshal(crm.UpdateDealRequest{Stage: &badStage, Version: deal.Version})
	req = httptest.NewRequest(http.MethodPut, "/api/crm/deals/"+deal.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
