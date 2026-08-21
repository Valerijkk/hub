package calendar

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"

	"github.com/valerii/hub/backend/internal/platform/events"
)

const (
	eventCallDueMaxTitleBytes       = 255
	eventCallDueMaxDurationMinutes = 1440
	eventCallDueMaxCorrelationBytes = 128
	eventCallDueOccurrenceKeyBytes  = 20
	eventCallDueKeyPrefix           = "hub.calendar.scheduled_due.v1\x00"
)

// EventCallDueCallSnapshot is the immutable currently persisted scheduled-call state.
type EventCallDueCallSnapshot struct {
	ScheduledCallID uuid.UUID
	CalendarEventID uuid.UUID
	OrganizationID  uuid.UUID
	ProjectID       uuid.UUID
	ChannelID       uuid.UUID
	CreatorID       uuid.UUID
	ScheduledAt     time.Time
	Active          bool
	Deleted         bool
}

// EventCallDueEnvelopeInput is the caller-owned candidate envelope identity for an active evaluation.
type EventCallDueEnvelopeInput struct {
	EventID       uuid.UUID
	OccurredAt    time.Time
	CorrelationID string
}

// EventCallDueProjectionInput is the raw canonical data consumed by the accepted 04J policy.
type EventCallDueProjectionInput struct {
	Event      EventCallProjectionEvent
	Target     EventCallProjectionTarget
	Creator    EventCallProjectionCreator
	Attendees  []EventCallProjectionAttendee
	Exceptions []EventCallProjectionException
}

// EventCallDueInput contains the complete pure due-policy input.
type EventCallDueInput struct {
	Call       EventCallDueCallSnapshot
	Now        time.Time
	Projection *EventCallDueProjectionInput
	Envelope   EventCallDueEnvelopeInput
}

// EventCallScheduledDuePayload is the exact calls.channel scheduled-due payload contract.
type EventCallScheduledDuePayload struct {
	ScheduledCallID uuid.UUID `json:"scheduled_call_id"`
	CalendarEventID uuid.UUID `json:"calendar_event_id"`
	OccurrenceKey   string    `json:"occurrence_key"`
	RecurrenceID    *string   `json:"recurrence_id"`
	ChannelID       uuid.UUID `json:"channel_id"`
	ProjectID       uuid.UUID `json:"project_id"`
	Title           string    `json:"title"`
	ScheduledAt     time.Time `json:"scheduled_at"`
	DurationMinutes int       `json:"duration_minutes"`
}

// EventCallDueDisposition is the only allowed future scheduled-call command intent.
type EventCallDueDisposition string

const (
	EventCallDueNotDue     EventCallDueDisposition = "not_due"
	EventCallDueReschedule EventCallDueDisposition = "reschedule"
	EventCallDueAdvance    EventCallDueDisposition = "advance"
	EventCallDueDeactivate EventCallDueDisposition = "deactivate"
	EventCallDueNoop       EventCallDueDisposition = "noop"
)

// EventCallDueSnapshot is a copied read model of a due-policy decision.
type EventCallDueSnapshot struct {
	Disposition     EventCallDueDisposition
	ScheduledCallID uuid.UUID
	CalendarEventID uuid.UUID
	OrganizationID  uuid.UUID
	ProjectID       uuid.UUID
	ChannelID       uuid.UUID
	CreatorID       uuid.UUID
	OccurrenceKey   string
	ScheduledAt     *time.Time
	Active          bool
}

// EventCallDueDecision keeps result data and any descriptor private and immutable.
type EventCallDueDecision struct {
	snapshot EventCallDueSnapshot
	outbox   *EventOutboxDescriptor
}

// Snapshot returns a fresh copy of the due decision.
func (decision EventCallDueDecision) Snapshot() EventCallDueSnapshot {
	return cloneEventCallDueSnapshot(decision.snapshot)
}

// Outbox returns a complete frozen descriptor only when one stored occurrence is due.
func (decision EventCallDueDecision) Outbox() (EventOutboxDescriptor, bool) {
	if decision.outbox == nil {
		return EventOutboxDescriptor{}, false
	}
	return *decision.outbox, true
}

// EventCallDueErrorKind classifies content-free due-policy failures.
type EventCallDueErrorKind string

const (
	EventCallDueInvalid  EventCallDueErrorKind = "invalid"
	EventCallDueBoundary EventCallDueErrorKind = "boundary"
	EventCallDueTooLarge EventCallDueErrorKind = "too_large"
)

