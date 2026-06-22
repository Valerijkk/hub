package marketplace_test
import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_UpdateListing_Success(t *testing.T) {
	app, _, memberID := setupMarketplaceTest(t)

	createBody, _ := json.Marshal(map[string]any{
		"title": "Original", "description": "Desc", "category": "books",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/marketplace/listings", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	var created ListingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	updateBody, _ := json.Marshal(map[string]any{
		"title": "Updated title", "version": created.Version,
	})
	url := fmt.Sprintf("/api/marketplace/listings/%s", created.ID)
	req = httptest.NewRequest(http.MethodPut, url, bytes.NewReader(updateBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var updated ListingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&updated))
	assert.Equal(t, "Updated title", updated.Title)
}

func TestHandler_DeleteListing_Success(t *testing.T) {
	app, adminID, memberID := setupMarketplaceTest(t)

	createBody, _ := json.Marshal(map[string]any{
		"title": "To delete", "description": "Desc", "category": "equipment",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/marketplace/listings", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	var created ListingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	url := fmt.Sprintf("/api/marketplace/listings/%s", created.ID)
	req = httptest.NewRequest(http.MethodDelete, url, nil)
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestHandler_ListListings_Pagination(t *testing.T) {
	app, _, memberID := setupMarketplaceTest(t)

	req := httptest.NewRequest(http.MethodGet, "/api/marketplace/listings?page=2&per_page=10", nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
