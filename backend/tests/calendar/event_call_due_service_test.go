package calendar_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/calendar"
)

var (
	dueCalendarEventID = uuid.MustParse("00112233-4455-6677-8899-aabbccddeeff")
	dueScheduledCallID = uuid.MustParse("10213243-5465-7687-8899-aabbccddeeff")
	dueEnvelopeEventID = uuid.MustParse("20314253-6475-8697-a8b9-cadbecfd0e1f")
	dueOrganizationID  = uuid.MustParse("30415263-7485-8697-a8b9-cadbecfd0e1f")
	dueProjectID       = uuid.MustParse("40516273-8495-a6b7-c8d9-eafbec0d1e2f")
	dueChannelID       = uuid.MustParse("50617283-94a5-b6c7-d8e9-fa0b1c2d3e4f")
	dueCreatorID       = uuid.MustParse("60718293-a4b5-c6d7-e8f9-0a1b2c3d4e5f")
	dueParticipantID   = uuid.MustParse("708192a3-b4c5-d6e7-f809-1a2b3c4d5e6f")
)

const (
	dueNominalKey = "2026-08-21T09:00:00Z"
	dueKey        = "02ccb08c267ddd7eb7049ea6ab52d72bbe3cdecfac0b61758119aaa81316d0ab"
)

func TestEventCallDueABI_IsFrozen(t *testing.T) {
	assertDueStruct(t, reflect.TypeOf(calendar.EventCallDueCallSnapshot{}), []dueField{
		{"ScheduledCallID", reflect.TypeOf(uuid.UUID{}), ""}, {"CalendarEventID", reflect.TypeOf(uuid.UUID{}), ""},
		{"OrganizationID", reflect.TypeOf(uuid.UUID{}), ""}, {"ProjectID", reflect.TypeOf(uuid.UUID{}), ""},
		{"ChannelID", reflect.TypeOf(uuid.UUID{}), ""}, {"CreatorID", reflect.TypeOf(uuid.UUID{}), ""},
		{"ScheduledAt", reflect.TypeOf(time.Time{}), ""}, {"Active", reflect.TypeOf(false), ""}, {"Deleted", reflect.TypeOf(false), ""},
	})
	assertDueStruct(t, reflect.TypeOf(calendar.EventCallDueEnvelopeInput{}), []dueField{
		{"EventID", reflect.TypeOf(uuid.UUID{}), ""}, {"OccurredAt", reflect.TypeOf(time.Time{}), ""}, {"CorrelationID", reflect.TypeOf(""), ""},
	})
	assertDueStruct(t, reflect.TypeOf(calendar.EventCallDueProjectionInput{}), []dueField{
		{"Event", reflect.TypeOf(calendar.EventCallProjectionEvent{}), ""}, {"Target", reflect.TypeOf(calendar.EventCallProjectionTarget{}), ""},
		{"Creator", reflect.TypeOf(calendar.EventCallProjectionCreator{}), ""}, {"Attendees", reflect.TypeOf([]calendar.EventCallProjectionAttendee{}), ""},
		{"Exceptions", reflect.TypeOf([]calendar.EventCallProjectionException{}), ""},
	})
	assertDueStruct(t, reflect.TypeOf(calendar.EventCallDueInput{}), []dueField{
		{"Call", reflect.TypeOf(calendar.EventCallDueCallSnapshot{}), ""}, {"Now", reflect.TypeOf(time.Time{}), ""},
		{"Projection", reflect.TypeOf((*calendar.EventCallDueProjectionInput)(nil)), ""}, {"Envelope", reflect.TypeOf(calendar.EventCallDueEnvelopeInput{}), ""},
	})
	assertDueStruct(t, reflect.TypeOf(calendar.EventCallScheduledDuePayload{}), []dueField{
		{"ScheduledCallID", reflect.TypeOf(uuid.UUID{}), "json:\"scheduled_call_id\""}, {"CalendarEventID", reflect.TypeOf(uuid.UUID{}), "json:\"calendar_event_id\""},
		{"OccurrenceKey", reflect.TypeOf(""), "json:\"occurrence_key\""}, {"RecurrenceID", reflect.TypeOf((*string)(nil)), "json:\"recurrence_id\""},
		{"ChannelID", reflect.TypeOf(uuid.UUID{}), "json:\"channel_id\""}, {"ProjectID", reflect.TypeOf(uuid.UUID{}), "json:\"project_id\""},
		{"Title", reflect.TypeOf(""), "json:\"title\""}, {"ScheduledAt", reflect.TypeOf(time.Time{}), "json:\"scheduled_at\""},
		{"DurationMinutes", reflect.TypeOf(int(0)), "json:\"duration_minutes\""},
	})
	assertDueStruct(t, reflect.TypeOf(calendar.EventCallDueSnapshot{}), []dueField{
		{"Disposition", reflect.TypeOf(calendar.EventCallDueDisposition("")), ""}, {"ScheduledCallID", reflect.TypeOf(uuid.UUID{}), ""},
		{"CalendarEventID", reflect.TypeOf(uuid.UUID{}), ""}, {"OrganizationID", reflect.TypeOf(uuid.UUID{}), ""},
		{"ProjectID", reflect.TypeOf(uuid.UUID{}), ""}, {"ChannelID", reflect.TypeOf(uuid.UUID{}), ""}, {"CreatorID", reflect.TypeOf(uuid.UUID{}), ""},
		{"OccurrenceKey", reflect.TypeOf(""), ""}, {"ScheduledAt", reflect.TypeOf((*time.Time)(nil)), ""}, {"Active", reflect.TypeOf(false), ""},
	})

	decisionType := reflect.TypeOf(calendar.EventCallDueDecision{})
	errorType := reflect.TypeOf((*calendar.EventCallDueError)(nil))
	snapshotMethod, snapshotFound := decisionType.MethodByName("Snapshot")
	outboxMethod, outboxFound := decisionType.MethodByName("Outbox")
	errorMethod, errorFound := errorType.MethodByName("Error")
	isMethod, isFound := errorType.MethodByName("Is")
	kindMethod, kindFound := errorType.MethodByName("Kind")
	unwrapMethod, unwrapFound := errorType.MethodByName("Unwrap")
	if !snapshotFound || !outboxFound || !errorFound || !isFound || !kindFound || !unwrapFound {
		t.Fatal("frozen exported method is absent")
	}
	for _, check := range []struct {
		name string
		value any
		want  string
	}{
		{"Snapshot", snapshotMethod.Type, "func(calendar.EventCallDueDecision) calendar.EventCallDueSnapshot"},
		{"Outbox", outboxMethod.Type, "func(calendar.EventCallDueDecision) (calendar.EventOutboxDescriptor, bool)"},
		{"Error", errorMethod.Type, "func(*calendar.EventCallDueError) string"},
		{"Is", isMethod.Type, "func(*calendar.EventCallDueError, error) bool"},
		{"Kind", kindMethod.Type, "func(*calendar.EventCallDueError) calendar.EventCallDueErrorKind"},
		{"Unwrap", unwrapMethod.Type, "func(*calendar.EventCallDueError) error"},
		{"ProjectEventCallDue", reflect.TypeOf(calendar.ProjectEventCallDue), "func(calendar.EventCallDueInput) (calendar.EventCallDueDecision, error)"},
	} {
		if got := check.value.(reflect.Type).String(); got != check.want {
			t.Fatalf("%s type = %s, want %s", check.name, got, check.want)
		}
	}
	if got := reflect.TypeOf(calendar.EventCallDueDecision{}).NumField(); got != 2 {
		t.Fatalf("opaque decision field count = %d, want 2", got)
	}
	for index := 0; index < reflect.TypeOf(calendar.EventCallDueDecision{}).NumField(); index++ {
		if field := reflect.TypeOf(calendar.EventCallDueDecision{}).Field(index); field.PkgPath == "" {
			t.Fatalf("opaque decision exposes %q", field.Name)
		}
	}
	if calendar.EventCallDueNotDue != "not_due" || calendar.EventCallDueReschedule != "reschedule" || calendar.EventCallDueAdvance != "advance" || calendar.EventCallDueDeactivate != "deactivate" || calendar.EventCallDueNoop != "noop" {
		t.Fatal("due disposition literals changed")
	}
	if calendar.EventCallDueInvalid != "invalid" || calendar.EventCallDueBoundary != "boundary" || calendar.EventCallDueTooLarge != "too_large" {
		t.Fatal("due error kind literals changed")
	}
}

