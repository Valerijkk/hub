package calendar_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
)

var (
	callPolicyOrganizationID = uuid.MustParse("10101010-1010-4010-8010-101010101010")
	callPolicyProjectID      = uuid.MustParse("20202020-2020-4020-8020-202020202020")
	callPolicyChannelID      = uuid.MustParse("30303030-3030-4030-8030-303030303030")
	callPolicyEventID        = uuid.MustParse("40404040-4040-4040-8040-404040404040")
	callPolicyScheduledID    = uuid.MustParse("50505050-5050-4050-8050-505050505050")
	callPolicyCreatorID      = uuid.MustParse("60606060-6060-4060-8060-606060606060")
	callPolicyParticipantID  = uuid.MustParse("70707070-7070-4070-8070-707070707070")
	callPolicyOccurrenceID   = uuid.MustParse("80808080-8080-4080-8080-808080808080")
)

func TestEventCallProjectionABI(t *testing.T) {
	type fields []struct {
		name string
		typ  reflect.Type
	}

	assertFields := func(value any, want fields) {
		t.Helper()
		typ := reflect.TypeOf(value)
		if typ.NumField() != len(want) {
			t.Fatalf("%s field count = %d, want %d", typ.Name(), typ.NumField(), len(want))
		}
		for index, expected := range want {
			field := typ.Field(index)
			if field.Name != expected.name || field.Type != expected.typ {
				t.Fatalf("%s field[%d] = %s %s, want %s %s", typ.Name(), index, field.Name, field.Type, expected.name, expected.typ)
			}
		}
	}

	assertFields(calendar.EventCallProjectionCreator{}, fields{
		{"UserID", reflect.TypeOf(uuid.UUID{})},
		{"Email", reflect.TypeOf("")},
		{"OrganizationID", reflect.TypeOf(uuid.UUID{})},
		{"ProjectID", reflect.TypeOf(uuid.UUID{})},
		{"ChannelID", reflect.TypeOf(uuid.UUID{})},
		{"Active", reflect.TypeOf(false)},
		{"ActiveProjectMember", reflect.TypeOf(false)},
		{"ActiveChannelMember", reflect.TypeOf(false)},
	})
	assertFields(calendar.EventCallProjectionAttendee{}, fields{
		{"UserID", reflect.TypeOf((*uuid.UUID)(nil))},
		{"Email", reflect.TypeOf("")},
		{"OrganizationID", reflect.TypeOf(uuid.UUID{})},
		{"ProjectID", reflect.TypeOf(uuid.UUID{})},
		{"Role", reflect.TypeOf(calendar.CalendarEventAttendeeRole(""))},
		{"Status", reflect.TypeOf(calendar.RSVPStatus(""))},
		{"RecurrenceID", reflect.TypeOf("")},
		{"Active", reflect.TypeOf(false)},
		{"ActiveProjectMember", reflect.TypeOf(false)},
	})
	assertFields(calendar.EventCallProjectionException{}, fields{
		{"RecurrenceID", reflect.TypeOf("")},
		{"Temporal", reflect.TypeOf((*calendar.CalendarEvent)(nil))},
		{"Status", reflect.TypeOf(calendar.CalendarEventStatus(""))},
		{"Cancelled", reflect.TypeOf(false)},
	})
	assertFields(calendar.EventCallProjectionEvent{}, fields{
		{"EventID", reflect.TypeOf(uuid.UUID{})},
		{"ScheduledCallID", reflect.TypeOf(uuid.UUID{})},
		{"OrganizationID", reflect.TypeOf(uuid.UUID{})},
		{"ProjectID", reflect.TypeOf(uuid.UUID{})},
		{"ChannelID", reflect.TypeOf(uuid.UUID{})},
		{"OrganizerID", reflect.TypeOf(uuid.UUID{})},
		{"OrganizerEmail", reflect.TypeOf("")},
		{"Source", reflect.TypeOf(calendar.CalendarEventSource(""))},
		{"Status", reflect.TypeOf(calendar.CalendarEventStatus(""))},
		{"Deleted", reflect.TypeOf(false)},
		{"Title", reflect.TypeOf("")},
		{"Description", reflect.TypeOf("")},
		{"Temporal", reflect.TypeOf(calendar.CalendarEvent{})},
		{"RRULE", reflect.TypeOf("")},
		{"Exdates", reflect.TypeOf([]string(nil))},
		{"ReminderMinutes", reflect.TypeOf((*int)(nil))},
	})
	assertFields(calendar.EventCallProjectionTarget{}, fields{
		{"ScheduledCallID", reflect.TypeOf(uuid.UUID{})},
		{"EventID", reflect.TypeOf(uuid.UUID{})},
		{"CreatorID", reflect.TypeOf(uuid.UUID{})},
		{"OrganizationID", reflect.TypeOf(uuid.UUID{})},
		{"ProjectID", reflect.TypeOf(uuid.UUID{})},
		{"ChannelID", reflect.TypeOf(uuid.UUID{})},
		{"Exists", reflect.TypeOf(false)},
		{"Active", reflect.TypeOf(false)},
		{"Deleted", reflect.TypeOf(false)},
	})
	assertFields(calendar.EventCallProjectionBoundary{}, fields{
		{"At", reflect.TypeOf(time.Time{})},
		{"Inclusive", reflect.TypeOf(false)},
	})
	assertFields(calendar.EventCallProjectionInput{}, fields{
		{"Action", reflect.TypeOf(calendar.EventCallProjectionAction(""))},
		{"Event", reflect.TypeOf(calendar.EventCallProjectionEvent{})},
		{"Target", reflect.TypeOf(calendar.EventCallProjectionTarget{})},
		{"Creator", reflect.TypeOf(calendar.EventCallProjectionCreator{})},
		{"Attendees", reflect.TypeOf([]calendar.EventCallProjectionAttendee(nil))},
		{"Exceptions", reflect.TypeOf([]calendar.EventCallProjectionException(nil))},
		{"Boundary", reflect.TypeOf(calendar.EventCallProjectionBoundary{})},
	})
	assertFields(calendar.EventCallProjectionOccurrence{}, fields{
		{"Key", reflect.TypeOf("")},
		{"StartsAt", reflect.TypeOf(time.Time{})},
		{"EndsAt", reflect.TypeOf(time.Time{})},
	})
	assertFields(calendar.EventCallProjectionSnapshot{}, fields{
		{"Disposition", reflect.TypeOf(calendar.EventCallProjectionDisposition(""))},
		{"ScheduledCallID", reflect.TypeOf(uuid.UUID{})},
		{"EventID", reflect.TypeOf(uuid.UUID{})},
		{"OrganizationID", reflect.TypeOf(uuid.UUID{})},
		{"ProjectID", reflect.TypeOf(uuid.UUID{})},
		{"ChannelID", reflect.TypeOf(uuid.UUID{})},
		{"CreatorID", reflect.TypeOf(uuid.UUID{})},
		{"Title", reflect.TypeOf("")},
		{"Description", reflect.TypeOf("")},
		{"DurationMinutes", reflect.TypeOf(0)},
		{"ReminderMinutes", reflect.TypeOf((*int)(nil))},
		{"RRULE", reflect.TypeOf("")},
		{"Occurrence", reflect.TypeOf((*calendar.EventCallProjectionOccurrence)(nil))},
		{"ParticipantIDs", reflect.TypeOf([]uuid.UUID(nil))},
	})

	if got := []calendar.EventCallProjectionAction{
		calendar.EventCallProjectionReconcile,
		calendar.EventCallProjectionPause,
		calendar.EventCallProjectionResume,
		calendar.EventCallProjectionDelete,
	}; !reflect.DeepEqual(got, []calendar.EventCallProjectionAction{"reconcile", "pause", "resume", "delete"}) {
		t.Fatalf("actions = %#v", got)
	}
	if got := []calendar.EventCallProjectionDisposition{
		calendar.EventCallProjectionUpsert,
		calendar.EventCallProjectionPaused,
		calendar.EventCallProjectionCompleted,
		calendar.EventCallProjectionDeleted,
	}; !reflect.DeepEqual(got, []calendar.EventCallProjectionDisposition{"upsert", "paused", "completed", "deleted"}) {
		t.Fatalf("dispositions = %#v", got)
	}
	if got := []calendar.EventCallProjectionErrorKind{
		calendar.EventCallProjectionInvalid,
		calendar.EventCallProjectionBoundary,
		calendar.EventCallProjectionTooLarge,
	}; !reflect.DeepEqual(got, []calendar.EventCallProjectionErrorKind{"invalid", "boundary", "too_large"}) {
		t.Fatalf("error kinds = %#v", got)
	}

	function := reflect.TypeOf(calendar.ProjectEventCall)
	if function.NumIn() != 1 || function.In(0) != reflect.TypeOf(calendar.EventCallProjectionInput{}) ||
		function.NumOut() != 2 || function.Out(0) != reflect.TypeOf(calendar.EventCallProjection{}) || function.Out(1) != reflect.TypeOf((*error)(nil)).Elem() {
		t.Fatalf("ProjectEventCall type = %s", function)
	}
	projectionType := reflect.TypeOf(calendar.EventCallProjection{})
	if projectionType.NumMethod() != 1 || projectionType.Method(0).Name != "Snapshot" || projectionType.Method(0).Type.String() != "func(calendar.EventCallProjection) calendar.EventCallProjectionSnapshot" {
		t.Fatalf("EventCallProjection methods = %d / %#v", projectionType.NumMethod(), projectionType.Method(0))
	}
	if projectionType.NumField() != 1 || projectionType.Field(0).PkgPath == "" || projectionType.Field(0).Type != reflect.TypeOf(calendar.EventCallProjectionSnapshot{}) {
		t.Fatalf("EventCallProjection must retain exactly one opaque snapshot field")
	}
	errorType := reflect.TypeOf((*calendar.EventCallProjectionError)(nil))
	errorValueType := errorType.Elem()
	if errorValueType.NumField() != 1 || errorValueType.Field(0).PkgPath == "" || errorValueType.Field(0).Type != reflect.TypeOf(calendar.EventCallProjectionErrorKind("")) || errorType.NumMethod() != 4 {
		t.Fatalf("EventCallProjectionError must retain exactly one opaque kind field and four methods")
	}
	for name, want := range map[string]reflect.Type{
		"Error":  reflect.TypeOf((*calendar.EventCallProjectionError).Error),
		"Kind":   reflect.TypeOf((*calendar.EventCallProjectionError).Kind),
		"Unwrap": reflect.TypeOf((*calendar.EventCallProjectionError).Unwrap),
		"Is":     reflect.TypeOf((*calendar.EventCallProjectionError).Is),
	} {
		method, ok := errorType.MethodByName(name)
		if !ok || method.Type != want {
			t.Fatalf("EventCallProjectionError.%s type = %v, want %v", name, method.Type, want)
		}
	}
}

