package bases_test
import (
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

func TestHandler_CreateTable_InvalidJSON(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Tables")

	req := httptest.NewRequest(http.MethodPost, "/api/bases/"+detail.ID.String()+"/tables", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateTable_EmptyName(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Tables")

	body, _ := json.Marshal(map[string]string{"name": "   "})
	req := httptest.NewRequest(http.MethodPost, "/api/bases/"+detail.ID.String()+"/tables", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateTable_BaseNotFound(t *testing.T) {
	app, _ := setupBasesTest(t)

	body, _ := json.Marshal(map[string]string{"name": "Extra"})
	req := httptest.NewRequest(http.MethodPost, "/api/bases/"+uuid.New().String()+"/tables", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestHandler_CreateColumn_InvalidJSON(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Columns")
	tableID := detail.Tables[0].ID

	req := httptest.NewRequest(http.MethodPost,
		"/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/columns",
		bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateColumn_EmptyName(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Columns")
	tableID := detail.Tables[0].ID

	body, _ := json.Marshal(map[string]string{"name": "", "col_type": "text"})
	req := httptest.NewRequest(http.MethodPost,
		"/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/columns",
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateRecord_InvalidJSON(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Records")
	tableID := detail.Tables[0].ID

	req := httptest.NewRequest(http.MethodPost,
		"/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records",
		bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_UpdateRecord_InvalidJSON(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Update")
	tableID := detail.Tables[0].ID

	req := httptest.NewRequest(http.MethodPut,
		"/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records/"+uuid.New().String(),
		bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_UpdateRecord_NotFound(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Update")
	tableID := detail.Tables[0].ID

	body, _ := json.Marshal(map[string]any{"cells": map[string]any{}})
	req := httptest.NewRequest(http.MethodPut,
		"/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records/"+uuid.New().String(),
		bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestHandler_GetTableView_TableNotFound(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "View")

	req := httptest.NewRequest(http.MethodGet,
		"/api/bases/"+detail.ID.String()+"/tables/"+uuid.New().String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}
