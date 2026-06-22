package project_test

// Тесты HTTP-обработчика проектов (волна 122): интеграционные и unit-сценарии testframework.
import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Фаза 122 — team CRUD and archive lifecycle via HTTP..

func TestHandler_Wave122_UnarchiveProject(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "UNAR")

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/archive", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/projects/"+created.ID+"/unarchive", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave122_ListInvites_Empty проверяет HTTP-слой: HTTP-обработчик wave 122 список invites пустой.
func TestHandler_Wave122_ListInvites_Empty(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)
	created := createProjectViaHandler(t, app, "INVL")

	req := httptest.NewRequest(http.MethodGet, "/api/projects/"+created.ID+"/invites", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_Wave122_ListTrash_Empty проверяет HTTP-слой: HTTP-обработчик wave 122 список корзина пустой.
func TestHandler_Wave122_ListTrash_Empty(t *testing.T) {
	app, _ := setupProjectHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/trash", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
