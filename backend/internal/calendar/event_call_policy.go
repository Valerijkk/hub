package calendar

import (
	"bytes"
	"errors"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/google/uuid"
)

const (
	eventCallProjectionMaxAttendees   = 200
	eventCallProjectionMaxTitleBytes  = 255
	eventCallProjectionMaxDescription = 2000
	eventCallProjectionMaxEmailBytes  = 320
	eventCallProjectionMaxTimezone    = 64
	eventCallProjectionMaxRRuleBytes  = 1000
	eventCallProjectionMaxExdates     = 1000
	eventCallProjectionMaxExceptions  = 1000
	eventCallProjectionMaxReminder    = 40320
	eventCallProjectionMaxDuration    = 1440
	eventCallProjectionKeyBytes       = 20
)

// EventCallProjectionAction selects the requested call projection transition.
type EventCallProjectionAction string

const (
	EventCallProjectionReconcile EventCallProjectionAction = "reconcile"
	EventCallProjectionPause     EventCallProjectionAction = "pause"
	EventCallProjectionResume    EventCallProjectionAction = "resume"
	EventCallProjectionDelete    EventCallProjectionAction = "delete"
)

// EventCallProjectionCreator is the resolved organizer membership snapshot.
type EventCallProjectionCreator struct {
	UserID              uuid.UUID
	Email               string
	OrganizationID      uuid.UUID
	ProjectID           uuid.UUID
	ChannelID           uuid.UUID
	Active              bool
	ActiveProjectMember bool
	ActiveChannelMember bool
}

// EventCallProjectionAttendee is a resolved calendar attendee snapshot.
type EventCallProjectionAttendee struct {
	UserID              *uuid.UUID
	Email               string
	OrganizationID      uuid.UUID
	ProjectID           uuid.UUID
	Role                CalendarEventAttendeeRole
	Status              RSVPStatus
	RecurrenceID        string
	Active              bool
	ActiveProjectMember bool
}

// EventCallProjectionException describes one canonical recurrence exception.
type EventCallProjectionException struct {
	RecurrenceID string
	Temporal     *CalendarEvent
	Status       CalendarEventStatus
	Cancelled    bool
}

// EventCallProjectionEvent is the canonical calendar event snapshot.
type EventCallProjectionEvent struct {
	EventID         uuid.UUID
	ScheduledCallID uuid.UUID
	OrganizationID  uuid.UUID
	ProjectID       uuid.UUID
	ChannelID       uuid.UUID
	OrganizerID     uuid.UUID
	OrganizerEmail  string
	Source          CalendarEventSource
	Status          CalendarEventStatus
	Deleted         bool
	Title           string
	Description     string
	Temporal        CalendarEvent
	RRULE           string
	Exdates         []string
	ReminderMinutes *int
}

// EventCallProjectionTarget is the observed scheduled-call projection state.
type EventCallProjectionTarget struct {
	ScheduledCallID uuid.UUID
	EventID         uuid.UUID
	CreatorID       uuid.UUID
	OrganizationID  uuid.UUID
	ProjectID       uuid.UUID
	ChannelID       uuid.UUID
	Exists          bool
	Active          bool
	Deleted         bool
}

// EventCallProjectionBoundary controls whether an occurrence at At is selected.
type EventCallProjectionBoundary struct {
	At        time.Time
	Inclusive bool
}

// EventCallProjectionInput contains the complete immutable policy input.
type EventCallProjectionInput struct {
	Action     EventCallProjectionAction
	Event      EventCallProjectionEvent
	Target     EventCallProjectionTarget
	Creator    EventCallProjectionCreator
	Attendees  []EventCallProjectionAttendee
	Exceptions []EventCallProjectionException
	Boundary   EventCallProjectionBoundary
}

// EventCallProjectionDisposition is the only allowed projected call state.
type EventCallProjectionDisposition string

const (
	EventCallProjectionUpsert    EventCallProjectionDisposition = "upsert"
	EventCallProjectionPaused    EventCallProjectionDisposition = "paused"
	EventCallProjectionCompleted EventCallProjectionDisposition = "completed"
	EventCallProjectionDeleted   EventCallProjectionDisposition = "deleted"
)

// EventCallProjectionOccurrence is an effective call occurrence keyed by its nominal start.
type EventCallProjectionOccurrence struct {
	Key      string
	StartsAt time.Time
	EndsAt   time.Time
}