func TestProjectEventCallDue_AdvanceBuildsExactDescriptor(t *testing.T) {
	input := dueValidInput()
	decision, err := calendar.ProjectEventCallDue(input)
	if err != nil {
		t.Fatalf("ProjectEventCallDue() error = %v", err)
	}
	assertDueSnapshot(t, decision.Snapshot(), calendar.EventCallDueAdvance, true, dueStart().AddDate(0, 0, 1), dueKey)
	descriptor, ok := decision.Outbox()
	if !ok {
		t.Fatal("Outbox() absent for due advance")
	}
	envelope := descriptor.Envelope()
	if got := envelope.Subject(); got != "calls.channel.50617283-94a5-b6c7-d8e9-fa0b1c2d3e4f.scheduled_due.v1" {
		t.Fatalf("subject = %q", got)
	}
	if got := envelope.EventID(); got != dueEnvelopeEventID {
		t.Fatalf("envelope EventID() = %s", got)
	}
	encoded, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("json.Marshal(envelope) error = %v", err)
	}
	var complete struct {
		Subject        string          `json:"subject"`
		SchemaVersion  int             `json:"schema_version"`
		EventID        uuid.UUID       `json:"event_id"`
		OccurredAt     time.Time       `json:"occurred_at"`
		OrganizationID uuid.UUID       `json:"organization_id"`
		ActorID        *uuid.UUID      `json:"actor_id"`
		CorrelationID  string          `json:"correlation_id"`
		Payload        json.RawMessage `json:"payload"`
	}
	if err := json.Unmarshal(encoded, &complete); err != nil {
		t.Fatalf("json.Unmarshal(envelope) error = %v", err)
	}
	var outer map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &outer); err != nil || len(outer) != 8 {
		t.Fatalf("outer envelope = %#v, %v; want exactly eight fields", outer, err)
	}
	if complete.Subject != envelope.Subject() || complete.SchemaVersion != 1 || complete.EventID != dueEnvelopeEventID || !complete.OccurredAt.Equal(dueStart()) || complete.OrganizationID != dueOrganizationID || complete.ActorID != nil || complete.CorrelationID != "calendar-due-04k" {
		t.Fatalf("envelope binding = %#v", complete)
	}
	var payload map[string]json.RawMessage
	if err := json.Unmarshal(complete.Payload, &payload); err != nil {
		t.Fatalf("json.Unmarshal(payload) error = %v", err)
	}
	if len(payload) != 9 {
		t.Fatalf("payload field count = %d, want 9: %s", len(payload), complete.Payload)
	}
	for field, want := range map[string]string{
		"scheduled_call_id": `"10213243-5465-7687-8899-aabbccddeeff"`, "calendar_event_id": `"00112233-4455-6677-8899-aabbccddeeff"`,
		"occurrence_key": `"` + dueKey + `"`, "recurrence_id": `"2026-08-21T09:00:00Z"`,
		"channel_id": `"50617283-94a5-b6c7-d8e9-fa0b1c2d3e4f"`, "project_id": `"40516273-8495-a6b7-c8d9-eafbec0d1e2f"`,
		"title": `"Daily planning"`, "scheduled_at": `"2026-08-21T09:00:00Z"`, "duration_minutes": `60`,
	} {
		if got := string(payload[field]); got != want {
			t.Fatalf("payload.%s = %s, want %s", field, got, want)
		}
	}
	copyOne, ok := decision.Outbox()
	if !ok {
		t.Fatal("second Outbox() absent")
	}
	copyBytes, _ := json.Marshal(copyOne.Envelope())
	if !bytes.Equal(encoded, copyBytes) {
		t.Fatalf("repeated Outbox() changed bytes: first=%s second=%s", encoded, copyBytes)
	}
}

