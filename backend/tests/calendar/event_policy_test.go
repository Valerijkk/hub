package calendar_test

import (
	"errors"
	"testing"
	"time"

	"github.com/valerii/hub/backend/internal/calendar"
)

func TestCalendarEvent_Validate(t *testing.T) {
	start := time.Date(2026, time.August, 20, 9, 0, 0, 0, time.UTC)
	end := time.Date(2026, time.August, 20, 10, 0, 0, 0, time.UTC)

	tests := []struct {
		name    string
		event   calendar.CalendarEvent
		wantErr error
	}{
		{
			name: "valid timed event",
			event: calendar.CalendarEvent{
				StartsAt: start,
				EndsAt:   end,
				Timezone: "Europe/Moscow",
			},
		},
		{
			name: "timed event requires start",
			event: calendar.CalendarEvent{
				EndsAt:   end,
				Timezone: "Europe/Moscow",
			},
			wantErr: calendar.ErrInvalidTimedEvent,
		},
		{
			name: "timed event requires exclusive end",
			event: calendar.CalendarEvent{
				StartsAt: start,
				EndsAt:   start,
				Timezone: "Europe/Moscow",
			},
			wantErr: calendar.ErrInvalidTimedEvent,
		},
		{
			name: "timed event rejects all day dates",
			event: calendar.CalendarEvent{
				StartsAt:  start,
				EndsAt:    end,
				StartDate: "2026-08-20",
				Timezone:  "Europe/Moscow",
			},
			wantErr: calendar.ErrInvalidTimedEvent,
		},
		{
			name: "valid all day event",
			event: calendar.CalendarEvent{
				AllDay:           true,
				StartDate:        "2026-08-20",
				EndDateExclusive: "2026-08-21",
				Timezone:         "Europe/Moscow",
			},
		},
		{
			name: "all day event rejects timed instants",
			event: calendar.CalendarEvent{
				AllDay:           true,
				StartsAt:         start,
				EndsAt:           end,
				StartDate:        "2026-08-20",
				EndDateExclusive: "2026-08-21",
				Timezone:         "Europe/Moscow",
			},
			wantErr: calendar.ErrInvalidAllDayEvent,
		},
		{
			name: "all day event requires dates",
			event: calendar.CalendarEvent{
				AllDay:   true,
				Timezone: "Europe/Moscow",
			},
			wantErr: calendar.ErrInvalidAllDayEvent,
		},
		{
			name: "all day event requires exclusive end date",
			event: calendar.CalendarEvent{
				AllDay:           true,
				StartDate:        "2026-08-20",
				EndDateExclusive: "2026-08-20",
				Timezone:         "Europe/Moscow",
			},
			wantErr: calendar.ErrInvalidAllDayEvent,
		},
		{
			name: "all day event rejects malformed start date",
			event: calendar.CalendarEvent{
				AllDay:           true,
				StartDate:        "2026-02-30",
				EndDateExclusive: "2026-03-01",
				Timezone:         "Europe/Moscow",
			},
			wantErr: calendar.ErrInvalidAllDayEvent,
		},
		{
			name: "all day event rejects timestamp as date",
			event: calendar.CalendarEvent{
				AllDay:           true,
				StartDate:        "2026-08-20T00:00:00Z",
				EndDateExclusive: "2026-08-21",
				Timezone:         "Europe/Moscow",
			},
			wantErr: calendar.ErrInvalidAllDayEvent,
		},
		{
			name: "all day event rejects end before start",
			event: calendar.CalendarEvent{
				AllDay:           true,
				StartDate:        "2026-08-21",
				EndDateExclusive: "2026-08-20",
				Timezone:         "Europe/Moscow",
			},
			wantErr: calendar.ErrInvalidAllDayEvent,
		},
		{
			name: "timed event rejects only all day exclusive end date",
			event: calendar.CalendarEvent{
				StartsAt:         start,
				EndsAt:           end,
				EndDateExclusive: "2026-08-21",
				Timezone:         "Europe/Moscow",
			},
			wantErr: calendar.ErrInvalidTimedEvent,
		},
		{
			name: "timezone must not be empty",
			event: calendar.CalendarEvent{
				StartsAt: start,
				EndsAt:   end,
			},
			wantErr: calendar.ErrInvalidTimezone,
		},
		{
			name: "timezone must be a loadable IANA name",
			event: calendar.CalendarEvent{
				StartsAt: start,
				EndsAt:   end,
				Timezone: "UTC+03:00",
			},
			wantErr: calendar.ErrInvalidTimezone,
		},
		{
			name: "timezone rejects unknown IANA name",
			event: calendar.CalendarEvent{
				StartsAt: start,
				EndsAt:   end,
				Timezone: "Mars/Olympus_Mons",
			},
			wantErr: calendar.ErrInvalidTimezone,
		},
		{
			name: "timezone must not use process local zone",
			event: calendar.CalendarEvent{
				StartsAt: start,
				EndsAt:   end,
				Timezone: "Local",
			},
			wantErr: calendar.ErrInvalidTimezone,
		},
		{
			name: "all day event validates IANA timezone",
			event: calendar.CalendarEvent{
				AllDay:           true,
				StartDate:        "2026-08-20",
				EndDateExclusive: "2026-08-21",
				Timezone:         "Mars/Olympus_Mons",
			},
			wantErr: calendar.ErrInvalidTimezone,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.event.Validate()
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Validate() error = %v, want errors.Is(..., %v)", err, tc.wantErr)
			}
		})
	}
}

