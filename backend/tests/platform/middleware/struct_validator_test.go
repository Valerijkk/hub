package middleware_test

// Тесты StructValidator (P2-BE-01): теги validate исполняются после Bind.
import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	apperrors "github.com/valerii/hub/backend/internal/platform/errors"
	"github.com/valerii/hub/backend/internal/platform/middleware"
)

// sampleDTO зеркалит типичные теги Hub DTO (title + email).
type sampleDTO struct {
	Title string `json:"title" validate:"required,min=1,max=500"`
	Email string `json:"email" validate:"required,email"`
}

func decodeTZError(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	body := make(map[string]any)
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	return body
}

func mountSampleBind(app *fiber.App) {
	app.Post("/bind", func(c fiber.Ctx) error {
		var req sampleDTO
		if err := c.Bind().JSON(&req); err != nil {
			return fiber.NewError(fiber.StatusBadRequest, apperrors.MsgBadRequest)
		}
		return c.Status(fiber.StatusCreated).JSON(req)
	})
}

// TestStructValidator_EmptyTitleEmail_Mutation: с validator → 400; без → 201.
func TestStructValidator_EmptyTitleEmail_Mutation(t *testing.T) {
	payload := `{"title":"","email":"not-an-email"}`

	t.Run("with_validator_rejected", func(t *testing.T) {
		app := fiber.New(fiber.Config{
			ErrorHandler:    middleware.ErrorHandler,
			StructValidator: middleware.NewStructValidator(),
		})
		app.Use(middleware.RequestID())
		mountSampleBind(app)

		req := httptest.NewRequest(http.MethodPost, "/bind", bytes.NewReader([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body := decodeTZError(t, resp)
		assert.NotEmpty(t, body["error"])
	})

	t.Run("without_validator_accepted", func(t *testing.T) {
		app := fiber.New(fiber.Config{ErrorHandler: middleware.ErrorHandler})
		app.Use(middleware.RequestID())
		mountSampleBind(app)

		req := httptest.NewRequest(http.MethodPost, "/bind", bytes.NewReader([]byte(payload)))
		req.Header.Set("Content-Type", "application/json")
		resp, err := app.Test(req)
		require.NoError(t, err)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusCreated, resp.StatusCode)
		var got sampleDTO
		require.NoError(t, json.NewDecoder(resp.Body).Decode(&got))
		assert.Empty(t, got.Title)
		assert.Equal(t, "not-an-email", got.Email)
	})
}

// TestStructValidator_ValidPayload_OK — happy path с validator.
func TestStructValidator_ValidPayload_OK(t *testing.T) {
	app := fiber.New(fiber.Config{
		ErrorHandler:    middleware.ErrorHandler,
		StructValidator: middleware.NewStructValidator(),
	})
	mountSampleBind(app)

	payload := `{"title":"Task","email":"user@example.com"}`
	req := httptest.NewRequest(http.MethodPost, "/bind", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)
}
