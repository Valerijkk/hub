package calendar

import (
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
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
	calendarCreateFingerprintVersion = 1
	calendarCreateMaxAttendees       = 200
	calendarCreateMaxTitleBytes      = 500
	calendarCreateMaxDescription     = 200000
	calendarCreateMaxLocationBytes   = 500
	calendarCreateMaxNameRunes       = 160
	calendarCreateMaxNameBytes       = 640
	calendarCreateMaxTimezoneBytes   = 64
	calendarCreateMaxRRuleBytes      = 1000
	calendarCreateMaxExdates         = 1000
	calendarCreateMaxEmailBytes      = 320
	calendarCreateMaxIdempotencyKey  = 128
	calendarCreateMinIdempotencyKey  = 16
	calendarCreateOccurrenceCap      = 1000
	calendarCreateCandidateWorkGuard = 1001
)

// CalendarEventCreateInput contains untrusted content and identity candidates.
// The in-transaction resolver is the only organizer/attendee authority.
type CalendarEventCreateInput struct {
	IdempotencyKey  string
	Title           string
	Description     string
	Location        string
	Temporal        CalendarEvent
	RRULE           string
	Exdates         []string
	ScheduledCallID *uuid.UUID
	ReminderMinutes *int
	Attendees       []CalendarEventCreateAttendeeCandidate
}

// CalendarEventCreateCommand is an immutable normalized create request.
type CalendarEventCreateCommand struct {
	input       CalendarEventCreateInput
	keyHash     MutationReceiptHash
	requestHash MutationReceiptHash
}

func NewCalendarEventCreateCommand(input CalendarEventCreateInput) (CalendarEventCreateCommand, error) {
	if !validCalendarCreateIdempotencyKey(input.IdempotencyKey) ||
		!validCalendarCreateText(input.Title, calendarCreateMaxTitleBytes, true) ||
		!validCalendarCreateText(input.Description, calendarCreateMaxDescription, false) ||
		!validCalendarCreateText(input.Location, calendarCreateMaxLocationBytes, false) ||
		!utf8.ValidString(input.Temporal.Timezone) ||
		len(input.Temporal.Timezone) > calendarCreateMaxTimezoneBytes ||
		input.Temporal.Validate() != nil ||
		len(input.Attendees) > calendarCreateMaxAttendees {
		return CalendarEventCreateCommand{}, ErrEventInvalid
	}
	if input.ScheduledCallID != nil && *input.ScheduledCallID == uuid.Nil {
		return CalendarEventCreateCommand{}, ErrEventInvalid
	}
	if input.ReminderMinutes != nil && (*input.ReminderMinutes < 0 || *input.ReminderMinutes > 40320) {
		return CalendarEventCreateCommand{}, ErrEventInvalid
	}
	if input.RRULE != "" {
		if len(input.RRULE) > calendarCreateMaxRRuleBytes || !utf8.ValidString(input.RRULE) {
			return CalendarEventCreateCommand{}, ErrEventInvalid
		}
		normalized, err := normalizeCalendarCreateRRULE(input.RRULE, input.Temporal)
		if err != nil {
			return CalendarEventCreateCommand{}, ErrEventInvalid
		}
		input.RRULE = normalized
	}
	var err error
	input.Exdates, err = canonicalizeCalendarCreateExdates(input.Temporal, input.Exdates)
	if err != nil {
		return CalendarEventCreateCommand{}, ErrEventInvalid
	}
	input.ScheduledCallID = copyUUID(input.ScheduledCallID)
	input.ReminderMinutes = copyInt(input.ReminderMinutes)
	attendees := make([]CalendarEventCreateAttendeeCandidate, len(input.Attendees))
	seenEmails := make(map[string]struct{}, len(input.Attendees))
	seenUsers := make(map[uuid.UUID]struct{}, len(input.Attendees))
	for index, candidate := range input.Attendees {
		if !validCalendarCreateEmail(candidate.Email) || !validCalendarCreateDisplayName(candidate.DisplayName) || !validCalendarEventAttendeeRole(candidate.Role) {
			return CalendarEventCreateCommand{}, ErrEventInvalid
		}
		candidate.Email = normalizeCalendarCreateEmail(candidate.Email)
		if _, exists := seenEmails[candidate.Email]; exists {
			return CalendarEventCreateCommand{}, ErrEventInvalid
		}
		seenEmails[candidate.Email] = struct{}{}
		if candidate.UserID != nil {
			if *candidate.UserID == uuid.Nil {
				return CalendarEventCreateCommand{}, ErrEventInvalid
			}
			if _, exists := seenUsers[*candidate.UserID]; exists {
				return CalendarEventCreateCommand{}, ErrEventInvalid
			}
			seenUsers[*candidate.UserID] = struct{}{}
		}
		attendees[index] = cloneCalendarEventCreateAttendeeCandidate(candidate)
	}
	input.Attendees = attendees
	requestHash, err := calendarCreateRequestHash(input)
	if err != nil {
		return CalendarEventCreateCommand{}, ErrEventInvalid
	}
	return CalendarEventCreateCommand{
		input: input, keyHash: MutationReceiptHash(sha256.Sum256([]byte(input.IdempotencyKey))), requestHash: requestHash,
	}, nil
}

