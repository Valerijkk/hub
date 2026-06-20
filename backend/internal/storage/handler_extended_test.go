package storage

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_ConfirmUpload_InvalidFileID(t *testing.T) {
	client, _, userID, _ := newStorageTestClient(t)
	h := &Handler{storage: &StorageService{entClient: client, bucket: "hub-files"}}

	app := fiber.New()
	t.Cleanup(func() { _ = app.Shutdown() })
	app.Post("/files/:fileId/confirm", func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return h.ConfirmUpload(c)
	})

	body := `{"entityType":"issue","entityId":"iss-1"}`
	req := httptest.NewRequest(http.MethodPost, "/files/not-uuid/confirm", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ConfirmUpload_FileNotFound(t *testing.T) {
	client, _, userID, _ := newStorageTestClient(t)
	h := &Handler{storage: &StorageService{entClient: client, bucket: "hub-files"}}

	app := fiber.New()
	t.Cleanup(func() { _ = app.Shutdown() })
	app.Post("/files/:fileId/confirm", func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return h.ConfirmUpload(c)
	})

	body := `{"entityType":"issue","entityId":"iss-1"}`
	req := httptest.NewRequest(http.MethodPost, "/files/"+uuid.New().String()+"/confirm", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestHandler_DeleteFile_InvalidID(t *testing.T) {
	h := &Handler{storage: &StorageService{}}
	app := fiber.New()
	t.Cleanup(func() { _ = app.Shutdown() })
	app.Delete("/files/:fileId", func(c fiber.Ctx) error {
		c.Locals("userID", uuid.New().String())
		return h.DeleteFile(c)
	})

	req := httptest.NewRequest(http.MethodDelete, "/files/bad-id", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_GetDownloadURL_FileNotFound(t *testing.T) {
	client, _, userID, _ := newStorageTestClient(t)
	h := &Handler{storage: &StorageService{entClient: client}}

	app := fiber.New()
	t.Cleanup(func() { _ = app.Shutdown() })
	app.Get("/files/:fileId/url", func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return h.GetDownloadURL(c)
	})

	req := httptest.NewRequest(http.MethodGet, "/files/"+uuid.New().String()+"/url", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
}

func TestHandler_GenerateUploadURL_InvalidProjectID(t *testing.T) {
	h := &Handler{storage: &StorageService{}}
	app := fiber.New()
	t.Cleanup(func() { _ = app.Shutdown() })
	app.Post("/upload-url", func(c fiber.Ctx) error {
		c.Locals("userID", uuid.New().String())
		return h.GenerateUploadURL(c)
	})

	payload := `{"filename":"a.txt","contentType":"text/plain","projectId":"bad"}`
	req := httptest.NewRequest(http.MethodPost, "/upload-url", bytes.NewReader([]byte(payload)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