func TestProjectEventCall_ReconcileProjectsFirstEffectiveOccurrence(t *testing.T) {
	input := callPolicyValidInput()
	projection, err := calendar.ProjectEventCall(input)
	if err != nil {
		t.Fatalf("ProjectEventCall() error = %v", err)
	}
	reminder := 15
	want := calendar.EventCallProjectionSnapshot{
		Disposition:    calendar.EventCallProjectionUpsert,
		ScheduledCallID: callPolicyScheduledID,
		EventID:        callPolicyEventID,
		OrganizationID: callPolicyOrganizationID,
		ProjectID:      callPolicyProjectID,
		ChannelID:      callPolicyChannelID,
		CreatorID:      callPolicyCreatorID,
		Title:          "Daily planning",
		Description:    "Discuss the next delivery.",
		DurationMinutes: 60,
		ReminderMinutes: &reminder,
		RRULE:          "FREQ=DAILY;INTERVAL=1;COUNT=3",
		Occurrence: &calendar.EventCallProjectionOccurrence{
			Key:      "2026-08-20T09:00:00Z",
			StartsAt: callPolicyStart(),
			EndsAt:   callPolicyStart().Add(time.Hour),
		},
		ParticipantIDs: []uuid.UUID{callPolicyParticipantID},
	}
	if got := projection.Snapshot(); !reflect.DeepEqual(got, want) {
		t.Fatalf("Snapshot() = %#v, want %#v", got, want)
	}
}

func TestProjectEventCall_BoundaryAndTerminalActions(t *testing.T) {
	t.Run("inclusive selects the occurrence at the boundary", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Boundary = calendar.EventCallProjectionBoundary{At: callPolicyStart(), Inclusive: true}
		projection, err := calendar.ProjectEventCall(input)
		if err != nil || projection.Snapshot().Occurrence == nil || !projection.Snapshot().Occurrence.StartsAt.Equal(callPolicyStart()) {
			t.Fatalf("inclusive result/error = %#v / %v", projection.Snapshot(), err)
		}
	})
	t.Run("strict selects the next occurrence", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Boundary = calendar.EventCallProjectionBoundary{At: callPolicyStart(), Inclusive: false}
		projection, err := calendar.ProjectEventCall(input)
		want := callPolicyStart().Add(24 * time.Hour)
		if err != nil || projection.Snapshot().Occurrence == nil || !projection.Snapshot().Occurrence.StartsAt.Equal(want) {
			t.Fatalf("strict result/error = %#v / %v", projection.Snapshot(), err)
		}
	})
	t.Run("exhausted reconcile is completed identity only", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Boundary = calendar.EventCallProjectionBoundary{At: callPolicyStart().Add(4 * 24 * time.Hour), Inclusive: true}
		projection, err := calendar.ProjectEventCall(input)
		callPolicyRequireIdentityOnly(t, projection.Snapshot(), calendar.EventCallProjectionCompleted)
		if err != nil {
			t.Fatalf("ProjectEventCall() error = %v", err)
		}
	})
	t.Run("pause uses existing active target", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Action = calendar.EventCallProjectionPause
		input.Target.Exists, input.Target.Active = true, true
		projection, err := calendar.ProjectEventCall(input)
		callPolicyRequireIdentityOnly(t, projection.Snapshot(), calendar.EventCallProjectionPaused)
		if err != nil {
			t.Fatalf("ProjectEventCall() error = %v", err)
		}
	})
	t.Run("resume uses existing inactive target", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Action = calendar.EventCallProjectionResume
		input.Target.Exists = true
		projection, err := calendar.ProjectEventCall(input)
		if err != nil || projection.Snapshot().Disposition != calendar.EventCallProjectionUpsert {
			t.Fatalf("resume result/error = %#v / %v", projection.Snapshot(), err)
		}
	})
	t.Run("delete requires a terminal canonical event", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Action = calendar.EventCallProjectionDelete
		input.Target.Exists, input.Target.Active = true, false
		input.Event.Status = calendar.CalendarEventStatusCancelled
		projection, err := calendar.ProjectEventCall(input)
		callPolicyRequireIdentityOnly(t, projection.Snapshot(), calendar.EventCallProjectionDeleted)
		if err != nil {
			t.Fatalf("ProjectEventCall() error = %v", err)
		}
	})
}

