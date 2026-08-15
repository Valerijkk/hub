package calendar_test

import (
	"reflect"
	"strings"
	"testing"

	platformnats "github.com/valerii/hub/backend/internal/platform/nats"
)

func TestStreamSubjects_AdmitsCalendarExactlyOnce(t *testing.T) {
	want := []string{
		"project.>",
		"tracker.>",
		"messenger.channel.>",
		"tms.>",
		"calls.channel.>",
		"calls.recording.>",
		"auth.session.>",
		"sprint.>",
		"notification.>",
		"audit.>",
		"calendar.>",
	}

	if !reflect.DeepEqual(platformnats.StreamSubjects, want) {
		t.Fatalf("StreamSubjects = %#v, want %#v", platformnats.StreamSubjects, want)
	}

	calendarSubjects := 0
	for _, subject := range platformnats.StreamSubjects {
		if subject == "calendar.>" {
			calendarSubjects++
		}
		if subject == ">" || strings.HasPrefix(subject, "*.") {
			t.Fatalf("StreamSubjects contains broader wildcard %q", subject)
		}
	}
	if calendarSubjects != 1 {
		t.Fatalf("calendar.> occurs %d times, want 1", calendarSubjects)
	}
}

func TestStreamSubjects_CalendarSubjectsAreAdmittedWithoutMailOrLinkedCallExpansion(t *testing.T) {
	calendarSubjects := []string{
		"calendar.invite.created.v1",
		"calendar.invite.updated.v1",
		"calendar.invite.cancelled.v1",
		"calendar.rsvp.changed.v1",
		"calendar.reminder.due.v1",
	}

	for _, subject := range calendarSubjects {
		if !strings.HasPrefix(subject, "calendar.") || strings.TrimPrefix(subject, "calendar.") == "" {
			t.Fatalf("calendar subject %q is not covered by calendar.>", subject)
		}
	}

	for _, subject := range platformnats.StreamSubjects {
		if subject == "mail.>" || (strings.HasPrefix(subject, "calls.channel.") && strings.HasSuffix(subject, ".scheduled_due.v1")) {
			t.Fatalf("StreamSubjects contains subject outside 143-04G admission: %q", subject)
		}
	}
}

func TestWebhookFilterSubjects_StayExactAndExcludeCalendar(t *testing.T) {
	want := []string{
		"project.*.issue.>",
		"tracker.*.issue.>",
		"project.created",
		"tms.run.completed.>",
		"tms.testrun.completed",
		"messenger.channel.*.message_sent",
	}

	if !reflect.DeepEqual(platformnats.WebhookFilterSubjects, want) {
		t.Fatalf("WebhookFilterSubjects = %#v, want %#v", platformnats.WebhookFilterSubjects, want)
	}
	for _, subject := range platformnats.WebhookFilterSubjects {
		if strings.HasPrefix(subject, "calendar.") {
			t.Fatalf("WebhookFilterSubjects exposes calendar subject %q", subject)
		}
	}
}
