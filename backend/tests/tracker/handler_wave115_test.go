package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 115): интеграционные и unit-сценарии testframework.
import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHandler_GetIssue_InvalidProjectID проверяет отклонение при HTTP-обработчик получение issue некорректный project id.
func TestHandler_GetIssue_InvalidProjectID(t *testing.T) {
	app, _, _, _ := setupTrackerHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/not-uuid/issues/HUB-1", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}

// TestHandler_ListIssues_InvalidProjectID проверяет отклонение при HTTP-обработчик список issues некорректный project id.
func TestHandler_ListIssues_InvalidProjectID(t *testing.T) {
	app, _, _, _ := setupTrackerHandlerApp(t)

	req := httptest.NewRequest(http.MethodGet, "/api/projects/bad-id/issues", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

// TestHandler_StartTimer_InvalidProjectID проверяет отклонение при HTTP-обработчик start timer некорректный project id.
func TestHandler_StartTimer_InvalidProjectID(t *testing.T) {
	app, _, _, _ := setupTrackerHandlerApp(t)

	req := httptest.NewRequest(http.MethodPost, "/api/projects/not-uuid/issues/HUB-1/timer/start", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.NotEqual(t, http.StatusOK, resp.StatusCode)
}