func TestProjectEventCall_ValidatesStatesAndBounds(t *testing.T) {
	tests := []struct {
		name string
		mutate func(*calendar.EventCallProjectionInput)
		want error
	}{
		{
			name: "non hub source",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Source = calendar.CalendarEventSourceMailImport },
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "deleted confirmed event",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Deleted = true },
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "target event identity mismatch",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Target.EventID = uuid.New() },
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "missing creator channel membership",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Creator.ActiveChannelMember = false },
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "pause on a new target",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Action = calendar.EventCallProjectionPause },
			want: calendar.ErrEventCallProjectionBoundary,
		},
		{
			name: "delete confirmed event",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Action = calendar.EventCallProjectionDelete
				input.Target.Exists = true
			},
			want: calendar.ErrEventCallProjectionBoundary,
		},
		{
			name: "title byte 256",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Title = strings.Repeat("a", 256) },
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "attendee count 201",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Attendees = make([]calendar.EventCallProjectionAttendee, 201)
			},
			want: calendar.ErrEventCallProjectionTooLarge,
		},
	} {
		t.Run(tests.name, func(t *testing.T) {
			input := callPolicyValidInput()
			tests.mutate(&input)
			projection, err := calendar.ProjectEventCall(input)
			if !errors.Is(err, tests.want) {
				t.Fatalf("ProjectEventCall() error = %v, want errors.Is(..., %v)", err, tests.want)
			}
			if got := projection.Snapshot(); !reflect.DeepEqual(got, calendar.EventCallProjectionSnapshot{}) {
				t.Fatalf("invalid projection = %#v, want zero", got)
			}
		})
	}
}

func TestProjectEventCall_ValidatesCanonicalTemporalEmailAndTimezoneBeforeProjection(t *testing.T) {
	tests := []struct {
		name string
		mutate func(*calendar.EventCallProjectionInput)
		want error
	}{
		{
			name: "master fixed zero offset location is not semantic UTC",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.Temporal.StartsAt = time.Date(2026, time.August, 20, 9, 0, 0, 0, time.FixedZone("UTC", 0))
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "master fractional second",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Temporal.EndsAt = callPolicyStart().Add(time.Hour + time.Nanosecond) },
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "moved exception non UTC instant",
			mutate: func(input *calendar.EventCallProjectionInput) {
				moved := input.Event.Temporal
				moved.StartsAt = time.Date(2026, time.August, 21, 9, 0, 0, 0, time.FixedZone("UTC", 0))
				moved.EndsAt = moved.StartsAt.Add(time.Hour)
				input.Exceptions = []calendar.EventCallProjectionException{{
					RecurrenceID: "2026-08-21T09:00:00Z", Temporal: &moved, Status: calendar.CalendarEventStatusConfirmed,
				}}
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "organizer email is not canonical",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.OrganizerEmail = "Owner@example.test"
				input.Creator.Email = "Owner@example.test"
				input.Attendees[0].Email = "Owner@example.test"
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "attendee email byte 321",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Attendees[1].Email = strings.Repeat("a", 309) + "@example.test" },
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "timezone byte 65",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Temporal.Timezone = strings.Repeat("a", 65) },
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "timezone control byte",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Temporal.Timezone = "UTC\x00" },
			want: calendar.ErrEventCallProjectionInvalid,
		},
	} {
		t.Run(tests.name, func(t *testing.T) {
			input := callPolicyValidInput()
			tests.mutate(&input)
			_, err := calendar.ProjectEventCall(input)
			if !errors.Is(err, tests.want) {
				t.Fatalf("ProjectEventCall() error = %v, want errors.Is(..., %v)", err, tests.want)
			}
		})
	}
}

func TestProjectEventCall_RecurrenceExceptionsAndOccurrenceAttendees(t *testing.T) {
	t.Run("exdate skips the nominal master", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Event.Exdates = []string{"2026-08-20T09:00:00Z"}
		projection, err := calendar.ProjectEventCall(input)
		if err != nil || projection.Snapshot().Occurrence == nil || projection.Snapshot().Occurrence.Key != "2026-08-21T09:00:00Z" {
			t.Fatalf("exdate result/error = %#v / %v", projection.Snapshot(), err)
		}
	})
	t.Run("cancelled exception removes its nominal occurrence", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Exceptions = []calendar.EventCallProjectionException{{
			RecurrenceID: "2026-08-20T09:00:00Z", Status: calendar.CalendarEventStatusCancelled, Cancelled: true,
		}}
		projection, err := calendar.ProjectEventCall(input)
		if err != nil || projection.Snapshot().Occurrence == nil || projection.Snapshot().Occurrence.Key != "2026-08-21T09:00:00Z" {
			t.Fatalf("cancelled result/error = %#v / %v", projection.Snapshot(), err)
		}
	})
	t.Run("nil temporal exception keeps the nominal interval", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Exceptions = []calendar.EventCallProjectionException{{
			RecurrenceID: "2026-08-20T09:00:00Z", Status: calendar.CalendarEventStatusTentative,
		}}
		projection, err := calendar.ProjectEventCall(input)
		if err != nil || projection.Snapshot().Occurrence == nil || !projection.Snapshot().Occurrence.StartsAt.Equal(callPolicyStart()) {
			t.Fatalf("nil temporal result/error = %#v / %v", projection.Snapshot(), err)
		}
	})
	t.Run("moved exception sorts on the actual start", func(t *testing.T) {
		input := callPolicyValidInput()
		moved := input.Event.Temporal
		moved.StartsAt = callPolicyStart().Add(-time.Hour)
		moved.EndsAt = moved.StartsAt.Add(30 * time.Minute)
		input.Exceptions = []calendar.EventCallProjectionException{{
			RecurrenceID: "2026-08-21T09:00:00Z", Temporal: &moved, Status: calendar.CalendarEventStatusConfirmed,
		}}
		input.Boundary = calendar.EventCallProjectionBoundary{At: callPolicyStart().Add(-2 * time.Hour), Inclusive: true}
		projection, err := calendar.ProjectEventCall(input)
		if err != nil || projection.Snapshot().Occurrence == nil || projection.Snapshot().Occurrence.Key != "2026-08-21T09:00:00Z" || !projection.Snapshot().Occurrence.StartsAt.Equal(moved.StartsAt) || projection.Snapshot().DurationMinutes != 30 {
			t.Fatalf("moved result/error = %#v / %v", projection.Snapshot(), err)
		}
	})
	t.Run("occurrence attendee is never a participant", func(t *testing.T) {
		input := callPolicyValidInput()
		id := callPolicyOccurrenceID
		input.Attendees = append(input.Attendees, calendar.EventCallProjectionAttendee{
			UserID: &id, Email: "occurrence@example.test", OrganizationID: callPolicyOrganizationID, ProjectID: callPolicyProjectID,
			Role: calendar.CalendarEventAttendeeRoleOptional, Status: calendar.RSVPAccepted, RecurrenceID: "2026-08-21T09:00:00Z", Active: true, ActiveProjectMember: true,
		})
		projection, err := calendar.ProjectEventCall(input)
		if err != nil || !reflect.DeepEqual(projection.Snapshot().ParticipantIDs, []uuid.UUID{callPolicyParticipantID}) {
			t.Fatalf("occurrence attendee result/error = %#v / %v", projection.Snapshot(), err)
		}
	})
	t.Run("foreign occurrence attendee recurrence id rejects", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Attendees[1].RecurrenceID = "2026-08-25T09:00:00Z"
		_, err := calendar.ProjectEventCall(input)
		if !errors.Is(err, calendar.ErrEventCallProjectionInvalid) {
			t.Fatalf("ProjectEventCall() error = %v", err)
		}
	})
	t.Run("one time event rejects occurrence attendees", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Event.RRULE = ""
		input.Attendees[1].RecurrenceID = "2026-08-20T09:00:00Z"
		_, err := calendar.ProjectEventCall(input)
		if !errors.Is(err, calendar.ErrEventCallProjectionInvalid) {
			t.Fatalf("ProjectEventCall() error = %v", err)
		}
	})
}