func TestProjectEventCallDue_StateTable(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*calendar.EventCallDueInput)
		want   calendar.EventCallDueDisposition
		active bool
		at     *time.Time
		key    string
		outbox bool
	}{
		{
			name: "inactive is identity only noop", mutate: func(input *calendar.EventCallDueInput) {
				input.Call.Active = false; input.Projection = nil; input.Envelope = calendar.EventCallDueEnvelopeInput{}
			}, want: calendar.EventCallDueNoop, active: false,
		},
		{
			name: "deleted inactive is noop", mutate: func(input *calendar.EventCallDueInput) {
				input.Call.Active, input.Call.Deleted = false, true; input.Projection = nil; input.Envelope = calendar.EventCallDueEnvelopeInput{}
			}, want: calendar.EventCallDueNoop, active: false,
		},
		{
			name: "early equal occurrence is not due", mutate: func(input *calendar.EventCallDueInput) {
				input.Now = dueStart().Add(-time.Second); input.Envelope.OccurredAt = input.Now
			}, want: calendar.EventCallDueNotDue, active: true, at: dueTime(dueStart()),
		},
		{
			name: "early moved earlier is reschedule", mutate: func(input *calendar.EventCallDueInput) {
				input.Now = dueStart().Add(-time.Hour); input.Envelope.OccurredAt = input.Now
				input.Projection.Exceptions = []calendar.EventCallProjectionException{{RecurrenceID: dueNominalKey, Status: calendar.CalendarEventStatusConfirmed, Temporal: dueTemporal(dueStart().Add(-30 * time.Minute), 60)}}
			}, want: calendar.EventCallDueReschedule, active: true, at: dueTime(dueStart().Add(-30 * time.Minute)),
		},
		{
			name: "early exdate skip is reschedule", mutate: func(input *calendar.EventCallDueInput) {
				input.Now = dueStart().Add(-time.Hour); input.Envelope.OccurredAt = input.Now; input.Projection.Event.Exdates = []string{dueNominalKey}
			}, want: calendar.EventCallDueReschedule, active: true, at: dueTime(dueStart().AddDate(0, 0, 1)),
		},
		{
			name: "early completed deactivates without outbox", mutate: func(input *calendar.EventCallDueInput) {
				input.Now = dueStart().Add(-time.Second); input.Envelope.OccurredAt = input.Now; input.Projection.Event.Exdates = []string{dueNominalKey, dueStart().AddDate(0, 0, 1).Format(time.RFC3339), dueStart().AddDate(0, 0, 2).Format(time.RFC3339)}; input.Projection.Exceptions = nil
			}, want: calendar.EventCallDueDeactivate, active: false,
		},
		{
			name: "due moved current reschedules", mutate: func(input *calendar.EventCallDueInput) {
				input.Projection.Exceptions = []calendar.EventCallProjectionException{{RecurrenceID: dueNominalKey, Status: calendar.CalendarEventStatusConfirmed, Temporal: dueTemporal(dueStart().Add(30 * time.Minute), 60)}}
			}, want: calendar.EventCallDueReschedule, active: true, at: dueTime(dueStart().Add(30 * time.Minute)),
		},
		{
			name: "due current and after completed deactivates without outbox", mutate: func(input *calendar.EventCallDueInput) {
				input.Projection.Event.Exdates = []string{dueNominalKey, dueStart().AddDate(0, 0, 1).Format(time.RFC3339), dueStart().AddDate(0, 0, 2).Format(time.RFC3339)}; input.Projection.Exceptions = nil
			}, want: calendar.EventCallDueDeactivate, active: false,
		},
		{
			name: "stale stored after one time exhaustion deactivates without outbox", mutate: func(input *calendar.EventCallDueInput) {
				input.Projection.Event.RRULE = ""; input.Projection.Exceptions = nil; input.Call.ScheduledAt = dueStart().AddDate(0, 0, 1); input.Now = input.Call.ScheduledAt.AddDate(0, 0, 1); input.Envelope.OccurredAt = input.Now
			}, want: calendar.EventCallDueDeactivate, active: false,
		},
		{
			name: "one time due deactivates with descriptor", mutate: func(input *calendar.EventCallDueInput) {
				input.Projection.Event.RRULE = ""; input.Projection.Exceptions = nil
			}, want: calendar.EventCallDueDeactivate, active: false, key: dueKey, outbox: true,
		},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := dueValidInput()
			testCase.mutate(&input)
			decision, err := calendar.ProjectEventCallDue(input)
			if err != nil {
				t.Fatalf("ProjectEventCallDue() error = %v", err)
			}
			assertDueSnapshot(t, decision.Snapshot(), testCase.want, testCase.active, derefDueTime(testCase.at), testCase.key)
			_, got := decision.Outbox()
			if got != testCase.outbox {
				t.Fatalf("Outbox() present = %t, want %t", got, testCase.outbox)
			}
		})
	}
}

func TestProjectEventCallDue_FailsClosedAndMapsErrors(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*calendar.EventCallDueInput)
		want   error
	}{
		{name: "deleted active call", mutate: func(input *calendar.EventCallDueInput) { input.Call.Deleted = true }, want: calendar.ErrEventCallDueInvalid},
		{name: "inactive projection", mutate: func(input *calendar.EventCallDueInput) { input.Call.Active = false }, want: calendar.ErrEventCallDueInvalid},
		{name: "active nil projection", mutate: func(input *calendar.EventCallDueInput) { input.Projection = nil }, want: calendar.ErrEventCallDueInvalid},
		{name: "target missing", mutate: func(input *calendar.EventCallDueInput) { input.Projection.Target.Exists = false }, want: calendar.ErrEventCallDueBoundary},
		{name: "target deleted", mutate: func(input *calendar.EventCallDueInput) { input.Projection.Target.Deleted = true }, want: calendar.ErrEventCallDueBoundary},
		{name: "target identity mismatch", mutate: func(input *calendar.EventCallDueInput) { input.Projection.Target.ProjectID = uuid.New() }, want: calendar.ErrEventCallDueBoundary},
		{name: "projection zero identity", mutate: func(input *calendar.EventCallDueInput) { input.Projection.Event.EventID = uuid.Nil }, want: calendar.ErrEventCallDueInvalid},
		{name: "occurred at differs from now", mutate: func(input *calendar.EventCallDueInput) { input.Envelope.OccurredAt = input.Now.Add(time.Second) }, want: calendar.ErrEventCallDueInvalid},
		{name: "non canonical correlation", mutate: func(input *calendar.EventCallDueInput) { input.Envelope.CorrelationID = " calendar-due-04k " }, want: calendar.ErrEventCallDueInvalid},
		{name: "projection invalid maps invalid", mutate: func(input *calendar.EventCallDueInput) { input.Projection.Event.Source = calendar.CalendarEventSourceICalFeed }, want: calendar.ErrEventCallDueInvalid},
		{name: "projection too large maps too large", mutate: func(input *calendar.EventCallDueInput) { input.Projection.Event.Title = strings.Repeat("x", 256) }, want: calendar.ErrEventCallDueTooLarge},
		{name: "now must be exact utc second", mutate: func(input *calendar.EventCallDueInput) { input.Now = input.Now.Add(time.Nanosecond); input.Envelope.OccurredAt = input.Now }, want: calendar.ErrEventCallDueInvalid},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := dueValidInput()
			testCase.mutate(&input)
			decision, err := calendar.ProjectEventCallDue(input)
			if !errors.Is(err, testCase.want) {
				t.Fatalf("ProjectEventCallDue() error = %v, want %v", err, testCase.want)
			}
			if snapshot := decision.Snapshot(); !reflect.DeepEqual(snapshot, calendar.EventCallDueSnapshot{}) {
				t.Fatalf("failure returned decision = %#v", snapshot)
			}
			if _, ok := decision.Outbox(); ok {
				t.Fatal("failure returned outbox")
			}
		})
	}
}