func (command CalendarEventCreateCommand) valid() bool {
	rebuilt, err := NewCalendarEventCreateCommand(command.input)
	return err == nil && rebuilt.keyHash == command.keyHash && rebuilt.requestHash == command.requestHash
}

type CalendarEventCreateResult struct {
	response MutationReceiptResponse
	replayed bool
}

func (result CalendarEventCreateResult) Response() MutationReceiptResponse { return result.response }
func (result CalendarEventCreateResult) Replayed() bool { return result.replayed }
func (result CalendarEventCreateResult) valid() bool { return result.response.valid() }

// CreateCalendarEvent orders fresh flow as scope, receipt reservation/replay,
// authorization and writes. Replay intentionally precedes authorization.
func (service *EventService) CreateCalendarEvent(ctx context.Context, actorID uuid.UUID, command CalendarEventCreateCommand) (CalendarEventCreateResult, error) {
	if service == nil || service.repository == nil || service.clock == nil || service.ids == nil || actorID == uuid.Nil || !command.valid() {
		return CalendarEventCreateResult{}, ErrEventInvalid
	}
	var result CalendarEventCreateResult
	err := service.repository.WithEventTransaction(ctx, func(transaction *EventTransaction) error {
		scope, err := transaction.ResolveMutationScope(ctx, actorID, MutationReceiptEventCreate)
		if err != nil {
			return err
		}
		reservation, err := NewMutationReceiptReservation(scope, command.keyHash, command.requestHash)
		if err != nil {
			return err
		}
		reserved, err := transaction.ReserveMutationReceipt(ctx, scope, reservation)
		if err != nil {
			return err
		}
		switch reserved.State() {
		case MutationReceiptReplayed:
			response, found := reserved.ReplayedResponse()
			if !found {
				return ErrEventInvalid
			}
			result = CalendarEventCreateResult{response: response, replayed: true}
			return nil
		case MutationReceiptReserved:
		default:
			return ErrEventInvalid
		}
		authorization, err := transaction.ResolveCalendarEventCreateAuthorization(ctx, scope, calendarCreateAuthorizationRequest(scope, command.input))
		if err != nil {
			return err
		}
		if err := validateCalendarCreateAuthorization(scope, command.input, authorization); err != nil {
			return err
		}
		scopeActorID, scopeOrganizationID := scope.ActorID(), scope.OrganizationID()
		eventID := calendarCreateUUID("hub.calendar.event.id.v1", scopeActorID[:], scopeOrganizationID[:], command.keyHash[:], command.requestHash[:])
		event, organizer, attendees, attendeeIDs, err := calendarCreatePersistenceCommands(scope, eventID, command.input, authorization)
		if err != nil {
			return err
		}
		persistedEvent, err := transaction.PersistCalendarEvent(ctx, scope, event)
		if err != nil {
			return err
		}
		if persistedEvent.EventID() != eventID || persistedEvent.Version() != 1 {
			return ErrEventInvalid
		}
		if err := calendarCreatePersistAttendee(ctx, transaction, scope, organizer); err != nil {
			return err
		}
		for _, attendee := range attendees {
			if err := calendarCreatePersistAttendee(ctx, transaction, scope, attendee); err != nil {
				return err
			}
		}
		sourceEnvelopeID := service.ids.NewEventID()
		if sourceEnvelopeID == uuid.Nil {
			return ErrEventInvalid
		}
		eventSnapshot := event.Snapshot()
		if eventSnapshot.EventID != eventID || eventSnapshot.OrganizerID == nil || *eventSnapshot.OrganizerID == uuid.Nil {
			return ErrEventInvalid
		}
		actor := scope.ActorID()
		envelope, err := NewCalendarInviteDeliveryEnvelopeV2(
			CalendarInviteCreatedSubjectV2, sourceEnvelopeID, service.clock.Now().UTC(), scope.OrganizationID(), &actor, eventID.String(),
			CalendarInviteDeliveryEnvelopeV2Input{
				EventID: eventID, EventSequence: eventSnapshot.Sequence, OrganizerID: *eventSnapshot.OrganizerID,
				RecurrenceID: nil, AttendeeIDs: attendeeIDs,
			},
		)
		if err != nil {
			return ErrEventInvalid
		}
		descriptor, err := NewEventOutboxDescriptor(envelope)
		if err != nil {
			return err
		}
		if err := transaction.AppendOutbox(ctx, scope, descriptor); err != nil {
			return err
		}
		response, err := calendarCreateResponse(eventID, command.input, authorization, attendees, attendeeIDs)
		if err != nil {
			return err
		}
		completion, err := NewMutationReceiptCompletion(reserved.ReceiptID(), &eventID, response)
		if err != nil {
			return err
		}
		if err := transaction.CompleteMutationReceipt(ctx, scope, completion); err != nil {
			return err
		}
		result = CalendarEventCreateResult{response: response}
		return nil
	})
	if err != nil {
		return CalendarEventCreateResult{}, calendarCreateSafeError(err)
	}
	if !result.valid() {
		return CalendarEventCreateResult{}, ErrEventInvalid
	}
	return result, nil
}

