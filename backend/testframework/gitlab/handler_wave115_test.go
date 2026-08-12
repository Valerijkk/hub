package gitlab_test
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

func TestHandler_TriggerPipeline_InvalidJSON(t *testing.T) {
	app := testGitLabHandlerApp(t, &gitlabMockRepo{})
	projectID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/gitlab/pipelines/trigger",
		bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CreateBranch_InvalidJSON(t *testing.T) {
	app := testGitLabHandlerApp(t, &gitlabMockRepo{})
	projectID := uuid.New()

	req := httptest.NewRequest(http.MethodPost, "/api/projects/"+projectID.String()+"/issues/HUB-1/gitlab/create-branch", bytes.NewReader([]byte("[")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