// EventCallDueError exposes only a frozen error classification.
type EventCallDueError struct {
	kind EventCallDueErrorKind
}

var (
	ErrEventCallDueInvalid  error = errors.New("calendar: event call due invalid")
	ErrEventCallDueBoundary error = errors.New("calendar: event call due boundary mismatch")
	ErrEventCallDueTooLarge error = errors.New("calendar: event call due too large")
)

// Error intentionally contains no identity, recurrence, title or payload data.
func (err *EventCallDueError) Error() string {
	if cause := err.Unwrap(); cause != nil {
		return cause.Error()
	}
	return "calendar: event call due failed"
}

// Kind returns a known kind only for a known failure.
func (err *EventCallDueError) Kind() EventCallDueErrorKind {
	if err == nil || eventCallDueSentinel(err.kind) == nil {
		return ""
	}
	return err.kind
}

// Unwrap exposes only one of the three frozen sentinels.
func (err *EventCallDueError) Unwrap() error {
	if err == nil {
		return nil
	}
	return eventCallDueSentinel(err.kind)
}

// Is supports errors.Is for the three frozen sentinels.
func (err *EventCallDueError) Is(target error) bool {
	cause := err.Unwrap()
	return cause != nil && target == cause
}

// ProjectEventCallDue derives one pure scheduled-call decision without I/O or generated values.
func ProjectEventCallDue(input EventCallDueInput) (EventCallDueDecision, error) {
	if !eventCallDueCallValid(input.Call) || !eventCallProjectionExactUTCSecond(input.Now) {
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueInvalid)
	}
	if !input.Call.Active || input.Call.Deleted {
		if input.Projection != nil || !eventCallDueZeroEnvelope(input.Envelope) {
			return EventCallDueDecision{}, eventCallDueFailure(EventCallDueInvalid)
		}
		return EventCallDueDecision{snapshot: eventCallDueIdentitySnapshot(input.Call, EventCallDueNoop, false, nil, "")}, nil
	}
	if input.Projection == nil {
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueInvalid)
	}
	if kind := eventCallDueProjectionCaps(*input.Projection); kind != "" {
		return EventCallDueDecision{}, eventCallDueFailure(kind)
	}
	if kind := eventCallDueEnvelopeKind(input.Envelope, input.Now); kind != "" {
		return EventCallDueDecision{}, eventCallDueFailure(kind)
	}
	if kind := eventCallDueProjectionPreflight(input.Call, *input.Projection); kind != "" {
		return EventCallDueDecision{}, eventCallDueFailure(kind)
	}

	projection := cloneEventCallDueProjectionInput(*input.Projection)
	current, err := ProjectEventCall(EventCallProjectionInput{
		Action: EventCallProjectionReconcile, Event: projection.Event, Target: projection.Target,
		Creator: projection.Creator, Attendees: projection.Attendees, Exceptions: projection.Exceptions,
		Boundary: EventCallProjectionBoundary{At: input.Call.ScheduledAt, Inclusive: true},
	})
	if err != nil {
		return EventCallDueDecision{}, eventCallDueProjectionFailure(err)
	}
	afterNow, err := ProjectEventCall(EventCallProjectionInput{
		Action: EventCallProjectionReconcile, Event: projection.Event, Target: projection.Target,
		Creator: projection.Creator, Attendees: projection.Attendees, Exceptions: projection.Exceptions,
		Boundary: EventCallProjectionBoundary{At: input.Now, Inclusive: false},
	})
	if err != nil {
		return EventCallDueDecision{}, eventCallDueProjectionFailure(err)
	}

	currentSnapshot := current.Snapshot()
	afterSnapshot := afterNow.Snapshot()
	if !eventCallDueProjectionSnapshotIdentityValid(input.Call, currentSnapshot) || !eventCallDueProjectionSnapshotIdentityValid(input.Call, afterSnapshot) {
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueBoundary)
	}
	if !eventCallDueProjectionDispositionAllowed(currentSnapshot.Disposition) || !eventCallDueProjectionDispositionAllowed(afterSnapshot.Disposition) {
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueBoundary)
	}
	if eventCallDueEqualInstantDifferentKey(currentSnapshot, afterSnapshot) {
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueBoundary)
	}

	if input.Now.Before(input.Call.ScheduledAt) {
		return eventCallDueEarlyDecision(input, currentSnapshot, afterSnapshot)
	}
	return eventCallDueDueDecision(input, currentSnapshot, afterSnapshot)
}