func calendarCreateAuthorizationRequest(scope EventMutationScope, input CalendarEventCreateInput) CalendarEventCreateAuthorizationRequest {
	organizerID := scope.ActorID()
	attendees := make([]CalendarEventCreateAttendeeCandidate, len(input.Attendees))
	for index, candidate := range input.Attendees {
		attendees[index] = cloneCalendarEventCreateAttendeeCandidate(candidate)
	}
	return CalendarEventCreateAuthorizationRequest{
		Organizer: CalendarEventCreateAttendeeCandidate{UserID: &organizerID, Role: CalendarEventAttendeeRoleRequired}, Attendees: attendees,
		Title: input.Title, Description: input.Description, Location: input.Location, Temporal: input.Temporal, RRULE: input.RRULE,
		Exdates: append([]string(nil), input.Exdates...), ReminderMinutes: copyInt(input.ReminderMinutes), ScheduledCallID: copyUUID(input.ScheduledCallID),
	}
}

func validateCalendarCreateAuthorization(scope EventMutationScope, input CalendarEventCreateInput, authorization CalendarEventCreateAuthorization) error {
	if !sameCalendarCreateUUID(input.ScheduledCallID, authorization.ScheduledCallID) {
		return newCalendarEventCreateConflict(CalendarEventCreateConflictScheduledCallLinkMismatch)
	}
	if authorization.Organizer.UserID != scope.ActorID() || !validCalendarCreateEmail(authorization.Organizer.Email) ||
		!validCalendarCreateDisplayName(authorization.Organizer.DisplayName) || len(authorization.Attendees) != len(input.Attendees) {
		return ErrEventConflict
	}
	if authorization.ScheduledCallID != nil {
		for _, resolved := range authorization.Attendees {
			if resolved.UserID == nil || *resolved.UserID == uuid.Nil {
				return newCalendarEventCreateConflict(CalendarEventCreateConflictScheduledCallNotAttachable)
			}
		}
	}
	seenEmails := map[string]struct{}{normalizeCalendarCreateEmail(authorization.Organizer.Email): {}}
	seenUsers := map[uuid.UUID]struct{}{authorization.Organizer.UserID: {}}
	for index, resolved := range authorization.Attendees {
		candidate := input.Attendees[index]
		if !validCalendarCreateEmail(resolved.Email) || !validCalendarCreateDisplayName(resolved.DisplayName) ||
			!validCalendarEventAttendeeRole(resolved.Role) || resolved.Role != candidate.Role ||
			normalizeCalendarCreateEmail(resolved.Email) != candidate.Email {
			return ErrEventConflict
		}
		if candidate.UserID != nil && (resolved.UserID == nil || *resolved.UserID != *candidate.UserID) {
			return ErrEventConflict
		}
		if resolved.UserID != nil {
			if *resolved.UserID == uuid.Nil {
				return ErrEventConflict
			}
			if _, exists := seenUsers[*resolved.UserID]; exists {
				return ErrEventConflict
			}
			seenUsers[*resolved.UserID] = struct{}{}
		}
		email := normalizeCalendarCreateEmail(resolved.Email)
		if _, exists := seenEmails[email]; exists {
			return ErrEventConflict
		}
		seenEmails[email] = struct{}{}
	}
	return nil
}

