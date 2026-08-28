package calendar_test

import (
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
)

func TestCalendarAvailabilityReservationOccurrencesMatchUsesEffectiveStartKey(t *testing.T) {
	sourceBytes, err := os.ReadFile("../../internal/calendar/availability_reservation_ent.go")
	if err != nil {
		t.Fatalf("read availability reservation source: %v", err)
	}
	source := string(sourceBytes)
	start := strings.Index(source, "func calendarAvailabilityReservationOccurrencesMatch")
	end := strings.Index(source, "func (backend *eventEntTransactionBackend) lockActiveCalendarAvailabilityReservations")
	if start < 0 || end <= start {
		t.Fatal("availability occurrence matcher source bounds are missing")
	}
	matcher := source[start:end]
	for _, required := range []string{
		"occurrence.key == \"\"",
		"occurrence.from.IsZero()",
		"occurrence.from.Location() != time.UTC",
		"occurrence.to.IsZero()",
		"!occurrence.to.After(occurrence.from)",
		"effectiveKey := occurrence.from.UTC().Format(time.RFC3339Nano)",
		"byKey[effectiveKey]",
		"actualOccurrence, found := byKey[occurrence.OccurrenceKey]",
	} {
		if !strings.Contains(matcher, required) {
			t.Fatalf("availability occurrence matcher is missing %q", required)
		}
	}
	if strings.Contains(matcher, "byKey[occurrence.key]") {
		t.Fatal("availability occurrence matcher must not index reservations by nominal recurrence key")
	}

	nominalStart := time.Date(2031, time.January, 2, 9, 0, 0, 0, time.UTC)
	effectiveStart := nominalStart.Add(2 * time.Hour)
	effectiveEnd := effectiveStart.Add(time.Hour)
	_, err = calendar.NewCalendarAvailabilityReservationSet(
		uuid.MustParse("a1111111-1111-4111-8111-111111111111"),
		uuid.MustParse("a2222222-2222-4222-8222-222222222222"),
		uuid.MustParse("a3333333-3333-4333-8333-333333333333"),
		nil,
		[]calendar.CalendarAvailabilityReservationOccurrence{{
			OccurrenceKey: effectiveStart.Format(time.RFC3339Nano),
			StartsAt:      effectiveStart,
			EndsAt:        effectiveEnd,
		}},
	)
	if err != nil {
		t.Fatalf("moved effective reservation occurrence error = %v", err)
	}

	_, err = calendar.NewCalendarAvailabilityReservationSet(
		uuid.MustParse("a1111111-1111-4111-8111-111111111111"),
		uuid.MustParse("a2222222-2222-4222-8222-222222222222"),
		uuid.MustParse("a3333333-3333-4333-8333-333333333333"),
		nil,
		[]calendar.CalendarAvailabilityReservationOccurrence{{
			OccurrenceKey: nominalStart.Format(time.RFC3339Nano),
			StartsAt:      effectiveStart,
			EndsAt:        effectiveEnd,
		}},
	)
	if !errors.Is(err, calendar.ErrAvailabilitySourceInvalid) {
		t.Fatalf("nominal key with moved effective start error = %v, want %v", err, calendar.ErrAvailabilitySourceInvalid)
	}
}