func TestProjectEventCall_SnapshotAndInputDoNotAlias(t *testing.T) {
	input := callPolicyValidInput()
	projection, err := calendar.ProjectEventCall(input)
	if err != nil {
		t.Fatalf("ProjectEventCall() error = %v", err)
	}
	first := projection.Snapshot()
	*first.ReminderMinutes = 99
	first.Occurrence.Key = "mutated"
	first.ParticipantIDs[0] = uuid.Nil
	input.Event.Title = "mutated input"
	*input.Event.ReminderMinutes = 98
	input.Event.ReminderMinutes = nil
	input.Attendees[1].Email = "mutated@example.test"

	second := projection.Snapshot()
	if second.Title != "Daily planning" || second.ReminderMinutes == nil || *second.ReminderMinutes != 15 ||
		second.Occurrence == nil || second.Occurrence.Key != "2026-08-20T09:00:00Z" ||
		!reflect.DeepEqual(second.ParticipantIDs, []uuid.UUID{callPolicyParticipantID}) {
		t.Fatalf("second Snapshot() retained an alias: %#v", second)
	}
}

func TestProjectEventCall_DoesNotAliasExceptionTemporal(t *testing.T) {
	input := callPolicyValidInput()
	moved := input.Event.Temporal
	moved.StartsAt = callPolicyStart().Add(-time.Hour)
	moved.EndsAt = moved.StartsAt.Add(30 * time.Minute)
	input.Exceptions = []calendar.EventCallProjectionException{{
		RecurrenceID: "2026-08-21T09:00:00Z", Temporal: &moved, Status: calendar.CalendarEventStatusConfirmed,
	}}
	input.Boundary = calendar.EventCallProjectionBoundary{At: callPolicyStart().Add(-2 * time.Hour), Inclusive: true}
	projection, err := calendar.ProjectEventCall(input)
	if err != nil {
		t.Fatalf("ProjectEventCall() error = %v", err)
	}
	moved.StartsAt = callPolicyStart().Add(10 * time.Hour)
	moved.EndsAt = moved.StartsAt.Add(time.Minute)
	input.Exceptions[0].Temporal = nil
	snapshot := projection.Snapshot()
	if snapshot.Occurrence == nil || snapshot.Occurrence.Key != "2026-08-21T09:00:00Z" || !snapshot.Occurrence.StartsAt.Equal(callPolicyStart().Add(-time.Hour)) || snapshot.DurationMinutes != 30 {
		t.Fatalf("projection retained an exception alias: %#v", snapshot)
	}
}

func TestProjectEventCall_IsDeterministicPureValuePolicy(t *testing.T) {
	input := callPolicyValidInput()
	changed := input.Event.Temporal
	changed.StartsAt = callPolicyStart().Add(25 * time.Hour)
	changed.EndsAt = changed.StartsAt.Add(30 * time.Minute)
	input.Exceptions = []calendar.EventCallProjectionException{{
		RecurrenceID: "2026-08-21T09:00:00Z", Temporal: &changed, Status: calendar.CalendarEventStatusTentative,
	}}
	before := cloneCallPolicyInput(input)
	first, firstErr := calendar.ProjectEventCall(input)
	second, secondErr := calendar.ProjectEventCall(input)
	if firstErr != nil || secondErr != nil || !reflect.DeepEqual(first.Snapshot(), second.Snapshot()) {
		t.Fatalf("pure repeated projection = %#v/%v then %#v/%v", first.Snapshot(), firstErr, second.Snapshot(), secondErr)
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("ProjectEventCall mutated its caller-owned input")
	}
}

func TestEventCallProjectionErrorsAreContentFreeAndTyped(t *testing.T) {
	for _, testCase := range []struct {
		name string
		input calendar.EventCallProjectionInput
		want error
		text string
		kind calendar.EventCallProjectionErrorKind
	}{
		{
			name: "invalid", input: func() calendar.EventCallProjectionInput { input := callPolicyValidInput(); input.Event.Title = ""; return input }(),
			want: calendar.ErrEventCallProjectionInvalid, text: "calendar: event call projection invalid", kind: calendar.EventCallProjectionInvalid,
		},
		{
			name: "boundary", input: func() calendar.EventCallProjectionInput { input := callPolicyValidInput(); input.Action = calendar.EventCallProjectionPause; return input }(),
			want: calendar.ErrEventCallProjectionBoundary, text: "calendar: event call projection boundary mismatch", kind: calendar.EventCallProjectionBoundary,
		},
		{
			name: "too large", input: func() calendar.EventCallProjectionInput { input := callPolicyValidInput(); input.Event.Title = strings.Repeat("x", 256); return input }(),
			want: calendar.ErrEventCallProjectionTooLarge, text: "calendar: event call projection too large", kind: calendar.EventCallProjectionTooLarge,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			_, err := calendar.ProjectEventCall(testCase.input)
			if !errors.Is(err, testCase.want) || err.Error() != testCase.text {
				t.Fatalf("error = %v, want %q and errors.Is(..., %v)", err, testCase.text, testCase.want)
			}
			typed, ok := err.(*calendar.EventCallProjectionError)
			if !ok || typed.Kind() != testCase.kind || typed.Unwrap() != testCase.want {
				t.Fatalf("typed error = %#v", err)
			}
		})
	}

	var nilError *calendar.EventCallProjectionError
	if nilError.Error() != "calendar: event call projection failed" || nilError.Kind() != "" || nilError.Unwrap() != nil || errors.Is(nilError, calendar.ErrEventCallProjectionInvalid) {
		t.Fatalf("nil error behavior = %q / %q / %v", nilError.Error(), nilError.Kind(), nilError.Unwrap())
	}
	unknown := &calendar.EventCallProjectionError{}
	if unknown.Error() != "calendar: event call projection failed" || unknown.Kind() != "" || unknown.Unwrap() != nil || errors.Is(unknown, calendar.ErrEventCallProjectionTooLarge) {
		t.Fatalf("unknown error behavior = %q / %q / %v", unknown.Error(), unknown.Kind(), unknown.Unwrap())
	}
}