// EventCallProjectionSnapshot is the deep-copying read model of a projection.
type EventCallProjectionSnapshot struct {
	Disposition     EventCallProjectionDisposition
	DeletedAt       *time.Time
	ScheduledCallID uuid.UUID
	EventID         uuid.UUID
	OrganizationID  uuid.UUID
	ProjectID       uuid.UUID
	ChannelID       uuid.UUID
	CreatorID       uuid.UUID
	Title           string
	Description     string
	DurationMinutes int
	ReminderMinutes *int
	RRULE           string
	Occurrence      *EventCallProjectionOccurrence
	ParticipantIDs  []uuid.UUID
}

// EventCallProjection keeps its snapshot opaque to prevent aliasing its policy result.
type EventCallProjection struct {
	snapshot EventCallProjectionSnapshot
}

// Snapshot returns a fresh deep copy of the policy result.
func (projection EventCallProjection) Snapshot() EventCallProjectionSnapshot {
	return cloneEventCallProjectionSnapshot(projection.snapshot)
}

// EventCallProjectionErrorKind classifies content-free policy errors.
type EventCallProjectionErrorKind string

const (
	EventCallProjectionInvalid  EventCallProjectionErrorKind = "invalid"
	EventCallProjectionBoundary EventCallProjectionErrorKind = "boundary"
	EventCallProjectionTooLarge EventCallProjectionErrorKind = "too_large"
)

// EventCallProjectionError keeps the failure classification opaque to callers.
type EventCallProjectionError struct {
	kind EventCallProjectionErrorKind
}

var (
	ErrEventCallProjectionInvalid  error = errors.New("calendar: event call projection invalid")
	ErrEventCallProjectionBoundary error = errors.New("calendar: event call projection boundary mismatch")
	ErrEventCallProjectionTooLarge error = errors.New("calendar: event call projection too large")
)

// Error intentionally contains no event, identity, recurrence or content data.
func (err *EventCallProjectionError) Error() string {
	if cause := err.Unwrap(); cause != nil {
		return cause.Error()
	}
	return "calendar: event call projection failed"
}

// Kind returns a known kind only for a known internal policy failure.
func (err *EventCallProjectionError) Kind() EventCallProjectionErrorKind {
	if err == nil || eventCallProjectionSentinel(err.kind) == nil {
		return ""
	}
	return err.kind
}

// Unwrap exposes only the frozen content-free sentinel.
func (err *EventCallProjectionError) Unwrap() error {
	if err == nil {
		return nil
	}
	return eventCallProjectionSentinel(err.kind)
}

// Is supports errors.Is for the three frozen sentinels.
func (err *EventCallProjectionError) Is(target error) bool {
	cause := err.Unwrap()
	return cause != nil && target == cause
}

func eventCallProjectionSentinel(kind EventCallProjectionErrorKind) error {
	switch kind {
	case EventCallProjectionInvalid:
		return ErrEventCallProjectionInvalid
	case EventCallProjectionBoundary:
		return ErrEventCallProjectionBoundary
	case EventCallProjectionTooLarge:
		return ErrEventCallProjectionTooLarge
	default:
		return nil
	}
}

func eventCallProjectionFailure(kind EventCallProjectionErrorKind) error {
	return &EventCallProjectionError{kind: kind}
}

