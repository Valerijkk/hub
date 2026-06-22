package crm_test
import (
	"github.com/valerii/hub/backend/internal/crm"
	"bytes"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_ListContacts(t *testing.T) {
	app, _, _ := setupCRMTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/crm/contacts", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestMapServiceError_AllBranches(t *testing.T) {
	assert.Equal(t, fiber.StatusNotFound, mapServiceError(crm.ErrNotFound).(*fiber.Error).Code)
	assert.Equal(t, fiber.StatusBadRequest, mapServiceError(crm.ErrInvalidInput).(*fiber.Error).Code)
	assert.Equal(t, fiber.StatusConflict, mapServiceError(crm.ErrVersionConflict).(*fiber.Error).Code)
	assert.Equal(t, fiber.StatusForbidden, mapServiceError(crm.ErrProjectForbidden).(*fiber.Error).Code)
	assert.Equal(t, fiber.StatusInternalServerError, mapServiceError(errors.New("other")).(*fiber.Error).Code)
}

func TestHandler_UpdateContact_NotFound(t *testing.T) {
	app, _, _ := setupCRMTest(t)

	req := httptest.NewRequest(http.MethodPut, "/api/crm/contacts/"+uuid.New().String(), bytes.NewReader([]byte(`{"name":"Ghost"}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_DeleteContact_InvalidID(t *testing.T) {
	app, _, _ := setupCRMTest(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/crm/contacts/not-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