func sameCalendarCreateUUID(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func calendarCreatePersistAttendee(ctx context.Context, transaction *EventTransaction, scope EventMutationScope, attendee CalendarEventAttendeeCommand) error {
	persisted, err := transaction.PersistCalendarEventAttendee(ctx, scope, attendee)
	if err != nil {
		return err
	}
	if persisted.AttendeeID() != attendee.AttendeeID() || persisted.Version() != 1 {
		return ErrEventInvalid
	}
	return nil
}

func calendarCreatePersistenceCommands(scope EventMutationScope, eventID uuid.UUID, input CalendarEventCreateInput, authorization CalendarEventCreateAuthorization) (CalendarEventCommand, CalendarEventAttendeeCommand, []CalendarEventAttendeeCommand, []uuid.UUID, error) {
	organizerID := authorization.Organizer.UserID
	event, err := NewCalendarEventCommand(CalendarEventCommandInput{
		EventID: eventID, OrganizationID: scope.OrganizationID(), OwnerID: scope.ActorID(), UID: "hub-" + eventID.String(),
		Title: input.Title, Description: input.Description, Location: input.Location, Temporal: input.Temporal, RRULE: input.RRULE, Exdates: input.Exdates,
		Sequence: 0, Version: 1, Source: CalendarEventSourceHub, Status: CalendarEventStatusConfirmed, OrganizerID: &organizerID,
		OrganizerEmail: normalizeCalendarCreateEmail(authorization.Organizer.Email), ScheduledCallID: copyUUID(authorization.ScheduledCallID), ReminderMinutes: copyInt(input.ReminderMinutes),
	})
	if err != nil {
		return CalendarEventCommand{}, CalendarEventAttendeeCommand{}, nil, nil, err
	}
	organizerAttendeeID := calendarCreateUUID("hub.calendar.attendee.id.v1", eventID[:], []byte("organizer"))
	organizer, err := newCalendarEventOrganizerAttendeeCommand(CalendarEventAttendeeCommandInput{
		AttendeeID: organizerAttendeeID, EventID: eventID, UserID: &organizerID, Email: normalizeCalendarCreateEmail(authorization.Organizer.Email),
		DisplayName: calendarCreateDisplayName(authorization.Organizer.DisplayName), Role: CalendarEventAttendeeRoleOrganizer, Status: RSVPAccepted, Version: 1,
	})
	if err != nil {
		return CalendarEventCommand{}, CalendarEventAttendeeCommand{}, nil, nil, err
	}
	attendees := make([]CalendarEventAttendeeCommand, 0, len(authorization.Attendees))
	attendeeIDs := make([]uuid.UUID, 0, len(authorization.Attendees))
	for index, resolved := range authorization.Attendees {
		attendeeID := calendarCreateUUID("hub.calendar.attendee.id.v1", eventID[:], calendarCreateIndexBytes(index))
		attendee, err := NewCalendarEventAttendeeCommand(CalendarEventAttendeeCommandInput{
			AttendeeID: attendeeID, EventID: eventID, UserID: copyUUID(resolved.UserID), Email: normalizeCalendarCreateEmail(resolved.Email),
			DisplayName: calendarCreateDisplayName(resolved.DisplayName), Role: resolved.Role, Status: RSVPNeedsAction, Version: 1,
		})
		if err != nil {
			return CalendarEventCommand{}, CalendarEventAttendeeCommand{}, nil, nil, err
		}
		attendees = append(attendees, attendee)
		attendeeIDs = append(attendeeIDs, attendeeID)
	}
	return event, organizer, attendees, attendeeIDs, nil
}

func calendarCreateDisplayName(value *string) string {
	if value == nil { return "" }
	return *value
}

func calendarCreateResponse(eventID uuid.UUID, input CalendarEventCreateInput, authorization CalendarEventCreateAuthorization, attendees []CalendarEventAttendeeCommand, attendeeIDs []uuid.UUID) (MutationReceiptResponse, error) {
	responseAttendees := make([]calendarCreateResponseAttendee, 0, len(authorization.Attendees))
	for index, resolved := range authorization.Attendees {
		responseAttendees = append(responseAttendees, calendarCreateResponseAttendee{
			ID: attendeeIDs[index], UserID: copyUUID(resolved.UserID), Email: normalizeCalendarCreateEmail(resolved.Email), DisplayName: cloneCalendarEventCreateString(resolved.DisplayName),
			Role: resolved.Role, RSVPStatus: RSVPNeedsAction, ReplyDeliveryStatus: "not_required", Version: attendees[index].Version(),
		})
	}
	startsAt, endsAt, startDate, endDateExclusive := calendarCreateTemporalResponse(input.Temporal)
	organizerID := authorization.Organizer.UserID
	encoded, err := json.Marshal(calendarCreateResponseBody{
		ID: eventID, UID: "hub-" + eventID.String(), Title: input.Title, Description: input.Description, Location: input.Location,
		StartsAt: startsAt, EndsAt: endsAt, StartDate: startDate, EndDateExclusive: endDateExclusive, AllDay: input.Temporal.AllDay,
		Timezone: input.Temporal.Timezone, RRULE: input.RRULE, Exdates: append([]string(nil), input.Exdates...), Sequence: 0,
		Source: CalendarEventSourceHub, Status: CalendarEventStatusConfirmed, OrganizerID: &organizerID, OrganizerEmail: normalizeCalendarCreateEmail(authorization.Organizer.Email),
		ScheduledCallID: copyUUID(authorization.ScheduledCallID), ReminderMinutes: copyInt(input.ReminderMinutes), Attendees: responseAttendees, Version: 1,
	})
	if err != nil { return MutationReceiptResponse{}, ErrEventInvalid }
	return NewMutationReceiptResponse(201, encoded)
}

type calendarCreateResponseBody struct {
	ID uuid.UUID `json:"id"`; UID string `json:"uid"`; Title string `json:"title"`; Description string `json:"description,omitempty"`; Location string `json:"location,omitempty"`
	StartsAt *time.Time `json:"starts_at"`; EndsAt *time.Time `json:"ends_at"`; StartDate *string `json:"start_date"`; EndDateExclusive *string `json:"end_date_exclusive"`
	AllDay bool `json:"all_day"`; Timezone string `json:"timezone"`; RRULE string `json:"rrule,omitempty"`; Exdates []string `json:"exdates,omitempty"`; Sequence uint `json:"sequence"`
	Source CalendarEventSource `json:"source"`; Status CalendarEventStatus `json:"status"`; OrganizerID *uuid.UUID `json:"organizer_id"`; OrganizerEmail string `json:"organizer_email"`
	ScheduledCallID *uuid.UUID `json:"scheduled_call_id"`; ReminderMinutes *int `json:"reminder_minutes"`; Attendees []calendarCreateResponseAttendee `json:"attendees"`; Version uint `json:"version"`
}

type calendarCreateResponseAttendee struct {
	ID uuid.UUID `json:"id"`; UserID *uuid.UUID `json:"user_id"`; Email string `json:"email"`; DisplayName *string `json:"display_name"`; Role CalendarEventAttendeeRole `json:"role"`
	RSVPStatus RSVPStatus `json:"rsvp_status"`; RecurrenceID *string `json:"recurrence_id"`; RespondedAt *time.Time `json:"responded_at"`; Version uint `json:"version"`; ReplyDeliveryStatus string `json:"reply_delivery_status"`
}

func calendarCreateTemporalResponse(temporal CalendarEvent) (*time.Time, *time.Time, *string, *string) {
	if temporal.AllDay { startDate, endDateExclusive := temporal.StartDate, temporal.EndDateExclusive; return nil, nil, &startDate, &endDateExclusive }
	startsAt, endsAt := temporal.StartsAt.UTC(), temporal.EndsAt.UTC()
	return &startsAt, &endsAt, nil, nil
}

// normalizeCalendarCreateRRULE implements the closed section 3.11 create grammar.
// It deliberately does not share the inbound iTIP parser or its repair rules.
func normalizeCalendarCreateRRULE(value string, temporal CalendarEvent) (string, error) {
	if value == "" || len(value) > calendarCreateMaxRRuleBytes || !utf8.ValidString(value) {
		return "", ErrEventInvalid
	}
	fields := make(map[string]string, 6)
	for _, part := range strings.Split(value, ";") {
		key, rawValue, found := strings.Cut(part, "=")
		if !found || key == "" || rawValue == "" || strings.Contains(rawValue, "=") {
			return "", ErrEventInvalid
		}
		switch key {
		case "FREQ", "INTERVAL", "BYDAY", "BYMONTHDAY", "COUNT", "UNTIL":
		default:
			return "", ErrEventInvalid
		}
		if _, duplicate := fields[key]; duplicate {
			return "", ErrEventInvalid
		}
		fields[key] = rawValue
	}

	frequency, found := fields["FREQ"]
	if !found || (frequency != "DAILY" && frequency != "WEEKLY" && frequency != "MONTHLY") {
		return "", ErrEventInvalid
	}
	interval := 1
	if rawInterval, found := fields["INTERVAL"]; found {
		parsed, err := parseCalendarCreatePositiveInteger(rawInterval, 365)
		if err != nil {
			return "", ErrEventInvalid
		}
		interval = parsed
	}

	byday := ""
	byMonthDay := ""
	switch frequency {
	case "DAILY":
		if _, found := fields["BYDAY"]; found {
			return "", ErrEventInvalid
		}
		if _, found := fields["BYMONTHDAY"]; found {
			return "", ErrEventInvalid
		}
	case "WEEKLY":
		rawDays, found := fields["BYDAY"]
		if !found {
			return "", ErrEventInvalid
		}
		if _, found := fields["BYMONTHDAY"]; found {
			return "", ErrEventInvalid
		}
		normalizedDays, err := normalizeCalendarCreateWeekdays(rawDays)
		if err != nil {
			return "", ErrEventInvalid
		}
		byday = normalizedDays
	case "MONTHLY":
		rawMonthDay, found := fields["BYMONTHDAY"]
		if !found {
			return "", ErrEventInvalid
		}
		if _, found := fields["BYDAY"]; found {
			return "", ErrEventInvalid
		}
		parsed, err := parseCalendarCreatePositiveInteger(rawMonthDay, 31)
		if err != nil {
			return "", ErrEventInvalid
		}
		byMonthDay = strconv.Itoa(parsed)
	}

	rawCount, hasCount := fields["COUNT"]
	rawUntil, hasUntil := fields["UNTIL"]
	if hasCount == hasUntil {
		return "", ErrEventInvalid
	}
	count := ""
	if hasCount {
		parsed, err := parseCalendarCreatePositiveInteger(rawCount, 1000)
		if err != nil {
			return "", ErrEventInvalid
		}
		count = strconv.Itoa(parsed)
	}
	until := ""
	if hasUntil {
		var err error
		if temporal.AllDay {
			until, err = normalizeCalendarCreateAllDayUntil(rawUntil)
		} else {
			until, err = normalizeCalendarCreateTimedUntil(rawUntil)
		}
		if err != nil {
			return "", ErrEventInvalid
		}
		if err := validateCalendarCreateOccurrenceCap(temporal, frequency, interval, byday, byMonthDay, until); err != nil {
			return "", ErrEventInvalid
		}
	}

	normalized := "FREQ=" + frequency + ";INTERVAL=" + strconv.Itoa(interval)
	if byday != "" {
		normalized += ";BYDAY=" + byday
	}
	if byMonthDay != "" {
		normalized += ";BYMONTHDAY=" + byMonthDay
	}
	if count != "" {
		normalized += ";COUNT=" + count
	}
	if until != "" {
		normalized += ";UNTIL=" + until
	}
	return normalized, nil
}

func parseCalendarCreatePositiveInteger(value string, maximum int) (int, error) {
	if value == "" {
		return 0, ErrEventInvalid
	}
	for index := 0; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return 0, ErrEventInvalid
		}
	}
	parsed, err := strconv.Atoi(value)
	if err != nil || parsed < 1 || parsed > maximum {
		return 0, ErrEventInvalid
	}
	return parsed, nil
}