func eventCallDueEarlyDecision(input EventCallDueInput, current, after EventCallProjectionSnapshot) (EventCallDueDecision, error) {
	switch after.Disposition {
	case EventCallProjectionCompleted:
		return EventCallDueDecision{snapshot: eventCallDueIdentitySnapshot(input.Call, EventCallDueDeactivate, false, nil, "")}, nil
	case EventCallProjectionUpsert:
		if !eventCallDueOccurrenceValid(after.Occurrence) || after.Occurrence.StartsAt.Before(input.Now) {
			return EventCallDueDecision{}, eventCallDueFailure(EventCallDueBoundary)
		}
		if after.Occurrence.StartsAt.Equal(input.Call.ScheduledAt) {
			return EventCallDueDecision{snapshot: eventCallDueIdentitySnapshot(input.Call, EventCallDueNotDue, true, &input.Call.ScheduledAt, "")}, nil
		}
		return EventCallDueDecision{snapshot: eventCallDueIdentitySnapshot(input.Call, EventCallDueReschedule, true, &after.Occurrence.StartsAt, "")}, nil
	default:
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueBoundary)
	}
}

func eventCallDueDueDecision(input EventCallDueInput, current, after EventCallProjectionSnapshot) (EventCallDueDecision, error) {
	if current.Disposition == EventCallProjectionCompleted {
		if after.Disposition != EventCallProjectionCompleted {
			return EventCallDueDecision{}, eventCallDueFailure(EventCallDueBoundary)
		}
		return EventCallDueDecision{snapshot: eventCallDueIdentitySnapshot(input.Call, EventCallDueDeactivate, false, nil, "")}, nil
	}
	if current.Disposition != EventCallProjectionUpsert || !eventCallDueOccurrenceValid(current.Occurrence) {
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueBoundary)
	}
	if !current.Occurrence.StartsAt.Equal(input.Call.ScheduledAt) {
		return EventCallDueDecision{snapshot: eventCallDueIdentitySnapshot(input.Call, EventCallDueReschedule, true, &current.Occurrence.StartsAt, "")}, nil
	}

	switch after.Disposition {
	case EventCallProjectionCompleted:
		return eventCallDueWithOutbox(input, current, EventCallDueDeactivate, nil)
	case EventCallProjectionUpsert:
		if !eventCallDueOccurrenceValid(after.Occurrence) || !after.Occurrence.StartsAt.After(current.Occurrence.StartsAt) || after.Occurrence.Key == current.Occurrence.Key {
			return EventCallDueDecision{}, eventCallDueFailure(EventCallDueBoundary)
		}
		return eventCallDueWithOutbox(input, current, EventCallDueAdvance, &after.Occurrence.StartsAt)
	default:
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueBoundary)
	}
}

func eventCallDueWithOutbox(input EventCallDueInput, current EventCallProjectionSnapshot, disposition EventCallDueDisposition, scheduledAt *time.Time) (EventCallDueDecision, error) {
	if !eventCallDueTitleValid(current.Title) || current.DurationMinutes < 1 {
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueInvalid)
	}
	if current.DurationMinutes > eventCallDueMaxDurationMinutes {
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueTooLarge)
	}
	if !eventCallDueOccurrenceValid(current.Occurrence) {
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueInvalid)
	}
	key := eventCallDueKey(input.Call.CalendarEventID, input.Call.ScheduledCallID, current.Occurrence.Key)
	if key == "" {
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueInvalid)
	}
	var recurrenceID *string
	if current.RRULE != "" {
		value := current.Occurrence.Key
		recurrenceID = &value
	}
	payload := EventCallScheduledDuePayload{
		ScheduledCallID: input.Call.ScheduledCallID, CalendarEventID: input.Call.CalendarEventID, OccurrenceKey: key,
		RecurrenceID: recurrenceID, ChannelID: input.Call.ChannelID, ProjectID: input.Call.ProjectID,
		Title: current.Title, ScheduledAt: current.Occurrence.StartsAt, DurationMinutes: current.DurationMinutes,
	}
	subject := "calls.channel." + input.Call.ChannelID.String() + ".scheduled_due.v1"
	envelope, err := events.NewEnvelopeV1(subject, input.Envelope.EventID, input.Envelope.OccurredAt, input.Call.OrganizationID, nil, input.Envelope.CorrelationID, payload)
	if err != nil || !eventCallDueEnvelopeBindingValid(envelope, subject, input, payload) {
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueInvalid)
	}
	descriptor, err := NewEventOutboxDescriptor(envelope)
	if err != nil {
		return EventCallDueDecision{}, eventCallDueFailure(EventCallDueInvalid)
	}
	decision := EventCallDueDecision{
		snapshot: eventCallDueIdentitySnapshot(input.Call, disposition, disposition == EventCallDueAdvance, scheduledAt, key),
		outbox:   &descriptor,
	}
	return decision, nil
}

