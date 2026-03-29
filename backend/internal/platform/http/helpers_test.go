package http

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper to run a Fiber handler and return the response.
func testRequest(t *testing.T, app *fiber.App, method, path string) *http.Response {
	t.Helper()
	req := httptest.NewRequest(method, path, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	return resp
}

func TestUserIDFromCtx_Success(t *testing.T) {
	app := fiber.New()
	expected := uuid.New()

	app.Get("/test", func(c fiber.Ctx) error {
		c.Locals("userID", expected.String())
		id, err := UserIDFromCtx(c)
		require.NoError(t, err)
		assert.Equal(t, expected, id)
		return c.SendStatus(200)
	})

	resp := testRequest(t, app,"GET", "/test")
	assert.Equal(t, 200, resp.StatusCode)
}

func TestUserIDFromCtx_MissingLocal(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		_, err := UserIDFromCtx(c)
		require.Error(t, err)
		var fiberErr *fiber.Error
		require.ErrorAs(t, err, &fiberErr)
		assert.Equal(t, fiber.StatusUnauthorized, fiberErr.Code)
		return c.SendStatus(200)
	})

	testRequest(t, app,"GET", "/test")
}

func TestUserIDFromCtx_EmptyString(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		c.Locals("userID", "")
		_, err := UserIDFromCtx(c)
		require.Error(t, err)
		var fiberErr *fiber.Error
		require.ErrorAs(t, err, &fiberErr)
		assert.Equal(t, fiber.StatusUnauthorized, fiberErr.Code)
		return c.SendStatus(200)
	})

	testRequest(t, app,"GET", "/test")
}

func TestUserIDFromCtx_InvalidUUID(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		c.Locals("userID", "not-a-uuid")
		id, err := UserIDFromCtx(c)
		require.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
		var fiberErr *fiber.Error
		require.ErrorAs(t, err, &fiberErr)
		assert.Equal(t, fiber.StatusUnauthorized, fiberErr.Code)
		return c.SendStatus(200)
	})

	testRequest(t, app,"GET", "/test")
}

func TestUserIDFromCtx_WrongType(t *testing.T) {
	app := fiber.New()

	app.Get("/test", func(c fiber.Ctx) error {
		c.Locals("userID", 12345) // int, not string
		_, err := UserIDFromCtx(c)
		require.Error(t, err)
		return c.SendStatus(200)
	})

	testRequest(t, app,"GET", "/test")
}

func TestParseUUIDParam_Success(t *testing.T) {
	app := fiber.New()
	expected := uuid.New()

	app.Get("/items/:itemId", func(c fiber.Ctx) error {
		id, err := ParseUUIDParam(c, "itemId")
		require.NoError(t, err)
		assert.Equal(t, expected, id)
		return c.SendStatus(200)
	})

	resp := testRequest(t, app,"GET", "/items/"+expected.String())
	assert.Equal(t, 200, resp.StatusCode)
}

func TestParseUUIDParam_MissingParam(t *testing.T) {
	app := fiber.New()

	app.Get("/items", func(c fiber.Ctx) error {
		_, err := ParseUUIDParam(c, "itemId")
		require.Error(t, err)
		var fiberErr *fiber.Error
		require.ErrorAs(t, err, &fiberErr)
		assert.Equal(t, fiber.StatusBadRequest, fiberErr.Code)
		assert.Contains(t, fiberErr.Message, "itemId")
		return c.SendStatus(200)
	})

	testRequest(t, app,"GET", "/items")
}

func TestParseUUIDParam_InvalidUUID(t *testing.T) {
	app := fiber.New()

	app.Get("/items/:itemId", func(c fiber.Ctx) error {
		id, err := ParseUUIDParam(c, "itemId")
		require.Error(t, err)
		assert.Equal(t, uuid.Nil, id)
		var fiberErr *fiber.Error
		require.ErrorAs(t, err, &fiberErr)
		assert.Equal(t, fiber.StatusBadRequest, fiberErr.Code)
		assert.Contains(t, fiberErr.Message, "invalid")
		return c.SendStatus(200)
	})

	testRequest(t, app,"GET", "/items/not-valid")
}

func TestParseProjectID_Success(t *testing.T) {
	app := fiber.New()
	expected := uuid.New()

	app.Get("/projects/:projectId", func(c fiber.Ctx) error {
		id, err := ParseProjectID(c)
		require.NoError(t, err)
		assert.Equal(t, expected, id)
		return c.SendStatus(200)
	})

	resp := testRequest(t, app,"GET", "/projects/"+expected.String())
	assert.Equal(t, 200, resp.StatusCode)
}

func TestParseProjectID_Invalid(t *testing.T) {
	app := fiber.New()

	app.Get("/projects/:projectId", func(c fiber.Ctx) error {
		_, err := ParseProjectID(c)
		require.Error(t, err)
		var fiberErr *fiber.Error
		require.ErrorAs(t, err, &fiberErr)
		assert.Equal(t, fiber.StatusBadRequest, fiberErr.Code)
		return c.SendStatus(200)
	})

	testRequest(t, app,"GET", "/projects/abc")
}

func TestParseProjectID_Missing(t *testing.T) {
	app := fiber.New()

	app.Get("/projects", func(c fiber.Ctx) error {
		_, err := ParseProjectID(c)
		require.Error(t, err)
		return c.SendStatus(200)
	})

	testRequest(t, app,"GET", "/projects")
}

// Verify that testRequest helper drains and closes body.
func TestTestRequest_Helper(t *testing.T) {
	app := fiber.New()
	app.Get("/ping", func(c fiber.Ctx) error {
		return c.SendString("pong")
	})
	resp := testRequest(t, app,"GET", "/ping")
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, "pong", string(body))
}