func normalizeCalendarCreateWeekdays(value string) (string, error) {
	order := map[string]int{"MO": 0, "TU": 1, "WE": 2, "TH": 3, "FR": 4, "SA": 5, "SU": 6}
	days := strings.Split(value, ",")
	if len(days) < 1 || len(days) > 7 {
		return "", ErrEventInvalid
	}
	seen := make(map[string]struct{}, len(days))
	for _, day := range days {
		if _, found := order[day]; !found {
			return "", ErrEventInvalid
		}
		if _, duplicate := seen[day]; duplicate {
			return "", ErrEventInvalid
		}
		seen[day] = struct{}{}
	}
	sort.Slice(days, func(left, right int) bool { return order[days[left]] < order[days[right]] })
	return strings.Join(days, ","), nil
}

func normalizeCalendarCreateTimedUntil(value string) (string, error) {
	parsed, err := time.Parse("20060102T150405Z", value)
	if err != nil || parsed.Location() != time.UTC || parsed.Format("20060102T150405Z") != value {
		return "", ErrEventInvalid
	}
	return parsed.Format("20060102T150405Z"), nil
}

func normalizeCalendarCreateAllDayUntil(value string) (string, error) {
	parsed, err := time.Parse("20060102", value)
	if err != nil || parsed.Format("20060102") != value {
		return "", ErrEventInvalid
	}
	return parsed.Format("20060102"), nil
}