func eventCallDueCallValid(call EventCallDueCallSnapshot) bool {
	if call.ScheduledCallID == uuid.Nil || call.CalendarEventID == uuid.Nil || call.OrganizationID == uuid.Nil || call.ProjectID == uuid.Nil || call.ChannelID == uuid.Nil || call.CreatorID == uuid.Nil {
		return false
	}
	if !eventCallProjectionExactUTCSecond(call.ScheduledAt) {
		return false
	}
	return !call.Deleted || !call.Active
}

func eventCallDueZeroEnvelope(envelope EventCallDueEnvelopeInput) bool {
	return envelope.EventID == uuid.Nil && envelope.OccurredAt.IsZero() && envelope.CorrelationID == ""
}

func eventCallDueProjectionCaps(projection EventCallDueProjectionInput) EventCallDueErrorKind {
	if len(projection.Event.Exdates) > eventCallProjectionMaxExdates || len(projection.Attendees) > eventCallProjectionMaxAttendees || len(projection.Exceptions) > eventCallProjectionMaxExceptions {
		return EventCallDueTooLarge
	}
	return ""
}

func eventCallDueEnvelopeKind(envelope EventCallDueEnvelopeInput, now time.Time) EventCallDueErrorKind {
	if envelope.EventID == uuid.Nil || !eventCallProjectionExactUTCSecond(envelope.OccurredAt) || !envelope.OccurredAt.Equal(now) {
		return EventCallDueInvalid
	}
	return eventCallDueCorrelationKind(envelope.CorrelationID)
}

func eventCallDueCorrelationKind(value string) EventCallDueErrorKind {
	if len(value) > eventCallDueMaxCorrelationBytes {
		return EventCallDueTooLarge
	}
	if !utf8.ValidString(value) || value == "" || value != strings.TrimSpace(value) {
		return EventCallDueInvalid
	}
	for _, runeValue := range value {
		if unicode.IsControl(runeValue) {
			return EventCallDueInvalid
		}
	}
	return ""
}

func eventCallDueProjectionPreflight(call EventCallDueCallSnapshot, projection EventCallDueProjectionInput) EventCallDueErrorKind {
	event := projection.Event
	target := projection.Target
	creator := projection.Creator
	if event.EventID == uuid.Nil || event.ScheduledCallID == uuid.Nil || event.OrganizationID == uuid.Nil || event.ProjectID == uuid.Nil || event.ChannelID == uuid.Nil || event.OrganizerID == uuid.Nil ||
		target.ScheduledCallID == uuid.Nil || target.EventID == uuid.Nil || target.CreatorID == uuid.Nil || target.OrganizationID == uuid.Nil || target.ProjectID == uuid.Nil || target.ChannelID == uuid.Nil ||
		creator.UserID == uuid.Nil || creator.OrganizationID == uuid.Nil || creator.ProjectID == uuid.Nil || creator.ChannelID == uuid.Nil {
		return EventCallDueInvalid
	}
	if !target.Exists || !target.Active || target.Deleted || event.EventID != call.CalendarEventID || event.ScheduledCallID != call.ScheduledCallID || event.OrganizationID != call.OrganizationID || event.ProjectID != call.ProjectID || event.ChannelID != call.ChannelID || event.OrganizerID != call.CreatorID ||
		target.ScheduledCallID != call.ScheduledCallID || target.EventID != call.CalendarEventID || target.CreatorID != call.CreatorID || target.OrganizationID != call.OrganizationID || target.ProjectID != call.ProjectID || target.ChannelID != call.ChannelID ||
		creator.UserID != call.CreatorID || creator.OrganizationID != call.OrganizationID || creator.ProjectID != call.ProjectID || creator.ChannelID != call.ChannelID {
		return EventCallDueBoundary
	}
	return ""
}

