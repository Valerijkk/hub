package devdemo

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	calendarDemoDatasetVersion  = 1
	calendarDemoTimezone        = "Europe/Moscow"
	calendarDemoRequiredRole    = "required"
	calendarDemoAcceptedStatus  = "accepted"
	calendarDemoTentativeStatus = "tentative"
)

type calendarDemoDataset struct {
	version int
	events  []calendarDemoEventPlan
}

type calendarDemoEventPlan struct {
	actorID            uuid.UUID
	create             CalendarDemoCreateInput
	rsvps              []calendarDemoRSVPPlan
	cancelRecurrenceID time.Time
}

type calendarDemoRSVPPlan struct {
	actorID        uuid.UUID
	idempotencyKey string
	status         string
}

func newCalendarDemoDataset(userIDs map[string]uuid.UUID) (calendarDemoDataset, error) {
	pm, err := requiredCalendarDemoUserID(userIDs, "pm@hub.local")
	if err != nil {
		return calendarDemoDataset{}, err
	}
	dev1, err := requiredCalendarDemoUserID(userIDs, "dev1@hub.local")
	if err != nil {
		return calendarDemoDataset{}, err
	}
	dev2, err := requiredCalendarDemoUserID(userIDs, "dev2@hub.local")
	if err != nil {
		return calendarDemoDataset{}, err
	}
	tester1, err := requiredCalendarDemoUserID(userIDs, "tester1@hub.local")
	if err != nil {
		return calendarDemoDataset{}, err
	}
	tester2, err := requiredCalendarDemoUserID(userIDs, "tester2@hub.local")
	if err != nil {
		return calendarDemoDataset{}, err
	}

	weeklyReviewStart := time.Date(2026, time.August, 31, 9, 0, 0, 0, time.UTC)
	dataset := calendarDemoDataset{
		version: calendarDemoDatasetVersion,
		events: []calendarDemoEventPlan{
			{
				actorID: pm,
				create: CalendarDemoCreateInput{
					IdempotencyKey: "hub-devdemo-calendar-v1-create-team-sync",
					Title:          "Team sync",
					Description:    "Engineering and QA team sync",
					Location:       "Moscow Room A",
					StartsAt:       time.Date(2026, time.September, 1, 8, 0, 0, 0, time.UTC),
					EndsAt:         time.Date(2026, time.September, 1, 8, 30, 0, 0, time.UTC),
					Timezone:       calendarDemoTimezone,
					Attendees: []CalendarDemoAttendeeCandidate{
						{UserID: dev1, Email: "dev1@hub.local", Role: calendarDemoRequiredRole},
						{UserID: dev2, Email: "dev2@hub.local", Role: calendarDemoRequiredRole},
						{UserID: tester1, Email: "tester1@hub.local", Role: calendarDemoRequiredRole},
					},
				},
				rsvps: []calendarDemoRSVPPlan{
					{actorID: dev1, idempotencyKey: "hub-devdemo-calendar-v1-rsvp-team-sync-dev1", status: calendarDemoAcceptedStatus},
					{actorID: dev2, idempotencyKey: "hub-devdemo-calendar-v1-rsvp-team-sync-dev2", status: calendarDemoTentativeStatus},
				},
			},
			{
				actorID: dev2,
				create: CalendarDemoCreateInput{
					IdempotencyKey: "hub-devdemo-calendar-v1-create-dev2-focus",
					Title:          "Development focus",
					Description:    "Focused implementation session",
					Location:       "Focus Space",
					StartsAt:       time.Date(2026, time.September, 1, 11, 0, 0, 0, time.UTC),
					EndsAt:         time.Date(2026, time.September, 1, 12, 0, 0, 0, time.UTC),
					Timezone:       calendarDemoTimezone,
				},
			},
			{
				actorID: pm,
				create: CalendarDemoCreateInput{
					IdempotencyKey: "hub-devdemo-calendar-v1-create-release-planning",
					Title:          "Release planning",
					Description:    "Plan the upcoming release",
					Location:       "Moscow Room B",
					StartsAt:       time.Date(2026, time.September, 2, 12, 0, 0, 0, time.UTC),
					EndsAt:         time.Date(2026, time.September, 2, 13, 0, 0, 0, time.UTC),
					Timezone:       calendarDemoTimezone,
					Attendees: []CalendarDemoAttendeeCandidate{
						{UserID: dev1, Email: "dev1@hub.local", Role: calendarDemoRequiredRole},
						{UserID: tester2, Email: "tester2@hub.local", Role: calendarDemoRequiredRole},
					},
				},
				rsvps: []calendarDemoRSVPPlan{
					{actorID: dev1, idempotencyKey: "hub-devdemo-calendar-v1-rsvp-release-planning-dev1", status: calendarDemoAcceptedStatus},
					{actorID: tester2, idempotencyKey: "hub-devdemo-calendar-v1-rsvp-release-planning-tester2", status: calendarDemoAcceptedStatus},
				},
			},
			{
				actorID: pm,
				create: CalendarDemoCreateInput{
					IdempotencyKey: "hub-devdemo-calendar-v1-create-weekly-review",
					Title:          "Weekly review",
					Description:    "Review weekly delivery",
					Location:       "Review Room",
					StartsAt:       weeklyReviewStart,
					EndsAt:         weeklyReviewStart.Add(30 * time.Minute),
					Timezone:       calendarDemoTimezone,
					RRULE:          "FREQ=WEEKLY;COUNT=4",
					Attendees: []CalendarDemoAttendeeCandidate{
						{UserID: tester1, Email: "tester1@hub.local", Role: calendarDemoRequiredRole},
					},
				},
				rsvps: []calendarDemoRSVPPlan{
					{actorID: tester1, idempotencyKey: "hub-devdemo-calendar-v1-rsvp-weekly-review-tester1", status: calendarDemoAcceptedStatus},
				},
				cancelRecurrenceID: time.Date(2026, time.September, 7, 9, 0, 0, 0, time.UTC),
			},
		},
	}
	if err := validateCalendarDemoDataset(dataset); err != nil {
		return calendarDemoDataset{}, err
	}
	return dataset, nil
}

