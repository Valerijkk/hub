package bases_test
import (
	"github.com/valerii/hub/backend/internal/bases"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
)

func TestNormalizeCells_FiltersUnknownColumns(t *testing.T) {
	colID := uuid.New()
	cols := []*ent.HubColumn{{ID: colID, Name: "Title"}}
	cells := map[string]interface{}{
		colID.String():      "ok",
		uuid.New().String(): "drop",
	}
	out := normalizeCells(cells, cols)
	require.Len(t, out, 1)
	assert.Equal(t, "ok", out[colID.String()])
}

func TestNormalizeCells_NilCells(t *testing.T) {
	out := normalizeCells(nil, nil)
	require.NotNil(t, out)
	assert.Empty(t, out)
}

func TestMapServiceError_AllBranches(t *testing.T) {
	assert.Equal(t, 404, fiberStatus(mapServiceError(bases.ErrNotFound)))
	assert.Equal(t, 400, fiberStatus(mapServiceError(ErrInvalidInput)))
	assert.Equal(t, 409, fiberStatus(mapServiceError(ErrVersionConflict)))
	assert.Equal(t, 500, fiberStatus(mapServiceError(assert.AnError)))
}

func TestHandler_DeleteRecord_Success(t *testing.T) {
	app, _ := setupBasesTest(t)
	detail := createBaseViaAPI(t, app, "Delete Me")
	tableID := detail.Tables[0].ID

	req := httptest.NewRequest(http.MethodGet, "/api/bases/"+detail.ID.String()+"/tables/"+tableID.String(), nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	var view TableViewResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&view))
	resp.Body.Close()
	colID := view.Columns[0].ID.String()

	recBody, _ := json.Marshal(map[string]any{"cells": map[string]any{colID: "row"}})
	req = httptest.NewRequest(http.MethodPost,
		"/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records",
		bytes.NewReader(recBody))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var record RecordResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&record))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodDelete,
		"/api/bases/"+detail.ID.String()+"/tables/"+tableID.String()+"/records/"+record.ID.String(), nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandler_CreateBase_InvalidBody(t *testing.T) {
	app, _ := setupBasesTest(t)
	req := httptest.NewRequest(http.MethodPost, "/api/bases", bytes.NewReader([]byte(`{"name":""}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func fiberStatus(err error) int {
	if fe, ok := err.(*fiber.Error); ok {
		return fe.Code
	}
	return 0
}
