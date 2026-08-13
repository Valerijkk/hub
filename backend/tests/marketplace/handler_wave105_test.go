package marketplace_test
import (
	"github.com/valerii/hub/backend/internal/marketplace"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMapServiceError_AllBranches(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		err  error
		code int
	}{
		{"not found", marketplace.ErrNotFound, fiber.StatusNotFound},
		{"forbidden admin", ErrForbidden, fiber.StatusForbidden},
		{"not owner", ErrNotOwner, fiber.StatusForbidden},
		{"invalid input", ErrInvalidInput, fiber.StatusBadRequest},
		{"version conflict", ErrVersionConflict, fiber.StatusConflict},
		{"invalid status", ErrInvalidStatus, fiber.StatusBadRequest},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var fe *fiber.Error
			require.ErrorAs(t, mapServiceError(tc.err), &fe)
			assert.Equal(t, tc.code, fe.Code)
		})
	}

	var fe *fiber.Error
	require.ErrorAs(t, mapServiceError(errors.New("unexpected")), &fe)
	assert.Equal(t, fiber.StatusInternalServerError, fe.Code)
}

func TestHandler_WithdrawListing_ErrorPaths(t *testing.T) {
	app, adminID, memberID := setupMarketplaceTest(t)

	missingID := uuid.New()
	body, _ := json.Marshal(map[string]any{"version": 1})
	req := httptest.NewRequest(http.MethodPost, "/api/marketplace/listings/"+missingID.String()+"/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)

	createBody, _ := json.Marshal(map[string]any{
		"title": "Shared desk", "description": "Large", "category": "furniture",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	var created ListingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	stranger := uuid.New()
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings/"+created.ID.String()+"/withdraw", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", stranger.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	conflictBody, _ := json.Marshal(map[string]any{"version": created.Version + 99})
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings/"+created.ID.String()+"/withdraw", bytes.NewReader(conflictBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)

	approveBody, _ := json.Marshal(map[string]any{"version": created.Version})
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings/"+created.ID.String()+"/approve", bytes.NewReader(approveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	soldBody, _ := json.Marshal(map[string]any{"version": created.Version})
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings/"+created.ID.String()+"/sold", bytes.NewReader(soldBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	// Valid version so status guard runs (sold → ErrInvalidStatus → 400), not version<1 → 409.
	soldWithdrawBody, _ := json.Marshal(map[string]any{"version": created.Version})
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings/"+created.ID.String()+"/withdraw", bytes.NewReader(soldWithdrawBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateListing_InvalidInput(t *testing.T) {
	app, _, memberID := setupMarketplaceTest(t)

	body, _ := json.Marshal(map[string]any{
		"title": "", "description": "missing title", "category": "books",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/marketplace/listings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ApproveListing_VersionConflict(t *testing.T) {
	app, adminID, memberID := setupMarketplaceTest(t)

	createBody, _ := json.Marshal(map[string]any{
		"title": "Monitor", "description": "4K", "category": "equipment",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/marketplace/listings", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	var created ListingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()

	approveBody, _ := json.Marshal(map[string]any{"version": created.Version + 5})
	url := fmt.Sprintf("/api/marketplace/listings/%s/approve", created.ID)
	req = httptest.NewRequest(http.MethodPost, url, bytes.NewReader(approveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusConflict, resp.StatusCode)
}