func TestProjectEventCallDue_PreflightsProjectionCapsAndCorrelationKinds(t *testing.T) {
	for _, testCase := range []struct {
		name   string
		mutate func(*calendar.EventCallDueInput)
		want   error
	}{
		{name: "exdates over cap fail before recurrence", mutate: func(input *calendar.EventCallDueInput) { input.Projection.Event.Exdates = make([]string, 1001) }, want: calendar.ErrEventCallDueTooLarge},
		{name: "attendees over cap fail before authorization scan", mutate: func(input *calendar.EventCallDueInput) { input.Projection.Attendees = make([]calendar.EventCallProjectionAttendee, 201) }, want: calendar.ErrEventCallDueTooLarge},
		{name: "exceptions over cap fail before recurrence", mutate: func(input *calendar.EventCallDueInput) { input.Projection.Exceptions = make([]calendar.EventCallProjectionException, 1001) }, want: calendar.ErrEventCallDueTooLarge},
		{name: "single byte correlation passes", mutate: func(input *calendar.EventCallDueInput) { input.Envelope.CorrelationID = "x" }, want: nil},
		{name: "128 byte correlation passes", mutate: func(input *calendar.EventCallDueInput) { input.Envelope.CorrelationID = strings.Repeat("c", 128) }, want: nil},
		{name: "129 byte correlation is too large before scans", mutate: func(input *calendar.EventCallDueInput) { input.Envelope.CorrelationID = strings.Repeat("c", 129) }, want: calendar.ErrEventCallDueTooLarge},
		{name: "huge invalid utf8 correlation is too large without panic", mutate: func(input *calendar.EventCallDueInput) { input.Envelope.CorrelationID = string(append([]byte(strings.Repeat("x", 4096)), 0xff)) }, want: calendar.ErrEventCallDueTooLarge},
		{name: "empty correlation is invalid", mutate: func(input *calendar.EventCallDueInput) { input.Envelope.CorrelationID = "" }, want: calendar.ErrEventCallDueInvalid},
		{name: "trim changing correlation is invalid", mutate: func(input *calendar.EventCallDueInput) { input.Envelope.CorrelationID = " calendar-due-04k" }, want: calendar.ErrEventCallDueInvalid},
		{name: "control correlation is invalid", mutate: func(input *calendar.EventCallDueInput) { input.Envelope.CorrelationID = "calendar\n-due" }, want: calendar.ErrEventCallDueInvalid},
		{name: "short invalid utf8 correlation is invalid", mutate: func(input *calendar.EventCallDueInput) { input.Envelope.CorrelationID = string([]byte{'x', 0xff}) }, want: calendar.ErrEventCallDueInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := dueValidInput()
			testCase.mutate(&input)
			decision, err := projectDueWithoutPanic(t, input)
			if testCase.want == nil {
				if err != nil || !decision.Snapshot().Active { t.Fatalf("ProjectEventCallDue() = %#v, %v", decision.Snapshot(), err) }
				return
			}
			if !errors.Is(err, testCase.want) { t.Fatalf("ProjectEventCallDue() error = %v, want %v", err, testCase.want) }
		})
	}
}

func TestProjectEventCallDue_ExpandedOrderedStateTable(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*calendar.EventCallDueInput)
		want   calendar.EventCallDueDisposition
		active bool
		at     time.Time
		key    string
		outbox bool
	}{
		{name: "early moved later uses after now", mutate: func(input *calendar.EventCallDueInput) { input.Now = dueStart().Add(-time.Hour); input.Envelope.OccurredAt = input.Now; input.Projection.Exceptions = []calendar.EventCallProjectionException{{RecurrenceID: dueNominalKey, Status: calendar.CalendarEventStatusConfirmed, Temporal: dueTemporal(dueStart().Add(30*time.Minute), 60)}} }, want: calendar.EventCallDueReschedule, active: true, at: dueStart().Add(30*time.Minute), key: ""},
		{name: "early current and after completed deactivates without outbox", mutate: func(input *calendar.EventCallDueInput) { input.Now = dueStart().Add(-time.Second); input.Envelope.OccurredAt = input.Now; input.Projection.Event.Exdates = []string{dueNominalKey, dueStart().AddDate(0, 0, 1).Format(time.RFC3339), dueStart().AddDate(0, 0, 2).Format(time.RFC3339)}; input.Projection.Exceptions = nil }, want: calendar.EventCallDueDeactivate, active: false, key: ""},
		{name: "overdue advances one successor only", mutate: func(input *calendar.EventCallDueInput) { input.Now = dueStart().AddDate(0, 0, 1); input.Envelope.OccurredAt = input.Now }, want: calendar.EventCallDueAdvance, active: true, at: dueStart().AddDate(0, 0, 2), key: dueKey, outbox: true},
		{name: "overdue one time deactivates with one descriptor", mutate: func(input *calendar.EventCallDueInput) { input.Now = dueStart().AddDate(0, 0, 2); input.Envelope.OccurredAt = input.Now; input.Projection.Event.RRULE = ""; input.Projection.Exceptions = nil }, want: calendar.EventCallDueDeactivate, active: false, key: dueKey, outbox: true},
		{name: "finite recurrence exhaustion deactivates with descriptor", mutate: func(input *calendar.EventCallDueInput) { input.Projection.Event.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=1"; input.Projection.Exceptions = nil }, want: calendar.EventCallDueDeactivate, active: false, key: dueKey, outbox: true},
	}
	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			input := dueValidInput()
			testCase.mutate(&input)
			decision, err := calendar.ProjectEventCallDue(input)
			if err != nil { t.Fatalf("ProjectEventCallDue() error = %v", err) }
			got := decision.Snapshot()
			if got.Disposition != testCase.want || got.Active != testCase.active || got.OccurrenceKey != testCase.key { t.Fatalf("Snapshot() = %#v", got) }
			if testCase.at.IsZero() { if got.ScheduledAt != nil { t.Fatalf("ScheduledAt = %v, want nil", got.ScheduledAt) } } else if got.ScheduledAt == nil || !got.ScheduledAt.Equal(testCase.at) { t.Fatalf("ScheduledAt = %v, want %v", got.ScheduledAt, testCase.at) }
			_, hasOutbox := decision.Outbox()
			if hasOutbox != testCase.outbox { t.Fatalf("Outbox() = %t, want %t", hasOutbox, testCase.outbox) }
		})
	}
}

