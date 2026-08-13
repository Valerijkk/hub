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

func TestHandler_UpdateContact_InvalidLinkedUser(t *testing.T) {
	app, _, _ := setupCRMTest(t)

	body := `{"name":"Unlinked","email":"u@test.com"}`
	req := httptest.NewRequest(http.MethodPost, "/api/crm/contacts", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var contact crm.ContactResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&contact))
	resp.Body.Close()

	badID := uuid.New()
	updateBody, _ := json.Marshal(map[string]any{
		"linked_user_id": badID,
		"version":        contact.Version,
	})
	req = httptest.NewRequest(http.MethodPut, "/api/crm/contacts/"+contact.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateContact_InvalidJSON(t *testing.T) {
	app, _, _ := setupCRMTest(t)

	req := httptest.NewRequest(http.MethodPost, "/api/crm/contacts", bytes.NewReader([]byte("[")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_DeleteContact_NotFound(t *testing.T) {
	app, _, _ := setupCRMTest(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/crm/contacts/"+uuid.New().String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_UpdateDeal_InvalidContact(t *testing.T) {
	app, _, projID := setupCRMTest(t)

	dealBody, _ := json.Marshal(crm.CreateDealRequest{
		Title:     "Deal",
		ProjectID: &projID,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/crm/deals", bytes.NewReader(dealBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var deal crm.DealResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&deal))
	resp.Body.Close()

	badContact := uuid.New()
	updateBody, _ := json.Marshal(crm.UpdateDealRequest{
		ContactID: &badContact,
		Version:   deal.Version,
	})
	req = httptest.NewRequest(http.MethodPut, "/api/crm/deals/"+deal.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