// ProjectEventCall maps a fully resolved canonical event snapshot to one pure call projection.
func ProjectEventCall(input EventCallProjectionInput) (EventCallProjection, error) {
	if kind := eventCallProjectionPreflight(input); kind != "" {
		return EventCallProjection{}, eventCallProjectionFailure(kind)
	}
	if !eventCallProjectionActionKnown(input.Action) || !eventCallProjectionIdentityValid(input) || !eventCallProjectionEventStateValid(input.Event) ||
		!eventCallProjectionCreatorValid(input.Event, input.Creator) || !eventCallProjectionBoundaryValid(input.Boundary) {
		return EventCallProjection{}, eventCallProjectionFailure(EventCallProjectionInvalid)
	}

	if _, kind := eventCallProjectionValidateTemporal(input.Event.Temporal); kind != "" {
		return EventCallProjection{}, eventCallProjectionFailure(kind)
	}
	for _, exception := range input.Exceptions {
		if !eventCallProjectionExceptionStateValid(exception) {
			return EventCallProjection{}, eventCallProjectionFailure(EventCallProjectionInvalid)
		}
		if exception.Temporal != nil {
			if _, kind := eventCallProjectionValidateTemporal(*exception.Temporal); kind != "" {
				return EventCallProjection{}, eventCallProjectionFailure(kind)
			}
		}
	}
	if input.Event.RRULE == "" && len(input.Exceptions) != 0 {
		return EventCallProjection{}, eventCallProjectionFailure(EventCallProjectionInvalid)
	}

	nominal, kind := eventCallProjectionNominalOccurrences(input.Event)
	if kind != "" {
		return EventCallProjection{}, eventCallProjectionFailure(kind)
	}
	effective, kind := eventCallProjectionEffectiveOccurrences(nominal, input.Event.Exdates, input.Exceptions)
	if kind != "" {
		return EventCallProjection{}, eventCallProjectionFailure(kind)
	}
	participants, kind := eventCallProjectionParticipants(input.Event, input.Creator, input.Attendees, nominal)
	if kind != "" {
		return EventCallProjection{}, eventCallProjectionFailure(kind)
	}

	if !eventCallProjectionActionAllowed(input.Action, input.Event, input.Target) {
		return EventCallProjection{}, eventCallProjectionFailure(EventCallProjectionBoundary)
	}
	identity := eventCallProjectionIdentitySnapshot(input.Event, input.Creator)
	switch input.Action {
	case EventCallProjectionPause:
		identity.Disposition = EventCallProjectionPaused
		return EventCallProjection{snapshot: identity}, nil
	case EventCallProjectionDelete:
		identity.Disposition = EventCallProjectionDeleted
		deletedAt := input.Boundary.At
		identity.DeletedAt = &deletedAt
		return EventCallProjection{snapshot: identity}, nil
	case EventCallProjectionReconcile, EventCallProjectionResume:
		occurrence, found := eventCallProjectionSelectOccurrence(effective, input.Boundary)
		if !found {
			identity.Disposition = EventCallProjectionCompleted
			return EventCallProjection{snapshot: identity}, nil
		}
		return EventCallProjection{snapshot: EventCallProjectionSnapshot{
			Disposition:     EventCallProjectionUpsert,
			ScheduledCallID: input.Event.ScheduledCallID,
			EventID:         input.Event.EventID,
			OrganizationID:  input.Event.OrganizationID,
			ProjectID:       input.Event.ProjectID,
			ChannelID:       input.Event.ChannelID,
			CreatorID:       input.Creator.UserID,
			Title:           input.Event.Title,
			Description:     input.Event.Description,
			DurationMinutes: int(occurrence.EndsAt.Sub(occurrence.StartsAt) / time.Minute),
			ReminderMinutes: cloneEventCallProjectionInt(input.Event.ReminderMinutes),
			RRULE:           input.Event.RRULE,
			Occurrence:      &occurrence,
			ParticipantIDs:  append([]uuid.UUID(nil), participants...),
		}}, nil
	default:
		return EventCallProjection{}, eventCallProjectionFailure(EventCallProjectionInvalid)
	}
}

func eventCallProjectionPreflight(input EventCallProjectionInput) EventCallProjectionErrorKind {
	if len(input.Attendees) > eventCallProjectionMaxAttendees || len(input.Event.Exdates) > eventCallProjectionMaxExdates || len(input.Exceptions) > eventCallProjectionMaxExceptions ||
		len(input.Event.Title) > eventCallProjectionMaxTitleBytes || len(input.Event.Description) > eventCallProjectionMaxDescription || len(input.Event.RRULE) > eventCallProjectionMaxRRuleBytes ||
		len(input.Event.OrganizerEmail) > eventCallProjectionMaxEmailBytes || len(input.Creator.Email) > eventCallProjectionMaxEmailBytes || len(input.Event.Temporal.Timezone) > eventCallProjectionMaxTimezone {
		return EventCallProjectionTooLarge
	}
	if kind := eventCallProjectionEmailKind(input.Event.OrganizerEmail); kind != "" {
		return kind
	}
	if kind := eventCallProjectionEmailKind(input.Creator.Email); kind != "" {
		return kind
	}
	for _, attendee := range input.Attendees {
		if len(attendee.Email) > eventCallProjectionMaxEmailBytes {
			return EventCallProjectionTooLarge
		}
		if kind := eventCallProjectionEmailKind(attendee.Email); kind != "" {
			return kind
		}
	}
	if kind := eventCallProjectionTimezoneKind(input.Event.Temporal.Timezone); kind != "" {
		return kind
	}
	for _, exdate := range input.Event.Exdates {
		if kind := eventCallProjectionKeyKind(exdate); kind != "" {
			return kind
		}
	}
	for _, exception := range input.Exceptions {
		if kind := eventCallProjectionKeyKind(exception.RecurrenceID); kind != "" {
			return kind
		}
	}
	for _, attendee := range input.Attendees {
		if attendee.RecurrenceID != "" {
			if kind := eventCallProjectionKeyKind(attendee.RecurrenceID); kind != "" {
				return kind
			}
		}
	}
	for _, exception := range input.Exceptions {
		if exception.Temporal != nil {
			if len(exception.Temporal.Timezone) > eventCallProjectionMaxTimezone {
				return EventCallProjectionTooLarge
			}
			if kind := eventCallProjectionTimezoneKind(exception.Temporal.Timezone); kind != "" {
				return kind
			}
		}
	}
	if !eventCallProjectionTextValid(input.Event.Title, true) || !eventCallProjectionTextValid(input.Event.Description, false) {
		return EventCallProjectionInvalid
	}
	if input.Event.ReminderMinutes != nil {
		if *input.Event.ReminderMinutes < 0 {
			return EventCallProjectionInvalid
		}
		if *input.Event.ReminderMinutes > eventCallProjectionMaxReminder {
			return EventCallProjectionTooLarge
		}
	}
	return ""
}

