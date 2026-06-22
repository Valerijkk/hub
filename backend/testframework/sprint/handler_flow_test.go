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

func setupSprintHandlerApp(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID) {
	t.Helper()
	repo := newMockRepository()
	rdb, _ := newPokerTestRedis(t)
	svc := sprint.NewSprintService(repo, nil, nil)
	poker := sprint.NewPokerService(rdb, nil)
	h := sprint.NewHandler(svc, poker)
	projectID := uuid.New()
	userID := uuid.New()

	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", userID.String())
		return c.Next()
	}

	sprints := api.Group("/projects/:projectId/sprints", authMw)
	sprints.Post("/", h.CreateSprint)
	sprints.Get("/", h.ListSprints)
	sprints.Get("/:sprintId", h.GetSprint)
	sprints.Put("/:sprintId", h.UpdateSprint)
	sprints.Post("/:sprintId/start", h.StartSprint)
	sprints.Post("/:sprintId/complete", h.CompleteSprint)
	sprints.Put("/:sprintId/retro", h.UpdateRetroNotes)
	sprints.Get("/:sprintId/burndown", h.GetBurndown)
	sprints.Post("/:sprintId/poker", h.StartPokerSession)

	api.Get("/projects/:projectId/velocity", authMw, h.GetVelocity)

	pokerRoutes := api.Group("/sprints/poker", authMw)
	pokerRoutes.Post("/:sessionId/vote", h.CastVote)
	pokerRoutes.Post("/:sessionId/reveal", h.RevealVotes)
	pokerRoutes.Get("/:sessionId", h.GetPokerSession)

	return app, projectID, userID
}

func TestHandler_CreateListGetSprint(t *testing.T) {
	app, projectID, _ := setupSprintHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/sprints"

	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().AddDate(0, 0, 14).Format(time.RFC3339)
	body := fmt.Sprintf(`{"name":"sprint.Handler Sprint","goal":"ship it","start_date":%q,"end_date":%q}`, start, end)
	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var created sprint.SprintResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&created))
	require.NotEmpty(t, created.ID)

	req = httptest.NewRequest(http.MethodGet, base+"/", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/"+created.ID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_StartCompleteAndBurndown(t *testing.T) {
	app, projectID, _ := setupSprintHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/sprints"

	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().AddDate(0, 0, 14).Format(time.RFC3339)
	body := fmt.Sprintf(`{"name":"Lifecycle","start_date":%q,"end_date":%q}`, start, end)
	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var sp sprint.SprintResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sp))
	resp.Body.Close()

	req = httptest.NewRequest(http.MethodPost, base+"/"+sp.ID+"/start", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, base+"/"+sp.ID+"/burndown", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, base+"/"+sp.ID+"/complete", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_VelocityAndRetro(t *testing.T) {
	app, projectID, _ := setupSprintHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/sprints"

	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().AddDate(0, 0, 7).Format(time.RFC3339)
	body := fmt.Sprintf(`{"name":"Retro sprint","start_date":%q,"end_date":%q}`, start, end)
	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var sp sprint.SprintResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sp))
	resp.Body.Close()

	retroBody := `{"went_well":["fast tests"],"to_improve":["coverage"],"actions":["add handlers"]}`
	req = httptest.NewRequest(http.MethodPut, base+"/"+sp.ID+"/retro", bytes.NewReader([]byte(retroBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/projects/"+projectID.String()+"/velocity", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestHandler_PokerFlow(t *testing.T) {
	app, projectID, userID := setupSprintHandlerApp(t)
	base := "/api/projects/" + projectID.String() + "/sprints"

	start := time.Now().UTC().Format(time.RFC3339)
	end := time.Now().UTC().AddDate(0, 0, 7).Format(time.RFC3339)
	body := fmt.Sprintf(`{"name":"Poker sprint","start_date":%q,"end_date":%q}`, start, end)
	req := httptest.NewRequest(http.MethodPost, base+"/", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	var sp sprint.SprintResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&sp))
	resp.Body.Close()

	issueID := uuid.New()
	pokerBody := fmt.Sprintf(`{"issue_id":%q}`, issueID.String())
	req = httptest.NewRequest(http.MethodPost, base+"/"+sp.ID+"/poker", bytes.NewReader([]byte(pokerBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusCreated, resp.StatusCode)

	var session sprint.PokerSessionResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&session))
	require.NotEmpty(t, session.SessionID)

	voteBody := `{"value":"5"}`
	req = httptest.NewRequest(http.MethodPost, "/api/sprints/poker/"+session.SessionID+"/vote", bytes.NewReader([]byte(voteBody)))
	req.Header.Set("Content-Type", "application/json")
	resp, err = app.Test(req)
	require.NoError(t, err)
	resp.Body.Close()
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)

	req = httptest.NewRequest(http.MethodPost, "/api/sprints/poker/"+session.SessionID+"/reveal", nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	req = httptest.NewRequest(http.MethodGet, "/api/sprints/poker/"+session.SessionID, nil)
	resp, err = app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	_ = userID
}

func TestHandler_InvalidProjectID(t *testing.T) {
	h := sprint.NewHandler(sprint.NewSprintService(newMockRepository(), nil, nil), nil)
	app := fiber.New()
	app.Post("/api/projects/:projectId/sprints", h.CreateSprint)

	body := `{"name":"Bad project","start_date":"2026-06-01T00:00:00Z","end_date":"2026-06-15T00:00:00Z"}`
	req := httptest.NewRequest(http.MethodPost, "/api/projects/not-a-uuid/sprints", bytes.NewReader([]byte(body)))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, fiber.StatusBadRequest, resp.StatusCode)
}