func TestProjectEventCall_RecurrenceFormsAndDST(t *testing.T) {
	start := callPolicyStart()
	tests := []struct {
		name      string
		configure func(*calendar.EventCallProjectionInput)
		want      time.Time
	}{
		{
			name: "weekly count",
			configure: func(input *calendar.EventCallProjectionInput) {
				input.Event.RRULE = "FREQ=WEEKLY;INTERVAL=1;BYDAY=TH;COUNT=3"
				input.Boundary = calendar.EventCallProjectionBoundary{At: start, Inclusive: false}
			},
			want: start.AddDate(0, 0, 7),
		},
		{
			name: "monthly count",
			configure: func(input *calendar.EventCallProjectionInput) {
				input.Event.RRULE = "FREQ=MONTHLY;INTERVAL=1;BYMONTHDAY=20;COUNT=3"
				input.Boundary = calendar.EventCallProjectionBoundary{At: start, Inclusive: false}
			},
			want: time.Date(2026, time.September, 20, 9, 0, 0, 0, time.UTC),
		},
		{
			name: "until is inclusive",
			configure: func(input *calendar.EventCallProjectionInput) {
				input.Event.RRULE = "FREQ=DAILY;INTERVAL=1;UNTIL=20260821T090000Z"
				input.Boundary = calendar.EventCallProjectionBoundary{At: start.Add(24 * time.Hour), Inclusive: true}
			},
			want: start.Add(24 * time.Hour),
		},
		{
			name: "dst gap does not consume count",
			configure: func(input *calendar.EventCallProjectionInput) {
				gapStart := time.Date(2027, time.March, 13, 7, 30, 0, 0, time.UTC)
				input.Event.Temporal = calendar.CalendarEvent{StartsAt: gapStart, EndsAt: gapStart.Add(time.Hour), Timezone: "America/New_York"}
				input.Event.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=3"
				input.Boundary = calendar.EventCallProjectionBoundary{At: gapStart, Inclusive: false}
			},
			want: time.Date(2027, time.March, 15, 6, 30, 0, 0, time.UTC),
		},
		{
			name: "dst fold selects earliest instant",
			configure: func(input *calendar.EventCallProjectionInput) {
				foldStart := time.Date(2026, time.October, 31, 5, 30, 0, 0, time.UTC)
				input.Event.Temporal = calendar.CalendarEvent{StartsAt: foldStart, EndsAt: foldStart.Add(time.Hour), Timezone: "America/New_York"}
				input.Event.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=2"
				input.Boundary = calendar.EventCallProjectionBoundary{At: foldStart, Inclusive: false}
			},
			want: time.Date(2026, time.November, 1, 5, 30, 0, 0, time.UTC),
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := callPolicyValidInput()
			testCase.configure(&input)
			projection, err := calendar.ProjectEventCall(input)
			if err != nil || projection.Snapshot().Occurrence == nil || !projection.Snapshot().Occurrence.StartsAt.Equal(testCase.want) {
				t.Fatalf("ProjectEventCall() = %#v, %v; want occurrence at %s", projection.Snapshot(), err, testCase.want)
			}
		})
	}
}

func TestProjectEventCall_RecurrenceConflictMatrix(t *testing.T) {
	moved := func(input *calendar.EventCallProjectionInput, start time.Time, duration time.Duration) *calendar.CalendarEvent {
		temporal := input.Event.Temporal
		temporal.StartsAt = start
		temporal.EndsAt = start.Add(duration)
		return &temporal
	}
	tests := []struct {
		name   string
		mutate func(*calendar.EventCallProjectionInput)
		want   error
	}{
		{
			name: "exdates must be sorted",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.Exdates = []string{"2026-08-21T09:00:00Z", "2026-08-20T09:00:00Z"}
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "duplicate exdate",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.Exdates = []string{"2026-08-20T09:00:00Z", "2026-08-20T09:00:00Z"}
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "foreign exdate",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.Exdates = []string{"2026-08-25T09:00:00Z"}
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "exdate exception overlap",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.Exdates = []string{"2026-08-20T09:00:00Z"}
				input.Exceptions = []calendar.EventCallProjectionException{{RecurrenceID: "2026-08-20T09:00:00Z", Status: calendar.CalendarEventStatusConfirmed}}
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "duplicate exception key",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Exceptions = []calendar.EventCallProjectionException{
					{RecurrenceID: "2026-08-20T09:00:00Z", Status: calendar.CalendarEventStatusConfirmed},
					{RecurrenceID: "2026-08-20T09:00:00Z", Status: calendar.CalendarEventStatusTentative},
				}
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "foreign exception key",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Exceptions = []calendar.EventCallProjectionException{{RecurrenceID: "2026-08-25T09:00:00Z", Status: calendar.CalendarEventStatusConfirmed}}
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "cancelled false cannot have cancelled status",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Exceptions = []calendar.EventCallProjectionException{{RecurrenceID: "2026-08-20T09:00:00Z", Status: calendar.CalendarEventStatusCancelled}}
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "cancelled true requires cancelled status",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Exceptions = []calendar.EventCallProjectionException{{RecurrenceID: "2026-08-20T09:00:00Z", Status: calendar.CalendarEventStatusConfirmed, Cancelled: true}}
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "moved effective instant collision",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Exceptions = []calendar.EventCallProjectionException{{RecurrenceID: "2026-08-21T09:00:00Z", Temporal: moved(input, callPolicyStart(), time.Hour), Status: calendar.CalendarEventStatusConfirmed}}
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "moved duration over the cap",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Exceptions = []calendar.EventCallProjectionException{{RecurrenceID: "2026-08-21T09:00:00Z", Temporal: moved(input, callPolicyStart().Add(24*time.Hour), 1441*time.Minute), Status: calendar.CalendarEventStatusConfirmed}}
			},
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "one time event forbids exception",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.RRULE = ""
				input.Exceptions = []calendar.EventCallProjectionException{{RecurrenceID: "2026-08-20T09:00:00Z", Status: calendar.CalendarEventStatusConfirmed}}
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := callPolicyValidInput()
			testCase.mutate(&input)
			projection, err := calendar.ProjectEventCall(input)
			if !errors.Is(err, testCase.want) || !reflect.DeepEqual(projection.Snapshot(), calendar.EventCallProjectionSnapshot{}) {
				t.Fatalf("ProjectEventCall() = %#v, %v; want zero and %v", projection.Snapshot(), err, testCase.want)
			}
		})
	}
}