func TestRSVPStatus_Validate(t *testing.T) {
	tests := []struct {
		name    string
		status  calendar.RSVPStatus
		wantErr error
	}{
		{name: "needs action", status: calendar.RSVPNeedsAction},
		{name: "accepted", status: calendar.RSVPAccepted},
		{name: "declined", status: calendar.RSVPDeclined},
		{name: "tentative", status: calendar.RSVPTentative},
		{name: "empty", status: "", wantErr: calendar.ErrInvalidRSVPStatus},
		{name: "unknown", status: "maybe", wantErr: calendar.ErrInvalidRSVPStatus},
		{name: "wrong case", status: "Accepted", wantErr: calendar.ErrInvalidRSVPStatus},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := tc.status.Validate()
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("Validate() error = %v, want errors.Is(..., %v)", err, tc.wantErr)
			}
		})
	}
}

func TestClassifyEventUpdate(t *testing.T) {
	tests := []struct {
		name string
		input calendar.EventUpdate
		want  calendar.EventUpdatePolicy
	}{
		{
			name: "unchanged event has no mutation policy",
			want: calendar.EventUpdatePolicy{
				RSVPReset:    calendar.RSVPResetNone,
				CancelIntent: calendar.CancelIntentNone,
			},
		},
		{
			name:  "time change resets all non organizers",
			input: calendar.EventUpdate{TimeChanged: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:  1,
				SequenceIncrement: 1,
				RSVPReset:         calendar.RSVPResetAllNonOrganizers,
				CancelIntent:      calendar.CancelIntentNone,
				SendRequest:       true,
			},
		},
		{
			name:  "timezone change resets all non organizers",
			input: calendar.EventUpdate{TimezoneChanged: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:  1,
				SequenceIncrement: 1,
				RSVPReset:         calendar.RSVPResetAllNonOrganizers,
				CancelIntent:      calendar.CancelIntentNone,
				SendRequest:       true,
			},
		},
		{
			name:  "all day change resets all non organizers",
			input: calendar.EventUpdate{AllDayChanged: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:  1,
				SequenceIncrement: 1,
				RSVPReset:         calendar.RSVPResetAllNonOrganizers,
				CancelIntent:      calendar.CancelIntentNone,
				SendRequest:       true,
			},
		},
		{
			name:  "location change resets all non organizers",
			input: calendar.EventUpdate{LocationChanged: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:  1,
				SequenceIncrement: 1,
				RSVPReset:         calendar.RSVPResetAllNonOrganizers,
				CancelIntent:      calendar.CancelIntentNone,
				SendRequest:       true,
			},
		},
		{
			name:  "recurrence change resets affected non organizers",
			input: calendar.EventUpdate{RecurrenceChanged: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:  1,
				SequenceIncrement: 1,
				RSVPReset:         calendar.RSVPResetAffectedNonOrganizers,
				CancelIntent:      calendar.CancelIntentNone,
				SendRequest:       true,
			},
		},
		{
			name:  "title or agenda change preserves rsvp",
			input: calendar.EventUpdate{TitleOrAgendaChanged: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:       1,
				SequenceIncrement:      1,
				RSVPReset:              calendar.RSVPResetNone,
				CancelIntent:           calendar.CancelIntentNone,
				PreserveExistingRSVP:   true,
				SendRequest:            true,
			},
		},
		{
			name:  "notified description change preserves rsvp",
			input: calendar.EventUpdate{DescriptionChanged: true, NotifyAttendees: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:     1,
				SequenceIncrement:    1,
				RSVPReset:            calendar.RSVPResetNone,
				CancelIntent:         calendar.CancelIntentNone,
				PreserveExistingRSVP: true,
				SendRequest:          true,
			},
		},
		{
			name:  "silent description change preserves sequence and rsvp",
			input: calendar.EventUpdate{DescriptionChanged: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:     1,
				RSVPReset:            calendar.RSVPResetNone,
				CancelIntent:         calendar.CancelIntentNone,
				PreserveExistingRSVP: true,
			},
		},
		{
			name:  "reminder only increments version",
			input: calendar.EventUpdate{ReminderChanged: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:     1,
				RSVPReset:            calendar.RSVPResetNone,
				CancelIntent:         calendar.CancelIntentNone,
				PreserveExistingRSVP: true,
			},
		},
		{
			name:  "added attendee preserves existing and initializes added",
			input: calendar.EventUpdate{AddedAttendees: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:          1,
				SequenceIncrement:         1,
				RSVPReset:                 calendar.RSVPResetNone,
				CancelIntent:              calendar.CancelIntentNone,
				PreserveExistingRSVP:      true,
				InitializeAddedAttendees:  true,
				SendRequest:               true,
			},
		},
		{
			name:  "removed attendee preserves remaining rsvp and sends cancel",
			input: calendar.EventUpdate{RemovedAttendees: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:     1,
				SequenceIncrement:    1,
				RSVPReset:            calendar.RSVPResetNone,
				CancelIntent:         calendar.CancelIntentRemovedAttendees,
				PreserveExistingRSVP: true,
			},
		},
		{
			name:  "cancellation increments sequence and preserves historical rsvp",
			input: calendar.EventUpdate{Cancelled: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:     1,
				SequenceIncrement:    1,
				RSVPReset:            calendar.RSVPResetNone,
				CancelIntent:         calendar.CancelIntentWholeEvent,
				PreserveExistingRSVP: true,
			},
		},
		{
			name: "time and recurrence reset all non organizers",
			input: calendar.EventUpdate{TimeChanged: true, RecurrenceChanged: true},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:  1,
				SequenceIncrement: 1,
				RSVPReset:         calendar.RSVPResetAllNonOrganizers,
				CancelIntent:      calendar.CancelIntentNone,
				SendRequest:       true,
			},
		},
		{
			name: "reset with add and remove combines request and targeted cancel",
			input: calendar.EventUpdate{
				TimeChanged:      true,
				AddedAttendees:   true,
				RemovedAttendees: true,
			},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:         1,
				SequenceIncrement:        1,
				RSVPReset:                calendar.RSVPResetAllNonOrganizers,
				InitializeAddedAttendees: true,
				SendRequest:              true,
				CancelIntent:             calendar.CancelIntentRemovedAttendees,
			},
		},
		{
			name: "title and remove combines request and targeted cancel",
			input: calendar.EventUpdate{
				TitleOrAgendaChanged: true,
				RemovedAttendees:     true,
			},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:     1,
				SequenceIncrement:    1,
				RSVPReset:            calendar.RSVPResetNone,
				PreserveExistingRSVP: true,
				SendRequest:          true,
				CancelIntent:         calendar.CancelIntentRemovedAttendees,
			},
		},
		{
			name: "cancellation dominates all other flags",
			input: calendar.EventUpdate{
				TitleOrAgendaChanged: true,
				DescriptionChanged:   true,
				NotifyAttendees:      true,
				TimeChanged:          true,
				TimezoneChanged:      true,
				AllDayChanged:        true,
				LocationChanged:      true,
				RecurrenceChanged:    true,
				ReminderChanged:      true,
				AddedAttendees:       true,
				RemovedAttendees:     true,
				Cancelled:            true,
			},
			want: calendar.EventUpdatePolicy{
				VersionIncrement:     1,
				SequenceIncrement:    1,
				RSVPReset:            calendar.RSVPResetNone,
				PreserveExistingRSVP: true,
				CancelIntent:         calendar.CancelIntentWholeEvent,
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := calendar.ClassifyEventUpdate(tc.input); got != tc.want {
				t.Fatalf("ClassifyEventUpdate() = %#v, want %#v", got, tc.want)
			}
		})
	}
}