func eventCallProjectionIdentityValid(input EventCallProjectionInput) bool {
	event := input.Event
	target := input.Target
	creator := input.Creator
	if event.EventID == uuid.Nil || event.ScheduledCallID == uuid.Nil || event.OrganizationID == uuid.Nil || event.ProjectID == uuid.Nil || event.ChannelID == uuid.Nil || event.OrganizerID == uuid.Nil ||
		target.ScheduledCallID == uuid.Nil || target.EventID == uuid.Nil || target.CreatorID == uuid.Nil || target.OrganizationID == uuid.Nil || target.ProjectID == uuid.Nil || target.ChannelID == uuid.Nil ||
		creator.UserID == uuid.Nil || creator.OrganizationID == uuid.Nil || creator.ProjectID == uuid.Nil || creator.ChannelID == uuid.Nil {
		return false
	}
	if target.ScheduledCallID != event.ScheduledCallID || target.EventID != event.EventID || target.CreatorID != event.OrganizerID ||
		target.OrganizationID != event.OrganizationID || target.ProjectID != event.ProjectID || target.ChannelID != event.ChannelID ||
		creator.UserID != event.OrganizerID || creator.OrganizationID != event.OrganizationID || creator.ProjectID != event.ProjectID || creator.ChannelID != event.ChannelID {
		return false
	}
	if !target.Exists {
		return !target.Active && !target.Deleted
	}
	return !target.Deleted || !target.Active
}

func eventCallProjectionEventStateValid(event EventCallProjectionEvent) bool {
	if event.Source != CalendarEventSourceHub {
		return false
	}
	switch event.Status {
	case CalendarEventStatusConfirmed, CalendarEventStatusTentative:
		return !event.Deleted
	case CalendarEventStatusCancelled:
		return true
	default:
		return false
	}
}

func eventCallProjectionCreatorValid(event EventCallProjectionEvent, creator EventCallProjectionCreator) bool {
	return creator.Active && creator.ActiveProjectMember && creator.ActiveChannelMember && creator.UserID == event.OrganizerID && creator.Email == event.OrganizerEmail
}

func eventCallProjectionBoundaryValid(boundary EventCallProjectionBoundary) bool {
	return eventCallProjectionExactUTCSecond(boundary.At)
}

func eventCallProjectionEmailKind(value string) EventCallProjectionErrorKind {
	if len(value) > eventCallProjectionMaxEmailBytes {
		return EventCallProjectionTooLarge
	}
	if !utf8.ValidString(value) || value != normalizeCalendarCreateEmail(value) || !validCalendarCreateEmail(value) {
		return EventCallProjectionInvalid
	}
	return ""
}

func eventCallProjectionTimezoneKind(value string) EventCallProjectionErrorKind {
	if len(value) > eventCallProjectionMaxTimezone {
		return EventCallProjectionTooLarge
	}
	if value == "" || value == "Local" || !utf8.ValidString(value) {
		return EventCallProjectionInvalid
	}
	for _, runeValue := range value {
		if unicode.IsControl(runeValue) {
			return EventCallProjectionInvalid
		}
	}
	return ""
}

func eventCallProjectionTextValid(value string, required bool) bool {
	if !utf8.ValidString(value) || (required && strings.TrimSpace(value) == "") {
		return false
	}
	for _, runeValue := range value {
		if unicode.IsControl(runeValue) {
			return false
		}
	}
	return true
}