func TestProjectEventCall_TargetActionMatrix(t *testing.T) {
	tests := []struct {
		name        string
		mutate      func(*calendar.EventCallProjectionInput)
		disposition calendar.EventCallProjectionDisposition
		want        error
	}{
		{name: "reconcile existing active", mutate: func(input *calendar.EventCallProjectionInput) { input.Target.Exists, input.Target.Active = true, true }, disposition: calendar.EventCallProjectionUpsert},
		{name: "reconcile inactive is boundary", mutate: func(input *calendar.EventCallProjectionInput) { input.Target.Exists = true }, want: calendar.ErrEventCallProjectionBoundary},
		{name: "reconcile deleted is boundary", mutate: func(input *calendar.EventCallProjectionInput) { input.Target.Exists, input.Target.Deleted = true, true }, want: calendar.ErrEventCallProjectionBoundary},
		{name: "pause existing active", mutate: func(input *calendar.EventCallProjectionInput) { input.Action = calendar.EventCallProjectionPause; input.Target.Exists, input.Target.Active = true, true }, disposition: calendar.EventCallProjectionPaused},
		{name: "pause inactive is boundary", mutate: func(input *calendar.EventCallProjectionInput) { input.Action = calendar.EventCallProjectionPause; input.Target.Exists = true }, want: calendar.ErrEventCallProjectionBoundary},
		{name: "resume existing inactive", mutate: func(input *calendar.EventCallProjectionInput) { input.Action = calendar.EventCallProjectionResume; input.Target.Exists = true }, disposition: calendar.EventCallProjectionUpsert},
		{name: "resume active is boundary", mutate: func(input *calendar.EventCallProjectionInput) { input.Action = calendar.EventCallProjectionResume; input.Target.Exists, input.Target.Active = true, true }, want: calendar.ErrEventCallProjectionBoundary},
		{name: "delete cancelled canonical", mutate: func(input *calendar.EventCallProjectionInput) { input.Action = calendar.EventCallProjectionDelete; input.Target.Exists = true; input.Event.Status = calendar.CalendarEventStatusCancelled }, disposition: calendar.EventCallProjectionDeleted},
		{name: "delete deleted canonical", mutate: func(input *calendar.EventCallProjectionInput) { input.Action = calendar.EventCallProjectionDelete; input.Target.Exists = true; input.Event.Status = calendar.CalendarEventStatusCancelled; input.Event.Deleted = true }, disposition: calendar.EventCallProjectionDeleted},
		{name: "delete deleted target is boundary", mutate: func(input *calendar.EventCallProjectionInput) { input.Action = calendar.EventCallProjectionDelete; input.Target.Exists, input.Target.Deleted = true, true; input.Event.Status = calendar.CalendarEventStatusCancelled }, want: calendar.ErrEventCallProjectionBoundary},
		{name: "delete confirmed event is boundary", mutate: func(input *calendar.EventCallProjectionInput) { input.Action = calendar.EventCallProjectionDelete; input.Target.Exists = true }, want: calendar.ErrEventCallProjectionBoundary},
		{name: "unknown action is invalid", mutate: func(input *calendar.EventCallProjectionInput) { input.Action = "unknown" }, want: calendar.ErrEventCallProjectionInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := callPolicyValidInput()
			testCase.mutate(&input)
			projection, err := calendar.ProjectEventCall(input)
			if testCase.want != nil {
				if !errors.Is(err, testCase.want) {
					t.Fatalf("ProjectEventCall() error = %v, want %v", err, testCase.want)
				}
				return
			}
			if err != nil || projection.Snapshot().Disposition != testCase.disposition {
				t.Fatalf("ProjectEventCall() = %#v, %v; want %s", projection.Snapshot(), err, testCase.disposition)
			}
		})
	}
}

func TestProjectEventCall_PreflightCapsAndTemporalLimits(t *testing.T) {
	tests := []struct {
		name     string
		mutate   func(*calendar.EventCallProjectionInput)
		want     error
		duration int
	}{
		{
			name: "master duration exactly one minute is accepted",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.Temporal.EndsAt = input.Event.Temporal.StartsAt.Add(time.Minute)
			},
			duration: 1,
		},
		{
			name: "master duration exactly 1440 is accepted",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.Temporal.EndsAt = input.Event.Temporal.StartsAt.Add(1440 * time.Minute)
			},
			duration: 1440,
		},
		{
			name: "zero duration is invalid",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.Temporal.EndsAt = input.Event.Temporal.StartsAt
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "master duration over cap is too large",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.Temporal.EndsAt = input.Event.Temporal.StartsAt.Add(1441 * time.Minute)
			},
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "positive duration arithmetic overflow is too large",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.Temporal.EndsAt = input.Event.Temporal.StartsAt.AddDate(400, 0, 0)
			},
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "reminder over cap is too large",
			mutate: func(input *calendar.EventCallProjectionInput) {
				value := 40321
				input.Event.ReminderMinutes = &value
			},
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "count over nominal cap is too large",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=1001"
			},
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "until over nominal cap is too large",
			mutate: func(input *calendar.EventCallProjectionInput) {
				until := input.Event.Temporal.StartsAt.AddDate(0, 0, 1000).Format("20060102T150405Z")
				input.Event.RRULE = "FREQ=DAILY;INTERVAL=1;UNTIL=" + until
			},
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "exdate key over byte cap is too large before expansion",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.Exdates = []string{strings.Repeat("x", 21)}
			},
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "exception key over byte cap is too large before expansion",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Exceptions = []calendar.EventCallProjectionException{{RecurrenceID: strings.Repeat("x", 21), Status: calendar.CalendarEventStatusConfirmed}}
			},
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "attendee occurrence key over byte cap is too large before expansion",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Attendees[1].RecurrenceID = strings.Repeat("x", 21)
			},
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "year 10000 is not a literal utc key",
			mutate: func(input *calendar.EventCallProjectionInput) {
				start := time.Date(10000, time.January, 1, 0, 0, 0, 0, time.UTC)
				input.Event.Temporal = calendar.CalendarEvent{StartsAt: start, EndsAt: start.Add(time.Hour), Timezone: "UTC"}
				input.Event.RRULE = ""
				input.Boundary = calendar.EventCallProjectionBoundary{At: start, Inclusive: true}
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := callPolicyValidInput()
			testCase.mutate(&input)
			projection, err := calendar.ProjectEventCall(input)
			if testCase.want != nil {
				if !errors.Is(err, testCase.want) {
					t.Fatalf("ProjectEventCall() error = %v, want %v", err, testCase.want)
				}
				return
			}
			if err != nil || projection.Snapshot().DurationMinutes != testCase.duration {
				t.Fatalf("ProjectEventCall() = %#v, %v; want duration %d", projection.Snapshot(), err, testCase.duration)
			}
		})
	}

	t.Run("year zero is an exact twenty byte UTC key", func(t *testing.T) {
		input := callPolicyValidInput()
		start := time.Date(0, time.January, 2, 3, 4, 5, 0, time.UTC)
		input.Event.Temporal = calendar.CalendarEvent{StartsAt: start, EndsAt: start.Add(time.Minute), Timezone: "UTC"}
		input.Event.RRULE = ""
		input.Boundary = calendar.EventCallProjectionBoundary{At: start, Inclusive: true}
		projection, err := calendar.ProjectEventCall(input)
		if err != nil || projection.Snapshot().Occurrence == nil || projection.Snapshot().Occurrence.Key != "0000-01-02T03:04:05Z" {
			t.Fatalf("ProjectEventCall() = %#v, %v", projection.Snapshot(), err)
		}
	})
}

