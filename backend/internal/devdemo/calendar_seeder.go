package devdemo

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
)

type calendarSeeder struct {
	calendar CalendarDemoPort
}

type calendarSeedResult struct {
	datasetVersion       int
	eventsTotal          int
	eventsReplayed       int
	rsvpsTotal           int
	rsvpsReplayed        int
	occurrencesCancelled int
}

func newCalendarSeeder(calendar CalendarDemoPort) (*calendarSeeder, error) {
	if isNilCalendarDemoDependency(calendar) {
		return nil, errors.New("calendar demo seeder: calendar port is required")
	}
	return &calendarSeeder{calendar: calendar}, nil
}

func (seeder *calendarSeeder) seed(ctx context.Context, userIDs map[string]uuid.UUID) (calendarSeedResult, error) {
	if seeder == nil || isNilCalendarDemoDependency(seeder.calendar) {
		return calendarSeedResult{}, errors.New("calendar demo seed: seeder is unavailable")
	}
	if isNilCalendarDemoDependency(ctx) {
		return calendarSeedResult{}, errors.New("calendar demo seed: context is required")
	}

	dataset, err := newCalendarDemoDataset(userIDs)
	if err != nil {
		return calendarSeedResult{}, fmt.Errorf("calendar demo seed: dataset preparation failed: %w", err)
	}

	result := calendarSeedResult{datasetVersion: dataset.version}
	expectedRSVPs := 0
	expectedCancellations := 0
	for _, plan := range dataset.events {
		expectedRSVPs += len(plan.rsvps)
		if !plan.cancelRecurrenceID.IsZero() {
			expectedCancellations++
		}

		created, err := seeder.calendar.CreateEvent(ctx, plan.actorID, plan.create)
		if err != nil {
			return calendarSeedResult{}, fmt.Errorf("calendar demo seed: create event failed: %w", err)
		}
		if created.EventID == uuid.Nil {
			return calendarSeedResult{}, errors.New("calendar demo seed: create event response is invalid")
		}

		attendeeVersions, err := calendarDemoAttendeeVersions(created.Attendees)
		if err != nil {
			return calendarSeedResult{}, err
		}
		result.eventsTotal++
		if created.Replayed {
			result.eventsReplayed++
		}

		for _, rsvpPlan := range plan.rsvps {
			expectedVersion, ok := attendeeVersions[rsvpPlan.actorID]
			if !ok {
				return calendarSeedResult{}, errors.New("calendar demo seed: RSVP attendee version is unavailable")
			}

			rsvp, err := seeder.calendar.RespondSeries(ctx, rsvpPlan.actorID, CalendarDemoRSVPInput{
				IdempotencyKey:          rsvpPlan.idempotencyKey,
				EventID:                 created.EventID,
				ExpectedAttendeeVersion: expectedVersion,
				Status:                  rsvpPlan.status,
			})
			if err != nil {
				return calendarSeedResult{}, fmt.Errorf("calendar demo seed: respond to event failed: %w", err)
			}
			if rsvp.UserID == uuid.Nil || rsvp.UserID != rsvpPlan.actorID || rsvp.Status != rsvpPlan.status || rsvp.Version <= 0 {
				return calendarSeedResult{}, errors.New("calendar demo seed: RSVP response is invalid")
			}
			result.rsvpsTotal++
			if rsvp.Replayed {
				result.rsvpsReplayed++
			}
		}

		if !plan.cancelRecurrenceID.IsZero() {
			if err := seeder.cancelOccurrence(ctx, plan.actorID, created.EventID, plan.cancelRecurrenceID); err != nil {
				return calendarSeedResult{}, err
			}
			result.occurrencesCancelled++
		}
	}

	if result.eventsTotal != len(dataset.events) || result.rsvpsTotal != expectedRSVPs || result.occurrencesCancelled != expectedCancellations {
		return calendarSeedResult{}, errors.New("calendar demo seed: reconciliation totals are invalid")
	}
	return result, nil
}

func calendarDemoAttendeeVersions(attendees []CalendarDemoAttendeeState) (map[uuid.UUID]int, error) {
	versions := make(map[uuid.UUID]int, len(attendees))
	for _, attendee := range attendees {
		if attendee.UserID == uuid.Nil || attendee.Version <= 0 {
			return nil, errors.New("calendar demo seed: attendee response is invalid")
		}
		if _, exists := versions[attendee.UserID]; exists {
			return nil, errors.New("calendar demo seed: attendee response is duplicated")
		}
		versions[attendee.UserID] = attendee.Version
	}
	return versions, nil
}

func (seeder *calendarSeeder) cancelOccurrence(ctx context.Context, actorID, eventID uuid.UUID, recurrenceID time.Time) error {
	masterVersion, err := seeder.currentMasterVersion(ctx, actorID, eventID)
	if err != nil {
		return err
	}

	cancelled, err := seeder.calendar.CancelOccurrence(ctx, actorID, CalendarDemoOccurrenceCancelInput{
		EventID:               eventID,
		RecurrenceID:          recurrenceID,
		ExpectedMasterVersion: masterVersion,
	})
	if err == nil {
		return validateCalendarDemoOccurrenceCancellation(cancelled, eventID, recurrenceID)
	}
	if !errors.Is(err, ErrCalendarDemoVersionConflict) {
		return fmt.Errorf("calendar demo seed: cancel occurrence failed: %w", err)
	}

	masterVersion, err = seeder.currentMasterVersion(ctx, actorID, eventID)
	if err != nil {
		return err
	}
	cancelled, err = seeder.calendar.CancelOccurrence(ctx, actorID, CalendarDemoOccurrenceCancelInput{
		EventID:               eventID,
		RecurrenceID:          recurrenceID,
		ExpectedMasterVersion: masterVersion,
	})
	if err != nil {
		return fmt.Errorf("calendar demo seed: retry cancel occurrence failed: %w", err)
	}
	return validateCalendarDemoOccurrenceCancellation(cancelled, eventID, recurrenceID)
}

func (seeder *calendarSeeder) currentMasterVersion(ctx context.Context, actorID, eventID uuid.UUID) (int, error) {
	master, err := seeder.calendar.GetEvent(ctx, actorID, eventID)
	if err != nil {
		return 0, fmt.Errorf("calendar demo seed: get occurrence master failed: %w", err)
	}
	if master.EventID != eventID || master.Version <= 0 {
		return 0, errors.New("calendar demo seed: occurrence master response is invalid")
	}
	return master.Version, nil
}

func validateCalendarDemoOccurrenceCancellation(
	cancelled CalendarDemoOccurrenceCancelResult,
	eventID uuid.UUID,
	recurrenceID time.Time,
) error {
	if cancelled.EventID != eventID || !cancelled.RecurrenceID.Equal(recurrenceID) || cancelled.Status != "cancelled" ||
		cancelled.Sequence <= 0 || cancelled.EventVersion <= 0 {
		return errors.New("calendar demo seed: occurrence cancellation response is invalid")
	}
	return nil
}

func isNilCalendarDemoDependency(value any) bool {
	if value == nil {
		return true
	}

	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
