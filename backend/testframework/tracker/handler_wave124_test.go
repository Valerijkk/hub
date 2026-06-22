package tracker_test

// Тесты HTTP-обработчика трекера задач (волна 124): интеграционные и unit-сценарии testframework.
import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Фаза 124 — handler error branches for workflow delete..

func TestHandler_Wave124_DeleteWorkflowTransition_InvalidID(t *testing.T) {
	app, projectID, _, _ := setupTrackerExtendedApp(t)
	base := "/api/projects/" + projectID.String()

	req := httptest.NewRequest(http.MethodDelete, base+"/workflow/transitions/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
