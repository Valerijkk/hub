package marketplace_test
import (
	"github.com/valerii/hub/backend/internal/marketplace"
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
	"github.com/valerii/hub/backend/ent/user"
)

func setupMarketplaceTest(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:marketplace?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })

	admin, err := client.User.Create().
		SetEmail("admin@hub.local").
		SetSystemRole(user.SystemRoleAdmin).
		Save(context.Background())
	require.NoError(t, err)

	member, err := client.User.Create().
		SetEmail("member@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(context.Background())
	require.NoError(t, err)

	repo := NewEntRepository(client)
	svc := NewService(repo, nil, nil)
	h := marketplace.NewHandler(svc)

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		uid := c.Get("X-Test-User-ID")
		if uid != "" {
			c.Locals("userID", uid)
		}
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, admin.ID, member.ID
}

func TestHandler_CreateApproveListings(t *testing.T) {
	app, adminID, memberID := setupMarketplaceTest(t)

	body, _ := json.Marshal(map[string]any{
		"title":       "MacBook Pro 14",
		"description": "M3, 16GB, отличное состояние",
		"category":    "equipment",
		"price_cents": 12000000,
	})
	req := httptest.NewRequest(http.MethodPost, "/api/marketplace/listings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created ListingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	assert.Equal(t, "pending", string(created.Status))

	req = httptest.NewRequest(http.MethodGet, "/api/marketplace/listings", nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	var public ListingListResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&public))
	assert.Len(t, public.Items, 0)

	approveBody, _ := json.Marshal(map[string]any{"version": created.Version})
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings/"+created.ID.String()+"/approve", bytes.NewReader(approveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/marketplace/listings", nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&public))
	require.Len(t, public.Items, 1)
	assert.Equal(t, "MacBook Pro 14", public.Items[0].Title)
}

func TestHandler_PendingQueue_AdminOnly(t *testing.T) {
	app, adminID, memberID := setupMarketplaceTest(t)

	body, _ := json.Marshal(map[string]any{
		"title":       "Книга Go",
		"description": "Effective Go",
		"category":    "books",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/marketplace/listings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/marketplace/listings/pending", nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/marketplace/listings/pending", nil)
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_RejectSoldWithdrawMine(t *testing.T) {
	app, adminID, memberID := setupMarketplaceTest(t)

	body, _ := json.Marshal(map[string]any{
		"title":       "Keyboard",
		"description": "Mechanical",
		"category":    "equipment",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/marketplace/listings", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	var created ListingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	rejectBody, _ := json.Marshal(map[string]any{"version": created.Version, "review_note": "duplicate"})
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings/"+created.ID.String()+"/reject", bytes.NewReader(rejectBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body2, _ := json.Marshal(map[string]any{
		"title":       "Desk",
		"description": "Standing desk",
		"category":    "furniture",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings", bytes.NewReader(body2))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	approveBody, _ := json.Marshal(map[string]any{"version": created.Version})
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings/"+created.ID.String()+"/approve", bytes.NewReader(approveBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", adminID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	soldBody, _ := json.Marshal(map[string]any{"version": created.Version})
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings/"+created.ID.String()+"/sold", bytes.NewReader(soldBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	body3, _ := json.Marshal(map[string]any{
		"title":       "Chair",
		"description": "Ergonomic",
		"category":    "furniture",
	})
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings", bytes.NewReader(body3))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))

	withdrawBody, _ := json.Marshal(map[string]any{"version": created.Version})
	req = httptest.NewRequest(http.MethodPost, "/api/marketplace/listings/"+created.ID.String()+"/withdraw", bytes.NewReader(withdrawBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/marketplace/listings/mine", nil)
	req.Header.Set("X-Test-User-ID", memberID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
