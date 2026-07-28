package sprint_test
import (
	"github.com/valerii/hub/backend/internal/sprint"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandler_CreateSprint_InvalidJSON(t *testing.T) {
	app, _, projectID, _ := setupSprintHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/sprints"

	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_StartSprint_NotFound(t *testing.T) {
	// W1-A4-016 / P3-BE-02: точный 404 + тело error; NotEqual(OK) пропускал 500/400.
	app, _, projectID, _ := setupSprintHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/sprints"

	req := httptest.NewRequest(http.MethodPost, base+"/"+uuid.New().String()+"/start", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusNotFound, resp.StatusCode)
	var body map[string]string
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	assert.Contains(t, body["error"], "не найден")
}

func TestHandler_UpdateRetroNotes_InvalidJSON(t *testing.T) {
	app, _, projectID, _ := setupSprintHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/sprints"

	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().AddDate(0, 0, 7).Format(time.RFC3339)
	body := fmt.Sprintf(`{"name":"Retro bad json","start_date":%q,"end_date":%q}`, start, end)
	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var sp sprint.SprintResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sp))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPut, base+"/"+sp.ID+"/retro", bytes.NewReader([]byte("[")))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_CastVote_InvalidJSON(t *testing.T) {
	app, repo, projectID, _ := setupSprintHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/sprints"

	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().AddDate(0, 0, 7).Format(time.RFC3339)
	body := fmt.Sprintf(`{"name":"Vote sprint","start_date":%q,"end_date":%q}`, start, end)
	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var sp sprint.SprintResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sp))
	resp.Body.Close()

	issueID := uuid.New()
	repo.issueProject[issueID] = projectID
	pokerBody := fmt.Sprintf(`{"issue_id":%q}`, issueID.String())
	req = httptest.NewRequest(http.MethodPost, base+"/"+sp.ID+"/poker", bytes.NewReader([]byte(pokerBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	var session sprint.PokerSessionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&session))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, "/api/sprints/poker/"+session.SessionID+"/vote", bytes.NewReader([]byte("{")))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}

func TestHandler_GetSprint_InvalidID(t *testing.T) {
	app, _, projectID, _ := setupSprintHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/sprints"

	req := httptest.NewRequest(http.MethodGet, base+"/not-a-uuid", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