func TestClassifyEventUpdate_SequenceIncrementAtMostOne(t *testing.T) {
	tests := []struct {
		name   string
		update calendar.EventUpdate
	}{
		{name: "all material changes", update: calendar.EventUpdate{
			TitleOrAgendaChanged: true,
			DescriptionChanged:   true,
			NotifyAttendees:      true,
			TimeChanged:          true,
			TimezoneChanged:      true,
			AllDayChanged:        true,
			LocationChanged:      true,
			RecurrenceChanged:    true,
			AddedAttendees:       true,
			RemovedAttendees:     true,
		}},
		{name: "recurrence with attendee changes", update: calendar.EventUpdate{
			RecurrenceChanged: true,
			AddedAttendees:    true,
			RemovedAttendees:  true,
		}},
		{name: "reminder with notified description", update: calendar.EventUpdate{
			DescriptionChanged: true,
			NotifyAttendees:    true,
			ReminderChanged:    true,
		}},
		{name: "cancellation with all changes", update: calendar.EventUpdate{
			TimeChanged:      true,
			AddedAttendees:   true,
			RemovedAttendees: true,
			Cancelled:        true,
		}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if got := calendar.ClassifyEventUpdate(tc.update).SequenceIncrement; got > 1 {
				t.Fatalf("SequenceIncrement = %d, want at most 1", got)
			}
		})
	}
}