// validateCalendarCreateOccurrenceCap proves an UNTIL-bounded series has no
// more than the global nominal occurrence cap without materializing it.
func validateCalendarCreateOccurrenceCap(temporal CalendarEvent, frequency string, interval int, byday, byMonthDay, until string) error {
	if temporal.AllDay {
		return validateCalendarCreateAllDayOccurrenceCap(temporal, frequency, interval, byday, byMonthDay, until)
	}
	untilTime, err := time.Parse("20060102T150405Z", until)
	if err != nil || temporal.StartsAt.After(untilTime) {
		return ErrEventInvalid
	}
	location, err := time.LoadLocation(temporal.Timezone)
	if err != nil {
		return ErrEventInvalid
	}
	return validateCalendarCreateTimedOccurrenceCap(temporal.StartsAt, untilTime, location, frequency, interval, byday, byMonthDay)
}

func validateCalendarCreateAllDayOccurrenceCap(temporal CalendarEvent, frequency string, interval int, byday, byMonthDay, until string) error {
	startDate, err := parseCalendarDate(temporal.StartDate)
	if err != nil {
		return ErrEventInvalid
	}
	untilDate, err := time.Parse("20060102", until)
	if err != nil || untilDate.Before(startDate) {
		return ErrEventInvalid
	}
	work := 1 // DTSTART is the first nominal occurrence.
	occurrences := 1
	process := func(candidate time.Time) (bool, error) {
		if !candidate.After(startDate) {
			return false, nil
		}
		if candidate.After(untilDate) {
			return true, nil
		}
		work++
		if work >= calendarCreateCandidateWorkGuard {
			return false, ErrEventInvalid
		}
		occurrences++
		if occurrences > calendarCreateOccurrenceCap {
			return false, ErrEventInvalid
		}
		return false, nil
	}

	switch frequency {
	case "DAILY":
		for sequence := 1; ; sequence++ {
			candidate := startDate.AddDate(0, 0, sequence*interval)
			stop, err := process(candidate)
			if err != nil || stop { return err }
		}
	case "WEEKLY":
		weekdays, err := calendarCreateWeekdayOffsets(byday)
		if err != nil { return ErrEventInvalid }
		weekStart := calendarCreateMonday(startDate)
		for group := 0; ; group++ {
			base := weekStart.AddDate(0, 0, group*interval*7)
			for _, weekday := range weekdays {
				stop, err := process(base.AddDate(0, 0, weekday))
				if err != nil || stop { return err }
			}
		}
	case "MONTHLY":
		monthDay, err := strconv.Atoi(byMonthDay)
		if err != nil { return ErrEventInvalid }
		for sequence := 0; ; sequence++ {
			month := time.Date(startDate.Year(), startDate.Month()+time.Month(sequence*interval), 1, 0, 0, 0, 0, time.UTC)
			if monthDay > calendarCreateDaysInMonth(month.Year(), month.Month()) {
				if month.After(untilDate) { return nil }
				work++
				if work >= calendarCreateCandidateWorkGuard { return ErrEventInvalid }
				continue
			}
			candidate := time.Date(month.Year(), month.Month(), monthDay, 0, 0, 0, 0, time.UTC)
			stop, err := process(candidate)
			if err != nil || stop { return err }
		}
	}
	return ErrEventInvalid
}