func eventCallProjectionValidateTemporal(temporal CalendarEvent) (int, EventCallProjectionErrorKind) {
	if temporal.AllDay || temporal.StartsAt.IsZero() || temporal.EndsAt.IsZero() || temporal.StartDate != "" || temporal.EndDateExclusive != "" ||
		!eventCallProjectionExactUTCSecond(temporal.StartsAt) || !eventCallProjectionExactUTCSecond(temporal.EndsAt) {
		return 0, EventCallProjectionInvalid
	}
	if kind := eventCallProjectionTimezoneKind(temporal.Timezone); kind != "" {
		return 0, kind
	}
	if _, err := time.LoadLocation(temporal.Timezone); err != nil {
		return 0, EventCallProjectionInvalid
	}
	duration := temporal.EndsAt.Sub(temporal.StartsAt)
	if duration <= 0 {
		return 0, EventCallProjectionInvalid
	}
	minutes := duration / time.Minute
	if minutes > eventCallProjectionMaxDuration {
		return 0, EventCallProjectionTooLarge
	}
	if duration%time.Minute != 0 || minutes < 1 {
		return 0, EventCallProjectionInvalid
	}
	return int(minutes), ""
}

func eventCallProjectionExactUTCSecond(value time.Time) bool {
	key := value.Format("2006-01-02T15:04:05Z")
	return !value.IsZero() && value.Location() == time.UTC && value.Year() >= 0 && value.Year() <= 9999 && value.Nanosecond() == 0 && len(key) == eventCallProjectionKeyBytes && value.Format(time.RFC3339) == key
}

func eventCallProjectionExceptionStateValid(exception EventCallProjectionException) bool {
	if _, ok := eventCallProjectionParseKey(exception.RecurrenceID); !ok {
		return false
	}
	if exception.Cancelled {
		return exception.Status == CalendarEventStatusCancelled
	}
	return exception.Status == CalendarEventStatusConfirmed || exception.Status == CalendarEventStatusTentative
}

func eventCallProjectionNominalOccurrences(event EventCallProjectionEvent) ([]EventCallProjectionOccurrence, EventCallProjectionErrorKind) {
	master := EventCallProjectionOccurrence{Key: eventCallProjectionKey(event.Temporal.StartsAt), StartsAt: event.Temporal.StartsAt, EndsAt: event.Temporal.EndsAt}
	if event.RRULE == "" {
		if len(event.Exdates) != 0 {
			return nil, EventCallProjectionInvalid
		}
		return []EventCallProjectionOccurrence{master}, ""
	}
	normalized, err := normalizeCalendarCreateRRULE(event.RRULE, event.Temporal)
	if err != nil {
		if kind := eventCallProjectionRRuleLimitKind(event); kind != "" {
			return nil, kind
		}
		return nil, EventCallProjectionInvalid
	}
	if normalized != event.RRULE {
		return nil, EventCallProjectionInvalid
	}
	fields := eventCallProjectionRRuleFields(event.RRULE)
	frequency := fields["FREQ"]
	interval, _ := strconv.Atoi(fields["INTERVAL"])
	count := 0
	if value, found := fields["COUNT"]; found {
		count, _ = strconv.Atoi(value)
	}
	var until time.Time
	hasUntil := false
	if value, found := fields["UNTIL"]; found {
		var ok bool
		until, ok = eventCallProjectionParseUntil(value)
		if !ok {
			return nil, EventCallProjectionInvalid
		}
		hasUntil = true
	}
	location, err := time.LoadLocation(event.Temporal.Timezone)
	if err != nil {
		return nil, EventCallProjectionInvalid
	}
	duration := event.Temporal.EndsAt.Sub(event.Temporal.StartsAt)
	occurrences := []EventCallProjectionOccurrence{master}
	localUntil := calendarCreateCivilTime{}
	if hasUntil {
		localUntil = calendarCreateCivilTimeFromTime(until.In(location))
	}
	appendCivil := func(civil calendarCreateCivilTime) (bool, EventCallProjectionErrorKind) {
		if count > 0 && len(occurrences) >= count {
			return true, ""
		}
		candidate, exists := calendarCreateResolveCivilTime(location, civil)
		if !exists {
			if hasUntil && eventCallProjectionCivilCompare(civil, localUntil) > 0 {
				return true, ""
			}
			return false, ""
		}
		candidate = candidate.UTC()
		if hasUntil && candidate.After(until) {
			return true, ""
		}
		candidateEndsAt := candidate.Add(duration)
		if !eventCallProjectionExactUTCSecond(candidate) || !eventCallProjectionExactUTCSecond(candidateEndsAt) {
			return false, EventCallProjectionTooLarge
		}
		if len(occurrences) >= calendarCreateOccurrenceCap {
			return false, EventCallProjectionTooLarge
		}
		occurrences = append(occurrences, EventCallProjectionOccurrence{Key: eventCallProjectionKey(candidate), StartsAt: candidate, EndsAt: candidate.Add(duration)})
		return false, ""
	}
	localStart := calendarCreateCivilTimeFromTime(event.Temporal.StartsAt.In(location))
	work := 0
	switch frequency {
	case "DAILY":
		for sequence := 1; ; sequence++ {
			work++
			if work > calendarCreateCandidateWorkGuard {
				return nil, EventCallProjectionTooLarge
			}
			date := time.Date(localStart.Year, localStart.Month, localStart.Day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, sequence*interval)
			civil := localStart
			civil.Year, civil.Month, civil.Day = date.Date()
			stop, kind := appendCivil(civil)
			if kind != "" || stop {
				return occurrences, kind
			}
		}
	case "WEEKLY":
		weekdays, err := calendarCreateWeekdayOffsets(fields["BYDAY"])
		if err != nil {
			return nil, EventCallProjectionInvalid
		}
		weekStart := calendarCreateMonday(time.Date(localStart.Year, localStart.Month, localStart.Day, 0, 0, 0, 0, time.UTC))
		for group := 0; ; group++ {
			for _, weekday := range weekdays {
				date := weekStart.AddDate(0, 0, group*interval*7+weekday)
				civil := localStart
				civil.Year, civil.Month, civil.Day = date.Date()
				if eventCallProjectionCivilCompare(civil, localStart) <= 0 {
					continue
				}
				work++
				if work > calendarCreateCandidateWorkGuard {
					return nil, EventCallProjectionTooLarge
				}
				stop, kind := appendCivil(civil)
				if kind != "" || stop {
					return occurrences, kind
				}
			}
		}
	case "MONTHLY":
		monthDay, err := strconv.Atoi(fields["BYMONTHDAY"])
		if err != nil {
			return nil, EventCallProjectionInvalid
		}
		for sequence := 0; ; sequence++ {
			month := time.Date(localStart.Year, localStart.Month+time.Month(sequence*interval), 1, 0, 0, 0, 0, time.UTC)
			if monthDay > calendarCreateDaysInMonth(month.Year(), month.Month()) {
				work++
				if work > calendarCreateCandidateWorkGuard {
					return nil, EventCallProjectionTooLarge
				}
				continue
			}
			civil := localStart
			civil.Year, civil.Month, civil.Day = month.Year(), month.Month(), monthDay
			if eventCallProjectionCivilCompare(civil, localStart) <= 0 {
				continue
			}
			work++
			if work > calendarCreateCandidateWorkGuard {
				return nil, EventCallProjectionTooLarge
			}
			stop, kind := appendCivil(civil)
			if kind != "" || stop {
				return occurrences, kind
			}
		}
	default:
		return nil, EventCallProjectionInvalid
	}
}