func requiredCalendarDemoUserID(userIDs map[string]uuid.UUID, email string) (uuid.UUID, error) {
	if userIDs == nil {
		return uuid.Nil, errors.New("calendar demo dataset: users are required")
	}
	userID, ok := userIDs[email]
	if !ok || userID == uuid.Nil {
		return uuid.Nil, errors.New("calendar demo dataset: required user is unavailable")
	}
	return userID, nil
}

func validateCalendarDemoDataset(dataset calendarDemoDataset) error {
	if dataset.version != calendarDemoDatasetVersion || len(dataset.events) == 0 {
		return errors.New("calendar demo dataset: invalid version or events")
	}
	keys := make(map[string]struct{})
	for _, event := range dataset.events {
		if event.actorID == uuid.Nil || !calendarDemoUTCInterval(event.create.StartsAt, event.create.EndsAt) {
			return errors.New("calendar demo dataset: invalid event identity or interval")
		}
		if event.create.Timezone != calendarDemoTimezone || !calendarDemoUniqueKey(keys, event.create.IdempotencyKey) {
			return errors.New("calendar demo dataset: invalid event timezone or key")
		}
		attendees := make(map[uuid.UUID]struct{}, len(event.create.Attendees))
		for _, attendee := range event.create.Attendees {
			if attendee.UserID == uuid.Nil || attendee.Role != calendarDemoRequiredRole {
				return errors.New("calendar demo dataset: invalid attendee")
			}
			attendees[attendee.UserID] = struct{}{}
		}
		for _, rsvp := range event.rsvps {
			if rsvp.actorID == uuid.Nil || !calendarDemoUniqueKey(keys, rsvp.idempotencyKey) {
				return errors.New("calendar demo dataset: invalid RSVP identity or key")
			}
			if _, ok := attendees[rsvp.actorID]; !ok {
				return errors.New("calendar demo dataset: RSVP actor is not an attendee")
			}
			if rsvp.status != calendarDemoAcceptedStatus && rsvp.status != calendarDemoTentativeStatus {
				return errors.New("calendar demo dataset: invalid RSVP status")
			}
		}
		if !event.cancelRecurrenceID.IsZero() {
			if event.create.RRULE != "FREQ=WEEKLY;COUNT=4" || event.cancelRecurrenceID.Location() != time.UTC ||
				!event.cancelRecurrenceID.Equal(event.create.StartsAt.Add(7*24*time.Hour)) {
				return errors.New("calendar demo dataset: invalid recurrence cancellation")
			}
		}
	}
	return nil
}

func calendarDemoUTCInterval(startsAt, endsAt time.Time) bool {
	return startsAt.Location() == time.UTC && endsAt.Location() == time.UTC && startsAt.Before(endsAt)
}

func calendarDemoUniqueKey(keys map[string]struct{}, key string) bool {
	if len(key) < 16 || len(key) > 128 {
		return false
	}
	if _, exists := keys[key]; exists {
		return false
	}
	keys[key] = struct{}{}
	return true
}
