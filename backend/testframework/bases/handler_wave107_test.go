package bases_test
import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_UpdateBase_VersionConflict(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Conflict")

	updateBody, _ := json.Marshal(map[string]any{
		"name":    "Nope",
		"version": detail.Version + 10,
	})
	req := httptest.NewRequest(http.MethodPut, "/api/bases/"+detail.ID.String(), bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}

func TestHandler_DeleteBase(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Gone")

	req := httptest.NewRequest(http.MethodDelete, "/api/bases/"+detail.ID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/bases/"+detail.ID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_DeleteBase_NotFound(t *testing.T) {
	app, _ := setupBasesTest(t)

	req := httptest.NewRequest(http.MethodDelete, "/api/bases/"+uuid.New().String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestHandler_GetTableView_EmptyRecords(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "View")
	tableID := detail.Tables[0].ID

	req := httptest.NewRequest(http.MethodGet, "/api/bases/"+detail.ID.String()+"/tables/"+tableID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var view TableViewResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&view))
	assert.NotEmpty(t, view.Columns)
	assert.Empty(t, view.Records)
}