func eventCallProjectionActionKnown(action EventCallProjectionAction) bool {
	switch action {
	case EventCallProjectionReconcile, EventCallProjectionPause, EventCallProjectionResume, EventCallProjectionDelete:
		return true
	default:
		return false
	}
}

func eventCallProjectionRRuleFields(value string) map[string]string {
	fields := make(map[string]string, 6)
	for _, part := range strings.Split(value, ";") {
		key, rawValue, _ := strings.Cut(part, "=")
		fields[key] = rawValue
	}
	return fields
}

func eventCallProjectionRRuleLimitKind(event EventCallProjectionEvent) EventCallProjectionErrorKind {
	parts := strings.Split(event.RRULE, ";")
	countIndex := -1
	countValue := ""
	untilValue := ""
	for index, part := range parts {
		key, value, found := strings.Cut(part, "=")
		if !found {
			return ""
		}
		switch key {
		case "COUNT":
			if countIndex >= 0 {
				return ""
			}
			countIndex, countValue = index, value
		case "UNTIL":
			if untilValue != "" {
				return ""
			}
			untilValue = value
		}
	}
	if countIndex >= 0 && eventCallProjectionPositiveCountOverCap(countValue) {
		candidate := append([]string(nil), parts...)
		candidate[countIndex] = "COUNT=1"
		candidateRule := strings.Join(candidate, ";")
		if normalized, err := normalizeCalendarCreateRRULE(candidateRule, event.Temporal); err == nil && normalized == candidateRule {
			return EventCallProjectionTooLarge
		}
	}
	if untilValue != "" {
		until, ok := eventCallProjectionParseUntil(untilValue)
		if !ok {
			return ""
		}
		if until.Before(event.Temporal.StartsAt) {
			return ""
		}
		temporal := event.Temporal
		duration := temporal.EndsAt.Sub(temporal.StartsAt)
		temporal.StartsAt = until
		temporal.EndsAt = until.Add(duration)
		if normalized, err := normalizeCalendarCreateRRULE(event.RRULE, temporal); err == nil && normalized == event.RRULE {
			return EventCallProjectionTooLarge
		}
	}
	return ""
}