func validateCalendarCreateTimedOccurrenceCap(start, until time.Time, location *time.Location, frequency string, interval int, byday, byMonthDay string) error {
	start = start.UTC()
	localStart := calendarCreateCivilTimeFromTime(start.In(location))
	localUntil := calendarCreateCivilTimeFromTime(until.In(location))
	work := 1 // DTSTART counts once, before exclusions outside this pure packet.
	occurrences := 1
	process := func(civil calendarCreateCivilTime) (bool, error) {
		if calendarCreateCompareCivilTime(civil, localStart) <= 0 {
			return false, nil
		}
		candidate, exists := calendarCreateResolveCivilTime(location, civil)
		if exists && candidate.After(until) {
			return true, nil
		}
		if !exists && calendarCreateCompareCivilTime(civil, localUntil) > 0 {
			return true, nil
		}
		work++
		if work >= calendarCreateCandidateWorkGuard {
			return false, ErrEventInvalid
		}
		if !exists { // RFC 5545: a nonexistent local wall time is ignored.
			return false, nil
		}
		occurrences++
		if occurrences > calendarCreateOccurrenceCap {
			return false, ErrEventInvalid
		}
		return false, nil
	}

	switch frequency {
	case "DAILY":
		for sequence := 1; ; sequence++ {
			date := time.Date(localStart.Year, localStart.Month, localStart.Day, 0, 0, 0, 0, time.UTC).AddDate(0, 0, sequence*interval)
			civil := localStart
			civil.Year, civil.Month, civil.Day = date.Date()
			stop, err := process(civil)
			if err != nil || stop { return err }
		}
	case "WEEKLY":
		weekdays, err := calendarCreateWeekdayOffsets(byday)
		if err != nil { return ErrEventInvalid }
		weekStart := calendarCreateMonday(time.Date(localStart.Year, localStart.Month, localStart.Day, 0, 0, 0, 0, time.UTC))
		for group := 0; ; group++ {
			base := weekStart.AddDate(0, 0, group*interval*7)
			for _, weekday := range weekdays {
				civil := localStart
				civil.Year, civil.Month, civil.Day = base.AddDate(0, 0, weekday).Date()
				stop, err := process(civil)
				if err != nil || stop { return err }
			}
		}
	case "MONTHLY":
		monthDay, err := strconv.Atoi(byMonthDay)
		if err != nil { return ErrEventInvalid }
		for sequence := 0; ; sequence++ {
			month := time.Date(localStart.Year, localStart.Month+time.Month(sequence*interval), 1, 0, 0, 0, 0, time.UTC)
			if monthDay > calendarCreateDaysInMonth(month.Year(), month.Month()) {
				if calendarCreateCompareCivilTime(calendarCreateCivilTime{Year: month.Year(), Month: month.Month(), Day: 1}, localUntil) > 0 { return nil }
				work++
				if work >= calendarCreateCandidateWorkGuard { return ErrEventInvalid }
				continue
			}
			civil := localStart
			civil.Year, civil.Month, civil.Day = month.Year(), month.Month(), monthDay
			stop, err := process(civil)
			if err != nil || stop { return err }
		}
	}
	return ErrEventInvalid
}

type calendarCreateCivilTime struct {
	Year int
	Month time.Month
	Day int
	Hour int
	Minute int
	Second int
	Nanosecond int
}

func calendarCreateCivilTimeFromTime(value time.Time) calendarCreateCivilTime {
	return calendarCreateCivilTime{Year: value.Year(), Month: value.Month(), Day: value.Day(), Hour: value.Hour(), Minute: value.Minute(), Second: value.Second(), Nanosecond: value.Nanosecond()}
}

func calendarCreateCompareCivilTime(left, right calendarCreateCivilTime) int {
	for _, pair := range [][2]int{{left.Year, right.Year}, {int(left.Month), int(right.Month)}, {left.Day, right.Day}, {left.Hour, right.Hour}, {left.Minute, right.Minute}, {left.Second, right.Second}, {left.Nanosecond, right.Nanosecond}} {
		if pair[0] < pair[1] { return -1 }
		if pair[0] > pair[1] { return 1 }
	}
	return 0
}

func calendarCreateResolveCivilTime(location *time.Location, civil calendarCreateCivilTime) (time.Time, bool) {
	naive := time.Date(civil.Year, civil.Month, civil.Day, civil.Hour, civil.Minute, civil.Second, civil.Nanosecond, time.UTC)
	offsets := make(map[int]struct{}, 17)
	for hours := -48; hours <= 48; hours += 6 {
		_, offset := naive.Add(time.Duration(hours) * time.Hour).In(location).Zone()
		offsets[offset] = struct{}{}
	}
	var earliest time.Time
	found := false
	for offset := range offsets {
		candidate := naive.Add(-time.Duration(offset) * time.Second)
		if calendarCreateCompareCivilTime(calendarCreateCivilTimeFromTime(candidate.In(location)), civil) != 0 { continue }
		if !found || candidate.Before(earliest) {
			earliest, found = candidate, true
		}
	}
	return earliest, found
}

func calendarCreateWeekdayOffsets(value string) ([]int, error) {
	offsets := map[string]int{"MO": 0, "TU": 1, "WE": 2, "TH": 3, "FR": 4, "SA": 5, "SU": 6}
	parts := strings.Split(value, ",")
	result := make([]int, len(parts))
	for index, part := range parts {
		offset, found := offsets[part]
		if !found { return nil, ErrEventInvalid }
		result[index] = offset
	}
	return result, nil
}

func calendarCreateMonday(value time.Time) time.Time {
	offset := (int(value.Weekday()) + 6) % 7
	return value.AddDate(0, 0, -offset)
}