func TestProjectEventCallDue_Maps04JBoundaryAndValidatesAllRawIdentities(t *testing.T) {
	boundaryInput := dueValidInput()
	boundaryInput.Projection.Event.Status = calendar.CalendarEventStatusCancelled
	if _, err := calendar.ProjectEventCallDue(boundaryInput); !errors.Is(err, calendar.ErrEventCallDueBoundary) { t.Fatalf("04J boundary error = %v", err) }

	tests := []struct { name string; mutate func(*calendar.EventCallDueInput); want error }{
		{"call scheduled id zero", func(input *calendar.EventCallDueInput) { input.Call.ScheduledCallID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"call event id zero", func(input *calendar.EventCallDueInput) { input.Call.CalendarEventID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"call organization zero", func(input *calendar.EventCallDueInput) { input.Call.OrganizationID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"call project zero", func(input *calendar.EventCallDueInput) { input.Call.ProjectID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"call channel zero", func(input *calendar.EventCallDueInput) { input.Call.ChannelID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"call creator zero", func(input *calendar.EventCallDueInput) { input.Call.CreatorID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"event scheduled call zero", func(input *calendar.EventCallDueInput) { input.Projection.Event.ScheduledCallID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"event event zero", func(input *calendar.EventCallDueInput) { input.Projection.Event.EventID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"event organization zero", func(input *calendar.EventCallDueInput) { input.Projection.Event.OrganizationID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"event project zero", func(input *calendar.EventCallDueInput) { input.Projection.Event.ProjectID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"event channel zero", func(input *calendar.EventCallDueInput) { input.Projection.Event.ChannelID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"event organizer zero", func(input *calendar.EventCallDueInput) { input.Projection.Event.OrganizerID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"target scheduled call zero", func(input *calendar.EventCallDueInput) { input.Projection.Target.ScheduledCallID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"target event zero", func(input *calendar.EventCallDueInput) { input.Projection.Target.EventID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"target creator zero", func(input *calendar.EventCallDueInput) { input.Projection.Target.CreatorID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"target organization zero", func(input *calendar.EventCallDueInput) { input.Projection.Target.OrganizationID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"target project zero", func(input *calendar.EventCallDueInput) { input.Projection.Target.ProjectID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"target channel zero", func(input *calendar.EventCallDueInput) { input.Projection.Target.ChannelID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"creator user zero", func(input *calendar.EventCallDueInput) { input.Projection.Creator.UserID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"creator organization zero", func(input *calendar.EventCallDueInput) { input.Projection.Creator.OrganizationID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"creator project zero", func(input *calendar.EventCallDueInput) { input.Projection.Creator.ProjectID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"creator channel zero", func(input *calendar.EventCallDueInput) { input.Projection.Creator.ChannelID = uuid.Nil }, calendar.ErrEventCallDueInvalid},
		{"event scheduled call mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Event.ScheduledCallID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"event identifier mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Event.EventID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"event organization mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Event.OrganizationID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"event project mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Event.ProjectID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"event channel mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Event.ChannelID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"event organizer mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Event.OrganizerID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"target scheduled call mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Target.ScheduledCallID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"target identifier mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Target.EventID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"target creator mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Target.CreatorID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"target organization mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Target.OrganizationID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"target project mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Target.ProjectID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"target channel mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Target.ChannelID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"creator user mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Creator.UserID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"creator organization mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Creator.OrganizationID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"creator project mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Creator.ProjectID = uuid.New() }, calendar.ErrEventCallDueBoundary},
		{"creator channel mismatch", func(input *calendar.EventCallDueInput) { input.Projection.Creator.ChannelID = uuid.New() }, calendar.ErrEventCallDueBoundary},
	}
	for _, testCase := range tests { t.Run(testCase.name, func(t *testing.T) { input := dueValidInput(); testCase.mutate(&input); if _, err := calendar.ProjectEventCallDue(input); !errors.Is(err, testCase.want) { t.Fatalf("ProjectEventCallDue() error = %v, want %v", err, testCase.want) } }) }
}

func TestProjectEventCallDue_ValidationEdgesAndRealProjectionRelations(t *testing.T) {
	for _, testCase := range []struct { name string; mutate func(*calendar.EventCallDueInput); want error }{
		{"255 byte title passes", func(input *calendar.EventCallDueInput) { input.Projection.Event.Title = strings.Repeat("t", 255) }, nil},
		{"256 byte title maps too large", func(input *calendar.EventCallDueInput) { input.Projection.Event.Title = strings.Repeat("t", 256) }, calendar.ErrEventCallDueTooLarge},
		{"control title maps invalid", func(input *calendar.EventCallDueInput) { input.Projection.Event.Title = "bad\nname" }, calendar.ErrEventCallDueInvalid},
		{"1440 minute duration passes", func(input *calendar.EventCallDueInput) { input.Projection.Event.Temporal.EndsAt = input.Projection.Event.Temporal.StartsAt.Add(1440 * time.Minute) }, nil},
		{"1441 minute duration maps too large", func(input *calendar.EventCallDueInput) { input.Projection.Event.Temporal.EndsAt = input.Projection.Event.Temporal.StartsAt.Add(1441 * time.Minute) }, calendar.ErrEventCallDueTooLarge},
		{"fractional duration maps invalid", func(input *calendar.EventCallDueInput) { input.Projection.Event.Temporal.EndsAt = input.Projection.Event.Temporal.StartsAt.Add(time.Minute + time.Nanosecond) }, calendar.ErrEventCallDueInvalid},
		{"fixed zero offset call time is invalid", func(input *calendar.EventCallDueInput) { input.Call.ScheduledAt = time.Date(2026, time.August, 21, 9, 0, 0, 0, time.FixedZone("UTC", 0)) }, calendar.ErrEventCallDueInvalid},
		{"fractional call time is invalid", func(input *calendar.EventCallDueInput) { input.Call.ScheduledAt = input.Call.ScheduledAt.Add(time.Nanosecond) }, calendar.ErrEventCallDueInvalid},
		{"fixed zero offset now is invalid", func(input *calendar.EventCallDueInput) { input.Now = time.Date(2026, time.August, 21, 9, 0, 0, 0, time.FixedZone("UTC", 0)); input.Envelope.OccurredAt = input.Now }, calendar.ErrEventCallDueInvalid},
		{"inactive malformed call is not fast path bypass", func(input *calendar.EventCallDueInput) { input.Call.Active = false; input.Projection = nil; input.Envelope = calendar.EventCallDueEnvelopeInput{}; input.Call.ScheduledAt = time.Time{} }, calendar.ErrEventCallDueInvalid},
		{"inactive nonzero envelope is invalid", func(input *calendar.EventCallDueInput) { input.Call.Active = false; input.Projection = nil }, calendar.ErrEventCallDueInvalid},
		{"effective collision is rejected by real 04J and maps invalid", func(input *calendar.EventCallDueInput) { input.Projection.Exceptions = append(input.Projection.Exceptions, calendar.EventCallProjectionException{RecurrenceID: dueStart().AddDate(0, 0, 1).Format(time.RFC3339), Status: calendar.CalendarEventStatusConfirmed, Temporal: dueTemporal(dueStart(), 60)}) }, calendar.ErrEventCallDueInvalid},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := dueValidInput()
			testCase.mutate(&input)
			decision, err := calendar.ProjectEventCallDue(input)
			if testCase.want == nil {
				if err != nil || !decision.Snapshot().Active { t.Fatalf("ProjectEventCallDue() = %#v, %v", decision.Snapshot(), err) }
				return
			}
			if !errors.Is(err, testCase.want) { t.Fatalf("ProjectEventCallDue() error = %v, want %v", err, testCase.want) }
		})
	}
}

func TestProjectEventCallDue_DeepCopiesAndReplayIsDeterministic(t *testing.T) {
	input := dueValidInput()
	before := cloneDueInput(input)
	first, err := calendar.ProjectEventCallDue(input)
	if err != nil {
		t.Fatalf("first ProjectEventCallDue() error = %v", err)
	}
	if !reflect.DeepEqual(input, before) {
		t.Fatalf("ProjectEventCallDue() mutated input: got %#v want %#v", input, before)
	}
	firstSnapshot := first.Snapshot()
	input.Projection.Event.Title = "mutated"
	input.Projection.Event.Exdates = append(input.Projection.Event.Exdates, dueNominalKey)
	*input.Projection.Event.ReminderMinutes = 40320
	input.Projection.Attendees[0].Email = "mutated@example.test"
	*input.Projection.Attendees[0].UserID = uuid.New()
	input.Projection.Exceptions[0].Temporal.StartsAt = input.Projection.Exceptions[0].Temporal.StartsAt.Add(time.Hour)
	if got := first.Snapshot(); !reflect.DeepEqual(got, firstSnapshot) {
		t.Fatalf("decision aliases input: got %#v want %#v", got, firstSnapshot)
	}
	returned := first.Snapshot()
	if returned.ScheduledAt == nil { t.Fatal("advance snapshot ScheduledAt is nil") }
	*returned.ScheduledAt = returned.ScheduledAt.Add(24 * time.Hour)
	if reread := first.Snapshot(); reread.ScheduledAt == nil || !reread.ScheduledAt.Equal(firstSnapshot.ScheduledAt.UTC()) {
		t.Fatalf("Snapshot() scheduled-at aliases prior read: %#v", reread)
	}
	firstOutbox, ok := first.Outbox()
	if !ok {
		t.Fatal("first Outbox() absent")
	}
	firstBytes, _ := json.Marshal(firstOutbox.Envelope())
	second, err := calendar.ProjectEventCallDue(before)
	if err != nil {
		t.Fatalf("replay ProjectEventCallDue() error = %v", err)
	}
	secondOutbox, ok := second.Outbox()
	if !ok {
		t.Fatal("replay Outbox() absent")
	}
	secondBytes, _ := json.Marshal(secondOutbox.Envelope())
	if !bytes.Equal(firstBytes, secondBytes) || !reflect.DeepEqual(first.Snapshot(), second.Snapshot()) {
		t.Fatalf("same full input was nondeterministic: first=%s second=%s", firstBytes, secondBytes)
	}

	changedEnvelope := cloneDueInput(before)
	changedEnvelope.Envelope.EventID = uuid.MustParse("807192a3-b4c5-d6e7-f809-1a2b3c4d5e6f")
	third, err := calendar.ProjectEventCallDue(changedEnvelope)
	if err != nil {
		t.Fatalf("changed envelope ProjectEventCallDue() error = %v", err)
	}
	thirdOutbox, _ := third.Outbox()
	thirdBytes, _ := json.Marshal(thirdOutbox.Envelope())
	if third.Snapshot().OccurrenceKey != first.Snapshot().OccurrenceKey || bytes.Equal(firstBytes, thirdBytes) {
		t.Fatal("Envelope.EventID must change envelope bytes but not occurrence key")
	}
	var changedEnvelopeJSON struct { EventID uuid.UUID `json:"event_id"` }
	if err := json.Unmarshal(thirdBytes, &changedEnvelopeJSON); err != nil || changedEnvelopeJSON.EventID != changedEnvelope.Envelope.EventID {
		t.Fatalf("changed envelope event_id = %s, %v; want %s", changedEnvelopeJSON.EventID, err, changedEnvelope.Envelope.EventID)
	}

	changedCalendarEvent := cloneDueInput(before)
	changedCalendarEvent.Call.CalendarEventID = uuid.MustParse("907192a3-b4c5-d6e7-f809-1a2b3c4d5e6f")
	changedCalendarEvent.Projection.Event.EventID = changedCalendarEvent.Call.CalendarEventID
	changedCalendarEvent.Projection.Target.EventID = changedCalendarEvent.Call.CalendarEventID
	fourth, err := calendar.ProjectEventCallDue(changedCalendarEvent)
	if err != nil {
		t.Fatalf("changed calendar event ProjectEventCallDue() error = %v", err)
	}
	if fourth.Snapshot().OccurrenceKey == first.Snapshot().OccurrenceKey {
		t.Fatal("CalendarEventID must change occurrence key")
	}
	fourthOutbox, ok := fourth.Outbox()
	if !ok { t.Fatal("changed calendar event Outbox() absent") }
	fourthBytes, _ := json.Marshal(fourthOutbox.Envelope())
	var changedCalendarEventJSON struct { Payload calendar.EventCallScheduledDuePayload `json:"payload"` }
	if err := json.Unmarshal(fourthBytes, &changedCalendarEventJSON); err != nil || changedCalendarEventJSON.Payload.CalendarEventID != changedCalendarEvent.Call.CalendarEventID {
		t.Fatalf("changed payload calendar_event_id = %s, %v; want %s", changedCalendarEventJSON.Payload.CalendarEventID, err, changedCalendarEvent.Call.CalendarEventID)
	}
}

func TestProjectEventCallDue_PayloadVariantsAndDST(t *testing.T) {
	t.Run("one time payload emits recurrence null", func(t *testing.T) {
		input := dueValidInput()
		input.Projection.Event.RRULE = ""
		input.Projection.Exceptions = nil
		decision, err := calendar.ProjectEventCallDue(input)
		if err != nil { t.Fatalf("ProjectEventCallDue() error = %v", err) }
		descriptor, ok := decision.Outbox()
		if !ok { t.Fatal("one-time due descriptor absent") }
		encoded, _ := json.Marshal(descriptor.Envelope())
		var envelope struct { Payload map[string]json.RawMessage `json:"payload"` }
		if err := json.Unmarshal(encoded, &envelope); err != nil { t.Fatalf("json.Unmarshal() error = %v", err) }
		if got := string(envelope.Payload["recurrence_id"]); got != "null" { t.Fatalf("recurrence_id = %s, want null", got) }
	})

	t.Run("moved effective duration retains nominal key", func(t *testing.T) {
		input := dueValidInput()
		input.Call.ScheduledAt = dueStart().Add(30 * time.Minute)
		input.Now = input.Call.ScheduledAt
		input.Envelope.OccurredAt = input.Now
		input.Projection.Exceptions = []calendar.EventCallProjectionException{{RecurrenceID: dueNominalKey, Status: calendar.CalendarEventStatusConfirmed, Temporal: dueTemporal(input.Call.ScheduledAt, 90)}}
		decision, err := calendar.ProjectEventCallDue(input)
		if err != nil { t.Fatalf("ProjectEventCallDue() error = %v", err) }
		if got := decision.Snapshot().OccurrenceKey; got != dueKey { t.Fatalf("OccurrenceKey = %q, want nominal key hash %q", got, dueKey) }
		descriptor, ok := decision.Outbox()
		if !ok { t.Fatal("moved due descriptor absent") }
		encoded, _ := json.Marshal(descriptor.Envelope())
		var envelope struct { Payload calendar.EventCallScheduledDuePayload `json:"payload"` }
		if err := json.Unmarshal(encoded, &envelope); err != nil { t.Fatalf("json.Unmarshal() error = %v", err) }
		if envelope.Payload.DurationMinutes != 90 || !envelope.Payload.ScheduledAt.Equal(input.Call.ScheduledAt) || envelope.Payload.RecurrenceID == nil || *envelope.Payload.RecurrenceID != dueNominalKey { t.Fatalf("moved payload = %#v", envelope.Payload) }
	})

	for _, testCase := range []struct { name string; start, next time.Time }{
		{"dst gap skips nonexistent local time", time.Date(2026, time.March, 7, 7, 30, 0, 0, time.UTC), time.Date(2026, time.March, 9, 6, 30, 0, 0, time.UTC)},
		{"dst fold chooses earliest local instant", time.Date(2026, time.October, 31, 5, 30, 0, 0, time.UTC), time.Date(2026, time.November, 1, 5, 30, 0, 0, time.UTC)},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			input := dueValidInput()
			input.Call.ScheduledAt, input.Now = testCase.start, testCase.start
			input.Envelope.OccurredAt = input.Now
			input.Projection.Event.Temporal = calendar.CalendarEvent{StartsAt: testCase.start, EndsAt: testCase.start.Add(time.Hour), Timezone: "America/New_York"}
			input.Projection.Event.RRULE = "FREQ=DAILY;INTERVAL=1;COUNT=2"
			input.Projection.Exceptions = nil
			decision, err := calendar.ProjectEventCallDue(input)
			if err != nil { t.Fatalf("ProjectEventCallDue() error = %v", err) }
			got := decision.Snapshot()
			if got.Disposition != calendar.EventCallDueAdvance || got.ScheduledAt == nil || !got.ScheduledAt.Equal(testCase.next) { t.Fatalf("DST decision = %#v, want %v", got, testCase.next) }
		})
	}
}

func TestEventCallDueError_IsContentFree(t *testing.T) {
	for _, testCase := range []struct {
		kind calendar.EventCallDueErrorKind
		want error
		text string
	}{
		{calendar.EventCallDueInvalid, calendar.ErrEventCallDueInvalid, "calendar: event call due invalid"},
		{calendar.EventCallDueBoundary, calendar.ErrEventCallDueBoundary, "calendar: event call due boundary mismatch"},
		{calendar.EventCallDueTooLarge, calendar.ErrEventCallDueTooLarge, "calendar: event call due too large"},
	} {
		input := dueValidInput()
		if testCase.kind == calendar.EventCallDueInvalid { input.Call.ScheduledCallID = uuid.Nil } else if testCase.kind == calendar.EventCallDueBoundary { input.Projection.Target.Active = false } else { input.Projection.Event.Title = strings.Repeat("x", 256) }
		_, err := calendar.ProjectEventCallDue(input)
		if !errors.Is(err, testCase.want) || err.Error() != testCase.text {
			t.Fatalf("typed error = %v, want %q", err, testCase.text)
		}
	}
	var nilError *calendar.EventCallDueError
	if nilError.Error() != "calendar: event call due failed" || nilError.Kind() != "" || nilError.Unwrap() != nil || errors.Is(nilError, calendar.ErrEventCallDueInvalid) {
		t.Fatal("nil due error behavior changed")
	}
	unknown := &calendar.EventCallDueError{}
	if unknown.Error() != "calendar: event call due failed" || unknown.Kind() != "" || unknown.Unwrap() != nil || errors.Is(unknown, calendar.ErrEventCallDueBoundary) {
		t.Fatal("unknown due error behavior changed")
	}
}

type dueField struct { name string; typ reflect.Type; tag string }

func assertDueStruct(t *testing.T, got reflect.Type, want []dueField) {
	t.Helper()
	if got.NumField() != len(want) { t.Fatalf("%s fields = %d, want %d", got, got.NumField(), len(want)) }
	for index, expected := range want {
		field := got.Field(index)
		if field.Name != expected.name || field.Type != expected.typ || string(field.Tag) != expected.tag {
			t.Fatalf("%s field %d = %s %s %q, want %s %s %q", got, index, field.Name, field.Type, field.Tag, expected.name, expected.typ, expected.tag)
		}
	}
}

func dueValidInput() calendar.EventCallDueInput {
	creatorID := dueCreatorID
	participantID := dueParticipantID
	reminder := 15
	projection := &calendar.EventCallDueProjectionInput{
		Event: calendar.EventCallProjectionEvent{
			EventID: dueCalendarEventID, ScheduledCallID: dueScheduledCallID, OrganizationID: dueOrganizationID, ProjectID: dueProjectID, ChannelID: dueChannelID,
			OrganizerID: dueCreatorID, OrganizerEmail: "owner@example.test", Source: calendar.CalendarEventSourceHub, Status: calendar.CalendarEventStatusConfirmed,
			Title: "Daily planning", Description: "Discuss next delivery.", Temporal: calendar.CalendarEvent{StartsAt: dueStart(), EndsAt: dueStart().Add(time.Hour), Timezone: "UTC"}, RRULE: "FREQ=DAILY;INTERVAL=1;COUNT=3", ReminderMinutes: &reminder,
		},
		Target: calendar.EventCallProjectionTarget{ScheduledCallID: dueScheduledCallID, EventID: dueCalendarEventID, CreatorID: dueCreatorID, OrganizationID: dueOrganizationID, ProjectID: dueProjectID, ChannelID: dueChannelID, Exists: true, Active: true},
		Creator: calendar.EventCallProjectionCreator{UserID: dueCreatorID, Email: "owner@example.test", OrganizationID: dueOrganizationID, ProjectID: dueProjectID, ChannelID: dueChannelID, Active: true, ActiveProjectMember: true, ActiveChannelMember: true},
		Attendees: []calendar.EventCallProjectionAttendee{
			{UserID: &creatorID, Email: "owner@example.test", OrganizationID: dueOrganizationID, ProjectID: dueProjectID, Role: calendar.CalendarEventAttendeeRoleRequired, Status: calendar.RSVPAccepted, Active: true, ActiveProjectMember: true},
			{UserID: &participantID, Email: "participant@example.test", OrganizationID: dueOrganizationID, ProjectID: dueProjectID, Role: calendar.CalendarEventAttendeeRoleRequired, Status: calendar.RSVPNeedsAction, Active: true, ActiveProjectMember: true},
		},
		Exceptions: []calendar.EventCallProjectionException{{RecurrenceID: dueStart().AddDate(0, 0, 2).Format(time.RFC3339), Status: calendar.CalendarEventStatusConfirmed, Temporal: dueTemporal(dueStart().AddDate(0, 0, 2), 60)}},
	}
	return calendar.EventCallDueInput{
		Call: calendar.EventCallDueCallSnapshot{ScheduledCallID: dueScheduledCallID, CalendarEventID: dueCalendarEventID, OrganizationID: dueOrganizationID, ProjectID: dueProjectID, ChannelID: dueChannelID, CreatorID: dueCreatorID, ScheduledAt: dueStart(), Active: true},
		Now: dueStart(), Projection: projection, Envelope: calendar.EventCallDueEnvelopeInput{EventID: dueEnvelopeEventID, OccurredAt: dueStart(), CorrelationID: "calendar-due-04k"},
	}
}

func dueStart() time.Time { return time.Date(2026, time.August, 21, 9, 0, 0, 0, time.UTC) }
func dueTemporal(start time.Time, minutes int) *calendar.CalendarEvent { return &calendar.CalendarEvent{StartsAt: start, EndsAt: start.Add(time.Duration(minutes) * time.Minute), Timezone: "UTC"} }
func dueTime(value time.Time) *time.Time { copy := value; return &copy }
func derefDueTime(value *time.Time) time.Time { if value == nil { return time.Time{} }; return *value }

func assertDueSnapshot(t *testing.T, got calendar.EventCallDueSnapshot, disposition calendar.EventCallDueDisposition, active bool, scheduledAt time.Time, key string) {
	t.Helper()
	if got.Disposition != disposition || got.ScheduledCallID != dueScheduledCallID || got.CalendarEventID != dueCalendarEventID || got.OrganizationID != dueOrganizationID || got.ProjectID != dueProjectID || got.ChannelID != dueChannelID || got.CreatorID != dueCreatorID || got.Active != active || got.OccurrenceKey != key {
		t.Fatalf("Snapshot() = %#v", got)
	}
	if scheduledAt.IsZero() {
		if got.ScheduledAt != nil { t.Fatalf("Snapshot().ScheduledAt = %v, want nil", got.ScheduledAt) }
		return
	}
	if got.ScheduledAt == nil || !got.ScheduledAt.Equal(scheduledAt) { t.Fatalf("Snapshot().ScheduledAt = %v, want %v", got.ScheduledAt, scheduledAt) }
	*got.ScheduledAt = got.ScheduledAt.Add(time.Hour)
	if got.ScheduledAt.Equal(scheduledAt) { t.Fatal("test failed to mutate returned scheduled-at copy") }
}

func cloneDueInput(input calendar.EventCallDueInput) calendar.EventCallDueInput {
	clone := input
	if input.Projection == nil { return clone }
	projection := *input.Projection
	projection.Event.Exdates = append([]string(nil), input.Projection.Event.Exdates...)
	if input.Projection.Event.ReminderMinutes != nil { value := *input.Projection.Event.ReminderMinutes; projection.Event.ReminderMinutes = &value }
	projection.Attendees = make([]calendar.EventCallProjectionAttendee, len(input.Projection.Attendees))
	for index, attendee := range input.Projection.Attendees { projection.Attendees[index] = attendee; if attendee.UserID != nil { value := *attendee.UserID; projection.Attendees[index].UserID = &value } }
	projection.Exceptions = make([]calendar.EventCallProjectionException, len(input.Projection.Exceptions))
	for index, exception := range input.Projection.Exceptions { projection.Exceptions[index] = exception; if exception.Temporal != nil { temporal := *exception.Temporal; projection.Exceptions[index].Temporal = &temporal } }
	clone.Projection = &projection
	return clone
}

func projectDueWithoutPanic(t *testing.T, input calendar.EventCallDueInput) (decision calendar.EventCallDueDecision, err error) {
	t.Helper()
	defer func() {
		if recovered := recover(); recovered != nil {
			t.Fatalf("ProjectEventCallDue() panicked: %v", recovered)
		}
	}()
	return calendar.ProjectEventCallDue(input)
}
