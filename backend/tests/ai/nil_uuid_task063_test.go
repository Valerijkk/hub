//go:build hubtest

package ai_test

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

	"github.com/valerii/hub/backend/internal/ai"
)

// Mutation: пропустить uuid.Nil check → не 400 invalid id.
func TestAdminUpdateProvider_NilUUID_P3BE09(t *testing.T) {
	h := ai.NewAdminHandler(ai.NewService(&mockRepo{}, &mockAdapter{}, nil), func(_ fiber.Ctx, _ uuid.UUID) (bool, error) {
		return true, nil
	})
	app := fiber.New(fiber.Config{ErrorHandler: func(c fiber.Ctx, err error) error {
		if e, ok := err.(*fiber.Error); ok {
			return c.Status(e.Code).SendString(e.Message)
		}
		return c.Status(500).SendString(err.Error())
	}})
	api := app.Group("/api")
	auth := func(c fiber.Ctx) error {
		c.Locals("userID", uuid.New().String())
		return c.Next()
	}
	h.RegisterRoutes(api, auth)

	req := httptest.NewRequest(http.MethodPatch, "/api/admin/ai/providers/"+uuid.Nil.String(), bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
	assert.Contains(t, string(body), "invalid id")
}
