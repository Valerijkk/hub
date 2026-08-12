//go:build hubtest

package search_test

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/internal/search"
)

// Mutation: вернуть err.Error() в 500 body → NotContains "create failed" краснеет.
func TestHandler_CreateFilter_SafeErrorBody_P3BE08(t *testing.T) {
	svc := search.NewSearchService(&failCreateFilterRepo{}, newScopeTestDB(t), nil)
	h := search.NewHandler(svc)
	userID := uuid.New()
	app := fiber.New(fiber.Config{ErrorHandler: func(c fiber.Ctx, err error) error {
		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).SendString(e.Message)
		}
		return c.Status(500).SendString(err.Error())
	}})
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)

	body := `{"name":"Broken","filter_config":{},"project_id":"` + uuid.New().String() + `"}`
	req := httptest.NewRequest(http.MethodPost, "/api/filters/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	assert.Equal(t, fiber.StatusInternalServerError, resp.StatusCode)
	assert.Equal(t, "не удалось создать фильтр", string(raw))
	assert.NotContains(t, string(raw), "create failed")
}
