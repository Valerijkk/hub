package bases_test
import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func createBaseViaAPI(t *testing.T, app *fiber.App, name string) BaseDetailResponse {
	t.Helper()
	body := `{"name":"` + name + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/bases", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var detail BaseDetailResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&detail))
	resp.Body.Close()
	return detail
}

func TestHandler_GetBase(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Inventory")

	req := httptest.NewRequest(http.MethodGet, "/api/bases/"+detail.ID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var got BaseDetailResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
	assert.Equal(t, detail.ID, got.ID)
	assert.Equal(t, "Inventory", got.Name)
}

func TestHandler_GetBase_NotFound(t *testing.T) {
	app, _ := setupBasesTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/bases/00000000-0000-0000-0000-000000000099", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_CreateTableAndColumn(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "CRM")

	tableBody, _ := json.Marshal(map[string]string{"name": "Contacts"})
	req := httptest.NewRequest(http.MethodPost, "/api/bases/"+detail.ID.String()+"/tables", bytes.NewReader(tableBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	require.Equal(t, http.StatusCreated, resp.StatusCode)
	var table TableSummary
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&table))
	resp.Body.Close()

	colBody, _ := json.Marshal(map[string]string{"name": "Email", "col_type": "text"})
	req = httptest.NewRequest(http.MethodPost, "/api/bases/"+detail.ID.String()+"/tables/"+table.ID.String()+"/columns", bytes.NewReader(colBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}

func TestHandler_UpdateRecord(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Tasks")
	tableID := detail.Tables[0].ID

	req := httptest.NewRequest(http.MethodGet, "/api/bases/"+detail.ID.String()+"/tables/"+tableID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	var view TableViewResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&view))
	resp.Body.Close()
	colID := view.Columns[0].ID.String()

	recBody, _ := json.Marshal(map[string]any{"cells": map[string]any{colID: "Todo"}})
	req = httptest.NewRequest(http.MethodPost, "/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records", bytes.NewReader(recBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var record RecordResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&record))
	resp.Body.Close()

	updateBody, _ := json.Marshal(map[string]any{
		"cells":   map[string]any{colID: "Done"},
		"version": record.Version,
	})
	req = httptest.NewRequest(http.MethodPut, "/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records/"+record.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_UpdateRecord_VersionConflict(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Conflict DB")
	tableID := detail.Tables[0].ID

	req := httptest.NewRequest(http.MethodGet, "/api/bases/"+detail.ID.String()+"/tables/"+tableID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	var view TableViewResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&view))
	resp.Body.Close()
	colID := view.Columns[0].ID.String()

	recBody, _ := json.Marshal(map[string]any{"cells": map[string]any{colID: "v1"}})
	req = httptest.NewRequest(http.MethodPost, "/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records", bytes.NewReader(recBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var record RecordResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&record))
	resp.Body.Close()

	updateBody, _ := json.Marshal(map[string]any{"cells": map[string]any{colID: "v2"}, "version": record.Version + 99})
	req = httptest.NewRequest(http.MethodPut, "/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records/"+record.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestHandler_InvalidUUID(t *testing.T) {
	app, _ := setupBasesTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/bases/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
