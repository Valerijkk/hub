//go:build hubtest

package marketplace_test

// TASK-059 / P3-BE-03 / W7-A5-016: HTTP DELETE stranger → 403 (не только service-layer).
// Mutation: убрать owner/admin check в DeleteListing → 204 → FAIL.

import (
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
	"github.com/valerii/hub/backend/internal/marketplace"
	"github.com/valerii/hub/backend/internal/platform/middleware"
)

func TestHandler_DeleteListing_Forbidden_W7A5016(t *testing.T) {
	dsn := "file:mp_del_059_" + uuid.New().String() + "?mode=memory&cache=shared&_fk=1"
	client := enttest.Open(t, "sqlite3", dsn)
	t.Cleanup(func() { _ = client.Close() })
	ctx := context.Background()

	owner, err := client.User.Create().
		SetEmail("owner-del-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)
	stranger, err := client.User.Create().
		SetEmail("stranger-del-" + uuid.New().String()[:8] + "@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	svc := NewService(NewEntRepository(client), nil, nil)
	h := marketplace.NewHandler(svc)
	app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		if uid := c.Get("X-Test-User-ID"); uid != "" {
			c.Locals("userID", uid)
		}
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	createBody, _ := json.Marshal(map[string]any{
		"title": "Owned desk", "description": "Mine", "category": "furniture",
	})
	req := httptest.NewRequest(http.MethodPost, "/api/marketplace/listings", bytes.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Test-User-ID", owner.ID.String())
	resp, err := app.Test(req)
	require.NoError(t, err)
	var created ListingResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	resp.Body.Close()
	require.Equal(t, http.StatusCreated, resp.StatusCode)

	req = httptest.NewRequest(http.MethodDelete, "/api/marketplace/listings/"+created.ID.String(), nil)
	req.Header.Set("X-Test-User-ID", stranger.ID.String())
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
}