func eventCallDueProjectionFailure(err error) error {
	switch {
	case errors.Is(err, ErrEventCallProjectionInvalid):
		return eventCallDueFailure(EventCallDueInvalid)
	case errors.Is(err, ErrEventCallProjectionBoundary):
		return eventCallDueFailure(EventCallDueBoundary)
	case errors.Is(err, ErrEventCallProjectionTooLarge):
		return eventCallDueFailure(EventCallDueTooLarge)
	default:
		return eventCallDueFailure(EventCallDueInvalid)
	}
}

func eventCallDueProjectionSnapshotIdentityValid(call EventCallDueCallSnapshot, snapshot EventCallProjectionSnapshot) bool {
	return snapshot.ScheduledCallID == call.ScheduledCallID && snapshot.EventID == call.CalendarEventID && snapshot.OrganizationID == call.OrganizationID && snapshot.ProjectID == call.ProjectID && snapshot.ChannelID == call.ChannelID && snapshot.CreatorID == call.CreatorID
}

func eventCallDueProjectionDispositionAllowed(disposition EventCallProjectionDisposition) bool {
	return disposition == EventCallProjectionUpsert || disposition == EventCallProjectionCompleted
}

func eventCallDueEqualInstantDifferentKey(left, right EventCallProjectionSnapshot) bool {
	return left.Disposition == EventCallProjectionUpsert && right.Disposition == EventCallProjectionUpsert && left.Occurrence != nil && right.Occurrence != nil && left.Occurrence.StartsAt.Equal(right.Occurrence.StartsAt) && left.Occurrence.Key != right.Occurrence.Key
}

func eventCallDueOccurrenceValid(occurrence *EventCallProjectionOccurrence) bool {
	if occurrence == nil || occurrence.Key == "" || !eventCallProjectionExactUTCSecond(occurrence.StartsAt) || !eventCallProjectionExactUTCSecond(occurrence.EndsAt) || !occurrence.EndsAt.After(occurrence.StartsAt) {
		return false
	}
	return len(occurrence.Key) == eventCallDueOccurrenceKeyBytes && eventCallProjectionKey(occurrence.StartsAt) != "" && eventCallDueNominalKeyValid(occurrence.Key)
}

func eventCallDueNominalKeyValid(value string) bool {
	if len(value) != eventCallDueOccurrenceKeyBytes {
		return false
	}
	parsed, ok := eventCallProjectionParseKey(value)
	return ok && eventCallProjectionKey(parsed) == value
}

func eventCallDueKey(calendarEventID, scheduledCallID uuid.UUID, nominalKey string) string {
	if calendarEventID == uuid.Nil || scheduledCallID == uuid.Nil || !eventCallDueNominalKeyValid(nominalKey) {
		return ""
	}
	preimage := make([]byte, 0, len(eventCallDueKeyPrefix)+len(calendarEventID)+len(scheduledCallID)+len(nominalKey))
	preimage = append(preimage, eventCallDueKeyPrefix...)
	preimage = append(preimage, calendarEventID[:]...)
	preimage = append(preimage, scheduledCallID[:]...)
	preimage = append(preimage, nominalKey...)
	digest := sha256.Sum256(preimage)
	return hex.EncodeToString(digest[:])
}

func eventCallDueTitleValid(value string) bool {
	if len(value) == 0 || len(value) > eventCallDueMaxTitleBytes || !utf8.ValidString(value) || strings.TrimSpace(value) == "" {
		return false
	}
	for _, runeValue := range value {
		if unicode.IsControl(runeValue) {
			return false
		}
	}
	return true
}