func calendarCreateDaysInMonth(year int, month time.Month) int {
	return time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

func canonicalizeCalendarCreateExdates(temporal CalendarEvent, exdates []string) ([]string, error) {
	if len(exdates) > calendarCreateMaxExdates { return nil, ErrEventInvalid }
	if len(exdates) == 0 { return nil, nil }
	canonical := make([]string, len(exdates))
	for index, value := range exdates {
		if temporal.AllDay {
			if _, err := parseCalendarDate(value); err != nil { return nil, ErrEventInvalid }
		} else {
			parsed, err := time.Parse("2006-01-02T15:04:05Z", value)
			if err != nil || parsed.Location() != time.UTC || parsed.Format("2006-01-02T15:04:05Z") != value { return nil, ErrEventInvalid }
		}
		canonical[index] = value
	}
	sort.Strings(canonical)
	for index := 1; index < len(canonical); index++ { if canonical[index-1] == canonical[index] { return nil, ErrEventInvalid } }
	return canonical, nil
}

func calendarCreateRequestHash(input CalendarEventCreateInput) (MutationReceiptHash, error) {
	encoded, err := json.Marshal(calendarCreateFingerprint{Version: calendarCreateFingerprintVersion, Title: input.Title, Description: input.Description, Location: input.Location,
		Temporal: input.Temporal, RRULE: input.RRULE, Exdates: input.Exdates, ScheduledCallID: input.ScheduledCallID, ReminderMinutes: input.ReminderMinutes, Attendees: input.Attendees})
	if err != nil { return MutationReceiptHash{}, err }
	return MutationReceiptHash(sha256.Sum256(encoded)), nil
}

type calendarCreateFingerprint struct {
	Version int `json:"version"`; Title string `json:"title"`; Description string `json:"description"`; Location string `json:"location"`; Temporal CalendarEvent `json:"temporal"`
	RRULE string `json:"rrule"`; Exdates []string `json:"exdates"`; ScheduledCallID *uuid.UUID `json:"scheduled_call_id"`; ReminderMinutes *int `json:"reminder_minutes"`; Attendees []CalendarEventCreateAttendeeCandidate `json:"attendees"`
}

func calendarCreateUUID(domain string, components ...[]byte) uuid.UUID {
	hash := sha256.New(); _, _ = hash.Write([]byte(domain)); _, _ = hash.Write([]byte{0})
	for _, component := range components { var length [4]byte; binary.BigEndian.PutUint32(length[:], uint32(len(component))); _, _ = hash.Write(length[:]); _, _ = hash.Write(component) }
	sum := hash.Sum(nil); var identifier uuid.UUID; copy(identifier[:], sum); identifier[6] = (identifier[6] & 0x0f) | 0x50; identifier[8] = (identifier[8] & 0x3f) | 0x80
	return identifier
}

func calendarCreateIndexBytes(index int) []byte { var value [4]byte; binary.BigEndian.PutUint32(value[:], uint32(index)); return value[:] }

func calendarCreateSafeError(err error) error {
	var createConflict *CalendarEventCreateConflictError
	if errors.As(err, &createConflict) && createConflict != nil && validCalendarEventCreateConflictKind(createConflict.Kind()) {
		return newCalendarEventCreateConflict(createConflict.Kind())
	}
	switch { case errors.Is(err, ErrEventNotFound): return ErrEventNotFound; case errors.Is(err, ErrEventForbidden): return ErrEventForbidden; case errors.Is(err, ErrEventConflict): return ErrEventConflict; case errors.Is(err, ErrEventTransactionClosed): return ErrEventTransactionClosed; default: return ErrEventInvalid }
}

func validCalendarCreateIdempotencyKey(value string) bool { return utf8.ValidString(value) && len(value) >= calendarCreateMinIdempotencyKey && len(value) <= calendarCreateMaxIdempotencyKey && strings.TrimSpace(value) == value }

func validCalendarCreateText(value string, maximum int, required bool) bool {
	if !utf8.ValidString(value) || len(value) > maximum || (required && strings.TrimSpace(value) == "") { return false }
	for _, runeValue := range value { if unicode.IsControl(runeValue) { return false } }
	return true
}

func validCalendarCreateDisplayName(value *string) bool {
	return value == nil || (utf8.ValidString(*value) && len(*value) <= calendarCreateMaxNameBytes && utf8.RuneCountInString(*value) <= calendarCreateMaxNameRunes)
}

func validCalendarCreateEmail(value string) bool {
	value = normalizeCalendarCreateEmail(value)
	if !utf8.ValidString(value) || len(value) == 0 || len(value) > calendarCreateMaxEmailBytes || strings.Count(value, "@") != 1 { return false }
	for _, runeValue := range value { if runeValue > unicode.MaxASCII || unicode.IsSpace(runeValue) || unicode.IsControl(runeValue) { return false } }
	local, domain, found := strings.Cut(value, "@")
	return found && local != "" && domain != "" && !strings.HasPrefix(domain, ".") && !strings.HasSuffix(domain, ".") && !strings.Contains(domain, "..")
}

func normalizeCalendarCreateEmail(value string) string { return strings.ToLower(strings.TrimSpace(value)) }
