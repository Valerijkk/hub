package calendar_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/valerii/hub/backend/ent"
	"github.com/valerii/hub/backend/ent/calendarevent"
	"github.com/valerii/hub/backend/ent/calendareventexception"
	"github.com/valerii/hub/backend/internal/calendar"
)

func TestService_ListEvents_MergesCanonicalEventOccurrences(t *testing.T) {
	actorID := uuid.MustParse("10000000-0000-4000-8000-000000000001")
	eventID := uuid.MustParse("20000000-0000-4000-8000-000000000002")
	organizationID := uuid.MustParse("30000000-0000-4000-8000-000000000003")
	startsAt := time.Date(2027, time.February, 1, 10, 0, 0, 0, time.UTC)
	endsAt := startsAt.Add(time.Hour)
	from := startsAt.Add(30 * time.Minute)
	to := startsAt.AddDate(0, 0, 4)
	rrule := "FREQ=DAILY;COUNT=3"
	cancelled := calendareventexception.StatusCancelled
	thirdOccurrence := startsAt.Add(48 * time.Hour)
	tieEventID := uuid.MustParse("50000000-0000-4000-8000-000000000005")
	boundaryEventID := uuid.MustParse("60000000-0000-4000-8000-000000000006")
	boundaryEndsAt := to.Add(time.Hour)

	repository := &canonicalCalendarListRepository{
		events: []*ent.CalendarEvent{{
			ID:              eventID,
			OrganizationID:  organizationID,
			CalendarOwnerID: actorID,
			Title:           "Architecture review",
			Description:     "Review canonical calendar list",
			StartsAt:        &startsAt,
			EndsAt:          &endsAt,
			Timezone:        "UTC",
			Rrule:           &rrule,
			Exdates:         []string{startsAt.Add(24 * time.Hour).Format(time.RFC3339)},
			Status:          calendarevent.StatusConfirmed,
			Edges: ent.CalendarEventEdges{Exceptions: []*ent.CalendarEventException{{
				ID:              uuid.MustParse("40000000-0000-4000-8000-000000000004"),
				CalendarEventID: eventID,
				RecurrenceID:    thirdOccurrence.Format(time.RFC3339),
				Status:          &cancelled,
			}}},
		}, {
			ID:              tieEventID,
			OrganizationID:  organizationID,
			CalendarOwnerID: actorID,
			Title:           "Architecture review",
			StartsAt:        &startsAt,
			EndsAt:          &endsAt,
			Timezone:        "UTC",
			Status:          calendarevent.StatusConfirmed,
		}, {
			ID:              boundaryEventID,
			OrganizationID:  organizationID,
			CalendarOwnerID: actorID,
			Title:           "Boundary event",
			StartsAt:        &to,
			EndsAt:          &boundaryEndsAt,
			Timezone:        "UTC",
			Status:          calendarevent.StatusConfirmed,
		}},
	}

	response, err := calendar.NewService(repository).ListEvents(
		context.Background(),
		actorID,
		from,
		to,
	)

	require.NoError(t, err)
	assert.Equal(t, from, repository.from)
	assert.Equal(t, to, repository.to)
	require.Len(t, response.Events, 2)
	event := response.Events[0]
	assert.Equal(t, calendar.EventType("event"), event.Type)
	assert.Equal(t, "event-"+eventID.String()+"-"+startsAt.Format(time.RFC3339), event.ID)
	assert.Equal(t, "Architecture review", event.Title)
	assert.Equal(t, "Review canonical calendar list", event.Description)
	assert.Equal(t, startsAt, event.StartsAt)
	require.NotNil(t, event.EndsAt)
	assert.Equal(t, endsAt, *event.EndsAt)
	assert.False(t, event.AllDay)
	assert.Equal(t, "/calendar/events/"+eventID.String(), event.URL)
	assert.Equal(t, "event-"+tieEventID.String()+"-"+startsAt.Format(time.RFC3339), response.Events[1].ID)
	assert.Equal(t, startsAt, response.Events[1].StartsAt)
	assert.Equal(t, "Architecture review", response.Events[1].Title)
}

type canonicalCalendarListRepository struct {
	events []*ent.CalendarEvent
	from   time.Time
	to     time.Time
}

func (repository canonicalCalendarListRepository) ListUserProjectIDs(context.Context, uuid.UUID) ([]uuid.UUID, error) {
	return nil, nil
}

func (repository canonicalCalendarListRepository) ListScheduledCalls(context.Context, []uuid.UUID) ([]*ent.ScheduledCall, error) {
	return nil, nil
}

func (repository canonicalCalendarListRepository) ListSprintsInRange(context.Context, []uuid.UUID, time.Time, time.Time) ([]*ent.Sprint, error) {
	return nil, nil
}

func (repository canonicalCalendarListRepository) ListIssuesDueInRange(context.Context, []uuid.UUID, time.Time, time.Time) ([]*ent.Issue, error) {
	return nil, nil
}

func (repository canonicalCalendarListRepository) ListApprovedLeaveInRange(context.Context, uuid.UUID, []uuid.UUID, time.Time, time.Time) ([]*ent.LeaveRequest, error) {
	return nil, nil
}

func (repository *canonicalCalendarListRepository) ListCalendarEventsForActor(_ context.Context, _ uuid.UUID, from, to time.Time) ([]*ent.CalendarEvent, error) {
	repository.from = from
	repository.to = to
	return repository.events, nil
}
