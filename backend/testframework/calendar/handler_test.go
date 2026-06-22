package calendar_test
import (
	"github.com/valerii/hub/backend/internal/calendar"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	_ "github.com/mattn/go-sqlite3"

	"github.com/valerii/hub/backend/ent/enttest"
	"github.com/valerii/hub/backend/ent/issue"
	"github.com/valerii/hub/backend/ent/project"
	"github.com/valerii/hub/backend/ent/user"
)

func setupCalendarTest(t *testing.T) (*fiber.App, uuid.UUID, uuid.UUID) {
	t.Helper()
	client := enttest.Open(t, "sqlite3", "file:cal?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	ctx := context.Background()

	member, err := client.User.Create().
		SetEmail("dev@hub.local").
		SetSystemRole(user.SystemRoleMember).
		Save(ctx)
	require.NoError(t, err)

	proj, err := client.Project.Create().
		SetName("Hub").
		SetKey("HUB").
		SetStatus(project.StatusActive).
		Save(ctx)
	require.NoError(t, err)

	_, err = client.ProjectMember.Create().
		SetProjectID(proj.ID).
		SetUserID(member.ID).
		Save(ctx)
	require.NoError(t, err)

	status, err := client.WorkflowStatus.Create().
		SetName("Todo").
		SetCategory("todo").
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 14, 0, 0, 0, 0, time.UTC)
	_, err = client.Sprint.Create().
		SetName("Sprint 1").
		SetGoal("Ship calendar").
		SetStartDate(start).
		SetEndDate(end).
		SetProjectID(proj.ID).
		Save(ctx)
	require.NoError(t, err)

	due := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	_, err = client.Issue.Create().
		SetIssueNumber(1).
		SetIssueKey("HUB-1").
		SetTitle("Calendar deadline").
		SetIssueType(issue.IssueTypeTask).
		SetDueDate(due).
		SetProjectID(proj.ID).
		SetStatusID(status.ID).
		Save(ctx)
	require.NoError(t, err)

	repo := calendar.NewEntRepository(client)
	svc := calendar.NewService(repo)
	h := calendar.NewHandler(svc)
	app := fiber.New()
	api := app.Group("/api")
	authMw := func(c fiber.Ctx) error {
		c.Locals("userID", member.ID.String())
		return c.Next()
	}
	h.RegisterRoutes(api, authMw)
	return app, member.ID, proj.ID
}

func TestHandler_ListEvents_IncludesSprintAndDeadline(t *testing.T) {
	app, _, _ := setupCalendarTest(t)
	from := "2026-06-01T00:00:00Z"
	to := "2026-06-30T23:59:59Z"

	req := httptest.NewRequest(http.MethodGet, "/api/calendar/events?from="+from+"&to="+to, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	var body calendar.EventsResponse
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&body))
	require.NotEmpty(t, body.Events)

	types := map[calendar.EventType]bool{}
	for _, ev := range body.Events {
		types[ev.Type] = true
	}
	assert.True(t, types[calendar.EventTypeSprint])
	assert.True(t, types[calendar.EventTypeDeadline])
}

func TestHandler_ExportICal_ReturnsCalendar(t *testing.T) {
	app, _, _ := setupCalendarTest(t)
	from := "2026-06-01T00:00:00Z"
	to := "2026-06-30T23:59:59Z"

	req := httptest.NewRequest(http.MethodGet, "/api/calendar/export.ics?from="+from+"&to="+to, nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Contains(t, resp.Header.Get("Content-Type"), "text/calendar")
}

func TestService_ListEvents_InvalidRange(t *testing.T) {
	client := enttest.Open(t, "sqlite3", "file:cal2?mode=memory&cache=shared&_fk=1")
	t.Cleanup(func() { client.Close() })
	svc := calendar.NewService(calendar.NewEntRepository(client))
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	_, err := svc.ListEvents(context.Background(), uuid.New(), from, to)
	assert.ErrorIs(t, err, calendar.ErrInvalidRange)
}