func eventCallProjectionPositiveCountOverCap(value string) bool {
	if value == "" {
		return false
	}
	for index := 0; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	parsed, err := strconv.ParseUint(value, 10, 64)
	return err != nil || parsed > calendarCreateOccurrenceCap
}

func eventCallProjectionParseUntil(value string) (time.Time, bool) {
	parsed, err := time.Parse("20060102T150405Z", value)
	return parsed, err == nil && parsed.Location() == time.UTC && parsed.Format("20060102T150405Z") == value
}

func eventCallProjectionCivilCompare(left, right calendarCreateCivilTime) int {
	return calendarCreateCompareCivilTime(left, right)
}

func eventCallProjectionEffectiveOccurrences(nominal []EventCallProjectionOccurrence, exdates []string, exceptions []EventCallProjectionException) ([]EventCallProjectionOccurrence, EventCallProjectionErrorKind) {
	byKey := make(map[string]EventCallProjectionOccurrence, len(nominal))
	for _, occurrence := range nominal {
		byKey[occurrence.Key] = occurrence
	}
	excluded := make(map[string]struct{}, len(exdates))
	previous := ""
	for _, exdate := range exdates {
		if _, ok := eventCallProjectionParseKey(exdate); !ok || exdate <= previous {
			return nil, EventCallProjectionInvalid
		}
		if _, exists := byKey[exdate]; !exists {
			return nil, EventCallProjectionInvalid
		}
		excluded[exdate] = struct{}{}
		previous = exdate
	}
	byException := make(map[string]EventCallProjectionException, len(exceptions))
	for _, exception := range exceptions {
		if _, exists := byKey[exception.RecurrenceID]; !exists {
			return nil, EventCallProjectionInvalid
		}
		if _, overlaps := excluded[exception.RecurrenceID]; overlaps {
			return nil, EventCallProjectionInvalid
		}
		if _, duplicate := byException[exception.RecurrenceID]; duplicate {
			return nil, EventCallProjectionInvalid
		}
		byException[exception.RecurrenceID] = exception
	}
	effective := make([]EventCallProjectionOccurrence, 0, len(nominal))
	starts := make(map[string]struct{}, len(nominal))
	for _, occurrence := range nominal {
		if _, isExcluded := excluded[occurrence.Key]; isExcluded {
			continue
		}
		if exception, found := byException[occurrence.Key]; found {
			if exception.Cancelled {
				continue
			}
			if exception.Temporal != nil {
				occurrence.StartsAt = exception.Temporal.StartsAt
				occurrence.EndsAt = exception.Temporal.EndsAt
			}
		}
		startKey := eventCallProjectionKey(occurrence.StartsAt)
		if _, duplicate := starts[startKey]; duplicate {
			return nil, EventCallProjectionInvalid
		}
		starts[startKey] = struct{}{}
		effective = append(effective, occurrence)
	}
	sort.Slice(effective, func(left, right int) bool {
		if effective[left].StartsAt.Equal(effective[right].StartsAt) {
			return effective[left].Key < effective[right].Key
		}
		return effective[left].StartsAt.Before(effective[right].StartsAt)
	})
	return effective, ""
}

func eventCallProjectionParticipants(event EventCallProjectionEvent, creator EventCallProjectionCreator, attendees []EventCallProjectionAttendee, nominal []EventCallProjectionOccurrence) ([]uuid.UUID, EventCallProjectionErrorKind) {
	nominalKeys := make(map[string]struct{}, len(nominal))
	for _, occurrence := range nominal {
		nominalKeys[occurrence.Key] = struct{}{}
	}
	seenUsers := make(map[uuid.UUID]struct{}, len(attendees))
	seenEmails := make(map[string]struct{}, len(attendees))
	participants := make([]uuid.UUID, 0, len(attendees))
	organizers := 0
	for _, attendee := range attendees {
		if attendee.UserID == nil || *attendee.UserID == uuid.Nil || attendee.OrganizationID != event.OrganizationID || attendee.ProjectID != event.ProjectID || !attendee.Active || !attendee.ActiveProjectMember ||
			!validCalendarEventStoredAttendeeRoleState(attendee.Role, attendee.Status, attendee.RecurrenceID) {
			return nil, EventCallProjectionInvalid
		}
		if _, duplicate := seenUsers[*attendee.UserID]; duplicate {
			return nil, EventCallProjectionInvalid
		}
		if _, duplicate := seenEmails[attendee.Email]; duplicate {
			return nil, EventCallProjectionInvalid
		}
		seenUsers[*attendee.UserID] = struct{}{}
		seenEmails[attendee.Email] = struct{}{}
		if attendee.RecurrenceID != "" {
			if event.RRULE == "" {
				return nil, EventCallProjectionInvalid
			}
			if _, found := nominalKeys[attendee.RecurrenceID]; !found {
				return nil, EventCallProjectionInvalid
			}
			continue
		}
		organizerIdentity := *attendee.UserID == event.OrganizerID && *attendee.UserID == creator.UserID
		if attendee.Role == CalendarEventAttendeeRoleOrganizer || organizerIdentity {
			if attendee.Email != event.OrganizerEmail || attendee.Email != creator.Email ||
				!validCalendarEventStoredOrganizerAttendeeState(attendee.Role, attendee.Status, attendee.RecurrenceID, attendee.UserID, event.OrganizerID, event.OrganizerID) {
				return nil, EventCallProjectionInvalid
			}
			organizers++
			continue
		}
		if attendee.Status == RSVPNeedsAction || attendee.Status == RSVPAccepted || attendee.Status == RSVPTentative {
			participants = append(participants, *attendee.UserID)
		}
	}
	if organizers != 1 {
		return nil, EventCallProjectionInvalid
	}
	sort.Slice(participants, func(left, right int) bool { return bytes.Compare(participants[left][:], participants[right][:]) < 0 })
	return participants, ""
}