func eventCallDueEnvelopeBindingValid(envelope events.EnvelopeV1, subject string, input EventCallDueInput, want EventCallScheduledDuePayload) bool {
	encoded, err := envelope.MarshalJSON()
	if err != nil || envelope.Subject() != subject || envelope.EventID() != input.Envelope.EventID {
		return false
	}
	var complete map[string]json.RawMessage
	if json.Unmarshal(encoded, &complete) != nil || len(complete) != 8 {
		return false
	}
	for _, field := range []string{"subject", "schema_version", "event_id", "occurred_at", "organization_id", "actor_id", "correlation_id", "payload"} {
		if _, found := complete[field]; !found { return false }
	}
	var decoded struct {
		Subject        string          `json:"subject"`
		SchemaVersion  int             `json:"schema_version"`
		EventID        uuid.UUID       `json:"event_id"`
		OccurredAt     time.Time       `json:"occurred_at"`
		OrganizationID uuid.UUID       `json:"organization_id"`
		ActorID        *uuid.UUID      `json:"actor_id"`
		CorrelationID  string          `json:"correlation_id"`
		Payload        json.RawMessage `json:"payload"`
	}
	if json.Unmarshal(encoded, &decoded) != nil || decoded.Subject != subject || decoded.SchemaVersion != 1 || decoded.EventID != input.Envelope.EventID || !decoded.OccurredAt.Equal(input.Envelope.OccurredAt) || decoded.OrganizationID != input.Call.OrganizationID || decoded.ActorID != nil || decoded.CorrelationID != input.Envelope.CorrelationID {
		return false
	}
	var rawPayload map[string]json.RawMessage
	if json.Unmarshal(decoded.Payload, &rawPayload) != nil || len(rawPayload) != 9 {
		return false
	}
	for _, field := range []string{"scheduled_call_id", "calendar_event_id", "occurrence_key", "recurrence_id", "channel_id", "project_id", "title", "scheduled_at", "duration_minutes"} {
		if _, found := rawPayload[field]; !found { return false }
	}
	var got EventCallScheduledDuePayload
	if json.Unmarshal(decoded.Payload, &got) != nil || got.ScheduledCallID != want.ScheduledCallID || got.CalendarEventID != want.CalendarEventID || got.OccurrenceKey != want.OccurrenceKey || got.ChannelID != want.ChannelID || got.ProjectID != want.ProjectID || got.Title != want.Title || !got.ScheduledAt.Equal(want.ScheduledAt) || got.DurationMinutes != want.DurationMinutes {
		return false
	}
	if (got.RecurrenceID == nil) != (want.RecurrenceID == nil) || got.RecurrenceID != nil && *got.RecurrenceID != *want.RecurrenceID {
		return false
	}
	return got.ScheduledCallID == input.Call.ScheduledCallID && got.CalendarEventID == input.Call.CalendarEventID && got.ProjectID == input.Call.ProjectID && got.ChannelID == input.Call.ChannelID
}

func eventCallDueIdentitySnapshot(call EventCallDueCallSnapshot, disposition EventCallDueDisposition, active bool, scheduledAt *time.Time, key string) EventCallDueSnapshot {
	return EventCallDueSnapshot{
		Disposition: disposition, ScheduledCallID: call.ScheduledCallID, CalendarEventID: call.CalendarEventID, OrganizationID: call.OrganizationID,
		ProjectID: call.ProjectID, ChannelID: call.ChannelID, CreatorID: call.CreatorID, OccurrenceKey: key,
		ScheduledAt: cloneEventCallDueTime(scheduledAt), Active: active,
	}
}

func cloneEventCallDueSnapshot(snapshot EventCallDueSnapshot) EventCallDueSnapshot {
	snapshot.ScheduledAt = cloneEventCallDueTime(snapshot.ScheduledAt)
	return snapshot
}

func cloneEventCallDueTime(value *time.Time) *time.Time {
	if value == nil { return nil }
	clone := *value
	return &clone
}

func cloneEventCallDueProjectionInput(input EventCallDueProjectionInput) EventCallDueProjectionInput {
	clone := input
	clone.Event.Exdates = append([]string(nil), input.Event.Exdates...)
	clone.Event.ReminderMinutes = cloneEventCallProjectionInt(input.Event.ReminderMinutes)
	clone.Attendees = make([]EventCallProjectionAttendee, len(input.Attendees))
	for index, attendee := range input.Attendees {
		clone.Attendees[index] = attendee
		if attendee.UserID != nil { identifier := *attendee.UserID; clone.Attendees[index].UserID = &identifier }
	}
	clone.Exceptions = make([]EventCallProjectionException, len(input.Exceptions))
	for index, exception := range input.Exceptions {
		clone.Exceptions[index] = exception
		if exception.Temporal != nil { temporal := *exception.Temporal; clone.Exceptions[index].Temporal = &temporal }
	}
	return clone
}

func eventCallDueSentinel(kind EventCallDueErrorKind) error {
	switch kind {
	case EventCallDueInvalid: return ErrEventCallDueInvalid
	case EventCallDueBoundary: return ErrEventCallDueBoundary
	case EventCallDueTooLarge: return ErrEventCallDueTooLarge
	default: return nil
	}
}

func eventCallDueFailure(kind EventCallDueErrorKind) error { return &EventCallDueError{kind: kind} }