func TestProjectEventCall_RecurrenceYearAndUntilClassification(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*calendar.EventCallProjectionInput)
		want   error
	}{
		{
			name: "generated recurrence crossing year 9999 is too large",
			mutate: func(input *calendar.EventCallProjectionInput) {
				start := time.Date(9999, time.December, 31, 9, 0, 0, 0, time.UTC)
				input.Event.Temporal = calendar.CalendarEvent{StartsAt: start, EndsAt: start.Add(time.Hour), Timezone: "UTC"}
				input.Event.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=2"
				input.Boundary = calendar.EventCallProjectionBoundary{At: start, Inclusive: true}
			},
			want: calendar.ErrEventCallProjectionTooLarge,
		},
		{
			name: "until before dtstart is invalid rather than too large",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.RRULE = "FREQ=DAILY;INTERVAL=1;UNTIL=20260819T090000Z"
			},
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "until equal dtstart is a valid one occurrence series",
			mutate: func(input *calendar.EventCallProjectionInput) {
				input.Event.RRULE = "FREQ=DAILY;INTERVAL=1;UNTIL=20260820T090000Z"
			},
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := callPolicyValidInput()
			testCase.mutate(&input)
			projection, err := calendar.ProjectEventCall(input)
			if testCase.want != nil {
				if !errors.Is(err, testCase.want) {
					t.Fatalf("ProjectEventCall() error = %v, want %v", err, testCase.want)
				}
				return
			}
			if err != nil || projection.Snapshot().Occurrence == nil || projection.Snapshot().Occurrence.Key != "2026-08-20T09:00:00Z" {
				t.Fatalf("ProjectEventCall() = %#v, %v", projection.Snapshot(), err)
			}
		})
	}
}

func TestProjectEventCall_AttendeeAuthorizationMatrix(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*calendar.EventCallProjectionInput)
		want   error
	}{
		{
			name: "email only attendee is invalid",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Attendees[1].UserID = nil },
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "foreign organization attendee is invalid",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Attendees[1].OrganizationID = uuid.New() },
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "inactive attendee is invalid",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Attendees[1].Active = false },
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "non member attendee is invalid",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Attendees[1].ActiveProjectMember = false },
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "duplicate attendee email is invalid",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Attendees[1].Email = "owner@example.test" },
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "duplicate attendee user is invalid",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Attendees[1].UserID = input.Attendees[0].UserID },
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "unknown attendee role is invalid",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Attendees[1].Role = "unknown" },
			want: calendar.ErrEventCallProjectionInvalid,
		},
		{
			name: "unknown attendee status is invalid",
			mutate: func(input *calendar.EventCallProjectionInput) { input.Attendees[1].Status = "unknown" },
			want: calendar.ErrEventCallProjectionInvalid,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := callPolicyValidInput()
			testCase.mutate(&input)
			_, err := calendar.ProjectEventCall(input)
			if !errors.Is(err, testCase.want) {
				t.Fatalf("ProjectEventCall() error = %v, want %v", err, testCase.want)
			}
		})
	}

	t.Run("declined series attendee remains valid but does not project", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Attendees[1].Status = calendar.RSVPDeclined
		projection, err := calendar.ProjectEventCall(input)
		if err != nil || len(projection.Snapshot().ParticipantIDs) != 0 {
			t.Fatalf("ProjectEventCall() = %#v, %v", projection.Snapshot(), err)
		}
	})
}