func eventCallProjectionActionAllowed(action EventCallProjectionAction, event EventCallProjectionEvent, target EventCallProjectionTarget) bool {
	activeEvent := !event.Deleted && (event.Status == CalendarEventStatusConfirmed || event.Status == CalendarEventStatusTentative)
	switch action {
	case EventCallProjectionReconcile:
		return activeEvent && (!target.Exists || (target.Active && !target.Deleted))
	case EventCallProjectionPause:
		return activeEvent && target.Exists && target.Active && !target.Deleted
	case EventCallProjectionResume:
		return activeEvent && target.Exists && !target.Active && !target.Deleted
	case EventCallProjectionDelete:
		return target.Exists && !target.Deleted && (event.Deleted || event.Status == CalendarEventStatusCancelled)
	default:
		return false
	}
}

func eventCallProjectionSelectOccurrence(occurrences []EventCallProjectionOccurrence, boundary EventCallProjectionBoundary) (EventCallProjectionOccurrence, bool) {
	for _, occurrence := range occurrences {
		if occurrence.StartsAt.After(boundary.At) || (boundary.Inclusive && occurrence.StartsAt.Equal(boundary.At)) {
			return occurrence, true
		}
	}
	return EventCallProjectionOccurrence{}, false
}

func eventCallProjectionIdentitySnapshot(event EventCallProjectionEvent, creator EventCallProjectionCreator) EventCallProjectionSnapshot {
	return EventCallProjectionSnapshot{
		ScheduledCallID: event.ScheduledCallID,
		EventID:         event.EventID,
		OrganizationID:  event.OrganizationID,
		ProjectID:       event.ProjectID,
		ChannelID:       event.ChannelID,
		CreatorID:       creator.UserID,
	}
}

func eventCallProjectionParseKey(value string) (time.Time, bool) {
	if len(value) != eventCallProjectionKeyBytes {
		return time.Time{}, false
	}
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil || parsed.Location() != time.UTC || parsed.Year() < 0 || parsed.Year() > 9999 || parsed.Nanosecond() != 0 || parsed.Format("2006-01-02T15:04:05Z") != value {
		return time.Time{}, false
	}
	return parsed, true
}

func eventCallProjectionKeyKind(value string) EventCallProjectionErrorKind {
	if len(value) > eventCallProjectionKeyBytes {
		return EventCallProjectionTooLarge
	}
	if _, ok := eventCallProjectionParseKey(value); !ok {
		return EventCallProjectionInvalid
	}
	return ""
}

func eventCallProjectionKey(value time.Time) string {
	return value.Format("2006-01-02T15:04:05Z")
}

func cloneEventCallProjectionSnapshot(snapshot EventCallProjectionSnapshot) EventCallProjectionSnapshot {
	snapshot.DeletedAt = cloneEventCallProjectionTime(snapshot.DeletedAt)
	snapshot.ReminderMinutes = cloneEventCallProjectionInt(snapshot.ReminderMinutes)
	if snapshot.Occurrence != nil {
		occurrence := *snapshot.Occurrence
		snapshot.Occurrence = &occurrence
	}
	snapshot.ParticipantIDs = append([]uuid.UUID(nil), snapshot.ParticipantIDs...)
	return snapshot
}

func cloneEventCallProjectionInt(value *int) *int {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func cloneEventCallProjectionTime(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}
