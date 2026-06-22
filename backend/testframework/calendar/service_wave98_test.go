package calendar_test
import (
	"github.com/valerii/hub/backend/internal/calendar"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/leaverequest"
)

func TestExpandCallEvents_OneOff(t *testing.T) {
	at := time.Date(2026, 6, 15, 14, 0, 0, 0, time.UTC)
	sc := &ent.ScheduledCall{
		ID:          uuid.New(),
		Title:       "Sync call",
		ChannelID:   uuid.New(),
		ProjectID:   uuid.New(),
		ScheduledAt: at,
		Edges:       ent.ScheduledCallEdges{Project: &ent.Project{Key: "HUB"}},
	}
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	events := calendar.ExportExpandCallEvents(sc, from, to)
	require.Len(t, events, 1)
	assert.Equal(t, calendar.EventTypeCall, events[0].Type)
	assert.Equal(t, "Sync call", events[0].Title)
}

func TestExpandCallEvents_RecurringRRULE(t *testing.T) {
	base := time.Date(2026, 6, 2, 9, 0, 0, 0, time.UTC)
	sc := &ent.ScheduledCall{
		ID:          uuid.New(),
		Title:       "Weekly standup",
		ChannelID:   uuid.New(),
		ProjectID:   uuid.New(),
		ScheduledAt: base,
		Rrule:       "FREQ=WEEKLY;COUNT=3",
		Edges:       ent.ScheduledCallEdges{Project: &ent.Project{Key: "HUB"}},
	}
	from := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2026, 6, 30, 0, 0, 0, 0, time.UTC)

	events := calendar.ExportExpandCallEvents(sc, from, to)
	assert.GreaterOrEqual(t, len(events), 2)
}

func TestParseRRule_AddsDTSTART(t *testing.T) {
	base := time.Date(2026, 6, 1, 10, 0, 0, 0, time.UTC)
	rule, err := calendar.ExportParseRRule(base, "FREQ=DAILY;COUNT=2")
	require.NoError(t, err)
	require.NotNil(t, rule)
}

func TestLeaveEvent_FormatsTitle(t *testing.T) {
	start := time.Date(2026, 6, 5, 0, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	ev := calendar.ExportLeaveEvent(&ent.LeaveRequest{
		ID:        uuid.New(),
		LeaveType: leaverequest.LeaveTypeVacation,
		StartDate: start,
		EndDate:   end,
		Edges: ent.LeaveRequestEdges{
			Requester: &ent.User{Name: "Alice", Email: "alice@hub.local"},
		},
	})
	assert.Equal(t, calendar.EventTypeLeave, ev.Type)
	assert.Contains(t, ev.Title, "Alice")
	assert.Contains(t, ev.Title, "Отпуск")
}

func TestValidateRange_TooWide(t *testing.T) {
	from := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC)
	to := from.Add(400 * 24 * time.Hour)
	assert.ErrorIs(t, calendar.ExportValidateRange(from, to), calendar.ErrRangeTooWide)
}

func TestHandler_ListEvents_InvalidRange(t *testing.T) {
	app, _, _ := setupCalendarTest(t)
	req := httptest.NewRequest(http.MethodGet,
		"/api/calendar/events?from=2026-07-01T00:00:00Z&to=2026-06-01T00:00:00Z", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ListEvents_MissingParams(t *testing.T) {
	app, _, _ := setupCalendarTest(t)
	req := httptest.NewRequest(http.MethodGet, "/api/calendar/events", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestHandler_ListEvents_RangeTooWide(t *testing.T) {
	app, _, _ := setupCalendarTest(t)
	req := httptest.NewRequest(http.MethodGet,
		"/api/calendar/events?from=2026-01-01T00:00:00Z&to=2027-06-01T00:00:00Z", nil)
	resp, err := app.Test(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