func TestProjectEventCall_IdentityTargetEventAndCreatorMatrix(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*calendar.EventCallProjectionInput)
	}{
		{name: "target scheduled call mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Target.ScheduledCallID = uuid.New() }},
		{name: "target event mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Target.EventID = uuid.New() }},
		{name: "target creator mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Target.CreatorID = uuid.New() }},
		{name: "target organization mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Target.OrganizationID = uuid.New() }},
		{name: "target project mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Target.ProjectID = uuid.New() }},
		{name: "target channel mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Target.ChannelID = uuid.New() }},
		{name: "event identifier mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.EventID = uuid.New() }},
		{name: "event scheduled call mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.ScheduledCallID = uuid.New() }},
		{name: "event organization mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.OrganizationID = uuid.New() }},
		{name: "event project mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.ProjectID = uuid.New() }},
		{name: "event channel mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.ChannelID = uuid.New() }},
		{name: "event organizer mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.OrganizerID = uuid.New() }},
		{name: "event organizer email mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.OrganizerEmail = "other@example.test" }},
		{name: "creator user mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Creator.UserID = uuid.New() }},
		{name: "creator email mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Creator.Email = "other@example.test" }},
		{name: "creator organization mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Creator.OrganizationID = uuid.New() }},
		{name: "creator project mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Creator.ProjectID = uuid.New() }},
		{name: "creator channel mismatch", mutate: func(input *calendar.EventCallProjectionInput) { input.Creator.ChannelID = uuid.New() }},
		{name: "new target cannot be active", mutate: func(input *calendar.EventCallProjectionInput) { input.Target.Active = true }},
		{name: "new target cannot be deleted", mutate: func(input *calendar.EventCallProjectionInput) { input.Target.Deleted = true }},
		{name: "deleted target cannot be active", mutate: func(input *calendar.EventCallProjectionInput) { input.Target.Exists, input.Target.Deleted, input.Target.Active = true, true, true }},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := callPolicyValidInput()
			testCase.mutate(&input)
			projection, err := calendar.ProjectEventCall(input)
			if !errors.Is(err, calendar.ErrEventCallProjectionInvalid) || !reflect.DeepEqual(projection.Snapshot(), calendar.EventCallProjectionSnapshot{}) {
				t.Fatalf("ProjectEventCall() = %#v, %v; want invalid zero projection", projection.Snapshot(), err)
			}
		})
	}
}

func TestProjectEventCall_SourceStatusBoundaryAndExactEdges(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*calendar.EventCallProjectionInput)
		want   error
	}{
		{name: "ical source", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Source = calendar.CalendarEventSourceICalFeed }, want: calendar.ErrEventCallProjectionInvalid},
		{name: "unknown source", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Source = "unknown" }, want: calendar.ErrEventCallProjectionInvalid},
		{name: "confirmed deleted", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Deleted = true }, want: calendar.ErrEventCallProjectionInvalid},
		{name: "tentative deleted", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Status = calendar.CalendarEventStatusTentative; input.Event.Deleted = true }, want: calendar.ErrEventCallProjectionInvalid},
		{name: "unknown event status", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Status = "unknown" }, want: calendar.ErrEventCallProjectionInvalid},
		{name: "zero boundary", mutate: func(input *calendar.EventCallProjectionInput) { input.Boundary.At = time.Time{} }, want: calendar.ErrEventCallProjectionInvalid},
		{name: "fixed zone boundary", mutate: func(input *calendar.EventCallProjectionInput) { input.Boundary.At = time.Date(2026, time.August, 20, 9, 0, 0, 0, time.FixedZone("UTC", 0)) }, want: calendar.ErrEventCallProjectionInvalid},
		{name: "fractional boundary", mutate: func(input *calendar.EventCallProjectionInput) { input.Boundary.At = callPolicyStart().Add(time.Nanosecond) }, want: calendar.ErrEventCallProjectionInvalid},
		{name: "title exact cap", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Title = strings.Repeat("t", 255) }},
		{name: "description exact cap", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Description = strings.Repeat("d", 2000) }},
		{name: "description over cap", mutate: func(input *calendar.EventCallProjectionInput) { input.Event.Description = strings.Repeat("d", 2001) }, want: calendar.ErrEventCallProjectionTooLarge},
		{name: "reminder zero", mutate: func(input *calendar.EventCallProjectionInput) { value := 0; input.Event.ReminderMinutes = &value }},
		{name: "reminder exact cap", mutate: func(input *calendar.EventCallProjectionInput) { value := 40320; input.Event.ReminderMinutes = &value }},
		{name: "negative reminder", mutate: func(input *calendar.EventCallProjectionInput) { value := -1; input.Event.ReminderMinutes = &value }, want: calendar.ErrEventCallProjectionInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := callPolicyValidInput()
			testCase.mutate(&input)
			projection, err := calendar.ProjectEventCall(input)
			if testCase.want != nil {
				if !errors.Is(err, testCase.want) {
					t.Fatalf("ProjectEventCall() error = %v, want %v", err, testCase.want)
				}
				return
			}
			if err != nil || projection.Snapshot().Disposition != calendar.EventCallProjectionUpsert {
				t.Fatalf("ProjectEventCall() = %#v, %v; want upsert", projection.Snapshot(), err)
			}
		})
	}
}

func TestProjectEventCall_OrganizerProjectAndParticipantRows(t *testing.T) {
	t.Run("zero series organizers is invalid", func(t *testing.T) {
		input := callPolicyValidInput()
		otherID := uuid.New()
		input.Attendees[0].UserID = &otherID
		input.Attendees[0].Email = "other@example.test"
		_, err := calendar.ProjectEventCall(input)
		if !errors.Is(err, calendar.ErrEventCallProjectionInvalid) {
			t.Fatalf("ProjectEventCall() error = %v", err)
		}
	})
	t.Run("two series organizers is invalid", func(t *testing.T) {
		input := callPolicyValidInput()
		creatorID := callPolicyCreatorID
		input.Attendees = append(input.Attendees, calendar.EventCallProjectionAttendee{
			UserID: &creatorID, Email: "owner@example.test", OrganizationID: callPolicyOrganizationID, ProjectID: callPolicyProjectID,
			Role: calendar.CalendarEventAttendeeRoleRequired, Status: calendar.RSVPAccepted, Active: true, ActiveProjectMember: true,
		})
		_, err := calendar.ProjectEventCall(input)
		if !errors.Is(err, calendar.ErrEventCallProjectionInvalid) {
			t.Fatalf("ProjectEventCall() error = %v", err)
		}
	})
	t.Run("cross project attendee is invalid", func(t *testing.T) {
		input := callPolicyValidInput()
		input.Attendees[1].ProjectID = uuid.New()
		_, err := calendar.ProjectEventCall(input)
		if !errors.Is(err, calendar.ErrEventCallProjectionInvalid) {
			t.Fatalf("ProjectEventCall() error = %v", err)
		}
	})
	for _, status := range []calendar.RSVPStatus{calendar.RSVPAccepted, calendar.RSVPTentative} {
		t.Run("participant "+string(status), func(t *testing.T) {
			input := callPolicyValidInput()
			input.Attendees[1].Status = status
			projection, err := calendar.ProjectEventCall(input)
			if err != nil || !reflect.DeepEqual(projection.Snapshot().ParticipantIDs, []uuid.UUID{callPolicyParticipantID}) {
				t.Fatalf("ProjectEventCall() = %#v, %v", projection.Snapshot(), err)
			}
		})
	}
}

func cloneCallPolicyInput(input calendar.EventCallProjectionInput) calendar.EventCallProjectionInput {
	clone := input
	clone.Event.Exdates = append([]string(nil), input.Event.Exdates...)
	if input.Event.ReminderMinutes != nil {
		value := *input.Event.ReminderMinutes
		clone.Event.ReminderMinutes = &value
	}
	clone.Attendees = make([]calendar.EventCallProjectionAttendee, len(input.Attendees))
	for index, attendee := range input.Attendees {
		clone.Attendees[index] = attendee
		if attendee.UserID != nil {
			identifier := *attendee.UserID
			clone.Attendees[index].UserID = &identifier
		}
	}
	clone.Exceptions = make([]calendar.EventCallProjectionException, len(input.Exceptions))
	for index, exception := range input.Exceptions {
		clone.Exceptions[index] = exception
		if exception.Temporal != nil {
			temporal := *exception.Temporal
			clone.Exceptions[index].Temporal = &temporal
		}
	}
	return clone
}

func callPolicyValidInput() calendar.EventCallProjectionInput {
	reminder := 15
	creatorID := callPolicyCreatorID
	participantID := callPolicyParticipantID
	return calendar.EventCallProjectionInput{
		Action: calendar.EventCallProjectionReconcile,
		Event: calendar.EventCallProjectionEvent{
			EventID: callPolicyEventID, ScheduledCallID: callPolicyScheduledID,
			OrganizationID: callPolicyOrganizationID, ProjectID: callPolicyProjectID, ChannelID: callPolicyChannelID,
			OrganizerID: callPolicyCreatorID, OrganizerEmail: "owner@example.test", Source: calendar.CalendarEventSourceHub,
			Status: calendar.CalendarEventStatusConfirmed, Title: "Daily planning", Description: "Discuss the next delivery.",
			Temporal: calendar.CalendarEvent{StartsAt: callPolicyStart(), EndsAt: callPolicyStart().Add(time.Hour), Timezone: "UTC"},
			RRULE: "FREQ=DAILY;INTERVAL=1;COUNT=3", ReminderMinutes: &reminder,
		},
		Target: calendar.EventCallProjectionTarget{
			ScheduledCallID: callPolicyScheduledID, EventID: callPolicyEventID, CreatorID: callPolicyCreatorID,
			OrganizationID: callPolicyOrganizationID, ProjectID: callPolicyProjectID, ChannelID: callPolicyChannelID,
		},
		Creator: calendar.EventCallProjectionCreator{
			UserID: callPolicyCreatorID, Email: "owner@example.test", OrganizationID: callPolicyOrganizationID,
			ProjectID: callPolicyProjectID, ChannelID: callPolicyChannelID, Active: true, ActiveProjectMember: true, ActiveChannelMember: true,
		},
		Attendees: []calendar.EventCallProjectionAttendee{
			{UserID: &creatorID, Email: "owner@example.test", OrganizationID: callPolicyOrganizationID, ProjectID: callPolicyProjectID, Role: calendar.CalendarEventAttendeeRoleRequired, Status: calendar.RSVPAccepted, Active: true, ActiveProjectMember: true},
			{UserID: &participantID, Email: "participant@example.test", OrganizationID: callPolicyOrganizationID, ProjectID: callPolicyProjectID, Role: calendar.CalendarEventAttendeeRoleRequired, Status: calendar.RSVPNeedsAction, Active: true, ActiveProjectMember: true},
		},
		Boundary: calendar.EventCallProjectionBoundary{At: callPolicyStart(), Inclusive: true},
	}
}

func callPolicyStart() time.Time {
	return time.Date(2026, time.August, 20, 9, 0, 0, 0, time.UTC)
}

func callPolicyRequireIdentityOnly(t *testing.T, got calendar.EventCallProjectionSnapshot, disposition calendar.EventCallProjectionDisposition) {
	t.Helper()
	want := calendar.EventCallProjectionSnapshot{
		Disposition: disposition, ScheduledCallID: callPolicyScheduledID, EventID: callPolicyEventID,
		OrganizationID: callPolicyOrganizationID, ProjectID: callPolicyProjectID, ChannelID: callPolicyChannelID, CreatorID: callPolicyCreatorID,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("identity-only Snapshot() = %#v, want %#v", got, want)
	}
}
